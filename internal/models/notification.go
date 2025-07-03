package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Type       string                 `bson:"type" json:"type"`
	Title      string                 `bson:"title" json:"title"`
	Message    string                 `bson:"message" json:"message"`
	Read       bool                   `bson:"read" json:"read"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
	ActionLink string                 `bson:"action_link,omitempty" json:"action_link,omitempty"`
	Meta       map[string]interface{} `bson:"meta,omitempty" json:"meta,omitempty"`
}