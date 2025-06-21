// Package testutil provides reusable helpers for unit-testing HTTP handlers.
package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// The returned recorder can be inspected after the handler runs.
func NewContext(method, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()

	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

// pairs pre-set in the context (simulating middleware-injected values).
func NewContextWithParams(method, path string, body any, contextKeys map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := NewContext(method, path, body)
	for k, v := range contextKeys {
		c.Set(k, v)
	}
	return c, w
}

// DecodeBody unmarshals the recorder body into dst.
func DecodeBody(w *httptest.ResponseRecorder, dst any) error {
	return json.NewDecoder(w.Body).Decode(dst)
}

// MustMarshal serialises v to JSON or panics.
func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// JSONRequest builds an *http.Request with a JSON body.
func JSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}
