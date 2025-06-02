package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type Approval struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TransactionID  primitive.ObjectID `json:"transaction_id" bson:"transaction_id"`
	ApproverID     primitive.ObjectID `json:"approver_id" bson:"approver_id"`
	Status         ApprovalStatus     `json:"status" bson:"status"`
	ContributionID primitive.ObjectID `json:"contribution_id" bson:"contribution_id"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}