package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `bson:"username" json:"username,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	IsAdmin   bool               `json:"is_admin" bson:"is_admin"`
	Phone     string             `json:"phone" bson:"phone"`
	BVN       string             `json:"bvn" bson:"bvn,omitempty"`
	Verified  bool               `json:"verified" bson:"verified"`
	VerificationToken string    `json:"verification_token" bson:"verification_token"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type UserResponse struct {
	ID        primitive.ObjectID `json:"_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	IsAdmin   bool               `json:"is_admin"`
	Phone     string             `json:"phone"`
	BVN       string             `json:"bvn"`
	Verified  bool               `json:"verified"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Profile   *Profile           `json:"profile"`
	Wallet    *Wallet            `json:"wallet"`
}