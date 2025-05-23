package services

import (
	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserProfile(db *mongo.Collection, userId int) (*models.Profile, error) {
	return repository.GetUserProfile(db, userId)
}
