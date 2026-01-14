# CMS AI - PowerPoint Template Generation Platform

A modern web-based Content Management System for creating, managing, and exporting PowerPoint templates generated from natural-language prompts with AI.

## 🚀 Features

### AI-Powered Template Generation
- **Natural Language to Templates**: Convert prompts into professional PowerPoint templates
- **Hugging Face Integration**: Uses Mixtral-8x7B model for intelligent template generation
- **Brand Kit Support**: Incorporate organization branding and style guidelines
- **Multi-Language Support**: Generate templates in English and Arabic with RTL support

### Advanced Visual Editor
- **Drag-and-Drop Interface**: Visually position and resize placeholders
- **Real-Time Preview**: See changes instantly as you edit
- **Theme Customization**: Full control over colors, fonts, and spacing
- **Validation System**: Smart constraints and warnings for layout issues

### Complete Template Management
- **Version Control**: Track template changes with immutable versions
- **Organization Management**: Team collaboration with role-based permissions
- **Job Processing**: Asynchronous rendering with real-time status updates
- **Asset Storage**: Scalable object storage with signed URLs

### Production Ready
- **Modern Tech Stack**: Go backend, Next.js frontend, PostgreSQL database
- **Deployable**: Railway deployment with automated CI/CD
- **Comprehensive Testing**: Unit tests, integration tests, and E2E workflows
- **Monitoring**: Health checks, logging, and error tracking

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Next.js App   │    │   Go API Server │    │  Hugging Face   │
│   (Frontend)    │◄──►│   (Backend)     │◄──►│     AI Model    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   PostgreSQL    │    │  Object Storage │
                       │   Database      │    │   (S3/GCS)     │
                       └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm
- Go 1.21+
- Hugging Face API key (free tier available)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/ziyad00/cms-ai.git
   cd cms-ai
   ```

2. **Set up environment variables**
   ```bash
   # Frontend
   cd web
   cp .env.example .env.local
   
   # Backend
   cd ../server
   export HUGGINGFACE_API_KEY=your_huggingface_api_key
   ```

3. **Start the frontend**
   ```bash
   cd web
   npm install
   npm run dev
   ```

4. **Start the backend**
   ```bash
   cd server
   go mod download
   go run ./cmd/server/main.go
   ```

5. **Access the application**
   - Frontend: http://localhost:3000
   - API: http://localhost:8080

### Railway Deployment

1. **Install Railway CLI**
   ```bash
   npm install -g @railway/cli
   ```

2. **Deploy**
   ```bash
   ./scripts/deploy-railway.sh production
   ```

## 📚 Documentation

- [Implementation Plan](./IMPLEMENTATION_PLAN.md) - Detailed technical specifications
- [Authentication Guide](./docs/AUTHENTICATION.md) - NextAuth.js setup
- [AI Integration](./docs/HUGGINGFACE_AI_INTEGRATION.md) - Hugging Face configuration
- [Railway Deployment](./docs/RAILWAY_DEPLOYMENT.md) - Production deployment guide
- [Cost & Pricing](./docs/COST_AND_PRICING.md) - Usage costs and pricing models

## 🧪 Testing

### Backend Tests
```bash
cd server
go test ./... -v
```

### Frontend Tests
```bash
cd web
npm test
```

### End-to-End Tests
```bash
cd web
node --test
```

## 🎯 Core Workflows

### 1. AI Template Generation
```bash
curl -X POST http://localhost:8080/v1/templates/generate \
  -H "Content-Type: application/json" \
  -H "X-User-ID: test-user" \
  -H "X-Org-ID: test-org" \
  -H "X-User-Role: editor" \
  -d '{
    "prompt": "Create a professional business presentation with title and agenda slides",
    "name": "Business Template",
    "tone": "corporate",
    "language": "EN"
  }'
```

### 2. Template Export
```bash
# Trigger export job
curl -X POST http://localhost:8080/v1/templates/{id}/export \
  -H "Content-Type: application/json" \
  -H "X-User-ID: test-user" \
  -H "X-Org-ID: test-org" \
  -H "X-User-Role: editor"

# Download generated PPTX
curl -H "X-User-ID: test-user" \
     -H "X-Org-ID: test-org" \
     -H "X-User-Role: editor" \
     http://localhost:8080/v1/jobs/{jobId}/assets/export.pptx
```

## 🔧 Configuration

### Environment Variables

**Backend**
- `HUGGINGFACE_API_KEY`: Hugging Face API key
- `DATABASE_URL`: PostgreSQL connection string
- `STORAGE_TYPE`: `local` | `s3` | `gcs`
- `S3_BUCKET`, `AWS_REGION`: S3 configuration
- `NEXTAUTH_SECRET`: JWT secret for authentication

**Frontend**
- `NEXTAUTH_URL`: Auth callback URL
- `NEXT_PUBLIC_API_URL`: Backend API URL

## 📊 API Endpoints

### Templates
- `GET /v1/templates` - List templates
- `POST /v1/templates/generate` - Generate from AI
- `GET /v1/templates/{id}` - Get template
- `POST /v1/templates/{id}/export` - Export PPTX

### Organizations
- `GET /v1/organizations/{id}` - Get organization
- `POST /v1/organizations/{id}/members` - Add member
- `PUT /v1/organizations/{id}` - Update settings

### Jobs
- `GET /v1/jobs/{id}` - Get job status
- `GET /v1/jobs/{id}/assets/{filename}` - Download assets

## 🎨 Frontend Components

- **Visual Editor**: Drag-and-drop template editor
- **Theme Editor**: Color and typography customization
- **Organization Manager**: Team and permissions management
- **Job Status Dashboard**: Real-time job progress tracking

## 🚀 Deployment Options

- **Railway** (Recommended): One-click deployment with managed database
- **Render.com**: Alternative container platform
- **Docker**: Self-hosted with docker-compose
- **Kubernetes**: Enterprise deployment

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Hugging Face](https://huggingface.co) - AI model platform
- [Next.js](https://nextjs.org) - React framework
- [Go](https://go.dev) - Backend programming language
- [Railway](https://railway.app) - Deployment platform

---

**Built with ❤️ for modern template management**
