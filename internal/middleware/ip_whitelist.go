package middleware

import (
	"net/http"
	"strings"

	"github.com/KreaZcy/kzcy-config/env"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type WhitelistProvider interface {
	GetDashboardIPWhitelistNoCtx() []string
	GetAPIIPWhitelistNoCtx() []string
}

func isWhitelisted(ip string, whitelisted []string) bool {
	for _, w := range whitelisted {
		if w == ip || w == "*" {
			return true
		}
	}
	return false
}

func APIWhitelistMiddleware(provider WhitelistProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if env.IsDevelopment() {
			c.Next()
			return
		}

		ip := c.ClientIP()
		apiWhitelist := provider.GetAPIIPWhitelistNoCtx()
		dashboardWhitelist := provider.GetDashboardIPWhitelistNoCtx()

		allWhitelisted := append(apiWhitelist, dashboardWhitelist...)
		if len(allWhitelisted) == 0 {
			c.Next()
			return
		}

		if !isWhitelisted(ip, allWhitelisted) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not whitelisted"})
			return
		}

		c.Next()
	}
}

func DashboardWhitelistMiddleware(provider WhitelistProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if env.IsDevelopment() {
			c.Next()
			return
		}

		ip := c.ClientIP()
		dashboardWhitelist := provider.GetDashboardIPWhitelistNoCtx()

		if len(dashboardWhitelist) == 0 {
			c.Next()
			return
		}

		if !isWhitelisted(ip, dashboardWhitelist) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not whitelisted"})
			return
		}

		c.Next()
	}
}

func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user", claims)
		}

		c.Next()
	}
}
