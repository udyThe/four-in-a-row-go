package game

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"
)

// StartMetricsEmitter starts a background goroutine that emits periodic system metrics to Kafka
func (m *Manager) StartMetricsEmitter() {
	if m.kafkaProducer == nil {
		log.Println("Kafka producer not available, metrics emitter disabled")
		return
	}

	log.Println("Starting metrics emitter (sends stats every 60 seconds)")

	// Emit metrics every minute
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		// Send initial metrics immediately
		m.emitSystemMetrics()

		for range ticker.C {
			m.emitSystemMetrics()
		}
	}()
}

func (m *Manager) emitSystemMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate metrics
	activeGamesCount := 0
	inProgressGames := 0
	waitingGames := 0
	finishedGames := 0
	botGames := 0
	humanGames := 0

	for _, game := range m.games {
		activeGamesCount++
		switch game.Status {
		case StatusWaiting:
			waitingGames++
		case StatusInProgress:
			inProgressGames++
		case StatusFinished:
			finishedGames++
		}

		if game.Player2 != nil && game.Player2.Username == "Bot" {
			botGames++
		} else if game.Player2 != nil {
			humanGames++
		}
	}

	totalPlayers := len(m.playerGames)
	connectedPlayers := 0
	disconnectedPlayers := 0

	for playerID := range m.playerGames {
		game, err := m.GetGameByPlayer(playerID)
		if err == nil {
			if game.Player1 != nil && game.Player1.ID == playerID && game.Player1.Connected {
				connectedPlayers++
			} else if game.Player2 != nil && game.Player2.ID == playerID && game.Player2.Connected {
				connectedPlayers++
			} else {
				disconnectedPlayers++
			}
		}
	}

	// System metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	now := time.Now()

	event := map[string]interface{}{
		"event_type":    "system_metrics",
		"timestamp":     now.Unix(),
		"timestamp_iso": now.Format(time.RFC3339),
		"hour_of_day":   now.Hour(),
		"day_of_week":   now.Weekday().String(),
		"date":          now.Format("2006-01-02"),

		// Game metrics
		"total_active_games":   activeGamesCount,
		"games_in_progress":    inProgressGames,
		"games_waiting":        waitingGames,
		"games_finished_cache": finishedGames,
		"bot_games":            botGames,
		"human_vs_human_games": humanGames,

		// Player metrics
		"total_players":        totalPlayers,
		"connected_players":    connectedPlayers,
		"disconnected_players": disconnectedPlayers,

		// System metrics
		"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
		"memory_total_mb": memStats.TotalAlloc / 1024 / 1024,
		"num_goroutines":  runtime.NumGoroutine(),
		"num_gc_cycles":   memStats.NumGC,

		// Hourly rate calculations (estimated)
		"estimated_games_per_hour":    inProgressGames * 60,     // rough estimate
		"estimated_requests_per_hour": (connectedPlayers * 120), // ~2 requests/min per player
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal system metrics: %v", err)
		return
	}

	err = m.kafkaProducer.SendMessage(context.Background(), "game-events", data)
	if err != nil {
		log.Printf("Failed to send system metrics to Kafka: %v", err)
	} else {
		log.Printf("System metrics emitted: %d active games, %d connected players", activeGamesCount, connectedPlayers)
	}
}
