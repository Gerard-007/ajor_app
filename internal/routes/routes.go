package routes

import (
	"github.com/Gerard-007/ajor_app/internal/auth"
	"github.com/Gerard-007/ajor_app/internal/handlers"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(router *gin.Engine, db *mongo.Database, pg payment.PaymentGateway) {
	usersCollection := db.Collection("users")
	// Authentication routes
	router.POST("/login", handlers.LoginHandler(usersCollection))
	router.POST("/register", handlers.RegisterHandler(db, pg))
	router.POST("/logout", handlers.LogoutHandler(db))

	// Authenticated routes
	authenticated := router.Group("/")
	authenticated.Use(auth.AuthMiddleware(db))
	{
		// User routes
		authenticated.GET("/users/:id", handlers.GetUserByIdHandler(db))
		authenticated.GET("/admin/users", handlers.GetAllUsersHandler(usersCollection))
		authenticated.GET("/profile/:id", handlers.GetUserProfileHandler(db))
		authenticated.PUT("/profile/:id", handlers.UpdateUserProfileHandler(db))
		authenticated.PUT("/users/:id", handlers.UpdateUserHandler(db))
		authenticated.DELETE("/users/:id", handlers.DeleteUserHandler(db))
		// Contribution routes
		authenticated.POST("/contributions", handlers.CreateContributionHandler(db, pg))
		authenticated.GET("/contributions/:id", handlers.GetContributionHandler(db))
		authenticated.GET("/contributions/:id/wallet", handlers.GetContributionWalletHandler(db, pg))
		authenticated.GET("/contributions/:id/transactions", handlers.GetContributionTransactionsHandler(db))
		authenticated.GET("/contributions", handlers.GetUserContributionsHandler(db))
		authenticated.PUT("/contributions/:id", handlers.UpdateContributionHandler(db))
		authenticated.POST("/contributions/join", handlers.JoinContributionHandler(db))
		authenticated.DELETE("/contributions/:id/:user_id", handlers.RemoveMemberHandler(db))
		authenticated.POST("/contributions/:id/contribute", handlers.RecordContributionHandler(db))
		authenticated.POST("/contributions/:id/payout", handlers.RecordPayoutHandler(db))
		authenticated.GET("/notifications", handlers.GetUserNotificationsHandler(db))
		authenticated.GET("/admin/contributions", handlers.GetAllContributionsHandler(db))
		// Collection routes
		authenticated.POST("/contributions/:id/collections", handlers.CreateCollectionHandler(db))
		authenticated.GET("/contributions/:id/collections", handlers.GetCollectionsHandler(db))
		// Approval routes
		authenticated.PUT("/approvals/:approval_id", handlers.ApprovePayoutHandler(db))
		authenticated.GET("/approvals", handlers.GetPendingApprovalsHandler(db))
		// Wallet routes
		authenticated.GET("/wallet", handlers.GetUserWalletHandler(db, pg))
		authenticated.POST("/wallet/fund", handlers.FundWalletHandler(db, pg))
		authenticated.GET("/wallet/transactions", handlers.GetUserTransactionsHandler(db))
		authenticated.DELETE("/wallet", handlers.DeleteWalletHandler(db, pg))
	}
}