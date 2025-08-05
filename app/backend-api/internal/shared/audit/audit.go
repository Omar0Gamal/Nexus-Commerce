package audit

import (
	"context"
	"encoding/json"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
)

// Entry describes a single action to record in the audit log.
type Entry struct {
	ShopID       string
	ActorUserID  string
	ActorName    string
	Action       string // e.g. "product.delete"
	ResourceType string // e.g. "product"
	ResourceID   string
	Changes      map[string]any
}

// Write persists an audit log entry fire-and-forget.
// Errors are silently dropped — audit logging must never block or fail an
// otherwise successful request.
func Write(q *db.Queries, e Entry) {
	shopID, err := uuid.Parse(e.ShopID)
	if err != nil {
		return
	}
	actorID, err := uuid.Parse(e.ActorUserID)
	if err != nil {
		return
	}
	resourceID, err := uuid.Parse(e.ResourceID)
	if err != nil {
		return
	}

	params := db.CreateAuditLogParams{
		ShopID:       pgutil.ToUUID(shopID),
		ActorUserID:  pgutil.ToUUID(actorID),
		ActorName:    pgutil.ToText(e.ActorName),
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   pgutil.ToUUID(resourceID),
		IpAddress:    nil,
	}
	if e.Changes != nil {
		b, _ := json.Marshal(e.Changes)
		params.Changes = b
	} else {
		params.Changes = []byte("{}")
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = q.CreateAuditLog(ctx, params)
	}()
}
