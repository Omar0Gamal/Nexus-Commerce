package ginutil

import (
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ShopUUID parses the "shop_id" claim set by the auth middleware.
// On failure it writes a 400 response and returns false.
func ShopUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop_id")
		return uuid.UUID{}, false
	}
	return id, true
}

// CustomerUUID parses the "user_id" claim set by the auth middleware.
// On failure it writes a 400 response and returns false.
func CustomerUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return uuid.UUID{}, false
	}
	return id, true
}

// ParsePage returns the "page" query parameter as an integer (min 1).
func ParsePage(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if v < 1 {
		v = 1
	}
	return v
}

// ParsePerPage returns the "per_page" query parameter as an integer (min 1).
func ParsePerPage(c *gin.Context, defaultVal int) int {
	v, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultVal)))
	if v < 1 {
		v = defaultVal
	}
	return v
}
