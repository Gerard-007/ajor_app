package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserProfileHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		profile, err := services.GetUserProfile(db, userID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user profile"})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

func UpdateUserProfileHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse user ID from URL
		userIDStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Verify authenticated user matches the user ID
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

		// Check if user is admin
		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin status not found"})
			return
		}

		// Non-admins can only update their own profile
		if !isAdmin.(bool) && authUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this profile"})
			return
		}

		// Bind JSON input
		var profileUpdate models.Profile
		if err := c.ShouldBindJSON(&profileUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Set UserID for the update
		profileUpdate.UserID = userID

		// Update profile
		err = services.UpdateUserProfile(db, userID, &profileUpdate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

func UpdateUserProfilePictureHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse user ID from URL
		userIDStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Verify authenticated user matches the user ID
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

		// Check if user is admin
		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin status not found"})
			return
		}

		// Non-admins can only update their own profile picture
		if !isAdmin.(bool) && authUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this profile picture"})
			return
		}

		file, err := c.FormFile("profile_pic")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
			return
		}

		if err := os.MkdirAll(os.Getenv("PROFILE_PIC_DIR"), os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for profile picture"})
			return
		}

		fileName := fmt.Sprintf("%s/%s", os.Getenv("PROFILE_PIC_DIR"), file.Filename)
		filePath := filepath.Join(os.Getenv("PROFILE_PIC_DIR"), fileName)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile picture"})
			return
		}

		err = services.UpdateUserProfilePicture(db, userID, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile picture"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile picture updated successfully"})
	}
}