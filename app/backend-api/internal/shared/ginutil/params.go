package ginutil

import (
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ShopUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop_id")
		return uuid.UUID{}, false
	}
	return id, true
}

func CustomerUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return uuid.UUID{}, false
	}
	return id, true
}

func ParsePage(c *gin.Context) int {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	return page
}

func ParsePerPage(c *gin.Context, defaultPerPage int) int {
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage)))
	if perPage < 1 {
		perPage = defaultPerPage
	}
	return perPage
}
