package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateCollection(ctx context.Context, db *mongo.Database, contributionID, collectorID, groupAdminID primitive.ObjectID, collectionDate *time.Time) error {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return err
	}

	if contribution.GroupAdmin != groupAdminID {
		return errors.New("only group admin can create collections")
	}

	if !containsUser(contribution.YetToCollectMembers, collectorID) {
		return errors.New("collector not in contribution")
	}

	var finalCollectionDate time.Time
	if collectionDate == nil {
		finalCollectionDate = computeCollectionDate(contribution.Cycle, contribution.CreatedAt)
	} else {
		finalCollectionDate = *collectionDate
	}

	collection := &models.Collection{
		ContributionID: contributionID,
		Collector:      collectorID,
		CollectionDate: finalCollectionDate,
	}

	err = repository.CreateCollection(ctx, db, collection)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		UserID:         collectorID,
		ContributionID: contributionID,
		Message:        fmt.Sprintf("You are scheduled to collect for group: %s on %s", contribution.Name, finalCollectionDate.Format("2006-01-02")),
		Type:           models.NotificationInfo,
	}
	return repository.CreateNotification(ctx, db, notification)
}

func GetCollections(ctx context.Context, db *mongo.Database, contributionID, userID primitive.ObjectID) ([]*models.Collection, error) {
	contribution, err := repository.GetContributionByID(ctx, db, contributionID)
	if err != nil {
		return nil, err
	}

	if contribution.GroupAdmin != userID &&
		!containsUser(contribution.YetToCollectMembers, userID) &&
		!containsUser(contribution.AlreadyCollectedMembers, userID) {
		return nil, errors.New("unauthorized access to contribution")
	}

	return repository.GetCollectionsByContribution(ctx, db, contributionID)
}

func computeCollectionDate(cycle models.ContributionCycle, baseDate time.Time) time.Time {
	switch cycle {
	case models.CycleDaily:
		return baseDate.Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleWeekly:
		daysUntilSunday := (7 - int(baseDate.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		return baseDate.Add(time.Duration(daysUntilSunday) * 24 * time.Hour).Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleMonthly:
		nextMonth := baseDate.AddDate(0, 1, 0)
		lastDay := nextMonth.AddDate(0, 0, -nextMonth.Day())
		return lastDay.Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case models.CycleYearly:
		return time.Date(baseDate.Year(), 12, 31, 23, 59, 59, 0, baseDate.Location())
	default:
		return time.Now()
	}
}