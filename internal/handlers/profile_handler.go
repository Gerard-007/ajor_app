package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/Gerard-007/ajor_app/internal/services"
)

func GetUserProfileHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID",
			})
			return
		}
		profile, err := services.GetUserProfile(db, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user profile",
			})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}