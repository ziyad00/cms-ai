#!/bin/bash

# Initialize PostgreSQL database for Railway
set -e

echo "🗄️  Initializing CMS AI database..."

# Create extensions
psql "$DATABASE_URL" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
psql "$DATABASE_URL" -c "CREATE EXTENSION IF NOT EXISTS \"pg_trgm\";"

# Run your schema initialization here
# psql "$DATABASE_URL" -f /app/schema/schema.sql

echo "✅ Database initialization completed!"