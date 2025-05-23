package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" bson:"_id,omitempty"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"` // Omit from JSON
	Username  string    `json:"username" bson:"username"`
	Phone     string    `json:"phone" bson:"phone"`
	Verified  bool      `json:"verified" bson:"verified"`
	IsAdmin   bool      `json:"is_admin" bson:"is_admin"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
