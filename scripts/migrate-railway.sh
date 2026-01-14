#!/bin/bash

# Database migration script for Railway
set -e

echo "🗄️  Running database migrations..."

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL not set. Please configure your database connection."
    exit 1
fi

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
until pg_isready -d "$DATABASE_URL"; do
    echo "Database unavailable, sleeping..."
    sleep 2
done

echo "✅ Database is ready!"

# Run migrations (assuming you have a migration tool)
# Uncomment and adjust based on your migration setup
# cd /app/server
# go run ./cmd/migrate up

echo "🎉 Database migrations completed!"