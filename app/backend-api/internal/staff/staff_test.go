package staff

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateInviteToken(t *testing.T) {
	tok1, err := generateInviteToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tok1) != 64 {
		t.Fatalf("expected token length 64, got %d", len(tok1))
	}

	decoded, err := hex.DecodeString(tok1)
	if err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("expected 32 decoded bytes, got %d", len(decoded))
	}

	tok2, _ := generateInviteToken()
	if tok1 == tok2 {
		t.Fatal("two consecutive tokens should not be identical")
	}
}

func TestHandler_MissingShopContext(t *testing.T) {
	svc := &Service{}
	h := NewHandler(svc)

	r := gin.New()
	dummyMW := func(c *gin.Context) { c.Next() }
	h.RegisterRoutes(r.Group("/api/v1"), dummyMW, dummyMW)

	// GET /api/v1/staff without shop context
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/staff", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing shop context, got %d", w.Code)
	}

	// POST /api/v1/staff/invite without shop context
	w = httptest.NewRecorder()
	body, _ := json.Marshal(InviteStaffRequest{
		Email:  "test@example.com",
		RoleID: "00000000-0000-0000-0000-000000000001",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/staff/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing shop context, got %d", w.Code)
	}

	// PATCH /api/v1/staff/:id/role without shop context
	w = httptest.NewRecorder()
	body, _ = json.Marshal(UpdateStaffRoleRequest{
		RoleID: "00000000-0000-0000-0000-000000000001",
	})
	req, _ = http.NewRequest(http.MethodPatch, "/api/v1/staff/00000000-0000-0000-0000-000000000002/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing shop context, got %d", w.Code)
	}
}

func TestHandler_InvalidRequestBody(t *testing.T) {
	svc := &Service{}
	h := NewHandler(svc)

	r := gin.New()
	authMW := func(c *gin.Context) {
		c.Set("shop_id", "00000000-0000-0000-0000-000000000001")
		c.Next()
	}
	h.RegisterRoutes(r.Group("/api/v1"), authMW, authMW)

	// POST /api/v1/staff/invite with bad email
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{
		"email":   "not-an-email",
		"role_id": "00000000-0000-0000-0000-000000000002",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/staff/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid email format, got %d", w.Code)
	}

	// PATCH /api/v1/staff/:id/role with missing role_id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/v1/staff/00000000-0000-0000-0000-000000000002/role", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing role_id, got %d", w.Code)
	}
}

func TestHandler_HandleErrorMapping(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		err          error
		expectedCode int
	}{
		{ErrStaffNotFound, http.StatusNotFound},
		{ErrRoleNotFound, http.StatusNotFound},
		{ErrAlreadyStaff, http.StatusConflict},
		{ErrCannotChangeOwnerRole, http.StatusBadRequest},
		{ErrCannotRemoveOwner, http.StatusBadRequest},
		{ErrInvalidUUID, http.StatusBadRequest},
		{ErrInvalidEmail, http.StatusBadRequest},
		{ErrStaffLimit, http.StatusForbidden},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.handleError(c, tc.err)

		if w.Code != tc.expectedCode {
			t.Errorf("for error %v, expected status %d, got %d", tc.err, tc.expectedCode, w.Code)
		}
	}
}
