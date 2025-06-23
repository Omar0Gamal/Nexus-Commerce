// Package catalog implements the Catalog Module for multi-tenant product management.
// It provides CRUD operations for Products and Categories, strictly scoped by shop_id.
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
	// SEO fields
	SeoTitle       string `json:"seo_title,omitempty"`
	SeoDescription string `json:"seo_description,omitempty"`
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
	// SEO fields
	SeoTitle       *string `json:"seo_title,omitempty"`
	SeoDescription *string `json:"seo_description,omitempty"`
}

type CreateCategoryRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=255"`
	Slug            string `json:"slug" binding:"required,min=1,max=255"`
	ParentID        string `json:"parent_id,omitempty"` // UUID string
	MetaDescription string `json:"meta_description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name            *string `json:"name,omitempty"`
	Slug            *string `json:"slug,omitempty"`
	ParentID        *string `json:"parent_id,omitempty"`
	MetaDescription *string `json:"meta_description,omitempty"`
}

// Clean JSON responses — no pgtype.UUID/pgtype.Text leaking to the API consumer.

// ProductResponse is the JSON response for a single product.
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
	// SEO fields
	SeoTitle       string `json:"seo_title,omitempty"`
	SeoDescription string `json:"seo_description,omitempty"`
	// Timestamps
	CreatedAt string `json:"created_at,omitempty"`
}

// CategoryResponse is the JSON response for a single category.
type CategoryResponse struct {
	ID              string  `json:"id"`
	ShopID          string  `json:"shop_id"`
	ParentID        *string `json:"parent_id,omitempty"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	MetaDescription string  `json:"meta_description,omitempty"`
}

// ListQueryParams holds common pagination/filter/search parameters from query string.
type ListQueryParams struct {
	Page       int      `form:"page,default=1"`
	PerPage    int      `form:"per_page,default=20"`
	Status     string   `form:"status"`
	CategoryID string   `form:"category_id"`
	Search     string   `form:"search"`
	MinPrice   *float64 `form:"min_price"`
	MaxPrice   *float64 `form:"max_price"`
	SortBy     string   `form:"sort_by"` // price_asc | price_desc | (default: title asc)
	// Keyset pagination cursor (uses created_at DESC, id DESC ordering).
	// When provided, the response uses cursor-based next-page navigation
	// instead of offset/total-pages. Pass values from the previous
	// response's next_cursor field.
	CursorCreatedAt string `form:"cursor_created_at"` // RFC3339
	CursorID        string `form:"cursor_id"`         // UUID
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

// VariantResponse is the JSON response for a single variant.
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

// ImageResponse is the JSON response for a single image.
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
