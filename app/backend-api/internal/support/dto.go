package support

import "github.com/jackc/pgx/v5/pgtype"


// CreateTicketRequest is used by authenticated customers.
type CreateTicketRequest struct {
	Subject        string  `json:"subject"         binding:"required"`
	Body           string  `json:"body"            binding:"required"`
	Category       string  `json:"category"`
	RelatedOrderID *string `json:"related_order_id"`
}

// GuestTicketRequest is used by unauthenticated visitors.
type GuestTicketRequest struct {
	GuestEmail string `json:"guest_email" binding:"required,email"`
	Subject    string `json:"subject"     binding:"required"`
	Body       string `json:"body"        binding:"required"`
	Category   string `json:"category"`
}

// UpdateTicketRequest is used by staff to patch a ticket.
type UpdateTicketRequest struct {
	Status   *string `json:"status"`
	Priority *string `json:"priority"`
}

// AddMessageRequest is used to add a reply to a ticket.
type AddMessageRequest struct {
	Body           string `json:"body"             binding:"required"`
	IsInternalNote bool   `json:"is_internal_note"`
}

// TicketFilter is used by staff to filter tickets.
type TicketFilter struct {
	Status   string
	Priority string
	StaffID  pgtype.UUID
	Limit    int32
	Offset   int32
}


// CreateKBArticleRequest is used by staff to create an article.
type CreateKBArticleRequest struct {
	Category    string `json:"category"     binding:"required"`
	Title       string `json:"title"        binding:"required"`
	Slug        string `json:"slug"         binding:"required"`
	Body        string `json:"body"         binding:"required"`
	IsPublished bool   `json:"is_published"`
	Position    int32  `json:"position"`
}

// UpdateKBArticleRequest is used by staff to patch an article.
type UpdateKBArticleRequest struct {
	Category    *string `json:"category"`
	Title       *string `json:"title"`
	Slug        *string `json:"slug"`
	Body        *string `json:"body"`
	IsPublished *bool   `json:"is_published"`
	Position    *int32  `json:"position"`
}
