package inventory

import (
	"strconv"

	"backend-api/internal/shared/ginutil"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	inv := rg.Group("/inventory")
	inv.Use(requireAuth, requireStaff)
	{
		inv.GET("/warehouses", h.GetWarehouses)
		inv.POST("/warehouses", h.CreateWarehouse)
		inv.GET("/levels/:variantId", h.GetInventoryLevels)
		inv.POST("/adjustments", h.AdjustStock)
		inv.GET("/low-stock", h.GetLowStockVariants)
		inv.GET("/movements/:variantId", h.GetStockMovements)
		inv.GET("/bundles/:productId", h.GetBundleComponents)
		inv.POST("/bundles", h.UpsertBundleComponent)
		inv.GET("/reorder-rules", h.GetReorderRules)
		inv.POST("/reorder-rules", h.UpsertReorderRule)
		inv.GET("/bundles/:productId/availability", h.GetBundleAvailability)
	}
}

func (h *Handler) GetWarehouses(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	warehouses, err := h.svc.GetWarehouses(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, warehouses)
}

func (h *Handler) CreateWarehouse(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	wh, err := h.svc.CreateWarehouse(c.Request.Context(), sid, req.Name, req.Address, req.IsDefault)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, wh)
}

func (h *Handler) GetInventoryLevels(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	variantID, err := uuid.Parse(c.Param("variantId"))
	if err != nil {
		response.BadRequest(c, "invalid variantId")
		return
	}
	levels, err := h.svc.GetLevelsByVariant(c.Request.Context(), sid, variantID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, levels)
}

func (h *Handler) AdjustStock(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	variantID, err := uuid.Parse(req.VariantID)
	if err != nil {
		response.BadRequest(c, "invalid variant_id")
		return
	}
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		response.BadRequest(c, "invalid warehouse_id")
		return
	}
	var refID *uuid.UUID
	if req.ReferenceID != nil {
		parsed, err := uuid.Parse(*req.ReferenceID)
		if err != nil {
			response.BadRequest(c, "invalid reference_id")
			return
		}
		refID = &parsed
	}
	createdBy := c.GetString("user_id")
	if err := h.svc.AdjustStock(c.Request.Context(), sid, variantID, warehouseID,
		req.Delta, req.MovementType, refID, req.ReferenceType, req.Note, createdBy); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"adjusted": true})
}

func (h *Handler) GetLowStockVariants(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	rows, err := h.svc.GetLowStockVariants(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) GetStockMovements(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	variantID, err := uuid.Parse(c.Param("variantId"))
	if err != nil {
		response.BadRequest(c, "invalid variantId")
		return
	}
	limit := int32(50)
	offset := int32(0)
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	movements, err := h.svc.GetStockMovements(c.Request.Context(), sid, variantID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, movements)
}

func (h *Handler) GetBundleComponents(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.BadRequest(c, "invalid productId")
		return
	}
	components, err := h.svc.GetProductBundleComponents(c.Request.Context(), sid, productID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, components)
}

func (h *Handler) UpsertBundleComponent(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req UpsertBundleComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	bundleID, err := uuid.Parse(req.BundleProductID)
	if err != nil {
		response.BadRequest(c, "invalid bundle_product_id")
		return
	}
	componentID, err := uuid.Parse(req.ComponentVariantID)
	if err != nil {
		response.BadRequest(c, "invalid component_variant_id")
		return
	}
	comp, err := h.svc.UpsertBundleComponent(c.Request.Context(), sid, bundleID, componentID, req.Quantity)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, comp)
}

func (h *Handler) GetReorderRules(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	rules, err := h.svc.GetReorderRules(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rules)
}

func (h *Handler) UpsertReorderRule(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req UpsertReorderRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	variantID, err := uuid.Parse(req.VariantID)
	if err != nil {
		response.BadRequest(c, "invalid variant_id")
		return
	}
	rule, err := h.svc.UpsertReorderRule(c.Request.Context(), sid, variantID, req.ReorderPoint, req.ReorderQty, req.SupplierEmail)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rule)
}

func (h *Handler) GetBundleAvailability(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.BadRequest(c, "invalid productId")
		return
	}
	available, err := h.svc.ComputeBundleAvailability(c.Request.Context(), sid, productID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"available": available})
}
