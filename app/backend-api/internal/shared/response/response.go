// Package response provides standard API response helpers for consistent JSON output.
package response

import (
	"errors"
	"net/http"

	"backend-api/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

// OK sends a 200 response with the given data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

// OKWithMeta sends a 200 response with data + pagination/meta info.
func OKWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: meta})
}

// Created sends a 201 response with the created resource.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data})
}

// NoContent sends a 204 response (e.g., successful delete).
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest sends a 400 with an error message.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Envelope{Error: msg})
}

// Unauthorized sends a 401 with an error message.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Envelope{Error: msg})
}

// Forbidden sends a 403 with an error message.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Envelope{Error: msg})
}

// NotFound sends a 404 with an error message.
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Envelope{Error: msg})
}

// Conflict sends a 409 with an error message.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Envelope{Error: msg})
}

// UnprocessableEntity sends a 422 when the request is understood but cannot be processed.
func UnprocessableEntity(c *gin.Context, msg string) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{Error: msg})
}

func TooManyRequests(c *gin.Context, msg string) {
	c.JSON(http.StatusTooManyRequests, Envelope{Error: msg})
}

// InternalError sends a 500 with a generic message (never leak internals).
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Envelope{Error: "An internal error occurred"})
}

// PaginationMeta holds pagination metadata.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

// HandleError maps a domain service error to the appropriate HTTP response.
// It checks against apperr sentinels first; unknown errors become 500.
func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		NotFound(c, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		Conflict(c, err.Error())
	case errors.Is(err, apperr.ErrInvalidUUID):
		BadRequest(c, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		Forbidden(c, err.Error())
	case errors.Is(err, apperr.ErrUnprocessable):
		UnprocessableEntity(c, err.Error())
	default:
		InternalError(c)
	}
}

// CalcTotalPages computes the ceiling-divided number of pages.
func CalcTotalPages(total, perPage int64) int64 {
	if perPage <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}
