#!/bin/bash
# scripts/gatekeeper.sh
# Run this before every commit/push!
set -e

echo "🚧 GATEKEEPER: Pre-flight Checks..."

# 1. Check Backend Compilation
echo "📦 Checking Backend Compilation..."
cd server
go build ./...
echo "✅ Compilation OK"

# 2. Run Backend Unit Tests
echo "🧪 Running Backend Unit Tests..."
export JWT_SECRET=this_is_a_very_long_secret_at_least_32_characters
go test ./internal/...
echo "✅ Backend Tests OK"

# 3. Check Frontend Compilation
echo "🌐 Checking Frontend Compilation..."
cd ../web
npm run build
echo "✅ Frontend Build OK"

echo "🎉 ALL SYSTEMS GO. SAFE TO PUSH."
