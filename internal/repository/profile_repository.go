package repository

import (
	"context"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserProfile(db *mongo.Collection, userId int) (*models.Profile, error) {
	// Get user profile by user ID
	var profile models.Profile
	err := db.FindOne(context.TODO(), map[string]any{"user_id": userId}).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
