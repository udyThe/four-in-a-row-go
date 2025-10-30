package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourusername/4-in-a-row/internal/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSClient struct {
	conn     *websocket.Conn
	playerID string
	gameID   string
	send     chan []byte
	server   *Server
	mu       sync.Mutex
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &WSClient{
		conn:   conn,
		send:   make(chan []byte, 256),
		server: s,
	}

	s.registerClient(client)

	go client.writePump()
	go client.readPump()
}

func (s *Server) registerClient(client *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = true
	// Log registration for debugging websocket issues
	if client.conn != nil {
		log.Printf("Registered client from %s", client.conn.RemoteAddr().String())
	} else {
		log.Printf("Registered client (no conn available yet): %v", client)
	}
}

func (s *Server) unregisterClient(client *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		close(client.send)

		// Mark player as disconnected
		if client.playerID != "" {
			s.gameManager.SetPlayerDisconnected(client.playerID)
		}
	}
}

func (client *WSClient) readPump() {
	defer func() {
		client.server.unregisterClient(client)
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		client.handleMessage(message)
	}
}

func (client *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *WSClient) handleMessage(message []byte) {
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		client.sendError("Invalid message format")
		return
	}

	switch wsMsg.Type {
	case "join":
		client.handleJoin(wsMsg.Payload)
	case "move":
		client.handleMove(wsMsg.Payload)
	case "reconnect":
		client.handleReconnect(wsMsg.Payload)
	case "heartbeat":
		client.handleHeartbeat()
	default:
		client.sendError("Unknown message type")
	}
}

func (client *WSClient) handleJoin(payload json.RawMessage) {
	var data struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		client.sendError("Invalid join payload")
		return
	}

	if data.Username == "" {
		client.sendError("Username is required")
		return
	}

	// Add player to matchmaking. matchmaker now returns a matched flag to
	// indicate whether a second player was found immediately. We defer
	// calling JoinGame until after we set the WS client fields so the
	// game update callback can find both clients.
	player, gameObj, matched := client.server.matchmaker.AddPlayer(data.Username)

	// Assign client identifiers immediately so the client is discoverable
	// by server-level broadcasts.
	client.playerID = player.ID
	client.gameID = gameObj.ID

	log.Printf("Client joined: player_id=%s game_id=%s remote=%s", client.playerID, client.gameID, client.conn.RemoteAddr().String())

	// If a match was found, explicitly join the game now (this will emit
	// events and trigger the onGameUpdate callback which will broadcast
	// to connected clients). Doing this after setting client.playerID and
	// client.gameID avoids the race where the callback fires before the
	// WS client is ready to receive it.
	if matched {
		if err := client.server.gameManager.JoinGame(gameObj.ID, player); err != nil {
			log.Printf("Error joining matched game: %v", err)
			client.sendError("Failed to join matched game")
			return
		}
	}

	// Send player info including session token for reconnect
	client.sendMessage("player_info", map[string]interface{}{
		"player_id":     player.ID,
		"game_id":       gameObj.ID,
		"username":      player.Username,
		"session_token": player.SessionToken,
	})

	// Always broadcast current game state to ensure all connected clients get the update
	// This is critical when player 2 joins - both clients need to transition from waiting to playing
	client.broadcastGameState(gameObj)

	// If game is waiting for player 2, also send waiting message to this client
	if gameObj.Player2 == nil {
		client.sendMessage("waiting", map[string]interface{}{
			"message": "Waiting for opponent...",
		})
	}

	// If playing against bot and bot's turn, make bot move
	if gameObj.Player2 != nil && gameObj.Player2.IsBot && gameObj.CurrentTurn == game.Player2 {
		go func() {
			time.Sleep(1 * time.Second)
			client.server.gameManager.HandleBotMove(gameObj.ID)
			if g, err := client.server.gameManager.GetGame(gameObj.ID); err == nil {
				client.broadcastGameState(g)
			}
		}()
	}
}

func (client *WSClient) handleMove(payload json.RawMessage) {
	var data struct {
		Column int `json:"column"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		client.sendError("Invalid move payload")
		return
	}

	if client.gameID == "" || client.playerID == "" {
		client.sendError("Not in a game")
		return
	}

	// Make the move
	_, err := client.server.gameManager.MakeMove(client.gameID, client.playerID, data.Column)
	if err != nil {
		client.sendError(err.Error())
		return
	}

	// Get updated game state
	gameObj, err := client.server.gameManager.GetGame(client.gameID)
	if err != nil {
		client.sendError("Game not found")
		return
	}

	// Broadcast updated state
	client.broadcastGameState(gameObj)

	// If bot's turn, make bot move
	if gameObj.Status == game.StatusInProgress &&
		gameObj.Player2 != nil &&
		gameObj.Player2.IsBot &&
		gameObj.CurrentTurn == game.Player2 {
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := client.server.gameManager.HandleBotMove(client.gameID); err != nil {
				return
			}
			if g, err := client.server.gameManager.GetGame(client.gameID); err == nil {
				client.broadcastGameState(g)
			}
		}()
	}
}

func (client *WSClient) handleReconnect(payload json.RawMessage) {
	var data struct {
		SessionToken string `json:"session_token"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		client.sendError("Invalid reconnect payload")
		return
	}

	if data.SessionToken == "" {
		client.sendError("Session token is required")
		return
	}

	// Try to reconnect using session token
	gameObj, player, err := client.server.gameManager.ReconnectPlayer(data.SessionToken)
	if err != nil {
		log.Printf("Reconnect failed for session %s: %v", data.SessionToken, err)
		client.sendError(fmt.Sprintf("Reconnect failed: %v", err))
		return
	}

	// Set client identifiers
	client.playerID = player.ID
	client.gameID = gameObj.ID

	log.Printf("Client reconnected: player=%s game=%s session=%s", player.Username, gameObj.ID, data.SessionToken)

	// Send reconnection success with full player info
	client.sendMessage("reconnected", map[string]interface{}{
		"game_id":       gameObj.ID,
		"player_id":     player.ID,
		"username":      player.Username,
		"session_token": player.SessionToken,
	})

	// Send current game state
	client.broadcastGameState(gameObj)
}

func (client *WSClient) handleHeartbeat() {
	if client.playerID != "" {
		client.server.gameManager.UpdatePlayerHeartbeat(client.playerID)
		// Also push current game state to help clients transition from 'waiting' to active
		if client.gameID != "" {
			if g, err := client.server.gameManager.GetGame(client.gameID); err == nil {
				// If a bot or second player joined since last update, this ensures the client receives it
				client.broadcastGameState(g)
			}
		}
	}
}

func (client *WSClient) sendMessage(msgType string, payload interface{}) {
	msg := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.send <- data:
	default:
		// Client buffer full, disconnect
		client.server.unregisterClient(client)
	}
}

func (client *WSClient) sendError(message string) {
	client.sendMessage("error", map[string]interface{}{
		"message": message,
	})
}

func (client *WSClient) broadcastGameState(gameObj *game.Game) {
	gameData, err := gameObj.ToJSON()
	if err != nil {
		return
	}

	var gameMap map[string]interface{}
	json.Unmarshal(gameData, &gameMap)

	// Broadcast to all clients in this game
	client.server.mu.RLock()
	defer client.server.mu.RUnlock()

	for c := range client.server.clients {
		if c.gameID == gameObj.ID {
			c.sendMessage("game_update", gameMap)
		}
	}
}
