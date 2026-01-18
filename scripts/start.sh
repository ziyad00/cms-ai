#!/bin/sh
set -e

# Start Go backend on port 8080 in background
GO_API_PORT=8080
export PORT=$GO_API_PORT
/usr/local/bin/server &

# Wait for Go backend to be ready
sleep 2
until curl -f http://localhost:$GO_API_PORT/healthz >/dev/null 2>&1; do
  echo "Waiting for Go backend to be ready..."
  sleep 1
done

echo "Go backend is ready on port $GO_API_PORT"

# Start Next.js on PORT (Railway's assigned port or 3000)
cd /app/web
export PORT=${PORT:-3000}
export GO_API_BASE_URL=http://localhost:$GO_API_PORT
exec npm start
