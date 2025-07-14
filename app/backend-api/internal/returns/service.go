package returns

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"backend-api/internal/db"
	"backend-api/internal/paymob"
	"backend-api/internal/shared/apperr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound        = apperr.ErrNotFound
	ErrInvalidUUID     = apperr.ErrInvalidUUID
	ErrInvalidStatus   = errors.New("invalid return status: must be approved, rejected, or refunded")
	ErrAlreadyRefunded = errors.New("return has already been refunded")
	ErrBadTransition   = errors.New("invalid status transition")
)

// validStatuses maps every user-settable status for a return request.
var validStatuses = map[string]bool{
	"approved": true,
	"rejected": true,
}

// Service handles return request business logic.
type Service struct {
	q      *db.Queries
	paymob *paymob.Client
}

func NewService(q *db.Queries, paymobClient *paymob.Client) *Service {
	return &Service{q: q, paymob: paymobClient}
}


// CreateReturn opens a new return request for an order.
func (s *Service) CreateReturn(ctx context.Context, shopID string, req CreateReturnRequest) (*ReturnResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	params := db.CreateReturnParams{
		ShopID:  shopUUID,
		OrderID: orderUUID,
		Reason:  strings.TrimSpace(req.Reason),
	}
	if req.CustomerID != "" {
		custUUID, err := uuid.Parse(req.CustomerID)
		if err != nil {
			return nil, ErrInvalidUUID
		}
		params.CustomerID = pgtype.UUID{Bytes: custUUID, Valid: true}
	}

	r, err := s.q.CreateReturn(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create return: %w", err)
	}
	resp := mapReturn(r)
	return &resp, nil
}

// GetReturn retrieves a single return request by ID.
func (s *Service) GetReturn(ctx context.Context, shopID, returnID string) (*ReturnResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	retUUID, err := uuid.Parse(returnID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	r, err := s.q.GetReturn(ctx, db.GetReturnParams{ID: retUUID, ShopID: shopUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get return: %w", err)
	}
	resp := mapReturn(r)
	return &resp, nil
}

// ListReturns returns all return requests for a shop.
func (s *Service) ListReturns(ctx context.Context, shopID string, page, perPage int) ([]ReturnResponse, int64, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, 0, ErrInvalidUUID
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	rows, err := s.q.ListReturns(ctx, db.ListReturnsParams{
		ShopID: shopUUID,
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list returns: %w", err)
	}
	total, _ := s.q.CountReturns(ctx, shopUUID)

	result := make([]ReturnResponse, len(rows))
	for i, r := range rows {
		result[i] = mapReturn(r)
	}
	return result, total, nil
}

// ListByOrder returns all return requests for a specific order.
func (s *Service) ListByOrder(ctx context.Context, shopID, orderID string) ([]ReturnResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.q.ListReturnsByOrder(ctx, db.ListReturnsByOrderParams{
		OrderID: orderUUID,
		ShopID:  shopUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list returns by order: %w", err)
	}
	result := make([]ReturnResponse, len(rows))
	for i, r := range rows {
		result[i] = mapReturn(r)
	}
	return result, nil
}

// UpdateStatus transitions a return to approved or rejected.
func (s *Service) UpdateStatus(ctx context.Context, shopID, returnID string, req UpdateReturnRequest) (*ReturnResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	retUUID, err := uuid.Parse(returnID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	status := strings.ToLower(req.Status)
	if !validStatuses[status] {
		return nil, ErrInvalidStatus
	}

	// Guard: cannot change status if already refunded
	existing, err := s.q.GetReturn(ctx, db.GetReturnParams{ID: retUUID, ShopID: shopUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get return: %w", err)
	}
	if existing.Status == db.ReturnStatusRefunded {
		return nil, ErrAlreadyRefunded
	}

	params := db.UpdateReturnStatusParams{
		ID:     retUUID,
		ShopID: shopUUID,
		Status: db.ReturnStatus(status),
	}
	if req.Notes != nil {
		params.Notes = pgtype.Text{String: *req.Notes, Valid: true}
	}

	r, err := s.q.UpdateReturnStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update return status: %w", err)
	}
	resp := mapReturn(r)
	return &resp, nil
}

// ProcessRefund issues a Paymob refund for an approved return request.
// It calls the Paymob API (if configured) and then marks the return as refunded.
func (s *Service) ProcessRefund(ctx context.Context, shopID, returnID string, req ProcessRefundRequest) (*ReturnResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	retUUID, err := uuid.Parse(returnID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	existing, err := s.q.GetReturn(ctx, db.GetReturnParams{ID: retUUID, ShopID: shopUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get return: %w", err)
	}
	if existing.Status == db.ReturnStatusRefunded {
		return nil, ErrAlreadyRefunded
	}
	if existing.Status != db.ReturnStatusApproved {
		return nil, fmt.Errorf("%w: return must be approved before refunding", ErrBadTransition)
	}

	// Determine refund amount
	var amountCents int64
	if req.RefundAmount != nil {
		amountCents = int64(math.Round(*req.RefundAmount * 100))
	} else if req.AmountCents > 0 {
		amountCents = req.AmountCents
	}

	// Attempt Paymob refund (optional — if no client, mark refunded directly)
	var paymobRefundID int64
	if s.paymob != nil && amountCents > 0 {
		id, err := s.paymob.Refund(existing.OrderID.String(), amountCents)
		if err != nil {
			return nil, fmt.Errorf("paymob refund: %w", err)
		}
		paymobRefundID = id
	}

	params := db.SetReturnRefundedParams{
		ID:     retUUID,
		ShopID: shopUUID,
	}
	if paymobRefundID > 0 {
		params.PaymobRefundID = pgtype.Int8{Int64: paymobRefundID, Valid: true}
	}
	if amountCents > 0 {
		params.RefundAmount = pgtype.Numeric{Int: big.NewInt(amountCents), Exp: -2, Valid: true}
	}

	r, err := s.q.SetReturnRefunded(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("set return refunded: %w", err)
	}
	resp := mapReturn(r)
	return &resp, nil
}


func mapReturn(r db.OrderReturn) ReturnResponse {
	resp := ReturnResponse{
		ID:        r.ID.String(),
		ShopID:    r.ShopID.String(),
		OrderID:   r.OrderID.String(),
		Reason:    r.Reason,
		Status:    string(r.Status),
		CreatedAt: "",
		UpdatedAt: "",
	}
	if r.CustomerID.Valid {
		s := uuid.UUID(r.CustomerID.Bytes).String()
		resp.CustomerID = &s
	}
	if r.PaymobRefundID.Valid {
		resp.PaymobRefundID = &r.PaymobRefundID.Int64
	}
	if r.RefundAmount.Valid {
		v, err := r.RefundAmount.Float64Value()
		if err == nil && v.Valid {
			s := fmt.Sprintf("%.2f", v.Float64)
			resp.RefundAmount = &s
		}
	}
	if r.Notes.Valid && r.Notes.String != "" {
		resp.Notes = &r.Notes.String
	}
	if r.CreatedAt.Valid {
		resp.CreatedAt = r.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	if r.UpdatedAt.Valid {
		resp.UpdatedAt = r.UpdatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return resp
}
