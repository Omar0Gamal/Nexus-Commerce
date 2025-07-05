package storage_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-api/internal/storage"
)

// newTestServer creates a minimal S3-compatible mock server that accepts
// PutObject and DeleteObject requests and records them.
type recordedRequest struct {
	method string
	path   string
}

func newMockS3Server(t *testing.T) (*httptest.Server, *[]recordedRequest) {
	t.Helper()
	var reqs []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs = append(reqs, recordedRequest{method: r.Method, path: r.URL.Path})
		// S3 PutObject returns 200 with an ETag header.
		w.Header().Set("ETag", `"abc123"`)
		w.WriteHeader(http.StatusOK)
	}))
	return srv, &reqs
}

// newClientWithEndpoint builds a storage client pointing at a custom endpoint
// so tests don't need real R2 credentials.
func newClientWithEndpoint(t *testing.T, endpoint, publicURL string) *storage.Client {
	t.Helper()
	c, err := storage.NewWithEndpoint(endpoint, "access", "secret", "test-bucket", publicURL)
	if err != nil {
		t.Fatalf("storage.NewWithEndpoint: %v", err)
	}
	return c
}

func TestPublicURL(t *testing.T) {
	c, err := storage.New("acct123", "key", "secret", "my-bucket", "https://assets.example.com")
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	got := c.PublicURL("shop1/prod1/img.webp")
	want := "https://assets.example.com/shop1/prod1/img.webp"
	if got != want {
		t.Errorf("PublicURL = %q, want %q", got, want)
	}
}

func TestPublicURL_LeadingSlash(t *testing.T) {
	c, _ := storage.New("acct", "k", "s", "b", "https://cdn.example.com")
	got := c.PublicURL("/shop/img.jpg")
	if !strings.HasPrefix(got, "https://cdn.example.com/") {
		t.Errorf("unexpected URL: %q", got)
	}
	if strings.Contains(got, "//shop") {
		t.Errorf("double slash in URL: %q", got)
	}
}

func TestKeyFromURL_Valid(t *testing.T) {
	c, _ := storage.New("acct", "k", "s", "b", "https://assets.example.com")
	key := c.KeyFromURL("https://assets.example.com/shop1/prod1/img.webp")
	want := "shop1/prod1/img.webp"
	if key != want {
		t.Errorf("KeyFromURL = %q, want %q", key, want)
	}
}

func TestKeyFromURL_Invalid(t *testing.T) {
	c, _ := storage.New("acct", "k", "s", "b", "https://assets.example.com")
	key := c.KeyFromURL("https://other.cdn.com/shop1/img.webp")
	if key != "" {
		t.Errorf("KeyFromURL from wrong domain = %q, want empty", key)
	}
}

func TestNew_MissingFields(t *testing.T) {
	cases := []struct {
		name, accountID, key, secret, bucket string
	}{
		{"missing accountID", "", "k", "s", "b"},
		{"missing key", "acct", "", "s", "b"},
		{"missing secret", "acct", "k", "", "b"},
		{"missing bucket", "acct", "k", "s", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := storage.New(tc.accountID, tc.key, tc.secret, tc.bucket, "https://cdn.example.com")
			if err == nil {
				t.Error("expected error for missing field")
			}
		})
	}
}

func TestUpload_CallsPutObject(t *testing.T) {
	srv, reqs := newMockS3Server(t)
	defer srv.Close()

	c := newClientWithEndpoint(t, srv.URL, "https://assets.example.com")
	url, err := c.Upload(context.Background(), "shop/prod/img.webp", "image/webp", strings.NewReader("fake-image-data"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if !strings.HasSuffix(url, "shop/prod/img.webp") {
		t.Errorf("unexpected URL: %q", url)
	}
	if len(*reqs) == 0 {
		t.Fatal("expected at least one request to mock server")
	}
	if (*reqs)[0].method != http.MethodPut {
		t.Errorf("expected PUT, got %s", (*reqs)[0].method)
	}
}

func TestDelete_CallsDeleteObject(t *testing.T) {
	srv, reqs := newMockS3Server(t)
	defer srv.Close()

	c := newClientWithEndpoint(t, srv.URL, "https://assets.example.com")
	err := c.Delete(context.Background(), "shop/prod/img.webp")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	found := false
	for _, r := range *reqs {
		if r.method == http.MethodDelete {
			found = true
		}
	}
	if !found {
		t.Error("expected DELETE request to mock server")
	}
}

// Ensure Upload returns the io.Reader content properly (no content truncation).
func TestUpload_ReadsBody(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			b, _ := io.ReadAll(r.Body)
			received = b
		}
		w.Header().Set("ETag", `"abc"`)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClientWithEndpoint(t, srv.URL, "https://cdn.example.com")
	content := "hello world image bytes"
	_, err := c.Upload(context.Background(), "key.jpg", "image/jpeg", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if string(received) != content {
		t.Errorf("body received = %q, want %q", received, content)
	}
}
