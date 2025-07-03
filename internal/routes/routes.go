package routes

import (
	"github.com/Gerard-007/ajor_app/internal/auth"
	"github.com/Gerard-007/ajor_app/internal/handlers"
	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/services"
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

	notifRepo := repository.NewNotificationRepository(db)
	notifService := services.NewNotificationService(notifRepo)
	notifHandler := handlers.NewNotificationHandler(notifService)

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
		authenticated.POST("/contributions/join", handlers.JoinContributionHandler(db, notifService))
		authenticated.DELETE("/contributions/:id/:user_id", handlers.RemoveMemberHandler(db, notifService))
		authenticated.POST("/contributions/:id/contribute", handlers.RecordContributionHandler(db, notifService))
		authenticated.POST("/contributions/:id/payout", handlers.RecordPayoutHandler(db, notifService))
		authenticated.GET("/notifications", notifHandler.GetAll)
		authenticated.GET("/notifications/unread", notifHandler.GetUnread)
		authenticated.POST("/notifications/mark-read", notifHandler.MarkAsRead)
		authenticated.POST("/notifications/mark-unread", notifHandler.MarkAsUnread)
		authenticated.DELETE("/notifications", notifHandler.Delete)
		authenticated.GET("/admin/contributions", handlers.GetAllContributionsHandler(db))
		authenticated.GET("/contributions/groups/:user_id", handlers.GetUserContributionsByUserIdHandler(db))
		// Collection routes
		authenticated.POST("/contributions/:id/collections", handlers.CreateCollectionHandler(db, notifService))
		authenticated.GET("/contributions/:id/collections", handlers.GetCollectionsHandler(db))
		// Approval routes
		authenticated.PUT("/approvals/:approval_id", handlers.ApprovePayoutHandler(db, notifService))
		authenticated.GET("/approvals", handlers.GetPendingApprovalsHandler(db))
		// Wallet routes
		authenticated.GET("/wallet", handlers.GetUserWalletHandler(db, pg))
		authenticated.POST("/wallet/fund", handlers.FundWalletHandler(db, pg))
		authenticated.GET("/wallet/transactions", handlers.GetUserTransactionsHandler(db))
		authenticated.DELETE("/wallet", handlers.DeleteWalletHandler(db, pg))
		authenticated.POST("/notifications/test", notifHandler.CreateTest)
		authenticated.POST("/wallet/simulate-fund", handlers.SimulateFundWalletHandler(db))
		authenticated.GET("/transactions/:id", handlers.GetTransactionByIdHandler(db))
		authenticated.POST("/users/change-password", handlers.ChangePasswordHandler(db))
		authenticated.POST("/webhook/flutterwave", handlers.FlutterwaveWebhookHandler(db, pg))
	}

	router.GET("/verify-email", handlers.VerifyEmailHandler(db))
}
