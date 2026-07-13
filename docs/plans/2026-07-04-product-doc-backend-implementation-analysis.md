# OPC V4 product doc backend implementation analysis

Date: 2026-07-04

Last updated: 2026-07-13

Source product material:

- `/Users/zzm/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_5iu3dbhushlw22_7043/msg/file/2026-06/详细版 2/智活AI · OPC V4 0 开发需求（详细版）— 逐页功能规格 e333234f28d347deba1378d995da963d.md`
- Same folder screenshots `001.jpg` through `120.jpg`

This note focuses on backend architecture and implementation scope. UI pixel matching is intentionally out of scope.

## 1. Product backend conclusion

The product is not a simple page CRUD system. The backend should be designed as a modular monolith with five shared foundations:

1. User profile context: optional onboarding/profile fields are reused by Copilot, project matching, sandbox, learning diagnosis, growth calculation, competitor queries, and task generation.
2. Content and operations CMS: project catalog, tools, insights, courses, community QR codes, enterprise cases, package/quota config, and script account pools must be operator-maintained rather than hardcoded.
3. Membership, entitlement, and quota engine: limited actions must be checked and charged server-side, with real reset cycles.
4. AI workflow layer: Copilot, project matching, sandbox, learning diagnosis, task generation, tool recommendation, insight retrieval, growth calculation, and report generation need structured prompts, validation, run records, and sometimes streaming.
5. Job/script workflow layer: competitor full-data scan, competitor monitoring, GEO analysis, export generation, file analysis, and provider-backed lead search should be asynchronous jobs with status, retry, evidence, and source traceability.

Keep the current Go modular monolith. Do not split services yet. Domain boundaries are still moving, while user, quota, AI run, job status, notifications, and tasks need shared transactional behavior.

## 2. Current implementation snapshot

The repository already has a substantial backend skeleton:

- API app wiring: `apps/api/main.go`
- Router: `internal/platform/httpserver/router.go`
- Shared infra: PostgreSQL migrations, Redis cache/session/code store, task queue client/mux, health checks
- Domain packages: `auth`, `account`, `membership`, `notifications`, `home`, `content`, `support`, `ai`, `copilot`, `analysis`, `projects`, `sandbox`, `tasks`, `competitor`, `growth`, `learning`, `leads`, `crm`, `dashboard`, `geo`, `enterprise`
- Frontend clients exist under `apps/web/src/lib/*Api.ts`, so most pages already have typed API entry points.

Current coverage by product area:

| Area | Current state | Main gap |
| --- | --- | --- |
| Auth/account | SMS login/register/refresh/logout, profile, onboarding JSON sections, preferences, account content/quotas skeleton | WeChat scan login, phone/WeChat binding, password update wiring, richer 6-group profile schema, profile write-back prompts |
| Membership | Plans, quotas, transactional check/consume/refund with idempotency, resettable usage cycles, orders, manual checkout, redemption | Real payment callback, subscription activation, broader feature coverage, admin config surface |
| Content/top nav | Articles, tools, help, community config, favorites/bookmarks, limited admin upserts | Full CMS fields, tool recommendation, insight search/RAG, community QR code variants, course/admin content management |
| Copilot | Threads/messages, compare, memories, quota gating, multipart file upload and extraction, file lifecycle, upstream SSE streaming, task/project tools, model options, AI run history | Embeddings/RAG, object storage, richer document parsers, additional module tools |
| Project market | Opportunity and evidence catalog, match sessions, persistent follow-up Q&A, compare, export, task sync, list/detail/favorite APIs | Remove remaining frontend static fallbacks, complete saved-project UX, paid unlock policy |
| Sandbox | Create/list/get/run session; AI JSON report | Draft step persistence, roles catalog, async progress, per-role conversation, quota check/charge, report evidence/labels |
| Tasks | CRUD, filters, stats | Reminder rules, notification delivery, AI task generation, source links from other modules, batch operations |
| Competitor | Scan create/list/get, monitoring read model | Currently uses generated default conclusions; needs script job queue, account pool, source evidence, status/progress, CRUD monitoring rules |
| Growth | Models plus derived scenarios/forecast/recommendations | Dynamic AI clarification, saved snapshots, transparent assumptions, export/share, quota/paid gates |
| Learning | Courses, progress read, diagnoses and derived gap/recommendation/plan/report | Assessment submit/read, progress writes, course materials, recommendation snapshots |
| GEO | Overview and queued analysis request skeleton | Product doc says GEO is mostly locked/placeholder in this version; if kept, mark example data and avoid pretending real execution |
| Leads/CRM/Dashboard | Lead tasks/results, CRM customer/follow-up board, dashboard summary | Batch import, provider health/error detail, CRM handoff depth, report generation |
| Enterprise | Overview read model | Product doc needs public intro, cases, inquiry form, consultant QR; current API is too dashboard-like |

### 2.1 Progress handoff (2026-07-13)

The latest completed large module is **Copilot deep closure**. Backend and frontend work were completed together in five small commits and pushed to `origin/feature/bootstrap`:

| Commit | Completed scope |
| --- | --- |
| `f5a410e` | Resettable Copilot message/compare quotas, idempotent consumption, failure refunds, frontend usage display |
| `684120d` | Real multipart upload, 10 MB limits, extension/UTF-8 checks, text and DOCX extraction |
| `38d3b99` | File detail/delete lifecycle, status/source/SHA256/extracted-character metadata, safe list responses |
| `8982fea` | Provider-backed SSE streaming, cancellation, incremental frontend rendering, persisted final messages |
| `15f56f5` | Whitelisted `create_task` and `project_match` tools, audited task source, persisted/rendered tool results |

Database migrations added:

- `000047_copilot_quotas`
- `000048_copilot_file_quota`
- `000049_copilot_file_metadata`

Final verification at `15f56f5`:

- `go test ./...`: passed.
- Frontend tests: 66 files and 358 tests passed.
- `npm run build`: passed.
- `npm run lint`: 0 errors; 2 pre-existing warnings in `CompetitorDataPage.tsx` and `CrmPage.tsx`.
- PostgreSQL 16 empty-database migration reached version 49.

No production deployment was performed for this module.

## 3. Backend design target

### 3.1 Shared tables/components to formalize

The current schema is mostly domain-local. The next backend work should introduce or harden these cross-cutting concepts:

| Component | Responsibility |
| --- | --- |
| `user_profile_facts` or richer `user_onboarding.sections` contract | Store the six profile groups as structured facts: identity, business/company, products, resources, goals, preferences. |
| `profile_writeback_suggestions` | Let modules propose profile updates, with user approval before persistence. |
| `entitlement_features` / `membership_usage` ledger | Server-side checks and usage charging for sandbox, competitor scan, dynamic monitoring, Copilot compare/deep thinking, exports. |
| `jobs` or consistent domain job tables | Shared statuses: `draft`, `queued`, `running`, `succeeded`, `failed`, `canceled`, plus progress, safe error, retry count. |
| `evidence_sources` or domain-specific evidence JSON | Store source URL/platform/screenshot/raw snapshot timestamp for project data, competitor data, insight answers, enterprise cases. |
| `admin_config`/CMS tables | Manage QR codes, package/quota config, source accounts, frequency controls, content publish state. |

### 3.2 Status model for long work

Use HTTP request creation plus polling for long tasks:

```text
POST /api/v1/<domain>/<resource>
GET  /api/v1/<domain>/<resource>/:id
POST /api/v1/<domain>/<resource>/:id/cancel
POST /api/v1/<domain>/<resource>/:id/retry
```

Use SSE only for token streaming where the UI needs live text:

- Copilot chat
- Optional sandbox/project/growth report generation

Do not introduce WebSocket now.

### 3.3 Entitlement rule

Every paid or limited action must check entitlement in the service layer, not only in React:

- sandbox run: free 1/month, member 20/month by configurable default
- competitor scan: free 5/month, member 200/month by configurable default
- competitor dynamic monitoring: can share scan quota at first
- Copilot compare/multi-model/deep thinking/file analysis
- exports and long-term history
- board-three locked "use" actions should return a clear paywall/upgrade response and should not create real jobs

Transactional check-and-consume/refund semantics now exist and are used by several implemented workflows, including Copilot. New limited actions must reuse this service rather than introduce domain-local counters or frontend-only checks.

## 4. Domain-by-domain backend implementation

### 4.1 Account/profile

Keep `internal/account`, but upgrade the data contract from the current flat fields plus arbitrary sections into a stable six-group schema:

- identity: nickname, avatar, identity type, industry, region
- business: company/project name, business model, stage, target customers, channels
- products: product/service list with name, price, selling points, status
- resources: skills, budget, time, team, existing tools/assets
- goals: short-term goal, long-term goal, bottleneck, desired help
- preferences: subscribed topics, risk preference, model, language

Add profile read context for AI modules:

```text
GET /api/v1/account/profile-context
POST /api/v1/account/profile-writebacks
PATCH /api/v1/account/profile-writebacks/:id/accept
```

### 4.2 Admin/CMS

The product doc marks admin as P0 because content, QR codes, quotas, and script accounts cannot go live by code edits.

Current `content.RegisterAdmin` has article/tool/community/brand upserts, but it is not enough. Recommended admin domains:

- content catalog: projects, tools, insights, courses, help articles
- community QR codes: member group, enterprise group, consultant QR
- membership plans and quotas
- competitor/script sources: platform accounts, frequency limits, script status
- user/order operations: manual permission adjustment, order correction, refunds
- source/evidence maintenance

First implementation can be API-only; no admin UI is required before C-side pages are wired.

### 4.3 Membership/quota

Turn `membership` into the gatekeeper for paid actions:

```go
CheckAndConsume(ctx, userID, featureKey, amount, idempotencyKey)
Refund(ctx, userID, featureKey, amount, reason, sourceID)
```

Suggested feature keys:

- `sandbox_runs`
- `competitor_scans`
- `competitor_monitoring`
- `copilot_messages`
- `copilot_compare_models`
- `project_matches`
- `exports`

Use idempotency keys so retries do not double-charge.

### 4.4 Project market

Add the missing catalog side. `projects` currently handles only AI matching sessions.

Needed backend:

```text
GET  /api/v1/projects/opportunities
GET  /api/v1/projects/opportunities/:slug
GET  /api/v1/projects/cases
POST /api/v1/projects/matches/:id/answers
GET  /api/v1/projects/matches/:id/results
POST /api/v1/projects/compare
POST /api/v1/projects/exports
```

Data model:

- opportunity: title, industry, tags, budget band, difficulty, resource requirements, heat stats
- opportunity section: success path, current data, strengths/weaknesses, lessons, avoid list
- case/evidence: source URL, title, publisher, captured at, credibility label
- unlock/export records

Do not fabricate heat/case/source data. Empty state is better than fake numbers.

### 4.5 Sandbox

Current sandbox runs synchronously and immediately stores a report. Product needs staged setup, multi-role simulation, quota, and report labeling.

Next backend slice:

- `GET /sandbox/roles`
- `PATCH /sandbox/sessions/:id/draft`
- `POST /sandbox/sessions/:id/run` should check quota and create a queued job if provider latency is high
- `GET /sandbox/sessions/:id/status`
- `POST /sandbox/sessions/:id/messages` for role follow-up
- report fields should explicitly mark model-derived assumptions, not real statistics

### 4.6 Competitor full-data scan and dynamic monitoring

This is the biggest backend risk in the product doc. The doc explicitly says "script login member account, non-official API". Current implementation does not do that yet; it creates completed scans with default generated conclusions.

Target architecture:

```text
User request
  -> competitor_scan_requests row
  -> quota check/consume
  -> enqueue script job
  -> worker selects platform account from account pool
  -> script runner captures raw result/evidence
  -> AI analysis creates conclusion/action suggestions
  -> notification + optional task creation
```

Needed tables:

- `competitor_scan_requests`: target, platform, status, progress, user_id
- `competitor_raw_snapshots`: scan_id, platform, raw JSON/blob pointer, captured_at
- `competitor_evidence_sources`: scan_id/event_id, source_url, screenshot/object key, source_type
- `competitor_script_accounts`: platform, account label, status, cooldown_until, failure_count
- `competitor_monitor_rules`: target, platforms, dimensions, schedule, status
- `competitor_monitor_events`: rule_id, occurred_at, source, title, detail, AI interpretation

Important: because ToS/封号 risk exists, keep platform-specific runner code behind interfaces and centralize account pool/frequency config in admin.

### 4.7 Growth calculator

Current derived model is a good base. The product doc wants a one-box input plus AI clarification.

Add:

- `POST /growth/drafts` with free text
- `POST /growth/drafts/:id/answers`
- `POST /growth/drafts/:id/calculate`
- `GET /growth/models/:id/snapshots`
- `POST /growth/models/:id/export`

Store assumptions separately from generated recommendations so reports can show transparent calculation口径.

### 4.8 Learning

Current learning diagnosis-derived endpoints are enough for basic UI. Missing writes:

- assessment submit/read
- diagnosis snapshot by id
- progress update
- course material list
- plan item completion

Keep courses in CMS/admin eventually. Learning should own user progress and diagnoses, not generic article content.

### 4.9 Copilot

Copilot deep closure is complete for the current release scope:

- `POST /api/v1/copilot/threads/:id/messages/stream` streams provider output over SSE and persists the final assistant message.
- OpenAI Responses, OpenAI-compatible chat completions, the model router, and the development provider implement streaming.
- Multipart uploads validate size/type/UTF-8 and extract supported text formats plus DOCX content.
- File list/detail/delete APIs expose lifecycle metadata without returning extracted full text in list responses.
- Membership quotas cover messages, model comparison, and file analysis with idempotent consumption and failure refunds.
- The whitelisted tool registry can create tasks and run project matching; tool results are stored in message metadata and rendered as frontend action cards.

Remaining Copilot depth is optional follow-up work rather than a release blocker: object storage, asynchronous parsing for large documents, PDF/spreadsheet parsers, embeddings/RAG, and additional tools such as sandbox or competitor scan. Any new tool must remain explicitly whitelisted and auditable.

### 4.10 Board-three locked pages and enterprise

The product doc says GEO, AI leads, dashboard, CRM are visible but not actually usable in this version, while enterprise consulting is real.

The current backend already has real-looking `geo`, `leads`, `dashboard`, and `crm` APIs. Decide product direction:

- If following the doc strictly, UI actions in these four pages should call a paywall/upgrade config endpoint and not execute real workflows.
- If keeping the implemented backend, mark any seeded/demo data clearly and avoid fake production claims.

For enterprise, current `/enterprise/overview` should be expanded toward the doc:

```text
GET  /api/v1/enterprise/public-overview
GET  /api/v1/enterprise/cases
GET  /api/v1/enterprise/cases/:slug
POST /api/v1/enterprise/inquiries
GET  /api/v1/enterprise/contact-config
```

## 5. Recommended next backend slices

### Slice A: Contract cleanup and paywall/usage foundation

1. Update `docs/api/backend-api-contract.md` with current routes and planned statuses.
2. Add membership `CheckAndConsume`/idempotency service methods.
3. Add feature keys and seeded quotas matching the product doc defaults.
4. Wire quota checks to sandbox run and competitor scan first.
5. Return consistent paywall errors for over-limit actions.

This slice is the highest leverage because many pages depend on honest paid/free behavior.

### Slice B: Profile schema and context reuse

1. Formalize six-group profile schema.
2. Add profile context endpoint for AI modules.
3. Add write-back suggestion/accept flow.
4. Use profile context in projects, sandbox, growth, learning, and Copilot prompts.

### Slice C: Competitor script-job architecture

1. Convert competitor scans from immediate completed defaults to queued/running/succeeded.
2. Add scan status/progress/evidence/source tables.
3. Add script account pool config API.
4. Add worker handler skeleton and provider interface.
5. Only then connect real scripts.

### Slice D: Content/admin catalog depth

1. Project opportunity catalog and detail sections.
2. Tool detail rich fields and recommendation endpoint.
3. Insight search/detail and AI Q&A with citations.
4. Community QR config variants.
5. Enterprise public cases and inquiry form.

### Slice E: AI workflow depth

1. Copilot SSE.
2. Sandbox draft/roles/follow-up.
3. Growth AI clarification.
4. Learning assessment writes.
5. Task AI generation and source links.

## 6. Immediate warnings

- Do not ship competitor scan as "real data" while it still uses default generated conclusions.
- Do not show board-three metrics as real unless they are backed by source data; mark as example or use empty states.
- Do not rely on frontend-only locks for paid features.
- Do not hardcode quotas in UI or static configs; the doc explicitly requires real resettable quotas.
- Do not let AI-generated scores/probabilities look like measured facts. Store and label them as model推演/模型测算 with assumptions.

## 7. Practical next step

The next large module is **Project Market transparency cleanup**, because the backend project workflow is substantially implemented while several frontend routes still display static compatibility data. Complete it as five backend/frontend-verified small modules:

1. Remove the static `/projects/results` fallback and add explicit loading, empty, and error states.
2. Replace the old static `/projects/detail` dashboard with opportunity, match, and evidence APIs.
3. Remove the static match-history fallback and render only persisted sessions.
4. Connect saved/favorite projects to real records and complete the save/unsave UX.
5. Remove static `ProjectCopilot` business recommendations, make displayed claims traceable, run backend/frontend tests and build, then push the completed large module.

After this large module, continue with the remaining domain sub-tasks in priority order: competitor script-job/evidence architecture, sandbox workflow depth, learning writes, content/admin catalog depth, and enterprise public conversion flows.
