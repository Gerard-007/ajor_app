package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApprovePayoutHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		approverID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		approvalID, err := primitive.ObjectIDFromHex(c.Param("approval_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval ID"})
			return
		}
		var request struct {
			Approve bool `json:"approve"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		err = services.ApprovePayout(c.Request.Context(), db, approvalID, approverID, request.Approve)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "already processed") {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process approval"})
			return
		}
		status := "rejected"
		if request.Approve {
			status = "approved"
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Payout %s successfully", status)})
	}
}

func GetPendingApprovalsHandler(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		approverID, err := getAuthUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		approvals, err := services.GetPendingApprovals(c.Request.Context(), db, approverID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending approvals"})
			return
		}
		c.JSON(http.StatusOK, approvals)
	}
}