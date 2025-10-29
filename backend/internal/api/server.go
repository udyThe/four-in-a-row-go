package api

import (
	"encoding/json"
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
	return &Server{
		config:      cfg,
		gameManager: gameManager,
		matchmaker:  matchmaker,
		db:          db,
		clients:     make(map[*WSClient]bool),
	}
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
		"status": "healthy",
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

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
