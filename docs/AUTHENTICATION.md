# Authentication Setup Guide

This document explains how to set up NextAuth.js authentication for the CMS application.

## Prerequisites

1. GitHub OAuth App (for authentication)
2. Environment variables configured

## Setup Steps

### 1. Create GitHub OAuth App

1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create a new OAuth App
3. Set Homepage URL: `http://localhost:3000`
4. Set Authorization callback URL: `http://localhost:3000/api/auth/callback/github`
5. Note the Client ID and generate a Client Secret

### 2. Configure Environment Variables

Copy `.env.example` to `.env.local`:

```bash
cp web/.env.example web/.env.local
```

Update the values:

```env
# NextAuth.js Configuration
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-secret-key-here-change-this-in-production

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Backend API URL
GO_API_BASE_URL=http://localhost:8080
```

### 3. Database Setup

If using PostgreSQL, run the migrations:

```bash
psql -d cms_ai -f server/migrations/001_initial.sql
psql -d cms_ai -f server/migrations/002_auth_update.sql
```

### 4. Start the Services

Start the backend server:

```bash
cd server
go run ./cmd/server
```

Start the frontend:

```bash
cd web
npm run dev
```

## How It Works

1. **Authentication Flow**: Users sign in with GitHub via NextAuth.js
2. **User Creation**: On first sign-in, a user and organization are automatically created
3. **Token Handling**: NextAuth.js manages JWT tokens for session handling
4. **API Security**: Frontend API routes include auth headers when calling the backend
5. **Backend Validation**: Go backend validates tokens and extracts user/org information

## Testing the Auth Flow

1. Navigate to `http://localhost:3000`
2. Click "Sign in with GitHub"
3. Complete the GitHub OAuth flow
4. You should be redirected back to the app with your user session

## Security Considerations

- In production, use a strong `NEXTAUTH_SECRET`
- Configure proper GitHub OAuth scopes as needed
- Consider adding additional providers (Google, email/password)
- Implement proper rate limiting and session management

## Development Notes

- The system currently uses header-based auth in the backend for compatibility
- Frontend routes automatically include auth headers via the `getAuthHeaders` utility
- User roles and permissions are handled through the existing RBAC system