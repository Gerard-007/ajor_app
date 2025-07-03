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

func UpdateTransactionStatus(ctx context.Context, db *mongo.Database, transactionID primitive.ObjectID, status models.TransactionStatus) error {
	collection := db.Collection("transactions")
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": transactionID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
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

func GetTransactions(ctx context.Context, db *mongo.Database, filter bson.M) ([]models.Transaction, error) {
	collection := db.Collection("transactions")
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return []models.Transaction{}, nil
	}

	return transactions, nil
}

func GetTransactionByTxRef(ctx context.Context, db *mongo.Database, txRef string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := db.Collection("transactions").FindOne(ctx, bson.M{"tx_ref": txRef}).Decode(&transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}