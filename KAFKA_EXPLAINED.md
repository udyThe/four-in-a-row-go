# Kafka Integration in 4-in-a-Row Game

## What is Kafka?

Apache Kafka is a **distributed event streaming platform** that acts as a message queue. Think of it as a post office:
- Backend sends messages (game events) to Kafka
- Analytics service receives these messages from Kafka
- Messages are stored reliably, so no events are lost

## How Kafka Works in This Project

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Backend   │ ──────> │    Kafka    │ ──────> │  Analytics  │ ──────> │  PostgreSQL │
│  (Go API)   │ Events  │   Broker    │ Events  │  Consumer   │  Store  │  Database   │
└─────────────┘         └─────────────┘         └─────────────┘         └─────────────┘
      │                       │                         │                       │
      │                       │                         │                       │
   Produces              Stores in                 Consumes                Persists
   Events                 Topics                   Events                  Data
```

## Architecture Flow

### 1. **Event Production (Backend)**
When game actions happen, the backend emits events:

**File**: `backend/internal/game/manager.go`

```go
// When game starts
emitGameStartedEvent() → Kafka Topic: "game-events"

// When player makes a move
emitMoveEvent() → Kafka Topic: "game-events"

// When game ends
emitGameEndedEvent() → Kafka Topic: "game-events"
```

**Event Types Sent**:
- `game_started`: When two players are matched
- `move`: Each time a disc is placed
- `game_ended`: When game finishes (win/draw)

### 2. **Kafka Broker**
**Service**: `kafka` container (port 9092)

- Receives events from backend
- Stores them in **"game-events"** topic
- Guarantees event delivery (no data loss)
- Runs on Zookeeper for coordination

### 3. **Event Consumption (Analytics)**
**File**: `analytics/main.go`

The analytics service:
- Continuously reads from "game-events" topic
- Processes each event
- Stores in PostgreSQL tables:
  - `game_analytics`: All events log
  - `game_stats`: Aggregated statistics

### 4. **Data Storage**
Processed data enables:
- Leaderboard rankings
- Win/loss statistics
- Game history tracking
- Player analytics

## Kafka Components in Docker Compose

```yaml
zookeeper:
  image: confluentinc/cp-zookeeper:7.5.0
  # Manages Kafka cluster coordination

kafka:
  image: confluentinc/cp-kafka:7.5.0
  # Message broker for event streaming

kafka-ui:
  image: provectuslabs/kafka-ui:latest
  # Web interface to visualize Kafka (http://localhost:8090)

analytics:
  build: ./analytics
  # Consumes events from Kafka
```

## Kafka Web UI (NEW!)

**Access**: http://localhost:8090

The Kafka UI provides a visual interface to:
- ✅ View all Kafka topics (including "game-events")
- ✅ See messages in real-time as they arrive
- ✅ Monitor consumer groups and lag
- ✅ Browse message content (JSON events)
- ✅ Track partition offsets
- ✅ View broker health and configuration

This is the **easiest proof** that Kafka is working - just open the browser and see live events!

## Event Flow Example

```
1. User plays game:
   Player1 vs Player2 → Makes moves → Game ends

2. Backend emits events:
   ┌──────────────────────────────────────┐
   │ Event: game_started                  │
   │ {                                    │
   │   "event_type": "game_started",      │
   │   "game_id": "abc123",               │
   │   "player1": "Alice",                │
   │   "player2": "Bob",                  │
   │   "timestamp": 1234567890            │
   │ }                                    │
   └──────────────────────────────────────┘
                    ↓
   ┌──────────────────────────────────────┐
   │ Event: move (x7 moves)               │
   │ {                                    │
   │   "event_type": "move",              │
   │   "player": "Alice",                 │
   │   "column": 3,                       │
   │   "row": 5                           │
   │ }                                    │
   └──────────────────────────────────────┘
                    ↓
   ┌──────────────────────────────────────┐
   │ Event: game_ended                    │
   │ {                                    │
   │   "event_type": "game_ended",        │
   │   "winner": "Alice",                 │
   │   "result": "win",                   │
   │   "duration": 45.2                   │
   │ }                                    │
   └──────────────────────────────────────┘

3. Kafka stores all events in order

4. Analytics reads events and updates database:
   - game_analytics table gets 9 rows (1 start + 7 moves + 1 end)
   - game_stats updated: Alice wins++, Bob losses++
   
5. Leaderboard displays updated stats
```

## Why Use Kafka?

### Benefits in This Project:

1. **Decoupling**: Backend doesn't wait for analytics processing
2. **Reliability**: Events stored even if analytics service is down
3. **Scalability**: Can add multiple analytics consumers
4. **Real-time**: Events processed as they happen
5. **Audit Trail**: Complete history of all game actions

### Without Kafka:
```
Backend → Database (direct write)
- Backend waits for DB write
- No event replay capability
- Tight coupling
```

### With Kafka:
```
Backend → Kafka → Analytics → Database
- Backend continues immediately
- Can replay events if needed
- Services independent
```

## Verification Methods

### Method 1: Check Kafka Logs
```powershell
docker-compose logs kafka | Select-String "game-events"
```

### Method 2: Check Analytics Processing
```powershell
docker-compose logs analytics
```
Look for: "Processing event: game_started", "Processing event: move"

### Method 3: Query Database
```powershell
docker-compose exec db psql -U postgres -d four_in_a_row -c "SELECT * FROM game_analytics LIMIT 10;"
```

### Method 4: Use Verification Script
```powershell
.\verify-kafka.ps1
```

## Proof of Kafka Working

To demonstrate Kafka is working for your submission:

### **BEST: Use Kafka Web UI** (Recommended!)
1. Open browser: http://localhost:8090
2. Navigate to "Topics" → "game-events"
3. Click "Messages" tab
4. Screenshot showing live events flowing through Kafka
5. Play a game and watch new messages appear in real-time!

### Alternative Methods:

1. **Screenshot Container Status**:
   ```powershell
   docker-compose ps
   ```
   Show: kafka, zookeeper, analytics all running

2. **Screenshot Analytics Logs**:
   ```powershell
   docker-compose logs analytics --tail=20
   ```
   Show: Events being consumed

3. **Screenshot Database Records**:
   ```powershell
   docker-compose exec db psql -U postgres -d four_in_a_row -c "SELECT event_type, COUNT(*) FROM game_analytics GROUP BY event_type;"
   ```
   Show: Events stored via Kafka

4. **Play a Game**:
   - Start game, make moves, finish
   - Check analytics logs show new events
   - Query database shows new records

## Files Involved

```
Project Structure:
├── backend/
│   └── internal/
│       ├── kafka/
│       │   └── producer.go        # Kafka producer
│       └── game/
│           └── manager.go         # Emits events
├── analytics/
│   └── main.go                    # Kafka consumer
├── docker-compose.yml             # Kafka setup
└── verify-kafka.ps1               # Verification script
```

## Technical Details

**Kafka Topic**: `game-events`
**Port**: 9092 (internal), exposed to services
**Message Format**: JSON
**Consumer Group**: `analytics-group`
**Persistence**: Messages stored until consumed

## Common Issues & Solutions

1. **Kafka not connecting**:
   - Check zookeeper is running first
   - Wait 30s after docker-compose up

2. **No events in database**:
   - Play a game to generate events
   - Check analytics container logs

3. **Analytics not consuming**:
   - Restart analytics: `docker-compose restart analytics`
   - Check Kafka broker accessible

## Summary

Kafka enables **event-driven architecture** in this project:
- Asynchronous event processing
- Reliable message delivery
- Scalable analytics pipeline
- Complete audit trail of game actions

This is professional, production-grade architecture used by companies like LinkedIn, Uber, and Netflix for real-time data streaming.
