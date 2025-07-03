package jobs

import (
	"context"
	"log"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProcessCollections checks for collections due today and processes them (e.g., sends notifications).
func ProcessCollections(db *mongo.Database, notificationService *services.NotificationService) error {
	ctx := context.Background()
	collectionColl := db.Collection("collections")

	// Find collections due today
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	filter := bson.M{
		"collection_date": bson.M{
			"$gte": today,
			"$lt":  tomorrow,
		},
	}

	cursor, err := collectionColl.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var collection models.Collection
		if err := cursor.Decode(&collection); err != nil {
			log.Printf("Failed to decode collection: %v", err)
			continue
		}

		// Get the associated contribution
		contribution, err := repository.GetContributionByID(ctx, db, collection.ContributionID)
		if err != nil {
			log.Printf("Failed to get contribution %s: %v", collection.ContributionID.Hex(), err)
			continue
		}

		// Use NotificationService
		n := &models.Notification{
			UserID:  collection.Collector,
			Type:    "collection_due",
			Title:   "Collection Due Today",
			Message: "Reminder: Collection due today for group: " + contribution.Name,
			Meta:    map[string]interface{}{ "group": contribution.Name, "date": today.Format("2006-01-02") },
		}
		if err := notificationService.Create(ctx, n); err != nil {
			log.Printf("Failed to create notification for user %s: %v", collection.Collector.Hex(), err)
			continue
		}

		log.Printf("Processed collection %s for contribution %s", collection.ID.Hex(), contribution.Name)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}