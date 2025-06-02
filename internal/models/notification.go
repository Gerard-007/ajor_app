package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotificationInfo    NotificationType = "info"
	NotificationWarning NotificationType = "warning"
	NotificationError   NotificationType = "error"
)

type Notification struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	ContributionID primitive.ObjectID `json:"contribution_id" bson:"contribution_id,omitempty"`
	Message        string             `json:"message" bson:"message"`
	Type           NotificationType   `json:"type" bson:"type"`
	Read           bool               `json:"read" bson:"read"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}