package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// rbacCacheTTL is how long we cache a staff member's permissions in Redis.
const rbacCacheTTL = 5 * time.Minute

// RequirePermission validates staff role permissions with Redis caching (is_owner bypasses all gates).
func (s *Service) RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
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

func (s *Service) loadPermissions(c *gin.Context, staffIDStr, shopIDStr string) ([]byte, error) {
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("rbac:%s:%s", shopIDStr, staffIDStr)

	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return cached, nil
	}

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
	_ = s.rdb.Set(ctx, cacheKey, permsJSON, rbacCacheTTL).Err()

	return permsJSON, nil
}

func (s *Service) InvalidateRBACCache(ctx context.Context, shopIDStr, staffIDStr string) {
	cacheKey := fmt.Sprintf("rbac:%s:%s", shopIDStr, staffIDStr)
	_ = s.rdb.Del(ctx, cacheKey).Err()
}

func (s *Service) InvalidateRBACCacheForShop(ctx context.Context, shopIDStr string) {
	pattern := fmt.Sprintf("rbac:%s:*", shopIDStr)
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	_ = s.rdb.Del(ctx, keys...).Err()
}

func hasPermission(permsJSON []byte, perm string) bool {
	nsWild := ""
	if dot := strings.LastIndex(perm, "."); dot != -1 {
		nsWild = perm[:dot] + ".*"
	}

	var arr []string
	if json.Unmarshal(permsJSON, &arr) == nil {
		for _, p := range arr {
			if p == "*" || p == perm {
				return true
			}
			if nsWild != "" && p == nsWild {
				return true
			}
		}
		return false
	}

	var obj map[string]any
	if json.Unmarshal(permsJSON, &obj) == nil {
		if isTruthy(obj["*"]) {
			return true
		}
		if isTruthy(obj[perm]) {
			return true
		}
		if nsWild != "" && isTruthy(obj[nsWild]) {
			return true
		}
	}

	return false
}

func isTruthy(v any) bool {
	switch typedVal := v.(type) {
	case bool:
		return typedVal
	case string:
		return typedVal == "true" || typedVal == "1"
	case float64:
		return typedVal != 0
	default:
		return false
	}
}

func isRedisNil(err error) bool {
	return err == redis.Nil
}
