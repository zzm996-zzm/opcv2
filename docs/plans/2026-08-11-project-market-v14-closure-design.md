# Project Market V1.4 Closure Design

## Objective

Complete Project Market V1.4 as a usable, evidence-backed workflow before starting the Commercial Sandbox V1.2 frontend migration. A batch is complete only when a user can operate it in the browser, refresh and recover persisted state, understand failures, and verify the result with automated tests.

## Scope

The closure covers the Project Market home and catalog, project and evidence-case details, favorites and comparison, match input and file ingestion, clarification, retrieval and web evidence, result generation, diagnosis, export, analytics, and content administration. The existing full-open product decision remains authoritative: public project content is not hidden behind membership overlays while the paywall feature flag is disabled.

Commercial Sandbox frontend work is deliberately excluded until the Project Market acceptance suite and browser workflow pass.

## Product Contract

### Catalog and evidence

- The frontend uses the V1.4 public APIs: `/api/v1/projects/home`, `/api/v1/projects`, `/api/v1/projects/:id`, `/api/v1/project-cases`, and `/api/v1/project-cases/:id`.
- Search, filters, sorting, and pagination are server-backed. Counts and pagination metadata come from the database, never from display constants.
- Project and case details expose source attribution and evidence state. Invalid or unverified source URLs are not represented as verified evidence.
- Favorites and comparison require authentication and persist across refreshes. Comparison is limited to five projects with an actionable error when the limit is exceeded.

### Match workflow

- A user can provide text and supported files. Uploaded files are owned by the authenticated user, validated by size and detected type, parsed asynchronously where necessary, and persisted.
- The clarification workflow persists the parsed profile, field provenance, answers, skipped assumptions, completeness, and progress. Reopening or refreshing a session restores the current state.
- Generation is asynchronous and idempotent. Progress is persisted and available through both a status endpoint and replayable SSE; cancellation is checked before expensive stages.
- Candidate retrieval uses configured retrieval providers. Web research uses configured search/fetch providers and degrades visibly when unavailable rather than fabricating sources.
- Match recommendations link claims to project knowledge and research evidence. Insufficient evidence produces a partial/insufficient result with a clear explanation.

### Export and operations

- Report export is an asynchronous PDF/link workflow with persisted status and a seven-day expiry. The exported document contains the same recommendations and evidence available in the product.
- Analytics records the PRD-defined user actions without blocking the user workflow. Aggregates do not expose personal data.
- Content administration supports import batches, validation, publication, rollback, and knowledge-base reindexing with auditable statuses.

## Architecture

The Go/Gin modular monolith remains the system boundary. PostgreSQL is the source of truth for catalog data, user state, uploads, match sessions, progress, evidence, exports, analytics, and administration jobs. Redis/Asynq runs expensive parsing, research, matching, export, and reindex work. React consumes explicit V1.4 DTOs and treats server state as authoritative.

Provider interfaces remain behind the existing project provider boundary. Development providers may be used only in explicit development mode. Production startup must reject unavailable or unsafe placeholder providers for any enabled capability. Provider failures are mapped to stable API errors and persisted degraded states.

## Delivery Slices

1. **V1.4 browsing slice:** switch catalog and case pages to V1.4 APIs; complete real filtering, counts, pagination, full-open presentation, favorites, and comparison.
2. **Persistent input slice:** enable upload, validation, parsing, ownership checks, and session recovery.
3. **Evidence matching slice:** connect retrieval and research providers, implement sufficiency rules, and return evidence-linked recommendations.
4. **Completion slice:** PDF/link export, analytics, content administration, rollback, and reindex.
5. **Release gate:** full frontend/backend checks and browser E2E across desktop and mobile. Only then begin the Sandbox V1.2 frontend migration.

Each slice is committed and pushed only after its focused tests pass. Existing unrelated user changes, including `apps/web/src/profile-content-reference.css`, stay outside these commits.

## Failure And Recovery Rules

- Validation errors identify the field and correction.
- Authentication and ownership errors never reveal whether another user's resource exists.
- Provider or network failures preserve completed work and expose retryable/non-retryable status.
- Duplicate commands use idempotency keys or deterministic task IDs and return the existing operation.
- SSE disconnects do not lose work; clients reconnect with the last event ID and may fall back to status polling.
- Refreshing any persisted workflow resumes from server state instead of restarting it.

## Acceptance Gate

Project Market is complete only when all of the following pass:

- Focused React tests and focused Go tests for every slice.
- `npm test`, `npm run build`, and `npm run lint` in `apps/web`.
- `go test ./...` and `go vet ./...` at repository root.
- `git diff --check`.
- Browser tests covering catalog search/filter/pagination, case evidence, favorite/compare persistence, upload and clarification recovery, matching progress/reconnect/result, cancellation and retry, and PDF export download.
- Desktop and mobile screenshots show no blocked controls, stale membership overlays, clipped text, or incoherent overlap.
- No mocked success path or hard-coded count remains in the tested production UI.

