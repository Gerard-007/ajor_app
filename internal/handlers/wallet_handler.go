package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func GetUserWalletHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
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

		wallet, err := repository.GetWalletByUserID(db, user.ID)
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

func FundWalletHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
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

		var input struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		err = services.FundWallet(c.Request.Context(), db, userID, input.Amount, pg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fund wallet: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Wallet funding initiated successfully"})
	}
}

func GetContributionWalletHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate userID from context
		userIDStr, exists := c.Get("userID")
		if !exists {
			log.Println("UserID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Extract and validate isAdmin from context
		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			log.Println("isAdmin not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		isAdminBool, ok := isAdmin.(bool)
		if !ok {
			log.Println("Invalid isAdmin type in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Extract and validate contribution ID from URL
		contributionIDStr := c.Param("id")
		if contributionIDStr == "" {
			log.Println("Contribution ID is empty")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Contribution ID is required"})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(contributionIDStr)
		if err != nil {
			log.Printf("Invalid contribution ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contribution ID"})
			return
		}

		// Fetch wallet using service
		wallet, err := services.GetContributionWallet(c.Request.Context(), db, pg, contributionID, userID, isAdminBool)
		if err != nil {
			log.Printf("Failed to fetch contribution wallet: %v", err)
			switch err.Error() {
			case "contribution not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "Contribution not found"})
			case "wallet not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			case "unauthorized access":
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallet: " + err.Error()})
			}
			return
		}

		// Return wallet as JSON
		log.Printf("Successfully fetched wallet for contribution %s", contributionIDStr)
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

		wallet, err := repository.GetWalletByID(db, user.ID)
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

func SimulateFundWalletHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}
		var req struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
			c.JSON(400, gin.H{"error": "Invalid amount"})
			return
		}
		// Find user's wallet
		var wallet models.Wallet
		err = db.Collection("wallets").FindOne(c, bson.M{"owner_id": userID, "type": models.WalletTypeUser}).Decode(&wallet)
		if err != nil {
			c.JSON(404, gin.H{"error": "Wallet not found"})
			return
		}
		// Insert transaction
		tx := models.Transaction{
			FromWallet:     wallet.ID,
			ToWallet:       wallet.ID,
			Amount:         req.Amount,
			Type:           models.TransactionContribution,
			Date:           time.Now(),
			PaymentMethod:  "manual",
			Status:         models.StatusSuccess,
		}
		_, err = db.Collection("transactions").InsertOne(c, tx)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert transaction"})
			return
		}
		// Update wallet balance
		_, err = db.Collection("wallets").UpdateOne(c, bson.M{"_id": wallet.ID}, bson.M{"$inc": bson.M{"balance": req.Amount}})
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update wallet balance"})
			return
		}
		// Create notification for wallet funding
		notification := &models.Notification{
			UserID:    userID,
			Message:   fmt.Sprintf("Your wallet was funded with ₦%.2f", req.Amount),
			Type:      "info",
			Read:      false,
			CreatedAt: time.Now(),
		}
		_, err = db.Collection("notifications").InsertOne(c, notification)
		if err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
		c.JSON(200, gin.H{"message": "Wallet funded (simulated)", "amount": req.Amount})
	}
}