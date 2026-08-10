# Commercial Sandbox V1.2 Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade the existing commercial sandbox backend to PRD V1.2 with isolated multi-role runs, replayable progress, synthesized reports, and compatible legacy APIs.

**Architecture:** Extend `sandbox_sessions` as the run aggregate and add normalized role, event, report, and export tables. Add V2 service and HTTP contracts alongside the existing `/sandbox/sessions` surface, then evolve the current Asynq worker into a bounded role orchestrator.

**Tech Stack:** Go 1.22, Gin, PostgreSQL 16, pgx, Redis, Asynq, JSON model providers, SSE.

---

### Task 1: Add the V1.2 persistence foundation

**Files:**
- Create: `migrations/000067_sandbox_v12_foundation.up.sql`
- Create: `migrations/000067_sandbox_v12_foundation.down.sql`
- Modify: `internal/sandbox/types.go`
- Test: `internal/sandbox/service_test.go`

**Steps:**
1. Add failing tests for the eight role codes, seven defaults, required `skeptic`, and V2 statuses.
2. Add the migration extending `sandbox_sessions` and creating role configs, run roles, events, reports, and exports.
3. Add V2 DTOs and role definitions without changing legacy DTO JSON.
4. Run `go test ./internal/sandbox` and commit.

### Task 2: Implement V2 draft, clarification, and role selection

**Files:**
- Create: `internal/sandbox/v2_workflow.go`
- Modify: `internal/sandbox/postgres_repository.go`
- Modify: `internal/sandbox/http.go`
- Test: `internal/sandbox/v2_workflow_test.go`
- Test: `internal/sandbox/http_test.go`

**Steps:**
1. Add failing tests for structured draft creation, up to three questions per round, completeness `>= 0.8`, three-round stop, skip assumptions, and role validation.
2. Implement repository methods scoped by user ownership.
3. Register `POST /sandbox-runs`, `POST /sandbox-runs/:id/answer`, `POST /sandbox-runs/:id/roles`, and `GET /sandbox-runs/:id`.
4. Run focused tests and commit.

### Task 3: Implement idempotent start and execution snapshots

**Files:**
- Create: `internal/sandbox/v2_execution.go`
- Modify: `internal/sandbox/postgres_repository.go`
- Modify: `internal/sandbox/http.go`
- Test: `internal/sandbox/v2_execution_test.go`

**Steps:**
1. Add failing tests for one active run per user, repeated start, frozen input hash, distinct role session IDs, usage count without blocking, and deterministic queue jobs.
2. Prepare role snapshots transactionally and enqueue one orchestration job.
3. Keep the old `RunSession` path unchanged for legacy clients.
4. Run focused tests and commit.

### Task 4: Add isolated role execution and persisted SSE events

**Files:**
- Create: `internal/sandbox/v2_orchestrator.go`
- Create: `internal/sandbox/v2_progress.go`
- Modify: `internal/sandbox/worker.go`
- Modify: `internal/sandbox/http.go`
- Test: `internal/sandbox/v2_orchestrator_test.go`
- Test: `internal/sandbox/http_test.go`

**Steps:**
1. Add failing tests proving role prompts do not contain other role output and concurrency never exceeds three.
2. Validate role JSON and required dimension coverage, retry once, and persist failure without aborting other roles.
3. Persist ordered events and implement SSE replay using `Last-Event-ID`.
4. Implement stop with partial preservation.
5. Run focused tests and commit.

### Task 5: Add report synthesis, history, and export

**Files:**
- Create: `internal/sandbox/v2_report.go`
- Modify: `internal/sandbox/postgres_repository.go`
- Modify: `internal/sandbox/http.go`
- Modify: `internal/ai/development_provider.go`
- Modify: `docs/api/backend-api-contract.md`
- Test: `internal/sandbox/v2_report_test.go`

**Steps:**
1. Add failing tests for a distinct report session, disagreements, missing roles, assumptions, model-generated labels, and report idempotency.
2. Add `POST|GET /sandbox-runs/:id/report`, history list/update/delete, and seven-day export metadata.
3. Add deterministic development-provider role and report responses.
4. Run focused tests and commit.

### Task 6: Final verification

**Files:**
- Modify only files required by failing checks.

**Steps:**
1. Run `go test ./...`.
2. Run `make lint`.
3. Run `make build`.
4. Verify migrations are sequential and reversible.
5. Commit any final fixes and push `feature/bootstrap`.
