// Package worker provides background job workers.
// This file contains the SupportAutomationWorker that auto-replies
// to support tickets based on keyword rules (Phase 9-D).
package worker

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"backend-api/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// SupportAutomationWorker polls for new open support tickets every 30 s
// and auto-replies based on the shop's active automation rules.
type SupportAutomationWorker struct {
	db     *db.Queries
	logger *zap.Logger
}

func NewSupportAutomationWorker(queries *db.Queries, logger *zap.Logger) *SupportAutomationWorker {
	return &SupportAutomationWorker{db: queries, logger: logger}
}

// Run blocks until ctx is cancelled, polling every 30 s.
func (w *SupportAutomationWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run immediately on start.
	w.processAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processAll(ctx)
		}
	}
}

func (w *SupportAutomationWorker) processAll(ctx context.Context) {
	// List open tickets across all shops — paginated by 100.
	tickets, err := w.db.ListSupportTickets(ctx, db.ListSupportTicketsParams{
		ShopID: pgtype.UUID{Valid: false}, // all shops (SQL must allow NULL shop_id for cross-shop)
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		w.logger.Warn("support automation: list tickets error", zap.Error(err))
		return
	}

	for _, ticket := range tickets {
		if ticket.Status.TicketStatus != db.TicketStatusOpen {
			continue
		}
		if err := w.processTicket(ctx, ticket); err != nil {
			w.logger.Warn("support automation: process ticket error",
				zap.String("ticket_id", ticket.ID.String()),
				zap.Error(err))
		}
	}
}

func (w *SupportAutomationWorker) processTicket(ctx context.Context, ticket db.SupportTicket) error {
	if !ticket.ShopID.Valid {
		return nil
	}

	rules, err := w.db.ListActiveSupportAutomationRules(ctx, ticket.ShopID)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}

	text := strings.ToLower(ticket.Subject)

	for _, rule := range rules {
		if !matchesRule(rule, text) {
			continue
		}

		// Send auto-reply message.
		senderType := db.MessageSenderSystem
		var noUUID pgtype.UUID
		if !rule.ActionAutoReplyText.Valid || strings.TrimSpace(rule.ActionAutoReplyText.String) == "" {
			w.logger.Warn("support automation: matched rule has empty auto-reply text; skipping",
				zap.String("ticket_id", ticket.ID.String()),
				zap.String("rule_id", rule.ID.String()),
			)
			break
		}
		replyBody := rule.ActionAutoReplyText.String

		_, err := w.db.CreateSupportMessage(ctx, db.CreateSupportMessageParams{
			TicketID:       pgtype.UUID{Bytes: ticket.ID, Valid: true},
			SenderType:     senderType,
			StaffID:        noUUID,
			CustomerID:     noUUID,
			MessageBody:    replyBody,
			Attachments:    nil,
			IsInternalNote: pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			w.logger.Warn("support automation: create reply error", zap.Error(err))
		}

		// Break after first matching rule.
		break
	}
	return nil
}

// matchesRule checks if the ticket text matches ANY keyword in the rule's conditions JSON field.
// The Conditions field is a JSONB blob; we parse it as a flat keyword array.
func matchesRule(rule db.SupportAutomationRule, text string) bool {
	// rule.Conditions is a []byte JSONB: ["refund","return","cancel"]
	// Parse keywords from JSON array.
	keywords := parseKeywords(rule.Conditions)
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// parseKeywords parses a JSON string array from a JSONB conditions field.
func parseKeywords(raw []byte) []string {
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
