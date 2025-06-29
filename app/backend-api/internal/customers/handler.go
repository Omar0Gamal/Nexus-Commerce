package customers

import (
	"net/http"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	g := rg.Group("/customers")
	g.Use(requireAuth, requireStaff)
	{
		g.GET("", h.ListCustomers)
		g.GET("/archived", h.ListArchivedCustomers)
		g.GET("/:id", h.GetCustomer)
		g.POST("/:id/archive", h.ArchiveCustomer)
		g.POST("/:id/restore", h.RestoreCustomer)
	}
}

func (h *Handler) ListCustomers(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var params ListCustomersParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	customers, total, err := h.svc.ListCustomers(c.Request.Context(), shopID, params)
	if err != nil {
		response.InternalError(c)
		return
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	totalPages := response.CalcTotalPages(total, int64(perPage))
	response.OKWithMeta(c, customers, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetCustomer(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.Param("id")

	detail, err := h.svc.GetCustomer(c.Request.Context(), shopID, customerID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, detail)
}

func (h *Handler) ArchiveCustomer(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.Param("id")

	if err := h.svc.SoftDeleteCustomer(c.Request.Context(), shopID, customerID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) RestoreCustomer(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.Param("id")

	customer, err := h.svc.RestoreCustomer(c.Request.Context(), shopID, customerID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, customer)
}

func (h *Handler) ListArchivedCustomers(c *gin.Context) {
	shopID := c.GetString("shop_id")

	customers, err := h.svc.ListArchivedCustomers(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, customers)
}

func (h *Handler) RegisterCustomerSelfServiceRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	g := rg.Group("/customer")
	g.Use(requireAuth, requireCustomer)
	{
		g.POST("/consent", h.LogConsent)
		g.GET("/data-export", h.DataExport)
		g.DELETE("/account", h.DeleteAccount)
	}
}

func (h *Handler) LogConsent(c *gin.Context) {
	customerID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if customerID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	var body struct {
		ConsentType  string `json:"consent_type" binding:"required"`
		TermsVersion string `json:"terms_version"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.svc.LogConsent(c.Request.Context(), customerID, shopID, body.ConsentType, ip, ua, body.TermsVersion); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"message": "consent recorded"})
}

// Returns a JSON snapshot of all personal data for the requesting customer.
func (h *Handler) DataExport(c *gin.Context) {
	customerID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if customerID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	export, err := h.svc.ExportCustomerData(c.Request.Context(), customerID, shopID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, export)
}

// Anonymises PII and soft-deletes the customer record.
func (h *Handler) DeleteAccount(c *gin.Context) {
	customerID := c.GetString("user_id")
	shopID := c.GetString("shop_id_from_token")
	if customerID == "" || shopID == "" {
		response.Unauthorized(c, "missing customer context")
		return
	}

	if err := h.svc.AnonymizeCustomer(c.Request.Context(), customerID, shopID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "account deleted"})
}
