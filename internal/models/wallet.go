package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WalletType string

const (
	WalletTypeUser         WalletType = "user"
	WalletTypeContribution WalletType = "contribution"
)

type Wallet struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OwnerID              primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	Type                 WalletType         `json:"type" bson:"type"`
	Balance              float64            `json:"balance" bson:"balance"`
	VirtualAccountID     string             `json:"virtual_account_id" bson:"virtual_account_id"`
	VirtualAccountNumber string             `json:"virtual_account_number" bson:"virtual_account_number"`
	VirtualBankName      string             `json:"virtual_bank_name" bson:"virtual_bank_name"`
	CreatedAt            time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" bson:"updated_at"`
}
