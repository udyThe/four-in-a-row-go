# Render.com Deployment Guide (Dashboard Method)

This guide provides step-by-step instructions for deploying the 4-in-a-Row game to Render.com using the Dashboard UI.

## Important Note About Kafka

**⚠️ Render.com Free Tier Limitation**: Render's free tier does NOT support running Kafka and Zookeeper as they require persistent disk storage and specific networking configurations that aren't available in the free tier.

For this project, you have two deployment options:

### Option A: Deploy Without Kafka (Recommended for Free Tier)
Deploy the game functionality (backend + frontend + database) which works perfectly. Kafka analytics will be disabled but the game is fully functional.

### Option B: Deploy Full Stack with Kafka (Requires Paid Plan or Alternative)
Use a paid Render plan ($7/month) or deploy Kafka to a separate service like Confluent Cloud (free tier available) or Upstash Kafka.

**This guide covers Option A (Game Only) which is free and fully functional.**

---

## Prerequisites

1. **GitHub Account**: Your code must be in a GitHub repository
2. **Render.com Account**: Sign up at [render.com](https://render.com) (free)
3. **Code Ready**: Push all your code to GitHub main branch

## Architecture Overview

We'll deploy 3 services:
- **PostgreSQL Database** (managed by Render)
- **Backend Service** (Go + WebSocket)
- **Frontend Service** (React + Nginx)

---

## Step-by-Step Deployment

### Step 1: Create Render Account

1. Go to [render.com](https://render.com)
2. Click **"Get Started"** or **"Sign Up"**
3. Choose **"Sign up with GitHub"** (recommended)
4. Authorize Render to access your repositories

---

### Step 2: Create PostgreSQL Database

1. From Render Dashboard, click **"New +"** button (top right)
2. Select **"PostgreSQL"**
3. Fill in the details:
   - **Name**: `four-in-a-row-db`
   - **Database**: `four_in_a_row`
   - **User**: `postgres` (default)
   - **Region**: Select closest to you
   - **PostgreSQL Version**: 16 (default)
   - **Plan**: **Free**
4. Click **"Create Database"**
5. Wait 2-3 minutes for provisioning
6. Once ready, go to the **"Info"** tab
7. **Copy the "Internal Database URL"** - you'll need this for the backend

---

### Step 3: Deploy Backend Service

1. From Dashboard, click **"New +"** → **"Web Service"**
2. Connect your GitHub repository:
   - Find your repository in the list
   - Click **"Connect"**
3. Fill in the configuration:
   - **Name**: `four-in-a-row-backend`
   - **Region**: Same as database
   - **Branch**: `main`
   - **Root Directory**: `backend`
   - **Runtime**: **Docker**
   - **Plan**: **Free**
   - **Dockerfile Path**: `./Dockerfile`
4. Add Environment Variables:
   - Click **"Advanced"** → **"Add Environment Variable"**
   - Add these variables:
     
     ```
     PORT = 8080
     DATABASE_URL = [paste Internal Database URL from Step 2]
     KAFKA_ENABLED = false
     ```
   
5. Under **"Health Check Path"**, enter: `/api/health`
6. Click **"Create Web Service"**
7. Wait 5-10 minutes for build and deploy
8. Once deployed, copy the service URL (e.g., `https://four-in-a-row-backend.onrender.com`)

---

### Step 4: Deploy Frontend Service

1. From Dashboard, click **"New +"** → **"Web Service"**
2. Connect the same GitHub repository
3. Fill in the configuration:
   - **Name**: `four-in-a-row-frontend`
   - **Region**: Same as backend
   - **Branch**: `main`
   - **Root Directory**: `frontend`
   - **Runtime**: **Docker**
   - **Plan**: **Free**
   - **Dockerfile Path**: `./Dockerfile`
4. Add Environment Variables:
   - Click **"Advanced"** → **"Add Environment Variable"**
   - Add these variables (replace with your actual backend URL):
     
     ```
     REACT_APP_API_URL = https://four-in-a-row-backend.onrender.com/api
     REACT_APP_WS_URL = wss://four-in-a-row-backend.onrender.com/ws
     ```
   
5. Click **"Create Web Service"**
6. Wait 5-10 minutes for build and deploy
7. Once deployed, copy the frontend URL (e.g., `https://four-in-a-row-frontend.onrender.com`)

---

### Step 5: Update Backend CORS Settings

The backend needs to allow requests from your frontend domain.

1. Open `backend/internal/api/server.go` in your code editor
2. Find the CORS configuration section
3. Add your frontend URL to allowed origins:
   
   ```go
   AllowedOrigins: []string{
       "http://localhost:3000",
       "https://four-in-a-row-frontend.onrender.com", // Add your actual frontend URL
   },
   ```

4. Commit and push to GitHub:
   ```bash
   git add backend/internal/api/server.go
   git commit -m "Update CORS for production frontend"
   git push
   ```

5. Render will automatically redeploy the backend (takes 3-5 minutes)

---

### Step 6: Test Your Deployment

1. Open your frontend URL in a browser
2. Click **"Play"** to start a game
3. Try making moves - you should be playing against the bot
4. Open a second browser/incognito window with the same URL
5. Click **"Play"** in both - they should match as opponents
6. Test the leaderboard to ensure database connection works

---

## Troubleshooting

### Backend Service Fails to Start

**Check Logs:**
1. Go to your backend service in Render Dashboard
2. Click **"Logs"** tab
3. Look for error messages

**Common Issues:**
- **Database connection failed**: Check that `DATABASE_URL` is correctly copied from the database Info tab
- **Port binding error**: Ensure `PORT` environment variable is set to `8080`

### Frontend Shows Connection Errors

**Check Environment Variables:**
1. Ensure `REACT_APP_API_URL` and `REACT_APP_WS_URL` match your actual backend URL
2. URLs must start with `https://` and `wss://` respectively
3. No trailing slashes

**CORS Errors:**
- Make sure you updated the backend CORS settings (Step 5)
- The backend must be redeployed after CORS changes

### Database Connection Issues

1. Verify the database is in "Available" status
2. Use the **Internal Database URL** (not External)
3. Check the backend logs for specific database errors

### Free Tier Limitations

Render free tier services:
- Spin down after 15 minutes of inactivity
- Take 30-60 seconds to spin back up on first request
- Have 750 hours/month limit (enough for one service to run continuously)

**Solutions:**
- Use a service like UptimeRobot to ping your services every 10 minutes
- Upgrade to paid plan ($7/month per service) for always-on

---

## URLs for Evaluator Submission

After successful deployment, you'll have these URLs:

1. **Frontend (Game UI)**: `https://four-in-a-row-frontend.onrender.com`
   - This is the main URL to share
   - Users can play the game here

2. **Backend API**: `https://four-in-a-row-backend.onrender.com`
   - Used internally by frontend
   - Can test with: `https://four-in-a-row-backend.onrender.com/api/health`

3. **Database**: Internal only (not publicly accessible)

**Note**: Since Kafka is not deployed in the free tier, you won't have a Kafka UI URL. The game works perfectly without Kafka - it just means analytics events aren't being processed.

---

## Alternative: Deploy Kafka Separately (Optional)

If you want to include Kafka functionality for the evaluator:

### Option 1: Upstash Kafka (Free Tier)

1. Sign up at [upstash.com](https://upstash.com)
2. Create a free Kafka cluster
3. Get the connection URL and credentials
4. Add to backend environment variables:
   ```
   KAFKA_ENABLED = true
   KAFKA_BROKER = [Upstash broker URL]
   KAFKA_USERNAME = [Upstash username]
   KAFKA_PASSWORD = [Upstash password]
   ```

### Option 2: Confluent Cloud (Free Tier)

1. Sign up at [confluent.cloud](https://confluent.cloud)
2. Create a Basic cluster (free for 30 days)
3. Create a topic called `game_events`
4. Get API key and secret
5. Add to backend environment variables

### Option 3: Railway.app (Alternative to Render)

Railway has better support for Docker Compose and can run Kafka:
1. Sign up at [railway.app](https://railway.app)
2. Deploy from GitHub
3. Railway will read your `docker-compose.yml`
4. All services including Kafka will deploy
5. More expensive but supports full stack

---

## Monitoring Your Services

### Check Service Health

1. Go to Render Dashboard
2. All services should show **"Live"** status in green
3. Click on each service to see:
   - Recent logs
   - Metrics (CPU, Memory)
   - Deploy history

### View Logs

Real-time logs are crucial for debugging:
1. Click on a service
2. Click **"Logs"** tab
3. Use **"Live"** toggle to see real-time logs
4. Filter by date/time if needed

### Set Up Notifications

1. Go to Account Settings → Notifications
2. Enable email notifications for:
   - Deploy failures
   - Service crashes
   - High resource usage

---

## Updating Your Deployment

After making code changes:

1. Commit and push to GitHub:
   ```bash
   git add .
   git commit -m "Your changes"
   git push
   ```

2. Render auto-deploys from GitHub:
   - Go to Render Dashboard
   - You'll see "Deploying..." status
   - Wait 3-10 minutes for build and deploy
   - Check logs if deployment fails

### Manual Redeploy

If auto-deploy doesn't trigger:
1. Go to the service in Render Dashboard
2. Click **"Manual Deploy"** → **"Deploy latest commit"**

---

## Cost Breakdown

**Free Tier (Recommended):**
- Database: Free PostgreSQL (1GB storage, shared CPU)
- Backend: Free (512MB RAM, shared CPU, spins down after 15min)
- Frontend: Free (512MB RAM, shared CPU, spins down after 15min)
- **Total: $0/month**

**Paid Tier (Always On):**
- Database: $7/month (shared CPU, 1GB storage)
- Backend: $7/month (always on, 512MB RAM)
- Frontend: $7/month (always on, 512MB RAM)
- **Total: $21/month**

---

## Security Best Practices

1. **Never commit secrets**: Use Render's environment variables for all sensitive data
2. **Use Internal URLs**: Database should use internal URL (faster, more secure)
3. **Enable HTTPS**: Render provides free SSL certificates automatically
4. **Restrict CORS**: Only allow your frontend domain in backend CORS settings

---

## Next Steps After Deployment

1. **Test thoroughly**: Try all game features
2. **Check leaderboard**: Ensure database writes work
3. **Test multiplayer**: Open two browsers and verify matchmaking
4. **Monitor logs**: Watch for any errors in production
5. **Share URL**: Give your frontend URL to the evaluator

---

## Getting Help

If you encounter issues:

1. **Check Render Docs**: [render.com/docs](https://render.com/docs)
2. **Render Community**: [community.render.com](https://community.render.com)
3. **GitHub Issues**: Check if others had similar problems
4. **Service Logs**: Always check logs first - they show exact errors

---

## Summary Checklist

- [ ] Database created and running
- [ ] Backend deployed with correct DATABASE_URL
- [ ] Frontend deployed with correct API and WebSocket URLs
- [ ] CORS settings updated in backend
- [ ] All services show "Live" status
- [ ] Game tested in browser (single player vs bot)
- [ ] Multiplayer tested (two browsers)
- [ ] Leaderboard working
- [ ] Frontend URL shared with evaluator

**Your deployment is complete!** 🎉
