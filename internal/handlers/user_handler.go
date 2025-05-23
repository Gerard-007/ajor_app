package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/Gerard-007/ajor_app/internal/services"
)

func GetUserByIdHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error": "Invalid user ID",
			})
			return
		}
		user, err := services.GetUserByID(db, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error": "Failed to fetch user",
			})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
