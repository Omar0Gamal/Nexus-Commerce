package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRecorder(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	engine.GET("/", handler)
	engine.ServeHTTP(w, c.Request)
	return w
}

func decodeEnvelope(t *testing.T, w *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env
}

func TestOK(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.OK(c, gin.H{"id": "123"})
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCreated(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.Created(c, gin.H{"id": "abc"})
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestNoContent(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.NoContent(c)
	})
	if w.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestBadRequest(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.BadRequest(c, "bad input")
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
	env := decodeEnvelope(t, w)
	if env.Error != "bad input" {
		t.Errorf("error: got %q, want %q", env.Error, "bad input")
	}
}

func TestNotFound(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.NotFound(c, "not found")
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestConflict(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.Conflict(c, "already exists")
	})
	if w.Code != http.StatusConflict {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestUnprocessableEntity(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.UnprocessableEntity(c, "invalid data")
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestInternalError(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.InternalError(c)
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleError_NotFound(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.HandleError(c, apperr.ErrNotFound)
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleError_InvalidUUID(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.HandleError(c, apperr.ErrInvalidUUID)
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleError_Conflict(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.HandleError(c, apperr.ErrConflict)
	})
	if w.Code != http.StatusConflict {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestHandleError_Forbidden(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.HandleError(c, apperr.ErrForbidden)
	})
	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleError_Unknown(t *testing.T) {
	w := newRecorder(func(c *gin.Context) {
		response.HandleError(c, errors.New("some internal error"))
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestCalcTotalPages(t *testing.T) {
	tests := []struct {
		total   int64
		perPage int64
		want    int64
	}{
		{0, 10, 0},
		{10, 10, 1},
		{11, 10, 2},
		{20, 10, 2},
		{21, 10, 3},
		{1, 1, 1},
		{100, 7, 15},
		{0, 0, 0},   // zero perPage
		{10, 0, 0},  // zero perPage
		{10, -1, 0}, // negative perPage
	}
	for _, tc := range tests {
		got := response.CalcTotalPages(tc.total, tc.perPage)
		if got != tc.want {
			t.Errorf("CalcTotalPages(%d, %d) = %d, want %d", tc.total, tc.perPage, got, tc.want)
		}
	}
}
