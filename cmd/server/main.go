package main

import (
	"log"
	"os"
	"time"

	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/routes"
	"github.com/Gerard-007/ajor_app/pkg/jobs"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/Gerard-007/ajor_app/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := repository.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	pg := payment.NewFlutterwaveGateway()

	server := gin.Default()

	// CORS middleware configuration
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Configure trusted proxies
	if err := server.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatal("Failed to set trusted proxies:", err)
	}

	routes.InitRoutes(server, db, pg)

	// Start cron job
	c := cron.New()
	notifRepo := repository.NewNotificationRepository(db)
	notifService := services.NewNotificationService(notifRepo)
	_, err = c.AddFunc("0 0 * * *", func() { // Runs daily at midnight
		if err := jobs.ProcessCollections(db, notifService); err != nil {
			log.Printf("Error processing collections: %v", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	c.Start()
	defer c.Stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Starting server on port %s", port)
	server.Run(":" + port)
}
