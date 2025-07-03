package services

import (
	"errors"
	"fmt"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllUsers(db *mongo.Collection) ([]*models.User, error) {
	return repository.GetAllUsers(db)
}

// func GetUserByID(db *mongo.Collection, id primitive.ObjectID) (*models.User, error) {
// 	return repository.GetUserByID(db, id)
// }

func GetUserByID(db *mongo.Database, id primitive.ObjectID) (*models.UserResponse, error) {
	//ctx := context.Background()

	// Fetch user
	usersCollection := db.Collection("users")
	user, err := repository.GetUserByID(usersCollection, id)
	if err != nil {
		return nil, err
	}

	fmt.Println("user", user)
	// Fetch profile
	profilesCollection := db.Collection("profiles")
	profile, err := repository.GetProfileByUserID(profilesCollection, id)
	if err != nil {
		return nil, err
	}

	// Fetch wallet
	walletsCollection := db.Collection("wallets")
	wallet, err := repository.GetWalletByOwnerID(walletsCollection, id)
	if err != nil {
		return nil, err
	}

	// Combine into UserResponse
	response := &models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		Phone:     user.Phone,
		BVN:       user.BVN,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Profile:   profile,
		Wallet:    wallet,
	}

	// Map wallet fields to match desired output
	response.Wallet.VirtualAccountID = wallet.VirtualAccountID
	response.Wallet.VirtualBankName = wallet.VirtualBankName
	response.Wallet.VirtualAccountNumber = wallet.VirtualAccountID

	return response, nil
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
