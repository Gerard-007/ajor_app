package auth

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthMiddleware(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from the request header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token is required",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Check if token is blacklisted
		blacklistCollection := db.Collection("blacklisted_tokens")
		isBlacklisted, err := repository.IsTokenBlacklisted(blacklistCollection, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
			return
		}
		if isBlacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid"})
			return
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// If valid, proceed to the next handler
		c.Set("userID", claims.UserID)
		c.Set("isAdmin", claims.IsAdmin)
		c.Next()
	}
}