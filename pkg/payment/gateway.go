package payment

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VirtualAccount struct {
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	AccountID     string `json:"account_id"`
}

type PaymentGateway interface {
	CreateVirtualAccount(ctx context.Context, ownerID primitive.ObjectID, email, phone, narration string, isPermanent bool, bvn string) (*VirtualAccount, error)
	GetVirtualAccount(ctx context.Context, accountID string) (*VirtualAccount, error)
	DeactivateVirtualAccount(ctx context.Context, accountID string) error
	Transfer(ctx context.Context, fromWalletID, toWalletID primitive.ObjectID, amount float64, reference string) error
}
