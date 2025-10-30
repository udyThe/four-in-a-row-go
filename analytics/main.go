package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	kafka "github.com/segmentio/kafka-go"
)

type GameEvent struct {
	EventType string  `json:"event_type"`
	GameID    string  `json:"game_id"`
	Player1   string  `json:"player1,omitempty"`
	Player2   string  `json:"player2,omitempty"`
	Player    string  `json:"player"`
	Winner    string  `json:"winner,omitempty"`
	Result    string  `json:"result,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
	Column    int     `json:"column,omitempty"`
	Row       int     `json:"row,omitempty"`
	Timestamp int64   `json:"timestamp"`
}

type AnalyticsService struct {
	reader *kafka.Reader
	db     *pgxpool.Pool
}

func main() {
	log.Println("Starting analytics service...")

	// Get configuration from environment
	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/four_in_a_row?sslmode=disable")

	// Connect to database
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize analytics tables
	if err := initAnalyticsTables(ctx, db); err != nil {
		log.Fatalf("Failed to initialize analytics tables: %v", err)
	}

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "game-events",
		GroupID:  "analytics-consumer",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	service := &AnalyticsService{
		reader: reader,
		db:     db,
	}

	// Start consuming messages
	go service.consumeMessages()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down analytics service...")
}

func (s *AnalyticsService) consumeMessages() {
	ctx := context.Background()

	for {
		msg, err := s.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event GameEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		if err := s.processEvent(ctx, &event); err != nil {
			log.Printf("Error processing event: %v", err)
		} else {
			log.Printf("Processed event: %s for game %s", event.EventType, event.GameID)
		}
	}
}

func (s *AnalyticsService) processEvent(ctx context.Context, event *GameEvent) error {
	switch event.EventType {
	case "game_started":
		return s.recordGameStarted(ctx, event)
	case "move_made":
		return s.recordMove(ctx, event)
	case "game_finished":
		return s.recordGameFinished(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
	}
	return nil
}

func (s *AnalyticsService) recordGameStarted(ctx context.Context, event *GameEvent) error {
	query := `
		INSERT INTO analytics_games (game_id, player1, player2, started_at)
		VALUES ($1, $2, $3, to_timestamp($4))
		ON CONFLICT (game_id) DO NOTHING
	`
	_, err := s.db.Exec(ctx, query, event.GameID, event.Player1, event.Player2, event.Timestamp)

	if err == nil {
		s.updateHourlyStats(ctx, event.Timestamp, "game_started")
		s.updateDailyStats(ctx, event.Timestamp, "game_started")
	}

	return err
}

func (s *AnalyticsService) recordMove(ctx context.Context, event *GameEvent) error {
	query := `
		INSERT INTO analytics_moves (game_id, player, col_index, row_index, move_time)
		VALUES ($1, $2, $3, $4, to_timestamp($5))
	`
	_, err := s.db.Exec(ctx, query, event.GameID, event.Player, event.Column, event.Row, event.Timestamp)

	if err == nil {
		s.updateHourlyStats(ctx, event.Timestamp, "move_made")
		s.updateDailyStats(ctx, event.Timestamp, "move_made")
	}

	return err
}

func (s *AnalyticsService) recordGameFinished(ctx context.Context, event *GameEvent) error {
	query := `
		UPDATE analytics_games
		SET winner = $1, result = $2, duration = $3, finished_at = to_timestamp($4)
		WHERE game_id = $5
	`
	_, err := s.db.Exec(ctx, query, event.Winner, event.Result, event.Duration, event.Timestamp, event.GameID)

	// Update player statistics
	if err == nil && event.Winner != "" {
		s.updatePlayerAnalytics(ctx, event.Winner, true)

		loser := event.Player1
		if event.Winner == event.Player1 {
			loser = event.Player2
		}
		if loser != "" {
			s.updatePlayerAnalytics(ctx, loser, false)
		}

		s.updateHourlyStats(ctx, event.Timestamp, "game_finished")
		s.updateDailyStats(ctx, event.Timestamp, "game_finished")
	}

	return err
}

func (s *AnalyticsService) updatePlayerAnalytics(ctx context.Context, username string, won bool) error {
	query := `
		INSERT INTO analytics_players (username, games_played, games_won, games_lost)
		VALUES ($1, 1, $2, $3)
		ON CONFLICT (username) DO UPDATE SET
			games_played = analytics_players.games_played + 1,
			games_won = analytics_players.games_won + $2,
			games_lost = analytics_players.games_lost + $3,
			last_played = CURRENT_TIMESTAMP
	`

	gamesWon := 0
	gamesLost := 0
	if won {
		gamesWon = 1
	} else {
		gamesLost = 1
	}

	_, err := s.db.Exec(ctx, query, username, gamesWon, gamesLost)
	return err
}

func initAnalyticsTables(ctx context.Context, db *pgxpool.Pool) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS analytics_games (
			id SERIAL PRIMARY KEY,
			game_id VARCHAR(255) UNIQUE NOT NULL,
			player1 VARCHAR(255) NOT NULL,
			player2 VARCHAR(255) NOT NULL,
			winner VARCHAR(255),
			result VARCHAR(50),
			duration FLOAT,
			started_at TIMESTAMP,
			finished_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_moves (
			id SERIAL PRIMARY KEY,
			game_id VARCHAR(255) NOT NULL,
			player VARCHAR(255) NOT NULL,
			col_index INTEGER NOT NULL,
			row_index INTEGER NOT NULL,
			move_time TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_players (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			games_played INTEGER DEFAULT 0,
			games_won INTEGER DEFAULT 0,
			games_lost INTEGER DEFAULT 0,
			last_played TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_hourly (
			id SERIAL PRIMARY KEY,
			hour_timestamp TIMESTAMP NOT NULL,
			games_started INTEGER DEFAULT 0,
			games_completed INTEGER DEFAULT 0,
			total_moves INTEGER DEFAULT 0,
			unique_players INTEGER DEFAULT 0,
			avg_game_duration FLOAT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(hour_timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_daily (
			id SERIAL PRIMARY KEY,
			date DATE NOT NULL,
			games_started INTEGER DEFAULT 0,
			games_completed INTEGER DEFAULT 0,
			total_moves INTEGER DEFAULT 0,
			unique_players INTEGER DEFAULT 0,
			avg_game_duration FLOAT DEFAULT 0,
			peak_hour INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_games_game_id ON analytics_games(game_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_games_started_at ON analytics_games(started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_games_finished_at ON analytics_games(finished_at)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_moves_game_id ON analytics_moves(game_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_moves_time ON analytics_moves(move_time)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_players_username ON analytics_players(username)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_hourly_timestamp ON analytics_hourly(hour_timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_daily_date ON analytics_daily(date)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(ctx, query); err != nil {
			return err
		}
	}

	log.Println("Analytics tables initialized successfully")
	return nil
}

func (s *AnalyticsService) updateHourlyStats(ctx context.Context, timestamp int64, eventType string) error {
	query := `
		INSERT INTO analytics_hourly (hour_timestamp, games_started, games_completed, total_moves)
		VALUES (date_trunc('hour', to_timestamp($1)), 
			CASE WHEN $2 = 'game_started' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'game_finished' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'move_made' THEN 1 ELSE 0 END)
		ON CONFLICT (hour_timestamp) DO UPDATE SET
			games_started = analytics_hourly.games_started + CASE WHEN $2 = 'game_started' THEN 1 ELSE 0 END,
			games_completed = analytics_hourly.games_completed + CASE WHEN $2 = 'game_finished' THEN 1 ELSE 0 END,
			total_moves = analytics_hourly.total_moves + CASE WHEN $2 = 'move_made' THEN 1 ELSE 0 END
	`
	_, err := s.db.Exec(ctx, query, timestamp, eventType)
	return err
}

func (s *AnalyticsService) updateDailyStats(ctx context.Context, timestamp int64, eventType string) error {
	query := `
		INSERT INTO analytics_daily (date, games_started, games_completed, total_moves)
		VALUES (date_trunc('day', to_timestamp($1)), 
			CASE WHEN $2 = 'game_started' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'game_finished' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'move_made' THEN 1 ELSE 0 END)
		ON CONFLICT (date) DO UPDATE SET
			games_started = analytics_daily.games_started + CASE WHEN $2 = 'game_started' THEN 1 ELSE 0 END,
			games_completed = analytics_daily.games_completed + CASE WHEN $2 = 'game_finished' THEN 1 ELSE 0 END,
			total_moves = analytics_daily.total_moves + CASE WHEN $2 = 'move_made' THEN 1 ELSE 0 END
	`
	_, err := s.db.Exec(ctx, query, timestamp, eventType)
	return err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
