package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateWallet(db *mongo.Database, wallet *models.Wallet) error {
	collection := db.Collection("wallets")
	ctx := context.Background()

	// Insert wallet with pre-assigned ID
	result, err := collection.InsertOne(ctx, wallet)
	if err != nil {
		return err
	}

	// Ensure wallet.ID is set
	if wallet.ID.IsZero() {
		wallet.ID = result.InsertedID.(primitive.ObjectID)
	}

	return nil
}

func GetWalletByUserID(db *mongo.Database, owner_id primitive.ObjectID) (*models.Wallet, error) {
	var wallet models.Wallet
	err := db.Collection("wallets").FindOne(context.TODO(), bson.M{"owner_id": owner_id}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func GetWalletByID(db *mongo.Database, wallet_id primitive.ObjectID) (*models.Wallet, error) {
	var wallet models.Wallet
	err := db.Collection("wallets").FindOne(context.TODO(), bson.M{"_id": wallet_id}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func GetContributionWalletByID(ctx context.Context, db *mongo.Database, walletID primitive.ObjectID) (*models.Wallet, error) {
	var wallet models.Wallet
	collection := db.Collection("wallets")
	err := collection.FindOne(ctx, bson.M{"_id": walletID}).Decode(&wallet)
	if err != nil {
		log.Printf("Error fetching wallet ID %s: %v", walletID.Hex(), err)
		return nil, err
	}
	return &wallet, nil
}

func GetWalletByContributionID(db *mongo.Database, contributionID primitive.ObjectID) (*models.Wallet, error) {
	// First, fetch the contribution to get its wallet_id
	var contribution models.Contribution
	err := db.Collection("contributions").FindOne(context.TODO(), bson.M{"_id": contributionID}).Decode(&contribution)
	if err != nil {
		return nil, err
	}
	// Now fetch the wallet by its _id
	var wallet models.Wallet
	err = db.Collection("wallets").FindOne(context.TODO(), bson.M{"_id": contribution.WalletID}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func UpdateWalletBalance(db *mongo.Database, walletID primitive.ObjectID, amount float64, isCredit bool) error {
	filter := bson.M{"_id": walletID}
	var update bson.M
	if isCredit {
		update = bson.M{
			"$inc": bson.M{"balance": amount},
			"$set": bson.M{"updated_at": time.Now()},
		}
	} else {
		update = bson.M{
			"$inc": bson.M{"balance": -amount},
			"$set": bson.M{"updated_at": time.Now()},
		}
	}
	result, err := db.Collection("wallets").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("wallet not found")
	}
	return nil
}

func UpdateWalletVirtualAccount(db *mongo.Database, walletID primitive.ObjectID, virtualAccountNumber, accountID, accountBank string) error {
	collection := db.Collection("wallets")
	ctx := context.Background()

	update := bson.M{
		"$set": bson.M{
			"virtual_account_number": virtualAccountNumber,
			"virtual_account_id":     accountID,
			"virtual_bank_name":      accountBank,
			"updated_at":             time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": walletID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("wallet not found")
	}

	return nil
}

func DeleteWallet(db *mongo.Database, walletID primitive.ObjectID) error {
	result, err := db.Collection("wallets").DeleteOne(context.TODO(), bson.M{"_id": walletID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("wallet not found")
	}
	return nil
}