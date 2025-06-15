package ginutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-api/internal/shared/ginutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCtxWithKeys(keys map[string]string, query string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	path := "/"
	if query != "" {
		path = "/?" + query
	}
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range keys {
		c.Set(k, v)
	}
	return c
}

func TestShopUUID_Valid(t *testing.T) {
	id := uuid.New()
	c := newCtxWithKeys(map[string]string{"shop_id": id.String()}, "")

	got, ok := ginutil.ShopUUID(c)
	if !ok {
		t.Fatal("ShopUUID: expected ok=true")
	}
	if got != id {
		t.Errorf("ShopUUID: got %v, want %v", got, id)
	}
}

func TestShopUUID_Invalid(t *testing.T) {
	c := newCtxWithKeys(map[string]string{"shop_id": "not-a-uuid"}, "")

	_, ok := ginutil.ShopUUID(c)
	if ok {
		t.Error("ShopUUID: expected ok=false for invalid UUID")
	}
}

func TestShopUUID_Missing(t *testing.T) {
	c := newCtxWithKeys(nil, "")

	_, ok := ginutil.ShopUUID(c)
	if ok {
		t.Error("ShopUUID: expected ok=false when shop_id is missing")
	}
}

func TestCustomerUUID_Valid(t *testing.T) {
	id := uuid.New()
	c := newCtxWithKeys(map[string]string{"user_id": id.String()}, "")

	got, ok := ginutil.CustomerUUID(c)
	if !ok {
		t.Fatal("CustomerUUID: expected ok=true")
	}
	if got != id {
		t.Errorf("CustomerUUID: got %v, want %v", got, id)
	}
}

func TestCustomerUUID_Invalid(t *testing.T) {
	c := newCtxWithKeys(map[string]string{"user_id": "bad"}, "")

	_, ok := ginutil.CustomerUUID(c)
	if ok {
		t.Error("CustomerUUID: expected ok=false for invalid UUID")
	}
}

func TestParsePage_Default(t *testing.T) {
	c := newCtxWithKeys(nil, "")
	if got := ginutil.ParsePage(c); got != 1 {
		t.Errorf("ParsePage default: got %d, want 1", got)
	}
}

func TestParsePage_Valid(t *testing.T) {
	c := newCtxWithKeys(nil, "page=5")
	if got := ginutil.ParsePage(c); got != 5 {
		t.Errorf("ParsePage: got %d, want 5", got)
	}
}

func TestParsePage_ZeroClampsToOne(t *testing.T) {
	c := newCtxWithKeys(nil, "page=0")
	if got := ginutil.ParsePage(c); got != 1 {
		t.Errorf("ParsePage(0): got %d, want 1", got)
	}
}

func TestParsePage_NegativeClampsToOne(t *testing.T) {
	c := newCtxWithKeys(nil, "page=-3")
	if got := ginutil.ParsePage(c); got != 1 {
		t.Errorf("ParsePage(-3): got %d, want 1", got)
	}
}

func TestParsePage_NonNumericClampsToOne(t *testing.T) {
	c := newCtxWithKeys(nil, "page=abc")
	if got := ginutil.ParsePage(c); got != 1 {
		t.Errorf("ParsePage(abc): got %d, want 1", got)
	}
}

func TestParsePerPage_Default(t *testing.T) {
	c := newCtxWithKeys(nil, "")
	if got := ginutil.ParsePerPage(c, 25); got != 25 {
		t.Errorf("ParsePerPage default: got %d, want 25", got)
	}
}

func TestParsePerPage_Valid(t *testing.T) {
	c := newCtxWithKeys(nil, "per_page=50")
	if got := ginutil.ParsePerPage(c, 25); got != 50 {
		t.Errorf("ParsePerPage: got %d, want 50", got)
	}
}

func TestParsePerPage_ZeroUsesDefault(t *testing.T) {
	c := newCtxWithKeys(nil, "per_page=0")
	if got := ginutil.ParsePerPage(c, 20); got != 20 {
		t.Errorf("ParsePerPage(0): got %d, want 20 (default)", got)
	}
}
