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

func ApprovePayout(ctx context.Context, db *mongo.Database, notificationService *NotificationService, approvalID, approverID primitive.ObjectID, approve bool) error {
	var approval models.Approval
	err := db.Collection("approvals").FindOne(ctx, bson.M{"_id": approvalID}).Decode(&approval)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("approval not found")
		}
		return err
	}

	if approval.ApproverID != approverID {
		return errors.New("unauthorized to approve this payout")
	}

	if approval.Status != models.ApprovalPending {
		return errors.New("approval already processed")
	}

	var status models.ApprovalStatus
	if approve {
		status = models.ApprovalApproved
	} else {
		status = models.ApprovalRejected
	}

	if err := repository.UpdateApproval(ctx, db, approvalID, status); err != nil {
		return err
	}

	if approve {
		var transaction models.Transaction
		err := db.Collection("transactions").FindOne(ctx, bson.M{"_id": approval.TransactionID}).Decode(&transaction)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return errors.New("transaction not found")
			}
			return err
		}

		// Update wallet balances
		if err := repository.UpdateWalletBalance(db, transaction.FromWallet, transaction.Amount, false); err != nil {
			return err
		}
		if err := repository.UpdateWalletBalance(db, transaction.ToWallet, transaction.Amount, true); err != nil {
			// Rollback
			repository.UpdateWalletBalance(db, transaction.FromWallet, transaction.Amount, true)
			return err
		}

		// Update transaction status
		_, err = db.Collection("transactions").UpdateOne(ctx, bson.M{"_id": transaction.ID}, bson.M{
			"$set": bson.M{
				"status":     models.StatusSuccess,
				"updated_at": time.Now(),
			},
		})
		if err != nil {
			// Rollback
			repository.UpdateWalletBalance(db, transaction.FromWallet, transaction.Amount, true)
			repository.UpdateWalletBalance(db, transaction.ToWallet, transaction.Amount, false)
			return err
		}

		// Mark member as collected
		if err := repository.MarkMemberCollected(ctx, db, approval.ContributionID, transaction.ToWallet); err != nil {
			return err
		}

		// Notify user
		n := &models.Notification{
			UserID:  transaction.ToWallet,
			Type:    "payout_approved",
			Title:   "Payout Approved",
			Message: fmt.Sprintf("Payout of %.2f approved for contribution", transaction.Amount),
			Meta:    map[string]interface{}{ "amount": transaction.Amount },
		}
		return notificationService.Create(ctx, n)
	}

	return nil
}

func GetPendingApprovals(ctx context.Context, db *mongo.Database, approverID primitive.ObjectID) ([]*models.Approval, error) {
	return repository.GetPendingApprovals(ctx, db, approverID)
}