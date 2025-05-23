package routes

import (
	"github.com/Gerard-007/ajor_app/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(router *gin.Engine, db *mongo.Collection) {
	// Authentication routes
	router.POST("/login", handlers.LoginHandler(db))
	router.POST("/register", handlers.RegisterHandler(db))

	// User routes
	router.GET("/users/:id", handlers.GetUserByIdHandler(db))
	// router.POST("/forgot-password", handlers.ForgotPasswordHandler(db))
	// router.POST("/reset-password", handlers.ResetPasswordHandler(db))
	// router.POST("/verify-email", handlers.VerifyEmailHandler(db))
	// router.POST("/verify-phone", handlers.VerifyPhoneHandler(db))
	// router.POST("/resend-verification-email", handlers.ResendVerificationEmailHandler(db))
}
