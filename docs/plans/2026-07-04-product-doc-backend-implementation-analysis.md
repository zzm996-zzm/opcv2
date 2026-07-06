# OPC V4 product doc backend implementation analysis

Date: 2026-07-04

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
| Membership | Plans, quotas table, usage read, orders, manual checkout, redemption | Real payment callback, subscription activation, real quota deduction/reset engine, admin config surface |
| Content/top nav | Articles, tools, help, community config, favorites/bookmarks, limited admin upserts | Full CMS fields, tool recommendation, insight search/RAG, community QR code variants, course/admin content management |
| Copilot | Threads, messages, compare, memories, files metadata, model options, AI run history | SSE streaming, real upload/object storage/parsing, RAG, entitlement/quota gating, tool/function execution into modules |
| Project market | Match sessions with AI JSON, list/detail/favorite | Opportunity catalog, project detail evidence/source model, paid unlock, compare, export, follow-up Q&A persistence |
| Sandbox | Create/list/get/run session; AI JSON report | Draft step persistence, roles catalog, async progress, per-role conversation, quota check/charge, report evidence/labels |
| Tasks | CRUD, filters, stats | Reminder rules, notification delivery, AI task generation, source links from other modules, batch operations |
| Competitor | Scan create/list/get, monitoring read model | Currently uses generated default conclusions; needs script job queue, account pool, source evidence, status/progress, CRUD monitoring rules |
| Growth | Models plus derived scenarios/forecast/recommendations | Dynamic AI clarification, saved snapshots, transparent assumptions, export/share, quota/paid gates |
| Learning | Courses, progress read, diagnoses and derived gap/recommendation/plan/report | Assessment submit/read, progress writes, course materials, recommendation snapshots |
| GEO | Overview and queued analysis request skeleton | Product doc says GEO is mostly locked/placeholder in this version; if kept, mark example data and avoid pretending real execution |
| Leads/CRM/Dashboard | Lead tasks/results, CRM customer/follow-up board, dashboard summary | Batch import, provider health/error detail, CRM handoff depth, report generation |
| Enterprise | Overview read model | Product doc needs public intro, cases, inquiry form, consultant QR; current API is too dashboard-like |

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

The existing `membership_usage` read table is not enough. Add transactional check-and-consume semantics.

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

Current Copilot has the right domain shape. Next depth:

- SSE endpoint for message streaming
- real file upload pipeline: object storage path, parser job, extracted text, embeddings later
- function/tool execution registry to call modules: create task, run project match, start sandbox, start competitor scan
- model entitlement checks before compare/deep thinking
- quota usage records tied to `ai_runs`

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

If backend work continues now, start with Slice A. It stabilizes the monetization and usage contract before adding more page depth, and it gives every later workflow a consistent way to check permission, consume quota, refund failed jobs, and show upgrade states.
