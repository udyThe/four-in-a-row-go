package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/4-in-a-row/internal/api"
	"github.com/yourusername/4-in-a-row/internal/config"
	"github.com/yourusername/4-in-a-row/internal/database"
	"github.com/yourusername/4-in-a-row/internal/game"
	"github.com/yourusername/4-in-a-row/internal/kafka"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Kafka producer (optional)
	var kafkaProducer *kafka.Producer
	if cfg.KafkaEnabled {
		kafkaProducer, err = kafka.NewProducer(cfg.KafkaBrokers)
		if err != nil {
			log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
			kafkaProducer = nil
		}
		if kafkaProducer != nil {
			defer kafkaProducer.Close()
			log.Println("Kafka analytics enabled")
		}
	} else {
		log.Println("Kafka analytics disabled (set KAFKA_ENABLED=true to enable)")
	}

	// Initialize game manager
	gameManager := game.NewManager(db, kafkaProducer)

	// Start metrics emitter (sends system metrics to Kafka every 60 seconds)
	gameManager.StartMetricsEmitter()

	// Initialize matchmaking (do NOT start it yet)
	matchmaker := game.NewMatchmaker(gameManager)

	// Initialize API server (this registers callbacks the matchmaker relies on)
	server := api.NewServer(cfg, gameManager, matchmaker, db)

	// Now start the matchmaker loop after server (and callbacks) are ready
	go matchmaker.Run()

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
