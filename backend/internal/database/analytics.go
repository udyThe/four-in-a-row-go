package database

import (
	"context"
	"time"
)

type HourlyAnalytics struct {
	Hour            time.Time `json:"hour"`
	GamesStarted    int       `json:"games_started"`
	GamesCompleted  int       `json:"games_completed"`
	TotalMoves      int       `json:"total_moves"`
	AvgGameDuration float64   `json:"avg_game_duration"`
}

type DailyAnalytics struct {
	Date            time.Time `json:"date"`
	GamesStarted    int       `json:"games_started"`
	GamesCompleted  int       `json:"games_completed"`
	TotalMoves      int       `json:"total_moves"`
	AvgGameDuration float64   `json:"avg_game_duration"`
	PeakHour        *int      `json:"peak_hour,omitempty"`
}

func (db *DB) GetHourlyAnalytics(ctx context.Context, hours int) ([]HourlyAnalytics, error) {
	query := `
		SELECT 
			hour_timestamp,
			COALESCE(games_started, 0),
			COALESCE(games_completed, 0),
			COALESCE(total_moves, 0),
			COALESCE(avg_game_duration, 0)
		FROM analytics_hourly
		WHERE hour_timestamp >= NOW() - INTERVAL '1 hour' * $1
		ORDER BY hour_timestamp DESC
	`

	rows, err := db.pool.Query(ctx, query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []HourlyAnalytics
	for rows.Next() {
		var a HourlyAnalytics
		err := rows.Scan(
			&a.Hour,
			&a.GamesStarted,
			&a.GamesCompleted,
			&a.TotalMoves,
			&a.AvgGameDuration,
		)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, a)
	}

	return analytics, rows.Err()
}

func (db *DB) GetDailyAnalytics(ctx context.Context, days int) ([]DailyAnalytics, error) {
	query := `
		SELECT 
			date,
			COALESCE(games_started, 0),
			COALESCE(games_completed, 0),
			COALESCE(total_moves, 0),
			COALESCE(avg_game_duration, 0),
			peak_hour
		FROM analytics_daily
		WHERE date >= CURRENT_DATE - INTERVAL '1 day' * $1
		ORDER BY date DESC
	`

	rows, err := db.pool.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []DailyAnalytics
	for rows.Next() {
		var a DailyAnalytics
		err := rows.Scan(
			&a.Date,
			&a.GamesStarted,
			&a.GamesCompleted,
			&a.TotalMoves,
			&a.AvgGameDuration,
			&a.PeakHour,
		)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, a)
	}

	return analytics, rows.Err()
}
