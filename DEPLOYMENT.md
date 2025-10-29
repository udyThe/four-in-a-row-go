# Deployment Guide

This guide covers various deployment options for the 4 in a Row application.

## Table of Contents
- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Cloud Deployment](#cloud-deployment)
- [Production Checklist](#production-checklist)

## Local Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Kafka (optional)

### Backend Setup

```powershell
# Navigate to backend
cd backend

# Install dependencies
go mod download

# Set environment variables
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/four_in_a_row?sslmode=disable"
$env:PORT="8080"
$env:KAFKA_BROKER="localhost:9092"

# Run the server
go run main.go
```

### Frontend Setup

```powershell
# Navigate to frontend
cd frontend

# Install dependencies
npm install

# Start development server
npm start
```

### Database Setup

```sql
-- Create database
CREATE DATABASE four_in_a_row;

-- Tables will be created automatically by the backend
```

## Docker Deployment

### Single-Host Deployment

```powershell
# Clone repository
git clone <repo-url>
cd UdayAssignment

# Start all services
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Custom Configuration

Create `.env` files:

**backend/.env**
```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/four_in_a_row?sslmode=disable
KAFKA_BROKER=kafka:29092
```

**frontend/.env**
```env
REACT_APP_API_URL=http://your-domain.com/api
REACT_APP_WS_URL=ws://your-domain.com/ws
```

## Cloud Deployment

### AWS Deployment

#### Option 1: ECS (Elastic Container Service)

1. **Create ECR Repositories**
```bash
aws ecr create-repository --repository-name 4-in-a-row-backend
aws ecr create-repository --repository-name 4-in-a-row-frontend
aws ecr create-repository --repository-name 4-in-a-row-analytics
```

2. **Build and Push Images**
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build images
docker build -t 4-in-a-row-backend ./backend
docker build -t 4-in-a-row-frontend ./frontend
docker build -t 4-in-a-row-analytics ./analytics

# Tag images
docker tag 4-in-a-row-backend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-backend:latest
docker tag 4-in-a-row-frontend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-frontend:latest
docker tag 4-in-a-row-analytics:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-analytics:latest

# Push images
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-backend:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-frontend:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row-analytics:latest
```

3. **Setup RDS PostgreSQL**
```bash
aws rds create-db-instance \
  --db-instance-identifier 4-in-a-row-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username postgres \
  --master-user-password <password> \
  --allocated-storage 20
```

4. **Setup MSK (Managed Kafka)**
```bash
aws kafka create-cluster \
  --cluster-name 4-in-a-row-kafka \
  --broker-node-group-info file://broker-info.json \
  --kafka-version 2.8.1
```

5. **Create ECS Cluster and Services**
- Use AWS Console or CloudFormation
- Configure task definitions
- Setup Application Load Balancer for WebSocket support

#### Option 2: EKS (Kubernetes)

1. **Create EKS Cluster**
```bash
eksctl create cluster \
  --name 4-in-a-row \
  --region us-east-1 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3
```

2. **Deploy Services**
```bash
kubectl apply -f k8s/
```

### Azure Deployment

#### Option 1: Azure Container Instances

```bash
# Create resource group
az group create --name 4-in-a-row-rg --location eastus

# Create container registry
az acr create --resource-group 4-in-a-row-rg --name 4inarowacr --sku Basic

# Build and push images
az acr build --registry 4inarowacr --image 4-in-a-row-backend:latest ./backend
az acr build --registry 4inarowacr --image 4-in-a-row-frontend:latest ./frontend
az acr build --registry 4inarowacr --image 4-in-a-row-analytics:latest ./analytics

# Create Azure Database for PostgreSQL
az postgres server create \
  --resource-group 4-in-a-row-rg \
  --name 4-in-a-row-db \
  --location eastus \
  --admin-user postgres \
  --admin-password <password> \
  --sku-name B_Gen5_1

# Deploy containers
az container create \
  --resource-group 4-in-a-row-rg \
  --name 4-in-a-row-backend \
  --image 4inarowacr.azurecr.io/4-in-a-row-backend:latest \
  --cpu 1 --memory 1 \
  --ports 8080 \
  --environment-variables DATABASE_URL=<connection-string>
```

#### Option 2: Azure Kubernetes Service (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group 4-in-a-row-rg \
  --name 4-in-a-row-aks \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group 4-in-a-row-rg --name 4-in-a-row-aks

# Deploy
kubectl apply -f k8s/
```

### Google Cloud Platform (GCP)

#### Option 1: Cloud Run

```bash
# Setup
gcloud auth login
gcloud config set project <project-id>

# Build images with Cloud Build
gcloud builds submit --tag gcr.io/<project-id>/4-in-a-row-backend ./backend
gcloud builds submit --tag gcr.io/<project-id>/4-in-a-row-frontend ./frontend
gcloud builds submit --tag gcr.io/<project-id>/4-in-a-row-analytics ./analytics

# Create Cloud SQL instance
gcloud sql instances create 4-in-a-row-db \
  --database-version=POSTGRES_14 \
  --tier=db-f1-micro \
  --region=us-central1

# Deploy to Cloud Run
gcloud run deploy 4-in-a-row-backend \
  --image gcr.io/<project-id>/4-in-a-row-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars DATABASE_URL=<connection-string>
```

#### Option 2: GKE (Google Kubernetes Engine)

```bash
# Create GKE cluster
gcloud container clusters create 4-in-a-row \
  --num-nodes=3 \
  --zone=us-central1-a

# Get credentials
gcloud container clusters get-credentials 4-in-a-row --zone=us-central1-a

# Deploy
kubectl apply -f k8s/
```

### Heroku Deployment

#### Backend

```bash
# Navigate to backend
cd backend

# Login to Heroku
heroku login

# Create app
heroku create 4-in-a-row-backend

# Add PostgreSQL addon
heroku addons:create heroku-postgresql:hobby-dev

# Add Kafka addon (optional)
heroku addons:create heroku-kafka:basic-0

# Create Procfile
echo "web: ./main" > Procfile

# Deploy
git init
git add .
git commit -m "Initial commit"
heroku git:remote -a 4-in-a-row-backend
git push heroku master
```

#### Frontend

```bash
# Navigate to frontend
cd frontend

# Create app
heroku create 4-in-a-row-frontend

# Add buildpack
heroku buildpacks:add heroku/nodejs

# Set environment variables
heroku config:set REACT_APP_API_URL=https://4-in-a-row-backend.herokuapp.com/api
heroku config:set REACT_APP_WS_URL=wss://4-in-a-row-backend.herokuapp.com/ws

# Create Procfile
echo "web: npm start" > Procfile

# Deploy
git init
git add .
git commit -m "Initial commit"
heroku git:remote -a 4-in-a-row-frontend
git push heroku master
```

### DigitalOcean

#### App Platform

1. **Connect Repository**
   - Login to DigitalOcean
   - Go to App Platform
   - Connect GitHub repository

2. **Configure Services**
   - Backend: Go app (Dockerfile)
   - Frontend: Static site or Node.js
   - Database: Managed PostgreSQL

3. **Set Environment Variables**
   - Add DATABASE_URL
   - Add KAFKA_BROKER (if using)

4. **Deploy**
   - Click "Deploy"
   - Wait for build and deployment

## Production Checklist

### Security
- [ ] Change default database credentials
- [ ] Use environment variables for secrets
- [ ] Enable HTTPS/WSS
- [ ] Configure CORS properly
- [ ] Add rate limiting
- [ ] Enable authentication (if needed)
- [ ] Setup firewall rules
- [ ] Use secrets manager (AWS Secrets Manager, Azure Key Vault)

### Performance
- [ ] Enable database connection pooling
- [ ] Setup CDN for frontend assets
- [ ] Configure caching (Redis)
- [ ] Enable gzip compression
- [ ] Optimize Docker images (multi-stage builds)
- [ ] Setup horizontal scaling
- [ ] Configure load balancer

### Monitoring
- [ ] Setup application monitoring (New Relic, Datadog)
- [ ] Configure log aggregation (ELK, CloudWatch)
- [ ] Setup alerts for errors
- [ ] Monitor WebSocket connections
- [ ] Track database performance
- [ ] Monitor Kafka lag

### Backup & Recovery
- [ ] Enable automated database backups
- [ ] Test restore procedures
- [ ] Document recovery process
- [ ] Setup disaster recovery plan

### CI/CD
- [ ] Setup GitHub Actions or similar
- [ ] Automated testing on PR
- [ ] Automated deployment on merge
- [ ] Blue-green deployment
- [ ] Rollback strategy

### Documentation
- [ ] API documentation
- [ ] Deployment runbook
- [ ] Troubleshooting guide
- [ ] Architecture diagrams
- [ ] Change log

## Environment Variables Reference

### Backend
```env
PORT=8080
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
KAFKA_BROKER=kafka-host:9092
LOG_LEVEL=info
CORS_ORIGINS=https://yourdomain.com
```

### Frontend
```env
REACT_APP_API_URL=https://api.yourdomain.com/api
REACT_APP_WS_URL=wss://api.yourdomain.com/ws
```

### Analytics
```env
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
KAFKA_BROKER=kafka-host:9092
KAFKA_GROUP_ID=analytics-consumer
```

## Troubleshooting

### WebSocket Connection Issues
- Ensure load balancer supports WebSocket (sticky sessions)
- Check CORS configuration
- Verify SSL/TLS certificates
- Check firewall rules

### Database Connection Issues
- Verify connection string
- Check network security groups
- Ensure database is accessible from application
- Verify credentials

### Kafka Issues
- Ensure topics are created
- Check broker connectivity
- Verify consumer group configuration
- Monitor consumer lag

## Support

For issues and questions:
- GitHub Issues: [Create an issue]
- Documentation: See README.md
- Architecture: See ARCHITECTURE.md
