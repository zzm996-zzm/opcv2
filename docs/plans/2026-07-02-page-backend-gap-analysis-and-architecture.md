# 页面后端接口缺口分析与架构设计

**Goal:** 对当前所有已实现前端页面逐页梳理后端接口缺口，先确定领域边界、API 形态和实施顺序，再进入接口实现。

**Status:** Updated after first backend slice implementation on 2026-07-02.

**Implementation plan:** `docs/plans/2026-07-02-first-backend-slice-implementation.md`

**Sources checked:**
- `apps/web/src/App.tsx`
- `apps/web/src/pages/*.tsx`
- `apps/web/src/lib/*Api.ts`
- `internal/*/http.go`
- `docs/api/backend-api-contract.md`
- `docs/ui-inventory/top-tabs.md`
- `docs/ui-inventory/side-nav.md`
- `docs/plans/2026-06-23-backend-architecture-implementation.md`

---

## 1. Current Backend Coverage

当前后端已经不是空壳，核心基础已经具备：

| Domain | Existing endpoints | Coverage judgment |
| --- | --- | --- |
| Auth | SMS, login, register, refresh, logout, me | 登录闭环基本可用；账号资料、偏好、配额、注销已由 `account` 补齐；密码/绑定仍未实现。 |
| Analysis | direction, sessions, session detail, action items | 免费分析主链路较完整；缺竞品 Tab B 的外部证据采集、分析过程状态和更细报告结构。 |
| Membership | current membership, redemption redeem, plans, usage, orders, checkout | 会员中心第一批页面已接入；仍缺真实支付、订单取消、权益明细和管理端。 |
| Content | articles, tools filters/detail/favorite, article bookmark, community config/join, brand, help articles; admin writes | 工具、社群、资讯、帮助第一批页面已接入；仍缺工具 AI 推荐、资讯文件分析、社群活动等深水区能力。 |
| Copilot | threads, messages, compare, summary, memories, files, models, AI runs | 覆盖度最高；缺流式输出、真实文件上传解析、向量检索/RAG、用量门禁。 |
| Projects | create/list/get match, favorite | AI 匹配骨架可用；缺机会库、追问、结果详情、对比、导出、付费解锁。 |
| Sandbox | create/run/list/get sessions | 沙盘基础会话可用；缺分步草稿、角色/追问、运行进度、报告明细、额度。 |
| Tasks | create/list/get/update, list filters, stats | 任务列表和统计已可支持任务中心基础视图；仍缺提醒规则、来源关联、批量生成。 |
| Competitor | scans create/list/get, monitoring | 有扫描和监控展示骨架；缺异步进度、证据源、监控对象/规则 CRUD。 |
| Growth | models create/list/get, scenarios, forecast, cost items, action items | 增长测算核心展示已可由模型派生；仍缺保存快照、导出分享和敏感性分析。 |
| Leads | tasks create/list/detail, progress summary, result list | 获客任务和结果读取已可支撑基础页面；仍缺取消/重试、CRM 批量导入、评分深化和 provider 错误详情。 |
| CRM | import lead, list/detail/search/edit customers, update stage, record/list follow-up, copy, due customers, pipeline stats, activity timeline | CRM 核心动作和读模型可用；缺批量导入/更新和更细的提醒规则。 |
| Learning | courses, course detail, progress, diagnoses | 课程/诊断骨架可用；缺评估答案、差距分析、推荐方案、计划和课程学习动作。 |
| Dashboard | summary | 聚合接口可用；缺周期参数、周报生成、跨域数据可信来源约定。 |

Remaining whole-domain gaps after the first slice:
- `geo`
- `enterprise`
- `geo`
- `enterprise`
- richer `admin` APIs for membership/content/site config

---

## 2. Architecture Decisions

### ADR-001: Continue modular monolith, not microservices

Keep the existing Go modular monolith:

```text
internal/<domain>/
  types.go
  service.go
  postgres_repository.go
  http.go
  *_test.go
```

Reason:
- The product still needs fast iteration and shared transactions for user, credits, jobs, and CRM.
- PostgreSQL + Redis + API + Worker already matches the current codebase.
- Microservices would add operational cost before product boundaries are stable.

### ADR-002: Prefer domain APIs over page-specific APIs

Do not build `/home/page-state`, `/projects/page-state`, etc. except for true aggregation surfaces such as dashboard and home workbench.

Reason:
- Most pages are different views of reusable domains: content, learning, project matches, CRM, notifications.
- Page-specific endpoints would duplicate data contracts and make later redesign expensive.

Allowed aggregation endpoints:
- `GET /api/v1/home/summary`
- `GET /api/v1/dashboard/summary`
- `POST /api/v1/dashboard/reports`

### ADR-003: Long-running work is a job with pollable status

Use `internal/jobs` + Redis/Asynq for:
- lead search
- competitor scan
- GEO acquisition run
- file analysis
- report/export generation
- expensive AI workflows that can exceed HTTP timeouts

HTTP request creates a domain record and returns an id. Frontend polls:

```text
POST /api/v1/<domain>/jobs
GET  /api/v1/<domain>/jobs/:id
GET  /api/v1/<domain>/jobs/:id/results
POST /api/v1/<domain>/jobs/:id/cancel
POST /api/v1/<domain>/jobs/:id/retry
```

### ADR-004: SSE only where user experience needs token streaming

Default to HTTP + polling. Use SSE only for:
- Copilot chat streaming
- analysis/report generation if the UI needs live generation text

Do not introduce WebSocket now.

### ADR-005: Entitlements are checked in services, not frontend only

Membership, quotas and credits affect:
- Copilot model/file usage
- project unlock/export
- sandbox run count
- lead task count/result limit
- competitor scan/monitoring quota
- GEO task count

Every paid/limited action should call a service-level entitlement check before creating work or charging credits.

---

## 3. Target Domain Map

| Domain | Owns | Should not own |
| --- | --- | --- |
| `account` | profile fields, onboarding details, password/binding/delete requests, preferences | membership plans, content history calculation |
| `membership` | plans, current subscription, quotas, credit ledger, orders, redemptions | feature-specific business records |
| `notifications` | message center, read state, notification preferences, delivery records | domain source objects such as tasks or reports |
| `content` | articles, tools, community config, brand config, help articles | user-specific learning/progress |
| `learning` | courses, progress, diagnosis, assessment, recommendations, plans | generic content articles |
| `projects` | opportunity catalog, match sessions, questions, results, favorites, compare, export | CRM customer lifecycle |
| `sandbox` | sandbox session drafts, roles, questions, run status, reports | project market matching |
| `tasks` | user tasks, reminders, source links, generated tasks | source domain business state |
| `competitor` | scan tasks, evidence, tracked targets, monitoring rules/events | general web search provider code |
| `growth` | growth models, scenarios, forecast snapshots, action items | dashboard aggregation |
| `geo` | GEO projects, keywords, engines, content tasks, lead signals | CRM customer records |
| `leads` | lead tasks, provider evidence, scoring, result list, CRM handoff | CRM follow-up state |
| `crm` | customers, stages, activities, follow-ups, reminders, AI copy | lead search provider records |
| `dashboard` | cross-domain read aggregation, report generation | ownership of source data |
| `copilot` | AI conversations, memories, files, model selection, compare | generic document storage for all domains |
| `enterprise` | inquiries, packages, delivery board, milestones, cases | membership subscription rules |

---

## 4. API Conventions

### Common response rules

- Base path stays `/api/v1`.
- Protected endpoints require Bearer token.
- List endpoints use `limit`, later add `cursor` for infinite lists.
- Empty collections return `[]`, never `null`.
- Domain records are user-scoped unless under `/admin`.
- IDs in URL are numeric unless a resource is intentionally addressed by `slug`.

### Status model for jobs

Use one shared vocabulary:

```text
draft -> queued -> running -> succeeded
                   |-> failed
                   |-> canceled
                   |-> refunded
```

Each job-backed domain can add `progress_percent`, `current_step`, `error_code`, `safe_error_message`.

### Frontend migration rule

For each page:
1. Add/extend typed client in `apps/web/src/lib/<domain>Api.ts`.
2. Add page test that mocks backend success and failure.
3. Replace static arrays with backend data while keeping fallback only for explicit empty/demo state.
4. Update `docs/api/backend-api-contract.md`.

---

## 5. Page-by-Page Gap Matrix

### Entry, auth, account

| Routes | Page | Current backend | Missing APIs |
| --- | --- | --- | --- |
| `/`, `/home/notice`, `/home/account`, `/assistant/settings`, `/assistant/files`, `/assistant/collapsed` | `HomePage` | `GET /home/summary`, notification summary, account summary; Copilot files exist separately | `GET /account/quick-links`; assistant files UI can later reuse `GET /copilot/files`. |
| `/login` | `LoginPage` | Auth endpoints exist | Production SMS provider hardening, login audit display if needed. |
| `/register/details` | `RegisterDetailsPage` | Onboarding APIs exist | Wire the page to `GET/PUT /account/onboarding` and `POST /account/onboarding/complete` if this flow is kept. |
| `/terms`, `/privacy` | `LegalPage` | Static | No backend needed for first version unless legal copy becomes CMS-managed. |
| `/profile`, `/profile/settings`, `/profile/content`, `/profile/preferences`, password/logout/delete overlays | `ProfilePage` | Profile, onboarding, preferences, quotas, content and delete APIs wired | `PATCH /account/password`, bindings, activity feed. |
| `/messages`, `/messages/:messageId` | `MessagesPage` | Notifications list/detail/read/read-all wired | Source deep-link consistency and richer notification filters. |
| `/help` | `HelpPage` | Help topics/articles and support tickets wired | Help article detail route/modal and support ticket attachments if needed. |

### Membership

| Routes | Page | Current backend | Missing APIs |
| --- | --- | --- | --- |
| `/membership`, `/membership/upgrade` | `MembershipPage` | Current membership, plans, usage, orders, checkout, redemption wired | Benefits detail, real payment provider, order cancel, admin plan/code/order management. |

### Top tabs

| Routes | Page | Current backend | Missing APIs |
| --- | --- | --- | --- |
| `/tools`, `/tools/all`, `/tools/recommend`, `/tools/recommendation-plan`, `/tools/detail` | `ToolsPage` | Tool list filters/search/sort, detail and favorite wired | `POST /tools/recommendations`, `GET /tools/recommendation-plans/:id`. |
| `/copilot`, `/copilot/*` | `CopilotPage` | Strong coverage: threads, messages, compare, memories, files, models, runs | SSE streaming, binary upload + parsing, object storage, embeddings/RAG, file delete/update, thread search, usage quota checks. |
| `/learning`, diagnosis, assessment, gap-analysis, recommendation, plan, report, courses, course intro/detail, recommended courses, history | `Learning*Page` | courses, detail, progress, diagnosis latest/create; latest diagnosis gaps/recommendations/plan/report derived endpoints wired | `POST /learning/assessments`, `GET /learning/assessments/:id`, diagnosis-id-specific reads if historical reports need stable snapshots, `POST /learning/progress`, `PATCH /learning/progress/:courseSlug`, `GET /learning/history`, `GET /learning/courses/:slug/materials`. |
| `/community`, `/community/members`, `/community/enterprise` | `Community*Page` | Community config and join requests wired | Extend config schema for plans/features/posts/events/stats/QR codes, `GET /community/events`, `POST /community/events/:id/register`. |
| `/insights`, `/insights/detail`, `/insights/file-analysis` | `InsightsPage` | Article list/detail and bookmark wired through query slug | Proper `/insights/:slug` route, article filters, file analysis jobs/result APIs. |

### Project and execution

| Routes | Page | Current backend | Missing APIs |
| --- | --- | --- | --- |
| `/analysis`, `/analysis/history`, `/analysis/sessions/:sessionId` | `Analysis*Page` | Direction, history, detail, action items | Competitor Tab B evidence job, optional SSE/progress, source references, better report section contract, action item batch update. |
| `/projects`, match/explore/cases/questions/results/paywall/history/detail/compare/export | `ProjectsPage` | create/list/get match, favorite; history and `/projects/matches/:matchId` detail are wired to existing match APIs | `GET /projects/opportunities`, `GET /projects/opportunities/:slug`, `GET /projects/cases`, `POST /projects/matches/:id/questions`, `POST /projects/matches/:id/answers`, `GET /projects/matches/:id/results`, `POST /projects/matches/:id/unlock`, `POST /projects/compare`, `POST /projects/exports`, `GET /projects/exports/:id`. |
| `/sandbox`, setup/roles/start/questions/run/report/history/quota | `SandboxPage` | create/run/list/get sessions; history links and `/sandbox/sessions/:sessionId/report` are wired to session detail | `PATCH /sandbox/sessions/:id/draft`, `GET /sandbox/roles`, `POST /sandbox/sessions/:id/questions`, `POST /sandbox/sessions/:id/answers`, `GET /sandbox/sessions/:id/status`, `GET /sandbox/quota`. |
| `/tasks` | `TasksPage` | create/list/get/update, filters/search/status params, `GET /tasks/stats` | Reminders, source links, `POST /tasks/generate`, batch update, calendar/due views. |

### Growth and acquisition

| Routes | Page | Current backend | Missing APIs |
| --- | --- | --- | --- |
| `/competitor-data` | `CompetitorDataPage` | create/list/get scans | Scan status/progress, result evidence, competitor profiles, source health, conclusions, task flow, export, cancel/retry. |
| `/competitor-monitoring` | `CompetitorMonitoringPage` | `GET /competitor/monitoring` | `POST/PATCH/DELETE /competitor/watchlist`, alert rules CRUD, channel health, notifications integration, event read/triage actions. |
| `/growth-calculator` | `GrowthCalculatorPage` | growth model create/list/get, derived scenarios/forecast/recommendations | Saved snapshots, export/share, sensitivity analysis. |
| `/geo` | `GeoAcquisitionPage` | None | New `geo` module: projects, keywords, engines, content tasks, lead signals, run jobs, result evidence, CRM handoff. |
| `/leads` | `LeadDevelopmentPage` | lead task create/list/detail, progress summary, results list | Cancel/retry, deeper scoring, source health, `POST /leads/tasks/:id/import-crm`, batch import, provider error details. |
| `/dashboard` | `DashboardPage` | summary | Period filters, source freshness, report generation, action deep links, cross-domain aggregation consistency rules. |
| `/crm`, `/crm/follow-ups` | `CrmPage` | customer list/search/edit/detail API, customer activities, due customers, follow-up list, pipeline stats, import lead, stage update, follow-up, AI copy | Richer reminders, batch import/update, deeper activity filters. |
| `/enterprise` | `EnterprisePage` | None | `GET /enterprise/plans`, `GET /enterprise/cases`, `POST /enterprise/inquiries`, `GET /enterprise/delivery-board`, `GET /enterprise/milestones`. |

---

## 6. Proposed Implementation Phases

### Phase 0: Contract freeze before coding

Deliverables:
- Update `docs/api/backend-api-contract.md` with new endpoint groups and response shapes.
- Add TypeScript client method stubs only after contracts are accepted.
- Decide which pages must be real-data in next release vs allowed config/static.

No database migration or handler implementation in this phase.

### Phase 1: Low-risk CRUD/config gaps

Implement first because they remove many static arrays without provider complexity:

1. `account`: profile, onboarding, preferences, quotas read model.
2. `notifications`: message center list/detail/read.
3. `membership`: plans, orders, usage; wire current membership and redemption into UI.
4. `content`: tools detail/search/favorite, community config expansion, help articles, insights slug pages.

Expected affected pages:
- Home, Profile, Messages, Help, Membership
- Tools, Community, Insights

### Phase 2: Existing workflow depth

Extend domains that already exist:

1. `projects`: opportunities/cases, questions/answers, results, compare, export.
2. `sandbox`: draft steps, status/report/quota, question/answer persistence.
3. `learning`: assessment, gap, recommendation, plan, report, progress update.
4. `tasks`: stats, filters, reminders, source links.
5. `crm`: richer reminders, batch import/update, deeper activity filters.
6. `growth`: scenarios, forecast, action items. Done for derived model views; saved snapshots and export remain later work.

Expected affected pages:
- Projects, Sandbox, Learning, Tasks, CRM, Growth.

### Phase 3: Job-backed acquisition and monitoring

Implement job-heavy modules after contracts and UI polling conventions are stable:

1. `leads`: result detail and progress are wired; cancel/retry, provider error details and CRM batch import remain.
2. `competitor`: scan progress/evidence, watchlist/rules CRUD.
3. `geo`: new module with projects, keywords, engines, content tasks, lead signals.
4. `insights`: file analysis jobs.

Expected affected pages:
- Leads, Competitor Data, Competitor Monitoring, GEO, Insights file analysis.

### Phase 4: AI/paid experience polish

Add expensive or architecture-sensitive capabilities:

1. Copilot SSE streaming.
2. Binary file upload, parsing, object storage, embeddings/RAG.
3. Project/sandbox/report exports.
4. Entitlement gates and credit ledger checks for all expensive actions.
5. Dashboard weekly report generation.
6. Observability dashboards for AI runs, jobs and provider failures.

---

## 7. Recommended Next Step

Do not start implementation from a random page. Start Phase 0:

1. Choose the next release slice.
2. Update `docs/api/backend-api-contract.md` for that slice.
3. Add or update frontend typed API clients.
4. Implement backend by domain.
5. Replace page fallback data only after success/failure tests exist.

Recommended first release slice:

```text
account + notifications + membership + content/help
```

Reason:
- It unlocks many visible static pages.
- It has low dependency on AI providers and async jobs.
- It gives a stable foundation for Home/Profile/Messages/Membership/Help before deeper workflow work.

---

## 8. Phase 0 Contract Draft

这部分只定义第一批建议实现的 API 草案。接受后再同步到 `docs/api/backend-api-contract.md`，再写迁移、服务和页面接入。

### 8.1 Account

New module: `internal/account`.

Owned pages:
- `RegisterDetailsPage`
- `ProfilePage`
- Home account/preferences panels

Endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/account/profile` | 当前用户资料、企业资料、绑定状态。 |
| `PATCH` | `/api/v1/account/profile` | 更新昵称、企业、行业、角色、联系方式等资料字段。 |
| `GET` | `/api/v1/account/onboarding` | 注册后资料完善页草稿和完成状态。 |
| `PUT` | `/api/v1/account/onboarding` | 保存注册后资料完善表单。 |
| `POST` | `/api/v1/account/onboarding/complete` | 标记 onboarding 完成并进入工作台。 |
| `PATCH` | `/api/v1/account/password` | 修改密码。 |
| `GET` | `/api/v1/account/preferences` | 通知、默认模型、显示偏好。 |
| `PATCH` | `/api/v1/account/preferences` | 更新偏好。 |
| `GET` | `/api/v1/account/quotas` | AI、获客、沙盘、导出等额度读模型。 |
| `GET` | `/api/v1/account/content` | 用户最近内容：报告、匹配、沙盘、导出。 |
| `DELETE` | `/api/v1/account` | 账号注销请求。第一版可软删除或进入待处理状态。 |

Initial response model:

```json
{
  "profile": {
    "id": 42,
    "nickname": "张晨",
    "phone": "13800138000",
    "email": "",
    "wechat": "",
    "company": "智活AI科技有限公司",
    "industry": "企业服务",
    "role": "创始人",
    "onboarding_completed": true,
    "created_at": "2026-07-02T10:00:00Z",
    "updated_at": "2026-07-02T10:00:00Z"
  },
  "bindings": [
    { "type": "phone", "masked_value": "138****8000", "bound": true },
    { "type": "wechat", "masked_value": "", "bound": false }
  ]
}
```

Design notes:
- Use existing `users` where possible, add `user_profiles`, `user_preferences` only for fields that do not belong in auth.
- Do not store password change logic in `account`; call auth service through a narrow interface.

### 8.2 Notifications

New module: `internal/notifications`.

Owned pages:
- `MessagesPage`
- Home notice dropdown
- Notification hooks from tasks, reports, lead jobs, CRM reminders.

Endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/notifications?type=&status=&limit=` | 消息列表。 |
| `GET` | `/api/v1/notifications/:id` | 消息详情。 |
| `PATCH` | `/api/v1/notifications/:id/read` | 标记单条已读。 |
| `POST` | `/api/v1/notifications/read-all` | 批量已读。 |
| `GET` | `/api/v1/notifications/summary` | 未读数和分类计数，供首页使用。 |

Notification model:

```json
{
  "id": 101,
  "type": "task",
  "title": "任务提醒：AI 智能硬件项目拆解完成",
  "summary": "报告已生成，可继续查看拆解结果。",
  "body": "完整消息正文",
  "source_type": "analysis_session",
  "source_id": 99,
  "action_label": "查看拆解报告",
  "action_url": "/analysis/sessions/99",
  "read_at": null,
  "created_at": "2026-07-02T10:00:00Z"
}
```

Design notes:
- Source domains can create notification records, but notification module owns read state and delivery state.
- First version only needs in-app notifications; do not introduce email/SMS/push delivery yet.

### 8.3 Membership

Existing module to extend: `internal/membership`.

Owned pages:
- `MembershipPage`
- Upgrade modal
- Profile/Home quota views

Endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/membership/me` | Already exists; extend with plan, usage and expiry fields if needed. |
| `GET` | `/api/v1/membership/plans` | 套餐目录和权益对比。 |
| `GET` | `/api/v1/membership/usage` | 当前用户功能额度使用情况。 |
| `GET` | `/api/v1/membership/orders?limit=` | 订单记录。 |
| `POST` | `/api/v1/redemptions/redeem` | Already exists; keep. |
| `POST` | `/api/v1/membership/checkout` | 创建升级订单。真实支付未接入前返回 manual/pending 状态。 |

Plan model:

```json
{
  "plans": [
    {
      "id": 2,
      "code": "pro",
      "name": "会员版",
      "price_cents": 6900,
      "billing_cycle": "month",
      "features": ["线索数据实时更新", "去水印导出结果"],
      "quotas": [
        { "key": "lead_tasks", "label": "AI线索任务", "limit": 30, "unit": "次/月" }
      ],
      "recommended": true
    }
  ]
}
```

Design notes:
- First implementation can seed plans by migration and expose read APIs.
- Checkout can be manual order until payment is decided.
- Quota checks must eventually happen in each expensive domain service, not only in membership endpoints.

### 8.4 Content and Help

Existing module to extend: `internal/content`. New support ticket logic can live in `internal/support` or a small `support` submodule later; first version can start as `internal/content` help articles plus `internal/support` tickets.

Owned pages:
- `ToolsPage`
- `InsightsPage`
- `CommunityPage`, `CommunityMembersPage`, `CommunityEnterprisePage`
- `HelpPage`

Endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/content/tools?category=&q=&sort=&limit=` | 工具列表筛选搜索。 |
| `GET` | `/api/v1/content/tools/:slug` | 工具详情页。 |
| `POST` | `/api/v1/content/tools/:slug/favorite` | 收藏工具。 |
| `DELETE` | `/api/v1/content/tools/:slug/favorite` | 取消收藏。 |
| `POST` | `/api/v1/tools/recommendations` | 根据需求生成推荐工具。可先同步 AI，后续 job 化。 |
| `GET` | `/api/v1/tools/recommendation-plans/:id` | 工具组合方案。 |
| `GET` | `/api/v1/content/articles?category=&q=&limit=` | 咨询通文章列表。 |
| `GET` | `/api/v1/content/articles/:slug` | 文章详情。 |
| `POST` | `/api/v1/content/articles/:slug/bookmark` | 收藏文章。 |
| `DELETE` | `/api/v1/content/articles/:slug/bookmark` | 取消收藏。 |
| `GET` | `/api/v1/content/community` | Already exists; extend schema for three community pages. |
| `POST` | `/api/v1/community/join-requests` | 申请加入社群。 |
| `GET` | `/api/v1/help/topics` | 帮助主题。 |
| `GET` | `/api/v1/help/articles?topic=&q=&limit=` | 帮助文章搜索。 |
| `GET` | `/api/v1/help/articles/:slug` | 帮助文章详情。 |
| `POST` | `/api/v1/support/tickets` | 提交工单。 |
| `GET` | `/api/v1/support/tickets?limit=` | 我的工单列表。 |

Design notes:
- Community page content should stay CMS/config driven, not hard-coded in React arrays.
- Insights detail route should become `/insights/:slug` later; current `/insights/detail` can redirect to a default slug during transition.
- Tool recommendations are AI-generated user data, so they should not be mixed into static content tables.

### 8.5 Home Summary

New aggregation endpoint, implemented after the source modules above exist:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/home/summary` | 首页工作台聚合：推荐入口、最近任务、通知摘要、账号/额度摘要。 |

Response should only aggregate existing domain read models:

```json
{
  "hero_cards": [],
  "recommendations": [],
  "recent_tasks": [],
  "notification_summary": { "unread": 3, "latest": [] },
  "account_summary": { "plan_name": "会员版", "quota_warnings": [] }
}
```

Design notes:
- Home summary must not become the source of truth.
- It should tolerate partial domain failures and return degraded sections with safe empty arrays.

---

## 9. Implementation Checklist for the First Slice

Do this only after the contract draft is accepted.

1. Update `docs/api/backend-api-contract.md` with accepted endpoint contracts.
2. Add migrations for `user_profiles`, `user_preferences`, `notifications`, membership plan/order additions, help/support tables as needed.
3. Add backend tests first for repository, service and HTTP handlers.
4. Implement `internal/account`, `internal/notifications`, membership extensions, content/help extensions.
5. Add or extend frontend clients: `accountApi`, `notificationsApi`, `membershipApi`, `contentApi`, `supportApi`.
6. Wire pages in this order: `MembershipPage`, `MessagesPage`, `ProfilePage`, `HelpPage`, then Home panels.
7. Keep static fallback only as empty/demo state, not as silent replacement for failed backend calls.
8. Run `go test ./...`, `npm run lint`, `npm test -- --run`, `npm run build`.

---

## 10. Open Questions

1. 后台管理端是否现在就纳入第一版？`plan.md` 里有管理后台，但当前前端路由还没有 admin 页面。
2. 支付/订单是否先做手动订单和兑换码，还是需要接真实支付？
3. GEO 获客是否需要真实搜索引擎/AI 数据源，还是先做用户自建关键词和内容任务管理？
4. Copilot 文件是否下一阶段就做二进制上传/RAG，还是继续保留文本文件引用一段时间？
5. 哪些页面必须完全去掉 fallback 数据才能验收，哪些允许用后台配置种子数据？
