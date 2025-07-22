package reviews

import (
	"context"
	"errors"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound        = apperr.ErrNotFound
	ErrAlreadyReviewed = errors.New("you have already reviewed this product")
	ErrInvalidUUID     = apperr.ErrInvalidUUID
)

// Service handles product review business logic.
type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

// Submit creates a new review. Auto-sets is_verified_purchase if the customer
// completed an order containing the product. Respects pending/approved default
// based on the provided moderationEnabled flag (from shop settings).
func (s *Service) Submit(ctx context.Context, shopID string, req SubmitRequest, moderationEnabled bool) (*ReviewDTO, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	// Check verified purchase
	purchased, err := s.q.CustomerHasPurchasedProduct(ctx, db.CustomerHasPurchasedProductParams{
		ShopID:     shopUUID,
		CustomerID: pgtype.UUID{Bytes: customerUUID, Valid: true},
		ProductID:  pgtype.UUID{Bytes: productUUID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("check purchase: %w", err)
	}

	status := "approved"
	if moderationEnabled {
		status = "pending"
	}

	var orderUUID pgtype.UUID
	if req.OrderID != nil {
		oid, err := uuid.Parse(*req.OrderID)
		if err != nil {
			return nil, ErrInvalidUUID
		}
		orderUUID = pgtype.UUID{Bytes: oid, Valid: true}
	}

	var title pgtype.Text
	if req.Title != nil {
		title = pgtype.Text{String: *req.Title, Valid: true}
	}
	var body pgtype.Text
	if req.Body != nil {
		body = pgtype.Text{String: *req.Body, Valid: true}
	}

	review, err := s.q.CreateReview(ctx, db.CreateReviewParams{
		ShopID:             shopUUID,
		ProductID:          productUUID,
		CustomerID:         customerUUID,
		OrderID:            orderUUID,
		Rating:             req.Rating,
		Title:              title,
		Body:               body,
		Status:             status,
		IsVerifiedPurchase: purchased,
	})
	if err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	dto := reviewToDTO(review, "", "")
	return &dto, nil
}

// List returns approved reviews for a product, paginated.
func (s *Service) List(ctx context.Context, shopID, productID string, limit, offset int32) ([]ReviewDTO, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.q.GetReviewsByProduct(ctx, db.GetReviewsByProductParams{
		ShopID:    shopUUID,
		ProductID: productUUID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}

	dtos := make([]ReviewDTO, 0, len(rows))
	for _, r := range rows {
		dto := ReviewDTO{
			ID:                 r.ID.String(),
			ProductID:          r.ProductID.String(),
			CustomerID:         r.CustomerID.String(),
			Rating:             r.Rating,
			Status:             r.Status,
			IsVerifiedPurchase: r.IsVerifiedPurchase,
			HelpfulCount:       r.HelpfulCount,
			CreatedAt:          r.CreatedAt.Time,
		}
		if r.Title.Valid {
			dto.Title = &r.Title.String
		}
		if r.Body.Valid {
			dto.Body = &r.Body.String
		}
		if r.FirstName.Valid {
			dto.CustomerName = r.FirstName.String
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// Summary returns aggregate rating data for a product.
func (s *Service) Summary(ctx context.Context, shopID, productID string) (*ReviewSummary, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	row, err := s.q.GetReviewSummary(ctx, db.GetReviewSummaryParams{
		ShopID:    shopUUID,
		ProductID: productUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("review summary: %w", err)
	}

	avgFloat := float64(0)
	if row.AvgRating.Valid {
		f, _ := row.AvgRating.Float64Value()
		avgFloat = f.Float64
	}

	return &ReviewSummary{
		TotalReviews: int(row.TotalReviews),
		AvgRating:    avgFloat,
		Distribution: map[int]int{
			5: int(row.FiveStar),
			4: int(row.FourStar),
			3: int(row.ThreeStar),
			2: int(row.TwoStar),
			1: int(row.OneStar),
		},
	}, nil
}

// ListPending returns reviews awaiting moderation for a shop.
func (s *Service) ListPending(ctx context.Context, shopID string, limit, offset int32) ([]PendingReviewDTO, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.q.GetPendingReviews(ctx, db.GetPendingReviewsParams{
		ShopID: shopUUID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list pending reviews: %w", err)
	}

	dtos := make([]PendingReviewDTO, 0, len(rows))
	for _, r := range rows {
		dto := PendingReviewDTO{
			ID:           r.ID.String(),
			ProductID:    r.ProductID.String(),
			ProductTitle: r.ProductTitle,
			CustomerID:   r.CustomerID.String(),
			Rating:       r.Rating,
			CreatedAt:    r.CreatedAt.Time,
		}
		if r.FirstName.Valid {
			dto.CustomerName = r.FirstName.String
		}
		if r.Title.Valid {
			dto.Title = &r.Title.String
		}
		if r.Body.Valid {
			dto.Body = &r.Body.String
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// Moderate approves, rejects, or flags a review. Syncs rating cache on approval.
func (s *Service) Moderate(ctx context.Context, shopID, reviewID, status string) (*ReviewDTO, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	reviewUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	if status != "approved" && status != "rejected" {
		return nil, errors.New("status must be approved or rejected")
	}

	updated, err := s.q.UpdateReviewStatus(ctx, db.UpdateReviewStatusParams{
		ShopID: shopUUID,
		ID:     reviewUUID,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("moderate review: %w", err)
	}

	// Keep product rating cache in sync whenever a review status changes.
	_ = s.q.SyncProductRatingCache(ctx, db.SyncProductRatingCacheParams{
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
		ID:     updated.ProductID,
	})

	dto := reviewToDTO(updated, "", "")
	return &dto, nil
}

// MarkHelpful increments the helpful_count for a review.
func (s *Service) MarkHelpful(ctx context.Context, shopID, reviewID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	reviewUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return ErrInvalidUUID
	}

	return s.q.IncrementHelpfulCount(ctx, db.IncrementHelpfulCountParams{
		ShopID: shopUUID,
		ID:     reviewUUID,
	})
}


func reviewToDTO(r db.ProductReview, firstName, lastName string) ReviewDTO {
	dto := ReviewDTO{
		ID:                 r.ID.String(),
		ProductID:          r.ProductID.String(),
		CustomerID:         r.CustomerID.String(),
		Rating:             r.Rating,
		Status:             r.Status,
		IsVerifiedPurchase: r.IsVerifiedPurchase,
		HelpfulCount:       r.HelpfulCount,
		CreatedAt:          r.CreatedAt.Time,
	}
	if r.Title.Valid {
		dto.Title = &r.Title.String
	}
	if r.Body.Valid {
		dto.Body = &r.Body.String
	}
	if firstName != "" {
		dto.CustomerName = firstName
	}
	return dto
}
