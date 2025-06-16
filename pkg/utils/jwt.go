package utils

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTConfig struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.StandardClaims
	ExpiresAt int64 `json:"exp"`
}

func GenerateToken(username, email string, userID primitive.ObjectID, isAdmin bool) (string, error) {
	claims := JWTConfig{
		Email:    email,
		Username: username,
		UserID:   userID.Hex(),
		IsAdmin:  isAdmin,
		ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET"))) // Replace with your secret key
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateToken(tokenString string) (*JWTConfig, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTConfig{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil // Replace with your secret key
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTConfig)
	if !ok {
		return nil, jwt.NewValidationError("invalid token claims", jwt.ValidationErrorClaimsInvalid)
	}
	return claims, nil
}