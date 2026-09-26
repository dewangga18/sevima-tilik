#!/bin/bash
# ============================================
# Development Quick Start Script
# ============================================
# This script sets up and runs Tilik in development mode

set -e

echo "🚀 Starting Tilik in DEVELOPMENT mode..."

# Check if .env exists, if not copy from example
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ .env created. You can modify it if needed."
fi

# Build and start services
echo "🔨 Building and starting services..."
docker compose up --build

# Note: Services will be available at:
# - Web: http://localhost:5173
# - API: http://localhost:8080
# - DB: localhost:5432
