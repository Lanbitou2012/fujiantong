package middleware

import (
	"net/http"
	"strings"

	jwtpkg "fujiantong/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

func OptionalJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token != "" {
			if claims, err := jwtpkg.ParseToken(token); err == nil {
				ApplyClaims(c, claims)
			}
		}
		c.Next()
	}
}

func RequireJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		claims, err := jwtpkg.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效"})
			return
		}
		ApplyClaims(c, claims)
		c.Next()
	}
}

func RequireRole(roles ...int8) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, ok := c.Get(ContextRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		role, ok := roleValue.(int8)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "角色无效"})
			return
		}
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权访问"})
	}
}

func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
