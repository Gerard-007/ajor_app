package repository

import (
	"context"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateProfile(db *mongo.Database, profile *models.Profile) error {
	collection := db.Collection("profiles")
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	_, err := collection.InsertOne(context.TODO(), profile)
	return err
}

func GetUserProfile(collection *mongo.Collection, userID primitive.ObjectID) (*models.Profile, error) {
	var profile models.Profile
	err := collection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func UpdateUserProfile(collection *mongo.Collection, userID primitive.ObjectID, profileUpdate *models.Profile) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"bio":         profileUpdate.Bio,
			"location":    profileUpdate.Location,
			"profile_pic": profileUpdate.ProfilePic,
			"updated_at":  time.Now(),
		},
	}
	_, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	return err
}

func UpdateUserProfilePicture(collection *mongo.Collection, userID primitive.ObjectID, picturePath string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"profile_pic": picturePath,
			"updated_at":  time.Now(),
		},
	}
	_, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	return err
}
