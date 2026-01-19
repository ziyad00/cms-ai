# Implementation Plan

## Goal
Turn the current MVP (Phase 1 complete) into a production-ready SaaS platform (Phases 2–3).

---

## Phase 2: Production Hardening & Core Features (Next 2–4 weeks)

### 2.1 Real Renderer & Previews
- [ ] Install `python-pptx` and test end-to-end export.
- [ ] Implement actual preview thumbnail generation (PNG per slide).
- [ ] Wire job endpoints to create jobs; worker processes them.
- [ ] Add job status polling in frontend UI.
- [ ] Add file download endpoint serving rendered assets.
- [ ] Store preview thumbnails as assets and expose via API.

### 2.2 Authentication & Authorization
- [ ] Replace dev header auth with NextAuth or similar.
- [ ] Add login/signup flows.
- [ ] Add organization management UI.
- [ ] Implement role-based access checks in UI (hide actions).
- [ ] Add session management.

### 2.3 Job Queue & Workers
- [ ] Implement real job processing (render/preview/export) in worker.
- [ ] Add job deduplication keys.
- [ ] Add retry with exponential backoff.
- [ ] Add dead-letter queue for failed jobs.
- [ ] Add job progress updates (optional SSE/websocket).
- [ ] Add job cancellation support.

### 2.4 Asset Management & Storage
- [ ] Implement actual object storage abstraction (S3, GCS, or local).
- [ ] Add signed URL generation for private downloads.
- [ ] Add file upload with virus scanning/stub security checks.
- [ ] Store previews as immutable assets per version.

### 2.5 Monitoring & Observability
- [ ] Replace `log.Printf` with structured JSON logging.
- [ ] Add Prometheus metrics (HTTP, job queue latency, render errors).
- [ ] Add health checks for dependencies (DB, renderer, storage).
- [ ] Add alerting on job queue backlog.
- [ ] Add request tracing (IDs are present).
- [ ] Add rate limiting on auth and generation endpoints.

---

## Phase 3: Product & Scale (Next 4–8 weeks)

### 3.1 Advanced Generation & AI
- [ ] Add real AI orchestrator (prompt → spec) with GPT/vertex/claude.
- [ ] Add brand kit inference in generation.
- [ ] Add multi-step generation with repair loops per spec.
- [ ] Add quota enforcement and upgrade flows.
- [ ] Add real metering event types (preview, download, brand kit usage).

### 3.2 Advanced Editor
- [ ] Build visual spec editor (drag placeholders, edit geometry).
- [ ] Add live validation as user edits (overlap, contrast).
- [ ] Add component library (cards, dividers, accent bars).
- [ ] Add template gallery and sharing.
- [ ] Add “Create from brand kit” flow.

### 3.3 Brand Kits & Organization Features
- [ ] Brand kit CRUD UI (colors, fonts, logos, guidance).
- [ ] Org profile/settings management.
- - [ ] Member invite and role management UI.
- - [ ] Audit log viewer for admins.
- - [ ] Per-org usage dashboards.

### 3.4 Performance & Scale
- [ ] Redis layer for caching frequent data (templates, user sessions).
- - [ ] Background cleanup jobs (expire old assets, aggregates).
- - [ ] Horizontal scaling for workers (multiple pods).
- - [ ] CDN for static assets (Next.js export, public downloads).
- - [ ] Database connection pooling, read replicas.

---

## Phase 4: Enterprise & Ecosystem (Future)

### 4.1 Integrations
- [ ] External PPTX import/export.
- [ ] LMS integrations (training, onboarding).
- - [ ] API for third-party plugins (custom layouts).
- - [ ] Webhook notifications for job completion.

### 4.2 Analytics
- - [ ] Template usage analytics.
- - [ ] Generation success/failure metrics.
- - [ ] Per-org performance reporting.
- - [ ] A/B testing framework for templates.

### 4.3 Compliance & Security
- - [ ] SOC 2 Type 2 readiness checklist.
- - [ ] Data retention policies per org.
- - [ ] Content moderation for generated assets.
- - [ ] Automated security scanning (dependency, runtime).

---

## Immediate Next Run Checklist

- [ ] `pip install python-pptx`
- [ ] Set up Postgres instance and run `migrations/001_initial.sql`
- [ ] Run server with `DATABASE_URL`.
- [ ] Test end-to-end: generate → export → download.
- [ ] Verify worker processes jobs by adding logs to `processJobs`.
- [ ] Add structured logger in Go.

---

## File System & Directory Adjustments

- Move from header-based auth to JWT or session-based auth in `internal/auth/`.
- Add `internal/queue/` package for job queue abstraction.
- Add `internal/storage/` package for object storage interface.
- Add config management (`config/` directory with env defaults).
- Add Docker Compose for local development stack (Postgres, Redis, MinIO optional).
- Add deployment scripts and helm charts.

---

## Risks & Mitigations

| Risk | Mitigation |
|-------|------------|
| Renderer subprocess | Sandboxing, strict input validation, controlled paths |
| Job loss on restart | Persistent queue (Postgres) + retry |
| Unbounded Python process | Resource limits, timeouts, monitoring |
| Secrets in env | Centralize in secret manager at scale |
| Database lock contention | Connection pooling, read replicas for reads |
| Large file uploads | Streaming, size limits, chunked processing |
| Concurrent job processing | Worker pools, rate limiting, deduplication |
| Unsafe user inputs | Schema validation, sanitization, allowlists |

---

## Success Metrics (Phase 2)

- Export success rate ≥ 99.5%
- P95 job latency < 30s for PPTX generation
- Login/signup conversion rate > 80%
- No unauthenticated API access
- All critical paths covered by integration tests
- Production-ready Docker image < 200MB
- Zero secrets in repo (checked by CI)
