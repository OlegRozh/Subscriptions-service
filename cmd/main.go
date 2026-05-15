package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/OlegRozh/subscriptions-service/internal/server"
	"github.com/OlegRozh/subscriptions-service/internal/storage"
	"github.com/OlegRozh/subscriptions-service/migrations"
	"github.com/joho/godotenv"
)

// @title Subscriptions Service API
// @version 1.0
// @description API для управления подписками пользователей
// @host localhost:8080
// @BasePath /

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Start app")
	//Uncomment to local run and test
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found, using system environment variables")
	}
	//Comment to run docker compose
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	logger.Info("DEBUG", "DATABASE_URL", databaseURL)

	if err := migrations.Up(databaseURL, "./migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	logger.Info("Migrations applied")
	db, err := storage.NewStorage(databaseURL)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	defer db.Close()
	srv := server.New(db, logger, "8080")
	if err := srv.Run(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
