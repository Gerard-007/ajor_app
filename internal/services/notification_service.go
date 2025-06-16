package services

import (
	"context"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserNotifications(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) ([]*models.Notification, error) {
	return repository.GetUserNotifications(ctx, db, userID)
}