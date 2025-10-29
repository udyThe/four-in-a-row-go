# System Architecture

## Overview

This document describes the architecture of the 4 in a Row real-time multiplayer game.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER LAYER                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │   Browser    │  │   Browser    │  │   Browser    │           │
│  │  (Player 1)  │  │  (Player 2)  │  │  (Player N)  │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│         │                  │                  │                 │
│         └──────────────────┴──────────────────┘                 │
│                            │                                    │
└────────────────────────────┼────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    APPLICATION LAYER                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              FRONTEND (React + Nginx)                    │   │
│  │  • React Components  • WebSocket Client  • REST Client   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            │                                    │
│                            │ HTTP/WebSocket                     │
│                            ▼                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              BACKEND (Go Server)                         │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐          │   │
│  │  │ WebSocket  │  │ REST API   │  │   Game     │          │   │
│  │  │  Handler   │  │  Handler   │  │  Manager   │          │   │
│  │  └────────────┘  └────────────┘  └────────────┘          │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐          │   │
│  │  │ Matchmaker │  │  Bot AI    │  │   Kafka    │          │   │
│  │  │            │  │  (Minimax) │  │  Producer  │          │   │
│  │  └────────────┘  └────────────┘  └────────────┘          │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DATA LAYER                                 │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  PostgreSQL   │  │    Kafka     │  │  Analytics   │          │
│  │   Database    │  │   Message    │  │   Consumer   │          │
│  │               │  │    Queue     │  │   Service    │          │
│  └───────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Frontend (React)

**Technology**: React 18, WebSocket API, Axios

**Components**:
- `Game.js` - Main game orchestrator
- `GameBoard.js` - 7×6 grid display with animations
- `Leaderboard.js` - Rankings display
- `WebSocketService` - Real-time communication
- `API Service` - REST endpoint calls

**Responsibilities**:
- User interface rendering
- WebSocket connection management
- Game state visualization
- User input handling
- API communication

### 2. Backend (Go)

**Technology**: Go 1.21, Gorilla WebSocket, Gorilla Mux

#### 2.1 API Server (`internal/api`)
- HTTP routing (REST endpoints)
- WebSocket upgrade and management
- CORS handling
- Request validation

**Endpoints**:
- `GET /api/health` - Health check
- `GET /api/leaderboard` - Top players
- `GET /api/user/:username` - User statistics
- `GET /api/games/recent` - Recent games
- `GET /api/games/user/:username` - User's games
- `WS /ws` - WebSocket connection

#### 2.2 Game Engine (`internal/game`)

**Board.go**:
- 7×6 grid management
- Move validation
- Win detection (4 directions)
- Board state serialization

**Bot.go**:
- Minimax algorithm (depth 6)
- Alpha-beta pruning optimization
- Heuristic evaluation
- Position scoring

**Game.go**:
- Game state management
- Turn management
- Player synchronization
- Win/draw detection

**Manager.go**:
- Active game tracking
- Game lifecycle management
- Player-game mapping
- Disconnection handling

**Matchmaker.go**:
- Player queue management
- 10-second timeout enforcement
- Bot assignment
- Match creation

#### 2.3 Database Layer (`internal/database`)
- PostgreSQL connection pool
- Schema migrations
- CRUD operations
- Leaderboard queries
- Game history

#### 2.4 Kafka Integration (`internal/kafka`)
- Event production
- Topic management
- Error handling

### 3. Analytics Service

**Technology**: Go, Kafka Consumer, PostgreSQL

**Responsibilities**:
- Consume game events from Kafka
- Store analytics data
- Calculate metrics
- Track player statistics

**Events Consumed**:
- `game_started` - Game initiation
- `move_made` - Player moves
- `game_finished` - Game completion

**Analytics Tables**:
- `analytics_games` - Game metrics
- `analytics_moves` - Move history
- `analytics_players` - Player statistics

### 4. Database (PostgreSQL)

**Schema**:

```sql
users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    is_bot BOOLEAN,
    games_won INTEGER,
    games_lost INTEGER,
    games_drawn INTEGER,
    created_at TIMESTAMP
)

games (
    id VARCHAR(255) PRIMARY KEY,
    player1 VARCHAR(255) REFERENCES users(username),
    player2 VARCHAR(255) REFERENCES users(username),
    winner VARCHAR(255) REFERENCES users(username),
    result VARCHAR(50),
    board_state JSONB,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP
)

analytics_games (
    id SERIAL PRIMARY KEY,
    game_id VARCHAR(255) UNIQUE,
    player1 VARCHAR(255),
    player2 VARCHAR(255),
    winner VARCHAR(255),
    result VARCHAR(50),
    duration FLOAT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP
)

analytics_moves (
    id SERIAL PRIMARY KEY,
    game_id VARCHAR(255),
    player VARCHAR(255),
    column INTEGER,
    row INTEGER,
    move_time TIMESTAMP
)

analytics_players (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    games_played INTEGER,
    games_won INTEGER,
    games_lost INTEGER,
    last_played TIMESTAMP
)
```

### 5. Message Queue (Kafka)

**Topics**:
- `game-events` - All game events

**Event Types**:
```json
{
  "event_type": "game_started | move_made | game_finished",
  "game_id": "uuid",
  "player": "username",
  "timestamp": 1234567890
}
```

## Data Flow

### 1. Player Joins Game

```
Player → Frontend → WebSocket → Backend
                                   ↓
                              Matchmaker
                                   ↓
                         (Wait or Match with Bot)
                                   ↓
                              Game Manager
                                   ↓
                              WebSocket → Frontend → Player
```

### 2. Make Move

```
Player → Frontend → WebSocket → Backend
                                   ↓
                              Game Manager
                                   ↓
                         Validate & Update State
                                   ↓
                              ├─→ Database (Save)
                              ├─→ Kafka (Event)
                              └─→ WebSocket → All Players
```

### 3. Analytics Pipeline

```
Backend → Kafka Producer → Kafka Topic
                              ↓
                         Kafka Consumer
                              ↓
                         Analytics Service
                              ↓
                         PostgreSQL (Analytics Tables)
```

## Communication Protocols

### WebSocket Messages

**Client → Server**:
- `join` - Join matchmaking
- `move` - Make a move
- `reconnect` - Reconnect to game
- `heartbeat` - Keep connection alive

**Server → Client**:
- `player_info` - Player details
- `waiting` - Waiting for opponent
- `game_update` - Game state change
- `error` - Error message
- `reconnected` - Reconnection success

### REST API

Standard HTTP/JSON for:
- Leaderboard queries
- User statistics
- Game history
- Health checks

## Scaling Considerations

### Horizontal Scaling

1. **Frontend**: Multiple Nginx instances behind load balancer
2. **Backend**: Multiple Go servers with sticky sessions (WebSocket)
3. **Database**: Read replicas for queries
4. **Kafka**: Partitioned topics for throughput

### Caching Strategy

1. **Redis** for active game states
2. **CDN** for frontend assets
3. **Database** query caching for leaderboard

### Performance Optimizations

1. **WebSocket**: Binary protocol for efficiency
2. **Database**: Indexed queries, connection pooling
3. **Bot AI**: Depth-limited search with pruning
4. **Frontend**: React.memo, lazy loading

## Security

### Current Implementation
- CORS enabled for development
- WebSocket origin checking
- Input validation
- SQL injection prevention (parameterized queries)

### Production Recommendations
1. JWT authentication
2. Rate limiting
3. HTTPS/WSS only
4. Secret management (env vars)
5. Database credential rotation
6. API key for external access

## Monitoring & Observability

### Logs
- Structured logging in Go
- Log aggregation (ELK stack)

### Metrics
- Game duration
- Move frequency
- Player count
- Error rates
- WebSocket connections

### Health Checks
- `/api/health` endpoint
- Database connectivity
- Kafka connectivity

## Deployment

### Docker Compose (Development)
Single-host deployment with all services

### Kubernetes (Production)
- Pod for each service
- StatefulSet for database
- Ingress for routing
- ConfigMaps for configuration
- Secrets for credentials

### CI/CD Pipeline
1. Run tests
2. Build Docker images
3. Push to registry
4. Deploy to environment
5. Run smoke tests

## Future Enhancements

1. **Authentication**: User accounts and OAuth
2. **Tournaments**: Multi-player tournaments
3. **Replay**: Game replay system
4. **Chat**: In-game messaging
5. **Rankings**: ELO rating system
6. **Mobile**: Native mobile apps
7. **AI Levels**: Multiple difficulty levels
8. **Spectator Mode**: Watch live games
