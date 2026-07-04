# Second Backend Slice: Workflow Depth

**Status:** Started on 2026-07-02.

**Goal:** Extend existing workflow domains after the first low-risk CRUD/config
slice. This slice should deepen pages that already have backend modules, without
introducing job-heavy provider integrations too early.

**Architecture rule:** Prefer extending existing domain packages (`tasks`,
`projects`, `sandbox`, `learning`, `crm`, `growth`) over adding cross-cutting
tables. Keep expensive AI/job-backed actions for a later slice unless the
underlying workflow state is already stable.

## Task 1: Tasks Filters and Stats

Status: Done.

Scope:
- Extend `GET /api/v1/tasks` with `status`, `project`, `q`, and `limit`.
- Add `GET /api/v1/tasks/stats`.
- Wire `TasksPage` status filters to backend list filters.
- Render task overview numbers from backend stats when available.

Actual:
- Added `tasks.ListFilters` and `tasks.Stats`.
- Extended service, HTTP handler, Postgres repository and tests.
- Extended `tasksApi.listTasks` and added `tasksApi.stats`.
- Updated `TasksPage` tests and UI status filters.

Verification:

```bash
go test ./internal/tasks
npm test -- --run src/lib/tasksApi.test.ts src/pages/TasksPage.test.tsx
```

Expected: PASS.

## Task 2: Projects Workflow Detail

Status: Done.

Recommended scope:
- Wire `ProjectsPage` detail/history views to existing `GET /projects/matches/:id`.
- Add question/answer persistence only if the page needs editable follow-up
  state before AI result generation.
- Defer exports and paid unlock until entitlement gates are formalized.

Actual:
- Added `projectsApi.getMatch`.
- Added `/projects/matches/:matchId` route while keeping the static
  `/projects/detail` sample route.
- Updated history rows to link real sessions into match detail.
- Updated detail view to load a match session by id and render the first ranked
  project into the page hero.
- Documented the implemented Projects match endpoints in the backend API
  contract.

Verification:

```bash
npm test -- --run src/lib/projectsApi.test.ts src/pages/ProjectsPage.test.tsx
```

Expected: PASS.

## Task 3: Sandbox Draft and Report Depth

Status: Started.

Recommended scope:
- Add draft update endpoint for setup/roles/questions.
- Add report read endpoint backed by existing session result fields before
  adding richer report tables.
- Keep async progress minimal until job conventions are shared with leads and
  competitor scans.

Actual so far:
- Reused existing `GET /api/v1/sandbox/sessions/:id` as the report read source
  instead of adding a duplicate report endpoint.
- Added `/sandbox/sessions/:sessionId/report` route.
- Updated history rows to link real sessions into report detail.
- Updated report view to load a session by id and render its report.

Verification:

```bash
npm test -- --run src/lib/sandboxApi.test.ts src/pages/SandboxPage.test.tsx
```

Expected: PASS.

## Task 4: Learning Assessment Chain

Status: Started.

Recommended scope:
- Add assessment submit/read.
- Derive gap/recommendation/plan/report from diagnosis records first.
- Add progress write/update once course detail pages need real state.

Actual so far:
- Added latest-diagnosis derived endpoints:
  - `GET /api/v1/learning/diagnoses/latest/gaps`
  - `GET /api/v1/learning/diagnoses/latest/recommendations`
  - `GET /api/v1/learning/diagnoses/latest/plan`
  - `GET /api/v1/learning/diagnoses/latest/report`
- Kept derivation inside `internal/learning` service; no new table or
  migration was needed.
- Extended `learningApi` with typed clients for the derived views.
- Wired Gap Analysis, Recommendation, Plan, and Report pages to the new
  endpoints with static fallback when no diagnosis exists.

Verification:

```bash
go test ./internal/learning ./internal/platform/httpserver
npm test -- --run src/lib/learningApi.test.ts src/pages/LearningGapAnalysisPage.test.tsx src/pages/LearningRecommendationPage.test.tsx src/pages/LearningPlanPage.test.tsx src/pages/LearningReportPage.test.tsx
```

Expected: PASS.

## Task 5: CRM List and Follow-up Board

Status: Started.

Recommended scope:
- Add customers list/detail/search/edit.
- Add follow-up list and pipeline stats.
- Keep batch import/update separate from the first CRM UI wiring.

Actual so far:
- Added `GET /api/v1/crm/customers` with `stage`, `q`, and `limit`.
- Added `GET /api/v1/crm/customers/:id`.
- Added `PATCH /api/v1/crm/customers/:id` for contact/profile edits.
- Added `GET /api/v1/crm/customers/:id/activities`.
- Added `GET /api/v1/crm/follow-ups` with optional `customer_id`.
- Added `GET /api/v1/crm/pipeline-stats`.
- Reused existing `crm_customers`, `crm_followups`, and `crm_activities`
  tables; no migration was needed.
- Extended `crmApi` and wired `CrmPage` customer stats/list and follow-up list
  to backend data with static fallback.
- Wired customer detail timeline to customer activities when backend data is
  available.

Verification:

```bash
go test ./internal/crm ./internal/platform/httpserver
npm test -- --run src/lib/crmApi.test.ts src/pages/CrmPage.test.tsx
```

Expected: PASS.

## Task 6: Growth Derived Scenario Views

Status: Done.

Recommended scope:
- Derive scenarios, monthly forecast, cost items, and action items from the
  existing growth model first.
- Avoid new snapshot tables until users can edit and compare multiple saved
  versions.
- Defer export/share and sensitivity sliders until the base read model is
  stable.

Actual:
- Added `GET /api/v1/growth/models/:id/scenarios`.
- Added `GET /api/v1/growth/models/:id/forecast`.
- Added `GET /api/v1/growth/models/:id/recommendations`.
- Kept derivation inside `internal/growth` service; no new table or migration
  was needed.
- Extended `growthApi` with typed clients for the derived views.
- Wired `GrowthCalculatorPage` scenario, forecast, cost, and recommendation
  sections to backend data with static fallback.

Verification:

```bash
go test ./internal/growth
npm test -- --run src/lib/growthApi.test.ts src/pages/GrowthCalculatorPage.test.tsx
```

Expected: PASS.

## Task 7: Leads Result Detail Read Model

Status: Done.

Recommended scope:
- Reuse existing `lead_tasks` and `lead_results` tables.
- Add task detail, progress summary, and result list read APIs before adding
  cancel/retry or CRM import writes.
- Wire `/leads` to display completed task results with fallback to task/status
  cards.

Actual:
- Added `GET /api/v1/leads/tasks/:id`.
- Added `GET /api/v1/leads/tasks/:id/results`.
- Added service-level owner checks so users cannot read other users' tasks or
  results.
- Added Postgres result scanning with evidence JSON parsing.
- Extended `leadsApi` with typed clients for task detail and results.
- Wired `LeadDevelopmentPage` to render latest completed task results as lead
  company cards while preserving static/task fallback.

Verification:

```bash
go test ./internal/leads
npm test -- --run src/lib/leadsApi.test.ts src/pages/LeadDevelopmentPage.test.tsx
```

Expected: PASS.
