package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GenerateUsernameFromEmail generates a unique username from an email address.
// For example, "geetechlab@gmail.com" becomes "geetechlab". If "geetechlab" exists,
// it tries "geetechlab01", "geetechlab02", etc.
func GenerateUsernameFromEmail(db *mongo.Database, email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	// Extract the local part of the email (before @)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format")
	}
	baseUsername := strings.ToLower(strings.TrimSpace(parts[0]))

	usersCollection := db.Collection("users")
	ctx := context.Background()

	// Check if base username exists
	var existingUser models.User
	err := usersCollection.FindOne(ctx, bson.M{"username": baseUsername}).Decode(&existingUser)
	if err == mongo.ErrNoDocuments {
		return baseUsername, nil // Base username is available
	}
	if err != nil {
		return "", fmt.Errorf("failed to check username: %v", err)
	}

	// Try usernames with numeric suffixes (e.g., geetechlab01, geetechlab02)
	for i := 1; i <= 100; i++ { // Limit to 100 attempts to prevent infinite loops
		newUsername := fmt.Sprintf("%s%02d", baseUsername, i)
		err = usersCollection.FindOne(ctx, bson.M{"username": newUsername}).Decode(&existingUser)
		if err == mongo.ErrNoDocuments {
			return newUsername, nil // Unique username found
		}
		if err != nil {
			return "", fmt.Errorf("failed to check username: %v", err)
		}
	}

	return "", fmt.Errorf("could not generate a unique username after 100 attempts")
}