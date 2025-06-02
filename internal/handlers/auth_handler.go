package handlers

import (
	"net/http"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/Gerard-007/ajor_app/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := services.RegisterUser(db, &user, pg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

func LoginHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		token, err := services.LoginUser(db, user.Email, user.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func LogoutHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization token is required"})
			return
		}
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token to get expiration
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
			return
		}

		// Blacklist token
		blacklistCollection := db.Collection("blacklisted_tokens")
		err = repository.BlacklistToken(blacklistCollection, token, time.Unix(claims.ExpiresAt, 0))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}
