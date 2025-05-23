package services

import (
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
)

func GetUserByID(db *mongo.Collection, id int) (*models.User, error) {
	return repository.GetUserByID(db, id)
}
