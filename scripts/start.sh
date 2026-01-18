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
NEXTJS_PORT=${PORT:-3000}
export PORT=$NEXTJS_PORT
export NODE_ENV=production
export GO_API_BASE_URL=http://localhost:$GO_API_PORT

echo "Starting Next.js on port $NEXTJS_PORT..."
echo "GO_API_BASE_URL=$GO_API_BASE_URL"

# Start Next.js (this will block and keep container alive)
exec npm start
