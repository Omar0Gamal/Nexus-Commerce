package addresses

import (
	"errors"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "address not found")
	default:
		response.InternalError(c)
	}
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	addr := rg.Group("/addresses")
	addr.Use(requireAuth, requireCustomer)
	{
		addr.GET("", h.ListAddresses)
		addr.POST("", h.CreateAddress)
		addr.PUT("/:id", h.UpdateAddress)
		addr.DELETE("/:id", h.DeleteAddress)
		addr.PATCH("/:id/default", h.SetDefault)
	}
}

func (h *Handler) ListAddresses(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")

	addrs, err := h.svc.ListAddresses(c.Request.Context(), shopID, customerID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, addrs)
}

func (h *Handler) CreateAddress(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")

	var req CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	addr, err := h.svc.CreateAddress(c.Request.Context(), shopID, customerID, req)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, addr)
}

func (h *Handler) UpdateAddress(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	addressID := c.Param("id")

	var req UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	addr, err := h.svc.UpdateAddress(c.Request.Context(), shopID, customerID, addressID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, addr)
}

func (h *Handler) DeleteAddress(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	addressID := c.Param("id")

	if err := h.svc.DeleteAddress(c.Request.Context(), shopID, customerID, addressID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) SetDefault(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	addressID := c.Param("id")

	if err := h.svc.SetDefault(c.Request.Context(), shopID, customerID, addressID); err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "address set as default"})
}
