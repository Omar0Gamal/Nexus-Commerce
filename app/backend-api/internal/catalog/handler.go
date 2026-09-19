package catalog

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc, productQuotaMW ...gin.HandlerFunc) {
	catalog := rg.Group("/catalog")

	products := catalog.Group("/products")
	{
		products.GET("", h.ListProducts)
		createHandlers := append([]gin.HandlerFunc{requireAuth, requireStaff}, append(productQuotaMW, h.CreateProduct)...)
		products.POST("", createHandlers...)
		products.POST("/import", requireAuth, requireStaff, h.ImportProducts)
		products.POST("/import/shopify", requireAuth, requireStaff, h.ImportShopify)
		products.GET("/export", requireAuth, requireStaff, h.ExportProducts)
		products.GET("/archived", requireAuth, requireStaff, h.ListArchivedProducts)
		products.GET("/:id", h.GetProduct)
		products.PATCH("/:id", requireAuth, requireStaff, h.UpdateProduct)
		products.DELETE("/:id", requireAuth, requireStaff, h.DeleteProduct)
		products.POST("/:id/archive", requireAuth, requireStaff, h.ArchiveProduct)
		products.POST("/:id/restore", requireAuth, requireStaff, h.RestoreProduct)
		products.GET("/slug/:slug", h.GetProductBySlug)

		products.POST("/bulk-archive", requireAuth, requireStaff, h.BulkArchiveProducts)
		products.POST("/bulk-delete", requireAuth, requireStaff, h.BulkDeleteProducts)

		products.GET("/:id/variants", h.ListVariants)
		products.POST("/:id/variants", requireAuth, requireStaff, h.CreateVariant)
		products.PATCH("/:id/variants/:variantId", requireAuth, requireStaff, h.UpdateVariant)
		products.DELETE("/:id/variants/:variantId", requireAuth, requireStaff, h.DeleteVariant)

		products.GET("/:id/images", h.ListImages)
		products.POST("/:id/images/presign", requireAuth, requireStaff, h.PresignImageUpload)
		products.POST("/:id/images/link", requireAuth, requireStaff, h.LinkImage)
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

	// Keyset pagination: total == -1 indicates cursor-based paging.
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

type PresignImageUploadRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

func (h *Handler) PresignImageUpload(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	var req PresignImageUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !strings.HasPrefix(req.ContentType, "image/") {
		response.BadRequest(c, "Invalid content type. Must be an image.")
		return
	}

	if h.storage == nil {
		response.BadRequest(c, "Storage is not configured.")
		return
	}

	ext := strings.ToLower(path.Ext(req.Filename))
	if ext == "" {
		if req.ContentType == "image/jpeg" {
			ext = ".jpg"
		} else if req.ContentType == "image/png" {
			ext = ".png"
		} else if req.ContentType == "image/webp" {
			ext = ".webp"
		} else if req.ContentType == "image/gif" {
			ext = ".gif"
		} else {
			response.BadRequest(c, "Filename must have an extension.")
			return
		}
	}
	
	filename := uuid.New().String() + ext
	key := fmt.Sprintf("products/%s/%s/%s", shopID, productID, filename)

	uploadURL, publicURL, err := h.storage.PresignUpload(c.Request.Context(), key, req.ContentType, 300)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate presigned URL: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"upload_url": uploadURL,
		"public_url": publicURL,
		"key":        key,
	})
}

type LinkImageRequest struct {
	PublicURL  string `json:"public_url" binding:"required"`
	AltText    string `json:"alt_text"`
	VariantID  string `json:"variant_id"`
	Position   int32  `json:"position"`
	IsPrimary  bool   `json:"is_primary"`
}

func (h *Handler) LinkImage(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	var req LinkImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	image, err := h.svc.CreateImage(c.Request.Context(), shopID, productID, req.PublicURL, req.AltText, req.VariantID, req.Position, req.IsPrimary)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, image)
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

	mime, err := mimetype.DetectReader(file)
	if err != nil || !strings.HasPrefix(mime.String(), "image/") {
		response.BadRequest(c, "Invalid file type. Only image files are allowed.")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ext := strings.ToLower(path.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		response.BadRequest(c, "Invalid file type. Allowed: jpg, jpeg, png, gif, webp, svg")
		return
	}

	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "File too large. Maximum 10MB.")
		return
	}

	// Security: validate UUID format to prevent storage key path injection.
	if _, err := uuid.Parse(shopID); err != nil {
		response.BadRequest(c, "invalid shop ID")
		return
	}
	if _, err := uuid.Parse(productID); err != nil {
		response.BadRequest(c, "invalid product ID")
		return
	}

	fileBytes, err := io.ReadAll(io.LimitReader(file, 10*1024*1024))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	filename := uuid.New().String() + ext
	key := fmt.Sprintf("products/%s/%s/%s", shopID, productID, filename)

	var urlPath string
	if h.storage != nil {
		var uploadErr error
		urlPath, uploadErr = h.storage.Upload(c.Request.Context(), key, mime.String(), bytes.NewReader(fileBytes))
		if uploadErr != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Fallback placeholder when R2 object storage is not configured.
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

	if h.rdb != nil {
		if productsJSON, err2 := json.Marshal(products); err2 == nil {
			cacheKey := fmt.Sprintf("cache:related:%s:%s", shopID, productID)
			h.rdb.Set(c.Request.Context(), cacheKey, productsJSON, time.Hour)
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

func googleFeedCacheKey(shopID string) string {
	return fmt.Sprintf("feed:google:%s", shopID)
}

func (h *Handler) FacebookFeed(c *gin.Context) {
	shopID := c.GetString("shop_id")
	cacheKey := facebookFeedCacheKey(shopID)

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

	tsvBytes := buf.Bytes()
	if h.rdb != nil {
		_ = h.rdb.Set(c.Request.Context(), cacheKey, tsvBytes, feedCacheTTL).Err()
	}
	c.Data(200, "text/tab-separated-values; charset=utf-8", tsvBytes)
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

func (h *Handler) GoogleFeed(c *gin.Context) {
	shopID := c.GetString("shop_id")
	cacheKey := googleFeedCacheKey(shopID)

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
	xmlBytes := append([]byte(xml.Header), out...)

	if h.rdb != nil {
		_ = h.rdb.Set(c.Request.Context(), cacheKey, xmlBytes, feedCacheTTL).Err()
	}
	c.Data(200, "application/xml; charset=utf-8", xmlBytes)
}

type BulkActionRequest struct {
	ProductIDs []string `json:"product_ids" binding:"required"`
}

func (h *Handler) BulkArchiveProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	for _, pid := range req.ProductIDs {
		_ = h.svc.SoftDeleteProduct(c.Request.Context(), shopID, pid)
	}

	response.OK(c, gin.H{"message": "Products archived successfully"})
}

func (h *Handler) BulkDeleteProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	for _, pid := range req.ProductIDs {
		_ = h.svc.DeleteProduct(c.Request.Context(), shopID, pid)
	}

	response.OK(c, gin.H{"message": "Products deleted successfully"})
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



func (h *Handler) ExportProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")

	products, _, err := h.svc.ListProducts(c.Request.Context(), shopID, ListQueryParams{
		Page:    1,
		PerPage: 10000,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=products.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID", "Title", "Slug", "Price", "StockQuantity", "Status"})

	for _, p := range products {
		writer.Write([]string{
			p.ID,
			p.Title,
			p.Slug,
			p.Price,
			fmt.Sprintf("%d", p.StockQuantity),
			p.Status,
		})
	}
	writer.Flush()
}

func (h *Handler) ImportProducts(c *gin.Context) {
	shopID := c.GetString("shop_id")
	
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "Failed to get file from request")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		response.BadRequest(c, "Failed to parse CSV")
		return
	}

	if len(records) < 2 {
		response.BadRequest(c, "CSV must contain at least a header and one row")
		return
	}

	header := records[0]
	titleIdx, slugIdx, priceIdx, stockIdx, statusIdx := -1, -1, -1, -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "title": titleIdx = i
		case "slug": slugIdx = i
		case "price": priceIdx = i
		case "stockquantity": stockIdx = i
		case "status": statusIdx = i
		}
	}

	if titleIdx == -1 || slugIdx == -1 || priceIdx == -1 {
		response.BadRequest(c, "CSV missing required headers: Title, Slug, Price")
		return
	}

	successCount := 0
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= priceIdx { continue }
		
		price, _ := strconv.ParseFloat(row[priceIdx], 64)
		var stock int32
		if stockIdx != -1 && len(row) > stockIdx {
			s, _ := strconv.ParseInt(row[stockIdx], 10, 32)
			stock = int32(s)
		}
		status := "draft"
		if statusIdx != -1 && len(row) > statusIdx && row[statusIdx] != "" {
			status = row[statusIdx]
		}
		
		req := CreateProductRequest{
			Title: row[titleIdx],
			Slug: row[slugIdx],
			Price: price,
			TrackInventory: true,
			StockQuantity: stock,
			Status: status,
		}
		
		_, err := h.svc.CreateProduct(c.Request.Context(), shopID, req)
		if err == nil {
			successCount++
		}
	}

	response.OK(c, gin.H{"message": fmt.Sprintf("Imported %d products successfully", successCount)})
}

type ImportShopifyRequest struct {
	StoreURL    string `json:"store_url"`
	ShopURL     string `json:"shop_url,omitempty"`
	AccessToken string `json:"access_token"`
}

type shopifyProductImage struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Position  int32   `json:"position"`
	Src       string  `json:"src"`
	Alt       *string `json:"alt"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
}

type shopifyProductVariant struct {
	ID                  int64   `json:"id"`
	ProductID           int64   `json:"product_id"`
	Title               string  `json:"title"`
	Price               string  `json:"price"`
	Sku                 string  `json:"sku"`
	Position            int32   `json:"position"`
	CompareAtPrice      *string `json:"compare_at_price"`
	InventoryQuantity   int32   `json:"inventory_quantity"`
	InventoryManagement *string `json:"inventory_management"`
	Option1             *string `json:"option1"`
	Option2             *string `json:"option2"`
	Option3             *string `json:"option3"`
}

type shopifyProductOption struct {
	ID        int64    `json:"id"`
	ProductID int64    `json:"product_id"`
	Name      string   `json:"name"`
	Position  int32    `json:"position"`
	Values    []string `json:"values"`
}

type shopifyProduct struct {
	ID          int64                   `json:"id"`
	Title       string                  `json:"title"`
	BodyHTML    string                  `json:"body_html"`
	Vendor      string                  `json:"vendor"`
	ProductType string                  `json:"product_type"`
	Handle      string                  `json:"handle"`
	Status      string                  `json:"status"`
	Tags        string                  `json:"tags"`
	Variants    []shopifyProductVariant `json:"variants"`
	Options     []shopifyProductOption  `json:"options"`
	Images      []shopifyProductImage   `json:"images"`
}

type shopifyProductsResponse struct {
	Products []shopifyProduct `json:"products"`
}

func normalizeShopifyURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("store URL cannot be empty")
	}

	scheme := "https"
	if strings.HasPrefix(raw, "http://") && (strings.Contains(raw, "127.0.0.1") || strings.Contains(raw, "localhost")) {
		scheme = "http"
	}

	clean := raw
	if strings.HasPrefix(clean, "https://") {
		clean = strings.TrimPrefix(clean, "https://")
	} else if strings.HasPrefix(clean, "http://") {
		clean = strings.TrimPrefix(clean, "http://")
	}
	clean = strings.TrimRight(clean, "/")

	// If the user supplied a full admin API path
	if strings.Contains(clean, "/admin/api/") {
		urlStr := scheme + "://" + clean
		if !strings.Contains(urlStr, "limit=") {
			sep := "?"
			if strings.Contains(urlStr, "?") {
				sep = "&"
			}
			urlStr += sep + "limit=250"
		}
		return urlStr, nil
	}

	// Extract host
	parts := strings.SplitN(clean, "/", 2)
	host := parts[0]
	if !strings.Contains(host, ".") {
		host = host + ".myshopify.com"
	}
	return fmt.Sprintf("%s://%s/admin/api/2024-01/products.json?limit=250", scheme, host), nil
}

func parseShopifyNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	links := strings.Split(linkHeader, ",")
	for _, l := range links {
		parts := strings.Split(strings.TrimSpace(l), ";")
		if len(parts) >= 2 {
			for _, p := range parts[1:] {
				if strings.Contains(strings.TrimSpace(p), `rel="next"`) {
					return strings.Trim(strings.TrimSpace(parts[0]), "<>")
				}
			}
		}
	}
	return ""
}

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' {
			b.WriteRune('-')
		}
	}
	cleanSlug := b.String()
	for strings.Contains(cleanSlug, "--") {
		cleanSlug = strings.ReplaceAll(cleanSlug, "--", "-")
	}
	return strings.Trim(cleanSlug, "-")
}

func (h *Handler) ImportShopify(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "shop_id missing from context")
		return
	}

	var req ImportShopifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	storeURL := req.StoreURL
	if storeURL == "" {
		storeURL = req.ShopURL
	}
	if strings.TrimSpace(storeURL) == "" {
		response.BadRequest(c, "store_url is required")
		return
	}
	if strings.TrimSpace(req.AccessToken) == "" {
		response.BadRequest(c, "access_token is required")
		return
	}

	targetURL, err := normalizeShopifyURL(storeURL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	client := &http.Client{Timeout: 45 * time.Second}
	var allShopifyProducts []shopifyProduct
	currentURL := targetURL
	const maxPages = 10 // Safe cap: up to 2500 products

	for page := 0; page < maxPages && currentURL != ""; page++ {
		httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, currentURL, nil)
		if err != nil {
			response.BadRequest(c, fmt.Sprintf("Failed to construct Shopify request: %v", err))
			return
		}
		httpReq.Header.Set("X-Shopify-Access-Token", strings.TrimSpace(req.AccessToken))
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("User-Agent", "NexusCommerce-Importer/1.0")

		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Could not connect to Shopify store: %v", err)})
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			c.JSON(http.StatusBadGateway, gin.H{"error": "Shopify authentication failed. Please verify your Admin API Access Token."})
			return
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			c.JSON(http.StatusBadGateway, gin.H{"error": "Shopify store endpoint not found. Please verify your Store URL."})
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			resp.Body.Close()
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Shopify API returned error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))})
			return
		}

		var shopifyResp shopifyProductsResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&shopifyResp)
		linkHeader := resp.Header.Get("Link")
		resp.Body.Close()

		if decodeErr != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Failed to parse response from Shopify: %v", decodeErr)})
			return
		}

		allShopifyProducts = append(allShopifyProducts, shopifyResp.Products...)
		currentURL = parseShopifyNextLink(linkHeader)
	}

	if len(allShopifyProducts) == 0 {
		response.OK(c, gin.H{
			"message":        "No products found in the Shopify store.",
			"imported_count": 0,
			"total_found":    0,
		})
		return
	}

	successCount := 0
	var importErrors []string
	limitHit := false

	for _, sp := range allShopifyProducts {
		title := strings.TrimSpace(sp.Title)
		if title == "" {
			continue
		}
		if len(title) > 255 {
			title = title[:255]
		}

		slug := strings.TrimSpace(sp.Handle)
		if slug == "" {
			slug = title
		}
		slug = slugify(slug)
		if slug == "" {
			slug = fmt.Sprintf("shopify-product-%d", sp.ID)
		}
		if len(slug) > 255 {
			slug = slug[:255]
		}

		var price float64 = 0.01
		var compareAtPrice float64 = 0
		var totalStock int32 = 0
		sku := ""

		if len(sp.Variants) > 0 {
			firstV := sp.Variants[0]
			if pVal, err := strconv.ParseFloat(firstV.Price, 64); err == nil && pVal > 0 {
				price = pVal
			}
			if firstV.CompareAtPrice != nil && *firstV.CompareAtPrice != "" {
				if cVal, err := strconv.ParseFloat(*firstV.CompareAtPrice, 64); err == nil && cVal > 0 {
					compareAtPrice = cVal
				}
			}
			sku = firstV.Sku
			for _, v := range sp.Variants {
				totalStock += v.InventoryQuantity
			}
		}

		status := strings.ToLower(sp.Status)
		if status != "active" && status != "draft" && status != "archived" {
			status = "active"
		}

		desc := sp.BodyHTML

		createReq := CreateProductRequest{
			Title:          title,
			Slug:           slug,
			Description:    desc,
			Price:          price,
			CompareAtPrice: compareAtPrice,
			SKU:            sku,
			TrackInventory: true,
			StockQuantity:  totalStock,
			Status:         status,
			SeoTitle:       title,
		}

		prod, err := h.svc.CreateProduct(c.Request.Context(), shopID, createReq)
		if err != nil && errors.Is(err, ErrConflict) {
			// Deduplicate slug collision with random suffix.
			uniqueSuffix := uuid.New().String()[:6]
			if len(slug) > 248 {
				slug = slug[:248]
			}
			createReq.Slug = fmt.Sprintf("%s-%s", slug, uniqueSuffix)
			prod, err = h.svc.CreateProduct(c.Request.Context(), shopID, createReq)
		}

		if err != nil {
			if errors.Is(err, ErrProductLimit) {
				limitHit = true
				importErrors = append(importErrors, "Product limit reached for current plan")
				break
			}
			importErrors = append(importErrors, fmt.Sprintf("%s: %v", title, err))
			continue
		}

		if len(sp.Variants) > 0 {
			for _, v := range sp.Variants {
				if len(sp.Variants) == 1 && (v.Title == "Default Title" || v.Title == "") {
					continue
				}
				vTitle := v.Title
				if vTitle == "" {
					vTitle = "Default"
				}
				if len(vTitle) > 255 {
					vTitle = vTitle[:255]
				}

				vPrice := price
				if pVal, err := strconv.ParseFloat(v.Price, 64); err == nil && pVal > 0 {
					vPrice = pVal
				}
				var vCompare float64
				if v.CompareAtPrice != nil && *v.CompareAtPrice != "" {
					if cVal, err := strconv.ParseFloat(*v.CompareAtPrice, 64); err == nil && cVal > 0 {
						vCompare = cVal
					}
				}

				vOptions := make(map[string]interface{})
				if len(sp.Options) > 0 && v.Option1 != nil && *v.Option1 != "" {
					vOptions[sp.Options[0].Name] = *v.Option1
				}
				if len(sp.Options) > 1 && v.Option2 != nil && *v.Option2 != "" {
					vOptions[sp.Options[1].Name] = *v.Option2
				}
				if len(sp.Options) > 2 && v.Option3 != nil && *v.Option3 != "" {
					vOptions[sp.Options[2].Name] = *v.Option3
				}

				varReq := CreateVariantRequest{
					Title:          vTitle,
					SKU:            v.Sku,
					Price:          vPrice,
					CompareAtPrice: vCompare,
					StockQuantity:  v.InventoryQuantity,
					Options:        vOptions,
					Position:       v.Position,
				}
				_, _ = h.svc.CreateVariant(c.Request.Context(), shopID, prod.ID, varReq)
			}
		}

		for i, img := range sp.Images {
			src := strings.TrimSpace(img.Src)
			if src == "" {
				continue
			}
			if strings.HasPrefix(src, "//") {
				src = "https:" + src
			}
			alt := ""
			if img.Alt != nil {
				alt = *img.Alt
			}
			isPrimary := (i == 0 || img.Position == 1)

			finalURL := src
			if h.storage != nil {
				resp, err := http.Get(src)
				if err == nil && resp.StatusCode == 200 {
					contentType := resp.Header.Get("Content-Type")
					if contentType == "" {
						contentType = "image/jpeg"
					}
					key := fmt.Sprintf("shops/%s/products/%s/%d", shopID, prod.ID, img.ID)
					r2url, err := h.storage.Upload(c.Request.Context(), key, contentType, resp.Body)
					if err == nil && r2url != "" {
						finalURL = r2url
					}
					resp.Body.Close()
				} else if err == nil {
					resp.Body.Close()
				}
			}

			_, _ = h.svc.CreateImage(c.Request.Context(), shopID, prod.ID, finalURL, alt, "", img.Position, isPrimary)
		}

		successCount++
	}

	msg := fmt.Sprintf("Successfully imported %d of %d products from Shopify.", successCount, len(allShopifyProducts))
	if limitHit {
		msg += " Note: Product plan limit reached."
	}

	response.OK(c, gin.H{
		"message":        msg,
		"imported_count": successCount,
		"total_found":    len(allShopifyProducts),
		"errors":         importErrors,
	})
}

func (h *Handler) ImportShopifyProducts(c *gin.Context) {
	h.ImportShopify(c)
}
