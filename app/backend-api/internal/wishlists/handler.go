package wishlists

import (
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	wl := rg.Group("/wishlist", requireAuth, requireCustomer)
	{
		wl.GET("", h.GetWishlist)
		wl.POST("", h.Add)
		wl.DELETE("/:productId", h.Remove)
		wl.GET("/check/:productId", h.Check)
		wl.POST("/merge", h.Merge)
	}
}

func (h *Handler) GetWishlist(c *gin.Context) {
	shopID, customerID := c.GetString("shop_id"), c.GetString("user_id")

	items, err := h.svc.Get(c.Request.Context(), shopID, customerID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, items)
}

func (h *Handler) Add(c *gin.Context) {
	shopID, customerID := c.GetString("shop_id"), c.GetString("user_id")

	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Add(c.Request.Context(), shopID, customerID, req.ProductID, req.VariantID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) Remove(c *gin.Context) {
	shopID, customerID := c.GetString("shop_id"), c.GetString("user_id")
	productID := c.Param("productId")

	if err := h.svc.Remove(c.Request.Context(), shopID, customerID, productID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) Check(c *gin.Context) {
	shopID, customerID := c.GetString("shop_id"), c.GetString("user_id")
	productID := c.Param("productId")

	wishlisted, err := h.svc.Check(c.Request.Context(), shopID, customerID, productID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, gin.H{"wishlisted": wishlisted})
}

func (h *Handler) Merge(c *gin.Context) {
	shopID, customerID := c.GetString("shop_id"), c.GetString("user_id")

	var req MergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	for _, pid := range req.ProductIDs {
		_ = h.svc.Add(c.Request.Context(), shopID, customerID, pid, nil)
	}
	response.OK(c, gin.H{"merged": len(req.ProductIDs)})
}
