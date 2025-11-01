// @title           Project Management and Logging API
// @version         1.0
// @description     API for managing projects and retrieving associated logs with pagination.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	applicationLog "github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log"
	domainLog "github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
	httpRoutes "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http"
	httpControllersLog "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/controller/log"
	sse "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/http/sse"
	repoLog "github.com/rubensantoniorosa2704/LoggingSSE/internal/infrastructure/repository/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/pkg/mongodb"
)

const defaultPort = "8080"

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, relying on environment variables.")
	}

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI and MONGO_DB_NAME must be set.")
	}

	// Connect to MongoDB
	mongoClient, err := mongodb.ConnectSimple(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	log.Println("MongoDB connection successful.")
	// Initialize repository, usecase, and controller with dependency injection
	var logRepo domainLog.LogRepository = repoLog.NewLogRepository(mongoClient.Client, dbName)
	sseServer := sse.NewServer()
	logUsecase := applicationLog.NewLogUsecase(logRepo, sseServer)

	// Register routes and start server
	routerConfig := httpRoutes.RouterConfig{
		LogController: httpControllersLog.NewLogController(logUsecase),
		SSEServer:     sseServer,
	}
	router := httpRoutes.RegisterRoutes(routerConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
