#!/bin/bash

echo "🚀 Starting Centrifugo Messaging Application"
echo "============================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start Centrifugo and backend with Docker Compose
echo "📦 Starting Centrifugo and Backend services..."
cd docker
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 5

# Check if Centrifugo is ready
echo "🔍 Checking Centrifugo health..."
for i in {1..10}; do
    if curl -s http://localhost:8000/health > /dev/null; then
        echo "✅ Centrifugo is ready!"
        break
    fi
    echo "   Attempt $i/10 - waiting for Centrifugo..."
    sleep 2
done

# Check if Backend is ready
echo "🔍 Checking Backend health..."
for i in {1..10}; do
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "✅ Backend is ready!"
        break
    fi
    echo "   Attempt $i/10 - waiting for Backend..."
    sleep 2
done

echo ""
echo "🎉 Services are running!"
echo "📊 Centrifugo Admin: http://localhost:8000 (password: password)"
echo "🔧 Backend API: http://localhost:8080"
echo ""
echo "Next steps:"
echo "1. cd frontend"
echo "2. npm run dev"
echo "3. Open http://localhost:5173 in your browser"
echo ""
echo "To stop services: docker-compose down"