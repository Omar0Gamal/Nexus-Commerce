package inventory

type CreateWarehouseRequest struct {
	Name      string `json:"name"      binding:"required"`
	Address   string `json:"address"`
	IsDefault bool   `json:"is_default"`
}

type AdjustStockRequest struct {
	VariantID     string  `json:"variant_id"     binding:"required"`
	WarehouseID   string  `json:"warehouse_id"   binding:"required"`
	Delta         int     `json:"delta"          binding:"required"`
	MovementType  string  `json:"movement_type"  binding:"required"`
	ReferenceID   *string `json:"reference_id"`
	ReferenceType string  `json:"reference_type"`
	Note          string  `json:"note"`
}

type UpsertBundleComponentRequest struct {
	BundleProductID    string `json:"bundle_product_id"    binding:"required"`
	ComponentVariantID string `json:"component_variant_id" binding:"required"`
	Quantity           int16  `json:"quantity"             binding:"required"`
}

type UpsertReorderRuleRequest struct {
	VariantID     string `json:"variant_id"     binding:"required"`
	ReorderPoint  int32  `json:"reorder_point"  binding:"required"`
	ReorderQty    int32  `json:"reorder_qty"    binding:"required"`
	SupplierEmail string `json:"supplier_email"`
}
