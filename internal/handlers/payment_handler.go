package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/Gerard-007/ajor_app/internal/services"
)

// FlutterwaveWebhookHandler handles payment notifications from Flutterwave
func FlutterwaveWebhookHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Signature verification
		secret := os.Getenv("FLW_SECRET_KEY")
		if secret == "" {
			log.Println("FLW_SECRET_KEY not set")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration"})
			return
		}
		verifHash := c.GetHeader("verif-hash")
		if verifHash == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature header"})
			return
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))
		if verifHash != expectedMAC {
			log.Printf("Invalid webhook signature: got %s, expected %s", verifHash, expectedMAC)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}
		log.Printf("Received Flutterwave webhook: %s", string(body))

		var webhookData map[string]interface{}
		if err := json.Unmarshal(body, &webhookData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Example: Extract event type and transaction reference
		event, _ := webhookData["event"].(string)
		data, _ := webhookData["data"].(map[string]interface{})
		status, _ := data["status"].(string)
		txRef, _ := data["tx_ref"].(string)
		amount, _ := data["amount"].(float64)

		if event == "charge.completed" && status == "successful" && txRef != "" {
			ctx := c.Request.Context()
			// Find transaction by txRef
			transaction, err := repository.GetTransactionByTxRef(ctx, db, txRef)
			if err != nil {
				log.Printf("Transaction not found for txRef: %s", txRef)
				c.JSON(http.StatusOK, gin.H{"status": "ignored"})
				return
			}
			if transaction.Status == "success" {
				// Already processed
				c.JSON(http.StatusOK, gin.H{"status": "already_processed"})
				return
			}
			// Update transaction status
			err = repository.UpdateTransactionStatus(ctx, db, transaction.ID, "success")
			if err != nil {
				log.Printf("Failed to update transaction status: %v", err)
				c.JSON(http.StatusOK, gin.H{"status": "error"})
				return
			}
			// Update wallet balance
			err = repository.UpdateWalletBalance(db, transaction.ToWallet, amount, true)
			if err != nil {
				log.Printf("Failed to update wallet balance: %v", err)
				c.JSON(http.StatusOK, gin.H{"status": "error"})
				return
			}
			// Create notification for the user
			wallet, err := repository.GetWalletByID(db, transaction.ToWallet)
			if err == nil {
				notifRepo := repository.NewNotificationRepository(db)
				notifService := services.NewNotificationService(notifRepo)
				notifService.CreateWalletFundedNotification(ctx, wallet.OwnerID, amount)
			}
			log.Printf("Wallet funded: txRef=%s, amount=%.2f", txRef, amount)
		}

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	}
} 