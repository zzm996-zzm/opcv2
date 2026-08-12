# Project Market V1.4 Acceptance

Date: 2026-08-12 (Asia/Shanghai)

Branch: `feature/bootstrap`

Result: PASS for local release acceptance. Online deployment was not performed.

## Environment

- API: `http://127.0.0.1:8081`
- Web: `http://127.0.0.1:5175`
- PostgreSQL: `127.0.0.1:5432`, migration version `76`, dirty `false`
- Redis: `127.0.0.1:6390`
- Project file provider: local filesystem
- Retrieval provider: PostgreSQL
- Research provider: development deterministic provider
- Active project knowledge-base version after acceptance: `3`, with `8` documents

The development research provider proves orchestration, persistence, evidence handling, recovery, and export behavior. Production startup rejects placeholder project providers, but external production-provider credentials and connectivity still require verification in the deployment environment.

## Automated Gate

All required commands exited zero:

```text
cd apps/web && npm test
  71 test files passed
  427 tests passed

cd apps/web && npm run build
  TypeScript build passed
  Vite production build passed

cd apps/web && npm run lint
  passed

go test ./...
  passed

go vet ./...
  passed

OPCV2_TEST_DATABASE_URL=postgres://... go test ./internal/projects -count=1
  passed against local PostgreSQL

git diff --check
  passed
```

The production build reports a non-blocking warning that the main JavaScript and CSS bundles exceed Vite's default size recommendation. This does not affect correctness, but route-level code splitting should be tracked as a performance follow-up.

## Catalog And Cases

- Home returned the four configured featured projects with stable heat values.
- Catalog search for `Excel` updated the URL and reduced the list to the matching project.
- Favorite and compare state persisted after browser reload.
- Five projects could be added to compare; the sixth returned `409 compare_limit_reached` with `max: 5`.
- The public case API exposed three evidence-qualified cases: two success and one failure.
- Case detail rendered claim-level facts, explicitly generated analysis, and public source links.
- Both `/projects/cases` and the canonical `/project-cases` list route work; `/project-cases/:id` is directly addressable.

Verified evidence cases:

- Klarna AI customer service success, sourced to Klarna's official release.
- Coca-Cola United automation success, sourced to Microsoft customer and Power Platform publications.
- Amazon recruiting AI failure, sourced to Reuters and BBC.

Four legacy demo cases with `example.com` evidence remain hidden by the evidence gate.

## Analytics And Heat

- Posting the same analytics `event_id` twice returned `duplicate: false`, then `duplicate: true`.
- The raw visitor key was not stored. The stored 64-character value matched the SHA-256 digest.
- A valid `project_detail_view` inserted one project view and triggered asynchronous heat aggregation.
- Unknown event names and unknown properties returned `400 invalid_event`.
- The complete `match_result` payload, including `web_trigger_reason`, now returns `202`.

## Match Workflow

- Uploaded a Markdown file and observed parse completion; file metadata and restored attachments were persisted.
- Completed clarification and generation for multiple sessions.
- Reloading the result route restored the generated projects and eight evidence items.
- A queued generation was canceled at attempt 1, then retried through the same generate endpoint as attempt 2 and completed.
- Persisted events for the retry were `queued`, `analyzing`, `retrieving_kb`, `researching_web`, `merging`, `generating`, and `done`.
- SSE with `Last-Event-ID: 26` replayed only events 27 through 30. It did not repeat event 26.
- Invalid `Last-Event-ID` returned `400 invalid_last_event_id`.
- Ownership, unsupported-file, MIME mismatch, cross-user access, cancellation, retry, and stale-attempt behavior are covered by the automated repository, service, HTTP, and frontend suites.

## Content Operations

- A normal user received `403 admin_required` for admin content endpoints.
- A batch with one valid and one invalid item persisted validation details.
- Publishing that partial batch returned `422 publication_gate_failed`; rollback succeeded.
- A fully valid evidence-backed failure batch published, became publicly eligible, then rolled back to rejected status.
- Publish, rollback, and knowledge-base rebuild actions produced operation audit entries.
- An API-requested knowledge-base rebuild completed through the worker, moved the active version from `2` to `3`, and wrote `8` documents.
- The AI-answer audit endpoint returned recent successful answer records.
- Temporary QA users, batches, compare items, and test-only knowledge-base jobs were removed after acceptance.

## PDF Export

The asynchronous PDF export completed and downloaded successfully:

- Artifact: `/Users/zzm/Downloads/project-match-4.pdf`
- Size: 177,979 bytes
- Format: valid `%PDF-1.3`
- Pages: 2
- Page size: A4
- Rendered QA images:
  - `/Users/zzm/data/www/opcv2/tmp/pdfs/project-market-qa/fixed-page-1.png`
  - `/Users/zzm/data/www/opcv2/tmp/pdfs/project-market-qa/fixed-page-2.png`

Visual inspection found no blank pages, CJK font corruption, header collision, content overlap, or raw JSON artifacts in evidence text.

## Responsive Browser Acceptance

Checked at `390x844`:

- Opportunity catalog
- Evidence case list
- Evidence case detail
- Persisted match result
- Export modal

All pages kept document width at 390 CSS pixels. The export dialog remained within the viewport, and the browser console had no errors. The global navigation is intentionally horizontally scrollable on mobile and did not create document-level overflow.

## Defects Found During Acceptance

The gate found and fixed these release-blocking defects:

1. `match_result` analytics rejected the frontend's allowed research reason property.
2. Import batches with validation failures could pass the publication SQL gate.
3. Knowledge-base rebuild reused one PostgreSQL parameter as integer and text, causing pgx type inference failure.
4. The canonical `/project-cases` list route was missing while home and detail navigation used that namespace.

PostgreSQL integration tests now cover the publication gate and versioned knowledge-base rebuild, including cleanup that restores the original active version.

## Release Boundary

Project Market V1.4 is ready for deployment-environment verification. Before asking for online product testing, the deployment must still confirm:

- real project research provider configuration and credentials;
- object/file storage configuration and download reachability;
- Redis worker availability and queue processing;
- database migrations through version 76;
- production CORS, public base URL, and PDF font/runtime dependencies.

Online deployment remains a separate action and must use the repository root `update.sh`.
