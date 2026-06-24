# Backend Architecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the remaining backend without letting module boundaries drift, starting with the hardest shared foundations: AI provider orchestration, async task execution, entitlements/credits, and provider evidence.

**Architecture:** Keep the current Go modular monolith. Add small internal packages with explicit interfaces, repositories, and HTTP handlers; do not introduce microservices, global service locators, or framework rewrites. Shared complexity belongs in `internal/ai`, `internal/jobs`, and membership/credit gates, while product modules (`analysis`, `projects`, `leads`, `crm`) own their domain models and API contracts.

**Tech Stack:** Go 1.22, Gin, pgx, PostgreSQL, Redis, Asynq, Vitest/React frontend integration later, table-driven Go tests, SQL migrations.

---

## Current Backend Shape

Existing foundations:
- API entrypoint: `apps/api/main.go`
- Worker entrypoint: `apps/worker/main.go`
- Routing: `internal/platform/httpserver/router.go`
- Config: `internal/platform/config/config.go`
- Auth: `internal/auth`
- Membership and credits: `internal/membership`
- MVP analysis: `internal/analysis`
- Asynq wrapper: `internal/platform/taskqueue`
- Migrations: `migrations`

Current architecture rule:
- Continue with package-per-domain inside one deployable API and one Worker.
- Each domain package should expose `Service`, `Repository`, `HTTPHandler`, and tests where useful.
- External providers must be behind interfaces and must never leak raw provider DTOs into HTTP responses.
- Database writes that affect money, credits, task status, or idempotency must live in repository transactions.

## Non-Negotiable Guardrails

- TDD for backend behavior: write service/repository/handler tests before implementation for every task below.
- No feature code should call an external provider directly from an HTTP handler.
- No background job may deduct credits outside a transaction or without idempotency.
- No AI response may be trusted until schema validation succeeds.
- No user-facing output may claim verified contactability unless the source explicitly provides that fact.
- No secrets in migrations, tests, logs, docs, or `.env.example`.
- No new large framework. Prefer standard library, current dependencies, and small local helpers.
- No unbounded retries. Every provider retry needs max attempts, timeout, and classified final error.
- No cross-domain imports that create cycles. Product packages can depend on shared packages; shared packages cannot depend on product packages.

## Architecture Decisions

### ADR-001: Keep Modular Monolith

**Status:** Accepted

**Context:** The app already deploys as API, Worker, PostgreSQL, Redis, and Web. The remaining backend has shared auth, membership, credits, AI calls, and async work, but current scale does not justify distributed services.

**Decision:** Keep one Go module and separate domains by `internal/<domain>` packages.

**Consequences:**
- Positive: simpler transactions, faster local development, easier refactors.
- Negative: package discipline matters; careless imports can blur boundaries.
- Alternative rejected: microservices, because provider orchestration and credits need strong consistency more than independent deployment.

### ADR-002: Introduce Shared `internal/ai`

**Status:** Accepted

**Context:** Free analysis, project matching, project report generation, and CRM follow-up copy all need model calls, JSON schema validation, prompt versions, retries, and observability.

**Decision:** Create `internal/ai` as the shared provider orchestration layer. Product modules submit typed AI jobs through interfaces and receive validated typed results.

**Consequences:**
- Positive: one place for model config, schema retry, prompt versioning, run logs, latency and cost tracking.
- Negative: first implementation takes longer.
- Alternative rejected: each module owns its own provider, because that duplicates retry, parsing, logging, and error handling.

### ADR-003: Async Jobs Own Long-Running Work

**Status:** Accepted

**Context:** Lead search, web evidence enrichment, and long AI report generation can exceed request timeouts and need progress status.

**Decision:** Use Asynq through a small shared `internal/jobs` layer. HTTP creates domain records and enqueues jobs; Worker executes jobs and updates domain state.

**Consequences:**
- Positive: explicit task lifecycle, retry control, cancellation path, progress polling.
- Negative: requires careful idempotency and status transitions.
- Alternative rejected: do everything in HTTP requests, because provider calls are slow and failure-prone.

## Dependency Direction

Allowed:
- `apps/api` imports domain packages and platform packages.
- `apps/worker` imports domain worker handlers and platform packages.
- `internal/analysis`, `internal/projects`, `internal/leads`, `internal/crm` import `internal/ai`, `internal/jobs`, `internal/membership` interfaces when needed.
- `internal/ai` imports only platform/config/logging-safe helpers and stdlib.
- `internal/jobs` imports Asynq and stdlib, not product packages.

Forbidden:
- `internal/ai` importing `analysis`, `projects`, `leads`, or `crm`.
- HTTP handlers importing provider implementations directly.
- Repositories importing HTTP request types.
- Worker handlers reimplementing business rules that belong in services.

## Target Modules

### `internal/ai`

Responsibilities:
- Provider interface and provider selection.
- Prompt version metadata.
- JSON schema validation and one structured retry.
- AI run persistence and status tracking.
- Error classification: invalid input, provider timeout, rate limited, provider unavailable, invalid model JSON.
- Development provider for deterministic tests.

Not responsible:
- Product-specific prompts as giant strings mixed into provider code.
- HTTP routes.
- Business entitlement decisions.

### `internal/projects`

Responsibilities:
- Project-market AI matching sessions.
- Follow-up questions.
- Match results, project details, comparisons, exports.
- Persistence for history and saved projects.

Depends on:
- `internal/ai` for model calls.
- `internal/membership` for future entitlement checks.

### `internal/leads`

Responsibilities:
- Lead search task lifecycle.
- Provider-neutral lead/evidence model.
- Deduplication and result persistence.
- Credit charging and refund on system failure.

Depends on:
- `internal/jobs` for queueing.
- `internal/membership` for credit operations.
- Provider interfaces under `internal/leads`.

### `internal/crm`

Responsibilities:
- Import leads into customers.
- Customer list/filter/detail.
- Stage changes and follow-up records.
- Next-follow-up filtering.
- Later: AI follow-up copy through `internal/ai`.

Depends on:
- `internal/ai` only for AI copy, not for core CRM CRUD.

## Phase Order: Hard to Easy

1. AI Core.
2. Analysis migration to AI Core.
3. Project-market backend.
4. Jobs Core.
5. Lead provider and async lead tasks.
6. CRM core.
7. Content/config/admin CRUD.
8. Observability, security hardening, E2E.

This order is intentional. Do not start CRM or admin CRUD before AI Core and Jobs Core are in place unless explicitly asked.

## Implementation Progress

Updated 2026-06-23:
- Completed Task 1: `ai_runs` migration, `internal/ai` run types, Postgres repository, and repository tests.
- Completed Task 2: `internal/ai` provider interface, JSON generation service, one repair retry, error classification, development provider, and service tests.
- Completed Task 3: AI config fields, production config validation, `.env.example` entries, and API construction of AI service.
- Completed Task 4: `internal/analysis` now calls AI Core through a narrow JSON generator interface; development analysis output remains deterministic; invalid AI payloads map to safe analysis errors.
- Completed Task 5: analysis session history and detail restore APIs, user-scoped repository reads, and 404 behavior for missing/not-owned sessions.
- Completed Task 6: project-market match backend skeleton, migrations, AI-backed match creation, user-scoped history/detail, idempotent favorites, protected API routes, and development AI responses for project matching.
- Completed Task 7: shared jobs core with idempotency envelope, Asynq enqueue wrapper, registered task type validation, default registry, and worker mux wiring.
- Completed Task 8: lead task lifecycle with queued/running/succeeded/failed/refunded states, idempotent task creation, queue enqueue, development lead provider, repository/http/worker entry points, result persistence, and failure refund behavior through a credit ledger interface.
- Completed Task 9: Tianyancha and Serper HTTP provider adapters with provider-neutral mapping, timeout/rate-limit/quota classification, config/env wiring, and API/worker lead provider selection.
- Completed Task 10: CRM customer import, stage/activity tracking, follow-up scheduling, due/overdue queries, user-scoped repository access, migrations, protected HTTP routes, and API wiring.
- Completed Task 11: AI-generated CRM follow-up copy using owned customer context, schema validation, safe invalid-AI errors, protected HTTP route, and API wiring through AI Core.

Latest verification:
```bash
go test ./internal/ai
go test ./internal/platform/config
go test ./internal/analysis ./internal/ai
go test ./internal/analysis
go test ./internal/ai ./internal/projects ./internal/platform/httpserver ./internal/platform/migrations
go test ./internal/jobs ./internal/platform/taskqueue
go test ./internal/leads ./internal/jobs ./internal/membership ./internal/platform/httpserver ./internal/platform/migrations
go test ./...
```

All commands passed.

---

## Task 1: Add AI Core Data Model

**Files:**
- Create: `migrations/000006_ai_runs.up.sql`
- Create: `migrations/000006_ai_runs.down.sql`
- Create: `internal/ai/types.go`
- Create: `internal/ai/postgres_repository.go`
- Create: `internal/ai/postgres_repository_test.go`

**Step 1: Write repository tests**

Create tests that prove:
- `CreateRun` inserts a pending run with user id, feature, prompt version, model, request JSON.
- `CompleteRun` stores response JSON, latency, token counts, and status.
- `FailRun` stores classified error code and safe message.
- JSON fields round-trip as JSONB.

Use `pgxmock` as existing repository tests do.

**Step 2: Add migration**

Create `ai_runs`:
- `id BIGSERIAL PRIMARY KEY`
- `user_id BIGINT REFERENCES users(id) ON DELETE SET NULL`
- `feature TEXT NOT NULL`
- `prompt_version TEXT NOT NULL`
- `provider TEXT NOT NULL`
- `model TEXT NOT NULL`
- `status TEXT NOT NULL`
- `request JSONB NOT NULL DEFAULT '{}'::JSONB`
- `response JSONB NOT NULL DEFAULT '{}'::JSONB`
- `error_code TEXT NOT NULL DEFAULT ''`
- `error_message TEXT NOT NULL DEFAULT ''`
- `input_tokens INTEGER NOT NULL DEFAULT 0`
- `output_tokens INTEGER NOT NULL DEFAULT 0`
- `latency_ms INTEGER NOT NULL DEFAULT 0`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Indexes:
- `(user_id, created_at DESC)`
- `(feature, created_at DESC)`
- `(status, created_at DESC)`

**Step 3: Implement repository**

Expose:
- `CreateRun(ctx context.Context, run Run) (Run, error)`
- `CompleteRun(ctx context.Context, id int64, result RunResult) error`
- `FailRun(ctx context.Context, id int64, failure RunFailure) error`

**Step 4: Verify**

Run:
```bash
go test ./internal/ai
go test ./internal/platform/migrations
```

Expected:
- New AI tests pass.
- Migration tests pass.

## Task 2: Add AI Provider Interface and JSON Validation

**Files:**
- Create: `internal/ai/provider.go`
- Create: `internal/ai/service.go`
- Create: `internal/ai/service_test.go`
- Create: `internal/ai/development_provider.go`

**Step 1: Write service tests**

Cover:
- Service creates an `ai_runs` record before provider call.
- Valid JSON response returns typed payload and marks run completed.
- Invalid JSON triggers exactly one retry with a repair instruction.
- Second invalid JSON marks run failed with `invalid_model_json`.
- Provider timeout/rate-limit errors are classified.
- Request context cancellation stops the call and marks run failed.

**Step 2: Define interfaces**

```go
type Provider interface {
    Generate(ctx context.Context, request ProviderRequest) (ProviderResponse, error)
}

type Repository interface {
    CreateRun(ctx context.Context, run Run) (Run, error)
    CompleteRun(ctx context.Context, id int64, result RunResult) error
    FailRun(ctx context.Context, id int64, failure RunFailure) error
}
```

`ProviderRequest` must include:
- feature
- prompt version
- system prompt
- user prompt
- expected schema name
- retry attempt

**Step 3: Implement schema validator**

Keep it simple for now:
- Product modules pass a `func([]byte) error` validator.
- `internal/ai` does not need a full JSON Schema engine yet.
- Use `encoding/json.Decoder` with `DisallowUnknownFields` in product validators when strictness matters.

**Step 4: Verify**

Run:
```bash
go test ./internal/ai
```

Expected:
- All AI service tests pass.

## Task 3: Wire AI Config

**Files:**
- Modify: `internal/platform/config/config.go`
- Modify: `internal/platform/config/config_test.go`
- Modify: `.env.example`
- Modify: `apps/api/main.go`

**Step 1: Write config tests**

Cover:
- Defaults use development AI provider.
- Production requires explicit AI provider configuration.
- API key is required only for real providers.

**Step 2: Add config fields**

Add:
- `AIProvider`
- `AIModel`
- `AIAPIKey`
- `AIBaseURL`
- `AITimeoutSeconds`

Default:
- development: `AIProvider=development`
- production: no silent development fallback.

**Step 3: Wire in API**

In `apps/api/main.go`:
- Construct `aiRepository := ai.NewPostgresRepository(db)`.
- Construct provider from config.
- Pass `aiService` into downstream modules.

At this task, only construct it; do not migrate all analysis code yet.

**Step 4: Verify**

Run:
```bash
go test ./internal/platform/config
go test ./...
```

## Task 4: Migrate Analysis to AI Core

**Files:**
- Modify: `internal/analysis/service.go`
- Modify: `internal/analysis/service_test.go`
- Modify: `internal/analysis/development_provider.go`
- Modify: `apps/api/main.go`

**Step 1: Write failing tests**

Add tests for:
- Direction analysis calls AI Core when input is complete.
- Invalid AI payload returns safe internal error and creates failed AI run.
- Direction card validator requires all 7 card fields.
- Existing missing-question behavior remains unchanged.

**Step 2: Replace direct provider dependency**

Current:
- `analysis.Provider.GenerateDirection`

Target:
- `analysis` owns prompt construction and validator.
- `analysis` calls `ai.Service.GenerateJSON`.
- `analysis` maps typed AI result to `DirectionResult`.

**Step 3: Keep development behavior deterministic**

The development provider should return stable JSON for tests and local demos.

**Step 4: Verify**

Run:
```bash
go test ./internal/analysis ./internal/ai
go test ./...
```

## Task 5: Add Analysis Session Read APIs

**Files:**
- Modify: `internal/analysis/service.go`
- Modify: `internal/analysis/postgres_repository.go`
- Modify: `internal/analysis/http.go`
- Modify: `internal/analysis/*_test.go`

**Step 1: Write tests**

Cover:
- `GET /api/v1/analysis/sessions` returns only current user's sessions.
- `GET /api/v1/analysis/sessions/:id` returns 404 for another user's session.
- Stored report can be restored after refresh.

**Step 2: Repository additions**

Add:
- `ListSessions(ctx, userID, limit)`
- `GetSession(ctx, userID, id)`

**Step 3: HTTP routes**

Add protected routes:
- `GET /analysis/sessions`
- `GET /analysis/sessions/:id`

**Step 4: Verify**

Run:
```bash
go test ./internal/analysis
go test ./...
```

## Task 6: Add Project-Market Backend Skeleton

**Files:**
- Create: `internal/projects/types.go`
- Create: `internal/projects/service.go`
- Create: `internal/projects/service_test.go`
- Create: `internal/projects/http.go`
- Create: `internal/projects/http_test.go`
- Create: `internal/projects/postgres_repository.go`
- Create: `internal/projects/postgres_repository_test.go`
- Create: `migrations/000007_projects.up.sql`
- Create: `migrations/000007_projects.down.sql`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `apps/api/main.go`

**Step 1: Define MVP API**

Routes:
- `POST /api/v1/projects/matches`
- `GET /api/v1/projects/matches`
- `GET /api/v1/projects/matches/:id`
- `POST /api/v1/projects/matches/:id/favorite`

Do not implement export files yet. Return JSON that can feed the existing static UI.

**Step 2: Tests first**

Cover:
- Short/underspecified request returns follow-up questions.
- Complete request returns ranked match cards.
- History is scoped to current user.
- Favorite is idempotent.

**Step 3: Data model**

Create:
- `project_match_sessions`
- `project_match_favorites`

Store AI raw structured result in JSONB, but expose typed DTOs.

**Step 4: AI integration**

Use `internal/ai` with prompt version like `project_match_v1`.

**Step 5: Verify**

Run:
```bash
go test ./internal/projects ./internal/ai
go test ./...
```

## Task 7: Add Jobs Core

**Files:**
- Create: `internal/jobs/types.go`
- Create: `internal/jobs/client.go`
- Create: `internal/jobs/client_test.go`
- Modify: `internal/platform/taskqueue/mux.go`
- Modify: `apps/worker/main.go`

**Step 1: Write tests**

Cover:
- Job payloads include idempotency key.
- Enqueue options set max retry and timeout.
- Unknown job type is rejected by local registry.

**Step 2: Implement wrapper**

Add:
- `Client.Enqueue(ctx, Job) error`
- `Register(mux, Handler)`
- stable task type constants.

Do not put lead-specific logic here.

**Step 3: Verify**

Run:
```bash
go test ./internal/jobs ./internal/platform/taskqueue
go test ./...
```

## Task 8: Add Lead Task Lifecycle

**Files:**
- Create: `internal/leads/types.go`
- Create: `internal/leads/service.go`
- Create: `internal/leads/service_test.go`
- Create: `internal/leads/http.go`
- Create: `internal/leads/postgres_repository.go`
- Create: `internal/leads/worker.go`
- Create: `internal/leads/*_test.go`
- Create: `migrations/000008_leads.up.sql`
- Create: `migrations/000008_leads.down.sql`
- Modify: `apps/api/main.go`
- Modify: `apps/worker/main.go`
- Modify: `internal/platform/httpserver/router.go`

**Step 1: Define lifecycle**

Statuses:
- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`
- `refunded`

Transitions must be explicit and tested.

**Step 2: Tests first**

Cover:
- Creating a task deducts credits and enqueues exactly once.
- Duplicate idempotency key does not deduct twice.
- Worker success stores leads and evidence.
- Worker system failure refunds credits.
- Provider quota failure is not retried forever.
- User can list own tasks and not others.

**Step 3: Provider interfaces**

Define:
- `LeadProvider`
- `SearchProvider`

Use development providers first. Real Tianyancha/Serper adapters come after lifecycle works.

**Step 4: Verify**

Run:
```bash
go test ./internal/leads ./internal/jobs ./internal/membership
go test ./...
```

## Task 9: Add Real Provider Adapters

**Files:**
- Create: `internal/leads/tianyancha_provider.go`
- Create: `internal/leads/search_provider.go`
- Create: `internal/leads/provider_errors.go`
- Create: `internal/leads/*provider*_test.go`
- Modify: `internal/platform/config/config.go`
- Modify: `.env.example`

**Step 1: Tests with fake HTTP server**

Use `httptest.Server`, not real external calls.

Cover:
- success payload mapping
- quota/rate-limit classification
- timeout classification
- missing optional fields
- evidence source URLs preserved

**Step 2: Implement adapters**

Adapters return provider-neutral domain structs.

**Step 3: Verify**

Run:
```bash
go test ./internal/leads
go test ./...
```

## Task 10: Add CRM Core

**Files:**
- Create: `internal/crm/types.go`
- Create: `internal/crm/service.go`
- Create: `internal/crm/http.go`
- Create: `internal/crm/postgres_repository.go`
- Create: `internal/crm/*_test.go`
- Create: `migrations/000009_crm.up.sql`
- Create: `migrations/000009_crm.down.sql`
- Modify: `apps/api/main.go`
- Modify: `internal/platform/httpserver/router.go`

**Step 1: Tests first**

Cover:
- Importing same lead twice is idempotent.
- Stage update creates activity record.
- Follow-up record updates next follow-up date.
- Due and overdue filters work.
- User cannot access another user's customers.

**Step 2: Data model**

Create:
- `crm_customers`
- `crm_activities`
- `crm_followups`

Use deterministic unique keys for imported leads.

**Step 3: Verify**

Run:
```bash
go test ./internal/crm
go test ./...
```

## Task 11: AI Follow-Up Copy

**Files:**
- Modify: `internal/crm/service.go`
- Modify: `internal/crm/http.go`
- Modify: `internal/crm/*_test.go`

**Step 1: Tests first**

Cover:
- Generates copy from customer context through `internal/ai`.
- Does not include private fields unless the requesting user owns the customer.
- Invalid AI JSON returns a safe error.

**Step 2: Route**

Add:
- `POST /api/v1/crm/customers/:id/follow-up-copy`

**Step 3: Verify**

Run:
```bash
go test ./internal/crm ./internal/ai
go test ./...
```

## Task 12: Content and Config Admin CRUD

**Files:**
- Create: `internal/content`
- Create: `migrations/000010_content_config.up.sql`
- Create: `migrations/000010_content_config.down.sql`
- Modify: `apps/api/main.go`
- Modify: `internal/platform/httpserver/router.go`

**Step 1: Tests first**

Cover:
- Public content list/read endpoints.
- Admin-only write endpoints when admin role exists.
- Empty states return empty arrays, not errors.

**Step 2: Keep scope narrow**

Support:
- articles
- tools
- community config
- brand metrics/cases

Do not build a generic CMS framework.

## Task 13: Security and Observability Pass

**Files:**
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/config/config.go`
- Add focused tests in relevant packages.

**Checklist:**
- Request logging with PII redaction.
- Provider errors logged without raw keys or raw private payloads.
- Rate limits for expensive endpoints.
- CORS config explicit for production.
- JWT secret and provider keys required in production.
- AI and provider run failure metrics or structured log fields.
- Slow query candidates have indexes.

**Verify:**
```bash
go test ./...
npm test --prefix apps/web
npm run build --prefix apps/web
npm run lint --prefix apps/web
```

## API Error Contract

All new JSON API errors should use:

```json
{
  "error": "machine_readable_code",
  "message": "safe human-readable message"
}
```

Rules:
- Do not expose provider raw messages if they may contain request data.
- Preserve HTTP status semantics:
  - `400` invalid user input
  - `401` unauthenticated
  - `403` entitlement/ownership failure
  - `404` not found or not owned
  - `409` idempotency/conflict
  - `429` quota/rate limit
  - `500` internal
  - `502` provider unavailable
  - `504` provider timeout

## Database Rules

- Every new table needs `created_at` and usually `updated_at`.
- Every user-owned table needs `user_id` and a user-scoped index.
- Every async task table needs explicit status and retry/error fields.
- Every idempotent operation needs a unique key.
- JSONB is allowed for AI/provider payloads, but HTTP responses must use typed DTOs.
- Down migrations must reverse schema changes.

## Test Commands by Layer

Backend fast loop:
```bash
go test ./internal/ai ./internal/analysis ./internal/projects
```

Backend full:
```bash
go test ./...
```

Frontend after API contract changes:
```bash
npm test --prefix apps/web
npm run build --prefix apps/web
npm run lint --prefix apps/web
```

## Completion Criteria Before Starting Frontend Integration

- AI Core exists and analysis uses it.
- Project match endpoints can return the same data shape the static UI currently displays.
- Lead tasks can be created, processed by Worker, and inspected through API.
- Credit deduction/refund behavior is tested.
- CRM can import leads idempotently.
- `go test ./...` passes.

## Final Notes for Implementers

- Prefer small commits per task.
- Do not mix schema, provider, API, and frontend work in one commit.
- If a task reveals a missing abstraction, add it only after a test demonstrates duplicated or risky behavior.
- If a real provider is blocked by credentials, keep the interface, fake provider tests, and config validation complete; mark only the real adapter as blocked.
- Before marking any phase done, update this plan with actual implemented files and verification output.

## Implementation Record

Updated: 2026-06-24

Implemented:
- Tasks 1-3: AI run persistence, AI provider interface, JSON validation/retry/failure recording, and AI config fields. API now fails fast for unsupported AI providers instead of silently using development AI.
- Tasks 4-5: Analysis now calls AI Core for completed direction requests and exposes user-scoped session list/read APIs.
- Task 6: Project-market match APIs, match session persistence, favorites, and AI-backed ranked match results.
- Task 7: Jobs core wrapper and taskqueue registry/mux/server helpers.
- Tasks 8-9: Lead task lifecycle, worker processing, credit charge/refund behavior, Tianyancha lead provider, and Serper evidence provider.
- Tasks 10-11: CRM customers, activities, follow-ups, due-customer queries, lead import idempotency, and AI follow-up copy.
- Task 12: Content/config admin CRUD with public read endpoints, admin role checks, and content/config migrations.
- Task 13: Request logging with redaction, explicit CORS config, expensive endpoint rate limits, production secret/provider validation, and safe AI provider failure logging.

Verification run after implementation:
```bash
go test ./...
npm test --prefix apps/web
npm run build --prefix apps/web
npm run lint --prefix apps/web
```

Result:
- Backend Go tests passed.
- Frontend tests passed: 43 files, 108 tests.
- Frontend build passed with a Vite chunk-size warning.
- Frontend lint passed.

Known remaining integration note:
- `OPCV2_AI_PROVIDER=development` is the only currently wired AI provider. Non-development values fail fast until a real AI provider adapter is added.
