package support

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

var (
	ErrTicketNotFound  = errors.New("ticket not found")
	ErrArticleNotFound = errors.New("article not found")
	ErrForbidden       = errors.New("access denied")
)

type Service struct {
	queries *db.Queries
	logger  *zap.Logger
}

func NewService(queries *db.Queries, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{queries: queries, logger: logger}
}

func (s *Service) CreateTicket(ctx context.Context, shopID, customerID uuid.UUID, req CreateTicketRequest) (*db.SupportTicket, error) {
	params := db.CreateSupportTicketParams{
		ShopID:     pgtype.UUID{Bytes: shopID, Valid: true},
		CustomerID: pgtype.UUID{Bytes: customerID, Valid: true},
		Subject:    req.Subject,
		Status:     db.NullTicketStatus{TicketStatus: db.TicketStatusOpen, Valid: true},
		Priority:   db.NullTicketPriority{TicketPriority: db.TicketPriorityMedium, Valid: true},
	}
	if req.Category != "" {
		params.Category = pgtype.Text{String: req.Category, Valid: true}
	}
	if req.RelatedOrderID != nil {
		oid, err := uuid.Parse(*req.RelatedOrderID)
		if err == nil {
			params.RelatedOrderID = pgtype.UUID{Bytes: oid, Valid: true}
		}
	}
	ticket, err := s.queries.CreateSupportTicket(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	_, err = s.queries.CreateSupportMessage(ctx, db.CreateSupportMessageParams{
		TicketID:    pgtype.UUID{Bytes: ticket.ID, Valid: true},
		SenderType:  db.MessageSenderCustomer,
		CustomerID:  pgtype.UUID{Bytes: customerID, Valid: true},
		MessageBody: req.Body,
		Attachments: []byte("[]"),
	})
	if err != nil {
		return nil, fmt.Errorf("create first message: %w", err)
	}

	return &ticket, nil
}

func (s *Service) CreateGuestTicket(ctx context.Context, shopID uuid.UUID, req GuestTicketRequest) (*db.SupportTicket, error) {
	params := db.CreateSupportTicketParams{
		ShopID:     pgtype.UUID{Bytes: shopID, Valid: true},
		GuestEmail: pgtype.Text{String: req.GuestEmail, Valid: true},
		Subject:    req.Subject,
		Status:     db.NullTicketStatus{TicketStatus: db.TicketStatusOpen, Valid: true},
		Priority:   db.NullTicketPriority{TicketPriority: db.TicketPriorityMedium, Valid: true},
	}
	if req.Category != "" {
		params.Category = pgtype.Text{String: req.Category, Valid: true}
	}
	ticket, err := s.queries.CreateSupportTicket(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create guest ticket: %w", err)
	}

	_, err = s.queries.CreateSupportMessage(ctx, db.CreateSupportMessageParams{
		TicketID:    pgtype.UUID{Bytes: ticket.ID, Valid: true},
		SenderType:  db.MessageSenderCustomer,
		MessageBody: req.Body,
		Attachments: []byte("[]"),
	})
	if err != nil {
		return nil, fmt.Errorf("create first message: %w", err)
	}

	return &ticket, nil
}

func (s *Service) GetMyTickets(ctx context.Context, shopID, customerID uuid.UUID, limit, offset int32) ([]db.SupportTicket, error) {
	tickets, err := s.queries.ListSupportTicketsByCustomer(ctx, db.ListSupportTicketsByCustomerParams{
		ShopID:     pgtype.UUID{Bytes: shopID, Valid: true},
		CustomerID: pgtype.UUID{Bytes: customerID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return tickets, nil
}

func (s *Service) GetTicket(ctx context.Context, shopID, ticketID, requesterID uuid.UUID, isStaff bool) (*db.SupportTicket, []db.SupportMessage, error) {
	ticket, err := s.queries.GetSupportTicket(ctx, db.GetSupportTicketParams{
		ID:     ticketID,
		ShopID: pgtype.UUID{Bytes: shopID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrTicketNotFound
		}
		return nil, nil, fmt.Errorf("get ticket: %w", err)
	}

	if !isStaff && ticket.CustomerID.Bytes != requesterID {
		return nil, nil, ErrForbidden
	}

	var msgs []db.SupportMessage
	if isStaff {
		msgs, err = s.queries.ListSupportMessages(ctx, pgtype.UUID{Bytes: ticketID, Valid: true})
	} else {
		msgs, err = s.queries.ListPublicSupportMessages(ctx, pgtype.UUID{Bytes: ticketID, Valid: true})
	}
	if err != nil {
		return nil, nil, fmt.Errorf("list messages: %w", err)
	}

	return &ticket, msgs, nil
}

func (s *Service) AddMessage(ctx context.Context, shopID, ticketID, senderID uuid.UUID, senderType string, req AddMessageRequest) (*db.SupportMessage, error) {
	params := db.CreateSupportMessageParams{
		TicketID:       pgtype.UUID{Bytes: ticketID, Valid: true},
		MessageBody:    req.Body,
		IsInternalNote: pgtype.Bool{Bool: req.IsInternalNote, Valid: true},
		Attachments:    []byte("[]"),
	}
	switch senderType {
	case "customer":
		params.SenderType = db.MessageSenderCustomer
		params.CustomerID = pgtype.UUID{Bytes: senderID, Valid: true}
	case "staff":
		params.SenderType = db.MessageSenderStaff
		params.StaffID = pgtype.UUID{Bytes: senderID, Valid: true}
	default:
		params.SenderType = db.MessageSenderSystem
	}

	msg, err := s.queries.CreateSupportMessage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}
	return &msg, nil
}

func (s *Service) ListTickets(ctx context.Context, shopID uuid.UUID, filter TicketFilter) ([]db.SupportTicket, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.StaffID.Valid {
		return s.queries.ListSupportTicketsByStaff(ctx, filter.StaffID)
	}
	if filter.Status != "" {
		return s.queries.ListSupportTicketsByStatus(ctx, db.ListSupportTicketsByStatusParams{
			ShopID: pgtype.UUID{Bytes: shopID, Valid: true},
			Status: db.NullTicketStatus{TicketStatus: db.TicketStatus(filter.Status), Valid: true},
			Limit:  filter.Limit,
			Offset: filter.Offset,
		})
	}
	return s.queries.ListSupportTickets(ctx, db.ListSupportTicketsParams{
		ShopID: pgtype.UUID{Bytes: shopID, Valid: true},
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (s *Service) UpdateTicketStatus(ctx context.Context, shopID, ticketID uuid.UUID, req UpdateTicketRequest) (*db.SupportTicket, error) {
	params := db.UpdateSupportTicketParams{
		ID:     ticketID,
		ShopID: pgtype.UUID{Bytes: shopID, Valid: true},
	}
	if req.Status != nil {
		params.Status = db.NullTicketStatus{TicketStatus: db.TicketStatus(*req.Status), Valid: true}
	}
	if req.Priority != nil {
		params.Priority = db.NullTicketPriority{TicketPriority: db.TicketPriority(*req.Priority), Valid: true}
	}
	ticket, err := s.queries.UpdateSupportTicket(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("update ticket: %w", err)
	}
	return &ticket, nil
}

func (s *Service) AssignTicket(ctx context.Context, shopID, ticketID, staffID uuid.UUID) (*db.SupportTicket, error) {
	ticket, err := s.queries.AssignSupportTicket(ctx, db.AssignSupportTicketParams{
		AssignedStaffID: pgtype.UUID{Bytes: staffID, Valid: true},
		ID:              ticketID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("assign ticket: %w", err)
	}
	return &ticket, nil
}

func (s *Service) ListArticles(ctx context.Context, shopID uuid.UUID, category string, publishedOnly bool, limit, offset int32) ([]db.KbArticle, error) {
	if limit == 0 {
		limit = 20
	}
	params := db.ListKBArticlesParams{
		ShopID: shopID,
		Limit:  limit,
		Offset: offset,
	}
	if category != "" {
		params.Category = pgtype.Text{String: category, Valid: true}
	}
	if publishedOnly {
		params.PublishedOnly = pgtype.Bool{Bool: true, Valid: true}
	}
	return s.queries.ListKBArticles(ctx, params)
}

func (s *Service) GetArticle(ctx context.Context, shopID uuid.UUID, slug string) (*db.KbArticle, error) {
	art, err := s.queries.GetKBArticleBySlug(ctx, db.GetKBArticleBySlugParams{
		ShopID: shopID,
		Slug:   slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrArticleNotFound
		}
		return nil, fmt.Errorf("get article: %w", err)
	}
	// Increment view count asynchronously (fire and forget), but log failures.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.queries.IncrementKBViewCount(bgCtx, art.ID); err != nil {
			s.logger.Warn("failed to increment KB article view count",
				zap.String("article_id", art.ID.String()),
				zap.Error(err),
			)
		}
	}()
	return &art, nil
}

func (s *Service) MarkHelpful(ctx context.Context, articleID uuid.UUID) error {
	return s.queries.IncrementKBHelpfulCount(ctx, articleID)
}

func (s *Service) CreateArticle(ctx context.Context, shopID uuid.UUID, req CreateKBArticleRequest) (*db.KbArticle, error) {
	art, err := s.queries.CreateKBArticle(ctx, db.CreateKBArticleParams{
		ShopID:      shopID,
		Category:    req.Category,
		Title:       req.Title,
		Slug:        req.Slug,
		Body:        req.Body,
		IsPublished: req.IsPublished,
		Position:    req.Position,
	})
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return &art, nil
}

func (s *Service) UpdateArticle(ctx context.Context, shopID, articleID uuid.UUID, req UpdateKBArticleRequest) (*db.KbArticle, error) {
	params := db.UpdateKBArticleParams{
		ID:     articleID,
		ShopID: shopID,
	}
	if req.Category != nil {
		params.Category = pgtype.Text{String: *req.Category, Valid: true}
	}
	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}
	if req.Slug != nil {
		params.Slug = pgtype.Text{String: *req.Slug, Valid: true}
	}
	if req.Body != nil {
		params.Body = pgtype.Text{String: *req.Body, Valid: true}
	}
	if req.IsPublished != nil {
		params.IsPublished = pgtype.Bool{Bool: *req.IsPublished, Valid: true}
	}
	if req.Position != nil {
		params.Position = pgtype.Int4{Int32: *req.Position, Valid: true}
	}
	art, err := s.queries.UpdateKBArticle(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrArticleNotFound
		}
		return nil, fmt.Errorf("update article: %w", err)
	}
	return &art, nil
}

func (s *Service) DeleteArticle(ctx context.Context, shopID, articleID uuid.UUID) error {
	return s.queries.DeleteKBArticle(ctx, db.DeleteKBArticleParams{
		ID:     articleID,
		ShopID: shopID,
	})
}

func (s *Service) CreatePlatformTicket(ctx context.Context, shopID uuid.UUID, req CreatePlatformTicketRequest) (*db.PlatformTicket, error) {
	planName, err := s.queries.GetShopPlanName(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop plan: %w", err)
	}

	isPremium := false
	if planName == "premium" || planName == "Premium" {
		isPremium = true
	}

	ticket, err := s.queries.CreatePlatformTicket(ctx, db.CreatePlatformTicketParams{
		ShopID:   shopID,
		Subject:  req.Subject,
		Status:   "open",
		Priority: isPremium,
	})
	if err != nil {
		return nil, fmt.Errorf("create platform ticket: %w", err)
	}

	_, err = s.queries.CreatePlatformTicketMessage(ctx, db.CreatePlatformTicketMessageParams{
		TicketID:   ticket.ID,
		SenderType: "tenant",
		Message:    req.Message,
	})
	if err != nil {
		return nil, fmt.Errorf("create platform ticket message: %w", err)
	}

	return &ticket, nil
}

func (s *Service) GetPlatformTicketsByShop(ctx context.Context, shopID uuid.UUID) ([]db.PlatformTicket, error) {
	return s.queries.GetPlatformTicketsByShop(ctx, shopID)
}

func (s *Service) GetAllPlatformTickets(ctx context.Context) ([]db.PlatformTicket, error) {
	return s.queries.GetAllPlatformTickets(ctx)
}
