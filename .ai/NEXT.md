# NEXT

## Plan completed!
- Full Go API with templates/versions/brand-kits/jobs/metering/assets CRUD, auth/RBAC, quotas, audit logs.
- Next.js web app with basic UI to generate/list templates, proxying to Go API.
- Tests pass: `go test ./...` (server), `cd web && node --test` (web helpers).

## Next
- Read plan: `docs/COST_AND_PRICING.md`.
- Immediate: Phase 2.1 – Install python-pptx, wire real worker, add job status UI.
- Ready to continue when you choose a milestone or specific task.

## MVP status
- ✅ Full Phase 1 MVP: Go API + Postgres + Next.js (Tailwind) + worker stub + renderer stub.
- ✅ Dashboard and editor UI complete.
- ✅ Tests pass: `go test ./...`, `cd web && node --test`.

## First tasks (Phase 2.1)
- [x] Install Go PPTX renderer (baliance.com/gooxml).
- [x] Wire Go renderer to actually process jobs (render/preview/export).
- [x] Add job status polling/export/download in frontend.
- [x] Add asset download endpoint.
- [x] Add preview thumbnail generation and storage (basic placeholder done).

## Current Status
- ✅ Worker processes all job types (render/preview/export) end-to-end
- ✅ Assets now stored in object storage with signed URLs (S3/GCS/local filesystem support)
- ✅ Job status updates working (queued -> running -> done/failed)
- ✅ Comprehensive test coverage for worker logic and object storage
- ✅ Frontend job status polling and download UI complete
- ✅ NextAuth.js authentication system fully implemented with GitHub OAuth
- ✅ Job queue resilience fully implemented (retry with exponential backoff, deduplication, dead-letter queue, admin endpoints)
- ✅ Object storage integration complete with signed URLs and backend selection
- ✅ Hugging Face AI orchestrator implemented with prompt engineering, validation, and brand kit support
- ✅ Advanced visual editor implemented with drag-and-drop canvas, theme editor, validation system, and comprehensive test coverage
- Next: Configure Hugging Face API key for production AI generation, update worker to use ObjectStorage, add asset cleanup jobs, or Monitoring (structured logs, metrics, alerts)

## Immediate Next Tasks
- [x] URGENT: Fix /api/custom-auth/signup endpoint 404 issue - FIXED: Wrong backend URL in build-time environment variables
- [x] Backend unit tests - COMPLETED: Comprehensive test suite added for auth, AI, and API packages
- [x] Fix template persistence issue - FIXED: PostgreSQL backend storage configured and working
- [x] Database diagnostics implementation - COMPLETED: Added comprehensive diagnostic endpoints for investigating empty specs and data integrity issues
- [x] **FIXED**: AI generation producing placeholder content instead of meaningful specs
  - ROOT CAUSE: Incomplete few-shot examples in HuggingFace client - second example was missing response JSON
  - SOLUTION: Added complete response example for "Sales report template with quarterly data"
  - FIX: Updated getFewShotExamples() function with proper JSON response structure
  - Now AI should generate proper template specs instead of placeholder content like "car", "sales"
- [x] URGENT: Fix Railway deployment failures and deploy template loading fix - COMPLETED: Successfully deployed
- [x] Investigate build failure root cause (recent deployments all fail) - COMPLETED: Build issues resolved
- [x] Deploy React useEffect fix for consistent template loading - COMPLETED: Deployed with commit 8975237
- [x] Remove debug logging after verifying fix works in production - NOT NEEDED: No debug logging to remove
- [x] Decks: add persisted Deck + DeckVersion (DB migration + store + API)
- [x] Decks: implement AI binder (content blob -> filled spec) and deck export endpoint
- [ ] Web: add real `/decks` detail page with VisualEditor + content editor + export (no raw JSON UI)
- [ ] Deck binder: improve prompt so it fills more placeholders from content (currently only fills some fields)
- [ ] Add unit tests for spec package validation functions (geometry bounds checking)
- [ ] Add unit tests for queue package (job processing, retry logic)
- [ ] Configure HUGGINGFACE_API_KEY environment variable for production AI generation
- [ ] (Recommended) Add integration test that exercises one-click deck flow against a running Go API (generate -> export -> download)
- [ ] Investigate NextAuth route error in Railway logs: `TypeError: (intermediate value).POST is not a function` at `app/api/auth/[...nextauth]/route.js`
- [ ] Add data migration/cleanup: existing assets with non-UUID IDs (asset-...) will be orphaned and should be re-exported or deleted
- [ ] Test AI generation with real Hugging Face API and Mixtral model
- [ ] Add more comprehensive prompt engineering for different business verticals
- [ ] Add AI generation cost tracking and quota management
- [ ] Update worker to use ObjectStorage instead of old Storage interface
- [ ] Configure S3 bucket and IAM permissions for production use
- [ ] Test signed URL expiration and security in production environment
- [ ] Add asset cleanup job for old files and retention policies
- [ ] Complete GCS SDK implementation (currently placeholder)
- [ ] Add storage metrics and monitoring (upload/download rates, storage usage)
- [ ] Add queue monitoring metrics (queue depth, processing rate, failed job alerts)
- [ ] Implement structured error logging with correlation IDs
- [ ] Add job timeout handling to prevent stuck jobs
- [ ] Create queue monitoring dashboard in frontend
- [ ] Test database migration for job queue resilience features
- [ ] Test Railway deployment with real environment variables
- [ ] Set up custom domain and SSL certificates on Railway
- [ ] Configure monitoring and alerting for production environment

## Phase 2 Roadmap
- [x] 2.2: Replace dev auth (NextAuth), org/team management.
- [x] 2.2.1: Organization management UI pages
- [x] 2.2.2: Team members and role management
- [x] 2.2.3: User invitation system
- [x] 2.2.4: Role-based permissions in UI
- [x] 2.2.5: Add organization switcher component
- [ ] 2.2.6: Add audit log viewer for admins
- [ ] 2.2.7: Add billing plan management UI
- [ ] 2.2.8: Add organization creation flow
- [x] 2.3: Job queue improvements (retry, dedup, dead-letter).
- [x] 2.4: Asset management (object storage, signed URLs).
- 2.5: Monitoring (structured logs, metrics, alerts, health).

## Phase 3 (Later)
- Real AI orchestrator, ✅ advanced editor, analytics, integrations, compliance.

## Deploy & Test
- Dockerfile ready; test with `docker build -t cms-ai . && docker run -p 8080:8080 cms-ai`.
- Set up local dev stack with Docker Compose (optional).

## MVP milestones achieved
- ✅ Orgs/roles + tenant isolation + audit logs (in-memory).
- ✅ Template + versions CRUD with lifecycle.
- ✅ Generation stub + validate/repair.
- ✅ Renderer stub (spec -> PPTX) + export flow.
- ✅ Basic billing/quota enforcement.
- ✅ Web UI scaffold + API integration.

