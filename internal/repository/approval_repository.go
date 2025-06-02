package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateApproval(ctx context.Context, db *mongo.Database, approval *models.Approval) error {
	collection := db.Collection("approvals")
	approval.CreatedAt = time.Now()
	approval.UpdatedAt = time.Now()
	_, err := collection.InsertOne(ctx, approval)
	return err
}

func UpdateApproval(ctx context.Context, db *mongo.Database, approvalID primitive.ObjectID, status models.ApprovalStatus) error {
	filter := bson.M{"_id": approvalID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	result, err := db.Collection("approvals").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("approval not found")
	}
	return nil
}

func GetPendingApprovals(ctx context.Context, db *mongo.Database, approverID primitive.ObjectID) ([]*models.Approval, error) {
	var approvals []*models.Approval
	cursor, err := db.Collection("approvals").Find(ctx, bson.M{
		"approver_id": approverID,
		"status":      models.ApprovalPending,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var approval models.Approval
		if err := cursor.Decode(&approval); err != nil {
			return nil, err
		}
		approvals = append(approvals, &approval)
	}
	return approvals, nil
}