package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID             uuid.UUID `json:"id" bson:"_id,omitempty"`
	UserID         uuid.UUID `json:"user_id" bson:"user_id"`
	Bio            string    `json:"bio" bson:"bio"`
	Location       string    `json:"location" bson:"location"`
	ProfilePic     string    `json:"profile_pic" bson:"profile_pic"`
	Occupation     string    `json:"occupation" bson:"occupation"`
	IsSalaryEarner bool      `json:"is_salary_earner" bson:"is_salary_earner"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
}
