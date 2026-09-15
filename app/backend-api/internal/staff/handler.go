package staff

import (
	"errors"

	sharedaudit "backend-api/internal/shared/audit"
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
	staff := rg.Group("/staff")
	staff.Use(requireAuth, requireStaff)
	{
		staff.GET("", h.ListStaff)
		staff.POST("/invite", h.InviteStaff)
		staff.PATCH("/:id/role", h.UpdateStaffRole)
		staff.DELETE("/:id", h.RemoveStaff)
	}
}

func (h *Handler) getShopID(c *gin.Context) string {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		shopID = c.GetString("shop_id_from_token")
	}
	return shopID
}

func (h *Handler) ListStaff(c *gin.Context) {
	shopID := h.getShopID(c)
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}

	members, err := h.svc.ListStaff(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, members)
}

func (h *Handler) InviteStaff(c *gin.Context) {
	shopID := h.getShopID(c)
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}

	var req InviteStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	actorUserID := c.GetString("user_id")
	actorName := c.GetString("full_name")

	member, err := h.svc.InviteStaff(c.Request.Context(), shopID, actorUserID, actorName, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.svc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  actorUserID,
		ActorName:    actorName,
		Action:       "staff.invited",
		ResourceType: "staff",
		ResourceID:   member.ID,
		Changes: map[string]any{
			"email":   req.Email,
			"role_id": req.RoleID,
		},
	})

	response.Created(c, member)
}

func (h *Handler) UpdateStaffRole(c *gin.Context) {
	shopID := h.getShopID(c)
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}

	staffID := c.Param("id")
	if staffID == "" {
		response.BadRequest(c, "staff ID required")
		return
	}

	var req UpdateStaffRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.svc.UpdateStaffRole(c.Request.Context(), shopID, staffID, req.RoleID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.svc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("full_name"),
		Action:       "staff.role_assigned",
		ResourceType: "staff",
		ResourceID:   staffID,
		Changes: map[string]any{
			"role_id": req.RoleID,
		},
	})

	response.OK(c, updated)
}

func (h *Handler) RemoveStaff(c *gin.Context) {
	shopID := h.getShopID(c)
	if shopID == "" {
		response.BadRequest(c, "shop context required")
		return
	}

	staffID := c.Param("id")
	if staffID == "" {
		response.BadRequest(c, "staff ID required")
		return
	}

	if err := h.svc.RemoveStaff(c.Request.Context(), shopID, staffID); err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.svc.Queries(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("full_name"),
		Action:       "staff.removed",
		ResourceType: "staff",
		ResourceID:   staffID,
	})

	response.NoContent(c)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStaffNotFound), errors.Is(err, ErrRoleNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, ErrAlreadyStaff):
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrCannotChangeOwnerRole),
		errors.Is(err, ErrCannotRemoveOwner),
		errors.Is(err, ErrInvalidUUID),
		errors.Is(err, ErrInvalidEmail):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrStaffLimit):
		response.Forbidden(c, err.Error())
	default:
		response.InternalError(c)
	}
}
