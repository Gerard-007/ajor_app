package services

import (
	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserProfile(db *mongo.Database, userID primitive.ObjectID) (*models.Profile, error) {
	return repository.GetUserProfile(db.Collection("profiles"), userID)
}

func UpdateUserProfile(db *mongo.Database, userID primitive.ObjectID, profileUpdate *models.Profile) error {
	return repository.UpdateUserProfile(db.Collection("profiles"), userID, profileUpdate)
}

func UpdateUserProfilePicture(db *mongo.Database, userID primitive.ObjectID, picturePath string) error {
	return repository.UpdateUserProfilePicture(db.Collection("profiles"), userID, picturePath)
}
