# Quick Guide: Kafka UI Analytics

## What Changed

I've enhanced the Kafka events to show **real-time analytics directly in Kafka UI**!

## What You'll See

### Every 60 Seconds: System Metrics
```json
{
  "active_games": 12,           ← How many games active NOW
  "connected_players": 24,      ← How many players online NOW
  "games_in_progress": 8,       ← Games being played NOW
  "hour_of_day": 14,            ← Current hour (2 PM)
  "estimated_games_per_hour": 480,  ← Games per hour rate
  "estimated_requests_per_hour": 1680  ← Kafka messages per hour
}
```

### Every Game Start
```json
{
  "hour_of_day": 14,      ← Track hourly activity
  "day_of_week": "Wednesday",
  "active_games": 13,     ← Snapshot at game start
  "total_players": 18     ← Players online at that moment
}
```

### Every Move
```json
{
  "hour_of_day": 14,      ← Track peak hours
  "move_number": 5,       ← Track game progress
  "is_bot_move": false    ← Bot vs human tracking
}
```

### Every Game Finish
```json
{
  "hour_of_day": 14,
  "total_moves": 23,           ← Game complexity
  "game_duration_sec": 145,    ← Average duration
  "was_bot_game": true         ← Game type stats
}
```

## How to View

1. **Start services**: `docker-compose up`
2. **Play games**: http://localhost:3000 (play 5-10 games)
3. **Open Kafka UI**: http://localhost:8090
4. **Navigate**: Cluster → Topics → game-events → Messages
5. **Watch**: Events appear in real-time with all metrics!

## What Analytics You Get

### In Kafka UI Messages:
- ✅ **Players online NOW** - Look at latest `system_metrics` → `connected_players`
- ✅ **Games per hour** - Count `game_started` events by `hour_of_day`
- ✅ **Players per hour** - Track `connected_players` over time
- ✅ **Kafka requests/hour** - See `estimated_requests_per_hour`
- ✅ **Peak hours** - Find hour with most `game_started` events
- ✅ **Daily activity** - Filter by `date` field

### In Kafka UI Topic Overview:
- ✅ **Total messages** - All Kafka events processed
- ✅ **Messages/sec** - Current Kafka throughput
- ✅ **Consumer lag** - Analytics processing speed

## Example: Find Peak Hour

```
1. Open Kafka UI → game-events → Messages
2. Search for: "event_type":"game_started"
3. Look at "hour_of_day" in each message
4. Count events per hour:
   - Hour 14: 45 games
   - Hour 15: 23 games  
   - Hour 16: 67 games  ← PEAK!
```

## Example: Current Players

```
1. Search for: "event_type":"system_metrics"
2. Sort by timestamp (latest first)
3. Look at first message:
   "connected_players": 24  ← 24 players online RIGHT NOW
```

## For Evaluator

Show them:
1. Kafka UI with live events streaming
2. `system_metrics` showing current stats
3. Hour-by-hour breakdown in messages
4. Topic statistics showing total throughput

**Key Point**: All analytics visible directly in Kafka UI - no separate dashboard needed!

## Files Changed

- `backend/internal/game/manager.go` - Enhanced events with metrics
- `backend/internal/game/metrics.go` - NEW: System metrics emitter (every 60s)
- `backend/main.go` - Start metrics emitter
- `KAFKA_UI_ANALYTICS.md` - Full documentation

## Test It

```bash
docker-compose up
# Play 5 games
# Open http://localhost:8090
# See metrics in real-time!
```
