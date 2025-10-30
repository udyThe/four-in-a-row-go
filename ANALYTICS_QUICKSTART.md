# Quick Start: Testing Analytics Feature

## 🚀 In 5 Minutes

### Step 1: Start All Services (1 min)
```bash
cd UdayAssignment-2
docker-compose up
```

Wait for all services to start (you'll see "Server starting on port 8080").

### Step 2: Play Some Games (2 min)
1. Open browser: http://localhost:3000
2. Click "Play"
3. Enter username: "TestPlayer1"
4. Wait 10 seconds for bot opponent
5. Play 3-5 games quickly
6. Make various moves (don't worry about winning/losing)

### Step 3: View Analytics Dashboard (1 min)
1. Click "Analytics" in navigation bar
2. You'll see the analytics dashboard
3. Click "Hourly (24h)" to see last 24 hours
4. Click "Daily (30d)" to see daily trends

### Step 4: View Kafka Events (1 min)
1. Open new tab: http://localhost:8090
2. Click "four-in-a-row" cluster
3. Click "Topics" → "game-events"
4. Click "Messages" tab
5. See your game events streaming!

## ✅ What to Expect

### Analytics Dashboard Should Show:
- **Summary Cards**: 
  - Games Started: ~5
  - Games Completed: ~4
  - Total Moves: ~80-100
  - Avg Duration: ~30-60s

- **Data Table**:
  - Current hour with your game data
  - All metrics populated

- **Bar Chart**:
  - One or two bars showing activity
  - Hover to see exact numbers

### Kafka UI Should Show:
- **Topics**: `game-events` with messages
- **Consumer Groups**: `analytics-consumer` (active)
- **Messages**: JSON events like:
  ```json
  {
    "event_type": "game_started",
    "game_id": "abc123",
    "player1": "TestPlayer1",
    "player2": "Bot",
    "timestamp": 1730304000
  }
  ```

## 🎯 Quick Tests

### Test 1: Real-time Updates
1. Keep Analytics page open
2. Play 2 more games
3. Wait 30 seconds
4. Numbers should update automatically

### Test 2: Hourly View
1. On Analytics page, ensure "Hourly (24h)" is selected
2. See current hour with data
3. Previous hours should be 0 (no data yet)

### Test 3: Daily View
1. Click "Daily (30d)" button
2. See today's date with data
3. Previous days should be 0 (no data yet)

### Test 4: Kafka Pipeline
1. Open Kafka UI (http://localhost:8090)
2. Go to Topics → game-events → Messages
3. Play another game in different tab
4. Refresh Kafka UI
5. See new events appear

## 📊 Sample Output

After playing 5 games, you should see:

```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│      5      │      4      │     82      │    45.3s    │
│   Started   │  Completed  │   Moves     │  Duration   │
└─────────────┴─────────────┴─────────────┴─────────────┘

Peak Activity: Today, 2:00 PM - 5 games started

Hour              Started  Completed  Moves  Duration
Oct 30, 2:00 PM      5        4        82     45.3s
Oct 30, 1:00 PM      0        0         0       -
Oct 30, 12:00 PM     0        0         0       -
```

## 🔍 Verification Checklist

- [ ] Docker containers all running (7 services)
- [ ] Frontend accessible at http://localhost:3000
- [ ] Kafka UI accessible at http://localhost:8090
- [ ] Can play games successfully
- [ ] Analytics tab shows in navigation
- [ ] Summary cards display numbers > 0
- [ ] Data table has at least one row
- [ ] Bar chart shows bars
- [ ] Kafka UI shows game-events topic
- [ ] Kafka messages visible in Messages tab

## 🐛 Quick Troubleshooting

### Analytics Shows "No data"
**Solution**: Play 1-2 games first. Analytics needs events to aggregate.

### Kafka UI Not Loading
**Solution**: Wait 1-2 minutes for Kafka to fully start, then refresh.

### Games Not Starting
**Solution**: Check backend logs: `docker-compose logs backend`

### Frontend Connection Error
**Solution**: Ensure all services running: `docker-compose ps`

## 📸 Screenshots to Take

For evaluator submission, capture:
1. **Analytics Dashboard** - Showing summary cards and chart
2. **Kafka UI Topics** - Showing game-events topic
3. **Kafka Messages** - Showing actual event JSON
4. **Analytics Table** - Showing hourly/daily data

## 🎉 Success Indicators

You've successfully implemented and tested analytics if:
- ✅ Can toggle between Hourly/Daily views
- ✅ Summary cards show actual game data
- ✅ Chart displays bars for active hours
- ✅ Kafka UI shows game-events messages
- ✅ Data updates automatically (30s refresh)
- ✅ Peak activity time is identified

## 🔄 Reset and Test Again

To test with fresh data:
```bash
# Stop all services
docker-compose down -v

# Start fresh
docker-compose up

# Play games again
# View new analytics
```

This clears all data and lets you see the analytics build up from zero.

## 💡 Pro Tips

1. **Generate More Data**: Play 10-15 games to see better charts
2. **Test Peak Detection**: Play 5 games at once, then 1 game an hour later
3. **Check Consumer Lag**: In Kafka UI, go to Consumers → analytics-consumer
4. **Monitor Logs**: `docker-compose logs -f analytics` to see event processing
5. **API Testing**: `curl http://localhost:8080/api/analytics/hourly | jq`

## 📝 What This Proves

This feature demonstrates:
- ✅ Kafka event streaming (real-time)
- ✅ Stream processing (analytics service)
- ✅ Time-based aggregation (hourly/daily)
- ✅ Database persistence (PostgreSQL)
- ✅ RESTful APIs (Go backend)
- ✅ Modern UI (React + charts)
- ✅ Full stack integration (end-to-end)

Perfect for showing evaluators that Kafka is working and processing events!

---

**Total Time**: 5 minutes
**Difficulty**: Easy
**Result**: Beautiful analytics dashboard powered by Kafka! 🎊
