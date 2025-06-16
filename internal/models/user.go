package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	IsAdmin   bool               `json:"is_admin" bson:"is_admin"`
	WalletID  primitive.ObjectID `json:"wallet_id" bson:"wallet_id"`
	Phone     string             `json:"phone" bson:"phone"`
	BVN       string             `json:"bvn" bson:"bvn,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
