package handlers

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserNotificationsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		notifications, err := services.GetUserNotifications(c.Request.Context(), db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
			return
		}
		c.JSON(http.StatusOK, notifications)
	}
}