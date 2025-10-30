package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/yourusername/4-in-a-row/internal/config"
	"github.com/yourusername/4-in-a-row/internal/database"
	"github.com/yourusername/4-in-a-row/internal/game"
)

type Server struct {
	config      *config.Config
	gameManager *game.Manager
	matchmaker  *game.Matchmaker
	db          *database.DB
	clients     map[*WSClient]bool
	mu          sync.RWMutex
}

func NewServer(cfg *config.Config, gameManager *game.Manager, matchmaker *game.Matchmaker, db *database.DB) *Server {
	s := &Server{
		config:      cfg,
		gameManager: gameManager,
		matchmaker:  matchmaker,
		db:          db,
		clients:     make(map[*WSClient]bool),
	}

	// Register callback to broadcast game updates when state changes (e.g., bot joins)
	gameManager.SetGameUpdateCallback(func(gameID string) {
		log.Printf("GameUpdateCallback invoked for game %s", gameID)
		s.broadcastGameUpdate(gameID)
	})

	return s
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", s.handleWebSocket)

	// REST API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/leaderboard", s.handleLeaderboard).Methods("GET")
	api.HandleFunc("/user/{username}", s.handleUserStats).Methods("GET")
	api.HandleFunc("/games/recent", s.handleRecentGames).Methods("GET")
	api.HandleFunc("/games/user/{username}", s.handleUserGames).Methods("GET")
	api.HandleFunc("/analytics/hourly", s.handleHourlyAnalytics).Methods("GET")
	api.HandleFunc("/analytics/daily", s.handleDailyAnalytics).Methods("GET")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	return c.Handler(r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": "ok",
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	users, err := s.db.GetLeaderboard(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch leaderboard")
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (s *Server) handleUserStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := s.db.GetUserStats(r.Context(), username)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleRecentGames(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	games, err := s.db.GetRecentGames(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch games")
		return
	}

	respondJSON(w, http.StatusOK, games)
}

func (s *Server) handleUserGames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	games, err := s.db.GetUserGames(r.Context(), username, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch user games")
		return
	}

	respondJSON(w, http.StatusOK, games)
}

func (s *Server) handleHourlyAnalytics(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	analytics, err := s.db.GetHourlyAnalytics(r.Context(), hours)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch hourly analytics")
		return
	}

	respondJSON(w, http.StatusOK, analytics)
}

func (s *Server) handleDailyAnalytics(w http.ResponseWriter, r *http.Request) {
	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	analytics, err := s.db.GetDailyAnalytics(r.Context(), days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch daily analytics")
		return
	}

	respondJSON(w, http.StatusOK, analytics)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// broadcastGameUpdate sends game state update to all connected clients in a game
func (s *Server) broadcastGameUpdate(gameID string) {
	log.Printf("Broadcasting game update for game: %s", gameID)

	gameObj, err := s.gameManager.GetGame(gameID)
	if err != nil {
		log.Printf("Error getting game %s: %v", gameID, err)
		return
	}

	gameData, err := gameObj.ToJSON()
	if err != nil {
		log.Printf("Error converting game to JSON: %v", err)
		return
	}

	var gameMap map[string]interface{}
	json.Unmarshal(gameData, &gameMap)

	// Broadcast to all clients in this game
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientCount := 0
	for c := range s.clients {
		if c.gameID == gameID {
			clientCount++
			c.sendMessage("game_update", gameMap)
		}
	}

	log.Printf("Broadcast game_update to %d clients for game %s", clientCount, gameID)
}
