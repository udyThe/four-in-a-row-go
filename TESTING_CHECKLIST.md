# Testing Checklist

Use this checklist to verify that all features are working correctly.

## ✅ Setup Verification

### Prerequisites
- [ ] Docker installed and running
- [ ] Docker Compose installed
- [ ] Ports 3000, 8080, 5432, 9092 are available

### Initial Setup
- [ ] Repository cloned successfully
- [ ] All files present in correct locations
- [ ] docker-compose.yml exists in root directory

## ✅ Application Startup

### Docker Compose Start
- [ ] Run: `docker-compose up -d --build`
- [ ] Wait 60 seconds for all services to be ready
- [ ] Run: `docker-compose ps` - all services should be "Up"
- [ ] No error messages in logs: `docker-compose logs`

### Service Health Checks
- [ ] Backend health: `curl http://localhost:8080/api/health`
  - Expected: `{"status":"healthy"}`
- [ ] Frontend loads: Open http://localhost:3000
- [ ] No console errors in browser developer tools

## ✅ Game Functionality

### Single Player vs Bot
1. [ ] Open http://localhost:3000
2. [ ] Enter username "TestPlayer1"
3. [ ] Click "Join Game"
4. [ ] See "Waiting for opponent..." message
5. [ ] Wait 10 seconds
6. [ ] Bot joins automatically
7. [ ] Bot name shows "🤖 Bot"
8. [ ] Game starts (board becomes interactive)

### Making Moves
- [ ] Click on any column
- [ ] Disc drops with animation
- [ ] Disc lands in correct position
- [ ] Turn switches to bot
- [ ] Bot makes move within 1 second
- [ ] Turn switches back to player

### Win Conditions
- [ ] Play until you or bot connects 4
- [ ] Win message displays correctly
- [ ] Winner is announced
- [ ] "Play Again" button appears

### Draw Condition
- [ ] Fill board without connecting 4
- [ ] Draw message displays
- [ ] "Play Again" button appears

### Multiplayer (Two Browser Windows)
1. [ ] Open http://localhost:3000 in Window 1
2. [ ] Enter username "Player1"
3. [ ] Click "Join Game"
4. [ ] Immediately open http://localhost:3000 in Window 2
5. [ ] Enter username "Player2"
6. [ ] Click "Join Game"
7. [ ] Both players matched immediately
8. [ ] No bot joins
9. [ ] Game starts in both windows
10. [ ] Make move in Window 1 - appears in Window 2
11. [ ] Make move in Window 2 - appears in Window 1
12. [ ] Both windows show same game state

## ✅ Reconnection Feature

### Test Reconnection
1. [ ] Start a game (vs bot is easier)
2. [ ] Note your Game ID (check browser localStorage)
3. [ ] Close browser tab
4. [ ] Reopen http://localhost:3000 within 30 seconds
5. [ ] Game state should restore
6. [ ] Can continue playing

### Test Timeout
1. [ ] Start a game
2. [ ] Close browser tab
3. [ ] Wait 31 seconds
4. [ ] Reopen browser
5. [ ] Game should be forfeited
6. [ ] Opponent declared winner

## ✅ Leaderboard

### Access Leaderboard
- [ ] Click "🏆 Leaderboard" in navigation
- [ ] Leaderboard page loads
- [ ] Shows columns: Rank, Player, Wins, Losses, Draws, Win Rate

### Verify Data
- [ ] Play 2-3 games with different results
- [ ] Click "🔄 Refresh" on leaderboard
- [ ] Player appears in leaderboard
- [ ] Win/Loss/Draw counts are correct
- [ ] Win rate calculated correctly
- [ ] Rankings sorted by wins

## ✅ Bot Intelligence

### Strategic Behavior
1. [ ] Start game vs bot
2. [ ] Create 3-in-a-row situation (almost winning)
3. [ ] Bot blocks your winning move
4. [ ] Bot creates its own 3-in-a-row
5. [ ] Bot takes winning move when available
6. [ ] Bot doesn't make random moves

### Performance
- [ ] Bot responds within 1 second
- [ ] No noticeable lag
- [ ] Bot doesn't crash the game

## ✅ API Endpoints

### REST API
```bash
# Health Check
curl http://localhost:8080/api/health
# Expected: {"status":"healthy"}

# Leaderboard
curl http://localhost:8080/api/leaderboard
# Expected: JSON array of users

# User Stats (replace TestPlayer1 with actual username)
curl http://localhost:8080/api/user/TestPlayer1
# Expected: User object with stats

# Recent Games
curl http://localhost:8080/api/games/recent
# Expected: JSON array of recent games
```

- [ ] All endpoints return valid JSON
- [ ] No 500 errors
- [ ] Data matches expected format

## ✅ WebSocket Communication

### Browser Developer Tools
1. [ ] Open browser DevTools (F12)
2. [ ] Go to Network tab
3. [ ] Filter "WS" (WebSocket)
4. [ ] Join game
5. [ ] Check WebSocket connection
   - [ ] Connection status: "Connected"
   - [ ] Messages being sent/received
   - [ ] No connection errors

### Message Flow
- [ ] "join" message sent when joining
- [ ] "player_info" message received
- [ ] "game_update" messages on each move
- [ ] "heartbeat" messages every 10 seconds
- [ ] Error messages display in UI

## ✅ Analytics (Kafka)

### Verify Kafka Producer
```bash
# Check backend logs for Kafka messages
docker-compose logs backend | grep -i kafka
```
- [ ] "Kafka producer connected successfully" appears
- [ ] No Kafka connection errors

### Verify Analytics Consumer
```bash
# Check analytics service logs
docker-compose logs analytics
```
- [ ] "Starting analytics service..." appears
- [ ] "Analytics tables initialized successfully" appears
- [ ] Events being processed

### Database Verification
```bash
# Connect to database
docker exec -it four-in-a-row-db psql -U postgres -d four_in_a_row

# Check analytics tables
SELECT COUNT(*) FROM analytics_games;
SELECT COUNT(*) FROM analytics_moves;
SELECT COUNT(*) FROM analytics_players;
```
- [ ] Analytics tables exist
- [ ] Data is being inserted
- [ ] Counts increase after playing games

## ✅ Error Handling

### Invalid Moves
- [ ] Try clicking same column when full
- [ ] Error message displays
- [ ] Game doesn't crash

### Network Issues
- [ ] Stop backend: `docker-compose stop backend`
- [ ] Try to join game
- [ ] Error message displays
- [ ] Restart backend: `docker-compose start backend`
- [ ] Can join game again

### Database Issues
- [ ] Stop postgres: `docker-compose stop postgres`
- [ ] Backend should show connection errors
- [ ] Restart postgres: `docker-compose start postgres`
- [ ] Backend reconnects automatically

## ✅ Performance

### Response Times
- [ ] Page loads in < 3 seconds
- [ ] Moves register instantly (< 100ms)
- [ ] Bot responds in < 1 second
- [ ] Leaderboard loads in < 2 seconds

### Concurrent Games
1. [ ] Open 4 browser windows
2. [ ] Start 2 games (2 players each)
3. [ ] Play simultaneously
4. [ ] No lag or crashes
5. [ ] All games work independently

### Memory Usage
```bash
# Check container resource usage
docker stats
```
- [ ] Backend < 100MB RAM
- [ ] Frontend < 50MB RAM
- [ ] Database < 100MB RAM
- [ ] No memory leaks after extended use

## ✅ UI/UX

### Responsive Design
- [ ] Resize browser window
- [ ] Layout adjusts properly
- [ ] All elements remain accessible
- [ ] Text remains readable

### Mobile View
- [ ] Open on mobile browser (or use DevTools device emulation)
- [ ] Game board displays correctly
- [ ] Can click columns
- [ ] Navigation works

### Animations
- [ ] Disc dropping animation works
- [ ] Smooth transitions
- [ ] No visual glitches
- [ ] Hover effects work

### User Feedback
- [ ] Turn indicator shows whose turn it is
- [ ] Waiting message displays
- [ ] Win/loss/draw messages clear
- [ ] Error messages visible and helpful

## ✅ Documentation

### README.md
- [ ] All sections present
- [ ] Code examples work
- [ ] Links work
- [ ] Instructions clear

### Other Docs
- [ ] ARCHITECTURE.md exists and comprehensive
- [ ] DEPLOYMENT.md has deployment guides
- [ ] QUICK_START.md easy to follow
- [ ] PROJECT_SUMMARY.md complete

## ✅ Code Quality

### Backend
- [ ] Code compiles without warnings
- [ ] Proper error handling
- [ ] Logging in place
- [ ] Comments where needed

### Frontend
- [ ] No console errors
- [ ] No console warnings (except expected ones)
- [ ] Proper component structure
- [ ] Clean code style

## ✅ Cleanup & Restart

### Stop Services
```bash
docker-compose down
```
- [ ] All services stop cleanly
- [ ] No error messages

### Remove Volumes
```bash
docker-compose down -v
```
- [ ] Volumes removed
- [ ] Database data cleared

### Restart Fresh
```bash
docker-compose up -d --build
```
- [ ] All services start from clean state
- [ ] Database migrations run
- [ ] Application works as expected

## 📝 Test Results Summary

| Category | Pass | Fail | Notes |
|----------|------|------|-------|
| Setup | ☐ | ☐ | |
| Startup | ☐ | ☐ | |
| Game Functionality | ☐ | ☐ | |
| Reconnection | ☐ | ☐ | |
| Leaderboard | ☐ | ☐ | |
| Bot Intelligence | ☐ | ☐ | |
| API Endpoints | ☐ | ☐ | |
| WebSocket | ☐ | ☐ | |
| Analytics | ☐ | ☐ | |
| Error Handling | ☐ | ☐ | |
| Performance | ☐ | ☐ | |
| UI/UX | ☐ | ☐ | |
| Documentation | ☐ | ☐ | |
| Code Quality | ☐ | ☐ | |

## 🐛 Known Issues

Document any issues found during testing:

1. Issue:
   - Description:
   - Severity: [Critical/High/Medium/Low]
   - Workaround:

## ✅ Final Verification

Before submission:
- [ ] All critical tests pass
- [ ] No major bugs
- [ ] Documentation complete
- [ ] Code is clean
- [ ] Ready for deployment
- [ ] GitHub repository updated
- [ ] README has live URL (if deployed)

## 📊 Test Coverage

- **Total Tests**: 100+
- **Tests Passed**: ___
- **Tests Failed**: ___
- **Coverage**: ___%

---

**Tester**: _______________
**Date**: _______________
**Result**: ☐ PASS | ☐ FAIL
