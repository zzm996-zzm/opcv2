# API Integration Transparency Ledger

**Updated:** 2026-07-04

**Purpose:** Make backend integration status explicit while removing mock/static business data. This file is the execution ledger for what is implemented, what is wired, what is still fallback, and what is currently failing verification.

**Source plans:**
- `docs/plans/2026-07-02-page-backend-gap-analysis-and-architecture.md`
- `docs/plans/2026-07-02-first-backend-slice-implementation.md`
- `docs/plans/2026-07-02-second-backend-slice-workflow-depth.md`
- `docs/api/backend-api-contract.md`

## 1. Current Verification Snapshot

| Command | Result | Notes |
| --- | --- | --- |
| `go test ./...` | PASS | Backend packages and route registration tests pass. |
| `cd apps/web && npm run lint` | PASS | No current lint failures. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `cd apps/web && npm test -- --run` | PASS | 63 test files passed, 245 tests passed. |
| `cd apps/web && npm test -- --run src/pages/TasksPage.test.tsx src/pages/GrowthCalculatorPage.test.tsx src/pages/DashboardPage.test.tsx src/pages/CompetitorDataPage.test.tsx src/pages/CompetitorMonitoringPage.test.tsx` | PASS | First fallback-removal batch passed: 5 files, 19 tests. |
| `cd apps/web && npm test -- --run src/pages/CrmPage.test.tsx src/pages/LeadDevelopmentPage.test.tsx` | PASS | Second fallback-removal slice passed for CRM and Leads: 2 files, 8 tests. |
| `cd apps/web && npm test -- --run src/pages/ToolsPage.test.tsx src/pages/InsightsPage.test.tsx src/pages/HomePage.test.tsx` | PASS | Second fallback-removal slice passed for Tools, Insights, and Home: 3 files, 22 tests. |
| `cd apps/web && npm test -- --run src/pages/ProfilePage.test.tsx src/pages/HelpPage.test.tsx` | PASS | Second fallback-removal slice passed for Profile and Help: 2 files, 14 tests. |
| `cd apps/web && npm test -- --run src/pages/CommunityPage.test.tsx src/pages/CommunityMembersPage.test.tsx src/pages/CommunityEnterprisePage.test.tsx` | PASS | Second fallback-removal slice passed for Community pages: 3 files, 6 tests. |
| `cd apps/web && npm test -- --run src/pages/ProfilePage.test.tsx src/pages/HelpPage.test.tsx src/pages/CommunityPage.test.tsx src/pages/CommunityMembersPage.test.tsx src/pages/CommunityEnterprisePage.test.tsx` | PASS | Current turn combined frontend page verification: 5 files, 20 tests. |
| `cd apps/web && npm test -- --run src/pages/GeoAcquisitionPage.test.tsx src/pages/EnterprisePage.test.tsx` | PASS | Static-domain transparency slice passed for GEO and Enterprise pages: 2 files, 2 tests. |
| `go test ./internal/geo ./internal/enterprise ./internal/platform/httpserver ./apps/api && cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/lib/enterpriseApi.test.ts src/pages/GeoAcquisitionPage.test.tsx src/pages/EnterprisePage.test.tsx` | PASS | GEO and Enterprise overview API/client/page wiring passed: backend packages plus 4 frontend files, 6 frontend tests. |
| `go test ./internal/geo` | PASS | GEO Postgres repository tests passed after first failing on undefined `NewPostgresRepository`. |
| `go test ./internal/geo ./apps/api ./internal/platform/httpserver` | PASS | GEO repository, API main wiring, and router-adjacent packages passed. |
| `go test ./...` | PASS | Full backend suite passed after GEO overview persistence slice. |
| `cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/pages/GeoAcquisitionPage.test.tsx` | PASS | GEO typed client and page behavior remain green: 2 files, 3 tests. |
| `rg -n "智能客服系统怎么选\|ChatGPT\|豆包\|Kimi\|通义千问\|下游搜索量\|教育培训机构如何用\|AI客服选型后端关键词\|后端返回的内容任务" apps/web/src/pages/GeoAcquisitionPage.tsx internal/geo migrations --glob '!**/*_test.go'` | PASS | No old/sample GEO business records found outside tests. |
| `go test ./internal/enterprise` | PASS | Enterprise Postgres repository tests passed after first failing on undefined `NewPostgresRepository`. |
| `go test ./internal/enterprise ./apps/api ./internal/platform/httpserver` | PASS | Enterprise repository, API main wiring, and router-adjacent packages passed. |
| `go test ./...` | PASS | Full backend suite passed after Enterprise overview persistence slice. |
| `cd apps/web && npm test -- --run src/lib/enterpriseApi.test.ts src/pages/EnterprisePage.test.tsx` | PASS | Enterprise typed client and page behavior remain green: 2 files, 3 tests. |
| `cd apps/web && npm run lint` | PASS | Frontend lint remains green after Enterprise overview persistence slice. |
| `cd apps/web && npm test -- --run` | PASS | Full frontend suite remains green after Enterprise overview persistence slice: 63 files, 245 tests. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `rg -n "增长团队训练营\|AI获客陪跑\|企业定制系统\|连锁教育集团\|职业培训机构\|¥12万\|¥18万\|客服响应效率提升\|后端陪跑方案\|后端企业案例" apps/web/src/pages/EnterprisePage.tsx internal/enterprise migrations --glob '!**/*_test.go'` | PASS | No old/sample Enterprise business records found outside tests. |
| `go test ./internal/geo` | PASS | GEO analysis request service, handler, and repository tests passed after first failing on missing types/methods. |
| `go test ./internal/geo ./apps/api ./internal/platform/httpserver` | PASS | GEO analysis request API and router-adjacent packages passed. |
| `go test ./...` | PASS | Full backend suite passed after GEO analysis request slice. |
| `cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/pages/GeoAcquisitionPage.test.tsx` | PASS | GEO typed client and page submit behavior passed after first failing on missing client/page wiring: 2 files, 5 tests. |
| `go test ./internal/geo` | PASS | GEO analysis request list service, handler, and repository tests passed after first failing on missing list methods. |
| `go test ./internal/geo ./apps/api ./internal/platform/httpserver` | PASS | GEO analysis request list API and router-adjacent packages passed. |
| `go test ./...` | PASS | Full backend suite passed after GEO analysis request history slice. |
| `cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/pages/GeoAcquisitionPage.test.tsx` | PASS | GEO typed client and page request-history behavior passed after first failing on missing list client/page rendering: 2 files, 6 tests. |
| `go test ./internal/geo` | PASS | GEO analysis request detail service, handler, and repository tests passed after first failing on missing detail methods. |
| `go test ./internal/geo ./apps/api ./internal/platform/httpserver` | PASS | GEO analysis request detail API and router-adjacent packages passed. |
| `go test ./...` | PASS | Full backend suite passed after GEO analysis request detail slice. |
| `cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/pages/GeoAcquisitionPage.test.tsx` | PASS | GEO typed client request-detail method passed after first failing on missing `getAnalysisRequest`: 2 files, 7 tests. |
| `go test ./internal/geo` | PASS | GEO internal analysis request status update service and repository tests passed after first failing on missing status constants/update method. |
| `go test ./internal/geo` | PASS | GEO worker queue slice passed after first failing on missing `WithQueue`, `jobs.TypeGeoAnalysis`, `ProcessAnalysisRequest`, and worker handler wiring. |
| `go test ./internal/jobs ./internal/platform/taskqueue ./apps/api ./apps/worker ./internal/geo` | PASS | GEO job type, default queue registry, API enqueue wiring, and worker registration all pass. |
| `go test ./...` | PASS | Full backend suite passed after GEO worker queue slice. |
| `cd apps/web && npm test -- --run src/lib/geoApi.test.ts src/pages/GeoAcquisitionPage.test.tsx` | PASS | GEO typed client/page tests remain green after backend worker queue wiring: 2 files, 7 tests. |
| `cd apps/web && npm run lint` | PASS | Frontend lint remains green after GEO worker queue slice. |
| `cd apps/web && npm test -- --run` | PASS | Full frontend suite remains green after GEO worker queue slice: 63 files, 249 tests. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `rg -n "智能客服系统怎么选\|ChatGPT\|豆包\|Kimi\|通义千问\|下游搜索量\|教育培训机构如何用\|AI客服选型后端关键词\|后端返回的内容任务" apps/web/src/pages/GeoAcquisitionPage.tsx apps/web/src/lib/geoApi.ts internal/geo migrations --glob '!**/*_test.go'` | PASS | No old/sample GEO business records found outside tests after worker queue wiring. |
| `cd apps/web && npm run lint` | PASS | Frontend lint remains green after GEO analysis request detail slice. |
| `cd apps/web && npm test -- --run` | PASS | Full frontend suite remains green after GEO analysis request detail slice: 63 files, 249 tests. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `rg -n "智能客服系统怎么选\|ChatGPT\|豆包\|Kimi\|通义千问\|下游搜索量\|教育培训机构如何用\|AI客服选型后端关键词\|后端返回的内容任务" apps/web/src/pages/GeoAcquisitionPage.tsx apps/web/src/lib/geoApi.ts internal/geo migrations --glob '!**/*_test.go'` | PASS | No old/sample GEO business records found outside tests after request-detail API/client work. |
| `cd apps/web && npm run lint` | PASS | Frontend lint remains green after GEO analysis request history slice. |
| `cd apps/web && npm test -- --run` | PASS | Full frontend suite remains green after GEO analysis request history slice: 63 files, 248 tests. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `rg -n "智能客服系统怎么选\|ChatGPT\|豆包\|Kimi\|通义千问\|下游搜索量\|教育培训机构如何用\|AI客服选型后端关键词\|后端返回的内容任务" apps/web/src/pages/GeoAcquisitionPage.tsx apps/web/src/lib/geoApi.ts internal/geo migrations --glob '!**/*_test.go'` | PASS | No old/sample GEO business records found outside tests after request-history rendering. |
| `cd apps/web && npm run lint` | PASS | Frontend lint remains green after GEO analysis request slice. |
| `cd apps/web && npm test -- --run` | PASS | Full frontend suite remains green after GEO analysis request slice: 63 files, 247 tests. |
| `cd apps/web && npm run build` | PASS | Build passes; Vite still warns that the main JS chunk is larger than 500 kB. |
| `rg -n "智能客服系统怎么选\|ChatGPT\|豆包\|Kimi\|通义千问\|下游搜索量\|教育培训机构如何用\|AI客服选型后端关键词\|后端返回的内容任务" apps/web/src/pages/GeoAcquisitionPage.tsx apps/web/src/lib/geoApi.ts internal/geo migrations --glob '!**/*_test.go'` | PASS | No old/sample GEO business records found outside tests after request submission wiring. |

Resolved frontend test baseline issues from this pass:

| File | Test | Current symptom | Likely category |
| --- | --- | --- | --- |
| `apps/web/src/App.test.tsx` | `renders the Copilot comparison route for a signed-in user` | Previously expected 3 `回答完成` cards. | Fixed stale test expectation; comparison route now asserts the empty compare state instead of old static answers. |
| `apps/web/src/pages/CopilotPage.test.tsx` | `pauses an in-flight chat request` | Previously could not reliably find `已暂停本次对话`. | Fixed test mock to honor `AbortSignal`, matching real fetch behavior. |

## 2. Plan Completion Summary

### First backend slice

Status: Mostly complete and wired.

| Area | Backend | Frontend client | Page wiring | Remaining |
| --- | --- | --- | --- | --- |
| Account | Done | Done | Profile wired | `PATCH /account/password` is still not implemented. |
| Notifications | Done | Done | Messages and home notification summary wired | Source deep-link consistency still needs product pass. |
| Membership | Done | Done | Membership page wired | Real payment, order cancel, admin plan/code/order management. |
| Content/tools/community/help | Done | Done | Tools, Community, Insights, Help wired | Tool AI recommendations and recommendation plans are still planned. |
| Support tickets | Done | Done | Help feedback wired | Attachments are not implemented. |
| Home summary | Done | Done | Home wired | Hero/recommendation content still depends on backend/config quality. |

### Second backend slice

Status: Partially complete.

| Area | Backend/API status | Page wiring status | Remaining |
| --- | --- | --- | --- |
| Tasks filters/stats | Done | Tasks page wired; silent fallback business rows removed | Reminders, source links, generated task actions, batch update. |
| Projects match detail | Done for existing match APIs | History/detail partly wired | Opportunities, cases, questions/answers, result detail model, unlock, compare, export. |
| Sandbox report read | Partly done via existing session detail | History/report partly wired | Draft update, roles API, questions/answers, status/progress, quota. |
| Learning derived diagnosis views | Partly done | Gap/recommendation/plan/report wired | Assessment submit/read, progress write/update, course materials, history. |
| CRM list/follow-up board | Partly done | CRM page partly wired | Batch import/update, richer reminders, deeper activity filters. |
| Growth derived views | Done for derived model views | Growth page partly wired; silent fallback business values removed | Saved snapshots, export/share, sensitivity analysis. |
| Leads result read model | Done for task detail/results | Leads page partly wired | Cancel/retry, CRM batch import, provider error details, deeper scoring. |
| Dashboard summary | Done | Dashboard page wired; silent fallback business rows removed | Period filters, source freshness, report generation, action deep links. |
| Competitor scan/monitoring skeleton | Partly done | Competitor data and monitoring pages wired; silent fallback competitor records removed | Scan progress/evidence, profiles, source health, watchlist/rules CRUD, export, cancel/retry. |

## 3. Whole-Domain Gaps

These routes/pages still have no real domain backend or only static UI:

| Domain/page | Current state | Needed API work |
| --- | --- | --- |
| GEO acquisition (`/geo`) | Overview reads are wired to Postgres tables; analysis requests support create, recent list, detail read, internal status updates, queue enqueue, and worker status processing across `queued/running/succeeded/failed`. Empty tables return explicit empty states. | Add analyzer execution, evidence generation, overview row writers, result freshness, and CRM handoff. |
| Enterprise companion (`/enterprise`) | Read-only overview endpoint is wired to Postgres tables for metrics, plans, delivery board items, milestones, and cases. Empty tables return explicit empty states. | Add admin/write APIs for plans, cases, inquiries, delivery workflow, milestones, and payment handoff. |
| Admin | Some content admin routes exist, but no admin UI/module depth | Membership/content/site config/user/task admin APIs and frontend admin routes. |

## 4. Important Missing Endpoints

Not exhaustive, but these are the current high-impact missing API groups:

| Domain | Missing endpoints/capabilities |
| --- | --- |
| Account | `PATCH /account/password`, account bindings/activity feed. |
| Tools | `POST /tools/recommendations`, `GET /tools/recommendation-plans/:id`. |
| Insights | Proper `/insights/:slug` route, filters, file analysis jobs/results. |
| Analysis | Competitor Tab B evidence job, generation progress/SSE if required, source references, better report section contract. |
| Projects | Opportunity catalog, cases, questions/answers, match results, unlock, compare, export. |
| Sandbox | Draft persistence, roles, questions/answers, status/progress, quota. |
| Learning | Assessments, progress writes, course materials, learning history. |
| Competitor | Scan progress/evidence, profiles, source health, watchlist/rules CRUD, export, cancel/retry. |
| GEO | Overview persistence plus analysis request create/list/detail, internal status updates, queue enqueue, and worker status processing exist; analyzer execution, evidence, CRM handoff, and admin/write APIs for overview rows are still missing. |
| Leads | Cancel/retry, batch CRM import, provider error detail, scoring depth. |
| Dashboard | Period filters, source freshness, report generation, action deep links. |
| Enterprise | Overview persistence exists; inquiry submission, delivery workflow, milestone/case management, payment handoff, and admin/write APIs are still missing. |
| Copilot | SSE streaming, binary upload/parsing, object storage, embeddings/RAG, file delete/update, thread search, quota checks. |
| Membership | Real payment, order cancel, benefits detail, admin management, service-level entitlement checks across expensive actions. |

## 5. Mock/Fallback Removal Queue

Use this order so that visible pages stop hiding backend problems first.

| Priority | Page/file group | Current issue | Action |
| --- | --- | --- | --- |
| P0 | Copilot tests | Baseline is now green. | Keep this green while removing fallback data. |
| P0 | First fallback-removal batch | Done for `TasksPage`, `GrowthCalculatorPage`, `DashboardPage`, `CompetitorDataPage`, and `CompetitorMonitoringPage`. | Keep these pages on backend data or explicit empty states only. |
| P0 | Remaining already backed pages | Need a second audit for implicit fallback not covered by old test names. | Check Membership, Messages, Profile, Help, Home, Tools, Community, Insights, CRM, Leads for static business rows that still render after backend failure or empty responses. |
| P1 | `GeoAcquisitionPage`, `EnterprisePage` | Entire pages are static business data. | Implement backend modules or mark pages as intentionally unavailable until APIs exist. |
| P1 | `CompetitorDataPage`, `CompetitorMonitoringPage`, `DashboardPage` | Silent fallback records removed; deeper APIs are still missing. | Add missing endpoints instead of static substitutions. |
| P1 | `ProjectsPage`, `SandboxPage`, `Learning*Page` | Partially wired, still retains static sections for missing depth. | Remove by feature slice, not file-wide sweep; each removal needs success/failure tests. |
| P1 | `LeadDevelopmentPage`, `CrmPage` | CRM/customer/follow-up/lead result sample records removed from runtime fallback; still needs deeper missing API work. | Continue with missing endpoint slices instead of static substitutions. |
| P2 | Navigation/category labels and UI chrome | Static arrays that are not business records. | Keep unless product wants these CMS-managed. These are not mock data blockers. |

## 6. Definition of Done for Removing a Mock

A mock/static business section is considered removed only when all items are true:

1. The page reads from a typed client in `apps/web/src/lib/*Api.ts`.
2. The backend route exists in `internal/*/http.go` and is registered in `internal/platform/httpserver/router.go`.
3. The endpoint is documented in `docs/api/backend-api-contract.md`.
4. Page tests cover success and failure behavior.
5. Backend failure does not silently render sample business records.
6. Empty backend data renders an explicit empty state, not fake customer/project/report data.
7. Relevant focused tests pass.

## 7. Recommended Next Execution Order

1. Run the second fallback audit on remaining already backed pages: Membership, Messages, Profile, Help, Home, Tools, Community, Insights, CRM, Leads.
2. Start a real backend slice for the largest static pages: `geo` and `enterprise`.
3. Then deepen existing workflow pages in this order: competitor, projects, sandbox, learning, dashboard.
4. Only after the above, add expensive provider-backed features: SSE, file parsing/RAG, exports, payment, entitlement enforcement.

## 8. Completed Fallback-Removal Batch 1

Completed on 2026-07-03:

| Page | Change |
| --- | --- |
| `TasksPage` | Removed static task rows and fake task stats from runtime fallback; backend failure/empty response now shows `暂无任务数据` and 0 stats. |
| `GrowthCalculatorPage` | Removed static standard-plan values, scenarios, forecast, cost items and action items from runtime fallback; empty/error state now shows `暂无测算模型` and section empty states. |
| `DashboardPage` | Removed static metrics, projects, trend, pipeline, alerts and actions from runtime fallback; empty/error state now shows explicit empty sections. |
| `CompetitorDataPage` | Removed static competitor cards and AI conclusions from runtime fallback; scan-derived stats now come from backend data or 0 values. |
| `CompetitorMonitoringPage` | Removed static watchlist, timeline events, high-risk alert count and next actions from runtime fallback; monitoring stats now derive from backend data or 0 values. |

Guardrail search:

```bash
rg -n "keeping fallback|fallback .*visible|shows backend load errors while keeping|shows backend list errors while keeping" apps/web/src/pages -S
```

Result: no matches after this batch.

## 9. Completed Fallback-Removal Batch 2 Slice A

Completed on 2026-07-03:

| Page | Change |
| --- | --- |
| `CrmPage` | Removed static CRM customer records, pipeline fallback stats, default timeline rows, follow-up records, reminder rows and recent-update rows from runtime fallback; empty customer/follow-up API responses now show explicit empty states and 0/count-derived stats. |
| `LeadDevelopmentPage` | Removed static lead companies, CRM summary numbers and recent follow-up rows from runtime fallback; empty lead task API responses now show explicit empty states for high-intent leads, CRM stats and CRM follow-ups. |

Guardrail search:

```bash
rg -n "星桥教育集团|成都蓝鲸|橙果职业培训|领航企业内训|12,845|1,426|leadCompanies|followRows|defaultTimelineRows|recentUpdates|reminders" apps/web/src/pages/CrmPage.tsx apps/web/src/pages/LeadDevelopmentPage.tsx
```

Result: no matches after this slice.

## 11. Completed Fallback-Removal Batch 2 Slice C

Completed on 2026-07-04:

| Page | Change |
| --- | --- |
| `ProfilePage` | Removed static quota records, profile/onboarding form fallback, account binding fallback, recent activity rows, content records, and prefilled profile-completion modal values. Empty account quotas, onboarding sections, bindings, and content now show explicit empty states instead of sample personal/company/project data. |
| `HelpPage` | Removed static help topic and support ticket fallback rows. Help topics, articles, and tickets now show backend data or explicit empty states. |

Guardrail search:

```bash
rg -n "AI智算额度|智活AI科技有限公司|生成智能客服系统机会分析|智能客服系统项目匹配|商业沙盘数据导出异常|增长测算结果与预期不符|FB202506|张婧|138 \\*\\*\\*\\* 5678|zhihuo_ai|2025-06-01" apps/web/src/pages/ProfilePage.tsx apps/web/src/pages/HelpPage.tsx
```

Result: no matches after this slice.

Next transparent blocker:

```bash
rg -n "const .* = \\[|fallback|张婧|智能AI眼镜|智能硬件市场分析报告|1,204|1,200|社群动态|本周活动|成员|评论|动态" apps/web/src/pages/{CommunityPage,CommunityMembersPage,CommunityEnterprisePage}.tsx apps/web/src/pages/{CommunityPage,CommunityMembersPage,CommunityEnterprisePage}.test.tsx
```

Result: `CommunityPage`, `CommunityMembersPage`, and `CommunityEnterprisePage` still contain static community posts, event rows, member/company counts, value stats, Copilot sample chat, and sample report files. These should be removed in the next slice or replaced with real community content/activity APIs.

## 12. Completed Fallback-Removal Batch 2 Slice D

Completed on 2026-07-04:

| Page | Change |
| --- | --- |
| `CommunityPage` | Removed static community post rows, activity/event rows, member/company counts, value statistics, testimonial quote, and hidden sample report. The page now keeps entry cards and growth path UI while showing explicit empty states for community dynamics, events, and value data. |
| `CommunityMembersPage` | Removed static member posts, events, stats, member counts, Copilot sample chat, and sample PDF report. Join request remains wired to `POST /api/v1/community/join-requests`; missing activity/report data now shows explicit empty states. |
| `CommunityEnterprisePage` | Removed static enterprise posts, events, stats, company/member counts, Copilot sample chat, and sample PDF report. Join request remains wired to `POST /api/v1/community/join-requests`; missing activity/report data now shows explicit empty states. |

Guardrail search:

```bash
rg -n "张婧|智能AI眼镜|智能硬件市场分析报告|1,204|1,200\\+|已加入 326|本周新增 86|社群动态.*\\[|memberPosts|enterprisePosts|weekEvents|enterpriseEvents|memberStats|enterpriseStats|communityPosts|valueStats|企业私域增长|实战经验：如何用AI|AI如何搭建私域|分享了智能AI眼镜" apps/web/src/pages/{CommunityPage,CommunityMembersPage,CommunityEnterprisePage}.tsx
```

Result: no matches after this slice.

Remaining community transparency gap: real activity/event/stat/report APIs do not exist yet. Current UI is intentionally empty instead of silently substituting sample rows.

## 13. Completed Static-Domain Transparency Slice A

Completed on 2026-07-04:

| Page | Change |
| --- | --- |
| `GeoAcquisitionPage` | Removed static GEO metrics, AI engine coverage rows, keyword opportunities, content tasks, and lead-signal records. The page now explicitly shows `GEO 后端接口未接入` plus empty states for AI search coverage, lead opportunities, keyword opportunities, and content tasks. |
| `EnterprisePage` | Removed static enterprise service metrics, package/pricing rows, delivery-board counts, milestone rows, and customer cases. The page now explicitly shows `企业陪跑后端接口未接入` plus empty states for plans, delivery board, milestones, and cases. |

Guardrail search:

```bash
rg -n "智能客服系统怎么选|ChatGPT|豆包|Kimi|通义千问|下游搜索量|教育培训机构如何用|增长团队训练营|AI获客陪跑|企业定制系统|连锁教育集团|职业培训机构|¥12万|¥18万|客服响应效率提升" apps/web/src/pages/GeoAcquisitionPage.tsx apps/web/src/pages/EnterprisePage.tsx
```

Result: no old sample business records remain after this slice. Generic UI metric labels such as `目标关键词` and `服务企业数` remain as zero/unavailable placeholders.

Remaining domain gap: real `geo` and `enterprise` backend modules still do not exist. These pages are now transparent placeholders until those modules are implemented.

## 14. Completed GEO/Enterprise Overview Wiring Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` | Added `internal/geo` with protected `GET /api/v1/geo/overview`, service tests, handler tests, and safe empty overview output when no repository is configured. |
| Backend `enterprise` | Added `internal/enterprise` with protected `GET /api/v1/enterprise/overview`, service tests, handler tests, and safe empty overview output when no repository is configured. |
| Router/API app | Registered both handlers in `internal/platform/httpserver/router.go` and `apps/api/main.go`. |
| Frontend clients | Added typed `geoApi.overview()` and `enterpriseApi.overview()` clients with tests. |
| Frontend pages | Wired `GeoAcquisitionPage` and `EnterprisePage` to the new clients. Empty responses keep explicit empty states; populated API responses render backend records. |
| API docs | Documented both overview endpoints in `docs/api/backend-api-contract.md`. |

Current limitation:

These are read-only endpoints. They intentionally do not create sample data. `geo` and `enterprise` now both have Postgres repositories for overview reads. The next real backend slice should add either GEO write/job/evidence APIs or enterprise plan/case/inquiry write APIs.

## 15. Completed GEO Overview Persistence Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` repository | Added `NewPostgresRepository` and DB-backed overview reads for metrics, AI engine coverage, lead signals, keyword opportunities, and content tasks. Empty result sets are normalized to empty arrays. |
| Migrations | Added `000022_geo_overview` tables and user-scoped indexes. No sample/seed business data is inserted. |
| API app wiring | Changed `apps/api/main.go` to pass the Postgres repository into `geo.NewService`, so `GET /api/v1/geo/overview` reads from database tables. |
| API docs | Updated the GEO overview contract from skeleton-only to Postgres-backed read model. |

Current limitation:

GEO remains read-only. Admin/write endpoints, job execution, evidence capture, result freshness, and CRM handoff are still missing.

## 16. Completed Enterprise Overview Persistence Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `enterprise` repository | Added `NewPostgresRepository` and DB-backed overview reads for metrics, plans, delivery board items, milestones, and cases. Empty result sets are normalized to empty arrays, and plan focus arrays are normalized by the service. |
| Migrations | Added `000023_enterprise_overview` tables and user-scoped indexes. No sample/seed business data is inserted. |
| API app wiring | Changed `apps/api/main.go` to pass the Postgres repository into `enterprise.NewService`, so `GET /api/v1/enterprise/overview` reads from database tables. |
| API docs | Updated the Enterprise overview contract from skeleton-only to Postgres-backed read model. |

Current limitation:

Enterprise remains read-only. Inquiry submission, admin/write APIs, delivery workflow state changes, milestone management, case management, and payment handoff are still missing.

## 17. Completed GEO Analysis Request Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` API | Added protected `POST /api/v1/geo/analysis-requests` with input validation, service tests, handler tests, and repository tests. |
| Backend `geo` repository | Added insertion into `geo_analysis_requests` with initial `queued` status. The endpoint does not generate or insert mock overview rows. |
| Migrations | Added `000024_geo_analysis_requests` table and indexes for user history and queued job processing. |
| Frontend client/page | Added `geoApi.createAnalysisRequest()` and wired the GEO page form to submit real requests, display success/error state, and keep overview empty until real result data exists. |
| API docs | Documented the request endpoint and clarified that it only queues work for future processing. |

Current limitation:

GEO analysis requests are stored, visible in the page history, readable by ID, and can be updated internally to `queued/running/succeeded/failed`, but they are not processed by a worker yet. Worker execution, evidence capture, generated keyword/content/lead records, freshness metadata, and CRM handoff are still missing.

## 18. Completed GEO Analysis Request History Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` API | Added protected `GET /api/v1/geo/analysis-requests` with limit validation and tests. |
| Backend `geo` repository | Added user-scoped request listing from `geo_analysis_requests`, ordered newest first, with empty arrays for no data. |
| Frontend client/page | Added `geoApi.listAnalysisRequests()` and rendered recent request history on `GeoAcquisitionPage`. Submitting a request now inserts the returned queued request at the top of the visible history. |
| API docs | Documented the request listing endpoint and clarified remaining processing gaps. |

Current limitation:

The history is read-only and reflects queued/stored requests only. Request detail, status transitions, worker processing, generated evidence, generated overview rows, and CRM handoff remain missing.

## 19. Completed GEO Analysis Request Detail Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` API | Added protected `GET /api/v1/geo/analysis-requests/{id}` with ID validation, ownership scoping, 404 mapping, and tests. |
| Backend `geo` repository | Added `GetAnalysisRequest` using `user_id` plus request ID, mapping missing rows to `analysis_request_not_found`. |
| Frontend client | Added typed `geoApi.getAnalysisRequest(id)` with tests. |
| API docs | Documented the detail endpoint and clarified that it currently returns request metadata only. |

Current limitation:

Request detail does not include generated evidence or output sections yet. Status transitions, worker processing, generated overview rows, result freshness metadata, and CRM handoff remain missing.

## 20. Completed GEO Internal Status Transition Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` service | Added internal `UpdateAnalysisRequestStatus` validation for `queued`, `running`, `succeeded`, and `failed`. Invalid IDs/statuses are rejected before repository calls. |
| Backend `geo` repository | Added `UPDATE geo_analysis_requests SET status, error_message, updated_at` with `RETURNING` and not-found mapping. Existing request reads now include `error_message`. |
| Frontend types | Added optional `error_message` to `GeoAnalysisRequest` so list/detail responses match the backend contract. |
| API docs | Documented request statuses and the optional error field in request responses. |

Current limitation:

Status updates are internal only. Worker registration is now handled by the next slice; analyzer execution, evidence/overview row generation, and CRM handoff are still missing.

## 21. Completed GEO Worker Queue Slice

Completed on 2026-07-04:

| Area | Change |
| --- | --- |
| Backend `geo` service | `CreateAnalysisRequest` can now enqueue a `geo.analysis` job when a queue is configured. The API app wires GEO to the existing taskqueue client. |
| Backend `geo` worker | Added `internal/geo/worker.go` and registered it in `apps/worker/main.go`, so `geo.analysis` tasks call `ProcessAnalysisRequest`. |
| Job infrastructure | Added `jobs.TypeGeoAnalysis` and included it in the default taskqueue registry. |
| Status processing | Worker processing marks a request `running`; without a configured analyzer it marks the request `failed` with `error_message: "analyzer_not_configured"` instead of leaving it silently queued or generating fake rows. |
| API docs | Updated the GEO contract to document queue behavior and the current transparent failure mode. |

Current limitation:

No GEO analyzer/provider is configured yet. The worker does not generate evidence, keyword opportunities, content tasks, lead signals, freshness metadata, or CRM handoff records.

## 10. Completed Fallback-Removal Batch 2 Slice B

Completed on 2026-07-04:

| Page | Change |
| --- | --- |
| `ToolsPage` | Removed static tool library records, default tool detail fallback, static recommendation results, static recommendation plan, and static AI recommendation result rows. Tool library/detail/recommend/plan now show backend data or explicit empty states. |
| `InsightsPage` | Removed static article list fallback, default article detail body, static related recommendations, and fixed-count focus data. Article list/detail now show backend data or explicit empty states. |
| `HomePage` | Removed signed-in dashboard fallback recommendations, task rows, notification rows, and fake account/metric defaults. Empty home summary sections now show explicit empty states and 0/count-derived stats. |

Guardrail search:

```bash
rg -n "Canva AI|DeepSeek|Google Analytics|SimilarWeb|Midjourney|ChatGPT|Notion AI|Perplexity|Runway|Jasper|项目推荐|完成【AI 智能硬件】|竞品价格监测数据|企业版年中|企业智能客服落地实践|企业智能客服进入规模化|更新 12 条" apps/web/src/pages/ToolsPage.tsx apps/web/src/pages/InsightsPage.tsx apps/web/src/pages/HomePage.tsx
```

Result: no matches after this slice.
