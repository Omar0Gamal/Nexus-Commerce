package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend-api/internal/db"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound     = apperr.ErrNotFound
	ErrConflict     = apperr.ErrConflict
	ErrInvalidUUID  = apperr.ErrInvalidUUID
	ErrProductLimit = errors.New("product limit reached for current plan")
)

// PlanLimitChecker decouples catalog from billing to avoid import cycles.
type PlanLimitChecker interface {
	CheckProductLimit(ctx context.Context, shopID string) error
}

type Service struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	billing PlanLimitChecker
	rdb     *redis.Client
}

func NewService(queries *db.Queries, pool *pgxpool.Pool, billing PlanLimitChecker, rdb *redis.Client) *Service {
	return &Service{queries: queries, pool: pool, billing: billing, rdb: rdb}
}

func (s *Service) Q() *db.Queries { return s.queries }

func (s *Service) ListProducts(ctx context.Context, shopID string, params ListQueryParams) ([]ProductResponse, int64, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, 0, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	limit := int32(perPage)
	offset := int32((page - 1) * perPage)

	if params.Search != "" || params.MinPrice != nil || params.MaxPrice != nil {
		return s.listProductsSearch(ctx, shopPgUUID, params, limit, offset, shopUUID)
	}

	statusFilter := pgtype.Text{}
	if params.Status != "" {
		statusFilter = pgtype.Text{String: params.Status, Valid: true}
	}
	categoryFilter := pgtype.UUID{}
	if params.CategoryID != "" {
		if catID, err := uuid.Parse(params.CategoryID); err == nil {
			categoryFilter = pgtype.UUID{Bytes: catID, Valid: true}
		}
	}

	if params.CursorCreatedAt != "" || params.CursorID != "" {
		cursorCreatedAt := pgtype.Timestamptz{}
		if params.CursorCreatedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, params.CursorCreatedAt); err == nil {
				cursorCreatedAt = pgtype.Timestamptz{Time: t, Valid: true}
			}
		}
		cursorPgID := pgtype.UUID{}
		if params.CursorID != "" {
			if id, err := uuid.Parse(params.CursorID); err == nil {
				cursorPgID = pgtype.UUID{Bytes: id, Valid: true}
			}
		}

		rows, err := s.queries.ListProductsCursor(ctx, db.ListProductsCursorParams{
			ShopID:          shopPgUUID,
			Limit:           limit,
			Status:          statusFilter,
			CategoryID:      categoryFilter,
			CursorCreatedAt: cursorCreatedAt,
			CursorID:        cursorPgID,
		})
		if err != nil {
			return nil, 0, mapErr(err)
		}
		productIDs := make([]uuid.UUID, len(rows))
		for i, p := range rows {
			productIDs[i] = p.ID
		}
		variantMap, _ := s.buildVariantMap(ctx, productIDs, shopUUID)
		imageMap, _ := s.buildImageMap(ctx, productIDs, shopUUID)
		products := make([]ProductResponse, 0, len(rows))
		for _, p := range rows {
			products = append(products, mapProduct(p, variantMap[p.ID], imageMap[p.ID]))
		}
		// -1 signals keyset pagination mode (no total count calculated)
		return products, -1, nil
	}

	rows, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		ShopID:     shopPgUUID,
		Limit:      limit,
		Offset:     offset,
		Status:     statusFilter,
		CategoryID: categoryFilter,
	})
	if err != nil {
		return nil, 0, mapErr(err)
	}

	count, err := s.queries.CountProducts(ctx, db.CountProductsParams{
		ShopID:     shopPgUUID,
		Status:     statusFilter,
		CategoryID: categoryFilter,
	})
	if err != nil {
		return nil, 0, mapErr(err)
	}

	productIDs := make([]uuid.UUID, len(rows))
	for i, p := range rows {
		productIDs[i] = p.ID
	}
	variantMap, _ := s.buildVariantMap(ctx, productIDs, shopUUID)
	imageMap, _ := s.buildImageMap(ctx, productIDs, shopUUID)
	products := make([]ProductResponse, 0, len(rows))
	for _, p := range rows {
		products = append(products, mapProduct(p, variantMap[p.ID], imageMap[p.ID]))
	}
	return products, count, nil
}

func (s *Service) listProductsSearch(ctx context.Context, shopPgUUID pgtype.UUID, params ListQueryParams, limit, offset int32, shopUUID uuid.UUID) ([]ProductResponse, int64, error) {
	var minPrice, maxPrice pgtype.Numeric
	if params.MinPrice != nil {
		if err := minPrice.Scan(fmt.Sprintf("%.2f", *params.MinPrice)); err != nil {
			minPrice = pgtype.Numeric{}
		}
	}
	if params.MaxPrice != nil {
		if err := maxPrice.Scan(fmt.Sprintf("%.2f", *params.MaxPrice)); err != nil {
			maxPrice = pgtype.Numeric{}
		}
	}

	catID := uuid.Nil
	if params.CategoryID != "" {
		if id, err := uuid.Parse(params.CategoryID); err == nil {
			catID = id
		}
	}

	query := params.Search
	if query == "" {
		query = "*"
	}

	rows, err := s.queries.SearchProductsFTS(ctx, db.SearchProductsFTSParams{
		ShopID:     shopPgUUID,
		Query:      query,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		CategoryID: catID,
		Lim:        limit,
		Off:        offset,
	})
	if err != nil {
		return nil, 0, mapErr(err)
	}

	if s.rdb != nil && params.Search != "" {
		key := fmt.Sprintf("search:popular:%s", shopUUID.String())
		_ = s.rdb.ZIncrBy(ctx, key, 1, strings.ToLower(params.Search)).Err()
		sharedanalytics.Publish(ctx, s.rdb, shopUUID.String(), sharedanalytics.Event{
			Event:  "product_search",
			ShopID: shopUUID.String(),
			Payload: map[string]any{
				"query":        strings.ToLower(params.Search),
				"result_count": len(rows),
			},
		})
	}

	productIDs := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		productIDs[i] = r.ID
	}
	variantMap, _ := s.buildVariantMap(ctx, productIDs, shopUUID)
	imageMap, _ := s.buildImageMap(ctx, productIDs, shopUUID)

	products := make([]ProductResponse, 0, len(rows))
	for _, r := range rows {
		products = append(products, mapProduct(ftsRowToProduct(r), variantMap[r.ID], imageMap[r.ID]))
	}

	// Approximate pagination count for full-text search.
	var total int64
	if int32(len(rows)) < limit {
		total = int64(offset) + int64(len(rows))
	} else {
		countSQL := `
			SELECT COUNT(*) FROM products p
			WHERE p.shop_id = $1
			  AND p.deleted_at IS NULL
			  AND (
			      p.search_vector @@ plainto_tsquery('english', $2)
			      OR similarity(p.title, $2) > 0.2
			  )`
		countCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := s.pool.QueryRow(countCtx, countSQL, shopPgUUID, query).Scan(&total); err != nil {
			total = int64(offset) + int64(len(rows))
		}
		if total == 0 {
			total = int64(offset) + int64(len(rows))
		}
	}
	return products, total, nil
}

func ftsRowToProduct(r db.SearchProductsFTSRow) db.Product {
	return db.Product{
		ID:                r.ID,
		ShopID:            r.ShopID,
		CategoryID:        r.CategoryID,
		Title:             r.Title,
		Slug:              r.Slug,
		Description:       r.Description,
		Price:             r.Price,
		CompareAtPrice:    r.CompareAtPrice,
		TrackInventory:    r.TrackInventory,
		StockQuantity:     r.StockQuantity,
		LowStockThreshold: r.LowStockThreshold,
		Sku:               r.Sku,
		Status:            r.Status,
		SeoTitle:          r.SeoTitle,
		SeoDescription:    r.SeoDescription,
		CreatedAt:         r.CreatedAt,
		DeletedAt:         r.DeletedAt,
	}
}

type SearchSuggestion struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

func (s *Service) GetSearchSuggestions(ctx context.Context, shopID, prefix string, limit int) ([]SearchSuggestion, error) {
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	var results []SearchSuggestion

	if s.rdb != nil {
		key := fmt.Sprintf("search:popular:%s", shopID)
		popular, err := s.rdb.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
		if err == nil {
			lower := strings.ToLower(prefix)
			for _, term := range popular {
				if strings.HasPrefix(term, lower) {
					results = append(results, SearchSuggestion{Text: term, Type: "popular"})
					if len(results) >= limit {
						return results, nil
					}
				}
			}
		}
	}

	if len(results) < limit {
		shopPgUUID, err := pgutil.ParseUUID(shopID)
		if err != nil {
			return results, nil
		}
		remain := int32(limit - len(results))
		const suggestSQL = `
			SELECT DISTINCT title FROM products
			WHERE shop_id = $1
			  AND deleted_at IS NULL
			  AND title ILIKE $2
			ORDER BY title
			LIMIT $3`
		rows, err := s.pool.Query(ctx, suggestSQL, shopPgUUID, prefix+"%", remain)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var title string
				if rows.Scan(&title) == nil {
					results = append(results, SearchSuggestion{Text: title, Type: "product"})
				}
			}
		}
	}

	return results, nil
}

type FacetValue struct {
	Value string `json:"value"`
	Count int32  `json:"count"`
}

type Facet struct {
	Key    string       `json:"key"`
	Values []FacetValue `json:"values"`
}

func (s *Service) GetSearchFacets(ctx context.Context, shopID string) ([]Facet, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	rows, err := s.queries.GetSearchFacets(ctx, shopUUID)
	if err != nil {
		return nil, mapErr(err)
	}

	grouped := make(map[string][]FacetValue)
	order := make([]string, 0)
	for _, r := range rows {
		if _, exists := grouped[r.Key]; !exists {
			order = append(order, r.Key)
			grouped[r.Key] = nil
		}
		grouped[r.Key] = append(grouped[r.Key], FacetValue{Value: r.Value, Count: r.Count})
	}

	facets := make([]Facet, 0, len(order))
	for _, k := range order {
		facets = append(facets, Facet{Key: k, Values: grouped[k]})
	}
	return facets, nil
}

func (s *Service) GetRelatedProducts(ctx context.Context, shopID string, productID uuid.UUID, limit int) ([]ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	if limit <= 0 || limit > 20 {
		limit = 6
	}

	const copurchaseSQL = `
		SELECT p.id, p.shop_id, p.category_id, p.title, p.slug, p.description, p.price,
		       p.compare_at_price, p.track_inventory, p.stock_quantity, p.low_stock_threshold,
		       p.sku, p.status, p.seo_title, p.seo_description
		FROM order_items oi1
		JOIN order_items oi2 ON oi1.order_id = oi2.order_id
		JOIN orders o ON o.id = oi1.order_id
		JOIN products p ON p.id = oi2.product_id
		WHERE o.shop_id = $1
		  AND oi1.product_id = $2
		  AND oi2.product_id != $2
		  AND o.status = 'completed'
		  AND p.deleted_at IS NULL
		GROUP BY p.id, p.shop_id, p.category_id, p.title, p.slug, p.description, p.price,
		         p.compare_at_price, p.track_inventory, p.stock_quantity, p.low_stock_threshold,
		         p.sku, p.status, p.seo_title, p.seo_description
		ORDER BY COUNT(*) DESC
		LIMIT $3`

	rows, err := s.pool.Query(ctx, copurchaseSQL, shopPgUUID, productID, int32(limit))
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	var products []db.Product
	for rows.Next() {
		var p db.Product
		if err := rows.Scan(
			&p.ID, &p.ShopID, &p.CategoryID, &p.Title, &p.Slug,
			&p.Description, &p.Price, &p.CompareAtPrice, &p.TrackInventory,
			&p.StockQuantity, &p.LowStockThreshold, &p.Sku, &p.Status,
			&p.SeoTitle, &p.SeoDescription,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	relatedProductIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		relatedProductIDs[i] = p.ID
	}
	relatedVariantMap, _ := s.buildVariantMap(ctx, relatedProductIDs, shopUUID)
	relatedImageMap, _ := s.buildImageMap(ctx, relatedProductIDs, shopUUID)
	relatedProducts := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		relatedProducts = append(relatedProducts, mapProduct(p, relatedVariantMap[p.ID], relatedImageMap[p.ID]))
	}
	return relatedProducts, nil
}

func (s *Service) ListProductsForFeed(ctx context.Context, shopID string) ([]ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	statusActive := "active"
	rows, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		ShopID:     shopPgUUID,
		Status:     pgtype.Text{String: statusActive, Valid: true},
		CategoryID: pgtype.UUID{},
		Limit:      500,
		Offset:     0,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	productIDs := make([]uuid.UUID, len(rows))
	for i, p := range rows {
		productIDs[i] = p.ID
	}
	variantMap, _ := s.buildVariantMap(ctx, productIDs, shopUUID)
	imageMap, _ := s.buildImageMap(ctx, productIDs, shopUUID)

	products := make([]ProductResponse, 0, len(rows))
	for _, p := range rows {
		products = append(products, mapProduct(p, variantMap[p.ID], imageMap[p.ID]))
	}
	return products, nil
}

func (s *Service) CreateProduct(ctx context.Context, shopID string, req CreateProductRequest) (*ProductResponse, error) {
	if s.billing != nil {
		if err := s.billing.CheckProductLimit(ctx, shopID); err != nil {
			return nil, ErrProductLimit
		}
	}

	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	status := req.Status
	if status == "" {
		status = "draft"
	}

	req.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Slug), " ", "-"))

	categoryID := pgtype.UUID{}
	if req.CategoryID != "" {
		if id, err := uuid.Parse(req.CategoryID); err == nil {
			categoryID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	p, err := s.queries.CreateProduct(ctx, db.CreateProductParams{
		ShopID:            shopPgUUID,
		CategoryID:        categoryID,
		Title:             req.Title,
		Slug:              req.Slug,
		Description:       pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Price:             pgutil.Float64ToNumeric(req.Price),
		CompareAtPrice:    pgutil.Float64ToNumeric(req.CompareAtPrice),
		TrackInventory:    pgtype.Bool{Bool: req.TrackInventory, Valid: true},
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: lowStockThreshold(req.LowStockThreshold),
		Sku:               pgtype.Text{String: req.SKU, Valid: req.SKU != ""},
		Status:            pgtype.Text{String: status, Valid: true},
	})
	if err != nil {
		return nil, mapErr(err)
	}

	if req.SeoTitle != "" || req.SeoDescription != "" {
		p.SeoTitle = pgtype.Text{String: req.SeoTitle, Valid: req.SeoTitle != ""}
		p.SeoDescription = pgtype.Text{String: req.SeoDescription, Valid: req.SeoDescription != ""}
		_ = s.queries.UpdateProductSEO(ctx, db.UpdateProductSEOParams{
			ID:             p.ID,
			ShopID:         shopPgUUID,
			SeoTitle:       p.SeoTitle,
			SeoDescription: p.SeoDescription,
		})
	}
	variants, _ := s.getVariants(ctx, p.ID, shopUUID)
	images, _ := s.getImages(ctx, p.ID, shopUUID)
	resp := mapProduct(p, variants, images)
	return &resp, nil
}

// GetProduct retrieves a product by ID, scoped to shop.
func (s *Service) GetProduct(ctx context.Context, shopID, productID string) (*ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	p, err := s.queries.GetProduct(ctx, db.GetProductParams{
		ID:     prodUUID,
		ShopID: shopPgUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	variants, _ := s.getVariants(ctx, p.ID, shopUUID)
	images, _ := s.getImages(ctx, p.ID, shopUUID)
	resp := mapProduct(p, variants, images)
	return &resp, nil
}

// GetProductBySlug retrieves a product by slug, scoped to shop.
func (s *Service) GetProductBySlug(ctx context.Context, shopID, slug string) (*ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	p, err := s.queries.GetProductBySlug(ctx, db.GetProductBySlugParams{
		Slug:   slug,
		ShopID: shopPgUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	variants, _ := s.getVariants(ctx, p.ID, shopUUID)
	images, _ := s.getImages(ctx, p.ID, shopUUID)
	resp := mapProduct(p, variants, images)
	return &resp, nil
}

// UpdateProduct applies partial updates to a product.
func (s *Service) UpdateProduct(ctx context.Context, shopID, productID string, req UpdateProductRequest) (*ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)

	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	params := db.UpdateProductParams{
		ID:     prodUUID,
		ShopID: shopPgUUID,
	}
	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}
	if req.Slug != nil {
		params.Slug = pgtype.Text{String: *req.Slug, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.Price != nil {
		params.Price = pgutil.Float64ToNumeric(*req.Price)
	}
	if req.CompareAtPrice != nil {
		params.CompareAtPrice = pgutil.Float64ToNumeric(*req.CompareAtPrice)
	}
	if req.TrackInventory != nil {
		params.TrackInventory = pgtype.Bool{Bool: *req.TrackInventory, Valid: true}
	}
	if req.StockQuantity != nil {
		params.StockQuantity = pgtype.Int4{Int32: *req.StockQuantity, Valid: true}
	}
	if req.LowStockThreshold != nil {
		params.LowStockThreshold = pgtype.Int4{Int32: *req.LowStockThreshold, Valid: true}
	}
	if req.SKU != nil {
		params.Sku = pgtype.Text{String: *req.SKU, Valid: true}
	}
	if req.Status != nil {
		params.Status = pgtype.Text{String: *req.Status, Valid: true}
	}
	if req.CategoryID != nil {
		if id, err := uuid.Parse(*req.CategoryID); err == nil {
			params.CategoryID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	p, err := s.queries.UpdateProduct(ctx, params)
	if err != nil {
		return nil, mapErr(err)
	}

	if req.SeoTitle != nil || req.SeoDescription != nil {
		if req.SeoTitle != nil {
			p.SeoTitle = pgtype.Text{String: *req.SeoTitle, Valid: *req.SeoTitle != ""}
		}
		if req.SeoDescription != nil {
			p.SeoDescription = pgtype.Text{String: *req.SeoDescription, Valid: *req.SeoDescription != ""}
		}
		_ = s.queries.UpdateProductSEO(ctx, db.UpdateProductSEOParams{
			ID:             p.ID,
			ShopID:         shopPgUUID,
			SeoTitle:       p.SeoTitle,
			SeoDescription: p.SeoDescription,
		})
	}
	variants, _ := s.getVariants(ctx, p.ID, shopUUID)
	images, _ := s.getImages(ctx, p.ID, shopUUID)
	resp := mapProduct(p, variants, images)
	return &resp, nil
}

// DeleteProduct removes a product and its variants/images.
func (s *Service) DeleteProduct(ctx context.Context, shopID, productID string) error {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return ErrInvalidUUID
	}
	return mapErr(s.queries.DeleteProduct(ctx, db.DeleteProductParams{
		ID:     prodUUID,
		ShopID: shopPgUUID,
	}))
}

// SoftDeleteProduct moves a product to the archive (sets deleted_at).
func (s *Service) SoftDeleteProduct(ctx context.Context, shopID, productID string) error {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return ErrInvalidUUID
	}
	return mapErr(s.queries.SoftDeleteProduct(ctx, db.SoftDeleteProductParams{
		ID:     prodUUID,
		ShopID: shopPgUUID,
	}))
}

// RestoreProduct un-archives a previously soft-deleted product.
func (s *Service) RestoreProduct(ctx context.Context, shopID, productID string) (*ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)
	p, err := s.queries.RestoreProduct(ctx, db.RestoreProductParams{
		ID:     prodUUID,
		ShopID: shopPgUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	variants, _ := s.getVariants(ctx, p.ID, shopUUID)
	images, _ := s.getImages(ctx, p.ID, shopUUID)
	resp := mapProduct(p, variants, images)
	return &resp, nil
}

func (s *Service) ListArchivedProducts(ctx context.Context, shopID string) ([]ProductResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopUUID := uuid.UUID(shopPgUUID.Bytes)
	rows, err := s.queries.ListDeletedProducts(ctx, shopPgUUID)
	if err != nil {
		return nil, fmt.Errorf("list archived products: %w", err)
	}
	productIDs := make([]uuid.UUID, len(rows))
	for i, p := range rows {
		productIDs[i] = p.ID
	}
	variantMap, _ := s.buildVariantMap(ctx, productIDs, shopUUID)
	imageMap, _ := s.buildImageMap(ctx, productIDs, shopUUID)
	products := make([]ProductResponse, len(rows))
	for i, p := range rows {
		products[i] = mapProduct(p, variantMap[p.ID], imageMap[p.ID])
	}
	return products, nil
}

func (s *Service) CreateCategory(ctx context.Context, shopID string, req CreateCategoryRequest) (*CategoryResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	parentID := pgtype.UUID{}
	if req.ParentID != "" {
		if id, err := uuid.Parse(req.ParentID); err == nil {
			parentID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	cat, err := s.queries.CreateCategory(ctx, db.CreateCategoryParams{
		ShopID:   shopPgUUID,
		ParentID: parentID,
		Name:     req.Name,
		Slug:     strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Slug), " ", "-")),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	if req.MetaDescription != "" {
		cat.MetaDescription = pgtype.Text{String: req.MetaDescription, Valid: true}
		_ = s.queries.UpdateCategorySEO(ctx, db.UpdateCategorySEOParams{
			ID:              cat.ID,
			ShopID:          shopPgUUID,
			MetaDescription: cat.MetaDescription,
		})
	}
	resp := mapCategory(cat)
	return &resp, nil
}

func (s *Service) GetCategory(ctx context.Context, shopID, categoryID string) (*CategoryResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	catUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	cat, err := s.queries.GetCategory(ctx, db.GetCategoryParams{
		ID:     catUUID,
		ShopID: shopPgUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	resp := mapCategory(cat)
	return &resp, nil
}

func (s *Service) GetCategoryBySlug(ctx context.Context, shopID, slug string) (*CategoryResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	cat, err := s.queries.GetCategoryBySlug(ctx, db.GetCategoryBySlugParams{
		Slug:   slug,
		ShopID: shopPgUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	resp := mapCategory(cat)
	return &resp, nil
}

func (s *Service) ListCategories(ctx context.Context, shopID string) ([]CategoryResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	cats, err := s.queries.ListCategories(ctx, shopPgUUID)
	if err != nil {
		return nil, mapErr(err)
	}

	categories := make([]CategoryResponse, 0, len(cats))
	for _, c := range cats {
		categories = append(categories, mapCategory(c))
	}
	return categories, nil
}

func (s *Service) UpdateCategory(ctx context.Context, shopID, categoryID string, req UpdateCategoryRequest) (*CategoryResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	catUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	params := db.UpdateCategoryParams{
		ID:     catUUID,
		ShopID: shopPgUUID,
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Slug != nil {
		params.Slug = pgtype.Text{String: *req.Slug, Valid: true}
	}
	if req.ParentID != nil {
		if id, err := uuid.Parse(*req.ParentID); err == nil {
			params.ParentID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	cat, err := s.queries.UpdateCategory(ctx, params)
	if err != nil {
		return nil, mapErr(err)
	}
	if req.MetaDescription != nil {
		cat.MetaDescription = pgtype.Text{String: *req.MetaDescription, Valid: *req.MetaDescription != ""}
		_ = s.queries.UpdateCategorySEO(ctx, db.UpdateCategorySEOParams{
			ID:              cat.ID,
			ShopID:          shopPgUUID,
			MetaDescription: cat.MetaDescription,
		})
	}
	resp := mapCategory(cat)
	return &resp, nil
}

func (s *Service) DeleteCategory(ctx context.Context, shopID, categoryID string) error {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	catUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return ErrInvalidUUID
	}
	return mapErr(s.queries.DeleteCategory(ctx, db.DeleteCategoryParams{
		ID:     catUUID,
		ShopID: shopPgUUID,
	}))
}

func (s *Service) ListVariants(ctx context.Context, shopID, productID string) ([]VariantResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	variants, err := s.queries.ListVariantsByProduct(ctx, db.ListVariantsByProductParams{
		ProductID: prodUUID,
		ShopID:    shopUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	variantResponses := make([]VariantResponse, 0, len(variants))
	for _, v := range variants {
		variantResponses = append(variantResponses, mapVariant(v))
	}
	return variantResponses, nil
}

func (s *Service) CreateVariant(ctx context.Context, shopID, productID string, req CreateVariantRequest) (*VariantResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	optionsJSON, _ := json.Marshal(req.Options)

	v, err := s.queries.CreateVariant(ctx, db.CreateVariantParams{
		ProductID:      prodUUID,
		ShopID:         shopUUID,
		Title:          req.Title,
		Sku:            pgtype.Text{String: req.SKU, Valid: req.SKU != ""},
		Price:          pgutil.Float64ToNumeric(req.Price),
		CompareAtPrice: pgutil.Float64ToNumeric(req.CompareAtPrice),
		StockQuantity:  req.StockQuantity,
		Options:        optionsJSON,
		Position:       pgtype.Int4{Int32: req.Position, Valid: true},
	})
	if err != nil {
		return nil, mapErr(err)
	}

	resp := mapVariant(v)
	return &resp, nil
}

func (s *Service) UpdateVariant(ctx context.Context, shopID, variantID string, req UpdateVariantRequest) (*VariantResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	varUUID, err := uuid.Parse(variantID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	params := db.UpdateVariantParams{
		ID:     varUUID,
		ShopID: shopUUID,
	}
	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}
	if req.SKU != nil {
		params.Sku = pgtype.Text{String: *req.SKU, Valid: true}
	}
	if req.Price != nil {
		params.Price = pgutil.Float64ToNumeric(*req.Price)
	}
	if req.CompareAtPrice != nil {
		params.CompareAtPrice = pgutil.Float64ToNumeric(*req.CompareAtPrice)
	}
	if req.StockQuantity != nil {
		params.StockQuantity = pgtype.Int4{Int32: *req.StockQuantity, Valid: true}
	}
	if req.Options != nil {
		optionsJSON, _ := json.Marshal(req.Options)
		params.Options = optionsJSON
	}
	if req.Position != nil {
		params.Position = pgtype.Int4{Int32: *req.Position, Valid: true}
	}

	v, err := s.queries.UpdateVariant(ctx, params)
	if err != nil {
		return nil, mapErr(err)
	}

	resp := mapVariant(v)
	return &resp, nil
}

func (s *Service) DeleteVariant(ctx context.Context, shopID, variantID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	varUUID, err := uuid.Parse(variantID)
	if err != nil {
		return ErrInvalidUUID
	}
	return mapErr(s.queries.DeleteVariant(ctx, db.DeleteVariantParams{
		ID:     varUUID,
		ShopID: shopUUID,
	}))
}

func (s *Service) ListImages(ctx context.Context, shopID, productID string) ([]ImageResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	images, err := s.queries.ListImagesByProduct(ctx, db.ListImagesByProductParams{
		ProductID: prodUUID,
		ShopID:    shopUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	imageResponses := make([]ImageResponse, 0, len(images))
	for _, img := range images {
		imageResponses = append(imageResponses, mapImage(img))
	}
	return imageResponses, nil
}

func (s *Service) CreateImage(ctx context.Context, shopID, productID, urlPath, altText, variantIDStr string, position int32, isPrimary bool) (*ImageResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	prodUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	variantID := pgtype.UUID{}
	if variantIDStr != "" {
		if id, err := uuid.Parse(variantIDStr); err == nil {
			variantID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	img, err := s.queries.CreateImage(ctx, db.CreateImageParams{
		ProductID: prodUUID,
		ShopID:    shopUUID,
		VariantID: variantID,
		Url:       urlPath,
		AltText:   pgtype.Text{String: altText, Valid: altText != ""},
		Position:  pgtype.Int4{Int32: position, Valid: true},
		IsPrimary: pgtype.Bool{Bool: isPrimary, Valid: true},
	})
	if err != nil {
		return nil, mapErr(err)
	}

	resp := mapImage(img)
	return &resp, nil
}

func (s *Service) UpdateImage(ctx context.Context, shopID, imageID string, req UpdateImageRequest) (*ImageResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	imgUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	params := db.UpdateImageParams{
		ID:     imgUUID,
		ShopID: shopUUID,
	}
	if req.AltText != nil {
		params.AltText = pgtype.Text{String: *req.AltText, Valid: true}
	}
	if req.Position != nil {
		params.Position = pgtype.Int4{Int32: *req.Position, Valid: true}
	}
	if req.IsPrimary != nil {
		params.IsPrimary = pgtype.Bool{Bool: *req.IsPrimary, Valid: true}
	}
	if req.VariantID != nil {
		if id, err := uuid.Parse(*req.VariantID); err == nil {
			params.VariantID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	img, err := s.queries.UpdateImage(ctx, params)
	if err != nil {
		return nil, mapErr(err)
	}

	resp := mapImage(img)
	return &resp, nil
}

func (s *Service) DeleteImage(ctx context.Context, shopID, imageID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	imgUUID, err := uuid.Parse(imageID)
	if err != nil {
		return ErrInvalidUUID
	}
	return mapErr(s.queries.DeleteImage(ctx, db.DeleteImageParams{
		ID:     imgUUID,
		ShopID: shopUUID,
	}))
}

// Batch fetch all variants for a product set to avoid N+1 queries.
func (s *Service) buildVariantMap(ctx context.Context, productIDs []uuid.UUID, shopID uuid.UUID) (map[uuid.UUID][]VariantResponse, error) {
	if len(productIDs) == 0 {
		return map[uuid.UUID][]VariantResponse{}, nil
	}
	rows, err := s.queries.GetVariantsByProductIDs(ctx, db.GetVariantsByProductIDsParams{
		Column1: productIDs,
		ShopID:  shopID,
	})
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID][]VariantResponse, len(productIDs))
	for _, v := range rows {
		m[v.ProductID] = append(m[v.ProductID], mapVariant(v))
	}
	return m, nil
}

// Batch fetch all images for a product set to avoid N+1 queries.
func (s *Service) buildImageMap(ctx context.Context, productIDs []uuid.UUID, shopID uuid.UUID) (map[uuid.UUID][]ImageResponse, error) {
	if len(productIDs) == 0 {
		return map[uuid.UUID][]ImageResponse{}, nil
	}
	rows, err := s.queries.GetImagesByProductIDs(ctx, db.GetImagesByProductIDsParams{
		Column1: productIDs,
		ShopID:  shopID,
	})
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID][]ImageResponse, len(productIDs))
	for _, img := range rows {
		m[img.ProductID] = append(m[img.ProductID], mapImage(img))
	}
	return m, nil
}

func (s *Service) getVariants(ctx context.Context, productID, shopID uuid.UUID) ([]VariantResponse, error) {
	rows, err := s.queries.ListVariantsByProduct(ctx, db.ListVariantsByProductParams{
		ProductID: productID,
		ShopID:    shopID,
	})
	if err != nil {
		return nil, err
	}
	variantResponses := make([]VariantResponse, 0, len(rows))
	for _, v := range rows {
		variantResponses = append(variantResponses, mapVariant(v))
	}
	return variantResponses, nil
}

func (s *Service) getImages(ctx context.Context, productID, shopID uuid.UUID) ([]ImageResponse, error) {
	rows, err := s.queries.ListImagesByProduct(ctx, db.ListImagesByProductParams{
		ProductID: productID,
		ShopID:    shopID,
	})
	if err != nil {
		return nil, err
	}
	imageResponses := make([]ImageResponse, 0, len(rows))
	for _, img := range rows {
		imageResponses = append(imageResponses, mapImage(img))
	}
	return imageResponses, nil
}

func mapProduct(p db.Product, variants []VariantResponse, images []ImageResponse) ProductResponse {
	shopIDStr := uuid.UUID(p.ShopID.Bytes).String()

	var categoryIDPtr *string
	if p.CategoryID.Valid {
		s := uuid.UUID(p.CategoryID.Bytes).String()
		categoryIDPtr = &s
	}

	var descPtr *string
	if p.Description.Valid {
		descPtr = &p.Description.String
	}

	var skuPtr *string
	if p.Sku.Valid {
		skuPtr = &p.Sku.String
	}

	status := "draft"
	if p.Status.Valid {
		status = p.Status.String
	}

	return ProductResponse{
		ID:                p.ID.String(),
		ShopID:            shopIDStr,
		CategoryID:        categoryIDPtr,
		Title:             p.Title,
		Slug:              p.Slug,
		Description:       descPtr,
		Price:             pgutil.NumericToString(p.Price),
		CompareAtPrice:    pgutil.NumericToStringPtr(p.CompareAtPrice),
		TrackInventory:    p.TrackInventory.Bool,
		StockQuantity:     p.StockQuantity,
		LowStockThreshold: p.LowStockThreshold,
		LowStock:          p.TrackInventory.Bool && p.StockQuantity <= p.LowStockThreshold,
		SKU:               skuPtr,
		Status:            status,
		Variants:          variants,
		Images:            images,
		SeoTitle:          p.SeoTitle.String,
		SeoDescription:    p.SeoDescription.String,
		CreatedAt: func() string {
			if p.CreatedAt.Valid {
				return p.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			}
			return ""
		}(),
	}
}

func mapCategory(c db.Category) CategoryResponse {
	shopIDStr := uuid.UUID(c.ShopID.Bytes).String()

	var parentIDPtr *string
	if c.ParentID.Valid {
		s := uuid.UUID(c.ParentID.Bytes).String()
		parentIDPtr = &s
	}

	return CategoryResponse{
		ID:              c.ID.String(),
		ShopID:          shopIDStr,
		ParentID:        parentIDPtr,
		Name:            c.Name,
		Slug:            c.Slug,
		MetaDescription: c.MetaDescription.String,
	}
}

func mapVariant(v db.ProductVariant) VariantResponse {
	var options map[string]interface{}
	if len(v.Options) > 0 {
		_ = json.Unmarshal(v.Options, &options)
	}
	if options == nil {
		options = map[string]interface{}{}
	}

	var skuPtr *string
	if v.Sku.Valid {
		skuPtr = &v.Sku.String
	}

	return VariantResponse{
		ID:             v.ID.String(),
		ProductID:      v.ProductID.String(),
		ShopID:         v.ShopID.String(),
		Title:          v.Title,
		SKU:            skuPtr,
		Price:          pgutil.NumericToString(v.Price),
		CompareAtPrice: pgutil.NumericToStringPtr(v.CompareAtPrice),
		StockQuantity:  v.StockQuantity,
		Options:        options,
		Position:       pgutil.Int4ToInt32(v.Position),
		CreatedAt:      v.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapImage(img db.ProductImage) ImageResponse {
	var variantIDPtr *string
	if img.VariantID.Valid {
		s := uuid.UUID(img.VariantID.Bytes).String()
		variantIDPtr = &s
	}

	var altTextPtr *string
	if img.AltText.Valid {
		altTextPtr = &img.AltText.String
	}

	return ImageResponse{
		ID:        img.ID.String(),
		ProductID: img.ProductID.String(),
		ShopID:    img.ShopID.String(),
		VariantID: variantIDPtr,
		URL:       img.Url,
		AltText:   altTextPtr,
		Position:  pgutil.Int4ToInt32(img.Position),
		IsPrimary: img.IsPrimary.Bool,
		CreatedAt: img.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func lowStockThreshold(v int32) int32 {
	if v <= 0 {
		return 5
	}
	return v
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
