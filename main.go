package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"healthy_pay_backend/internal/config"
	"healthy_pay_backend/internal/database"
	"healthy_pay_backend/internal/middleware"
	"healthy_pay_backend/internal/routes"
	"healthy_pay_backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	
	db, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Start automated transaction queue
	queue := services.NewTransactionQueue(db)
	queue.Start()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		queue.Stop()
		os.Exit(0)
	}()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	routes.SetupRoutes(r, db)

	log.Fatal(r.Run(":" + cfg.Port))
}
