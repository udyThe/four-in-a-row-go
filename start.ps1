# Quick Start Script for Windows

Write-Host "4 in a Row - Quick Start" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is installed
Write-Host "Checking Docker installation..." -ForegroundColor Yellow
docker --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker is not installed. Please install Docker Desktop first." -ForegroundColor Red
    exit 1
}

Write-Host "SUCCESS: Docker is installed" -ForegroundColor Green
Write-Host ""

# Check if Docker Compose is installed
Write-Host "Checking Docker Compose installation..." -ForegroundColor Yellow
docker-compose --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker Compose is not installed." -ForegroundColor Red
    exit 1
}

Write-Host "SUCCESS: Docker Compose is installed" -ForegroundColor Green
Write-Host ""

# Stop any running containers
Write-Host "Stopping any existing containers..." -ForegroundColor Yellow
docker-compose down

# Build and start services
Write-Host "Building and starting all services..." -ForegroundColor Yellow
Write-Host "This may take a few minutes on first run..." -ForegroundColor Yellow
docker-compose up --build -d

# Wait for services to be ready
Write-Host ""
Write-Host "Waiting for services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service status
Write-Host ""
Write-Host "Service Status:" -ForegroundColor Cyan
docker-compose ps

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Application is ready!" -ForegroundColor Green
Write-Host ""
Write-Host "Access the application at:" -ForegroundColor Cyan
Write-Host "  Frontend:  http://localhost:3000" -ForegroundColor White
Write-Host "  Backend:   http://localhost:8080" -ForegroundColor White
Write-Host "  API Docs:  http://localhost:8080/api/health" -ForegroundColor White
Write-Host ""
# Write-Host "To view logs:" -ForegroundColor Yellow
# Write-Host "  docker-compose logs -f" -ForegroundColor White
# Write-Host ""
# Write-Host "To stop the application:" -ForegroundColor Yellow
# Write-Host "  docker-compose down" -ForegroundColor White
# Write-Host ""
