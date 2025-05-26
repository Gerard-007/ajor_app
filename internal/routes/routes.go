package routes

import (
	"github.com/Gerard-007/ajor_app/internal/auth"
	"github.com/Gerard-007/ajor_app/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(router *gin.Engine, db *mongo.Database) {
	usersCollection := db.Collection("users")
	// Authentication routes
	router.POST("/login", handlers.LoginHandler(usersCollection))
	router.POST("/register", handlers.RegisterHandler(db))
	router.POST("/logout", handlers.LogoutHandler(db))

	// User routes
	authenticated := router.Group("/")
	authenticated.Use(auth.AuthMiddleware(db))
	authenticated.GET("/users/:id", handlers.GetUserByIdHandler(usersCollection))
	authenticated.GET("/profile/:id", handlers.GetUserProfileHandler(db))
	authenticated.PUT("/profile/:id", handlers.UpdateUserProfileHandler(db))
	// router.DELETE("/users/:id", handlers.DeleteUserHandler(db))
	// router.GET("/users", handlers.GetAllUsersHandler(db))

	// router.POST("/forgot-password", handlers.ForgotPasswordHandler(db))
	// router.POST("/reset-password", handlers.ResetPasswordHandler(db))
	// router.POST("/verify-email", handlers.VerifyEmailHandler(db))
	// router.POST("/verify-phone", handlers.VerifyPhoneHandler(db))
	// router.POST("/resend-verification-email", handlers.ResendVerificationEmailHandler(db))
}
