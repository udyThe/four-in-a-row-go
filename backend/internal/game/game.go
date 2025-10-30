package game

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GameStatus string

const (
	StatusWaiting    GameStatus = "waiting"
	StatusInProgress GameStatus = "in_progress"
	StatusFinished   GameStatus = "finished"
	StatusAbandoned  GameStatus = "abandoned"
)

type GameResult string

const (
	ResultPlayer1Win GameResult = "player1_win"
	ResultPlayer2Win GameResult = "player2_win"
	ResultDraw       GameResult = "draw"
	ResultAbandoned  GameResult = "abandoned"
)

type Player struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	SessionToken   string     `json:"session_token"`
	IsBot          bool       `json:"is_bot"`
	Connected      bool       `json:"connected"`
	LastHeartbeat  time.Time  `json:"-"`
	DisconnectedAt *time.Time `json:"-"`
}

type Game struct {
	ID             string     `json:"id"`
	Player1        *Player    `json:"player1"`
	Player2        *Player    `json:"player2"`
	Board          *Board     `json:"board"`
	CurrentTurn    CellState  `json:"current_turn"`
	Status         GameStatus `json:"status"`
	Winner         *Player    `json:"winner,omitempty"`
	Result         GameResult `json:"result,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	LastMoveAt     time.Time  `json:"last_move_at"`
	TurnStartedAt  time.Time  `json:"turn_started_at"`
	TurnTimeoutSec int        `json:"turn_timeout_sec"`
	Bot            *Bot       `json:"-"`
	mu             sync.RWMutex
}

func NewGame(player1 *Player) *Game {
	now := time.Now()
	return &Game{
		ID:             uuid.New().String(),
		Player1:        player1,
		Board:          NewBoard(),
		CurrentTurn:    Player1,
		Status:         StatusWaiting,
		CreatedAt:      now,
		LastMoveAt:     now,
		TurnStartedAt:  now,
		TurnTimeoutSec: 30, // 30 seconds per turn
	}
}

// AddPlayer2 adds the second player to the game
func (g *Game) AddPlayer2(player2 *Player) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Player2 = player2
	now := time.Now()
	g.StartedAt = &now
	g.Status = StatusInProgress
	g.TurnStartedAt = now // Start timer for first turn

	// Initialize bot if player2 is a bot
	if player2.IsBot {
		g.Bot = NewBot(Player2)
	}
}

// MakeMove processes a move in the game
func (g *Game) MakeMove(playerID string, column int) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate game state
	if g.Status != StatusInProgress {
		return -1, ErrGameNotInProgress
	}

	// Validate player turn
	var currentPlayer CellState
	if g.Player1.ID == playerID {
		currentPlayer = Player1
	} else if g.Player2.ID == playerID {
		currentPlayer = Player2
	} else {
		return -1, ErrInvalidPlayer
	}

	if currentPlayer != g.CurrentTurn {
		return -1, ErrNotYourTurn
	}

	// Make the move
	row, err := g.Board.DropDisc(column, currentPlayer)
	if err != nil {
		return -1, err
	}

	now := time.Now()
	g.LastMoveAt = now

	// Check for win
	if g.Board.CheckWin(currentPlayer) {
		g.finishGame(currentPlayer)
		return row, nil
	}

	// Check for draw
	if g.Board.IsFull() {
		g.finishGameDraw()
		return row, nil
	}

	// Switch turn and reset turn timer
	if g.CurrentTurn == Player1 {
		g.CurrentTurn = Player2
	} else {
		g.CurrentTurn = Player1
	}
	g.TurnStartedAt = now // Reset turn timer for next player

	return row, nil
}

// GetBotMove gets the next move from the bot
func (g *Game) GetBotMove() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Bot == nil {
		return -1
	}

	return g.Bot.GetBestMove(g.Board)
}

// SkipTurn skips the current player's turn due to timeout
func (g *Game) SkipTurn() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusInProgress {
		return
	}

	now := time.Now()
	g.LastMoveAt = now

	// Switch turn without making a move
	if g.CurrentTurn == Player1 {
		g.CurrentTurn = Player2
	} else {
		g.CurrentTurn = Player1
	}
	g.TurnStartedAt = now // Reset turn timer for next player

	log.Printf("Turn skipped for game %s, now %v's turn", g.ID, g.CurrentTurn)
}

// finishGame marks the game as finished with a winner
func (g *Game) finishGame(winner CellState) {
	now := time.Now()
	g.FinishedAt = &now
	g.Status = StatusFinished

	if winner == Player1 {
		g.Winner = g.Player1
		g.Result = ResultPlayer1Win
	} else {
		g.Winner = g.Player2
		g.Result = ResultPlayer2Win
	}
}

// finishGameDraw marks the game as finished in a draw
func (g *Game) finishGameDraw() {
	now := time.Now()
	g.FinishedAt = &now
	g.Status = StatusFinished
	g.Result = ResultDraw
}

// AbandonGame marks the game as abandoned
func (g *Game) AbandonGame(disconnectedPlayerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status == StatusFinished {
		return
	}

	now := time.Now()
	g.FinishedAt = &now
	g.Status = StatusAbandoned
	g.Result = ResultAbandoned

	// Set winner as the other player
	if g.Player1.ID == disconnectedPlayerID && g.Player2 != nil {
		g.Winner = g.Player2
		if !g.Player2.IsBot {
			g.Result = ResultPlayer2Win
		}
	} else if g.Player2 != nil && g.Player2.ID == disconnectedPlayerID {
		g.Winner = g.Player1
		g.Result = ResultPlayer1Win
	}
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.CurrentTurn == Player1 {
		return g.Player1
	}
	return g.Player2
}

// IsPlayerTurn checks if it's the specified player's turn
func (g *Game) IsPlayerTurn(playerID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	currentPlayer := g.GetCurrentPlayer()
	return currentPlayer != nil && currentPlayer.ID == playerID
}

// UpdateHeartbeat updates the last heartbeat time for a player
func (g *Game) UpdateHeartbeat(playerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Player1.ID == playerID {
		g.Player1.LastHeartbeat = time.Now()
		g.Player1.Connected = true
	} else if g.Player2 != nil && g.Player2.ID == playerID {
		g.Player2.LastHeartbeat = time.Now()
		g.Player2.Connected = true
	}
}

// SetPlayerDisconnected marks a player as disconnected
func (g *Game) SetPlayerDisconnected(playerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	if g.Player1.ID == playerID {
		g.Player1.Connected = false
		g.Player1.DisconnectedAt = &now
	} else if g.Player2 != nil && g.Player2.ID == playerID {
		g.Player2.Connected = false
		g.Player2.DisconnectedAt = &now
	}
}

// ToJSON converts the game to JSON
func (g *Game) ToJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	type GameJSON struct {
		ID             string     `json:"id"`
		Player1        *Player    `json:"player1"`
		Player2        *Player    `json:"player2"`
		Board          [][]int    `json:"board"`
		CurrentTurn    int        `json:"current_turn"`
		Status         GameStatus `json:"status"`
		Winner         *Player    `json:"winner,omitempty"`
		Result         GameResult `json:"result,omitempty"`
		CreatedAt      time.Time  `json:"created_at"`
		StartedAt      *time.Time `json:"started_at,omitempty"`
		FinishedAt     *time.Time `json:"finished_at,omitempty"`
		LastMoveAt     time.Time  `json:"last_move_at"`
		TurnStartedAt  time.Time  `json:"turn_started_at"`
		TurnTimeoutSec int        `json:"turn_timeout_sec"`
	}

	gameJSON := GameJSON{
		ID:             g.ID,
		Player1:        g.Player1,
		Player2:        g.Player2,
		Board:          g.Board.ToArray(),
		CurrentTurn:    int(g.CurrentTurn),
		Status:         g.Status,
		Winner:         g.Winner,
		Result:         g.Result,
		CreatedAt:      g.CreatedAt,
		StartedAt:      g.StartedAt,
		FinishedAt:     g.FinishedAt,
		LastMoveAt:     g.LastMoveAt,
		TurnStartedAt:  g.TurnStartedAt,
		TurnTimeoutSec: g.TurnTimeoutSec,
	}

	return json.Marshal(gameJSON)
}
