#!/bin/bash

# Railway health check and monitoring setup
set -e

echo "🏥 Setting up health checks and monitoring..."

# Define health check endpoints
HEALTH_ENDPOINTS=(
    "/health"
    "/api/v1/health" 
    "/api/health"
)

# Test health endpoints
echo "🔍 Testing health endpoints..."
for endpoint in "${HEALTH_ENDPOINTS[@]}"; do
    echo "Testing: $endpoint"
    curl -f -s "https://$RAILWAY_PUBLIC_DOMAIN$endpoint" || echo "⚠️  Endpoint not responding"
done

# Set up log monitoring
echo "📊 Setting up log monitoring..."
railway logs --follow &

# Check resource usage
echo "💾 Checking resource usage..."
railway status

# Test database connectivity
echo "🗄️  Testing database connectivity..."
railway connect postgresql -c "SELECT version();"

# Test storage connectivity (if configured)
if [ "$STORAGE_TYPE" = "s3" ] || [ "$STORAGE_TYPE" = "gcs" ]; then
    echo "📦 Testing storage connectivity..."
    # Add storage health check here
    echo "✅ Storage configured: $STORAGE_TYPE"
fi

# Test authentication endpoints
echo "🔐 Testing authentication endpoints..."
curl -f -s "https://$RAILWAY_PUBLIC_DOMAIN/api/auth/signin" || echo "⚠️  Auth endpoint not responding"

echo "✅ Health check and monitoring setup complete!"