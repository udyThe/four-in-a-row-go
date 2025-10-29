# Render.com Deployment Checklist

Use this checklist while deploying to Render.com. Check off each step as you complete it.

## Pre-Deployment

- [ ] All code committed to GitHub
- [ ] Code pushed to main branch
- [ ] Render.com account created
- [ ] GitHub connected to Render

## Step 1: Database Deployment

- [ ] Click "New +" → "PostgreSQL"
- [ ] Name: `four-in-a-row-db`
- [ ] Database: `four_in_a_row`
- [ ] Region selected (closest to you)
- [ ] Plan: Free
- [ ] Click "Create Database"
- [ ] Wait for "Available" status (2-3 minutes)
- [ ] Copy **Internal Database URL** from Info tab
- [ ] Save URL somewhere safe (you'll need it for backend)

**Database URL format:**
```
postgres://username:password@host:5432/four_in_a_row
```

## Step 2: Backend Deployment

- [ ] Click "New +" → "Web Service"
- [ ] Select your GitHub repository
- [ ] Click "Connect"
- [ ] Name: `four-in-a-row-backend`
- [ ] Region: Same as database
- [ ] Branch: `main`
- [ ] Root Directory: `backend`
- [ ] Runtime: **Docker**
- [ ] Plan: Free
- [ ] Dockerfile Path: `./Dockerfile`

### Backend Environment Variables:
- [ ] Click "Advanced" → "Add Environment Variable"
- [ ] Add: `PORT` = `8080`
- [ ] Add: `DATABASE_URL` = [paste your database URL]
- [ ] Add: `KAFKA_ENABLED` = `false`

### Backend Health Check:
- [ ] Health Check Path: `/api/health`

### Deploy:
- [ ] Click "Create Web Service"
- [ ] Wait for deployment (5-10 minutes)
- [ ] Check logs for "Server starting on port 8080"
- [ ] Status shows "Live" (green)
- [ ] Copy backend URL (e.g., `https://four-in-a-row-backend.onrender.com`)

**Test backend:**
- [ ] Visit: `https://your-backend-url.onrender.com/api/health`
- [ ] Should return: `{"status":"ok"}`

## Step 3: Frontend Deployment

- [ ] Click "New +" → "Web Service"
- [ ] Select same GitHub repository
- [ ] Click "Connect"
- [ ] Name: `four-in-a-row-frontend`
- [ ] Region: Same as backend
- [ ] Branch: `main`
- [ ] Root Directory: `frontend`
- [ ] Runtime: **Docker**
- [ ] Plan: Free
- [ ] Dockerfile Path: `./Dockerfile`

### Frontend Environment Variables:
- [ ] Click "Advanced" → "Add Environment Variable"
- [ ] Add: `REACT_APP_API_URL` = `https://your-backend-url.onrender.com/api`
- [ ] Add: `REACT_APP_WS_URL` = `wss://your-backend-url.onrender.com/ws`

**⚠️ Important:**
- Replace `your-backend-url` with your actual backend URL
- Use `https://` for API URL
- Use `wss://` for WebSocket URL
- No trailing slashes!

### Deploy:
- [ ] Click "Create Web Service"
- [ ] Wait for deployment (5-10 minutes)
- [ ] Check logs for successful build
- [ ] Status shows "Live" (green)
- [ ] Copy frontend URL (e.g., `https://four-in-a-row-frontend.onrender.com`)

## Step 4: Update Backend CORS

Now you need to allow your frontend domain in the backend.

### In Your Code Editor:
- [ ] Open `backend/internal/api/server.go`
- [ ] Find the `AllowedOrigins` section
- [ ] Add your frontend URL:
  ```go
  AllowedOrigins: []string{
      "http://localhost:3000",
      "https://four-in-a-row-frontend.onrender.com", // Your actual URL
  },
  ```
- [ ] Save the file

### Commit and Push:
```bash
git add backend/internal/api/server.go
git commit -m "Add production CORS origin"
git push
```

### Wait for Redeploy:
- [ ] Go to backend service in Render Dashboard
- [ ] You'll see "Deploying..." status
- [ ] Wait 3-5 minutes
- [ ] Status returns to "Live"

## Step 5: Testing

### Test Single Player (vs Bot):
- [ ] Open frontend URL in browser
- [ ] Click "Play"
- [ ] Enter a username
- [ ] Wait for game to start (matchmaker timeout: 10 seconds)
- [ ] Bot should be opponent
- [ ] Make moves by clicking columns
- [ ] Verify board updates in real-time
- [ ] Play until game ends (win/loss/draw)

### Test Multiplayer:
- [ ] Open frontend URL in first browser
- [ ] Open same URL in incognito/private window
- [ ] Click "Play" in first browser, enter username
- [ ] Click "Play" in second browser, enter different username
- [ ] Both should match (within 10 seconds)
- [ ] Take turns making moves
- [ ] Verify both boards update simultaneously
- [ ] Complete the game

### Test Leaderboard:
- [ ] Click "Leaderboard" tab
- [ ] Should show players with their stats
- [ ] Verify your username appears
- [ ] Check wins/losses/draws counts

### Test Reconnection:
- [ ] Start a game
- [ ] Refresh the browser page
- [ ] Game should reconnect (if within 30 seconds)
- [ ] Continue playing

## Step 6: Monitoring

### Check Service Status:
- [ ] All services show "Live" status in Render Dashboard
- [ ] No error logs in any service

### View Logs:
- [ ] Backend logs show game events
- [ ] No error messages
- [ ] WebSocket connections working

### Performance:
- [ ] First load may take 30-60 seconds (cold start)
- [ ] Subsequent loads should be fast
- [ ] Game moves respond immediately

## Troubleshooting

If something doesn't work, check these:

### Backend Won't Start:
- [ ] Check DATABASE_URL is correct (from database Info tab)
- [ ] Verify PORT is set to 8080
- [ ] Review logs for specific error

### Frontend Shows Connection Error:
- [ ] Verify REACT_APP_API_URL is correct
- [ ] Verify REACT_APP_WS_URL is correct (wss:// not ws://)
- [ ] Check backend CORS settings updated
- [ ] Ensure backend is "Live" before testing frontend

### CORS Errors:
- [ ] Backend server.go has frontend URL in AllowedOrigins
- [ ] Code was committed and pushed
- [ ] Backend redeployed after CORS change

### Database Connection Failed:
- [ ] Database status is "Available"
- [ ] Using **Internal** Database URL (not External)
- [ ] URL has correct format (postgres://...)

### Game Won't Start:
- [ ] Both players connected
- [ ] Check backend logs for matchmaking events
- [ ] Try waiting full 10 seconds for bot matchmaking

## Submission Checklist

- [ ] Frontend URL is live and working
- [ ] Game fully functional (single and multiplayer)
- [ ] Leaderboard displays data
- [ ] No console errors in browser
- [ ] All Render services show "Live" status
- [ ] Screenshots taken (optional)

## URLs for Submission

**Frontend (Main URL):**
```
https://four-in-a-row-frontend.onrender.com
```

**Backend (API):**
```
https://four-in-a-row-backend.onrender.com
```

**Test Health Endpoint:**
```
https://four-in-a-row-backend.onrender.com/api/health
```

## Note About Kafka

Since Kafka can't run on Render's free tier:

**What to tell evaluator:**
> "Kafka analytics integration is implemented (see docker-compose.yml and KAFKA_EXPLAINED.md) but not deployed due to Render's free tier limitations. The game is fully functional without Kafka. To demonstrate Kafka, run `docker-compose up` locally and access the Kafka UI at http://localhost:8090."

**Optional - Include Local Kafka Screenshots:**
- [ ] Run `docker-compose up` locally
- [ ] Access http://localhost:8090
- [ ] Play some games
- [ ] Take screenshots of:
  - Kafka UI dashboard
  - Topics list (game-events)
  - Messages in game-events topic
  - Consumer groups
- [ ] Include in submission documentation

---

## Success! 🎉

If all checkboxes are checked and tests pass, your deployment is complete!

**Next Steps:**
1. Share your frontend URL
2. Optional: Add screenshots to documentation
3. Submit with confidence!

**Estimated Total Time:** 30-45 minutes (including build/deploy waits)

**Cost:** $0.00 (completely free!)
