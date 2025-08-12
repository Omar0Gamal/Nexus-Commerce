package seo

import (
	"context"
	"fmt"
	"strings"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service handles SEO business logic: audits, redirects, sitemaps.
type Service struct {
	q      *db.Queries
	rdb    *redis.Client
	logger *zap.Logger
}

func NewService(q *db.Queries, rdb *redis.Client, logger *zap.Logger) *Service {
	return &Service{q: q, rdb: rdb, logger: logger}
}


// RunSEOAudit recomputes seo_score for every active product in the shop.
// Returns the number of products scored.
func (s *Service) RunSEOAudit(ctx context.Context, shopID uuid.UUID) (int, error) {
	rows, err := s.q.GetAllProductsForSEOAudit(ctx, pgutil.ToUUID(shopID))
	if err != nil {
		return 0, fmt.Errorf("seo audit: fetch products: %w", err)
	}

	for _, r := range rows {
		score := ScoreProduct(r)
		if err := s.q.UpdateProductSEOScore(ctx, db.UpdateProductSEOScoreParams{
			ID:       r.ID,
			ShopID:   pgutil.ToUUID(shopID),
			SeoScore: score,
		}); err != nil {
			s.logger.Warn("seo audit: update score failed",
				zap.String("product_id", r.ID.String()), zap.Error(err))
		}
	}
	return len(rows), nil
}

// GetAuditReport returns a paginated list of products with SEO scores.
func (s *Service) GetAuditReport(ctx context.Context, shopID uuid.UUID, page, limit int) ([]db.GetProductsForSEOAuditRow, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	rows, err := s.q.GetProductsForSEOAudit(ctx, db.GetProductsForSEOAuditParams{
		ShopID: pgutil.ToUUID(shopID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("seo audit report: %w", err)
	}

	total, err := s.q.CountProductsForSEOAudit(ctx, pgutil.ToUUID(shopID))
	if err != nil {
		return nil, 0, fmt.Errorf("seo audit count: %w", err)
	}

	return rows, total, nil
}

// AutofillSEO populates blank seo_title and seo_description from the product's
// title and description respectively. Does not overwrite existing values.
// Returns the number of products updated.
func (s *Service) AutofillSEO(ctx context.Context, shopID uuid.UUID) (int, error) {
	rows, err := s.q.GetAllProductsForSEOAudit(ctx, pgutil.ToUUID(shopID))
	if err != nil {
		return 0, fmt.Errorf("seo autofill: fetch products: %w", err)
	}

	updated := 0
	for _, r := range rows {
		// Only update products missing at least one SEO field.
		if r.SeoTitle.Valid && r.SeoTitle.String != "" &&
			r.SeoDescription.Valid && r.SeoDescription.String != "" {
			continue
		}

		newTitle := r.SeoTitle
		if !newTitle.Valid || newTitle.String == "" {
			// Trim to 60 chars max
			t := r.Title
			if len([]rune(t)) > 60 {
				t = string([]rune(t)[:57]) + "..."
			}
			newTitle = pgutil.ToText(t)
		}

		newDesc := r.SeoDescription
		if !newDesc.Valid || newDesc.String == "" {
			raw := pgutil.TextToString(r.Description)
			raw = strings.TrimSpace(raw)
			if len([]rune(raw)) > 155 {
				raw = string([]rune(raw)[:152]) + "..."
			}
			newDesc = pgutil.ToText(raw)
		}

		if err := s.q.UpdateProductSEOFields(ctx, db.UpdateProductSEOFieldsParams{
			ID:             r.ID,
			ShopID:         pgutil.ToUUID(shopID),
			SeoTitle:       newTitle,
			SeoDescription: newDesc,
		}); err != nil {
			s.logger.Warn("seo autofill: update failed",
				zap.String("product_id", r.ID.String()), zap.Error(err))
			continue
		}
		updated++
	}

	// Invalidate sitemap after SEO field changes.
	s.InvalidateSitemapCache(ctx, shopID)
	return updated, nil
}


// CreateRedirect creates or replaces a URL redirect for the shop.
func (s *Service) CreateRedirect(ctx context.Context, shopID uuid.UUID, fromPath, toPath string, statusCode int16) (db.UrlRedirect, error) {
	if statusCode == 0 {
		statusCode = 301
	}
	r, err := s.q.CreateRedirect(ctx, db.CreateRedirectParams{
		ShopID:     shopID,
		FromPath:   fromPath,
		ToPath:     toPath,
		StatusCode: statusCode,
	})
	if err != nil {
		return db.UrlRedirect{}, fmt.Errorf("create redirect: %w", err)
	}
	// Invalidate cached redirect map so the new entry is picked up immediately.
	_ = s.rdb.Del(ctx, fmt.Sprintf("redirects:%s", shopID)).Err()
	return r, nil
}

// ListRedirects returns a paginated list of redirects for the shop.
func (s *Service) ListRedirects(ctx context.Context, shopID uuid.UUID, page, limit int) ([]db.UrlRedirect, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	rows, err := s.q.ListRedirects(ctx, db.ListRedirectsParams{
		ShopID: shopID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list redirects: %w", err)
	}

	total, err := s.q.CountRedirects(ctx, shopID)
	if err != nil {
		return nil, 0, fmt.Errorf("count redirects: %w", err)
	}

	return rows, total, nil
}

// DeleteRedirect removes a redirect by ID for the shop.
func (s *Service) DeleteRedirect(ctx context.Context, shopID uuid.UUID, id uuid.UUID) error {
	err := s.q.DeleteRedirect(ctx, db.DeleteRedirectParams{
		ID:     id,
		ShopID: shopID,
	})
	if err != nil {
		return apperr.ErrNotFound
	}
	_ = s.rdb.Del(ctx, fmt.Sprintf("redirects:%s", shopID)).Err()
	return nil
}
