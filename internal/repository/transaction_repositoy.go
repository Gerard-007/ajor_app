package repository

import (
	"context"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTransaction(ctx context.Context, db *mongo.Database, transaction *models.Transaction) error {
	collection := db.Collection("transactions")
	transaction.CreatedAt = time.Now()
	result, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return err
	}
	transaction.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func GetUserTransactions(ctx context.Context, db *mongo.Database, userID, contributionID primitive.ObjectID) ([]*models.Transaction, error) {
	var wallet models.Wallet
	err := db.Collection("wallets").FindOne(ctx, bson.M{"owner_id": userID, "type": models.WalletTypeUser}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"$or": []bson.M{
			{"from_wallet": wallet.ID},
			{"to_wallet": wallet.ID},
		},
		"contribution_id": contributionID,
	}
	var transactions []*models.Transaction
	cursor, err := db.Collection("transactions").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var transaction models.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}
	return transactions, nil
}