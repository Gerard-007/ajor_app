package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/Gerard-007/ajor_app/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *mongo.Database, user *models.User, pg payment.PaymentGateway) error {
	usersCollection := db.Collection("users")

	// Validate input
	if user.Username == "" || user.Email == "" || user.Password == "" || user.Phone == "" || user.BVN == "" {
		return errors.New("username, email, password, phone and BVN are required")
	}

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
	// user.Verified = false
	//user.IsAdmin = false // Enforce false for security; override via MongoDB or admin endpoint

	// Create user
	userResult, err := usersCollection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	user.ID = userResult.InsertedID.(primitive.ObjectID)

	// Create profile
	profile := models.Profile{
		UserID:     user.ID,
		Bio:        "",
		Location:   "",
		ProfilePic: "",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = repository.CreateProfile(db, &profile)
	if err != nil {
		// Clean up user if profile creation fails
		_, delErr := usersCollection.DeleteOne(context.Background(), bson.M{"_id": user.ID})
		if delErr != nil {
			log.Printf("Failed to clean up user after profile creation failure: %v", delErr)
		}
		return err
	}

	// Create wallet
	wallet := models.Wallet{
		ID:        primitive.NewObjectID(),
		OwnerID:   user.ID,
		Type:      models.WalletTypeUser,
		Balance:   0.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repository.CreateWallet(db, &wallet)
	if err != nil {
		// Clean up user and profile if wallet creation fails
		_, delErr := usersCollection.DeleteOne(context.Background(), bson.M{"_id": user.ID})
		if delErr != nil {
			log.Printf("Failed to clean up user after wallet creation failure: %v", delErr)
		}
		return err
	}
	log.Printf("Created wallet with ID: %s for user: %s", wallet.ID.Hex(), user.Email)


	// Create virtual account
	narration := fmt.Sprintf("Wallet for %s", user.Username)
	va, err := pg.CreateVirtualAccount(context.Background(), user.ID, user.Email, user.Phone, narration, true, user.BVN, 0.0)
	if err != nil {
		log.Printf("Failed to create virtual account for user %s: %v", user.Email, err)
		usersCollection.DeleteOne(context.Background(), bson.M{"_id": user.ID})
		db.Collection("wallets").DeleteOne(context.Background(), bson.M{"_id": wallet.ID})
		return fmt.Errorf("failed to create virtual account: %v", err)
	}

	// Update wallet with virtual account details
	if err := repository.UpdateWalletVirtualAccount(db, wallet.ID, va.AccountNumber, va.AccountID, va.BankName); err != nil {
		log.Printf("Failed to update wallet %s with virtual account for user %s: %v", wallet.ID.Hex(), user.Email, err)
		usersCollection.DeleteOne(context.Background(), bson.M{"_id": user.ID})
		db.Collection("wallets").DeleteOne(context.Background(), bson.M{"_id": wallet.ID})
		return fmt.Errorf("wallet not found: %v", err)
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
