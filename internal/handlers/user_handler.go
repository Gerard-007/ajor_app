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

func GetAllUsersHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin
		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin status not found"})
			return
		}
		isAdminBool, ok := isAdmin.(bool)
		if !ok || !isAdminBool {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view all users"})
			return
		}

		users, err := services.GetAllUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func UpdateUserHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse user ID
		userIDStr := c.Param("id")
		id, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get authenticated user info
		authUserIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		authUserID, err := primitive.ObjectIDFromHex(authUserIDStr.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user ID"})
			return
		}
		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin status not found"})
			return
		}
		isAdminBool, ok := isAdmin.(bool)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin status"})
			return
		}

		// Non-admins can only update their own user
		if !isAdminBool && authUserID != id {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this user"})
			return
		}

		// Parse user data from request body
		var userUpdate services.UserUpdate
		if err := c.ShouldBindJSON(&userUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		updatedUser, err := services.UpdateUser(db, id, &userUpdate, isAdminBool)
		if err != nil {
			if err.Error() == "user not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			if err.Error() == "email already exists" || err.Error() == "username already exists" {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, updatedUser)
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