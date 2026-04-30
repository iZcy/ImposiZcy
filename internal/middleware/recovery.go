package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.WithField("error", recovered).Error("Panic recovered")
		c.AbortWithStatusJSON(500, gin.H{"success": false, "error": "Internal server error"})
	})
}
