package repository

import (
	"context"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateNotification(ctx context.Context, db *mongo.Database, notification *models.Notification) error {
	collection := db.Collection("notifications")
	notification.CreatedAt = time.Now()
	_, err := collection.InsertOne(ctx, notification)
	return err
}

func GetUserNotifications(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	cursor, err := db.Collection("notifications").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var notification models.Notification
		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}
		notifications = append(notifications, &notification)
	}
	return notifications, nil
}