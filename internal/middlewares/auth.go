package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/ilam072/event-calendar/pkg/jwt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

const bearerPrefix = "Bearer "

func Auth(manager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
			log.Logger.Warn().Msg("invalid header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)

		claims, err := manager.ParseToken(token)
		if err != nil {
			log.Logger.Err(err).Msg("failed to parse token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
