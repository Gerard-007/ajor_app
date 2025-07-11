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
	"strconv"
)


func RegisterUser(db *mongo.Database, user *models.User, pg payment.PaymentGateway) (string, error) {
	usersCollection := db.Collection("users")

	// Generate username from email if not provided
	if user.Username == "" {
		generatedUsername, err := utils.GenerateUsernameFromEmail(db, user.Email)
		if err != nil {
			log.Printf("Failed to generate username: %v", err)
			return "", fmt.Errorf("failed to generate username: %v", err)
		}
		user.Username = generatedUsername
	}

	// Validate input
	if user.Username == "" {
		return "", errors.New("username is required")
	}
	if user.Email == "" {
		return "", errors.New("email is required")
	}
	if user.Password == "" {
		return "", errors.New("password is required")
	}
	if user.Phone == "" || len(user.Phone) < 11 {
		return "", errors.New("phone is required and must be at least 11 digits")
	}
	// Check if phone contains only digits
	if _, err := strconv.Atoi(user.Phone); err != nil {
		return "", errors.New("phone must contain only digits")
	}

	if user.BVN == "" || len(user.BVN) != 11 {
		return "", errors.New("BVN is required and must be 11 digits")
	}
	// Check if BVN contains only digits
	if _, err := strconv.Atoi(user.BVN); err != nil {
		return "", errors.New("BVN must contain only digits")
	}

	// Check if email, username, or phone exists
	var existingUser models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := usersCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		log.Printf("Email already registered: %s", user.Email)
		return "", errors.New("email already exists")
	}
	if err != mongo.ErrNoDocuments {
		log.Printf("Error checking email existence: %v", err)
		return "", err
	}

	err = usersCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		log.Printf("Username already taken: %s", user.Username)
		return "", errors.New("username already exists")
	}
	if err != mongo.ErrNoDocuments {
		log.Printf("Error checking username existence: %v", err)
		return "", err
	}

	err = usersCollection.FindOne(ctx, bson.M{"phone": user.Phone}).Decode(&existingUser)
	if err == nil {
		log.Printf("Phone already registered: %s", user.Phone)
		return "", errors.New("phone already exists")
	}
	if err != mongo.ErrNoDocuments {
		log.Printf("Error checking phone existence: %v", err)
		return "", err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsAdmin = false // Enforce false for security

	// Create user
	userResult, err := usersCollection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return "", err
	}
	user.ID = userResult.InsertedID.(primitive.ObjectID)

	// Create profile
	profile := models.Profile{
		ID:            primitive.NewObjectID(),
		UserID:        user.ID,
		Bio:           "",
		Location:      "",
		ProfilePic:    "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = repository.CreateProfile(db, &profile)
	if err != nil {
		log.Printf("Error creating profile for user %s: %v", user.Email, err)
		// Clean up user
		_, delErr := usersCollection.DeleteOne(ctx, bson.M{"_id": user.ID})
		if delErr != nil {
			log.Printf("Failed to clean up user after profile creation failure: %v", delErr)
		}
		return "", err
	}

	// Create wallet
	wallet := models.Wallet{
		ID:            primitive.NewObjectID(),
		OwnerID:       user.ID,
		Type:          models.WalletTypeUser,
		Balance:       0.0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = repository.CreateWallet(db, &wallet)
	if err != nil {
		log.Printf("Error creating wallet for user %s: %v", user.Email, err)
		// Clean up user and profile
		_, delErr := usersCollection.DeleteOne(ctx, bson.M{"_id": user.ID})
		if delErr != nil {
			log.Printf("Failed to clean up user after wallet creation failure: %v", delErr)
		}
		db.Collection("profiles").DeleteOne(ctx, bson.M{"user_id": user.ID})
		return "", err
	}
	log.Printf("Created wallet with ID: %s for user: %s", wallet.ID.Hex(), user.Email)

	// Create virtual account
	narration := fmt.Sprintf("Wallet for %s", user.Username)
	va, err := pg.CreateVirtualAccount(ctx, user.ID, user.Email, user.Phone, narration, true, user.BVN, 0.0)
	if err != nil {
		log.Printf("Failed to create virtual account for user %s: %v", user.Email, err)
		usersCollection.DeleteOne(ctx, bson.M{"_id": user.ID})
		db.Collection("profiles").DeleteOne(ctx, bson.M{"user_id": user.ID})
		db.Collection("wallets").DeleteOne(ctx, bson.M{"_id": wallet.ID})
		return "", fmt.Errorf("failed to create virtual account: %v", err)
	}

	// Update wallet with virtual account details
	if err := repository.UpdateWalletVirtualAccount(db, wallet.ID, va.AccountNumber, va.AccountID, va.BankName); err != nil {
		log.Printf("Failed to update wallet %s with virtual account for user %s: %v", wallet.ID.Hex(), user.Email, err)
		usersCollection.DeleteOne(ctx, bson.M{"_id": user.ID})
		db.Collection("profiles").DeleteOne(ctx, bson.M{"user_id": user.ID})
		db.Collection("wallets").DeleteOne(ctx, bson.M{"_id": wallet.ID})
		return "", fmt.Errorf("failed to update wallet with virtual account: %v", err)
	}

	// Generate JWT token for immediate login
	token, err := utils.GenerateToken(user.Username, user.Email, user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("Error generating JWT for user %s: %v", user.Email, err)
		// Clean up user, profile, and wallet
		usersCollection.DeleteOne(ctx, bson.M{"_id": user.ID})
		db.Collection("profiles").DeleteOne(ctx, bson.M{"user_id": user.ID})
		db.Collection("wallets").DeleteOne(ctx, bson.M{"_id": wallet.ID})
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	log.Printf("User registered and logged in successfully: %s", user.Email)
	return token, nil
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
