package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

type GameRecord struct {
	ID         string      `json:"id"`
	Player1    string      `json:"player1"`
	Player2    string      `json:"player2"`
	Winner     *string     `json:"winner,omitempty"`
	Result     string      `json:"result"`
	BoardState [][]int     `json:"board_state"`
	StartedAt  *time.Time  `json:"started_at,omitempty"`
	FinishedAt *time.Time  `json:"finished_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	IsBot     bool      `json:"is_bot"`
	GamesWon  int       `json:"games_won"`
	GamesLost int       `json:"games_lost"`
	GamesDrawn int      `json:"games_drawn"`
	CreatedAt time.Time `json:"created_at"`
}

func NewDB(databaseURL string) (*DB, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Migrate() error {
	ctx := context.Background()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			is_bot BOOLEAN DEFAULT FALSE,
			games_won INTEGER DEFAULT 0,
			games_lost INTEGER DEFAULT 0,
			games_drawn INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(255) PRIMARY KEY,
			player1 VARCHAR(255) NOT NULL,
			player2 VARCHAR(255),
			winner VARCHAR(255),
			result VARCHAR(50) NOT NULL,
			board_state JSONB NOT NULL,
			started_at TIMESTAMP,
			finished_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (player1) REFERENCES users(username) ON DELETE CASCADE,
			FOREIGN KEY (player2) REFERENCES users(username) ON DELETE CASCADE,
			FOREIGN KEY (winner) REFERENCES users(username) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_games_player1 ON games(player1)`,
		`CREATE INDEX IF NOT EXISTS idx_games_player2 ON games(player2)`,
		`CREATE INDEX IF NOT EXISTS idx_games_winner ON games(winner)`,
		`CREATE INDEX IF NOT EXISTS idx_games_created_at ON games(created_at)`,
	}

	for _, query := range queries {
		if _, err := db.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// UpsertUser creates or updates a user
func (db *DB) UpsertUser(ctx context.Context, username string, isBot bool) error {
	query := `
		INSERT INTO users (username, is_bot) 
		VALUES ($1, $2)
		ON CONFLICT (username) DO NOTHING
	`
	_, err := db.pool.Exec(ctx, query, username, isBot)
	return err
}

// SaveGame saves a completed game
func (db *DB) SaveGame(ctx context.Context, game *GameRecord) error {
	boardJSON, err := json.Marshal(game.BoardState)
	if err != nil {
		return fmt.Errorf("failed to marshal board state: %w", err)
	}

	query := `
		INSERT INTO games (id, player1, player2, winner, result, board_state, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			winner = EXCLUDED.winner,
			result = EXCLUDED.result,
			board_state = EXCLUDED.board_state,
			finished_at = EXCLUDED.finished_at
	`
	
	_, err = db.pool.Exec(ctx, query,
		game.ID,
		game.Player1,
		game.Player2,
		game.Winner,
		game.Result,
		boardJSON,
		game.StartedAt,
		game.FinishedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	// Update user statistics
	if game.Winner != nil && *game.Winner != "" {
		if err := db.updateUserStats(ctx, *game.Winner, true, false); err != nil {
			return err
		}
		
		// Update loser stats
		loser := game.Player1
		if *game.Winner == game.Player1 {
			loser = game.Player2
		}
		if loser != "" {
			if err := db.updateUserStats(ctx, loser, false, false); err != nil {
				return err
			}
		}
	} else if game.Result == "draw" {
		// Both players get a draw
		if err := db.updateUserStats(ctx, game.Player1, false, true); err != nil {
			return err
		}
		if game.Player2 != "" {
			if err := db.updateUserStats(ctx, game.Player2, false, true); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateUserStats updates win/loss/draw counts for a user
func (db *DB) updateUserStats(ctx context.Context, username string, won, drawn bool) error {
	var query string
	if won {
		query = `UPDATE users SET games_won = games_won + 1 WHERE username = $1`
	} else if drawn {
		query = `UPDATE users SET games_drawn = games_drawn + 1 WHERE username = $1`
	} else {
		query = `UPDATE users SET games_lost = games_lost + 1 WHERE username = $1`
	}
	
	_, err := db.pool.Exec(ctx, query, username)
	return err
}

// GetLeaderboard returns top players by wins
func (db *DB) GetLeaderboard(ctx context.Context, limit int) ([]User, error) {
	query := `
		SELECT id, username, is_bot, games_won, games_lost, games_drawn, created_at
		FROM users
		WHERE is_bot = FALSE
		ORDER BY games_won DESC, games_lost ASC
		LIMIT $1
	`
	
	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsBot,
			&user.GamesWon,
			&user.GamesLost,
			&user.GamesDrawn,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserStats returns statistics for a specific user
func (db *DB) GetUserStats(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, is_bot, games_won, games_lost, games_drawn, created_at
		FROM users
		WHERE username = $1
	`
	
	var user User
	err := db.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.IsBot,
		&user.GamesWon,
		&user.GamesLost,
		&user.GamesDrawn,
		&user.CreatedAt,
	)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	
	return &user, err
}

// GetRecentGames returns recent games
func (db *DB) GetRecentGames(ctx context.Context, limit int) ([]GameRecord, error) {
	query := `
		SELECT id, player1, player2, winner, result, board_state, started_at, finished_at, created_at
		FROM games
		ORDER BY created_at DESC
		LIMIT $1
	`
	
	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []GameRecord
	for rows.Next() {
		var game GameRecord
		var boardJSON []byte
		
		err := rows.Scan(
			&game.ID,
			&game.Player1,
			&game.Player2,
			&game.Winner,
			&game.Result,
			&boardJSON,
			&game.StartedAt,
			&game.FinishedAt,
			&game.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(boardJSON, &game.BoardState); err != nil {
			return nil, err
		}
		
		games = append(games, game)
	}

	return games, rows.Err()
}

// GetUserGames returns games for a specific user
func (db *DB) GetUserGames(ctx context.Context, username string, limit int) ([]GameRecord, error) {
	query := `
		SELECT id, player1, player2, winner, result, board_state, started_at, finished_at, created_at
		FROM games
		WHERE player1 = $1 OR player2 = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	rows, err := db.pool.Query(ctx, query, username, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []GameRecord
	for rows.Next() {
		var game GameRecord
		var boardJSON []byte
		
		err := rows.Scan(
			&game.ID,
			&game.Player1,
			&game.Player2,
			&game.Winner,
			&game.Result,
			&boardJSON,
			&game.StartedAt,
			&game.FinishedAt,
			&game.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(boardJSON, &game.BoardState); err != nil {
			return nil, err
		}
		
		games = append(games, game)
	}

	return games, rows.Err()
}
