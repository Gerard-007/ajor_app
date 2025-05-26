package services

import (
	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserByID(db *mongo.Collection, id primitive.ObjectID) (*models.User, error) {
	return repository.GetUserByID(db, id)
}

func DeleteUser(db *mongo.Database, userID primitive.ObjectID) error {
	return repository.DeleteUserAndProfile(db, userID)
}
