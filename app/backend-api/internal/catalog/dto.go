package catalog

type CreateProductRequest struct {
	Title             string  `json:"title" binding:"required,min=1,max=255"`
	Slug              string  `json:"slug" binding:"required,min=1,max=255"`
	Description       string  `json:"description,omitempty"`
	Price             float64 `json:"price" binding:"required,gt=0"`
	CompareAtPrice    float64 `json:"compare_at_price,omitempty"`
	CategoryID        string  `json:"category_id,omitempty"`
	SKU               string  `json:"sku,omitempty"`
	TrackInventory    bool    `json:"track_inventory,omitempty"`
	StockQuantity     int32   `json:"stock_quantity,omitempty"`
	LowStockThreshold int32   `json:"low_stock_threshold,omitempty"`
	Status            string  `json:"status,omitempty"`
	SeoTitle          string  `json:"seo_title,omitempty"`
	SeoDescription    string  `json:"seo_description,omitempty"`
}

type UpdateProductRequest struct {
	Title             *string  `json:"title,omitempty"`
	Slug              *string  `json:"slug,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Price             *float64 `json:"price,omitempty"`
	CompareAtPrice    *float64 `json:"compare_at_price,omitempty"`
	CategoryID        *string  `json:"category_id,omitempty"`
	SKU               *string  `json:"sku,omitempty"`
	TrackInventory    *bool    `json:"track_inventory,omitempty"`
	StockQuantity     *int32   `json:"stock_quantity,omitempty"`
	LowStockThreshold *int32   `json:"low_stock_threshold,omitempty"`
	Status            *string  `json:"status,omitempty"`
	SeoTitle          *string  `json:"seo_title,omitempty"`
	SeoDescription    *string  `json:"seo_description,omitempty"`
}

type CreateCategoryRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=255"`
	Slug            string `json:"slug" binding:"required,min=1,max=255"`
	ParentID        string `json:"parent_id,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name            *string `json:"name,omitempty"`
	Slug            *string `json:"slug,omitempty"`
	ParentID        *string `json:"parent_id,omitempty"`
	MetaDescription *string `json:"meta_description,omitempty"`
}

type ProductResponse struct {
	ID                string            `json:"id"`
	ShopID            string            `json:"shop_id"`
	CategoryID        *string           `json:"category_id,omitempty"`
	Title             string            `json:"title"`
	Slug              string            `json:"slug"`
	Description       *string           `json:"description,omitempty"`
	Price             string            `json:"price"`
	CompareAtPrice    *string           `json:"compare_at_price,omitempty"`
	TrackInventory    bool              `json:"track_inventory"`
	StockQuantity     int32             `json:"stock_quantity"`
	LowStockThreshold int32             `json:"low_stock_threshold"`
	LowStock          bool              `json:"low_stock"`
	SKU               *string           `json:"sku,omitempty"`
	Status            string            `json:"status"`
	Variants          []VariantResponse `json:"variants,omitempty"`
	Images            []ImageResponse   `json:"images,omitempty"`
	SeoTitle          string            `json:"seo_title,omitempty"`
	SeoDescription    string            `json:"seo_description,omitempty"`
	CreatedAt         string            `json:"created_at,omitempty"`
}

type CategoryResponse struct {
	ID              string  `json:"id"`
	ShopID          string  `json:"shop_id"`
	ParentID        *string `json:"parent_id,omitempty"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	MetaDescription string  `json:"meta_description,omitempty"`
}

type ListQueryParams struct {
	Page            int      `form:"page,default=1"`
	PerPage         int      `form:"per_page,default=20"`
	Status          string   `form:"status"`
	CategoryID      string   `form:"category_id"`
	Search          string   `form:"search"`
	MinPrice        *float64 `form:"min_price"`
	MaxPrice        *float64 `form:"max_price"`
	SortBy          string   `form:"sort_by"`
	// Keyset pagination ordering: (created_at DESC, id DESC).
	CursorCreatedAt string   `form:"cursor_created_at"`
	CursorID        string   `form:"cursor_id"`
}

type CreateVariantRequest struct {
	Title          string                 `json:"title" binding:"required,min=1,max=255"`
	SKU            string                 `json:"sku,omitempty"`
	Price          float64                `json:"price" binding:"required,gt=0"`
	CompareAtPrice float64                `json:"compare_at_price,omitempty"`
	StockQuantity  int32                  `json:"stock_quantity"`
	Options        map[string]interface{} `json:"options,omitempty"`
	Position       int32                  `json:"position,omitempty"`
}

type UpdateVariantRequest struct {
	Title          *string                `json:"title,omitempty"`
	SKU            *string                `json:"sku,omitempty"`
	Price          *float64               `json:"price,omitempty"`
	CompareAtPrice *float64               `json:"compare_at_price,omitempty"`
	StockQuantity  *int32                 `json:"stock_quantity,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
	Position       *int32                 `json:"position,omitempty"`
}

type VariantResponse struct {
	ID             string                 `json:"id"`
	ProductID      string                 `json:"product_id"`
	ShopID         string                 `json:"shop_id"`
	Title          string                 `json:"title"`
	SKU            *string                `json:"sku,omitempty"`
	Price          string                 `json:"price"`
	CompareAtPrice *string                `json:"compare_at_price,omitempty"`
	StockQuantity  int32                  `json:"stock_quantity"`
	Options        map[string]interface{} `json:"options"`
	Position       int32                  `json:"position"`
	CreatedAt      string                 `json:"created_at"`
}

type UpdateImageRequest struct {
	AltText   *string `json:"alt_text,omitempty"`
	Position  *int32  `json:"position,omitempty"`
	IsPrimary *bool   `json:"is_primary,omitempty"`
	VariantID *string `json:"variant_id,omitempty"`
}

type ImageResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	ShopID    string  `json:"shop_id"`
	VariantID *string `json:"variant_id,omitempty"`
	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	Position  int32   `json:"position"`
	IsPrimary bool    `json:"is_primary"`
	CreatedAt string  `json:"created_at"`
}
