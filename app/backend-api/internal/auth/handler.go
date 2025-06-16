package auth

import (
	"errors"
	"net/http"

	sharedaudit "backend-api/internal/shared/audit"
	"backend-api/internal/shared/response"
	"backend-api/internal/shared/token"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	staffSvc    *Service
	customerSvc *CustomerService
	adminSvc    *AdminService
	frontendURL string
	totpEncKey  string // AES-256-GCM key for TOTP secret encryption (hex, may be empty in dev)
}

func NewHandler(staffSvc *Service, customerSvc *CustomerService, adminSvc *AdminService, frontendURL, totpEncKey string) *Handler {
	return &Handler{
		staffSvc:    staffSvc,
		customerSvc: customerSvc,
		adminSvc:    adminSvc,
		frontendURL: frontendURL,
		totpEncKey:  totpEncKey,
	}
}

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, authMW, requireStaff, rateLimitMW gin.HandlerFunc) {
	staff := rg.Group("/auth/staff")

	// Public (rate-limited by IP)
	public := staff.Group("")
	public.Use(rateLimitMW)
	public.POST("/register", h.StaffRegister)
	public.POST("/login", h.StaffLogin)
	public.POST("/refresh", h.StaffRefreshToken)
	public.POST("/forgot-password", h.ForgotPassword)
	public.POST("/reset-password", h.ResetPassword)

	// Protected
	protected := staff.Group("")
	protected.Use(authMW, requireStaff)
	{
		protected.GET("/me", h.StaffMe)
		protected.PATCH("/me", h.StaffUpdateProfile)
		protected.PATCH("/me/password", h.StaffChangePassword)
		protected.POST("/select-shop", h.StaffSelectShop)

		// 2FA (require existing full auth to setup/disable)
		protected.POST("/2fa/setup", h.TwoFASetup)
		protected.POST("/2fa/verify", h.TwoFAVerify)
		protected.POST("/2fa/disable", h.TwoFADisable)
	}

	// 2FA challenge + recovery — unauthenticated (exchange login_token)
	staff.POST("/2fa/challenge", h.TwoFAChallenge)
	staff.POST("/2fa/recovery", h.TwoFARecovery)
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, authMW, requireCustomer, rateLimitMW gin.HandlerFunc) {
	cust := rg.Group("/auth/customer")

	// Public (rate-limited by IP)
	public := cust.Group("")
	public.Use(rateLimitMW)
	public.POST("/register", h.CustomerRegister)
	public.POST("/login", h.CustomerLogin)
	public.POST("/refresh", h.CustomerRefreshToken)

	// Protected
	protected := cust.Group("")
	protected.Use(authMW, requireCustomer)
	{
		protected.GET("/me", h.CustomerMe)
		protected.PATCH("/me", h.CustomerUpdateMe)
		protected.PATCH("/me/password", h.CustomerChangePassword)
	}
}

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup, authMW, requireAdmin, rateLimitMW gin.HandlerFunc) {
	admin := rg.Group("/auth/admin")

	// Public (rate-limited by IP)
	public := admin.Group("")
	public.Use(rateLimitMW)
	public.POST("/login", h.AdminLogin)
	public.POST("/refresh", h.AdminRefreshToken)

	// Protected
	protected := admin.Group("")
	protected.Use(authMW, requireAdmin)
	{
		protected.GET("/me", h.AdminMe)
	}
}

func (h *Handler) StaffRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.staffSvc.Register(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *Handler) StaffLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.staffSvc.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// If 2FA challenge, return 200 with requires_2fa payload
	if challenge, ok := result.(*TwoFAChallengeResponse); ok {
		c.JSON(http.StatusOK, challenge)
		return
	}

	response.OK(c, result)
}

func (h *Handler) StaffRefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.staffSvc.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) StaffMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}

	result, err := h.staffSvc.Me(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) StaffSelectShop(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}

	var req SelectShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.staffSvc.SelectShop(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) CustomerRegister(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required (X-Shop-ID header)")
		return
	}

	var req CustomerRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.customerSvc.Register(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *Handler) CustomerLogin(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required (X-Shop-ID header)")
		return
	}

	var req CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.customerSvc.Login(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) CustomerRefreshToken(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required (X-Shop-ID header)")
		return
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.customerSvc.RefreshToken(c.Request.Context(), shopID, req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) CustomerMe(c *gin.Context) {
	userID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if userID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	result, err := h.customerSvc.Me(c.Request.Context(), userID, shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) CustomerUpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if userID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	var req CustomerUpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.customerSvc.UpdateMe(c.Request.Context(), userID, shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) CustomerChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if userID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	var req CustomerChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.customerSvc.ChangePassword(c.Request.Context(), userID, shopID, req); err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "password updated"})
}

func (h *Handler) RegisterShopRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	shop := rg.Group("/shops")
	protected := shop.Group("")
	protected.Use(requireAuth, requireStaff)
	{
		protected.GET("/settings", h.GetShopSettings)
		protected.PATCH("/settings", h.UpdateShopSettings)
		protected.GET("/staff", h.ListStaff)
		protected.POST("/staff", h.AddStaff)
		protected.DELETE("/staff/:id", h.RemoveStaff)
	}
}

func (h *Handler) GetShopSettings(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required (X-Shop-ID header)")
		return
	}

	result, err := h.staffSvc.GetShopSettings(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) UpdateShopSettings(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required (X-Shop-ID header)")
		return
	}

	var req UpdateShopSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.staffSvc.UpdateShopSettings(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) ListStaff(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	result, err := h.staffSvc.ListStaff(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) AddStaff(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	var req AddStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.staffSvc.AddStaff(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.staffSvc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("user_name"),
		Action:       "staff.add",
		ResourceType: "shop_staff",
		ResourceID:   result.ID,
		Changes:      map[string]any{"email": req.Email},
	})

	response.OK(c, result)
}

func (h *Handler) RemoveStaff(c *gin.Context) {
	shopID := c.GetString("shop_id")
	staffID := c.Param("id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	if err := h.staffSvc.RemoveStaff(c.Request.Context(), shopID, staffID); err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.staffSvc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("user_name"),
		Action:       "staff.remove",
		ResourceType: "shop_staff",
		ResourceID:   staffID,
	})

	response.OK(c, gin.H{"message": "staff member removed"})
}

func (h *Handler) StaffUpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}
	var req StaffUpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.staffSvc.UpdateStaffProfile(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) StaffChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}
	var req StaffChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.staffSvc.StaffChangePassword(c.Request.Context(), userID, req); err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "password updated"})
}

func (h *Handler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminSvc.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) AdminRefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) AdminMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing admin context")
		return
	}

	result, err := h.adminSvc.Me(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Always return 200 to prevent email enumeration.
	_ = h.staffSvc.ForgotPassword(c.Request.Context(), ForgotPasswordRequest{
		Email:       body.Email,
		FrontendURL: h.frontendURL,
	})
	response.OK(c, gin.H{"message": "If that email is registered, a reset link has been sent."})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.staffSvc.ResetPassword(c.Request.Context(), req); err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "Password updated successfully."})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		response.Conflict(c, "email already registered")
	case errors.Is(err, ErrCustomerEmailTaken):
		response.Conflict(c, "email already registered for this store")
	case errors.Is(err, ErrSubdomainTaken):
		response.Conflict(c, "subdomain already taken")
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(c, "invalid email or password")
	case errors.Is(err, ErrAccountInactive):
		response.Unauthorized(c, "account is not active")
	case errors.Is(err, ErrNotStaff):
		response.Unauthorized(c, "not a member of this shop")
	case errors.Is(err, ErrPlanNotFound):
		response.BadRequest(c, "plan not found")
	case errors.Is(err, ErrShopNotFound):
		response.NotFound(c, "shop not found")
	case errors.Is(err, ErrUserNotFound):
		response.NotFound(c, "user not found")
	case errors.Is(err, ErrCustomerNotFound):
		response.NotFound(c, "customer not found")
	case errors.Is(err, ErrAdminNotFound):
		response.NotFound(c, "admin not found")
	case errors.Is(err, ErrAlreadyMember):
		response.Conflict(c, "user is already a member of this shop")
	case errors.Is(err, ErrStaffLimit):
		response.UnprocessableEntity(c, "You have reached the maximum number of staff accounts for your current plan. Please upgrade.")
	case errors.Is(err, ErrInvalidResetToken):
		response.BadRequest(c, "invalid or expired password reset token")
	case errors.Is(err, ErrRateLimited):
		response.TooManyRequests(c, "too many failed login attempts, please try again later")
	case errors.Is(err, Err2FANotEnabled):
		response.BadRequest(c, "2FA is not enabled on this account")
	case errors.Is(err, Err2FAAlreadyOn):
		response.Conflict(c, "2FA is already enabled")
	case errors.Is(err, ErrInvalid2FACode):
		response.Unauthorized(c, "invalid 2FA code")
	case errors.Is(err, ErrInvalidBackup):
		response.Unauthorized(c, "invalid or already-used backup code")
	case errors.Is(err, token.ErrInvalidToken), errors.Is(err, token.ErrTokenExpired):
		response.Unauthorized(c, "invalid or expired token")
	default:
		response.InternalError(c)
	}
}

func (h *Handler) TwoFASetup(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}
	result, err := h.staffSvc.Setup2FA(c.Request.Context(), userID, h.totpEncKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) TwoFAVerify(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.staffSvc.Verify2FA(c.Request.Context(), userID, req.Code, h.totpEncKey); err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "2FA enabled successfully"})
}

func (h *Handler) TwoFAChallenge(c *gin.Context) {
	var req Challenge2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.staffSvc.Challenge2FA(c.Request.Context(), req, h.totpEncKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) TwoFADisable(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "missing user context")
		return
	}
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.staffSvc.Disable2FA(c.Request.Context(), userID, req.Code, h.totpEncKey); err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "2FA disabled"})
}

func (h *Handler) TwoFARecovery(c *gin.Context) {
	var req RecoverAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.staffSvc.RecoverAccount(c.Request.Context(), req, h.totpEncKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}
