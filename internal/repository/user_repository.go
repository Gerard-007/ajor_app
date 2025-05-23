package repository

import (
	"context"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUser(db *mongo.Collection, user *models.User) error {
	// Implement the logic to create a user in the database
	_, err := db.InsertOne(context.TODO(), user)
	return err
}

func GetUserByID(db *mongo.Collection, id int) (*models.User, error) {
	// Get a user by ID from the database
	var user models.User
	err := db.FindOne(context.TODO(), map[string]any{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
