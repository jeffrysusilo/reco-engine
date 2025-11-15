# Setup script for Recommendation Engine
# This script helps you get started quickly

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Recommendation Engine Setup" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker
Write-Host "Checking prerequisites..." -ForegroundColor Yellow
$dockerInstalled = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerInstalled) {
    Write-Host "ERROR: Docker is not installed. Please install Docker Desktop first." -ForegroundColor Red
    Write-Host "Download from: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

$dockerComposeInstalled = Get-Command docker-compose -ErrorAction SilentlyContinue
if (-not $dockerComposeInstalled) {
    Write-Host "ERROR: Docker Compose is not installed." -ForegroundColor Red
    exit 1
}

Write-Host "✓ Docker is installed" -ForegroundColor Green
Write-Host "✓ Docker Compose is installed" -ForegroundColor Green
Write-Host ""

# Check if Docker is running
try {
    docker ps | Out-Null
    Write-Host "✓ Docker daemon is running" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Docker daemon is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Starting services..." -ForegroundColor Yellow
Write-Host ""

# Start services
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to start services" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Waiting for services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service health
Write-Host ""
Write-Host "Checking service health..." -ForegroundColor Yellow

$services = @(
    @{Name="PostgreSQL"; Port=5432},
    @{Name="Redis"; Port=6379},
    @{Name="Kafka"; Port=9092},
    @{Name="Ingest API"; Port=8080},
    @{Name="Recommendation API"; Port=8081}
)

foreach ($service in $services) {
    $connection = Test-NetConnection -ComputerName localhost -Port $service.Port -WarningAction SilentlyContinue
    if ($connection.TcpTestSucceeded) {
        Write-Host "✓ $($service.Name) is running on port $($service.Port)" -ForegroundColor Green
    } else {
        Write-Host "✗ $($service.Name) is not responding on port $($service.Port)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Testing APIs..." -ForegroundColor Yellow

# Test ingest API
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -TimeoutSec 5
    Write-Host "✓ Ingest API health check passed" -ForegroundColor Green
} catch {
    Write-Host "✗ Ingest API health check failed" -ForegroundColor Red
}

# Test recommendation API
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method Get -TimeoutSec 5
    Write-Host "✓ Recommendation API health check passed" -ForegroundColor Green
} catch {
    Write-Host "✗ Recommendation API health check failed" -ForegroundColor Red
}

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Services are running:" -ForegroundColor Yellow
Write-Host "  • Event Ingest API:       http://localhost:8080" -ForegroundColor White
Write-Host "  • Recommendation API:     http://localhost:8081" -ForegroundColor White
Write-Host "  • Prometheus:             http://localhost:9090" -ForegroundColor White
Write-Host "  • Grafana:                http://localhost:3000 (admin/admin)" -ForegroundColor White
Write-Host ""

Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Generate sample events:" -ForegroundColor White
Write-Host "     .\scripts\generate_events.ps1 -NumEvents 1000" -ForegroundColor Cyan
Write-Host ""
Write-Host "  2. Test the recommendation API:" -ForegroundColor White
Write-Host "     curl http://localhost:8081/recommendations?user_id=1&count=10" -ForegroundColor Cyan
Write-Host ""
Write-Host "  3. View logs:" -ForegroundColor White
Write-Host "     docker-compose logs -f" -ForegroundColor Cyan
Write-Host ""
Write-Host "  4. Stop services:" -ForegroundColor White
Write-Host "     docker-compose down" -ForegroundColor Cyan
Write-Host ""

Write-Host "Documentation:" -ForegroundColor Yellow
Write-Host "  • Quick Start:  docs\QUICKSTART.md" -ForegroundColor White
Write-Host "  • API Docs:     docs\API.md" -ForegroundColor White
Write-Host "  • Architecture: docs\ARCHITECTURE.md" -ForegroundColor White
Write-Host ""
