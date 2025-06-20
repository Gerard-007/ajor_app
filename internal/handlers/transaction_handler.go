package handlers

import (
	"fmt"
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// func GetUserTransactionsHandler(db *mongo.Database) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userID, err := getAuthUserID(c)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 			return
// 		}
// 		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
// 			return
// 		}
// 		transactions, err := services.GetUserTransactions(c.Request.Context(), db, userID, contributionID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, transactions)
// 	}
// }

func GetUserTransactionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		isAdmin, _ := c.Get("isAdmin")
		isAdminBool := isAdmin.(bool)

		transactions, err := services.GetUserTransactions(c.Request.Context(), db, userID, isAdminBool)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch transactions: %v", err)})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}

func GetContributionTransactionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		isAdmin, _ := c.Get("isAdmin")
		isAdminBool := isAdmin.(bool)

		contributionIDStr := c.Param("id")
		contributionID, err := primitive.ObjectIDFromHex(contributionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contribution ID"})
			return
		}

		transactions, err := services.GetContributionTransactions(c.Request.Context(), db, contributionID, userID, isAdminBool)
		if err != nil {
			switch err.Error() {
			case "contribution not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "Contribution not found"})
			case "unauthorized access":
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch transactions: %v", err)})
			}
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}
