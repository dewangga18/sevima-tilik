#!/bin/bash
# ============================================
# Production Build & Deploy Script
# ============================================
# This script builds and starts Tilik in production mode

set -e

echo "🏭 Starting Tilik PRODUCTION build..."

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo "❌ ERROR: .env.production not found!"
    echo "📝 Please create .env.production from .env.production.example"
    echo "   cp .env.production.example .env.production"
    echo "   # Then edit .env.production with your actual values"
    exit 1
fi

# Load production env
source .env.production

# Validate required variables
if [ -z "$DB_PASSWORD" ] || [ "$DB_PASSWORD" = "CHANGE_ME_STRONG_PASSWORD_HERE" ]; then
    echo "❌ ERROR: DB_PASSWORD not set or still using default value"
    echo "   Please set a strong password in .env.production"
    exit 1
fi

if [ -z "$VITE_API_URL" ]; then
    echo "❌ ERROR: VITE_API_URL not set in .env.production"
    exit 1
fi

if [ -z "$ALLOWED_ORIGIN" ]; then
    echo "❌ ERROR: ALLOWED_ORIGIN not set in .env.production"
    exit 1
fi

echo "✅ Environment validation passed"
echo ""
echo "📦 Building services..."
echo "   - API: production binary"
echo "   - Web: static build with API URL: $VITE_API_URL"
echo ""

# Build with production config
docker compose -f docker-compose.prod.yml --env-file .env.production build \
    --build-arg VITE_API_URL="$VITE_API_URL"

echo ""
echo "✅ Build complete!"
echo ""
echo "🚀 Starting services..."

# Start services
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

echo ""
echo "✅ Tilik is now running in PRODUCTION mode"
echo ""
echo "📊 Service status:"
docker compose -f docker-compose.prod.yml ps
echo ""
echo "⚠️  IMPORTANT: Set up HTTPS reverse proxy (nginx/Caddy/Traefik) to:"
echo "   - Expose web service on port 80/443"
echo "   - Expose api service on port 80/443"
echo "   - Enable SSL certificates (Let's Encrypt recommended)"
echo ""
echo "📝 View logs:"
echo "   docker compose -f docker-compose.prod.yml logs -f"
echo ""
echo "🛑 Stop services:"
echo "   docker compose -f docker-compose.prod.yml down"
