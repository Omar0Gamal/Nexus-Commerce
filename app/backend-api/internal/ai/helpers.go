package ai

import (
	"encoding/json"

	"github.com/google/uuid"
)

// mustParseUUID parses s into uuid.UUID, returning uuid.Nil on failure.
func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// unmarshalJSONB decodes a JSONB []byte into the target interface.
func unmarshalJSONB(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}
