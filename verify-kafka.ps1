# Kafka Verification Script
# This script demonstrates that Kafka is working in the 4-in-a-Row project

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Kafka Integration Verification" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if containers are running
Write-Host "1. Checking if Kafka containers are running..." -ForegroundColor Yellow
docker-compose ps | Select-String -Pattern "kafka|zookeeper|analytics"
Write-Host ""

# Check Kafka logs for successful connection
Write-Host "2. Checking backend Kafka producer connection..." -ForegroundColor Yellow
docker-compose logs backend | Select-String -Pattern "Kafka" | Select-Object -Last 5
Write-Host ""

# Check analytics service logs
Write-Host "3. Checking analytics consumer logs..." -ForegroundColor Yellow
docker-compose logs analytics | Select-String -Pattern "Starting|Consuming|event" | Select-Object -Last 10
Write-Host ""

# Query the analytics database for game events
Write-Host "4. Querying database for Kafka-processed events..." -ForegroundColor Yellow
Write-Host "   (Checking game_analytics table populated by Kafka consumer)" -ForegroundColor Gray
docker-compose exec -T db psql -U postgres -d four_in_a_row -c "SELECT COUNT(*) as total_events, event_type, COUNT(*) FROM game_analytics GROUP BY event_type;"
Write-Host ""

# Check for recent events
Write-Host "5. Recent game events processed by Kafka:" -ForegroundColor Yellow
docker-compose exec -T db psql -U postgres -d four_in_a_row -c "SELECT event_type, player, timestamp FROM game_analytics ORDER BY timestamp DESC LIMIT 10;"
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verification Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Kafka Web UI Access:" -ForegroundColor Yellow
Write-Host "  URL: http://localhost:8090" -ForegroundColor White
Write-Host "  View Topics, Messages, Consumers in real-time!" -ForegroundColor Gray
Write-Host ""
Write-Host "How Kafka works in this project:" -ForegroundColor Yellow
Write-Host "1. Backend emits game events (start, move, end) to Kafka" -ForegroundColor White
Write-Host "2. Kafka stores these events in 'game-events' topic" -ForegroundColor White
Write-Host "3. Analytics service consumes events from Kafka" -ForegroundColor White
Write-Host "4. Analytics stores processed events in PostgreSQL" -ForegroundColor White
Write-Host "5. Events are used for statistics and game history" -ForegroundColor White
Write-Host ""
