package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateWallet(db *mongo.Database, wallet *models.Wallet) error {
	collection := db.Collection("wallets")
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	_, err := collection.InsertOne(context.TODO(), wallet)
	return err
}

func GetWalletByID(db *mongo.Database, id primitive.ObjectID) (*models.Wallet, error) {
	var wallet models.Wallet
	err := db.Collection("wallets").FindOne(context.TODO(), bson.M{"_id": id}).Decode(&wallet)
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

func UpdateWalletVirtualAccount(db *mongo.Database, walletID primitive.ObjectID, va *payment.VirtualAccount) error {
	filter := bson.M{"_id": walletID}
	update := bson.M{
		"$set": bson.M{
			"virtual_account_id":     va.AccountID,
			"virtual_account_number": va.AccountNumber,
			"virtual_bank_name":      va.BankName,
			"updated_at":             time.Now(),
		},
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