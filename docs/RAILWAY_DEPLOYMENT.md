# Railway Deployment Guide

## Overview
This guide explains how to deploy the CMS AI application to Railway using the provided configuration files.

## Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **Railway CLI**: Install with `npm install -g @railway/cli`
3. **Domain**: Custom domain (optional)
4. **External Services**:
   - GitHub OAuth App (for authentication)
   - Hugging Face API key (for AI generation)
   - S3/GCS bucket (for asset storage, production only)

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone <your-repo>
   cd cms-ai
   ```

2. **Configure environment variables**:
   ```bash
   cp .env.railway.template .env
   # Edit .env with your values
   ```

3. **Deploy to Railway**:
   ```bash
   ./scripts/deploy-railway.sh production
   ```

## Configuration Files

### `railway.json`
Main Railway configuration file defining:
- Docker build settings
- Service definitions (API + Database)
- Environment variables for production/staging
- Health checks and restart policies

### `Dockerfile.railway`
Optimized multi-stage Docker build for Railway:
- Go backend build with Alpine Linux
- Next.js frontend build
- Production runtime with health checks
- Non-root user security

### `docker-compose.railway.yml`
Local development setup matching Railway environment:
- API service with health checks
- PostgreSQL database
- Redis (optional, for caching)

## Environment Variables

### Required for all environments:
- `DATABASE_URL`: PostgreSQL connection (Railway provides this)
- `NEXTAUTH_SECRET`: Secret for JWT signing
- `JWT_SECRET`: Secret for API JWT tokens

### Authentication:
- `GITHUB_CLIENT_ID`: GitHub OAuth App ID
- `GITHUB_CLIENT_SECRET`: GitHub OAuth App Secret

### AI Services:
- `HUGGINGFACE_API_KEY`: For AI template generation

### Storage (Production):
- `STORAGE_TYPE`: Set to 's3' or 'gcs'
- `AWS_*`: S3 credentials and bucket info
- `GCS_*`: GCS credentials and bucket info

## Custom Domain Setup

1. **Add custom domain in Railway Dashboard**:
   - Go to Settings → Domains
   - Add your custom domain

2. **Update DNS**:
   ```
   CNAME www.yourdomain.com -> up.railway.app
   A yourdomain.com -> 76.76.19.61
   ```

3. **Update environment variables**:
   ```bash
   railway variables set NEXTAUTH_URL=https://yourdomain.com
   railway variables set CORS_ORIGINS=https://yourdomain.com
   ```

## Database Setup

Railway automatically provides PostgreSQL. To initialize:

1. **Run migrations** (if you have them):
   ```bash
   ./scripts/migrate-railway.sh
   ```

2. **Manual setup** (if needed):
   ```bash
   railway connect postgresql
   # Run your SQL commands
   ```

## SSL and Security

- **SSL**: Railway provides automatic SSL certificates
- **Security headers**: Configure in your Go server
- **CORS**: Set via `CORS_ORIGINS` environment variable

## Monitoring and Logs

### Viewing logs:
```bash
railway logs
```

### Health checks:
- Endpoint: `/health`
- Checks every 30 seconds
- Automatic restarts on failure

### Monitoring setup:
- Add Sentry for error tracking (`SENTRY_DSN`)
- Use Railway's built-in metrics
- Set up alerts in Railway dashboard

## Deployment Workflow

### Production deployment:
```bash
./scripts/deploy-railway.sh production
```

### Staging deployment:
```bash
./scripts/deploy-railway.sh staging
```

### Manual deployment:
```bash
railway up
```

## Troubleshooting

### Common issues:

1. **Build failures**:
   - Check Dockerfile.railway for syntax
   - Verify Go modules are properly vendored
   - Ensure all dependencies in package.json

2. **Runtime errors**:
   - Check logs: `railway logs`
   - Verify environment variables
   - Test health endpoint: `curl https://your-app.railway.app/health`

3. **Database issues**:
   - Verify DATABASE_URL format
   - Check database migrations
   - Test connection manually

4. **Authentication issues**:
   - Verify GitHub OAuth configuration
   - Check NEXTAUTH_URL matches deployed URL
   - Ensure secrets are properly set

### Debug mode:
```bash
railway variables set LOG_LEVEL=debug
railway up
```

## Performance Optimization

### For production:
- Enable S3/GCS storage (not local filesystem)
- Configure CDN for static assets
- Set up Redis for caching
- Use Railway's GPU instances for AI workloads

### Environment-specific optimizations:
- Production: Minimize logs, enable monitoring
- Staging: Debug logging, full error traces

## CI/CD Integration

### GitHub Actions example:
```yaml
name: Deploy to Railway
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to Railway
        run: railway up
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

## Cost Optimization

- **Database**: Use shared PostgreSQL for small apps
- **Build**: Optimize Docker layers for faster builds
- **Storage**: Use S3/GCS with lifecycle policies
- **Monitoring**: Set up alerts to prevent overages

## Support

- **Railway docs**: [docs.railway.app](https://docs.railway.app)
- **Issue tracking**: Use Railway dashboard
- **Community**: Railway Discord community