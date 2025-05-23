package services

import (
	"context"
	"errors"
	// "time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *mongo.Collection, user *models.User) error {
	// Check if the email already exists
	var existingUser models.User
	err := db.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return errors.New("email already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return repository.CreateUser(db, user)
}


func LoginUser(db *mongo.Collection, email, password string) (*models.User, error) {
	// Find the user by email
	var user models.User
	err := db.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
        return nil, errors.New("invalid email or password")
    }
    if err != nil {
        return nil, err
    }

	// Compare the password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}