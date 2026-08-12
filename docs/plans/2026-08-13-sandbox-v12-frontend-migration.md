# Commercial Sandbox V1.2 Completion Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the legacy Commercial Sandbox web workflow with the PRD V1.2 run contract and close every P0/P1 user action with a real, tested backend workflow.

**Architecture:** The web application treats `/api/v1/sandbox-runs/*` as the only run source of truth and uses `/sandbox/new`, `/sandbox-runs/:id`, and `/sandbox-runs/:id/report` as canonical routes. The existing Go `sandbox` service remains the owner of run analytics and post-run follow-up, while explicit handoff endpoints call the existing tasks and growth application boundaries without duplicating their persistence rules.

**Tech Stack:** React 18, TypeScript, React Router, Vitest, Go 1.22, Gin, PostgreSQL 16, pgx, Redis, Asynq, authenticated SSE, server-rendered PDF.

---

### Task 1: Add the V1.2 TypeScript contract and authenticated transport

**Files:**
- Modify: `apps/web/src/lib/sandboxApi.ts`
- Modify: `apps/web/src/lib/sandboxApi.test.ts`

**Steps:**
1. Replace legacy session DTO tests with the exact `V2SandboxRun`, role, report, progress event, history, export, follow-up, task handoff, and growth handoff shapes.
2. Add failing transport tests for create, answer, role selection, start, stop, role retry, filtered history, rename, delete, report generation, authenticated SSE replay, PDF creation/download, follow-up, task batch creation, and growth handoff.
3. Implement a V1.2-only `sandboxApi` using `/api/v1/sandbox-runs/*`; keep no legacy DTO adapter.
4. Run `npm test -- src/lib/sandboxApi.test.ts` from `apps/web` and verify it passes.
5. Commit and push `feat(sandbox): add v1.2 web api contract`.

### Task 2: Make the four-step creation flow canonical

**Files:**
- Modify: `apps/web/src/App.tsx`
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Modify: `apps/web/src/pages/SandboxPage.test.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxHomeView.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxQuestionsView.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxSummaryView.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxRolesView.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxStartView.tsx`
- Modify: `apps/web/src/components/sandbox/SandboxStepper.tsx`
- Modify: `apps/web/src/components/sandbox/SandboxQuotaDialog.tsx`

**Steps:**
1. Add failing page tests for `/sandbox/new`, initial product parsing, bounded clarification, back navigation without data loss, all eight server roles, 3-8 selection, mandatory `skeptic`, confirmation data, and unlimited usage copy.
2. Mount `/sandbox/new` and canonical run/report routes; retain `/sandbox/start` only as a redirect-compatible alias.
3. Move the four steps into one run-aware controller driven by `V2SandboxRun.revision`, `next_questions`, and server role definitions.
4. Remove membership usage calls from normal sandbox execution. Keep the quota dialog shell disabled with `即将开放`, and never trigger it from a normal V1.2 response.
5. Run the focused page tests, build, and lint.
6. Commit and push `feat(sandbox): migrate creation flow to v1.2`.

### Task 3: Implement the real V1.2 run experience

**Files:**
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxRunView.tsx`
- Modify: `apps/web/src/components/sandbox/SandboxRoleCard.tsx`
- Modify: `apps/web/src/pages/SandboxPage.test.tsx`
- Modify: `apps/web/src/sandbox.css`

**Steps:**
1. Add failing tests for persisted role order, actual completed-role progress, stance labels, streamed content, failed role retry, stop preservation, terminal report action, and post-run follow-up.
2. Replace two-second legacy polling with authenticated SSE replay. Reconnect from the last event ID and refresh the authoritative run snapshot after lifecycle events.
3. Render each `run_role` independently using server status and output; never display provider/model names or fabricated interim progress.
4. Wire stop, one-role retry, report generation, and V2 follow-up to server APIs with visible pending/error states.
5. Verify desktop and 390px layout, then run focused tests, build, and lint.
6. Commit and push `feat(sandbox): stream isolated role runs`.

### Task 4: Close server-owned analytics and role follow-up

**Files:**
- Create: `migrations/000077_sandbox_v12_product_workflows.up.sql`
- Create: `migrations/000077_sandbox_v12_product_workflows.down.sql`
- Create: `internal/sandbox/v2_analytics.go`
- Create: `internal/sandbox/v2_followup.go`
- Modify: `internal/sandbox/types.go`
- Modify: `internal/sandbox/http.go`
- Modify: `internal/sandbox/v2_repository.go`
- Modify: `internal/sandbox/postgres_repository.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `docs/api/backend-api-contract.md`
- Test: `internal/sandbox/v2_analytics_test.go`
- Test: `internal/sandbox/v2_followup_test.go`
- Test: `internal/sandbox/http_test.go`

**Steps:**
1. Add failing tests for the PRD analytics whitelist/property schemas, idempotency, user ownership, safe visitor hashing, V2 role validation, frozen-output follow-up context, and no cross-role leakage.
2. Add user-scoped analytics and follow-up storage with indexes and reversible migration.
3. Implement `POST /sandbox/analytics`, `GET /sandbox-runs/:id/follow-ups`, and `POST /sandbox-runs/:id/follow-ups`.
4. Generate follow-up answers from the frozen selected role profile plus that role output only; reject unfinished/unknown roles.
5. Emit server lifecycle analytics for role start/done/failed using persisted role snapshot metadata.
6. Run focused unit and PostgreSQL repository tests.
7. Commit and push `feat(sandbox): add analytics and role followups`.

### Task 5: Add real report action handoffs

**Files:**
- Modify: `internal/tasks/types.go`
- Modify: `internal/tasks/service.go`
- Modify: `internal/tasks/http.go`
- Modify: `internal/tasks/postgres_repository.go`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `internal/sandbox/http.go`
- Modify: `docs/api/backend-api-contract.md`
- Test: `internal/tasks/http_test.go`
- Test: `internal/tasks/service_test.go`
- Test: `internal/sandbox/http_test.go`

**Steps:**
1. Add failing tests for transactional batch task creation, a maximum batch size, validation, ownership, idempotent sandbox source linkage, and preservation of `source_type=sandbox_session` plus `source_id=run_id`.
2. Add `POST /api/v1/tasks/batch` through the existing tasks application and repository boundaries.
3. Add a sandbox report task-handoff endpoint that converts selected report advice into validated task inputs and calls the tasks application.
4. Inspect and use the existing growth calculation draft contract. Add a sandbox growth-handoff endpoint that returns a canonical growth URL containing server-validated pricing and channel inputs, without creating synthetic growth results.
5. Run focused tests and the router contract tests.
6. Commit and push `feat(sandbox): connect report action handoffs`.

### Task 6: Complete report, history, and PDF workflows

**Files:**
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxReportView.tsx`
- Modify: `apps/web/src/pages/sandbox/SandboxHistoryView.tsx`
- Modify: `apps/web/src/pages/SandboxPage.test.tsx`
- Modify: `apps/web/src/sandbox.css`

**Steps:**
1. Add failing tests for model-simulation labels, feasibility basis, assumptions, missing roles, disagreements without averaging, role takeaways, selected task creation, growth handoff, authenticated server PDF download, rename, delete, filters, pagination, rerun confirmation, and missing-session dual exits.
2. Render the V1.2 report schema directly and expose only real action states.
3. Create/download the server PDF and derive its filename from the run name.
4. Implement server-driven history filters and pagination, plus rename/delete/rerun/export actions. Rerun copies the original product/context/roles into `/sandbox/new` and waits for explicit confirmation.
5. Ensure mobile controls wrap without overlap and all destructive actions require confirmation.
6. Run focused tests, build, and lint.
7. Commit and push `feat(sandbox): finish report and history workflows`.

### Task 7: Run the release acceptance gate

**Files:**
- Create: `docs/testing/sandbox-v12-acceptance.md`
- Modify only files required by failing checks.

**Steps:**
1. Run `npm test`, `npm run build`, and `npm run lint` in `apps/web`.
2. Run `go test ./...`, `go vet ./...`, and `git diff --check`.
3. Apply all migrations to a clean PostgreSQL database and run PostgreSQL-backed sandbox/tasks integration tests.
4. Run five consecutive authenticated sandbox runs and verify no quota response, one active-run enforcement, unique role session IDs, maximum three role calls in flight, role isolation, retry, stop, partial report, and SSE `Last-Event-ID` replay.
5. Verify `/sandbox`, `/sandbox/new`, run, report, and history in desktop and 390x844 browser viewports with no console errors or overlapping controls.
6. Generate and download a real report PDF, render it to page images, and visually inspect page content and wrapping.
7. Record exact commands, results, known non-blocking warnings, and deployment-environment prerequisites in the acceptance report.
8. Commit and push `test(sandbox): certify v1.2 acceptance workflow`.

