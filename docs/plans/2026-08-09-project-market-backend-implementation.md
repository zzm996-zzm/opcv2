# Project Market Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Evolve the existing project module into the evidence-backed project market specified by PRD V1.4 without breaking the current frontend.

**Architecture:** Keep the Go/Gin modular monolith and BIGINT identity model. Extend current project tables, split public and protected routes, add normalized evidence tables, and run matching/research through durable Asynq jobs with persisted progress.

**Tech Stack:** Go 1.22, Gin, PostgreSQL 16, Redis, Asynq, pgx, pgvector in a later provider-dependent slice, React/Vite API consumer.

---

### Task 1: Establish configuration and public route boundaries

**Files:**
- Modify: `internal/platform/config/config.go`
- Modify: `internal/platform/config/config_test.go`
- Modify: `.env.example`
- Modify: `.env.server.example`
- Modify: `internal/projects/http.go`
- Modify: `internal/platform/httpserver/router.go`
- Test: `internal/projects/http_test.go`
- Test: `internal/platform/httpserver/health_test.go`

**Step 1: Write failing tests**

Test that config defaults `FeaturePaywallEnabled` to false and reads `OPCV2_FEATURE_PAYWALL_ENABLED`. Test that catalog, detail, case, dictionary, and config reads work without authentication while match, favorite, comparison, and export commands remain protected.

**Step 2: Run tests to verify failure**

Run: `go test ./internal/platform/config ./internal/projects ./internal/platform/httpserver`

Expected: FAIL because the feature flag and split route registration do not exist.

**Step 3: Implement minimal configuration and registration**

Add `FeaturePaywallEnabled bool`, `RegisterPublic`, and `RegisterProtected`. Keep `Register` as a compatibility wrapper for unit tests until callers migrate. Register public project routes before the protected group.

**Step 4: Run tests and commit**

Run: `go test ./internal/platform/config ./internal/projects ./internal/platform/httpserver`

Expected: PASS.

Commit: `feat: establish project market route boundaries`

### Task 2: Add executable project market foundation migration

**Files:**
- Create: `migrations/000061_project_market_foundation.up.sql`
- Create: `migrations/000061_project_market_foundation.down.sql`
- Modify: `internal/platform/migrations/migrations_test.go` if migration assertions require extension

**Step 1: Write the migration contract**

Use BIGSERIAL/BIGINT throughout. Create dictionaries, startup failures and sources, rebuild plans/localizations, research jobs/sources, opportunity publication records, source mappings, project favorites/views/compare items, import batches, and corrections. Extend `project_opportunities`, `project_cases`, and `project_match_sessions` without duplicating them.

**Step 2: Validate SQL against a real PostgreSQL instance**

Run: `go test ./internal/platform/migrations`

Run when Docker PostgreSQL is available: `docker compose up -d postgres && go test ./internal/platform/migrations -run Integration -count=1`

Expected: migration applies once, is idempotent through the migration runner, and down migration removes only the new foundation.

**Step 3: Commit**

Commit: `feat: add project market foundation schema`

### Task 3: Add config, dictionaries, home, list, and detail contracts

**Files:**
- Modify: `internal/projects/types.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Modify: `internal/projects/postgres_repository.go`
- Test: `internal/projects/catalog_service_test.go`
- Test: `internal/projects/http_test.go`
- Test: `internal/projects/postgres_repository_test.go`
- Modify: `docs/api/backend-api-contract.md`

**Step 1: Write failing service and HTTP tests**

Cover `GET /config`, `GET /dicts`, `GET /projects/home`, paged `GET /projects`, and `GET /projects/:id`. Assert `locked_blocks` is empty and `is_unlocked` is true when paywall is disabled. Assert only valid HTTP(S) source URLs are returned as verified.

**Step 2: Implement types and repository queries**

Add explicit list/detail DTOs instead of returning database models. Support keyword, category, track, budget, difficulty, resource, sort, featured, page, and page size. Use stable pagination and published-only predicates.

**Step 3: Add compatibility behavior**

Keep `/projects/opportunities` and `/projects/opportunities/:slug` backed by existing DTOs while new consumers use `/projects` and `/projects/:id`.

**Step 4: Verify and commit**

Run: `go test ./internal/projects ./internal/platform/httpserver`

Expected: PASS.

Commit: `feat: add public project market catalog APIs`

### Task 4: Normalize case evidence and publishing rules

**Files:**
- Modify: `internal/projects/types.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/postgres_repository.go`
- Modify: `internal/projects/http.go`
- Test: `internal/projects/case_service_test.go`
- Test: `internal/projects/postgres_repository_test.go`

**Step 1: Write failing publication tests**

Verify a case cannot be exposed as verified without one primary/authority source or two independent reliable sources. Verify conflicts remain `needs_review`. Verify claim fields map to individual source IDs.

**Step 2: Implement case list/detail DTOs and source policy**

Return facts, analyses, and sources separately. Preserve the old `/projects/cases` payload for compatibility while adding `/project-cases` and `/project-cases/:id`.

**Step 3: Verify and commit**

Run: `go test ./internal/projects`

Expected: PASS.

Commit: `feat: add evidence-backed project cases`

### Task 5: Implement persistent match clarification state

**Files:**
- Modify: `internal/projects/types.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/postgres_repository.go`
- Modify: `internal/projects/http.go`
- Test: `internal/projects/service_test.go`
- Test: `internal/projects/postgres_repository_test.go`
- Test: `internal/projects/http_test.go`

**Step 1: Write failing state-machine tests**

Cover one-line input, profile merge, field sources, up to three questions per round, eight total questions, three rounds, completeness threshold, skipped assumptions, ownership, and idempotent commands.

**Step 2: Replace heuristic completion with structured analysis**

Persist `input_snapshot`, `parsed_profile`, `field_sources`, `analysis_summary`, `completeness`, `rounds`, questions, answers, and assumptions. Do not persist model chain-of-thought.

**Step 3: Verify and commit**

Run: `go test ./internal/projects`

Expected: PASS.

Commit: `feat: add project match clarification workflow`

### Task 6: Add asynchronous generation, progress, and cancellation

**Files:**
- Modify: `internal/jobs/types.go`
- Modify: `internal/platform/taskqueue/mux.go`
- Modify: `apps/worker/main.go`
- Create: `internal/projects/worker.go`
- Create: `internal/projects/progress.go`
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Test: `internal/projects/worker_test.go`
- Test: `internal/projects/http_test.go`
- Test: `internal/platform/taskqueue/mux_test.go`

**Step 1: Write failing queue and state tests**

Test deterministic task IDs, duplicate generate calls, legal transitions, persisted progress, SSE replay, reconnect, partial result, and cancellation before each expensive stage.

**Step 2: Implement the worker shell**

Add the match generation job type and orchestration interfaces for retrieval, research, and generation. Initially use the catalog retriever and existing model generator behind those interfaces.

**Step 3: Verify and commit**

Run: `go test ./internal/projects ./internal/platform/taskqueue ./apps/worker`

Expected: PASS.

Commit: `feat: run project matching asynchronously`

### Task 7: Add files, retrieval, and web evidence providers

**Files:**
- Create: `internal/projects/files/`
- Create: `internal/projects/retrieval/`
- Create: `internal/projects/research/`
- Create: provider-specific migrations after model selection
- Modify: `internal/platform/config/config.go`
- Modify: `apps/api/main.go`
- Modify: `apps/worker/main.go`

**Step 1: Define provider contracts and development fakes**

Add object storage, scanner, parser/OCR, embedding, search, fetch, canonicalization, and evidence extraction interfaces. Development implementations must be explicit and production startup must reject unsafe development providers.

**Step 2: Add ownership and security tests**

Test limits, MIME/header mismatch, expiry, prompt-injection text treatment, user isolation, canonical URL deduplication, quality threshold, and network failure degradation.

**Step 3: Add pgvector only after embedding selection**

Create the vector migration using the selected dimension and update the PostgreSQL image to include the extension. Reindex increments the knowledge-base version and invalidates retrieval cache keys.

**Step 4: Verify and commit in provider-sized changes**

Run: `go test ./...`

Expected: PASS.

### Task 8: Complete export, comparison, diagnostics, and observability

**Files:**
- Modify: `internal/projects/service.go`
- Modify: `internal/projects/http.go`
- Modify: `internal/projects/postgres_repository.go`
- Modify: `internal/membership/service.go`
- Modify: `docs/api/backend-api-contract.md`
- Test: project service, repository, HTTP, and worker tests

**Step 1: Add failing acceptance tests**

Cover five-item comparison, async PDF/link export with seven-day expiry, diagnosis output, usage count without blocking, view/heat aggregation, evidence audit records, and all PRD error states.

**Step 2: Implement the remaining workflows**

Keep paywall responses unreachable while the flag is false. Preserve future locked block shapes without storing locked body content in public responses.

**Step 3: Run final verification**

Run: `go test ./...`

Run: `make lint`

Run: `make build`

Expected: all commands pass.

**Step 4: Commit and push**

Commit: `feat: complete project market backend workflows`
