package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	// "time"

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
		return errors.New("Username, email, password, phone and BVN are required.")
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

	// Create wallet
	wallet := &models.Wallet{
		OwnerID: user.ID,
		Type:    models.WalletTypeUser,
		Balance: 0.0,
	}
	if err := repository.CreateWallet(db, wallet); err != nil {
		// Rollback user creation
		usersCollection.DeleteOne(context.TODO(), bson.M{"_id": user.ID})
		return err
	}

	// Create virtual account
	narration := fmt.Sprintf("Wallet for %s", user.Username)
	va, err := pg.CreateVirtualAccount(context.TODO(), user.ID, user.Email, user.Phone, narration, true, user.BVN)
	if err != nil {
		usersCollection.DeleteOne(context.TODO(), bson.M{"_id": user.ID})
		db.Collection("wallets").DeleteOne(context.TODO(), bson.M{"_id": wallet.ID})
		return err
	}

	// Update wallet with virtual account details
	if err := repository.UpdateWalletVirtualAccount(db, wallet.ID, va); err != nil {
		usersCollection.DeleteOne(context.TODO(), bson.M{"_id": user.ID})
		db.Collection("wallets").DeleteOne(context.TODO(), bson.M{"_id": wallet.ID})
		return err
	}

	// Update user with wallet ID
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"wallet_id": wallet.ID}}
	if _, err := usersCollection.UpdateOne(context.TODO(), filter, update); err != nil {
		// Rollback
		usersCollection.DeleteOne(context.TODO(), bson.M{"_id": user.ID})
		db.Collection("wallets").DeleteOne(context.TODO(), bson.M{"_id": wallet.ID})
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
