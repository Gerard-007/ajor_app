package handlers

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserByIdHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdStr := c.Param("id")
		userId, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid user ID",
			})
			return
		}
		user, err := services.GetUserByID(db, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to fetch user",
			})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func DeleteUserHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse user ID
		userIDStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if user is admin
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can delete users"})
			return
		}

		// Delete user and profile
		err = services.DeleteUser(db, userID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User and profile deleted successfully"})
	}
}
