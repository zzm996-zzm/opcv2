# Sandbox Design Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver the nine reference sandbox states as one API-backed production flow with desktop visual parity and responsive behavior.

**Architecture:** Extend the existing sandbox session with JSONB context and settings while preserving the current queue, quota, role, message, and report contracts. Replace the monolithic sandbox presentation with isolated `sb-*` components and styles so the reference geometry can be verified independently from legacy global CSS.

**Tech Stack:** Go, Gin, PostgreSQL, React, TypeScript, React Router, CSS, Vitest, Go tests, Playwright/in-app browser.

---

### Task 1: Lock the backend contract

**Files:**
- Modify: `internal/sandbox/types.go`
- Modify: `internal/sandbox/service.go`
- Modify: `internal/sandbox/http.go`
- Modify: `internal/sandbox/postgres_repository.go`
- Create: `migrations/000060_sandbox_flow_context.up.sql`
- Create: `migrations/000060_sandbox_flow_context.down.sql`
- Test: `internal/sandbox/*_test.go`

**Steps:**
1. Write failing tests for context/settings round trips, question answers, filtered pagination, and old-record defaults.
2. Add typed context, answer, settings, question, list-query, and list-result models.
3. Add the migration and update every repository query/scan in one change.
4. Add question and answer handlers; extend list and draft handlers without breaking current clients.
5. Run `go test ./internal/sandbox` and commit the backend slice.

### Task 2: Extend the web API client

**Files:**
- Modify: `apps/web/src/lib/sandboxApi.ts`
- Modify: `apps/web/src/lib/sandboxApi.test.ts`

**Steps:**
1. Write failing request-shape tests for questions, answers, filters, settings, and examples.
2. Add the TypeScript types and methods matching the Go contract.
3. Preserve the existing method signatures where current pages and tests depend on them.
4. Run the focused API tests and commit.

### Task 3: Build the shared sandbox frame

**Files:**
- Create: `apps/web/src/components/sandbox/SandboxFrame.tsx`
- Create: `apps/web/src/components/sandbox/SandboxCopilot.tsx`
- Create: `apps/web/src/components/sandbox/SandboxStepper.tsx`
- Create: `apps/web/src/components/sandbox/SandboxRoleCard.tsx`
- Create: `apps/web/src/sandbox-reference.css`
- Modify: `apps/web/src/main.tsx`

**Steps:**
1. Add component tests for active navigation, step states, Copilot variants, keyboard controls, and modal close behavior.
2. Implement a stable 1672px reference grid with the existing `V4PageShell` chrome.
3. Use existing sandbox image assets for the hero and five reference roles.
4. Add desktop, narrow-desktop, tablet, and mobile layouts.
5. Run component tests and commit.

### Task 4: Implement setup, questions, and completion

**Files:**
- Create: `apps/web/src/pages/sandbox/SandboxSetupView.tsx`
- Create: `apps/web/src/pages/sandbox/SandboxQuestionsView.tsx`
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Test: `apps/web/src/pages/SandboxPage.test.tsx`

**Steps:**
1. Write failing tests for draft creation, one-question-at-a-time saving, refresh recovery, skip, edit, and completion.
2. Make the homepage description create the draft instead of navigating to a hard-coded form.
3. Persist each answer through the API and derive completion from the returned session.
4. Route the completed summary to role selection with the session ID intact.
5. Run focused tests and commit.

### Task 5: Implement roles, startup settings, and quota

**Files:**
- Create: `apps/web/src/pages/sandbox/SandboxRolesView.tsx`
- Create: `apps/web/src/pages/sandbox/SandboxStartView.tsx`
- Create: `apps/web/src/components/sandbox/SandboxQuotaDialog.tsx`
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Test: `apps/web/src/pages/SandboxPage.test.tsx`

**Steps:**
1. Write failing tests for API roles, selection persistence, settings persistence, start, quota response, and dialog controls.
2. Match the five reference cards while allowing additional backend roles in a secondary row.
3. Implement depth, output style, outline toggle, edit, and advanced-settings interactions.
4. Open the quota dialog from actual usage or a 402 response and keep all dialog exits functional.
5. Run focused tests and commit.

### Task 6: Implement run, report, and history

**Files:**
- Create: `apps/web/src/pages/sandbox/SandboxRunView.tsx`
- Create: `apps/web/src/pages/sandbox/SandboxReportView.tsx`
- Create: `apps/web/src/pages/sandbox/SandboxHistoryView.tsx`
- Modify: `apps/web/src/pages/SandboxPage.tsx`
- Test: `apps/web/src/pages/SandboxPage.test.tsx`

**Steps:**
1. Write failing tests for queued/running/completed/failed/canceled states and eliminate the current stale draft regression.
2. Render role progress, recognized opportunities/risks, follow-up questions, retries, and report navigation from returned session data.
3. Render report metrics, conclusions, role summaries, actions, growth path, indicators, and timeline from the report contract with honest empty states.
4. Wire keyword/status/role/date filters and pagination to list requests; expose API-backed example records only through an explicit demo switch.
5. Run focused tests and commit.

### Task 7: Visual calibration and regression

**Files:**
- Modify: `apps/web/src/sandbox-reference.css`
- Remove or leave unreferenced: legacy sandbox presentation rules in `apps/web/src/styles.css`

**Steps:**
1. Run `npm test -- --run src/pages/SandboxPage.test.tsx src/lib/sandboxApi.test.ts`.
2. Run `go test ./internal/sandbox`, then `go test ./...`.
3. Run `npm run lint` and `npm run build`.
4. Start local services and capture all nine states at 1672 x 941 plus critical states at 390 x 844.
5. Measure shell, main column, Copilot, cards, and modal; fix overflow, blank canvas, overlap, and console errors.
6. Commit the final calibration and report any remaining external-service limitations.

