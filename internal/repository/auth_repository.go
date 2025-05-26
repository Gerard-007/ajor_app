package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func BlacklistToken(collection *mongo.Collection, token string, expiresAt time.Time) error {
	_, err := collection.InsertOne(context.TODO(), bson.M{
		"token":      token,
		"expires_at": expiresAt,
		"created_at": time.Now(),
	})
	return err
}

func IsTokenBlacklisted(collection *mongo.Collection, token string) (bool, error) {
	count, err := collection.CountDocuments(context.TODO(), bson.M{"token": token})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
