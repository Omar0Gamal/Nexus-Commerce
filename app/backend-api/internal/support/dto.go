package support

import "github.com/jackc/pgx/v5/pgtype"


type CreateTicketRequest struct {
	Subject        string  `json:"subject"         binding:"required"`
	Body           string  `json:"body"            binding:"required"`
	Category       string  `json:"category"`
	RelatedOrderID *string `json:"related_order_id"`
}

type GuestTicketRequest struct {
	GuestEmail string `json:"guest_email" binding:"required,email"`
	Subject    string `json:"subject"     binding:"required"`
	Body       string `json:"body"        binding:"required"`
	Category   string `json:"category"`
}

type UpdateTicketRequest struct {
	Status   *string `json:"status"`
	Priority *string `json:"priority"`
}

type AddMessageRequest struct {
	Body           string `json:"body"             binding:"required"`
	IsInternalNote bool   `json:"is_internal_note"`
}

type TicketFilter struct {
	Status   string
	Priority string
	StaffID  pgtype.UUID
	Limit    int32
	Offset   int32
}

type CreateKBArticleRequest struct {
	Category    string `json:"category"     binding:"required"`
	Title       string `json:"title"        binding:"required"`
	Slug        string `json:"slug"         binding:"required"`
	Body        string `json:"body"         binding:"required"`
	IsPublished bool   `json:"is_published"`
	Position    int32  `json:"position"`
}

type UpdateKBArticleRequest struct {
	Category    *string `json:"category"`
	Title       *string `json:"title"`
	Slug        *string `json:"slug"`
	Body        *string `json:"body"`
	IsPublished *bool   `json:"is_published"`
	Position    *int32  `json:"position"`
}

type CreatePlatformTicketRequest struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}
