package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Collection struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ContributionID primitive.ObjectID `json:"contribution_id" bson:"contribution_id"`
	Collector      primitive.ObjectID `json:"collector" bson:"collector"`
	CollectionDate time.Time          `json:"collection_date" bson:"collection_date"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}