package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoverWithLog() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("panic recovered: %v", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	})
}
