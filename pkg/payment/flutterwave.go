package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FlutterwaveGateway struct {
	APIKey  string
	BaseURL string
}

type createVirtualAccountRequest struct {
	Email        string   `json:"email"`
	Currency     string   `json:"currency"`
	IsPermanent  bool     `json:"is_permanent"`
	Narration    string   `json:"narration"`
	PhoneNumber  string   `json:"phone"`
	BVN          string   `json:"bvn,omitempty"`
	TxRef        string   `json:"tx_ref"`
	Amount       *float64 `json:"amount,omitempty"`
}

type createVirtualAccountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountNumber string `json:"account_number"`
		BankName      string `json:"bank_name"`
		OrderRef      string `json:"order_ref"`
	} `json:"data"`
}

type getVirtualAccountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountNumber string `json:"account_number"`
		BankName      string `json:"bank_name"`
		OrderRef      string `json:"id"`
	} `json:"data"`
}

type deactivateVirtualAccountRequest struct {
	Status string `json:"status"`
}

type fundVirtualAccountRequest struct {
	TxRef       string  `json:"tx_ref"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number"`
	Narration   string  `json:"narration"`
	RedirectURL string  `json:"redirect_url"`
}

type fundVirtualAccountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Link          string `json:"link"`
		TransactionID string `json:"id"`
	} `json:"data"`
}

type verifyTransactionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID            int     `json:"id"`
		TxRef         string  `json:"tx_ref"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		Status        string  `json:"status"`
	} `json:"data"`
}

func NewFlutterwaveGateway() *FlutterwaveGateway {
	apiKey := os.Getenv("FLW_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("FLW_SECRET_KEY environment variable not set")
	}
	log.Printf("Flutterwave API Key: %s", apiKey)
	return &FlutterwaveGateway{
		APIKey:  apiKey,
		BaseURL: "https://api.flutterwave.com/v3",
	}
}

func (f *FlutterwaveGateway) CreateVirtualAccount(ctx context.Context, ownerID primitive.ObjectID, email, phone, narration string, isPermanent bool, bvn string, amount float64) (*VirtualAccount, error) {
	url := f.BaseURL + "/virtual-account-numbers"
	txRef := fmt.Sprintf("ajor-%s-%d", ownerID.Hex(), time.Now().Unix())

	var amountPtr *float64
	if amount > 0 {
		amountPtr = &amount
	}

	payload := createVirtualAccountRequest{
		Email:       email,
		Currency:    "NGN",
		IsPermanent: isPermanent,
		Narration:   narration,
		PhoneNumber: phone,
		BVN:         bvn,
		TxRef:       txRef,
		Amount:      amountPtr,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response createVirtualAccountResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("failed to create virtual account: %s", response.Message)
	}

	return &VirtualAccount{
		AccountNumber: response.Data.AccountNumber,
		BankName:      response.Data.BankName,
		AccountID:     response.Data.OrderRef,
	}, nil
}

func (f *FlutterwaveGateway) GetVirtualAccount(ctx context.Context, accountID string) (*VirtualAccount, error) {
	url := fmt.Sprintf("%s/virtual-account-numbers/%s", f.BaseURL, accountID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response getVirtualAccountResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("failed to get virtual account: %s", response.Message)
	}

	return &VirtualAccount{
		AccountNumber: response.Data.AccountNumber,
		BankName:      response.Data.BankName,
		AccountID:     response.Data.OrderRef,
	}, nil
}

func (f *FlutterwaveGateway) DeactivateVirtualAccount(ctx context.Context, accountID string) error {
	url := fmt.Sprintf("%s/virtual-account-numbers/%s", f.BaseURL, accountID)

	payload := deactivateVirtualAccountRequest{Status: "inactive"}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != "success" {
		return fmt.Errorf("failed to deactivate virtual account: %s", response.Message)
	}

	return nil
}

func (f *FlutterwaveGateway) FundVirtualAccount(ctx context.Context, accountID string, req FundingRequest) (*TransactionResponse, error) {
	url := f.BaseURL + "/payments"

	payload := fundVirtualAccountRequest{
		TxRef:       req.TxRef,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Narration:   req.Narration,
		RedirectURL: "https://your-app.com/redirect", // Replace with actual redirect URL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+f.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response fundVirtualAccountResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("failed to initiate funding: %s", response.Message)
	}

	return &TransactionResponse{
		TransactionID: response.Data.TransactionID,
		Status:        "pending",
		Amount:        req.Amount,
	}, nil
}

func (f *FlutterwaveGateway) VerifyTransaction(ctx context.Context, transactionID string) (*TransactionResponse, error) {
	url := fmt.Sprintf("%s/transactions/%s/verify", f.BaseURL, transactionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response verifyTransactionResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("failed to verify transaction: %s", response.Message)
	}

	return &TransactionResponse{
		TransactionID: fmt.Sprintf("%d", response.Data.ID),
		Status:        response.Data.Status,
		Amount:        response.Data.Amount,
	}, nil
}

func (f *FlutterwaveGateway) Transfer(ctx context.Context, fromWalletID, toWalletID primitive.ObjectID, amount float64, reference string) error {
	// Placeholder: Implement Flutterwave transfer API
	// https://developer.flutterwave.com/reference/endpoints/transfers
	return errors.New("transfer not implemented")
}