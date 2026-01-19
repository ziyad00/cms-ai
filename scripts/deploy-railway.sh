#!/bin/bash

# Railway deployment script for CMS AI
set -e

echo "🚀 Deploying CMS AI to Railway..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing..."
    npm install -g @railway/cli
fi

# Login to Railway (if not already logged in)
if ! railway whoami &> /dev/null; then
    echo "🔐 Please login to Railway:"
    railway login
fi

# Create or select project
echo "📋 Setting up Railway project..."
if ! railway project &> /dev/null; then
    railway project create cms-ai
else
    railway project cms-ai
fi

# Set up environment
echo "🏗️  Setting up environment variables..."

# Check if we're in production or staging
ENVIRONMENT=${1:-production}

echo "Deploying to environment: $ENVIRONMENT"

# Set environment variables
railway variables set ENV=$ENVIRONMENT
railway variables set PORT=8080

# NextAuth variables
railway variables set NEXTAUTH_URL=https://$RAILWAY_PUBLIC_DOMAIN
railway variables set NEXTAUTH_SECRET=$NEXTAUTH_SECRET
railway variables set GITHUB_CLIENT_ID=$GITHUB_CLIENT_ID
railway variables set GITHUB_CLIENT_SECRET=$GITHUB_CLIENT_SECRET

# AI/ML variables
railway variables set HUGGINGFACE_API_KEY=$HUGGINGFACE_API_KEY

# Storage variables (for production)
if [ "$ENVIRONMENT" = "production" ]; then
    railway variables set STORAGE_TYPE=s3
    railway variables set AWS_REGION=$AWS_REGION
    railway variables set AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
    railway variables set AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
    railway variables set S3_BUCKET=$S3_BUCKET
    railway variables set S3_ENDPOINT=$S3_ENDPOINT
else
    railway variables set STORAGE_TYPE=local
fi

# Security variables
railway variables set JWT_SECRET=$JWT_SECRET
railway variables set CORS_ORIGINS=https://$RAILWAY_PUBLIC_DOMAIN

# Logging
railway variables set LOG_LEVEL=info

# Add PostgreSQL database
echo "🗄️  Setting up PostgreSQL database..."
railway add postgresql

# Deploy application
echo "🚀 Deploying application..."
railway up

echo "✅ Deployment complete!"
echo "🌐 Application URL: https://$RAILWAY_PUBLIC_DOMAIN"
echo "📊 View logs: railway logs"
echo "🔧 Open dashboard: railway open"