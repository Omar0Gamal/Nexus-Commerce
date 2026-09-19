package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeShopifyURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{
			input:    "mystore.myshopify.com",
			expected: "https://mystore.myshopify.com/admin/api/2024-01/products.json?limit=250",
		},
		{
			input:    "https://mystore.myshopify.com",
			expected: "https://mystore.myshopify.com/admin/api/2024-01/products.json?limit=250",
		},
		{
			input:    "http://mystore.myshopify.com/",
			expected: "https://mystore.myshopify.com/admin/api/2024-01/products.json?limit=250",
		},
		{
			input:    "quick-shop",
			expected: "https://quick-shop.myshopify.com/admin/api/2024-01/products.json?limit=250",
		},
		{
			input:    "https://example.com/admin/api/2024-01/products.json",
			expected: "https://example.com/admin/api/2024-01/products.json?limit=250",
		},
		{
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		got, err := normalizeShopifyURL(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("normalizeShopifyURL(%q) expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeShopifyURL(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.expected {
			t.Errorf("normalizeShopifyURL(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestParseShopifyNextLink(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{
			header:   `<https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=abc>; rel="next"`,
			expected: "https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=abc",
		},
		{
			header:   `<https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=prev>; rel="previous", <https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=next>; rel="next"`,
			expected: "https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=next",
		},
		{
			header:   `<https://shop.myshopify.com/admin/api/2024-01/products.json?page_info=prev>; rel="previous"`,
			expected: "",
		},
		{
			header:   "",
			expected: "",
		},
	}

	for _, tc := range tests {
		got := parseShopifyNextLink(tc.header)
		if got != tc.expected {
			t.Errorf("parseShopifyNextLink(%q) = %q; want %q", tc.header, got, tc.expected)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Awesome T-Shirt", "awesome-t-shirt"},
		{"Special #1 & Best!", "special-1-best"},
		{"---multiple---dashes---", "multiple-dashes"},
		{"Upper CASE Title", "upper-case-title"},
	}

	for _, tc := range tests {
		got := slugify(tc.input)
		if got != tc.expected {
			t.Errorf("slugify(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestImportShopify_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}

	reqBody := `{"access_token":"shpat_123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/catalog/import/shopify", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("shop_id", "00000000-0000-0000-0000-000000000001")

	h.ImportShopify(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}

	reqBody = `{"store_url":"shop.myshopify.com"}`
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/catalog/import/shopify", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("shop_id", "00000000-0000-0000-0000-000000000001")

	h.ImportShopify(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestImportShopify_ShopifyAPIAuthFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockShopify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"errors": "[API] Invalid API key or access token"})
	}))
	defer mockShopify.Close()

	h := &Handler{}
	reqBody := `{"store_url":"` + mockShopify.URL + `/admin/api/2024-01/products.json","access_token":"bad_token"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/catalog/import/shopify", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("shop_id", "00000000-0000-0000-0000-000000000001")

	h.ImportShopify(c)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway for upstream Shopify auth failure, got %d", w.Code)
	}
}

func TestImportShopify_ShopifyAPINotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockShopify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"errors": "Not Found"})
	}))
	defer mockShopify.Close()

	h := &Handler{}
	reqBody := `{"store_url":"` + mockShopify.URL + `/admin/api/2024-01/products.json","access_token":"token"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/catalog/import/shopify", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("shop_id", "00000000-0000-0000-0000-000000000001")

	h.ImportShopify(c)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway for 404, got %d", w.Code)
	}
}

func TestImportShopify_EmptyProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockShopify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"products": []interface{}{}})
	}))
	defer mockShopify.Close()

	h := &Handler{}
	reqBody := `{"store_url":"` + mockShopify.URL + `/admin/api/2024-01/products.json","access_token":"valid_token"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/catalog/import/shopify", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("shop_id", "00000000-0000-0000-0000-000000000001")

	h.ImportShopify(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for empty products, got %d", w.Code)
	}
}

