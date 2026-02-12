#!/bin/sh
set -e

echo "Starting CMS-AI services..."

# Start Go backend on internal port 8081 (avoid conflict with Railway's PORT)
GO_API_PORT=8081
export ADDR=:$GO_API_PORT
# Start Go backend
echo "Starting Go backend on port $GO_API_PORT..."
echo "🚀 BINARY INFO: $(ls -l /usr/local/bin/server)"
echo "🚀 BINARY MD5: $(md5sum /usr/local/bin/server || md5 /usr/local/bin/server)"

# Unset PORT temporarily so Go server uses ADDR instead
env -u PORT /usr/local/bin/server &
GO_PID=$!

# Wait for Go backend to be ready
echo "Waiting for Go backend (PID: $GO_PID)..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  if kill -0 $GO_PID 2>/dev/null; then
    if curl -f http://localhost:$GO_API_PORT/healthz >/dev/null 2>&1; then
      echo "✅ Go backend is ready on port $GO_API_PORT"
      break
    fi
  else
    echo "❌ Go backend CRASHED immediately! Check logs above."
    wait $GO_PID || true
    # We continue so Next.js starts and we can see the logs
    break
  fi
  echo "Waiting... ($i/10)"
  sleep 1
done

# Start Next.js on PORT (Railway's assigned port)
cd /app/web

# Railway sets PORT automatically for external routing.
# Next.js should use Railway's PORT, Go uses internal 8080
# Default to 3000 when PORT is not set (local dev)
if [ -z "$PORT" ]; then
  export PORT=3000
fi

export NODE_ENV=production
export GO_API_BASE_URL=http://127.0.0.1:$GO_API_PORT

echo "Starting Next.js on port $PORT..."
echo "GO_API_BASE_URL=$GO_API_BASE_URL"
echo "NODE_ENV=$NODE_ENV"

# Verify Next.js files exist
if [ ! -d "/app/web/.next" ]; then
  echo "ERROR: Next.js build directory not found!"
  ls -la /app/web/
  exit 1
fi

echo "Next.js files found, starting application..."
ls -la /app/web/.next/

# Start Next.js (this will block and keep container alive)
exec npm start
