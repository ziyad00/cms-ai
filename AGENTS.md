# Ralph AI Agent Learnings

## Iteration 1 - 2026-02-07

### Project Context
- Found existing CMS AI project with comprehensive development history
- Project has robust Next.js + Go backend architecture
- Extensive test suite and Railway deployment setup already exists
- Current PRD has sample structure with basic story

### Key Learnings
- Project already well-established with advanced features (AI generation, PPTX rendering, smart design)
- Need to understand if PRD needs updating for real project requirements
- Should focus on incremental improvements rather than basic setup

### Technical Observations
- Go backend with comprehensive API
- Next.js frontend with auth and visual editor
- PostgreSQL database with migrations
- Object storage integration (S3/GCS)
- Comprehensive test coverage

### Current Mission: Export Functionality Crisis
- CRITICAL: PPTX export with olama AI renderer is broken after Feb 5-6 deployments
- Root cause identified: Railway deployment architecture mismatch
- Solution path: Fix Go renderer paths + Python dependencies + end-to-end testing

### Technical Discovery
- Railway deploys Next.js frontend from web/ directory, not Go backend from server/
- Olama Python files needed to be copied from server/tools/ to web/tools/ (DONE)
- Go PythonPPTXRenderer still points to old server-relative paths
- Python dependencies may not be installed during Next.js build process

### ✅ Iteration 2 Complete - 2026-02-07
**MAJOR SUCCESS**: Fixed PPTX export crisis!

#### Key Achievements:
- ✅ CRITICAL-001 completed - System validation successful
- ✅ STORY-001 completed - Fixed Go renderer Python script paths
- ✅ Integration tests passing for AI-enhanced PPTX export
- ✅ Smart path fallback for Railway vs local development

#### Technical Solutions Applied:
1. **Path Resolution Fix**: Updated renderer.go with intelligent fallback logic
   - Primary: `/app/tools/renderer/render_pptx.py` (Railway)
   - Fallback: Navigate up directories to find local `tools/renderer/render_pptx.py`
2. **AI Renderer Fix**: Modified ai_enhanced_renderer.go to use default paths with fallback
3. **Validation**: Integration tests `TestCompleteAIPipeline` and `TestAIGenerationToRendering` now pass

### ✅ Iteration 3 Complete - 2026-02-07
**SUCCESS**: Railway Python dependency configuration complete!

#### Key Achievements:
- ✅ STORY-002 completed - Railway Python dependency installation configured
- ✅ Created web/nixpacks.toml with proper Python3, pip, and requirements.txt setup
- ✅ Tested build process simulation - Python modules install successfully
- ✅ All integration tests still passing after configuration changes

#### Technical Solutions Applied:
1. **Railway Configuration**: Created web/nixpacks.toml to handle Python dependencies in Next.js deployment
2. **Build Process**: Added pip install step for python-pptx and httpx during Railway build
3. **Validation**: Confirmed Python script execution works from web directory with all dependencies

#### Next Iteration Guidance:
- Focus on STORY-003: End-to-end workflow testing with real deployment
- Export functionality working locally with proper dependency management
- Railway deployment should now install Python packages automatically
- Validate that export job status shows "Completed" instead of "Queued" in production