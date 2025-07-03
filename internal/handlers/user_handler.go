package handlers

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"github.com/Gerard-007/ajor_app/internal/models"
)

func GetUserByIdHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid user ID",
			})
			return
		}

		// Get authenticated user ID and admin status
		authUserIDStr, _ := c.Get("userID")
		isAdmin, _ := c.Get("isAdmin")
		authUserID, err := primitive.ObjectIDFromHex(authUserIDStr.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Invalid authenticated user ID",
			})
			return
		}

		// Authorization check: user can only access their own data unless admin
		if !isAdmin.(bool) && authUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"status": "error",
				"error":  "Unauthorized access",
			})
			return
		}

		user, err := services.GetUserByID(db, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  err.Error(),
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

func ChangePasswordHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var req struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if len(req.NewPassword) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 8 characters"})
			return
		}
		var user models.User
		err = db.Collection("users").FindOne(c, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		_, err = db.Collection("users").UpdateOne(c, bson.M{"_id": userID}, bson.M{"$set": bson.M{"password": string(hash)}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}