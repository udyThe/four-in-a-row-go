# 🎮 Four in a Row (Connect Four)

A real-time multiplayer Connect Four game with AI opponent, built using Go, React, PostgreSQL, and Kafka.

## Features

- Play against AI bot or real players
- Real-time multiplayer via WebSocket
- Automatic matchmaking
- Leaderboard and analytics dashboard
- Event-driven architecture with Kafka

## Tech Stack

- **Backend**: Go 1.23, WebSocket, PostgreSQL
- **Frontend**: React 18.2
- **Infrastructure**: Kafka, Zookeeper, Docker

## Quick Start

**Prerequisites:** Docker and Docker Compose

1. **Clone and start**
   ```bash
   git clone git@github.com:udyThe/four-in-a-row-go.git
   cd four-in-a-row-go
   docker-compose up -d
   ```

2. **Access the app**
   - Game: http://localhost:3000
   - Kafka UI: http://localhost:8090

3. **Stop**
   ```bash
   docker-compose down
   ```

## How It Works

- **Backend** (Port 8080): Go service handling game logic, WebSocket connections, and REST API
- **Frontend** (Port 3000): React app with game interface, leaderboard, and analytics
- **Analytics** (Background): Consumes Kafka events to track game statistics
- **Database** (Port 5432): PostgreSQL stores games, players, and analytics
- **Kafka** (Port 9092): Event streaming for real-time analytics

## How to Play

1. Open http://localhost:3000
2. Enter your username
3. Choose "Play vs Bot" or "Find Match"
4. Click columns to drop discs
5. Connect 4 discs horizontally, vertically, or diagonally to win

## 👨‍💻 Author

Created by Uday

---

Enjoy playing Four in a Row! 🎉
