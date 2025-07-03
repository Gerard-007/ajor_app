package services

import (
	"context"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(ctx context.Context, n *models.Notification) error {
	return s.repo.Create(ctx, n)
}

func (s *NotificationService) GetByUser(ctx context.Context, userID primitive.ObjectID, unreadOnly bool) ([]models.Notification, error) {
	return s.repo.GetByUser(ctx, userID, unreadOnly)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}

func (s *NotificationService) MarkAsUnread(ctx context.Context, notificationID primitive.ObjectID) error {
	return s.repo.MarkAsUnread(ctx, notificationID)
}

func (s *NotificationService) Delete(ctx context.Context, notificationID primitive.ObjectID) error {
	return s.repo.Delete(ctx, notificationID)
}

// Helper for common notification types
func (s *NotificationService) CreateWalletFundedNotification(ctx context.Context, userID primitive.ObjectID, amount float64) error {
	n := &models.Notification{
		UserID:  userID,
		Type:    "wallet_funded",
		Title:   "Wallet Funded",
		Message: "Your wallet has been funded.",
		Meta:    map[string]interface{}{ "amount": amount },
	}
	return s.Create(ctx, n)
}

func (s *NotificationService) CreateGroupContributionNotification(ctx context.Context, userID primitive.ObjectID, groupName string, amount float64) error {
	n := &models.Notification{
		UserID:  userID,
		Type:    "group_contribution",
		Title:   "Group Contribution",
		Message: "You contributed to group " + groupName,
		Meta:    map[string]interface{}{ "group": groupName, "amount": amount },
	}
	return s.Create(ctx, n)
}