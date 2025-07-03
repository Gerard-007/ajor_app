package repository

import (
	"context"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationRepository struct {
	Collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		Collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	n.CreatedAt = time.Now()
	_, err := r.Collection.InsertOne(ctx, n)
	return err
}

func (r *NotificationRepository) GetByUser(ctx context.Context, userID primitive.ObjectID, unreadOnly bool) ([]models.Notification, error) {
	filter := bson.M{"user_id": userID}
	if unreadOnly {
		filter["read"] = false
	}
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var notifications []models.Notification
	for cur.Next(ctx) {
		var n models.Notification
		if err := cur.Decode(&n); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error {
	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": notificationID}, bson.M{"$set": bson.M{"read": true}})
	return err
}

func (r *NotificationRepository) MarkAsUnread(ctx context.Context, notificationID primitive.ObjectID) error {
	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": notificationID}, bson.M{"$set": bson.M{"read": false}})
	return err
}

func (r *NotificationRepository) Delete(ctx context.Context, notificationID primitive.ObjectID) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": notificationID})
	return err
}