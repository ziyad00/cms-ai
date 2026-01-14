# Cost and Pricing Guide

## Overview

This document defines the cost model for the PowerPoint Template Generation CMS, including per-slide costs, per-user daily costs, and pricing strategies for a SaaS product.

---

## 1. Per-Slide / Per-Image Costs (2025)

| Service | Model | Per Image | Notes |
|--------|-------|--------|--------|
| **DALL·E 3** | $0.040 | Standard |
| **DALL·E 3 HD** | $0.080 | Standard |
| **ChatGPT 4o (Aug 2025)** | ~$0.02 | Can generate text content first |
| **Stable Diffusion 3.5** | $0.015 | Fast, consistent |
| **OpenAI (Platform)** | Variable | Starting at ~$0.003/image |
| **Midjourney** | Subscription | Fast |
| **Stability AI DreamStudio** | Credit based | ~$0.02/image |
| **Stability AI (new)** | Credit based | ~$0.015/image |

### Image Generation Cost per Slide
- **Low-end**: $0.02–$0.05 per slide (e.g., Stable Diffusion 3.5 Medium)
- **Mid-range**: $0.10–$0.30 per slide (e.g., DALL·E 3)
- **High-end**: $0.40–$0.80 per slide (e.g., DALL·E 3 HD)

---

## 2. Template/Presentation-Specific Services

| Service | Model | Per Template | Slides Included | Monthly Cost | Notes |
|--------|-------|--------|----------|-------------|-------|
| **Beautiful.ai** | Custom AI | $12–$20 (Pro plan) | ~10–20 slides | $12–$20 |
| **Tome** | AI + Template | $5–$8 (Pro plan) | ~8–15 slides | $5–$8 |
| **SlidesGo** | AI Generator | $1–$3 (various plans) | ~3–5 slides | $1–$3 |
| **Pitch** | AI Generator | $8–$12 (various plans) | ~8–15 slides | $8–$12 |

---

## 3. Cost Estimates by Slide Count

| Slides | Low Cost | Mid Cost | High Cost |
|-------|----------|-----------|-----------|
| **10 slides** | $0.20–$0.40 | $0.60–$1.00 | $2.00+ |
| **20 slides** | $0.40–$0.80 | $1.20–$2.00 | $3.00+ |
| **50 slides** | $1.00–$2.00 | $5.00–$10.00 | $15.00+ |
| **100 slides** | $2.00–$4.00 | $10.00–$20.00 | $30.00+ |

---

## 4. Per-User / Per-Day Costs

Assuming unlimited generations:

| User Type | Slides/Day | Low Cost | Mid Cost | High Cost |
|-----------|-----------|-----------|-----------|
| **Heavy User** | 50 slides/day | $2–$5/day | $10–$25/day | $25+ |
| **Standard User** | 10 slides/day | $0.40–$1.00/day | $1.20–$3.00/day | $3.00+ |
| **Light User** | 5 slides/day | $0.20–$0.50/day | $0.60–$1.50/day | $1.50+ |

---

## 5. Recommended Pricing Model for SaaS

### Option 1: Credits Per Use (Recommended)
- **Revenue**: Credits × Unit Price
- **Unit Price**: $0.01/credit
- **Credits per slide**: 1–4 credits (depending on model)
- **Pros**: Simple, predictable scaling
- **Cons**: Revenue varies with AI model costs

### Option 2: Slide Packs
- **Basic Deck (10 slides)**: 2 credits
- **Standard Deck (20 slides)**: 5 credits
- **Professional Deck (50 slides)**: 12 credits
- **Enterprise Deck (100 slides)**: 25 credits
- **Cost per slide**: $0.10–$0.25

### Option 3: Tiered Subscription (Hybrid)
- **Basic**: $10/mo, up to 100 slides
- **Professional**: $20/mo, up to 500 slides
- **Enterprise**: $50/mo, unlimited slides
- **Overages**: $0.25/credit

---

## 6. Cost Prediction Before Generation

### 6.1 Input Validation
```typescript
interface GenerationRequest {
  slideCount: number;
  quality: 'basic' | 'standard' | 'premium';
  brandKitId?: string;
}
```

### 6.2 Cost Estimation
```typescript
function estimateCost(req: GenerationRequest): CostEstimate {
  const costs = {
    basic: { perSlide: 0.15, tokensPerSlide: 300, credits: 1 },
    standard: { perSlide: 0.30, tokensPerSlide: 500, credits: 2 },
    premium: { perSlide: 0.50, tokensPerSlide: 800, credits: 4 }
  };
  
  const model = costs[req.quality];
  return {
    slides: req.slideCount,
    model: req.quality,
    tokensRequired: req.slideCount * model.tokensPerSlide,
    estimatedCost: req.slideCount * model.credits
  };
}
```

### 6.3 Quota Check
```typescript
function checkQuota(org: Organization, req: GenerationRequest): boolean {
  const plan = getPlan(org);
  const cost = estimateCost(req);
  return (
    org.creditsUsed + cost.estimatedCost <= plan.creditsLimit &&
    org.slidesUsed + req.slideCount <= plan.maxSlidesPerMonth
  );
}
```

---

## 7. Usage Analytics

### 7.1 Metrics to Track
- Slides generated per organization
- Credits consumed per organization
- Cost per organization (monthly)
- Generations per user
- Failures and retries
- Model usage distribution

### 7.2 Dashboard Views
- Organization usage summary
- Cost breakdown by model
- User activity heatmap
- Revenue and quota status

---

## 8. Revenue Model Options

### 8.1 Pure Usage-Based
- **Revenue**: Credits × Unit Price
- **Pros**: Simple, predictable scaling
- **Cons**: Revenue varies with AI model costs

### 8.2 Tiered Subscription
- **Basic**: $5/mo, up to 100 slides
- **Professional**: $20/mo, up to 500 slides
- **Enterprise**: $50/mo, unlimited slides
- **Overages**: $0.25/credit

### 8.3 Hybrid (Recommended)
- **Base Subscription**: $10/mo (includes 50 slides)
- **Usage Credits**: $0.25/credit for overages
- **Premium Credits**: $1.00/credit for advanced features

---

## 9. Implementation Checklist

### 9.1 Backend
- [ ] Add `slideCount` to generation request payload.
- [ ] Add cost estimation before generation.
- [ ] Enforce quotas per organization.
- [ ] Record metering events with cost metadata.
- [ ] Add structured logging for cost tracking.

### 9.2 Frontend
- [ ] Add slide count input or preset templates.
- [ ] Show cost estimate before generation.
- [] Display usage and quota status.
- [] Add upgrade/overage flow when limits exceeded.

### 9.3 Admin
- [ ] Organization plan management (slide limits, quotas).
- [ ] Cost and usage analytics dashboard.
- [] Revenue reporting and forecasting.
- [] Audit logs for cost attribution.

---

## 10. Risks & Mitigations

| Risk | Mitigation |
|-------|------------|
| AI model price spikes | Lock model version for a billing period; add alerts on cost changes. |
| Unbounded generation | Hard limits per org; rate limit per user. |
| Large file uploads | Size limits, chunked uploads, virus scanning. |
| Secrets in env | Centralize in secret manager; rotate keys. |
| Job loss on restart | Persistent queue + retry with deduplication. |
| Renderer subprocess | Sandboxing, controlled paths, timeouts. |

---

## 11. Success Metrics (Phase 2)

- Export success rate ≥ 99.5%
- P95 job latency < 30s for PPTX generation
- Login/signup conversion rate > 80%
- No unauthenticated API access
- All critical paths covered by integration tests
- Zero secrets in repo (checked by CI)
- Cost per slide < $0.25 for 90% of generations

---

## 12. Next Steps

- Choose a pricing model (credits, packs, or tiers).
- Add cost estimation to generation flow.
- Implement quota enforcement and usage tracking.
- Add cost analytics dashboard.
- Set up billing and upgrade flows.

---

## 13. File Locations

- `docs/COST_AND_PRICING.md` – This file.
- `server/internal/pricing/` – Pricing engine (if you want code).
- `web/app/pricing/` – Frontend pricing UI (if you want UI).
- `server/internal/store/pricing.go` – Store pricing plans and usage.

---

## 14. References

- `specs.md` – Product specification.
- `IMPLEMENTATION_PLAN.md` – Implementation roadmap.
- `server/migrations/001_initial.sql` – Database schema.
- `server/internal/store/store.go` – Store interfaces.
- `server/internal/api/router_v1.go` – API endpoints.
- `web/app/page.js` – Frontend dashboard.

---

**Last updated**: 2025-01-13.