package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FundWallet initiates a funding request to the user's virtual account and updates the wallet balance upon success.
func FundWallet(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, amount float64, pg payment.PaymentGateway) error {
	// Get user and wallet
	user, err := repository.GetUserByID(db.Collection("users"), userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}
	wallet, err := repository.GetWalletByUserID(db, userID)
	if err != nil {
		return fmt.Errorf("wallet not found: %v", err)
	}
	if wallet.VirtualAccountID == "" {
		return fmt.Errorf("no virtual account linked to wallet")
	}

	// Initiate funding to virtual account
	txRef := fmt.Sprintf("fund-wallet-%s-%d", userID.Hex(), time.Now().UnixNano())
	fundingRequest := payment.FundingRequest{
		Email:        user.Email,
		Amount:       amount,
		TxRef:        txRef,
		Currency:     "NGN",
		IsPermanent:  false,
		Narration:    fmt.Sprintf("Fund wallet for %s", user.Username),
		PhoneNumber:  user.Phone,
	}

	transactionResponse, err := pg.FundVirtualAccount(ctx, wallet.VirtualAccountID, fundingRequest)
	if err != nil {
		return fmt.Errorf("failed to initiate funding: %v", err)
	}

	// Create pending transaction
	transaction := &models.Transaction{
		FromWallet:    primitive.ObjectID{}, // No source wallet for external funding
		ToWallet:      wallet.ID,
		Amount:        amount,
		Type:          models.TransactionWallet,
		Date:          time.Now(),
		PaymentMethod: models.PaymentBankTransfer,
		Status:        models.StatusPending,
		ContributionID: primitive.ObjectID{},
		TxRef:         txRef,
	}
	if err := repository.CreateTransaction(ctx, db, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// Verify transaction (in real-world, this would be via webhook; here we simulate verification)
	verifiedTx, err := pg.VerifyTransaction(ctx, transactionResponse.TransactionID)
	if err != nil {
		// Update transaction to failed
		repository.UpdateTransactionStatus(ctx, db, transaction.ID, models.StatusFailed)
		return fmt.Errorf("transaction verification failed: %v", err)
	}
	if verifiedTx.Status != "success" || verifiedTx.Amount != amount {
		repository.UpdateTransactionStatus(ctx, db, transaction.ID, models.StatusFailed)
		return fmt.Errorf("invalid transaction status or amount")
	}

	// Update wallet balance
	if err := repository.UpdateWalletBalance(db, wallet.ID, amount, true); err != nil {
		// Update transaction to failed (rollback attempt)
		repository.UpdateTransactionStatus(ctx, db, transaction.ID, models.StatusFailed)
		return fmt.Errorf("failed to update wallet balance: %v", err)
	}

	// Update transaction to success
	if err := repository.UpdateTransactionStatus(ctx, db, transaction.ID, models.StatusSuccess); err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}

	return nil
}


func GetContributionWallet(ctx context.Context, db *mongo.Database, pg payment.PaymentGateway, contributionID, userID primitive.ObjectID, isAdmin bool) (*models.Wallet, error) {
	log.Printf("Fetching contribution ID: %s for user ID: %s", contributionID.Hex(), userID.Hex())

	// Fetch contribution by ID
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Contribution not found: %s", contributionID.Hex())
			return nil, fmt.Errorf("contribution not found")
		}
		log.Printf("Error fetching contribution: %v", err)
		return nil, fmt.Errorf("failed to fetch contribution: %v", err)
	}

	// Check authorization
	log.Printf("Checking authorization for user ID: %s, isAdmin: %v", userID.Hex(), isAdmin)
	if contribution.GroupAdmin != userID && !isAdmin {
		// Ensure member arrays are not nil to prevent panic
		hasAccess := false
		if contribution.YetToCollectMembers != nil {
			for _, memberID := range contribution.YetToCollectMembers {
				if memberID == userID {
					hasAccess = true
					break
				}
			}
		}
		if !hasAccess && contribution.AlreadyCollectedMembers != nil {
			for _, memberID := range contribution.AlreadyCollectedMembers {
				if memberID == userID {
					hasAccess = true
					break
				}
			}
		}
		if !hasAccess {
			log.Printf("Unauthorized access for user ID: %s", userID.Hex())
			return nil, fmt.Errorf("unauthorized access")
		}
	}

	// Fetch wallet using WalletID from contribution
	if contribution.WalletID.IsZero() {
		log.Printf("No wallet ID associated with contribution: %s", contributionID.Hex())
		return nil, fmt.Errorf("wallet not found")
	}
	log.Printf("Fetching wallet ID: %s for contribution: %s", contribution.WalletID.Hex(), contributionID.Hex())
	wallet, err := repository.GetContributionWalletByID(ctx, db, contribution.WalletID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Wallet not found for ID: %s", contribution.WalletID.Hex())
			return nil, fmt.Errorf("wallet not found")
		}
		log.Printf("Error fetching wallet: %v", err)
		return nil, fmt.Errorf("failed to fetch wallet: %v", err)
	}

	// Fetch virtual account details if available
	if wallet.VirtualAccountID != "" {
		log.Printf("Fetching virtual account ID: %s", wallet.VirtualAccountID)
		va, err := pg.GetVirtualAccount(ctx, wallet.VirtualAccountID)
		if err != nil {
			log.Printf("Error fetching virtual account: %v", err)
			return nil, fmt.Errorf("failed to get virtual account: %v", err)
		}
		if va != nil {
			wallet.VirtualAccountNumber = va.AccountNumber
			wallet.VirtualBankName = va.BankName
		} else {
			log.Printf("Virtual account is nil for ID: %s", wallet.VirtualAccountID)
		}
	}

	log.Printf("Returning wallet: %+v", wallet)
	return wallet, nil
}