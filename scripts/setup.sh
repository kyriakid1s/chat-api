#!/bin/bash

# Go Chat API Setup Script

echo "🚀 Setting up Go Chat API with PostgreSQL..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create .env file from template if it doesn't exist
if [ ! -f .env ]; then
    echo "📄 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created. You can modify it if needed."
fi

# Start PostgreSQL database
echo "🐘 Starting PostgreSQL database..."
docker-compose up -d postgres

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Check if database is ready
until docker-compose exec postgres pg_isready -U postgres &> /dev/null; do
    echo "⏳ Still waiting for database..."
    sleep 2
done

echo "✅ Database is ready!"

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Build the application
echo "🔨 Building the application..."
go build -o bin/chatapi cmd/main.go

echo "🎉 Setup complete!"
echo ""
echo "To start the application:"
echo "  1. Make sure the database is running: docker-compose up -d postgres"
echo "  2. Run the application: go run cmd/main.go"
echo "  3. Or use the built binary: ./bin/chatapi"
echo ""
echo "The API will be available at: http://localhost:8080"
echo "PostgreSQL will be available at: localhost:5432"
echo ""
echo "To stop the database: docker-compose down"
