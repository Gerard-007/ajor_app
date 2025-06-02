package repository

import (
	"context"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateCollection(ctx context.Context, db *mongo.Database, collection *models.Collection) error {
	coll := db.Collection("collections")
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()
	_, err := coll.InsertOne(ctx, collection)
	return err
}

func GetCollectionsByContribution(ctx context.Context, db *mongo.Database, contributionID primitive.ObjectID) ([]*models.Collection, error) {
	var collections []*models.Collection
	cursor, err := db.Collection("collections").Find(ctx, bson.M{"contribution_id": contributionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var collection models.Collection
		if err := cursor.Decode(&collection); err != nil {
			return nil, err
		}
		collections = append(collections, &collection)
	}
	return collections, nil
}