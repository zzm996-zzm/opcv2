# Project Market V1.4 Closure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver a fully usable and verifiable Project Market V1.4 before migrating the Commercial Sandbox frontend to V1.2.

**Architecture:** Keep the existing Go/Gin modular monolith, PostgreSQL source of truth, Redis/Asynq jobs, and React/Vite frontend. Move the frontend to explicit V1.4 contracts, then connect durable file, retrieval, research, export, analytics, and administration workflows through existing provider boundaries with production-safe startup validation.

**Tech Stack:** Go 1.22, Gin, PostgreSQL 16, Redis, Asynq, React, TypeScript, Vite, Vitest, Playwright/browser-client, PDF renderer.

---

### Task 1: Switch catalog and evidence pages to V1.4 APIs

**Files:**
- Modify: `apps/web/src/pages/ProjectsPage.tsx`
- Modify: `apps/web/src/lib/api.ts`
- Modify: `apps/web/src/types/api.ts`
- Test: `apps/web/src/pages/ProjectsPage.test.tsx`
- Test: `internal/projects/http_test.go`

**Step 1: Write failing frontend contract tests**

Assert that home, catalog, detail, case list, and case detail use `/api/v1/projects/home`, `/api/v1/projects`, `/api/v1/projects/:id`, `/api/v1/project-cases`, and `/api/v1/project-cases/:id`. Cover query serialization for keyword, category, track, budget, difficulty, resource, sort, page, and page size.

**Step 2: Run the focused tests and confirm failure**

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Expected: FAIL because the page still calls legacy opportunity and case routes.

**Step 3: Add typed V1.4 client methods and adapt the page**

Use server `items`, `total`, `page`, and `page_size` values. Remove the hard-coded `120` result total and local reference-data fallback from successful API flows. Keep an explicit empty state and retryable error state.

**Step 4: Add catalog interaction tests**

Cover debounced keyword search, filters, sorting, page changes, request cancellation/race handling, direct detail navigation, empty results, failed requests, and refresh recovery from URL query parameters.

**Step 5: Run focused frontend and backend tests**

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Run: `go test ./internal/projects ./internal/platform/httpserver`

Expected: PASS.

**Step 6: Commit and push**

```bash
git add apps/web/src/pages/ProjectsPage.tsx apps/web/src/lib/api.ts apps/web/src/types/api.ts apps/web/src/pages/ProjectsPage.test.tsx internal/projects/http_test.go
git commit -m "feat(projects): connect v1.4 catalog and evidence pages"
git push origin feature/bootstrap
```

### Task 2: Complete full-open, favorite, and comparison behavior

**Files:**
- Modify: `apps/web/src/pages/ProjectsPage.tsx`
- Modify: `apps/web/src/lib/api.ts`
- Test: `apps/web/src/pages/ProjectsPage.test.tsx`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/postgres_repository.go`
- Test: `internal/projects/service_test.go`
- Test: `internal/projects/postgres_repository_test.go`

**Step 1: Write failing behavior tests**

Cover full-open detail rendering when `feature_paywall_enabled=false`, favorite persistence, removal, five-item comparison, and an actionable comparison-limit response.

**Step 2: Remove active membership gates from the full-open path**

Hide upgrade/unlock controls while the feature flag is false and render all returned blocks. Preserve future locked DTO shapes without making them active.

**Step 3: Connect favorite and comparison mutations**

Use authenticated APIs as the source of truth, roll back optimistic changes after failures, and reload saved state after refresh.

**Step 4: Verify, commit, and push**

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Run: `go test ./internal/projects`

Expected: PASS.

Commit: `feat(projects): complete open catalog interactions`

### Task 3: Implement persistent match file ingestion

**Files:**
- Modify: `apps/web/src/pages/ProjectsPage.tsx`
- Modify: `apps/web/src/lib/api.ts`
- Test: `apps/web/src/pages/ProjectsPage.test.tsx`
- Modify: `internal/projects/files/manager.go`
- Modify: `internal/projects/files/parser.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Modify: `internal/projects/postgres_repository.go`
- Test: `internal/projects/files/manager_test.go`
- Test: `internal/projects/files/parser_test.go`
- Test: `internal/projects/http_test.go`

**Step 1: Write failing upload and ownership tests**

Cover supported types, size limits, MIME/header mismatch, malicious names, missing files, expired files, cross-user access, duplicate attachment, parse failure, and successful parsed-text persistence.

**Step 2: Add upload lifecycle APIs**

Implement create/upload, status, attach-to-session, and delete endpoints using stable file IDs. Persist owner, detected MIME, checksum, object key, parse state, extracted text metadata, and failure detail.

**Step 3: Enable the frontend upload control**

Show upload progress, parse progress, success, retry, and removal. Restore attached files when a match session is reopened.

**Step 4: Verify, commit, and push**

Run: `go test ./internal/projects/...`

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Expected: PASS.

Commit: `feat(projects): add persistent match file ingestion`

### Task 4: Connect retrieval, web research, and evidence-backed matching

**Files:**
- Modify: `internal/platform/projectprovider/factory.go`
- Modify: `internal/platform/config/config.go`
- Modify: `.env.example`
- Modify: `.env.server.example`
- Modify: `apps/api/main.go`
- Modify: `apps/worker/main.go`
- Modify: `internal/projects/retrieval/`
- Modify: `internal/projects/research/`
- Modify: `internal/projects/worker.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/postgres_repository.go`
- Test: `internal/platform/projectprovider/factory_test.go`
- Test: `internal/projects/worker_test.go`
- Test: `internal/projects/service_test.go`

**Step 1: Write failing provider and pipeline tests**

Cover production startup rejection for unavailable placeholder providers, retrieval ranking, evidence deduplication, source quality, prompt-injection text isolation, provider timeout, partial research, insufficient evidence, cancellation, and idempotent retry.

**Step 2: Implement configured production adapters**

Connect the selected object store, vector/keyword retrieval, and web search/fetch providers. Validate all required configuration at startup and keep explicit local providers for development/test only.

**Step 3: Replace catalog-wide generation with the evidence pipeline**

Retrieve ranked candidates, assess sufficiency, research only identified gaps, persist sources and claims, and generate recommendations from the evidence bundle. Mark incomplete runs as partial or insufficient instead of returning unsupported certainty.

**Step 4: Expose durable progress and recovery in the frontend**

Reconnect SSE using the last event ID, fall back to status polling, restore current stages after refresh, and present cancel/retry controls according to persisted state.

**Step 5: Verify, commit, and push**

Run: `go test ./internal/platform/projectprovider ./internal/projects/... ./apps/api ./apps/worker`

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Expected: PASS.

Commit: `feat(projects): generate evidence-backed matches`

### Task 5: Complete PDF/link export

**Files:**
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Modify: `internal/projects/postgres_repository.go`
- Modify: `internal/projects/worker.go`
- Modify: `internal/projects/export/`
- Modify: `apps/web/src/pages/ProjectsPage.tsx`
- Modify: `apps/web/src/lib/api.ts`
- Test: `internal/projects/worker_test.go`
- Test: `internal/projects/http_test.go`
- Test: `apps/web/src/pages/ProjectsPage.test.tsx`

**Step 1: Write failing export lifecycle tests**

Cover ownership, idempotent creation, queued/running/succeeded/failed states, actual PDF bytes, evidence parity, signed download, seven-day expiry, retry, and unavailable-download errors.

**Step 2: Implement asynchronous rendering and download**

Render the persisted report through the real PDF renderer, store it through the configured object store, and issue an expiring download link. Do not treat JSON as a PDF export.

**Step 3: Connect frontend export status and download**

Show progress and clear failure/retry states. Enable download only after the server confirms success.

**Step 4: Verify, commit, and push**

Run: `go test ./internal/projects/...`

Run: `cd apps/web && npm test -- ProjectsPage.test.tsx`

Expected: PASS.

Commit: `feat(projects): add durable pdf report export`

### Task 6: Add analytics and content administration

**Files:**
- Modify: `internal/projects/types.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Modify: `internal/projects/postgres_repository.go`
- Modify: `internal/projects/worker.go`
- Modify: `internal/platform/taskqueue/mux.go`
- Modify: `apps/worker/main.go`
- Test: `internal/projects/service_test.go`
- Test: `internal/projects/http_test.go`
- Test: `internal/projects/postgres_repository_test.go`
- Test: `internal/projects/worker_test.go`
- Modify: `docs/api/backend-api-contract.md`

**Step 1: Write failing analytics tests**

Cover PRD events, non-blocking ingestion, duplicate suppression, view/heat aggregates, privacy boundaries, and malformed-event rejection.

**Step 2: Write failing administration tests**

Cover import validation, batch status, publication gates, partial failure reporting, rollback, scheduled content production, reindex version increments, cache invalidation, permissions, and audit history.

**Step 3: Implement analytics and administration APIs/jobs**

Persist raw events and aggregates separately. Implement deterministic import/reindex task IDs and reversible import batches; expose stable status and validation errors.

**Step 4: Document, verify, commit, and push**

Run: `go test ./internal/projects/... ./internal/platform/taskqueue ./apps/worker`

Expected: PASS.

Commit: `feat(projects): add analytics and content operations`

### Task 7: Run the Project Market release gate

**Files:**
- Modify as failures require: Project Market implementation and tests only
- Create: `docs/testing/project-market-v14-acceptance.md`

**Step 1: Run all automated checks**

Run: `cd apps/web && npm test`

Run: `cd apps/web && npm run build`

Run: `cd apps/web && npm run lint`

Run: `go test ./...`

Run: `go vet ./...`

Run: `git diff --check`

Expected: all commands exit zero.

**Step 2: Run browser acceptance tests**

Verify desktop and mobile catalog search/filter/sort/pagination, direct routes, case evidence, favorite/compare persistence, upload/parse/removal, clarification refresh recovery, SSE reconnect, cancellation, retry, result evidence, and PDF download.

**Step 3: Record reproducible evidence**

Document the environment, seed data, commands, scenarios, results, known non-blocking limitations, and paths to screenshots/PDF artifacts in `docs/testing/project-market-v14-acceptance.md`.

**Step 4: Commit and push**

Commit: `test(projects): certify v1.4 acceptance workflow`

### Task 8: Begin Commercial Sandbox V1.2 frontend migration

**Files:**
- Follow: `docs/plans/2026-08-10-sandbox-v12-backend-design.md`
- Follow: `docs/plans/2026-08-10-sandbox-v12-backend-implementation.md`
- Modify: Sandbox frontend files identified during the post-gate audit

**Step 1: Confirm the Project Market gate remains green**

Run the Task 7 checks and do not start this task if any required check fails.

**Step 2: Audit current Sandbox frontend contracts**

Map legacy `/sandbox/sessions/*` calls to `/api/v1/sandbox-runs/*`, replace polling with replayable SSE where appropriate, and replace client JSON export with the server report/PDF workflow.

**Step 3: Write a focused Sandbox frontend execution plan**

Save the plan under `docs/plans/` with exact routes, files, tests, browser scenarios, and commit boundaries before modifying Sandbox application code.

