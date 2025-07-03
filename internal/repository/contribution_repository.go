package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateContribution(ctx context.Context, db *mongo.Database, contribution *models.Contribution) error {
	collection := db.Collection("contributions")
	contribution.InviteCode = uuid.New().String()
	contribution.CreatedAt = time.Now()
	contribution.UpdatedAt = time.Now()
	result, err := collection.InsertOne(ctx, contribution)
	if err != nil {
		return err
	}
	contribution.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func GetAllContributions(db *mongo.Database) ([]*models.Contribution, error) {
	var contributions []*models.Contribution
	cursor, err := db.Collection("contributions").Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var contribution models.Contribution
		if err := cursor.Decode(&contribution); err != nil {
			return nil, err
		}
		contributions = append(contributions, &contribution)
	}
	return contributions, nil
}

func GetContributionByID(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*models.Contribution, error) {
	var contribution models.Contribution
	err := db.Collection("contributions").FindOne(ctx, bson.M{"_id": id}).Decode(&contribution)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("contribution not found")
	}
	if err != nil {
		return nil, err
	}
	return &contribution, nil
}

func GetContributionsByUserID(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) ([]*models.Contribution, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"yet_to_collect_members": userID},
			{"already_collected_members": userID},
		},
	}

	cursor, err := db.Collection("contributions").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var contributions []*models.Contribution
	for cursor.Next(ctx) {
		var c models.Contribution
		if err := cursor.Decode(&c); err != nil {
			return nil, err
		}
		contributions = append(contributions, &c)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return contributions, nil
}

func GetContributionByInviteCode(ctx context.Context, db *mongo.Database, inviteCode string) (*models.Contribution, error) {
	var contribution models.Contribution
	err := db.Collection("contributions").FindOne(ctx, bson.M{"invite_code": inviteCode}).Decode(&contribution)

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("contribution not found")
	}

	if err != nil {
		return nil, err
	}

	return &contribution, nil
}

func GetContributionsByUser(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) ([]*models.Contribution, error) {
	var contributions []*models.Contribution
	cursor, err := db.Collection("contributions").Find(ctx, bson.M{
		"$or": []bson.M{
			{"group_admin": userID},
			{"yet_to_collect_members": userID},
			{"already_collected_members": userID},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var contribution models.Contribution
		if err := cursor.Decode(&contribution); err != nil {
			return nil, err
		}
		contributions = append(contributions, &contribution)
	}
	return contributions, nil
}

func UpdateContribution(ctx context.Context, db *mongo.Database, id primitive.ObjectID, contribution *models.Contribution) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"name":                contribution.Name,
			"description":         contribution.Description,
			"cycle":               contribution.Cycle,
			"amount":              contribution.Amount,
			"cycle_count":         contribution.CycleCount,
			"collection_day":      contribution.CollectionDay,
			"collection_deadline": contribution.CollectionDeadline,
			"type":                contribution.Type,
			"penalty_amount":      contribution.PenaltyAmount,
			"updated_at":          time.Now(),
		},
	}
	result, err := db.Collection("contributions").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("contribution not found")
	}
	return nil
}

func DecrementCycleCount(ctx context.Context, db *mongo.Database, contributionID primitive.ObjectID) error {
	filter := bson.M{"_id": contributionID}
	update := bson.M{
		"$inc": bson.M{"cycle_count": -1},
		"$set": bson.M{"updated_at": time.Now()},
	}
	result, err := db.Collection("contributions").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("contribution not found")
	}
	return nil
}

func RemoveMember(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": contributionID}
	update := bson.M{
		"$pull": bson.M{
			"yet_to_collect_members":    userID,
			"already_collected_members": userID,
		},
		"$set": bson.M{"updated_at": time.Now()},
	}
	result, err := db.Collection("contributions").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("contribution not found")
	}
	return nil
}

func JoinContribution(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": contributionID}
	update := bson.M{
		"$addToSet": bson.M{"yet_to_collect_members": userID},
		"$set":      bson.M{"updated_at": time.Now()},
	}
	result, err := db.Collection("contributions").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("contribution not found")
	}
	return nil
}

func MarkMemberCollected(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": contributionID}
	update := bson.M{
		"$pull":     bson.M{"yet_to_collect_members": userID},
		"$addToSet": bson.M{"already_collected_members": userID},
		"$set":      bson.M{"updated_at": time.Now()},
	}
	result, err := db.Collection("contributions").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("contribution not found")
	}
	return nil
}

func UpdateContributionMemberUsernames(ctx context.Context, db *mongo.Database, contributionID primitive.ObjectID, memberUsernames map[primitive.ObjectID]string) error {
	collection := db.Collection("contributions")
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": contributionID},
		bson.M{"$set": bson.M{"member_usernames": memberUsernames}},
	)
	return err
}
