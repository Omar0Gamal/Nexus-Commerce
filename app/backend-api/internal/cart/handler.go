package cart

import (
	"context"
	"errors"
	"time"

	"backend-api/internal/orders"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	svc       *Service
	ordersSvc OrderPlacer
	logger    *zap.Logger
}

func NewHandler(svc *Service, ordersSvc OrderPlacer, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, ordersSvc: ordersSvc, logger: logger}
}

// optionalAuth parses JWT when present; unauthenticated requests fall back to X-Cart-Token.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, optionalAuth gin.HandlerFunc) {
	cart := rg.Group("/cart")
	cart.Use(optionalAuth)
	{
		cart.GET("", h.GetCart)
		cart.GET("/recover", h.RecoverCart)
		cart.DELETE("", h.ClearCart)
		cart.POST("/items", h.AddItem)
		cart.PATCH("/items", h.UpdateItem)
		cart.DELETE("/items", h.RemoveItem)
		cart.POST("/checkout", h.Checkout)
	}
}

func identityFromContext(c *gin.Context) (string, bool) {
	if actorType := c.GetString("actor_type"); actorType == "customer" {
		if id := c.GetString("user_id"); id != "" {
			return "customer:" + id, true
		}
	}

	token := c.GetHeader("X-Cart-Token")
	if token == "" {
		token = uuid.New().String()
		c.Header("X-Cart-Token", token)
	}
	return "guest:" + token, false
}

func (h *Handler) GetCart(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, _ := identityFromContext(c)

	cart, err := h.svc.GetCart(c.Request.Context(), shopID, identity)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, cart)
}

func (h *Handler) AddItem(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, isCustomer := identityFromContext(c)

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cart, err := h.svc.AddItem(c.Request.Context(), shopID, identity, isCustomer, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if isCustomer {
		if customerID := c.GetString("user_id"); customerID != "" {
			h.asyncSaveCart(shopID, customerID)
		}
	}

	response.OK(c, cart)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, isCustomer := identityFromContext(c)

	productID := c.Query("product_id")
	variantID := c.Query("variant_id")

	if productID == "" {
		response.BadRequest(c, "product_id query parameter is required")
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cart, err := h.svc.UpdateItem(c.Request.Context(), shopID, identity, isCustomer, productID, variantID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if isCustomer {
		if customerID := c.GetString("user_id"); customerID != "" {
			h.asyncSaveCart(shopID, customerID)
		}
	}

	response.OK(c, cart)
}

func (h *Handler) RemoveItem(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, isCustomer := identityFromContext(c)

	productID := c.Query("product_id")
	variantID := c.Query("variant_id")

	if productID == "" {
		response.BadRequest(c, "product_id query parameter is required")
		return
	}

	cart, err := h.svc.RemoveItem(c.Request.Context(), shopID, identity, isCustomer, productID, variantID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if isCustomer {
		if customerID := c.GetString("user_id"); customerID != "" {
			h.asyncSaveCart(shopID, customerID)
		}
	}

	response.OK(c, cart)
}

func (h *Handler) ClearCart(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, _ := identityFromContext(c)

	if err := h.svc.ClearCart(c.Request.Context(), shopID, identity); err != nil {
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

func (h *Handler) Checkout(c *gin.Context) {
	shopID := c.GetString("shop_id")
	identity, _ := identityFromContext(c)

	customerID := ""
	if c.GetString("actor_type") == "customer" {
		customerID = c.GetString("user_id")
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.svc.Checkout(c.Request.Context(), shopID, identity, customerID, req, h.ordersSvc)
	if err != nil {
		h.logger.Error("checkout error", zap.Error(err))
		h.handleError(c, err)
		return
	}

	if customerID != "" {
		go func() {
			deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.svc.DeleteSavedCart(deleteCtx, shopID, customerID); err != nil {
				h.logger.Warn("failed to delete saved cart after checkout",
					zap.String("shop_id", shopID),
					zap.String("customer_id", customerID),
					zap.Error(err),
				)
			}
		}()
	}

	response.Created(c, order)
}

func (h *Handler) RecoverCart(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}
	cart, _, err := h.svc.RecoverCart(c.Request.Context(), token)
	if err != nil {
		response.NotFound(c, "recovery token not found or already used")
		return
	}
	response.OK(c, cart)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProductNotFound):
		response.NotFound(c, "Product not found")
	case errors.Is(err, ErrVariantNotFound):
		response.NotFound(c, "Variant not found")
	case errors.Is(err, ErrEmptyCart):
		response.BadRequest(c, "Cart is empty")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "Invalid ID format")
	case errors.Is(err, ErrOutOfStock):
		response.BadRequest(c, err.Error())
	case errors.Is(err, orders.ErrInsufficientStock):
		response.BadRequest(c, "One or more items are out of stock")
	default:
		h.logger.Error("unhandled cart error", zap.Error(err))
		response.InternalError(c)
	}
}

func (h *Handler) asyncSaveCart(shopID, customerID string) {
	go func() {
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.svc.SaveCart(saveCtx, shopID, customerID); err != nil {
			h.logger.Warn("failed to save cart asynchronously",
				zap.String("shop_id", shopID),
				zap.String("customer_id", customerID),
				zap.Error(err),
			)
		}
	}()
}
