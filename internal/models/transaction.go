package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionType string
type TransactionStatus string
type PaymentMethod string

const (
	TransactionContribution TransactionType = "contribution"
	TransactionPayout       TransactionType = "payout"
	TransactionWallet       TransactionType = "wallet"
)

const (
	StatusPending TransactionStatus = "pending"
	StatusSuccess TransactionStatus = "success"
	StatusFailed  TransactionStatus = "failed"
)

const (
	PaymentBankTransfer PaymentMethod = "bank_transfer"
	PaymentMobileMoney  PaymentMethod = "mobile_money"
	PaymentCash         PaymentMethod = "cash"
	PaymentWallet         PaymentMethod = "wallet"
)

type Transaction struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FromWallet     primitive.ObjectID `json:"from_wallet" bson:"from_wallet"`
	ToWallet       primitive.ObjectID `json:"to_wallet" bson:"to_wallet"`
	Amount         float64            `json:"amount" bson:"amount"`
	Type           TransactionType    `json:"type" bson:"type"`
	Date           time.Time          `json:"date" bson:"date"`
	PaymentMethod  PaymentMethod      `json:"payment_method" bson:"payment_method"`
	Status         TransactionStatus  `json:"status" bson:"status"`
	ContributionID primitive.ObjectID `json:"contribution_id" bson:"contribution_id"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}
