package response

import (
	"errors"
	"net/http"

	"backend-api/internal/shared/apperr"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

func OK(c *gin.Context, payload any) {
	c.JSON(http.StatusOK, Envelope{Data: payload})
}

func OKWithMeta(c *gin.Context, payload any, meta any) {
	c.JSON(http.StatusOK, Envelope{Data: payload, Meta: meta})
}

func Created(c *gin.Context, payload any) {
	c.JSON(http.StatusCreated, Envelope{Data: payload})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Envelope{Error: msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Envelope{Error: msg})
}

func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Envelope{Error: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Envelope{Error: msg})
}

func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Envelope{Error: msg})
}

func UnprocessableEntity(c *gin.Context, msg string) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{Error: msg})
}

func TooManyRequests(c *gin.Context, msg string) {
	c.JSON(http.StatusTooManyRequests, Envelope{Error: msg})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Envelope{Error: "An internal error occurred"})
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

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

func CalcTotalPages(total, perPage int64) int64 {
	if perPage <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}
