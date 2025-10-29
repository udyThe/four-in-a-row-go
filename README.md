# 🎮 4 in a Row - Real-Time Multiplayer Game

A modern, real-time implementation of the classic **Connect Four** game built with **Go** (backend) and **React** (frontend). Features include competitive bot AI, player matchmaking, live gameplay via WebSockets, and game analytics powered by Kafka.

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Prerequisites](#-prerequisites)
- [Installation & Setup](#-installation--setup)
- [Running the Application](#-running-the-application)
- [API Documentation](#-api-documentation)
- [Game Rules](#-game-rules)
- [Project Structure](#-project-structure)
- [Deployment](#-deployment)
- [Screenshots](#-screenshots)

## ✨ Features

### Core Gameplay
- ✅ Real-time 1v1 multiplayer gameplay via WebSockets
- ✅ Automatic matchmaking with 10-second timeout
- ✅ Competitive bot opponent using Minimax algorithm with alpha-beta pruning
- ✅ Player reconnection support (30-second grace period)
- ✅ Live game state synchronization

### Bot Intelligence
- 🧠 Strategic decision-making using Minimax algorithm
- 🎯 Blocks opponent's winning moves
- 💡 Creates winning opportunities
- ⚡ Optimized with alpha-beta pruning (depth 6)
- 🏆 Evaluates board position with heuristics

### Analytics & Tracking
- 📊 Kafka-based event streaming for game analytics
- 📈 Track game duration, moves, and player statistics
- 🏅 Real-time leaderboard with win/loss/draw statistics
- 📉 Player-specific metrics and game history

### User Experience
- 🎨 Clean, responsive UI with animations
- 🔄 Automatic game state persistence
- ⏱️ Heartbeat system for connection monitoring
- 🎭 Visual feedback for turns and game status

## 🛠 Tech Stack

### Backend
- **Go 1.21+** - High-performance game server
- **Gorilla WebSocket** - Real-time bidirectional communication
- **Gorilla Mux** - HTTP routing
- **PostgreSQL** - Game and user data persistence
- **pgx/v5** - PostgreSQL driver
- **Kafka** - Event streaming for analytics
- **Docker** - Containerization

### Frontend
- **React 18** - UI framework
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **WebSocket API** - Real-time game updates
- **CSS3** - Styling and animations

### Infrastructure
- **Docker Compose** - Multi-container orchestration
- **Nginx** - Frontend serving and reverse proxy
- **Zookeeper** - Kafka coordination

## 🏗 Architecture

```
┌─────────────┐         WebSocket          ┌──────────────┐
│   React     │◄──────────────────────────►│   Go Server  │
│  Frontend   │         HTTP/REST          │   (Backend)  │
└─────────────┘◄──────────────────────────►└──────────────┘
                                                    │
                                                    ▼
                                            ┌──────────────┐
                                            │  PostgreSQL  │
                                            │   Database   │
                                            └──────────────┘
                                                    ▲
┌─────────────┐         Kafka Events              │
│  Analytics  │◄──────────────────────────────────┘
│  Consumer   │         (Game Metrics)
└─────────────┘
```

### Key Components

1. **Game Manager** - Manages active games in memory
2. **Matchmaker** - Handles player queuing and bot assignment
3. **WebSocket Handler** - Real-time communication
4. **Bot Engine** - AI opponent with Minimax algorithm
5. **Database Layer** - Persistence and leaderboard
6. **Kafka Producer/Consumer** - Analytics pipeline

## 📦 Prerequisites

### Required
- **Docker** (v20.10+) & **Docker Compose** (v2.0+)
- **Go** (v1.21+) - for local development
- **Node.js** (v18+) & **npm** - for frontend development

### Optional (for local development without Docker)
- **PostgreSQL** (v15+)
- **Apache Kafka** (v3.5+)

## 🚀 Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd UdayAssignment
```

### 2. Environment Configuration

Create `.env` files if needed (optional - defaults are configured):

**Backend (.env)**
```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/four_in_a_row?sslmode=disable
KAFKA_BROKER=kafka:29092
```

**Frontend (.env)**
```env
REACT_APP_API_URL=http://localhost:8080/api
REACT_APP_WS_URL=ws://localhost:8080/ws
```

## 🏃 Running the Application

### Using Docker Compose (Recommended)

This will start all services (Backend, Frontend, PostgreSQL, Kafka, Analytics):

```bash
# Build and start all services
docker-compose up --build

# Run in detached mode
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

**Access Points:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api
- WebSocket: ws://localhost:8080/ws
- **Kafka UI: http://localhost:8090** (View live event streaming)
- PostgreSQL: localhost:5432

### Local Development

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/four_in_a_row?sslmode=disable"
export KAFKA_BROKER="localhost:9092"
export PORT="8080"

# Run the server
go run main.go
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start
```

#### Analytics Service

```bash
cd analytics

# Install dependencies
go mod download

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/four_in_a_row?sslmode=disable"
export KAFKA_BROKER="localhost:9092"

# Run the analytics consumer
go run main.go
```

## 📡 API Documentation

### REST Endpoints

#### Health Check
```
GET /api/health
Response: { "status": "healthy", "timestamp": "ok" }
```

#### Get Leaderboard
```
GET /api/leaderboard?limit=10
Response: Array of users with stats
```

#### Get User Stats
```
GET /api/user/{username}
Response: User object with game statistics
```

#### Get Recent Games
```
GET /api/games/recent?limit=20
Response: Array of recent games
```

#### Get User Games
```
GET /api/games/user/{username}?limit=20
Response: Array of games for specific user
```

### WebSocket Messages

#### Client → Server

**Join Game**
```json
{
  "type": "join",
  "payload": {
    "username": "Player1"
  }
}
```

**Make Move**
```json
{
  "type": "move",
  "payload": {
    "column": 3
  }
}
```

**Reconnect**
```json
{
  "type": "reconnect",
  "payload": {
    "player_id": "uuid",
    "game_id": "uuid"
  }
}
```

**Heartbeat**
```json
{
  "type": "heartbeat",
  "payload": {}
}
```

#### Server → Client

**Player Info**
```json
{
  "type": "player_info",
  "payload": {
    "player_id": "uuid",
    "game_id": "uuid",
    "username": "Player1"
  }
}
```

**Game Update**
```json
{
  "type": "game_update",
  "payload": {
    "id": "uuid",
    "player1": {...},
    "player2": {...},
    "board": [[0,0,...], ...],
    "current_turn": 1,
    "status": "in_progress"
  }
}
```

**Error**
```json
{
  "type": "error",
  "payload": {
    "message": "Error description"
  }
}
```

## 🎲 Game Rules

1. **Board**: 7 columns × 6 rows grid
2. **Players**: 2 players take turns (Player 1: Red, Player 2: Yellow)
3. **Objective**: Connect 4 discs vertically, horizontally, or diagonally
4. **Turns**: Players alternate dropping discs into columns
5. **Gravity**: Discs fall to the lowest available position
6. **Win**: First to connect 4 wins
7. **Draw**: Board fills with no winner

### Matchmaking Rules
- Player joins and enters queue
- If another player waiting: instant match
- If no player after 10 seconds: bot joins
- Disconnected players have 30 seconds to reconnect
- After 30 seconds: opponent wins by forfeit

## 📁 Project Structure

```
UdayAssignment/
├── backend/                    # Go backend server
│   ├── internal/
│   │   ├── api/               # HTTP & WebSocket handlers
│   │   │   ├── server.go      # REST API server
│   │   │   └── websocket.go   # WebSocket handler
│   │   ├── game/              # Game logic
│   │   │   ├── board.go       # Board state & validation
│   │   │   ├── bot.go         # AI bot with Minimax
│   │   │   ├── game.go        # Game state management
│   │   │   ├── manager.go     # Game lifecycle
│   │   │   └── matchmaker.go  # Player matchmaking
│   │   ├── database/          # PostgreSQL layer
│   │   │   └── database.go    # DB operations
│   │   ├── kafka/             # Kafka integration
│   │   │   └── producer.go    # Event producer
│   │   └── config/            # Configuration
│   │       └── config.go
│   ├── main.go                # Entry point
│   ├── go.mod                 # Go dependencies
│   └── Dockerfile
│
├── frontend/                   # React frontend
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   ├── components/        # React components
│   │   │   ├── Game.js        # Main game component
│   │   │   ├── GameBoard.js   # Board UI
│   │   │   └── Leaderboard.js # Leaderboard display
│   │   ├── services/          # API & WebSocket clients
│   │   │   ├── api.js         # REST API client
│   │   │   └── websocket.js   # WebSocket service
│   │   ├── App.js             # Root component
│   │   ├── index.js           # Entry point
│   │   └── index.css          # Global styles
│   ├── package.json           # npm dependencies
│   ├── nginx.conf             # Nginx configuration
│   └── Dockerfile
│
├── analytics/                  # Kafka consumer service
│   ├── main.go                # Analytics consumer
│   ├── go.mod
│   └── Dockerfile
│
├── docker-compose.yml          # Multi-container setup
└── README.md                   # This file
```

## 🚢 Deployment

### Render.com Deployment (Recommended - Free Tier)

For detailed step-by-step instructions, see **[RENDER_DEPLOYMENT_GUIDE.md](./RENDER_DEPLOYMENT_GUIDE.md)**

**Quick Overview:**
1. Create Render account and connect GitHub
2. Deploy PostgreSQL database (free)
3. Deploy backend service (Docker, free tier)
4. Deploy frontend service (Docker, free tier)
5. Configure environment variables
6. Access your live game at `https://your-app.onrender.com`

**Note**: Render's free tier is perfect for this project but doesn't support Kafka. The game works perfectly without Kafka - you just won't have the analytics dashboard.

### Docker Hub Deployment

```bash
# Build images
docker build -t yourusername/4-in-a-row-backend:latest ./backend
docker build -t yourusername/4-in-a-row-frontend:latest ./frontend
docker build -t yourusername/4-in-a-row-analytics:latest ./analytics

# Push to Docker Hub
docker push yourusername/4-in-a-row-backend:latest
docker push yourusername/4-in-a-row-frontend:latest
docker push yourusername/4-in-a-row-analytics:latest
```

### Cloud Deployment (AWS, Azure, GCP)

1. **Container Orchestration**: Use Kubernetes or ECS
2. **Database**: Managed PostgreSQL (RDS, Cloud SQL)
3. **Message Queue**: Managed Kafka (MSK, Confluent Cloud)
4. **Load Balancer**: For WebSocket scaling
5. **Environment Variables**: Configure per environment

### Heroku Deployment

```bash
# Backend
heroku create 4-in-a-row-backend
heroku addons:create heroku-postgresql:hobby-dev
heroku addons:create heroku-kafka:basic-0
git subtree push --prefix backend heroku master

# Frontend
heroku create 4-in-a-row-frontend
git subtree push --prefix frontend heroku master
```

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./... -v
```

### Frontend Tests
```bash
cd frontend
npm test
```

### Load Testing
```bash
# Install artillery
npm install -g artillery

# Run WebSocket load test
artillery quick --count 10 --num 50 ws://localhost:8080/ws
```

## 🐛 Troubleshooting

### Common Issues

**Port already in use**
```bash
# Check what's using the port
netstat -ano | findstr :8080
# Kill the process
taskkill /PID <PID> /F
```

**Database connection failed**
```bash
# Verify PostgreSQL is running
docker-compose ps postgres
# Check logs
docker-compose logs postgres
```

**Kafka connection failed**
```bash
# Wait for Kafka to be fully ready (30-60 seconds)
docker-compose logs kafka
# Restart if needed
docker-compose restart kafka
```

**WebSocket not connecting**
- Check browser console for errors
- Verify backend is running: `curl http://localhost:8080/api/health`
- Check firewall/antivirus settings

## 📝 Development Notes

### Bot AI Details
The bot uses **Minimax with Alpha-Beta Pruning**:
- **Depth**: 6 moves ahead
- **Evaluation**: Heuristic scoring based on patterns
- **Priority**: Winning > Blocking > Building
- **Performance**: ~100ms per move

### Kafka Topics
- `game-events`: All game events (started, move_made, finished)

### Database Schema
- `users`: Player profiles and statistics
- `games`: Completed game records
- `analytics_games`: Game metrics
- `analytics_moves`: Move history
- `analytics_players`: Player analytics

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is part of a backend engineering assignment.

## 👤 Author

**Uday Assignment**
- Built with ❤️ using Go and React
- Real-time multiplayer game with competitive AI

## 🙏 Acknowledgments

- Minimax algorithm for game AI
- Gorilla WebSocket for real-time communication
- React community for frontend patterns
- Apache Kafka for event streaming

---

**⭐ If you like this project, please give it a star!**
