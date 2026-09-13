package audit

import (
	"context"
	"encoding/json"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
)

type Entry struct {
	ShopID       string
	ActorUserID  string
	ActorName    string
	Action       string
	ResourceType string
	ResourceID   string
	Changes      map[string]any
}

// Write persists an audit entry asynchronously so logging failures never block request handling.
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
