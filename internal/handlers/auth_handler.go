package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/Gerard-007/ajor_app/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Invalid registration input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}
		log.Printf("Bound user: %+v", user) // Debug log

		// Register user and get status
		status, err := services.RegisterUser(db, &user, pg)
		if err != nil {
			log.Printf("Registration error: %v", err)
			// Map specific errors to appropriate HTTP status codes
			switch err.Error() {
			case "email already exists", "username already exists", "phone already exists":
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case "username is required", "email is required", "password is required",
				"phone is required and must be at least 11 digits", "phone must contain only digits",
				"BVN is required and must be 11 digits", "BVN must contain only digits":
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user: " + err.Error()})
			}
			return
		}

		// Return success response with verification message
		log.Printf("User registered successfully with email: %s", user.Email)
		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully. Please check your email to verify your account.",
			"status": status,
		})
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

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing verification token"})
			return
		}
		usersCollection := db.Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"verification_token": token}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
			return
		}
		_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"verified": true, "verification_token": ""}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
	}
}
