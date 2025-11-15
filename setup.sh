#!/bin/bash

# Setup script for Recommendation Engine
# This script helps you get started quickly

echo "================================"
echo "Recommendation Engine Setup"
echo "================================"
echo ""

# Check Docker
echo "Checking prerequisites..."
if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed. Please install Docker first."
    echo "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "ERROR: Docker Compose is not installed."
    exit 1
fi

echo "✓ Docker is installed"
echo "✓ Docker Compose is installed"
echo ""

# Check if Docker is running
if ! docker ps &> /dev/null; then
    echo "ERROR: Docker daemon is not running. Please start Docker."
    exit 1
fi

echo "✓ Docker daemon is running"
echo ""

echo "Starting services..."
echo ""

# Start services
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to start services"
    exit 1
fi

echo ""
echo "Waiting for services to be ready..."
sleep 10

echo ""
echo "Checking service health..."

# Function to check if port is open
check_port() {
    nc -z localhost $1 &> /dev/null
    return $?
}

services=(
    "PostgreSQL:5432"
    "Redis:6379"
    "Kafka:9092"
    "Ingest API:8080"
    "Recommendation API:8081"
)

for service in "${services[@]}"; do
    name="${service%%:*}"
    port="${service##*:}"
    
    if check_port $port; then
        echo "✓ $name is running on port $port"
    else
        echo "✗ $name is not responding on port $port"
    fi
done

echo ""
echo "Testing APIs..."

# Test ingest API
if curl -s -f http://localhost:8080/health > /dev/null; then
    echo "✓ Ingest API health check passed"
else
    echo "✗ Ingest API health check failed"
fi

# Test recommendation API
if curl -s -f http://localhost:8081/health > /dev/null; then
    echo "✓ Recommendation API health check passed"
else
    echo "✗ Recommendation API health check failed"
fi

echo ""
echo "================================"
echo "Setup Complete!"
echo "================================"
echo ""

echo "Services are running:"
echo "  • Event Ingest API:       http://localhost:8080"
echo "  • Recommendation API:     http://localhost:8081"
echo "  • Prometheus:             http://localhost:9090"
echo "  • Grafana:                http://localhost:3000 (admin/admin)"
echo ""

echo "Next steps:"
echo "  1. Generate sample events:"
echo "     ./scripts/generate_events.sh http://localhost:8080 1000"
echo ""
echo "  2. Test the recommendation API:"
echo "     curl http://localhost:8081/recommendations?user_id=1&count=10"
echo ""
echo "  3. View logs:"
echo "     docker-compose logs -f"
echo ""
echo "  4. Stop services:"
echo "     docker-compose down"
echo ""

echo "Documentation:"
echo "  • Quick Start:  docs/QUICKSTART.md"
echo "  • API Docs:     docs/API.md"
echo "  • Architecture: docs/ARCHITECTURE.md"
echo ""
