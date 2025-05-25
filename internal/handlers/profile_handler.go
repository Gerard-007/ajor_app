package handlers

import (
	"net/http"
	"strconv"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserProfileHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID",
			})
			return
		}
		profile, err := services.GetUserProfile(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user profile",
			})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}


func UpdateUserProfileHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// // Get authenticated user from context (set by AuthMiddleware)
		// authUser, exists := c.Get("user")
		// if !exists {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		// 	return
		// }
		// authUserClaims, ok := authUser.(*utils.Claims)
		// if !ok {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		// 	return
		// }

		// // Ensure user can only update their own profile
		// if authUserClaims.UserID != userID.Hex() {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update another user's profile"})
		// 	return
		// }

		var profileUpdate struct {
			Bio        string `json:"bio"`
			Location   string `json:"location"`
			ProfilePic string `json:"profile_pic"`
		}
		if err := c.ShouldBindJSON(&profileUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		profile := models.Profile{
			UserID:     userID,
			Bio:        profileUpdate.Bio,
			Location:   profileUpdate.Location,
			ProfilePic: profileUpdate.ProfilePic,
		}

		err = services.UpdateUserProfile(db, &profile)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}
