# Quick Reference Guide

## 🚀 Quick Commands

### Docker Commands
```powershell
# Start everything
docker-compose up -d --build

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Stop everything
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Restart a service
docker-compose restart backend
```

### Local Development

#### Backend
```powershell
cd backend
go run main.go
```

#### Frontend
```powershell
cd frontend
npm start
```

#### Analytics
```powershell
cd analytics
go run main.go
```

## 🔧 Testing the Application

### 1. Start the Application
```powershell
docker-compose up -d --build
```

### 2. Wait for Services (30-60 seconds)
Check service health:
```powershell
docker-compose ps
```

### 3. Open Browser
- Go to: http://localhost:3000
- Enter a username
- Wait for opponent or bot (10 seconds)
- Play the game!

### 4. Test API
```powershell
# Health check
curl http://localhost:8080/api/health

# Get leaderboard
curl http://localhost:8080/api/leaderboard
```

## 🎮 Game Flow

1. **Join Game**
   - Enter username
   - Click "Join Game"

2. **Matchmaking**
   - Wait for another player (max 10 seconds)
   - Bot joins automatically if no player

3. **Playing**
   - Click on column to drop disc
   - Wait for opponent's turn
   - See real-time updates

4. **Game End**
   - Win: Connect 4 discs
   - Draw: Board fills up
   - Forfeit: Player disconnects for 30+ seconds

5. **Play Again**
   - Click "Play Again" button

## 🐛 Troubleshooting

### Backend not starting
```powershell
# Check logs
docker-compose logs backend

# Common issue: Database not ready
# Solution: Wait 30 seconds and restart
docker-compose restart backend
```

### Frontend not loading
```powershell
# Check logs
docker-compose logs frontend

# Rebuild frontend
docker-compose up -d --build frontend
```

### WebSocket not connecting
- Check browser console for errors
- Verify backend is running: `docker-compose ps`
- Check firewall settings

### Database connection error
```powershell
# Restart postgres
docker-compose restart postgres

# Check postgres logs
docker-compose logs postgres
```

### Kafka connection error
```powershell
# Kafka takes time to start (30-60 seconds)
# Check status
docker-compose logs kafka

# Restart if needed
docker-compose restart kafka
```

## 📊 Monitoring

### View All Logs
```powershell
docker-compose logs -f
```

### View Specific Service
```powershell
docker-compose logs -f backend
docker-compose logs -f kafka
docker-compose logs -f postgres
```

### Check Service Status
```powershell
docker-compose ps
```

### Database Access
```powershell
docker exec -it four-in-a-row-db psql -U postgres -d four_in_a_row
```

Useful SQL queries:
```sql
-- View all games
SELECT * FROM games ORDER BY created_at DESC LIMIT 10;

-- View leaderboard
SELECT * FROM users ORDER BY games_won DESC LIMIT 10;

-- View analytics
SELECT * FROM analytics_games ORDER BY created_at DESC LIMIT 10;
```

## 🔑 Key Endpoints

### Frontend
- **Main Game**: http://localhost:3000
- **Leaderboard**: http://localhost:3000/leaderboard

### Backend
- **Health Check**: http://localhost:8080/api/health
- **Leaderboard**: http://localhost:8080/api/leaderboard
- **User Stats**: http://localhost:8080/api/user/{username}
- **Recent Games**: http://localhost:8080/api/games/recent
- **WebSocket**: ws://localhost:8080/ws

## 📝 Environment Variables

### Backend (.env)
```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/four_in_a_row?sslmode=disable
KAFKA_BROKER=kafka:29092
```

### Frontend (.env)
```
REACT_APP_API_URL=http://localhost:8080/api
REACT_APP_WS_URL=ws://localhost:8080/ws
```

## 🎯 Performance Tips

1. **First startup** takes 2-3 minutes (downloading images)
2. **Subsequent startups** take 30-60 seconds
3. **Kafka** needs 30-60 seconds to be fully ready
4. **Database migrations** run automatically on backend start
5. Use `docker-compose down -v` to clean up and start fresh

## 🔐 Security Notes

For production deployment:
1. Change database credentials
2. Use environment variables for secrets
3. Enable CORS restrictions in backend
4. Use HTTPS/WSS for connections
5. Add authentication/authorization
6. Rate limiting on API endpoints

## 📦 Ports Used

- **3000**: Frontend (React)
- **8080**: Backend (Go API)
- **5432**: PostgreSQL
- **9092**: Kafka
- **2181**: Zookeeper

Make sure these ports are not in use before starting!
