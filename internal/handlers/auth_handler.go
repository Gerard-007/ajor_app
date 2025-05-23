package handlers

import (
	"net/http"

	"github.com/Gerard-007/ajor_app/internal/models"
	"github.com/Gerard-007/ajor_app/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func LoginHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement login logic here
		// var user models.User
	}
}

func RegisterHandler(db *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement registration logic here
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := services.RegisterUser(db, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "User created successfully", "user": user})

		// Example: Insert user directly into MongoDB collection
		// _, err := db.InsertOne(c, user)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }

		// c.JSON(http.StatusCreated, gin.H{"status": "success", "user": user})
	}
}
