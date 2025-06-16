package handlers

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserTransactionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		transactions, err := services.GetUserTransactions(c.Request.Context(), db, userID, contributionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
			return
		}
		c.JSON(http.StatusOK, transactions)
	}
}
