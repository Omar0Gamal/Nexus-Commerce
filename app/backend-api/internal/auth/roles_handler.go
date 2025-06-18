package auth

import (
	"errors"
	"net/http"

	sharedaudit "backend-api/internal/shared/audit"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// These are tenant-scoped (X-Shop-ID required) and staff-only.
//
//	GET    /api/v1/roles             — list roles
//	POST   /api/v1/roles             — create custom role
//	PATCH  /api/v1/roles/:id         — update role permissions
//	DELETE /api/v1/roles/:id         — delete custom role
//	PATCH  /api/v1/staff/:id/role    — assign role to staff member
func (h *Handler) RegisterRoleRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	roles := rg.Group("/roles")
	roles.Use(requireAuth, requireStaff)
	{
		roles.GET("", h.ListRoles)
		roles.POST("", h.CreateRole)
		roles.PATCH("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
	}

	staff := rg.Group("/staff")
	staff.Use(requireAuth, requireStaff)
	{
		staff.PATCH("/:id/role", h.AssignStaffRole)
	}
}

func (h *Handler) ListRoles(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	roles, err := h.staffSvc.ListRoles(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, roles)
}

func (h *Handler) CreateRole(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.staffSvc.CreateRole(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	shopID := c.GetString("shop_id")
	roleID := c.Param("id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.staffSvc.UpdateRole(c.Request.Context(), shopID, roleID, req)
	if err != nil {
		h.handleRoleError(c, err)
		return
	}
	response.OK(c, role)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	shopID := c.GetString("shop_id")
	roleID := c.Param("id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	if err := h.staffSvc.DeleteRole(c.Request.Context(), shopID, roleID); err != nil {
		h.handleRoleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) AssignStaffRole(c *gin.Context) {
	shopID := c.GetString("shop_id")
	staffID := c.Param("id")
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.staffSvc.AssignStaffRole(c.Request.Context(), shopID, staffID, req.RoleID); err != nil {
		h.handleRoleError(c, err)
		return
	}

	sharedaudit.Write(h.staffSvc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("user_name"),
		Action:       "staff.role_assign",
		ResourceType: "shop_staff",
		ResourceID:   staffID,
		Changes:      map[string]any{"role_id": req.RoleID},
	})

	response.OK(c, gin.H{"message": "role assigned"})
}

func (h *Handler) handleRoleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRoleNotFound), errors.Is(err, ErrStaffNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, ErrSystemRole):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrRoleInUse):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "invalid id format")
	default:
		response.InternalError(c)
	}
}
