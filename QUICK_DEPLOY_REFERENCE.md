# Quick Deployment Reference Card

## 🚀 Deploy in 5 Steps

### 1️⃣ Push to GitHub
```bash
git add .
git commit -m "Ready for deployment"
git push
```

### 2️⃣ Create Database
- Render Dashboard → New + → PostgreSQL
- Name: `four-in-a-row-db`
- Plan: **Free**
- Copy **Internal Database URL**

### 3️⃣ Deploy Backend
- New + → Web Service → Connect GitHub
- Root Directory: `backend`
- Runtime: **Docker**
- Environment Variables:
  ```
  PORT = 8080
  DATABASE_URL = [paste from step 2]
  KAFKA_ENABLED = false
  ```
- Health Check: `/api/health`
- Copy backend URL

### 4️⃣ Deploy Frontend
- New + → Web Service → Connect GitHub
- Root Directory: `frontend`
- Runtime: **Docker**
- Environment Variables:
  ```
  REACT_APP_API_URL = https://[your-backend].onrender.com/api
  REACT_APP_WS_URL = wss://[your-backend].onrender.com/ws
  ```

### 5️⃣ Update CORS & Test
- Edit `backend/internal/api/server.go`
- Add frontend URL to `AllowedOrigins`
- Commit and push
- Wait for redeploy
- Test at your frontend URL!

---

## ⏱️ Time Estimate
- Setup: 5 minutes
- Database: 3 minutes
- Backend: 8 minutes
- Frontend: 8 minutes
- CORS + Test: 6 minutes
- **Total: ~30 minutes**

---

## 💰 Cost
**$0.00** - Completely free!

---

## 📖 Full Guides
- **Step-by-Step:** `DEPLOYMENT_CHECKLIST.md`
- **Detailed Guide:** `RENDER_DEPLOYMENT_GUIDE.md`
- **Summary:** `DEPLOYMENT_READY.md`

---

## ✅ Test Checklist
- [ ] Click "Play" → plays vs bot
- [ ] Two browsers → multiplayer works
- [ ] Leaderboard shows stats
- [ ] No errors in browser console

---

## 🆘 Quick Troubleshooting

**Backend won't start?**
→ Check DATABASE_URL is correct

**Frontend connection error?**
→ Check REACT_APP_API_URL and REACT_APP_WS_URL

**CORS error?**
→ Update server.go with frontend URL, commit, push

**Database error?**
→ Use Internal DB URL, not External

---

## 📱 URLs to Share

After deployment, share your **frontend URL**:
```
https://four-in-a-row-frontend.onrender.com
```

Test backend health:
```
https://four-in-a-row-backend.onrender.com/api/health
```

---

## 📝 Note About Kafka

Kafka can't run on Render free tier. The game works perfectly without it!

**To demonstrate Kafka:**
```bash
docker-compose up
```
Then visit: http://localhost:8090

---

## 🎯 Success = All Green "Live" Status!

Good luck! 🚀
