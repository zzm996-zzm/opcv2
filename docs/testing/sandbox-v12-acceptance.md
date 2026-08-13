# Sandbox V1.2 Acceptance

Date: 2026-08-13 (Asia/Shanghai)

Branch: `feature/bootstrap`

Result: PASS for local release acceptance. Online deployment was not performed.

## Environment

- API: `http://127.0.0.1:18082`
- Web: `http://127.0.0.1:5174`
- PostgreSQL: `127.0.0.1:5432`, migration version `78`
- Redis: `127.0.0.1:6390`
- Worker: local `apps/worker` process
- Acceptance user ID: `17`
- Completed sandbox run ID: `5`
- AI provider: development deterministic provider

The development provider verifies orchestration, persistence, role isolation, streaming, report generation, export, and product handoffs. External model quality, credentials, latency, and provider availability remain deployment-environment checks.

## Automated Gate

The final release gate includes:

```text
go test ./...
go vet ./...
cd apps/web && npm test -- --run
cd apps/web && npm run build
cd apps/web && npm run lint
git diff --check
```

The sandbox frontend suite covers the V1.2 creation flow, clarification, role selection, SSE execution, report rendering, authenticated PDF export, task handoff, growth handoff, nullable historical reports, history search, missing-run recovery, and creating a distinct run for another simulation.

## Real Workflow

- Created run `5` and completed all seven selected roles through the worker.
- Received real progress events through authenticated SSE.
- Loaded the completed report through the canonical `/sandbox-runs/5/report` route.
- Asked a role-specific follow-up and received the persisted answer without leaking another role session.
- Rendered feasibility, purchase probability, opportunities, risks, role takeaways, disagreements, scenarios, assumptions, and advice.
- Saved selected advice as tasks through the report handoff endpoint.
- Confirmed repeated task submission is idempotent.
- Opened the growth calculator using the server-generated URL. The final route included `sandbox_run=5`, the validated project name, and channel input.
- Verified with a frontend regression test that another simulation posts a new run and navigates to the new run ID instead of reusing the completed run.

## Historical Report Compatibility

Real acceptance found an older report whose optional JSON collections contained `null`. This caused the report page to fail while reading collection lengths.

The backend now normalizes report and nested collections when generating and reading reports. The frontend also treats optional collections as empty arrays or objects. A regression fixture verifies that nullable assumptions, missing roles, disagreements, role key points, and related collections render without a blank page.

## PDF Export

The server-side PDF export completed and downloaded successfully:

- Artifact: `tmp/pdfs/sandbox-run-5-final.pdf`
- Size: 159,590 bytes
- Format: valid PDF 1.3
- Pages: 4
- Page size: A4
- Rendered QA image: `tmp/pdfs/sandbox-run-5-final-page1.png`

Visual inspection confirmed Chinese fonts, headers, footer page numbers, metrics, section hierarchy, and body content render correctly. A long Chinese project title now wraps by measured glyph width and is fully visible on two lines without clipping.

## Responsive Layout

The report stylesheet now collapses metrics, opportunity and risk cards, role cards, scenarios, advice, report status actions, export control, and footer actions to single-column mobile layouts below `760px`. Buttons use full available width, preventing the report handoff commands from overlapping at phone widths.

The in-app browser's viewport override did not apply to the active macOS browser surface during this run, so the final `390x844` screenshot could not be captured reliably. Responsive behavior is covered by the updated breakpoint rules, frontend regression suite, TypeScript build, and lint gate; a physical-device or deployment-browser pass remains part of online verification.

## Release Boundary

Sandbox V1.2 is ready for deployment-environment verification. Before asking for online product testing, deployment must confirm:

- production AI model routes, credentials, latency, and response schema stability;
- API and worker connectivity to the same PostgreSQL and Redis instances;
- migrations through version `78`;
- production PDF font availability and export download reachability;
- CORS, cookie, and public URL configuration;
- phone-width visual behavior in a real deployment browser.

Online deployment is a separate action and must use the repository root `update.sh`.
