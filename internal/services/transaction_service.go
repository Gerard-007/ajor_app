package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RecordContribution(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID, amount float64, paymentMethod models.PaymentMethod) error {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return err
	}
	if !containsUser(contribution.YetToCollectMembers, userID) && !containsUser(contribution.AlreadyCollectedMembers, userID) {
		return errors.New("user not in contribution")
	}
	if amount != contribution.Amount {
		return errors.New("contribution amount mismatch")
	}

	// Get wallets
	var user models.User
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return errors.New("user not found")
	}
	userWallet, err := repository.GetWalletByUserID(db, user.ID)
	if err != nil {
		return errors.New("user wallet not found")
	}
	fmt.Println("Wallet ID from contribution:", contribution.WalletID.Hex())

	groupWallet, err := repository.GetWalletByID(db, contribution.WalletID)
	if err != nil {
		return errors.New("group wallet not found")
	}

	// Check balance
	if userWallet.Balance < amount {
		return errors.New("insufficient balance")
	}

	// Update wallets
	if err := repository.UpdateWalletBalance(db, userWallet.ID, amount, false); err != nil {
		return err
	}
	if err := repository.UpdateWalletBalance(db, groupWallet.ID, amount, true); err != nil {
		// Rollback
		repository.UpdateWalletBalance(db, userWallet.ID, amount, true)
		return err
	}

	transaction := &models.Transaction{
		FromWallet:     userWallet.ID,
		ToWallet:       groupWallet.ID,
		Amount:         amount,
		Type:           models.TransactionContribution,
		Date:           time.Now(),
		PaymentMethod:  paymentMethod,
		Status:         models.StatusSuccess,
		ContributionID: contributionID,
	}
	if err := repository.CreateTransaction(ctx, db, transaction); err != nil {
		// Rollback
		repository.UpdateWalletBalance(db, userWallet.ID, amount, true)
		repository.UpdateWalletBalance(db, groupWallet.ID, amount, false)
		return err
	}

	if time.Now().After(contribution.CollectionDeadline) {
		notification := &models.Notification{
			UserID:         userID,
			ContributionID: contributionID,
			Message:        fmt.Sprintf("Late contribution recorded. Penalty applied: %.2f", contribution.PenaltyAmount),
			Type:           models.NotificationWarning,
		}
		return repository.CreateNotification(ctx, db, notification)
	}
	return nil
}

func RecordPayout(ctx context.Context, db *mongo.Database, contributionID, userID, groupAdminID primitive.ObjectID, amount float64, paymentMethod models.PaymentMethod) error {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return err
	}

	if contribution.GroupAdmin != groupAdminID {
		return errors.New("only group admin can record payouts")
	}

	if !containsUser(contribution.YetToCollectMembers, userID) {
		return errors.New("user not eligible for payout")
	}

	// Get wallets
	var user models.User
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return errors.New("user not found")
	}
	userWallet, err := repository.GetWalletByID(db, user.ID)
	if err != nil {
		return errors.New("user wallet not found")
	}
	groupWallet, err := repository.GetWalletByID(db, contribution.WalletID)
	if err != nil {
		return errors.New("group wallet not found")
	}

	// Check balance
	if groupWallet.Balance < amount {
		return errors.New("insufficient balance in group wallet")
	}

	// Create transaction (pending)
	transaction := &models.Transaction{
		FromWallet:     groupWallet.ID,
		ToWallet:       userWallet.ID,
		Amount:         amount,
		Type:           models.TransactionPayout,
		Date:           time.Now(),
		PaymentMethod:  paymentMethod,
		Status:         models.StatusPending,
		ContributionID: contributionID,
	}
	if err := repository.CreateTransaction(ctx, db, transaction); err != nil {
		return err
	}

	// Create approval
	approval := &models.Approval{
		TransactionID:  transaction.ID,
		ApproverID:     groupAdminID,
		Status:         models.ApprovalPending,
		ContributionID: contributionID,
	}
	if err := repository.CreateApproval(ctx, db, approval); err != nil {
		// Rollback: Delete transaction
		db.Collection("transactions").DeleteOne(ctx, bson.M{"_id": transaction.ID})
		return err
	}

	notification := &models.Notification{
		UserID:         userID,
		ContributionID: contributionID,
		Message:        fmt.Sprintf("Payout of %.2f requested for contribution: %s", amount, contribution.Name),
		Type:           models.NotificationInfo,
	}
	return repository.CreateNotification(ctx, db, notification)
}

// func GetUserTransactions(ctx context.Context, db *mongo.Database, userID, contributionID primitive.ObjectID) ([]*models.Transaction, error) {
// 	return repository.GetUserTransactions(ctx, db, userID, contributionID)
// }


func GetUserTransactions(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, isAdmin bool) ([]models.Transaction, error) {
	var filter bson.M
	if isAdmin {
		// System admins can view all transactions
		filter = bson.M{}
	} else {
		// Regular users can only view transactions involving their wallet
		wallet, err := repository.GetWalletByUserID(db, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch wallet: %v", err)
		}
		filter = bson.M{
			"$or": []bson.M{
				{"from_wallet": wallet.ID},
				{"to_wallet": wallet.ID},
			},
		}
	}

	transactions, err := repository.GetTransactions(ctx, db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	return transactions, nil
}

func GetContributionTransactions(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID, isAdmin bool) ([]models.Transaction, error) {
	// Fetch contribution to verify authorization
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("contribution not found")
		}
		return nil, fmt.Errorf("failed to fetch contribution: %v", err)
	}

	// Check authorization: user must be group admin, a member, or system admin
	if contribution.GroupAdmin != userID && !isAdmin && !containsUser(contribution.YetToCollectMembers, userID) && !containsUser(contribution.AlreadyCollectedMembers, userID) {
		return nil, fmt.Errorf("unauthorized access")
	}

	// Fetch wallet for the contribution
	wallet, err := repository.GetWalletByContributionID(db, contributionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wallet: %v", err)
	}

	// Fetch transactions involving the contribution's wallet
	filter := bson.M{
		"$or": []bson.M{
			{"from_wallet": wallet.ID},
			{"to_wallet": wallet.ID},
		},
	}
	transactions, err := repository.GetTransactions(ctx, db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	return transactions, nil
}