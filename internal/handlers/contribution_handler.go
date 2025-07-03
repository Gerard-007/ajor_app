package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateContributionHandler(db *mongo.Database, pg payment.PaymentGateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		var contribution models.Contribution
		if err := c.ShouldBindJSON(&contribution); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		err = services.CreateContribution(c.Request.Context(), db, pg, &contribution, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message":      "Contribution created successfully",
			"contribution": contribution,
			"invite_code":  contribution.InviteCode,
		})
	}
}

func GetUserContributionsByUserIdHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		contributions, err := services.GetUserContributions(c.Request.Context(), db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contributions"})
			return
		}

		c.JSON(http.StatusOK, contributions)
	}
}

func GetContributionHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contribution ID"})
			return
		}
		contribution, err := services.GetContribution(c.Request.Context(), db, contributionID, userID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contribution"})
			return
		}
		c.JSON(http.StatusOK, contribution)
	}
}

func GetUserContributionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributions, err := services.GetUserContributions(c.Request.Context(), db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contributions"})
			return
		}
		c.JSON(http.StatusOK, contributions)
	}
}

func UpdateContributionHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		contributionID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contribution ID"})
			return
		}
		var contribution models.Contribution
		if err := c.ShouldBindJSON(&contribution); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		err = services.UpdateContribution(c.Request.Context(), db, contributionID, userID, &contribution)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "only group admin") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contribution"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Contribution updated successfully"})
	}
}

func JoinContributionHandler(db *mongo.Database, notifService *services.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var req struct {
			InviteCode string `json:"invite_code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code is required"})
			return
		}
		contribution, err := services.FindContributionByInviteCode(c.Request.Context(), db, req.InviteCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite code"})
			return
		}
		err = services.JoinContribution(c.Request.Context(), db, notifService, contribution.ID, userID, req.InviteCode)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "already"):
				c.JSON(http.StatusBadRequest, gin.H{"error": "You are already in the group"})
			case strings.Contains(err.Error(), "not found"):
				c.JSON(http.StatusBadRequest, gin.H{"error": "Contribution not found"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the group"})
	}
}

func RemoveMemberHandler(db *mongo.Database, notifService *services.NotificationService) gin.HandlerFunc {
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
		userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		err = services.RemoveMember(c.Request.Context(), db, notifService, contributionID, userID, groupAdminID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "only group admin") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
	}
}

func RecordContributionHandler(db *mongo.Database, notifService *services.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getAuthUserID(c)
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
			Amount        float64              `json:"amount"`
			PaymentMethod models.PaymentMethod `json:"payment_method"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		err = services.RecordContribution(c.Request.Context(), db, notifService, contributionID, userID, request.Amount, request.PaymentMethod)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "amount mismatch") || strings.Contains(err.Error(), "insufficient balance") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record contribution"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Contribution recorded successfully"})
	}
}

func RecordPayoutHandler(db *mongo.Database, notifService *services.NotificationService) gin.HandlerFunc {
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
			UserID        primitive.ObjectID   `json:"user_id"`
			Amount        float64              `json:"amount"`
			PaymentMethod models.PaymentMethod `json:"payment_method"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		err = services.RecordPayout(c.Request.Context(), db, notifService, contributionID, request.UserID, groupAdminID, request.Amount, request.PaymentMethod)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "only group admin") || strings.Contains(err.Error(), "insufficient balance") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payout"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Payout recorded, pending approval"})
	}
}

func GetAllContributionsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only system admins can view all contributions"})
			return
		}
		contributions, err := services.GetAllContributions(db, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, contributions)
	}
}

func getAuthUserID(c *gin.Context) (primitive.ObjectID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return primitive.ObjectID{}, errors.New("user not authenticated")
	}
	return primitive.ObjectIDFromHex(userIDStr.(string))
}
