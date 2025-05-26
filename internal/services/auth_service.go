package services

import (
	"context"
	"errors"
	"log"
	"time"

	// "time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *mongo.Database, user *models.User) error {
	usersCollection := db.Collection("users")

	// Check if email or username exists
	var existingUser models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return errors.New("email already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}
	err = usersCollection.FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		return errors.New("username already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Verified = false
	//user.IsAdmin = false // Enforce false for security; override via MongoDB or admin endpoint

	// Create user
	userResult, err := usersCollection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	// Create profile
	profile := models.Profile{
		UserID:     userResult.InsertedID.(primitive.ObjectID),
		Bio:        "",
		Location:   "",
		ProfilePic: "",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = repository.CreateProfile(db, &profile)
	if err != nil {
		// Clean up user if profile creation fails
		_, delErr := usersCollection.DeleteOne(context.Background(), bson.M{"_id": userResult.InsertedID})
		if delErr != nil {
			log.Printf("Failed to clean up user after profile creation failure: %v", delErr)
		}
		return err
	}

	return nil
}

func LoginUser(db *mongo.Collection, email, password string) (string, error) {
	// Find the user by email
	user, err := repository.GetUserByEmail(db, email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("user not found")
		}
		return "", err
	}

	// Compare the provided password with the stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.Username, user.Email, user.ID, user.IsAdmin)
	if err != nil {
		return "", err
	}

	return token, nil
}
