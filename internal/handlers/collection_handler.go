package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateCollectionHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupAdminID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contribution ID"})
			return
		}
		var request struct {
			CollectorID    string     `json:"collector_id"`
			CollectionDate *time.Time `json:"collection_date,omitempty"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		collectorID, err := primitive.ObjectIDFromHex(request.CollectorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collector ID"})
			return
		}
		err = services.CreateCollection(c.Request.Context(), db, contributionID, collectorID, groupAdminID, request.CollectionDate)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "only group admin") || strings.Contains(err.Error(), "collector not in contribution") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Collection created successfully"})
	}
}

func GetCollectionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		collections, err := services.GetCollections(c.Request.Context(), db, contributionID, userID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collections"})
			return
		}
		c.JSON(http.StatusOK, collections)
	}
}