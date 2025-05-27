package services

import (
	"errors"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllUsers(db *mongo.Collection) ([]*models.User, error) {
	return repository.GetAllUsers(db)
}

func GetUserByID(db *mongo.Collection, id primitive.ObjectID) (*models.User, error) {
	return repository.GetUserByID(db, id)
}

type UserUpdate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Verified bool   `json:"verified"`
	IsAdmin  bool   `json:"is_admin"`
}

func UpdateUser(db *mongo.Database, id primitive.ObjectID, userUpdate *UserUpdate, isAdmin bool) (*models.User, error) {
	// Restrict Verified and IsAdmin updates to admins
	if !isAdmin {
		if userUpdate.Verified || userUpdate.IsAdmin {
			return nil, errors.New("only admins can update verified or admin status")
		}
		userUpdate.Verified = false
		userUpdate.IsAdmin = false
	}

	// Map to repository UserUpdate
	repoUpdate := &repository.UserUpdate{
		Email:    userUpdate.Email,
		Username: userUpdate.Username,
		Phone:    userUpdate.Phone,
		Verified: userUpdate.Verified,
		IsAdmin:  userUpdate.IsAdmin,
	}

	return repository.UpdateUser(db, id, repoUpdate)
}

func DeleteUser(db *mongo.Database, userID primitive.ObjectID) error {
	return repository.DeleteUserAndProfile(db, userID)
}
