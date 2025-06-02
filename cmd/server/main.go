package main

import (
	"log"
	"os"

	"github.com/Gerard-007/ajor_app/internal/repository"
	"github.com/Gerard-007/ajor_app/internal/routes"
	"github.com/Gerard-007/ajor_app/pkg/jobs"
	"github.com/Gerard-007/ajor_app/pkg/payment"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
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
	routes.InitRoutes(server, db, pg)
	
	// Start cron job
	c := cron.New()
	_, err = c.AddFunc("0 0 * * *", func() { // Runs daily at midnight
		if err := jobs.ProcessCollections(db); err != nil {
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
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)	
	server.Run(":" + port)
}