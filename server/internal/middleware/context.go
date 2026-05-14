package middleware

import (
	jwtpkg "fujiantong/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	ContextAuthorID    = "author_id"
	ContextAuthorAppID = "authorizer_appid"
	ContextRoleName    = "role_name"
)

func CurrentUserID(c *gin.Context) uint {
	value, ok := c.Get(ContextUserID)
	if !ok {
		return 0
	}
	if userID, ok := value.(uint); ok {
		return userID
	}
	return 0
}

func CurrentAuthorID(c *gin.Context) uint64 {
	if v, ok := c.Get(ContextAuthorID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func CurrentAuthorAppID(c *gin.Context) string {
	if v, ok := c.Get(ContextAuthorAppID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func CurrentRoleName(c *gin.Context) string {
	if v, ok := c.Get(ContextRoleName); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ApplyClaims 把 JWT Claims 写入 gin.Context
func ApplyClaims(c *gin.Context, claims *jwtpkg.CustomClaims) {
	c.Set(ContextUserID, claims.UserID)
	c.Set(ContextRole, claims.Role)
	c.Set(ContextRoleName, claims.RoleName)
	c.Set(ContextAuthorID, claims.AuthorID)
	c.Set(ContextAuthorAppID, claims.AuthorizerAppID)
}
