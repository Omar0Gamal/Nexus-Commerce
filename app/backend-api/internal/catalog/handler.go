package catalog

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	sharedanalytics "backend-api/internal/shared/analytics"
	sharedaudit "backend-api/internal/shared/audit"
	"backend-api/internal/shared/response"
	"backend-api/internal/storage"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	svc     *Service
	storage *storage.Client // nil → local disk fallback disabled (R2 required)
	rdb     *redis.Client
}

func NewHandler(svc *Service, store *storage.Client, rdb *redis.Client) *Handler {
	return &Handler{svc: svc, storage: store, rdb: rdb}
}

// Read (GET) routes are public; write routes require authenticated staff.
//
//	/api/v1/catalog/products          GET, POST
//	/api/v1/catalog/products/:id      GET, PATCH, DELETE
//	/api/v1/catalog/products/slug/:slug  GET
//	/api/v1/catalog/categories        GET, POST
//	/api/v1/catalog/categories/:id    GET, PATCH, DELETE
//	/api/v1/catalog/categories/slug/:slug  GET
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc, productQuotaMW ...gin.HandlerFunc) {
	catalog := rg.Group("/catalog")

	products := catalog.Group("/products")
	{
		products.GET("", h.ListProducts)
		createHandlers := append([]gin.HandlerFunc{requireAuth, requireStaff}, append(productQuotaMW, h.CreateProduct)...)
		products.POST("", createHandlers...)
		products.GET("/archived", requireAuth, requireStaff, h.ListArchivedProducts)
		products.GET("/:id", h.GetProduct)
		products.PATCH("/:id", requireAuth, requireStaff, h.UpdateProduct)
		products.DELETE("/:id", requireAuth, requireStaff, h.DeleteProduct)
		products.POST("/:id/archive", requireAuth, requireStaff, h.ArchiveProduct)
		products.POST("/:id/restore", requireAuth, requireStaff, h.RestoreProduct)
		products.GET("/slug/:slug", h.GetProductBySlug)

		products.GET("/:id/variants", h.ListVariants)
		products.POST("/:id/variants", requireAuth, requireStaff, h.CreateVariant)
		products.PATCH("/:id/variants/:variantId", requireAuth, requireStaff, h.UpdateVariant)
		products.DELETE("/:id/variants/:variantId", requireAuth, requireStaff, h.DeleteVariant)

		products.GET("/:id/images", h.ListImages)
		products.POST("/:id/images", requireAuth, requireStaff, h.UploadImage)
		products.PATCH("/:id/images/:imageId", requireAuth, requireStaff, h.UpdateImage)
		products.DELETE("/:id/images/:imageId", requireAuth, requireStaff, h.DeleteImage)

		products.GET("/:id/related", h.RelatedProducts)
	}

	categories := catalog.Group("/categories")
	{
		categories.GET("", h.ListCategories)
		categories.POST("", requireAuth, requireStaff, h.CreateCategory)
		categories.GET("/:id", h.GetCategory)
		categories.PATCH("/:id", requireAuth, requireStaff, h.UpdateCategory)
		categories.DELETE("/:id", requireAuth, requireStaff, h.DeleteCategory)
		categories.GET("/slug/:slug", h.GetCategoryBySlug)
	}

	catalog.GET("/search/suggest", h.SearchSuggest)
	catalog.GET("/filters", h.GetFilters)

	catalog.GET("/feed/facebook", h.FacebookFeed)
	catalog.GET("/feed/google", h.GoogleFeed)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := h.svc.CreateProduct(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, product)
}

func (h *Handler) GetProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	product, err := h.svc.GetProduct(c.Request.Context(), shopID, productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, product)
}

func (h *Handler) GetProductBySlug(c *gin.Context) {
	shopID := c.GetString("shop_id")
	slug := c.Param("slug")

	product, err := h.svc.GetProductBySlug(c.Request.Context(), shopID, slug)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, product)
}

// Supports two pagination modes:
//   - Offset mode (default): ?page=1&per_page=20 — returns total_items/total_pages.
//   - Keyset mode: ?cursor_created_at=...&cursor_id=... — returns next_cursor for
//     efficient deep-feed pagination without a COUNT query.
func (h *Handler) ListProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var params ListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, total, err := h.svc.ListProducts(c.Request.Context(), shopID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Publish product_search analytics event when a search query is present (fire-and-forget)
	if h.rdb != nil && params.Search != "" {
		sharedanalytics.Publish(c.Request.Context(), h.rdb, shopID, sharedanalytics.Event{
			Event:  "product_search",
			ShopID: shopID,
			Payload: map[string]any{
				"query":        params.Search,
				"result_count": total,
			},
		})
	}

	// Keyset mode: total == -1; build next_cursor from last item.
	if total == -1 {
		type cursorMeta struct {
			NextCursor *struct {
				CreatedAt string `json:"cursor_created_at"`
				ID        string `json:"cursor_id"`
			} `json:"next_cursor"`
			PerPage int `json:"per_page"`
		}
		meta := cursorMeta{PerPage: params.PerPage}
		if len(products) == params.PerPage {
			last := products[len(products)-1]
			meta.NextCursor = &struct {
				CreatedAt string `json:"cursor_created_at"`
				ID        string `json:"cursor_id"`
			}{
				CreatedAt: last.CreatedAt,
				ID:        last.ID,
			}
		}
		response.OKWithMeta(c, products, meta)
		return
	}

	// Offset mode.
	totalPages := total / int64(params.PerPage)
	if total%int64(params.PerPage) != 0 {
		totalPages++
	}
	response.OKWithMeta(c, products, response.PaginationMeta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := h.svc.UpdateProduct(c.Request.Context(), shopID, productID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, product)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	if err := h.svc.DeleteProduct(c.Request.Context(), shopID, productID); err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.svc.Q(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("user_name"),
		Action:       "product.delete",
		ResourceType: "product",
		ResourceID:   productID,
	})

	response.NoContent(c)
}

func (h *Handler) ArchiveProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	if err := h.svc.SoftDeleteProduct(c.Request.Context(), shopID, productID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) RestoreProduct(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	product, err := h.svc.RestoreProduct(c.Request.Context(), shopID, productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, product)
}

func (h *Handler) ListArchivedProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")

	products, err := h.svc.ListArchivedProducts(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, products)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.svc.CreateCategory(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, category)
}

func (h *Handler) GetCategory(c *gin.Context) {
	shopID := c.GetString("shop_id")
	categoryID := c.Param("id")

	category, err := h.svc.GetCategory(c.Request.Context(), shopID, categoryID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, category)
}

func (h *Handler) GetCategoryBySlug(c *gin.Context) {
	shopID := c.GetString("shop_id")
	slug := c.Param("slug")

	category, err := h.svc.GetCategoryBySlug(c.Request.Context(), shopID, slug)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, category)
}

func (h *Handler) ListCategories(c *gin.Context) {
	shopID := c.GetString("shop_id")

	categories, err := h.svc.ListCategories(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, categories)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	shopID := c.GetString("shop_id")
	categoryID := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.svc.UpdateCategory(c.Request.Context(), shopID, categoryID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	shopID := c.GetString("shop_id")
	categoryID := c.Param("id")

	if err := h.svc.DeleteCategory(c.Request.Context(), shopID, categoryID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) ListVariants(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	variants, err := h.svc.ListVariants(c.Request.Context(), shopID, productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, variants)
}

func (h *Handler) CreateVariant(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	var req CreateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	variant, err := h.svc.CreateVariant(c.Request.Context(), shopID, productID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, variant)
}

func (h *Handler) UpdateVariant(c *gin.Context) {
	shopID := c.GetString("shop_id")
	variantID := c.Param("variantId")

	var req UpdateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	variant, err := h.svc.UpdateVariant(c.Request.Context(), shopID, variantID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, variant)
}

func (h *Handler) DeleteVariant(c *gin.Context) {
	shopID := c.GetString("shop_id")
	variantID := c.Param("variantId")

	if err := h.svc.DeleteVariant(c.Request.Context(), shopID, variantID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) ListImages(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	images, err := h.svc.ListImages(c.Request.Context(), shopID, productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, images)
}

func (h *Handler) UploadImage(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded. Use multipart form field 'file'.")
		return
	}
	defer file.Close()

	// Validate MIME type (read the first 512 bytes, then seek back)
	mime, err := mimetype.DetectReader(file)
	if err != nil || !strings.HasPrefix(mime.String(), "image/") {
		response.BadRequest(c, "Invalid file type. Only image files are allowed.")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Validate file extension as a secondary check
	ext := strings.ToLower(path.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		response.BadRequest(c, "Invalid file type. Allowed: jpg, jpeg, png, gif, webp, svg")
		return
	}

	// Max 10MB
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "File too large. Maximum 10MB.")
		return
	}

	// Validate shopID and productID are valid UUIDs before using them in a storage key.
	if _, err := uuid.Parse(shopID); err != nil {
		response.BadRequest(c, "invalid shop ID")
		return
	}
	if _, err := uuid.Parse(productID); err != nil {
		response.BadRequest(c, "invalid product ID")
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, 10*1024*1024))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	filename := uuid.New().String() + ext
	key := fmt.Sprintf("products/%s/%s/%s", shopID, productID, filename)

	var urlPath string
	if h.storage != nil {
		var uploadErr error
		urlPath, uploadErr = h.storage.Upload(c.Request.Context(), key, mime.String(), bytes.NewReader(data))
		if uploadErr != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	} else {
		// R2 not configured — store raw URL key as placeholder so at least the
		// database record is created. Operators should configure R2 for production.
		urlPath = "/" + key
	}

	altText := c.PostForm("alt_text")
	variantID := c.PostForm("variant_id")
	positionStr := c.PostForm("position")
	isPrimaryStr := c.PostForm("is_primary")

	var position int32
	if positionStr != "" {
		if p, err := strconv.Atoi(positionStr); err == nil {
			position = int32(p)
		}
	}
	isPrimary := isPrimaryStr == "true" || isPrimaryStr == "1"

	image, err := h.svc.CreateImage(c.Request.Context(), shopID, productID, urlPath, altText, variantID, position, isPrimary)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, image)
}

func (h *Handler) UpdateImage(c *gin.Context) {
	shopID := c.GetString("shop_id")
	imageID := c.Param("imageId")

	var req UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	image, err := h.svc.UpdateImage(c.Request.Context(), shopID, imageID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, image)
}

func (h *Handler) DeleteImage(c *gin.Context) {
	shopID := c.GetString("shop_id")
	imageID := c.Param("imageId")

	if err := h.svc.DeleteImage(c.Request.Context(), shopID, imageID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

// Returns products frequently co-purchased with the given product.
// Results are cached in Redis for 1 hour (cache hit: no Postgres query).
func (h *Handler) RelatedProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	limit := 6
	if l, err2 := strconv.Atoi(c.DefaultQuery("limit", "6")); err2 == nil && l > 0 {
		limit = l
	}

	// Serve from cache if available.
	if h.rdb != nil {
		cacheKey := fmt.Sprintf("cache:related:%s:%s", shopID, productID)
		if cached, err2 := h.rdb.Get(c.Request.Context(), cacheKey).Bytes(); err2 == nil {
			c.Data(200, "application/json; charset=utf-8", cached)
			return
		}
	}

	products, err := h.svc.GetRelatedProducts(c.Request.Context(), shopID, productID, limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Populate cache for subsequent requests.
	if h.rdb != nil {
		if data, err2 := json.Marshal(products); err2 == nil {
			cacheKey := fmt.Sprintf("cache:related:%s:%s", shopID, productID)
			h.rdb.Set(c.Request.Context(), cacheKey, data, time.Hour)
		}
	}

	response.OK(c, products)
}

func (h *Handler) SearchSuggest(c *gin.Context) {
	shopID := c.GetString("shop_id")
	q := c.Query("q")
	if q == "" {
		response.OK(c, []any{})
		return
	}
	limitStr := c.DefaultQuery("limit", "5")
	limit := 5
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
		limit = n
	}

	suggestions, err := h.svc.GetSearchSuggestions(c.Request.Context(), shopID, q, limit)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if suggestions == nil {
		suggestions = []SearchSuggestion{}
	}
	response.OK(c, suggestions)
}

func (h *Handler) GetFilters(c *gin.Context) {
	shopID := c.GetString("shop_id")

	facets, err := h.svc.GetSearchFacets(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if facets == nil {
		facets = []Facet{}
	}
	response.OK(c, facets)
}

const feedCacheTTL = 6 * time.Hour

// facebookFeedCacheKey returns the Redis cache key for the Facebook TSV feed.
func facebookFeedCacheKey(shopID string) string {
	return fmt.Sprintf("feed:facebook:%s", shopID)
}

// googleFeedCacheKey returns the Redis cache key for the Google XML feed.
func googleFeedCacheKey(shopID string) string {
	return fmt.Sprintf("feed:google:%s", shopID)
}

// Returns a TSV (tab-separated values) product feed compatible with Meta/Facebook Catalog.
func (h *Handler) FacebookFeed(c *gin.Context) {
	shopID := c.GetString("shop_id")
	cacheKey := facebookFeedCacheKey(shopID)

	// Serve from cache if available
	if h.rdb != nil {
		if cached, err := h.rdb.Get(c.Request.Context(), cacheKey).Bytes(); err == nil {
			c.Data(200, "text/tab-separated-values; charset=utf-8", cached)
			return
		}
	}

	products, err := h.svc.ListProductsForFeed(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var buf bytes.Buffer
	buf.WriteString("id\ttitle\tdescription\tavailability\tcondition\tprice\tlink\timage_link\tbrand\n")
	for _, p := range products {
		if p.Status != "active" {
			continue
		}
		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}
		avail := "in stock"
		if p.StockQuantity <= 0 {
			avail = "out of stock"
		}
		imageURL := ""
		for _, img := range p.Images {
			if img.IsPrimary {
				imageURL = img.URL
				break
			}
		}
		if imageURL == "" && len(p.Images) > 0 {
			imageURL = p.Images[0].URL
		}
		link := fmt.Sprintf("/products/%s", p.Slug)
		price := fmt.Sprintf("%s EGP", p.Price)

		row := strings.Join([]string{
			p.ID, p.Title, desc, avail, "new", price, link, imageURL, "",
		}, "\t")
		buf.WriteString(row + "\n")
	}

	data := buf.Bytes()
	if h.rdb != nil {
		_ = h.rdb.Set(c.Request.Context(), cacheKey, data, feedCacheTTL).Err()
	}
	c.Data(200, "text/tab-separated-values; charset=utf-8", data)
}

type googleFeed struct {
	XMLName xml.Name      `xml:"rss"`
	Version string        `xml:"version,attr"`
	Xmlns   string        `xml:"xmlns:g,attr"`
	Channel googleChannel `xml:"channel"`
}

type googleChannel struct {
	Title string       `xml:"title"`
	Link  string       `xml:"link"`
	Items []googleItem `xml:"item"`
}

type googleItem struct {
	ID           string `xml:"g:id"`
	Title        string `xml:"g:title"`
	Description  string `xml:"g:description"`
	Link         string `xml:"g:link"`
	ImageLink    string `xml:"g:image_link"`
	Availability string `xml:"g:availability"`
	Price        string `xml:"g:price"`
	Condition    string `xml:"g:condition"`
	GTIN         string `xml:"g:gtin,omitempty"`
}

// Returns an XML product feed compatible with Google Merchant Center.
func (h *Handler) GoogleFeed(c *gin.Context) {
	shopID := c.GetString("shop_id")
	cacheKey := googleFeedCacheKey(shopID)

	// Serve from cache if available
	if h.rdb != nil {
		if cached, err := h.rdb.Get(c.Request.Context(), cacheKey).Bytes(); err == nil {
			c.Data(200, "application/xml; charset=utf-8", cached)
			return
		}
	}

	products, err := h.svc.ListProductsForFeed(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]googleItem, 0, len(products))
	for _, p := range products {
		if p.Status != "active" {
			continue
		}
		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}
		avail := "in stock"
		if p.StockQuantity <= 0 {
			avail = "out of stock"
		}
		imageURL := ""
		for _, img := range p.Images {
			if img.IsPrimary {
				imageURL = img.URL
				break
			}
		}
		if imageURL == "" && len(p.Images) > 0 {
			imageURL = p.Images[0].URL
		}
		items = append(items, googleItem{
			ID:           p.ID,
			Title:        p.Title,
			Description:  desc,
			Link:         fmt.Sprintf("/products/%s", p.Slug),
			ImageLink:    imageURL,
			Availability: avail,
			Price:        fmt.Sprintf("%s EGP", p.Price),
			Condition:    "new",
		})
	}

	feed := googleFeed{
		Version: "2.0",
		Xmlns:   "http://base.google.com/ns/1.0",
		Channel: googleChannel{
			Title: shopID,
			Link:  fmt.Sprintf("/shop/%s", shopID),
			Items: items,
		},
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		h.handleError(c, err)
		return
	}
	data := append([]byte(xml.Header), out...)

	if h.rdb != nil {
		_ = h.rdb.Set(c.Request.Context(), cacheKey, data, feedCacheTTL).Err()
	}
	c.Data(200, "application/xml; charset=utf-8", data)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "Resource not found")
	case errors.Is(err, ErrConflict):
		response.Conflict(c, "A resource with that slug already exists")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "Invalid ID format")
	case errors.Is(err, ErrProductLimit):
		response.UnprocessableEntity(c, "You have reached the maximum number of products for your current plan. Please upgrade.")
	default:
		fmt.Println("CATALOG ERROR: ", err)
		c.JSON(500, gin.H{"error": err.Error()})
	}
}



