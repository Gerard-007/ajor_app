package payment

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VirtualAccount struct {
	AccountNumber string
	AccountID     string
	BankName      string
}

type FundingRequest struct {
	Email        string
	Amount       float64
	TxRef        string
	Currency     string
	IsPermanent  bool
	Narration    string
	PhoneNumber  string
}

type TransactionResponse struct {
	TransactionID string
	Status        string
	Amount        float64
}

type PaymentGateway interface {
	CreateVirtualAccount(ctx context.Context, ownerID primitive.ObjectID, email, phone, narration string, isPermanent bool, bvn string, amount float64) (*VirtualAccount, error)
	GetVirtualAccount(ctx context.Context, accountID string) (*VirtualAccount, error)
	DeactivateVirtualAccount(ctx context.Context, accountID string) error
	FundVirtualAccount(ctx context.Context, accountID string, req FundingRequest) (*TransactionResponse, error)
	VerifyTransaction(ctx context.Context, transactionID string) (*TransactionResponse, error)
}