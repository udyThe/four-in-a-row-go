package game

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/yourusername/4-in-a-row/internal/database"
	"github.com/yourusername/4-in-a-row/internal/kafka"
)

type Manager struct {
	games         map[string]*Game
	playerGames   map[string]string // playerID -> gameID
	sessionGames  map[string]string // sessionToken -> gameID
	mu            sync.RWMutex
	db            *database.DB
	kafkaProducer *kafka.Producer
	onGameUpdate  func(gameID string) // Callback when game state changes
}

func NewManager(db *database.DB, kafkaProducer *kafka.Producer) *Manager {
	m := &Manager{
		games:         make(map[string]*Game),
		playerGames:   make(map[string]string),
		sessionGames:  make(map[string]string),
		db:            db,
		kafkaProducer: kafkaProducer,
	}

	// Start cleanup goroutine
	go m.cleanupDisconnectedGames()

	// Start turn timer monitoring
	go m.monitorTurnTimers()

	return m
}

// SetGameUpdateCallback sets a callback function to be called when game state changes
func (m *Manager) SetGameUpdateCallback(callback func(gameID string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onGameUpdate = callback
	log.Printf("SetGameUpdateCallback: callback registered successfully (callback is nil: %v)", callback == nil)
}

// CreateGame creates a new game with player1
func (m *Manager) CreateGame(player1 *Player) *Game {
	m.mu.Lock()
	defer m.mu.Unlock()

	game := NewGame(player1)
	m.games[game.ID] = game
	m.playerGames[player1.ID] = game.ID
	if player1.SessionToken != "" {
		m.sessionGames[player1.SessionToken] = game.ID
	}

	log.Printf("Game created: %s for player %s (session: %s)", game.ID, player1.Username, player1.SessionToken)

	return game
}

// JoinGame adds player2 to an existing game
func (m *Manager) JoinGame(gameID string, player2 *Player) error {
	// Acquire lock, capture state, then release before callbacks
	m.mu.Lock()

	log.Printf("JoinGame called: gameID=%s player2=%s callback_is_nil=%v", gameID, player2.Username, m.onGameUpdate == nil)

	game, exists := m.games[gameID]
	if !exists {
		m.mu.Unlock()
		return ErrGameNotFound
	}

	game.AddPlayer2(player2)
	m.playerGames[player2.ID] = gameID
	if player2.SessionToken != "" {
		m.sessionGames[player2.SessionToken] = gameID
	}

	log.Printf("Player %s joined game %s (session: %s)", player2.Username, gameID, player2.SessionToken)

	// Capture state needed for Kafka event before releasing lock
	activeGames := len(m.games)
	totalPlayers := len(m.playerGames)

	// Release lock before emitting events or calling callbacks to avoid deadlock
	m.mu.Unlock()

	// Emit game started event (uses captured state, no locking)
	m.emitGameStartedEventWithState(game, activeGames, totalPlayers)

	// Notify websocket layer that game state changed
	if m.onGameUpdate != nil {
		log.Printf("Triggering game update callback for game %s", gameID)
		go m.onGameUpdate(gameID)
	} else {
		log.Printf("WARNING: onGameUpdate callback is nil for game %s", gameID)
	}

	return nil
}

// ReconnectPlayer reconnects a player to their game using session token
func (m *Manager) ReconnectPlayer(sessionToken string) (*Game, *Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find game by session token
	gameID, exists := m.sessionGames[sessionToken]
	if !exists {
		return nil, nil, errors.New("session not found or expired")
	}

	game, exists := m.games[gameID]
	if !exists {
		// Clean up stale session
		delete(m.sessionGames, sessionToken)
		return nil, nil, errors.New("game no longer exists")
	}

	// Find player in game
	var player *Player
	if game.Player1.SessionToken == sessionToken {
		player = game.Player1
	} else if game.Player2 != nil && game.Player2.SessionToken == sessionToken {
		player = game.Player2
	}

	if player == nil {
		return nil, nil, errors.New("player not found in game")
	}

	// Check if reconnect window is still open (30 seconds)
	if player.DisconnectedAt != nil {
		elapsed := time.Since(*player.DisconnectedAt)
		if elapsed > 30*time.Second {
			log.Printf("Reconnect window expired for player %s (elapsed: %v)", player.Username, elapsed)
			return nil, nil, errors.New("reconnect window expired (>30 seconds)")
		}
	}

	// Reconnect player
	player.Connected = true
	player.LastHeartbeat = time.Now()
	player.DisconnectedAt = nil

	log.Printf("Player %s reconnected to game %s (was disconnected for %v)",
		player.Username, gameID,
		func() time.Duration {
			if player.DisconnectedAt != nil {
				return time.Since(*player.DisconnectedAt)
			}
			return 0
		}())

	return game, player, nil
}

// GetGame retrieves a game by ID
func (m *Manager) GetGame(gameID string) (*Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, exists := m.games[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}

	return game, nil
}

// GetGameByPlayer retrieves a game by player ID
func (m *Manager) GetGameByPlayer(playerID string) (*Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameID, exists := m.playerGames[playerID]
	if !exists {
		return nil, ErrGameNotFound
	}

	game, exists := m.games[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}

	return game, nil
}

// MakeMove processes a move in a game
func (m *Manager) MakeMove(gameID, playerID string, column int) (int, error) {
	game, err := m.GetGame(gameID)
	if err != nil {
		return -1, err
	}

	row, err := game.MakeMove(playerID, column)
	if err != nil {
		return -1, err
	}

	// Emit move event
	m.emitMoveEvent(game, playerID, column, row)

	// Check if game is finished
	if game.Status == StatusFinished {
		m.handleGameFinished(game)
	}

	return row, nil
}

// HandleBotMove processes a bot's move
func (m *Manager) HandleBotMove(gameID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	if game.Player2 == nil || !game.Player2.IsBot {
		return errors.New("game does not have a bot")
	}

	if game.CurrentTurn != Player2 {
		return errors.New("not bot's turn")
	}

	// Add small delay to make it more natural
	time.Sleep(500 * time.Millisecond)

	column := game.GetBotMove()
	if column == -1 {
		return errors.New("bot could not find valid move")
	}

	_, err = m.MakeMove(gameID, game.Player2.ID, column)
	return err
}

// UpdatePlayerHeartbeat updates player's last heartbeat
func (m *Manager) UpdatePlayerHeartbeat(playerID string) error {
	game, err := m.GetGameByPlayer(playerID)
	if err != nil {
		return err
	}

	game.UpdateHeartbeat(playerID)
	return nil
}

// SetPlayerDisconnected marks a player as disconnected
func (m *Manager) SetPlayerDisconnected(playerID string) {
	game, err := m.GetGameByPlayer(playerID)
	if err != nil {
		return
	}

	game.SetPlayerDisconnected(playerID)
	log.Printf("Player %s disconnected from game %s", playerID, game.ID)
}

// ReconnectPlayerByID reconnects a player to their game by playerID (legacy method)
func (m *Manager) ReconnectPlayerByID(playerID string) (*Game, error) {
	game, err := m.GetGameByPlayer(playerID)
	if err != nil {
		return nil, err
	}

	game.UpdateHeartbeat(playerID)
	log.Printf("Player %s reconnected to game %s by playerID", playerID, game.ID)

	return game, nil
}

// cleanupDisconnectedGames checks for abandoned games
func (m *Manager) cleanupDisconnectedGames() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		for gameID, game := range m.games {
			if game.Status != StatusInProgress {
				continue
			}

			// Check player1
			if !game.Player1.IsBot && game.Player1.Connected {
				if now.Sub(game.Player1.LastHeartbeat) > 30*time.Second {
					log.Printf("Player %s timed out in game %s", game.Player1.Username, gameID)
					game.AbandonGame(game.Player1.ID)
					m.handleGameFinished(game)
				}
			}

			// Check player2
			if game.Player2 != nil && !game.Player2.IsBot && game.Player2.Connected {
				if now.Sub(game.Player2.LastHeartbeat) > 30*time.Second {
					log.Printf("Player %s timed out in game %s", game.Player2.Username, gameID)
					game.AbandonGame(game.Player2.ID)
					m.handleGameFinished(game)
				}
			}
		}

		m.mu.Unlock()
	}
}

// monitorTurnTimers checks for turn timeouts (30s per turn, 60s inactivity forfeits)
func (m *Manager) monitorTurnTimers() {
	ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds for responsiveness
	defer ticker.Stop()

	for range ticker.C {
		m.mu.RLock()
		now := time.Now()
		gamesToUpdate := make(map[string]*Game)

		for gameID, game := range m.games {
			if game.Status != StatusInProgress {
				continue
			}

			// Check for turn timeout (30 seconds)
			turnElapsed := now.Sub(game.TurnStartedAt)
			if turnElapsed > time.Duration(game.TurnTimeoutSec)*time.Second {
				// Skip turn if no move in 30 seconds
				gamesToUpdate[gameID] = game
				log.Printf("Turn timeout in game %s, skipping turn (elapsed: %v)", gameID, turnElapsed)
			}

			// Check for total inactivity timeout (60 seconds = forfeit)
			inactivityElapsed := now.Sub(game.LastMoveAt)
			if inactivityElapsed > 60*time.Second {
				// Forfeit game due to inactivity
				var inactivePlayerID string
				if game.CurrentTurn == Player1 {
					inactivePlayerID = game.Player1.ID
				} else if game.Player2 != nil {
					inactivePlayerID = game.Player2.ID
				}

				if inactivePlayerID != "" {
					log.Printf("Inactivity timeout in game %s, forfeiting game (elapsed: %v)", gameID, inactivityElapsed)
					// Will handle forfeit after releasing read lock
					game.AbandonGame(inactivePlayerID)
					m.handleGameFinished(game)
				}
			}
		}

		m.mu.RUnlock()

		// Skip turns outside of read lock to allow game updates
		for gameID, game := range gamesToUpdate {
			game.SkipTurn()
			log.Printf("Turn skipped for game %s", gameID)

			// Trigger game update callback to notify clients
			if m.onGameUpdate != nil {
				go m.onGameUpdate(gameID)
			}
		}
	}
}

// handleGameFinished processes a finished game
func (m *Manager) handleGameFinished(game *Game) {
	log.Printf("Game %s finished: %s", game.ID, game.Result)

	// Save to database
	if err := m.saveGameToDB(game); err != nil {
		log.Printf("Error saving game to database: %v", err)
	}

	// Emit game finished event
	m.emitGameFinishedEvent(game)

	// Clean up after a delay
	go func() {
		time.Sleep(1 * time.Minute)
		m.removeGame(game.ID)
	}()
}

// removeGame removes a game from memory
func (m *Manager) removeGame(gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[gameID]
	if !exists {
		return
	}

	// Clean up player mappings
	delete(m.playerGames, game.Player1.ID)
	if game.Player1.SessionToken != "" {
		delete(m.sessionGames, game.Player1.SessionToken)
	}

	if game.Player2 != nil {
		delete(m.playerGames, game.Player2.ID)
		if game.Player2.SessionToken != "" {
			delete(m.sessionGames, game.Player2.SessionToken)
		}
	}

	delete(m.games, gameID)

	log.Printf("Game %s removed from memory (cleaned up session tokens)", gameID)
}

// saveGameToDB saves a completed game to the database
func (m *Manager) saveGameToDB(game *Game) error {
	ctx := context.Background()

	// Ensure players exist in database
	if err := m.db.UpsertUser(ctx, game.Player1.Username, game.Player1.IsBot); err != nil {
		return err
	}
	if game.Player2 != nil {
		if err := m.db.UpsertUser(ctx, game.Player2.Username, game.Player2.IsBot); err != nil {
			return err
		}
	}

	// Save game
	var winner *string
	if game.Winner != nil {
		winner = &game.Winner.Username
	}

	player2Username := ""
	if game.Player2 != nil {
		player2Username = game.Player2.Username
	}

	return m.db.SaveGame(ctx, &database.GameRecord{
		ID:         game.ID,
		Player1:    game.Player1.Username,
		Player2:    player2Username,
		Winner:     winner,
		Result:     string(game.Result),
		BoardState: game.Board.ToArray(),
		StartedAt:  game.StartedAt,
		FinishedAt: game.FinishedAt,
	})
}

// Kafka event emission methods
func (m *Manager) emitGameStartedEvent(game *Game) {
	if m.kafkaProducer == nil {
		return
	}

	m.mu.RLock()
	activeGames := len(m.games)
	totalPlayers := len(m.playerGames)
	m.mu.RUnlock()

	m.emitGameStartedEventWithState(game, activeGames, totalPlayers)
}

func (m *Manager) emitGameStartedEventWithState(game *Game, activeGames, totalPlayers int) {
	if m.kafkaProducer == nil {
		return
	}

	event := map[string]interface{}{
		"event_type":    "game_started",
		"game_id":       game.ID,
		"player1":       game.Player1.Username,
		"player2":       game.Player2.Username,
		"timestamp":     time.Now().Unix(),
		"timestamp_iso": time.Now().Format(time.RFC3339),
		"hour_of_day":   time.Now().Hour(),
		"day_of_week":   time.Now().Weekday().String(),
		// Kafka metrics visible in UI
		"active_games":  activeGames,
		"total_players": totalPlayers,
		"is_bot_game":   game.Player2.Username == "Bot",
	}

	data, _ := json.Marshal(event)
	m.kafkaProducer.SendMessage(context.Background(), "game-events", data)
}

func (m *Manager) emitMoveEvent(game *Game, playerID string, column, row int) {
	if m.kafkaProducer == nil {
		return
	}

	username := ""
	isBot := false
	if game.Player1.ID == playerID {
		username = game.Player1.Username
	} else if game.Player2 != nil {
		username = game.Player2.Username
		isBot = game.Player2.Username == "Bot"
	}

	// Count total moves in game
	totalMoves := 0
	for _, col := range game.Board.Grid {
		for _, cell := range col {
			if cell != 0 {
				totalMoves++
			}
		}
	}

	event := map[string]interface{}{
		"event_type":    "move_made",
		"game_id":       game.ID,
		"player":        username,
		"column":        column,
		"row":           row,
		"timestamp":     time.Now().Unix(),
		"timestamp_iso": time.Now().Format(time.RFC3339),
		"hour_of_day":   time.Now().Hour(),
		// Kafka metrics
		"move_number": totalMoves,
		"is_bot_move": isBot,
	}

	data, _ := json.Marshal(event)
	m.kafkaProducer.SendMessage(context.Background(), "game-events", data)
}

func (m *Manager) emitGameFinishedEvent(game *Game) {
	if m.kafkaProducer == nil {
		return
	}

	duration := 0.0
	if game.StartedAt != nil && game.FinishedAt != nil {
		duration = game.FinishedAt.Sub(*game.StartedAt).Seconds()
	}

	winner := ""
	if game.Winner != nil {
		winner = game.Winner.Username
	}

	// Count total moves in game
	totalMoves := 0
	for _, col := range game.Board.Grid {
		for _, cell := range col {
			if cell != 0 {
				totalMoves++
			}
		}
	}

	m.mu.RLock()
	activeGames := len(m.games)
	totalPlayers := len(m.playerGames)
	m.mu.RUnlock()

	event := map[string]interface{}{
		"event_type":    "game_finished",
		"game_id":       game.ID,
		"player1":       game.Player1.Username,
		"player2":       game.Player2.Username,
		"winner":        winner,
		"result":        string(game.Result),
		"duration":      duration,
		"timestamp":     time.Now().Unix(),
		"timestamp_iso": time.Now().Format(time.RFC3339),
		"hour_of_day":   time.Now().Hour(),
		"day_of_week":   time.Now().Weekday().String(),
		// Kafka metrics visible in UI
		"total_moves":       totalMoves,
		"active_games":      activeGames,
		"total_players":     totalPlayers,
		"game_duration_sec": int(duration),
		"was_bot_game":      game.Player2.Username == "Bot",
	}

	data, _ := json.Marshal(event)
	m.kafkaProducer.SendMessage(context.Background(), "game-events", data)
}
