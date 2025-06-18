package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// rbacCacheTTL is how long we cache a staff member's permissions in Redis.
const rbacCacheTTL = 5 * time.Minute

// RequirePermission returns a Gin middleware that ensures the authenticated
// staff member holds the given permission in their assigned role.
//
// Short-circuits immediately for:
//   - Staff with is_owner=true (JWT claim)
//   - Roles with ["*"] or {"*": true} wildcard permissions
//
// The role's permissions JSONB is cached in Redis under
// "rbac:{shopID}:{staffID}" for rbacCacheTTL to avoid per-request DB lookups.
func (s *Service) RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Owners bypass all permission gates.
		if c.GetBool("is_owner") {
			c.Next()
			return
		}

		staffIDStr := c.GetString("staff_id")
		shopIDStr := c.GetString("shop_id")
		if staffIDStr == "" || shopIDStr == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "staff context required"})
			return
		}

		permsJSON, err := s.loadPermissions(c, staffIDStr, shopIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		if !hasPermission(permsJSON, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "permission_denied",
				"permission": perm,
				"message":    "You do not have permission to perform this action.",
			})
			return
		}

		c.Next()
	}
}

// loadPermissions returns the raw permissions JSON for a staff member,
// using Redis as a short-lived cache.
func (s *Service) loadPermissions(c *gin.Context, staffIDStr, shopIDStr string) ([]byte, error) {
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("rbac:%s:%s", shopIDStr, staffIDStr)

	// Try cache first.
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return cached, nil
	}
	if !isRedisNil(err) {
		// Redis error — fall through to DB to stay available.
	}

	// Cache miss — load from DB.
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid staff context")
	}

	staff, err := s.Queries().GetShopStaff(ctx, staffID)
	if err != nil {
		return nil, fmt.Errorf("staff record not found")
	}
	if !staff.RoleID.Valid {
		return nil, fmt.Errorf("no role assigned to staff member")
	}

	role, err := s.Queries().GetShopRole(ctx, staff.RoleID.Bytes)
	if err != nil {
		return nil, fmt.Errorf("role not found")
	}

	permsJSON := role.Permissions

	// Warm cache, ignore errors.
	_ = s.rdb.Set(ctx, cacheKey, permsJSON, rbacCacheTTL).Err()

	return permsJSON, nil
}

// InvalidateRBACCache removes the cached permissions for a staff member.
// Call this whenever a staff member's role is changed.
func (s *Service) InvalidateRBACCache(ctx context.Context, shopIDStr, staffIDStr string) {
	cacheKey := fmt.Sprintf("rbac:%s:%s", shopIDStr, staffIDStr)
	_ = s.rdb.Del(ctx, cacheKey).Err()
}

// InvalidateRBACCacheForShop removes all cached RBAC entries for a shop.
// Used when a role's permissions are updated (affects all staff with that role).
func (s *Service) InvalidateRBACCacheForShop(ctx context.Context, shopIDStr string) {
	pattern := fmt.Sprintf("rbac:%s:*", shopIDStr)
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	_ = s.rdb.Del(ctx, keys...).Err()
}

// hasPermission checks whether the raw JSONB permissions blob grants perm.
// Supports two formats:
//   - Array:  ["*"]  or  ["products.read", "orders.write", ...]
//   - Object: {"*": true}  or  {"products.read": true, ...}
func hasPermission(permsJSON []byte, perm string) bool {
	// Try array format.
	var arr []string
	if json.Unmarshal(permsJSON, &arr) == nil {
		for _, p := range arr {
			if p == "*" || p == perm {
				return true
			}
		}
		return false
	}

	// Try object format.
	var obj map[string]any
	if json.Unmarshal(permsJSON, &obj) == nil {
		if isTruthy(obj["*"]) {
			return true
		}
		return isTruthy(obj[perm])
	}

	return false
}

// isTruthy returns whether a JSON value represents a truthy permission flag.
func isTruthy(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1"
	case float64:
		return val != 0
	default:
		return false
	}
}

func isRedisNil(err error) bool {
	return err == redis.Nil
}
