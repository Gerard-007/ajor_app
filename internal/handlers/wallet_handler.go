package handlers

import (
	"fmt"
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetWalletHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		user, err := repository.GetUserByID(db.Collection("users"), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		wallet, err := repository.GetWalletByID(db, user.WalletID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}

		if wallet.VirtualAccountID != "" {
			va, err := pg.GetVirtualAccount(c.Request.Context(), wallet.VirtualAccountID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get virtual account: %v", err)})
				return
			}
			wallet.VirtualAccountNumber = va.AccountNumber
			wallet.VirtualBankName = va.BankName
		}

		c.JSON(http.StatusOK, wallet)
	}
}

func DeleteWalletHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin status not found"})
			return
		}

		user, err := repository.GetUserByID(db.Collection("users"), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		wallet, err := repository.GetWalletByID(db, user.WalletID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}

		// Check if wallet belongs to a contribution
		var contribution models.Contribution
		err = db.Collection("contributions").FindOne(c.Request.Context(), bson.M{"wallet_id": wallet.ID}).Decode(&contribution)
		if err == nil && contribution.GroupAdmin != userID && !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only group admin or system admin can delete contribution wallet"})
			return
		}
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check contribution: %v", err)})
			return
		}

		if wallet.VirtualAccountID != "" {
			if err := pg.DeactivateVirtualAccount(c.Request.Context(), wallet.VirtualAccountID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to deactivate virtual account: %v", err)})
				return
			}
		}

		if err := repository.DeleteWallet(db, wallet.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete wallet"})
			return
		}

		if wallet.Type == models.WalletTypeUser {
			_, err = repository.UpdateUser(db, userID, &repository.UserUpdate{WalletID: primitive.ObjectID{}})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink wallet from user"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
	}
}