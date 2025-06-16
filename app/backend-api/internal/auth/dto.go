package auth


// Creates a new user account AND a new shop in one atomic operation.
type RegisterRequest struct {
	// User fields
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"full_name" binding:"required,min=1,max=100"`
	Phone    string `json:"phone,omitempty"`

	// Shop fields
	ShopName  string `json:"shop_name" binding:"required,min=1,max=100"`
	Subdomain string `json:"subdomain" binding:"required,min=3,max=63,alphanum"`
	Currency  string `json:"currency,omitempty"`  // default: EGP
	Timezone  string `json:"timezone,omitempty"`  // default: Africa/Cairo
	PlanName  string `json:"plan_name,omitempty"` // default: standard
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	// ShopID is optional — if provided, the token will include shop context.
	// If omitted, the user gets a "global" token and must select a shop later.
	ShopID string `json:"shop_id,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	// ShopID is optional — allows switching shop context on refresh.
	ShopID string `json:"shop_id,omitempty"`
}

type SelectShopRequest struct {
	ShopID string `json:"shop_id" binding:"required"`
}


// Creates a customer account scoped to the current shop (from X-Shop-ID).
type CustomerRegisterRequest struct {
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8,max=128"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type CustomerLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}


type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}


// AuthResponse is returned after successful login or register.
type AuthResponse struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int               `json:"expires_in"` // seconds
	ActorType    string            `json:"actor_type"` // "staff", "customer", "platform_admin"
	User         UserResponse      `json:"user"`
	Shop         *ShopBrief        `json:"shop,omitempty"`     // nil if no shop context selected (staff)
	Customer     *CustomerResponse `json:"customer,omitempty"` // set for customer auth
}

// UserResponse is a clean user representation (no password hash). Used for staff.
type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FullName        string `json:"full_name"`
	Phone           string `json:"phone,omitempty"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

// CustomerResponse is a clean customer representation. Used for customer auth.
type CustomerResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// AdminResponse is a clean platform admin representation.
type AdminResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// ShopBrief is a minimal shop representation included in auth responses.
type ShopBrief struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
	Role      string `json:"role"`
	IsOwner   bool   `json:"is_owner"`
}

// MeResponse is returned by GET /auth/me (staff).
type MeResponse struct {
	User  UserResponse `json:"user"`
	Shops []ShopBrief  `json:"shops"`
}

// CustomerMeResponse is returned by GET /auth/customer/me.
type CustomerMeResponse struct {
	Customer CustomerResponse `json:"customer"`
}

// AdminMeResponse is returned by GET /auth/admin/me.
type AdminMeResponse struct {
	Admin AdminResponse `json:"admin"`
}


type CustomerUpdateMeRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type CustomerChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}


type UpdateShopSettingsRequest struct {
	Name              string          `json:"name,omitempty"`
	Currency          string          `json:"currency,omitempty"`
	NotificationPrefs map[string]bool `json:"notification_prefs,omitempty"`
	// SEO fields (professional/premium only)
	SeoTitle       string `json:"seo_title,omitempty"`
	SeoDescription string `json:"seo_description,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
	// Custom domain (professional/premium only)
	CustomDomain string `json:"custom_domain,omitempty"`
}

// ShopSettingsResponse is returned after a successful shop settings update.
type ShopSettingsResponse struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Subdomain         string          `json:"subdomain"`
	CustomDomain      string          `json:"custom_domain,omitempty"`
	Currency          string          `json:"currency"`
	PlanName          string          `json:"plan_name,omitempty"`
	NotificationPrefs map[string]bool `json:"notification_prefs,omitempty"`
	// SEO fields
	SeoTitle       string `json:"seo_title,omitempty"`
	SeoDescription string `json:"seo_description,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
}


// TwoFAChallengeResponse is returned when login succeeds but 2FA is required.
type TwoFAChallengeResponse struct {
	Requires2FA bool   `json:"requires_2fa"`
	LoginToken  string `json:"login_token"` // short-lived JWT, exchange with Challenge2FA
}

// Setup2FAResponse is returned by POST /auth/2fa/setup.
type Setup2FAResponse struct {
	ProvisioningURI string   `json:"provisioning_uri"`
	QRCodeBase64    string   `json:"qr_code_base64"` // PNG data URI (one-time)
	BackupCodes     []string `json:"backup_codes"`   // plaintext (show once, then gone)
}

type Verify2FARequest struct {
	Code string `json:"code" binding:"required,min=6,max=8"`
}

type Challenge2FARequest struct {
	LoginToken string `json:"login_token" binding:"required"`
	Code       string `json:"code" binding:"required,min=6,max=8"`
}

type RecoverAccountRequest struct {
	LoginToken string `json:"login_token" binding:"required"`
	BackupCode string `json:"backup_code" binding:"required"`
}


// StaffMemberResponse is a staff member with user info.
type StaffMemberResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
	IsOwner  bool   `json:"is_owner"`
	JoinedAt string `json:"joined_at"`
}

type AddStaffRequest struct {
	Email string `json:"email" binding:"required,email"`
}


type StaffUpdateProfileRequest struct {
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type StaffChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}
