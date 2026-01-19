#!/bin/sh
set -e

echo "Starting services..."

# Start Go backend on fixed internal port 8080
GO_API_PORT=8080
export ADDR=:$GO_API_PORT
echo "Starting Go backend on port $GO_API_PORT..."
/usr/local/bin/server &

# Wait for Go backend to be ready
echo "Waiting for Go backend..."
sleep 3
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -f http://localhost:$GO_API_PORT/healthz >/dev/null 2>&1; then
    echo "Go backend is ready on port $GO_API_PORT"
    break
  fi
  echo "Waiting for Go backend... ($i/10)"
  sleep 1
done

# Start Next.js on PORT (Railway's assigned port)
cd /app/web

# Railway sets PORT automatically - but Go already uses 8080 internally.
# If Railway gives us PORT=8080, move Next.js to 3000 to avoid conflict.
# Default to 3000 when PORT is not set.
if [ -z "$PORT" ] || [ "$PORT" = "8080" ]; then
  export PORT=3000
fi

export NODE_ENV=production
export GO_API_BASE_URL=http://localhost:$GO_API_PORT

echo "Starting Next.js on port $PORT..."
echo "GO_API_BASE_URL=$GO_API_BASE_URL"
echo "NODE_ENV=$NODE_ENV"

# Verify Next.js files exist
if [ ! -d "/app/web/.next" ]; then
  echo "ERROR: Next.js build directory not found!"
  exit 1
fi

# Start Next.js (this will block and keep container alive)
exec npm start
