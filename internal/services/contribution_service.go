package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateContribution(ctx context.Context, db *mongo.Database, pg payment.PaymentGateway, contribution *models.Contribution, groupAdminID primitive.ObjectID) error {
	if contribution.Name == "" || contribution.Cycle == "" || contribution.Type == "" {
		return errors.New("name, cycle, and type are required")
	}
	if contribution.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if contribution.PenaltyAmount < 0 {
		return errors.New("penalty amount cannot be negative")
	}
	if contribution.CycleCount <= 0 {
		return errors.New("cycle count must be positive")
	}
	if !isValidCycle(contribution.Cycle) || !isValidType(contribution.Type) {
		return errors.New("invalid cycle or type")
	}

	// Set collection day and deadline
	switch contribution.Cycle {
	case models.CycleDaily:
		contribution.CollectionDay = "end of day"
		contribution.CollectionDeadline = time.Now().Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleWeekly:
		contribution.CollectionDay = "end of week"
		daysUntilSunday := (7 - int(time.Now().Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		contribution.CollectionDeadline = time.Now().Add(time.Duration(daysUntilSunday) * 24 * time.Hour).Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleMonthly:
		contribution.CollectionDay = "last day of month"
		nextMonth := time.Now().AddDate(0, 1, 0)
		lastDay := nextMonth.AddDate(0, 0, -nextMonth.Day())
		contribution.CollectionDeadline = lastDay.Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleYearly:
		contribution.CollectionDay = "last day of year"
		contribution.CollectionDeadline = time.Date(time.Now().Year(), 12, 31, 23, 59, 59, 0, time.Now().Location())
	}

	// Create wallet
	wallet := &models.Wallet{
		ID:      primitive.NewObjectID(),
		OwnerID: groupAdminID,
		Type:    models.WalletTypeContribution,
		Balance: 0.0,
	}
	if err := repository.CreateWallet(db, wallet); err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create virtual account
	var user models.User
	err := db.Collection("users").FindOne(ctx, bson.M{"_id": groupAdminID}).Decode(&user)
	if err != nil {
		repository.DeleteWallet(db, wallet.ID)
		return errors.New("group admin not found")
	}
	narration := fmt.Sprintf("Contribution %s", contribution.Name)
	va, err := pg.CreateVirtualAccount(ctx, groupAdminID, user.Email, user.Phone, narration, true, user.BVN)
	if err != nil {
		repository.DeleteWallet(db, wallet.ID)
		return fmt.Errorf("failed to create virtual account: %v", err)
	}
	if err := repository.UpdateWalletVirtualAccount(db, wallet.ID, va); err != nil {
		repository.DeleteWallet(db, wallet.ID)
		return fmt.Errorf("failed to update wallet with virtual account: %w", err)
	}

	// Set contribution fields
	contribution.GroupAdmin = groupAdminID
	contribution.WalletID = wallet.ID
	contribution.YetToCollectMembers = []primitive.ObjectID{groupAdminID}
	contribution.AlreadyCollectedMembers = []primitive.ObjectID{}

	return repository.CreateContribution(ctx, db, contribution)
}

func GetContribution(ctx context.Context, db *mongo.Database, id, userID primitive.ObjectID) (*models.Contribution, error) {
	contribution, err := repository.GetContributionByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if contribution.GroupAdmin != userID &&
		!containsUser(contribution.YetToCollectMembers, userID) &&
		!containsUser(contribution.AlreadyCollectedMembers, userID) {
		return nil, errors.New("unauthorized access to contribution")
	}

	return contribution, nil
}

func GetUserContributions(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) ([]*models.Contribution, error) {
	return repository.GetContributionsByUser(ctx, db, userID)
}

func UpdateContribution(ctx context.Context, db *mongo.Database, id, userID primitive.ObjectID, contribution *models.Contribution) error {
	existing, err := repository.GetContributionByID(ctx, db, id)
	if err != nil {
		return err
	}
	if existing.GroupAdmin != userID {
		return errors.New("only group admin can update contribution")
	}

	return repository.UpdateContribution(ctx, db, id, contribution)
}

func JoinContribution(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID, inviteCode string) error {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return err
	}

	if contribution.InviteCode != inviteCode {
		return errors.New("invalid invite code")
	}

	if containsUser(contribution.YetToCollectMembers, userID) || containsUser(contribution.AlreadyCollectedMembers, userID) {
		return errors.New("user already in contribution")
	}

	err = repository.JoinContribution(ctx, db, contributionID, userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		UserID:         contribution.GroupAdmin,
		ContributionID: contributionID,
		Message:        "A new member has joined your contribution group: " + contribution.Name,
		Type:           models.NotificationInfo,
	}
	return repository.CreateNotification(ctx, db, notification)
}

func RemoveMember(ctx context.Context, db *mongo.Database, contributionID, userID, groupAdminID primitive.ObjectID) error {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return err
	}
	if contribution.GroupAdmin != groupAdminID {
		return errors.New("only group admin can remove members")
	}

	err = repository.RemoveMember(ctx, db, contributionID, userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		UserID:         userID,
		ContributionID: contributionID,
		Message:        "You have been removed from the contribution group: " + contribution.Name,
		Type:           models.NotificationWarning,
	}
	return repository.CreateNotification(ctx, db, notification)
}

func isValidCycle(cycle models.ContributionCycle) bool {
	return cycle == models.CycleDaily || cycle == models.CycleWeekly || cycle == models.CycleMonthly || cycle == models.CycleYearly
}

func isValidType(contributionType models.ContributionType) bool {
	return contributionType == models.TypeDailySavings || contributionType == models.TypeGroupContribution
}

func containsUser(members []primitive.ObjectID, userID primitive.ObjectID) bool {
	for _, member := range members {
		if member == userID {
			return true
		}
	}
	return false
}

func GetAllContributions(db *mongo.Database, isSystemAdmin bool) ([]*models.Contribution, error) {
	if !isSystemAdmin {
		return nil, errors.New("only system admins can view all contributions")
	}
	return repository.GetAllContributions(db)
}