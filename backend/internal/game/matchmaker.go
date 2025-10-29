package game

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MatchmakingTimeout = 10 * time.Second
)

type MatchRequest struct {
	Player    *Player
	CreatedAt time.Time
}

type Matchmaker struct {
	queue       []*MatchRequest
	mu          sync.Mutex
	gameManager *Manager
}

func NewMatchmaker(gameManager *Manager) *Matchmaker {
	return &Matchmaker{
		queue:       make([]*MatchRequest, 0),
		gameManager: gameManager,
	}
}

// Run starts the matchmaker loop
func (mm *Matchmaker) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mm.processQueue()
	}
}

// AddPlayer adds a player to the matchmaking queue
func (mm *Matchmaker) AddPlayer(username string) (*Player, *Game) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	player := &Player{
		ID:            uuid.New().String(),
		Username:      username,
		IsBot:         false,
		Connected:     true,
		LastHeartbeat: time.Now(),
	}

	// Check if there's already someone waiting
	if len(mm.queue) > 0 {
		waitingRequest := mm.queue[0]
		mm.queue = mm.queue[1:]

		// Retrieve the existing game that was created when the first player joined.
		// The initial AddPlayer call creates a game for the waiting player, so reuse it
		// to avoid creating duplicate games and mismatched game IDs for clients.
		game, err := mm.gameManager.GetGameByPlayer(waitingRequest.Player.ID)
		if err != nil {
			// Fallback: if for some reason the game isn't found, create a new one.
			game = mm.gameManager.CreateGame(waitingRequest.Player)
		}

		game.AddPlayer2(player)
		mm.gameManager.playerGames[player.ID] = game.ID

		log.Printf("Matched players: %s vs %s", waitingRequest.Player.Username, player.Username)

		// Emit game started event
		mm.gameManager.emitGameStartedEvent(game)

		return player, game
	}

	// No one waiting, add to queue
	request := &MatchRequest{
		Player:    player,
		CreatedAt: time.Now(),
	}
	mm.queue = append(mm.queue, request)

	// Create game immediately for this player
	game := mm.gameManager.CreateGame(player)

	log.Printf("Player %s added to matchmaking queue", username)

	return player, game
}

// processQueue checks for timeout and matches with bot
func (mm *Matchmaker) processQueue() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	remainingQueue := make([]*MatchRequest, 0)

	for _, request := range mm.queue {
		if now.Sub(request.CreatedAt) >= MatchmakingTimeout {
			// Timeout - match with bot
			mm.matchWithBot(request.Player)
			log.Printf("Player %s matched with bot after timeout", request.Player.Username)
		} else {
			remainingQueue = append(remainingQueue, request)
		}
	}

	mm.queue = remainingQueue
}

// matchWithBot creates a bot opponent for a player
func (mm *Matchmaker) matchWithBot(player *Player) {
	bot := &Player{
		ID:            uuid.New().String(),
		Username:      "Bot",
		IsBot:         true,
		Connected:     true,
		LastHeartbeat: time.Now(),
	}

	// Find the player's game
	game, err := mm.gameManager.GetGameByPlayer(player.ID)
	if err != nil {
		log.Printf("Error finding game for player %s: %v", player.Username, err)
		return
	}

	// Add bot as player 2
	mm.gameManager.JoinGame(game.ID, bot)

	log.Printf("Bot joined game %s with player %s", game.ID, player.Username)
}

// RemovePlayer removes a player from the queue (if disconnected before match)
func (mm *Matchmaker) RemovePlayer(playerID string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for i, request := range mm.queue {
		if request.Player.ID == playerID {
			mm.queue = append(mm.queue[:i], mm.queue[i+1:]...)
			log.Printf("Player %s removed from matchmaking queue", request.Player.Username)
			return
		}
	}
}
