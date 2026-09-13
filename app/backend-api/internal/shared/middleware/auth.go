package middleware

import (
	"strings"

	"backend-api/internal/shared/response"
	"backend-api/internal/shared/token"

	"github.com/gin-gonic/gin"
)

func RequireAuth(tokenSvc *token.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := extractAndValidate(c, tokenSvc)
		if !ok {
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("full_name", claims.FullName)
		c.Set("actor_type", claims.ActorType)
		c.Set("shop_id_from_token", claims.ShopID)
		c.Set("staff_id", claims.StaffID)
		c.Set("role", claims.Role)
		c.Set("is_owner", claims.IsOwner)

		c.Next()
	}
}

func OptionalAuth(tokenSvc *token.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.ValidateToken(tokenStr)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("full_name", claims.FullName)
		c.Set("actor_type", claims.ActorType)
		c.Set("shop_id_from_token", claims.ShopID)
		c.Set("staff_id", claims.StaffID)
		c.Set("role", claims.Role)
		c.Set("is_owner", claims.IsOwner)

		c.Next()
	}
}

func RequireStaff() gin.HandlerFunc {
	return requireActorType(token.ActorStaff)
}

func RequireCustomer() gin.HandlerFunc {
	return requireActorType(token.ActorCustomer)
}

func RequirePlatformAdmin() gin.HandlerFunc {
	return requireActorType(token.ActorPlatformAdmin)
}

func requireActorType(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorType := c.GetString("actor_type")
		if actorType != expected {
			response.Forbidden(c, "access denied: requires "+expected+" role")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractAndValidate(c *gin.Context, tokenSvc *token.Service) (*token.Claims, bool) {
	header := c.GetHeader("Authorization")
	if header == "" {
		response.Unauthorized(c, "authorization header required")
		c.Abort()
		return nil, false
	}

	if !strings.HasPrefix(header, "Bearer ") {
		response.Unauthorized(c, "invalid authorization format, expected 'Bearer <token>'")
		c.Abort()
		return nil, false
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	claims, err := tokenSvc.ValidateToken(tokenStr)
	if err != nil {
		response.Unauthorized(c, "invalid or expired token")
		c.Abort()
		return nil, false
	}

	return claims, true
}
