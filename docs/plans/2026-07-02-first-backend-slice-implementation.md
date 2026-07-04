# First Backend Slice Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the first low-risk backend slice for account, notifications, membership catalog/usage, content/help, support tickets, and home aggregation so static account/content pages can start using real APIs.

**Architecture:** Keep the current Go modular monolith. Add new domain packages only where ownership is clear (`account`, `notifications`, `support`, `home`), extend existing packages where the data already belongs (`membership`, `content`), and wire everything through `internal/platform/httpserver/router.go` with auth middleware. Each domain owns its repository/service/http tests and SQL migrations.

**Tech Stack:** Go, Gin, pgx, PostgreSQL SQL migrations, React + TypeScript clients, Vitest/Testing Library, existing `apiRequest` wrapper.

---

## Preconditions

- Review and accept the planned contracts in `docs/api/backend-api-contract.md`.
- Keep the page gap design doc as the source for scope: `docs/plans/2026-07-02-page-backend-gap-analysis-and-architecture.md`.
- Use migration numbers starting at `000019` because `000018_copilot_files` is current.
- Do not implement payment, email/SMS delivery, binary file upload, RAG, or admin UI in this slice.

## Task 1: Add Account Domain

Status: Mostly completed on 2026-07-02. Implemented `internal/account`,
`000019_account_notifications`, protected account routes, API wiring, and tests.
Remaining: `PATCH /account/password` needs auth-service password update wiring.

**Files:**
- Create: `migrations/000019_account_notifications.up.sql`
- Create: `migrations/000019_account_notifications.down.sql`
- Create: `internal/account/types.go`
- Create: `internal/account/service.go`
- Create: `internal/account/service_test.go`
- Create: `internal/account/postgres_repository.go`
- Create: `internal/account/postgres_repository_test.go`
- Create: `internal/account/http.go`
- Create: `internal/account/http_test.go`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `apps/api/main.go`

**Step 1: Write repository tests**

Cover:
- `GetProfile` returns auth user fields plus optional profile fields.
- `UpdateProfile` trims strings and preserves immutable auth fields.
- `GetOnboarding`, `SaveOnboarding`, and `CompleteOnboarding` round-trip JSON sections.
- `GetPreferences` creates defaults when missing.
- `UpdatePreferences` only changes supplied fields.
- `GetQuotas` returns an empty list first; real quota aggregation can be filled by membership later.
- `ListContent` returns empty list first; later aggregation can be expanded.

**Step 2: Add migration**

Create:
- `user_profiles`
- `user_preferences`
- `user_onboarding`

Keep profile data separate from auth-sensitive fields in `users`.

**Step 3: Write service tests**

Cover:
- Reject missing user id.
- Validate optional email format.
- Trim editable strings.
- Delegate password update through a narrow auth-owned interface, or return a clear `ErrPasswordUnsupported` until auth integration is implemented.
- Account deletion returns `pending_deletion` without hard-deleting rows in this slice.

**Step 4: Write HTTP tests**

Cover:
- Protected routes require auth in router integration.
- `GET /account/profile`
- `PATCH /account/profile`
- `GET /account/onboarding`
- `PUT /account/onboarding`
- `POST /account/onboarding/complete`
- `GET/PATCH /account/preferences`
- `GET /account/quotas`
- `GET /account/content`
- `DELETE /account`

**Step 5: Implement account package**

Follow existing package shape from `tasks`, `growth`, and `crm`:
- `Application` interface in `http.go`.
- Service owns validation and defaults.
- Repository owns SQL and JSONB round-trips.
- HTTP handler maps known errors to stable codes.

**Step 6: Wire API**

Add:
- `account` import in `apps/api/main.go` and router.
- `Account *account.HTTPHandler` in `httpserver.Handlers`.
- Protected registration block in `NewRouter`.

**Step 7: Verify**

Run:

```bash
go test ./internal/account ./internal/platform/httpserver
```

Expected: PASS.

## Task 2: Add Notifications Domain

Status: Completed on 2026-07-02. Implemented `internal/notifications`, extended
`000019_account_notifications`, protected notification routes, API wiring, and tests.

**Files:**
- Modify: `migrations/000019_account_notifications.up.sql`
- Modify: `migrations/000019_account_notifications.down.sql`
- Create: `internal/notifications/types.go`
- Create: `internal/notifications/service.go`
- Create: `internal/notifications/service_test.go`
- Create: `internal/notifications/postgres_repository.go`
- Create: `internal/notifications/postgres_repository_test.go`
- Create: `internal/notifications/http.go`
- Create: `internal/notifications/http_test.go`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `apps/api/main.go`

**Step 1: Write repository tests**

Cover:
- List by user with `type`, `status`, and `limit`.
- Get by id is user-scoped.
- Mark one notification read is idempotent.
- Mark all read returns affected count.
- Summary returns total unread, counts by type, and latest items.

**Step 2: Extend migration**

Create `notifications`:
- `id BIGSERIAL PRIMARY KEY`
- `user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `type TEXT NOT NULL`
- `title TEXT NOT NULL`
- `summary TEXT NOT NULL DEFAULT ''`
- `body TEXT NOT NULL DEFAULT ''`
- `source_type TEXT NOT NULL DEFAULT ''`
- `source_id BIGINT`
- `action_label TEXT NOT NULL DEFAULT ''`
- `action_url TEXT NOT NULL DEFAULT ''`
- `read_at TIMESTAMPTZ`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Indexes:
- `(user_id, created_at DESC)`
- `(user_id, read_at, created_at DESC)`
- `(user_id, type, created_at DESC)`

**Step 3: Write service and HTTP tests**

Cover:
- Invalid type/status filters return `invalid_request`.
- Invalid ids return `invalid_notification_id`.
- Not-owned ids return `notification_not_found`.
- Empty lists return `[]`.

**Step 4: Implement package and route wiring**

Add protected endpoints:
- `GET /notifications`
- `GET /notifications/:id`
- `PATCH /notifications/:id/read`
- `POST /notifications/read-all`
- `GET /notifications/summary`

**Step 5: Verify**

Run:

```bash
go test ./internal/notifications ./internal/platform/httpserver
```

Expected: PASS.

## Task 3: Extend Membership Plans, Usage, Orders, Checkout

Status: Completed on 2026-07-02. Implemented membership catalog, usage, orders,
manual checkout, `000020_membership_catalog_orders`, HTTP routes, and tests.

**Files:**
- Create: `migrations/000020_membership_catalog_orders.up.sql`
- Create: `migrations/000020_membership_catalog_orders.down.sql`
- Modify: `internal/membership/service.go`
- Modify: `internal/membership/service_test.go`
- Modify: `internal/membership/postgres_repository.go`
- Modify: `internal/membership/postgres_repository_test.go`
- Modify: `internal/membership/http.go`
- Modify: `internal/membership/http_test.go`
- Modify: `docs/api/backend-api-contract.md` after implementation status changes.

**Step 1: Write tests for new service methods**

Add methods:
- `ListPlans(ctx)`
- `CurrentUsage(ctx, userID)`
- `ListOrders(ctx, userID, limit)`
- `CreateCheckout(ctx, CheckoutInput)`

Cover:
- Plans come from database seed rows.
- Usage returns stable empty/default rows when no usage records exist.
- Orders are user-scoped and sorted newest first.
- Checkout validates active plan and creates a manual pending order.

**Step 2: Add migration**

Create:
- `membership_plans`
- `membership_plan_quotas`
- `membership_usage`
- `membership_orders`

Seed at least:
- `free`
- `pro`

Keep existing `PlanCatalog` temporarily for backward compatibility, but prefer DB rows in new list/checkout APIs.

**Step 3: Extend repository**

Add SQL methods matching the service methods. Keep current redemption and snapshot behavior intact.

**Step 4: Extend HTTP handler**

Add protected endpoints:
- `GET /membership/plans`
- `GET /membership/usage`
- `GET /membership/orders`
- `POST /membership/checkout`

Do not change the behavior of:
- `GET /membership/me`
- `POST /redemptions/redeem`

**Step 5: Verify**

Run:

```bash
go test ./internal/membership ./internal/platform/httpserver
```

Expected: PASS.

## Task 4: Extend Content for Tools, Community, and Help

Status: Completed on 2026-07-02. Implemented tool filters/detail/favorites,
community join requests, help topics/articles, `000021_content_help_support_extensions`,
HTTP routes, and tests.

**Files:**
- Create: `migrations/000021_content_help_support_extensions.up.sql`
- Create: `migrations/000021_content_help_support_extensions.down.sql`
- Modify: `internal/content/types.go`
- Modify: `internal/content/service.go`
- Modify: `internal/content/service_test.go`
- Modify: `internal/content/postgres_repository.go`
- Modify: `internal/content/postgres_repository_test.go`
- Modify: `internal/content/http.go`
- Modify: `internal/content/http_test.go`
- Modify: `docs/api/backend-api-contract.md` after implementation status changes.

**Step 1: Write tests for tool filters and detail**

Cover:
- `ListTools` accepts category, search query, sort, and limit.
- Existing `GET /content/tools` remains compatible.
- `GetTool(slug)` returns `tool_not_found` for missing slug.
- Favorite/unfavorite is user-scoped and idempotent.

**Step 2: Write tests for community join requests**

Cover:
- Valid `members` and `enterprise` communities.
- Empty contact returns `invalid_request`.
- Submitted request is user-scoped.

**Step 3: Write tests for help articles**

Cover:
- `GET /help/topics`
- `GET /help/articles?topic=&q=&limit=`
- `GET /help/articles/:slug`
- Empty lists return `[]`.

**Step 4: Add migration**

Extend or create:
- `content_tool_favorites`
- `community_join_requests`
- `help_topics`
- `help_articles`

Seed a small help-topic set matching the Help page categories.

**Step 5: Implement service/repository/http**

Public endpoints:
- `GET /content/tools?category=&q=&sort=&limit=`
- `GET /content/tools/:slug`
- `GET /help/topics`
- `GET /help/articles`
- `GET /help/articles/:slug`

Protected endpoints:
- `POST /content/tools/:slug/favorite`
- `DELETE /content/tools/:slug/favorite`
- `POST /community/join-requests`

**Step 6: Verify**

Run:

```bash
go test ./internal/content ./internal/platform/httpserver
```

Expected: PASS.

## Task 5: Add Support Tickets Domain

Status: Completed on 2026-07-02. Implemented `internal/support`, support ticket
create/list routes, shared `000021_content_help_support_extensions` migration,
API wiring, and tests.

**Files:**
- Modify: `migrations/000021_content_help_support_extensions.up.sql`
- Modify: `migrations/000021_content_help_support_extensions.down.sql`
- Create: `internal/support/types.go`
- Create: `internal/support/service.go`
- Create: `internal/support/service_test.go`
- Create: `internal/support/postgres_repository.go`
- Create: `internal/support/postgres_repository_test.go`
- Create: `internal/support/http.go`
- Create: `internal/support/http_test.go`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `apps/api/main.go`

**Step 1: Write tests**

Cover:
- Create ticket validates title/body.
- List tickets is user-scoped.
- Empty ticket list returns `[]`.
- Limit validation follows `httpapi.QueryLimit`.

**Step 2: Extend migration**

Create `support_tickets`:
- `id BIGSERIAL PRIMARY KEY`
- `user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `topic TEXT NOT NULL DEFAULT ''`
- `title TEXT NOT NULL`
- `body TEXT NOT NULL`
- `status TEXT NOT NULL DEFAULT 'open'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Index `(user_id, created_at DESC)`.

**Step 3: Implement package and protected routes**

Add:
- `POST /support/tickets`
- `GET /support/tickets`

**Step 4: Verify**

Run:

```bash
go test ./internal/support ./internal/platform/httpserver
```

Expected: PASS.

## Task 6: Add Home Aggregation Endpoint

Status: Completed on 2026-07-02. Implemented `internal/home`, protected
`GET /home/summary`, API wiring, dependency-degraded summary behavior, and tests.

**Files:**
- Create: `internal/home/types.go`
- Create: `internal/home/service.go`
- Create: `internal/home/service_test.go`
- Create: `internal/home/http.go`
- Create: `internal/home/http_test.go`
- Modify: `internal/platform/httpserver/router.go`
- Modify: `internal/platform/httpserver/health_test.go`
- Modify: `apps/api/main.go`

**Step 1: Write service tests**

Start with simple dependency interfaces:
- Notification summary reader.
- Membership/account summary reader.
- Recent task reader.

Cover:
- Empty dependencies return safe empty arrays.
- Partial dependency failures degrade the affected section instead of failing the whole response.
- User id is required.

**Step 2: Implement HTTP endpoint**

Add:
- `GET /home/summary`

Response must match the planned contract:
- `hero_cards`
- `recommendations`
- `recent_tasks`
- `notification_summary`
- `account_summary`

**Step 3: Wire dependencies conservatively**

First version can:
- Use notifications summary.
- Use membership current snapshot/usage for account summary.
- Use tasks list for recent tasks if a narrow interface is easy.
- Return empty hero/recommendations until product copy is CMS-backed.

**Step 4: Verify**

Run:

```bash
go test ./internal/home ./internal/platform/httpserver
```

Expected: PASS.

## Task 7: Frontend API Clients

Status: Done.

**Files:**
- Create: `apps/web/src/lib/accountApi.ts`
- Create: `apps/web/src/lib/accountApi.test.ts`
- Create: `apps/web/src/lib/notificationsApi.ts`
- Create: `apps/web/src/lib/notificationsApi.test.ts`
- Create: `apps/web/src/lib/supportApi.ts`
- Create: `apps/web/src/lib/supportApi.test.ts`
- Modify: `apps/web/src/lib/membershipApi.ts`
- Modify: `apps/web/src/lib/membershipApi.test.ts`
- Modify: `apps/web/src/lib/contentApi.ts`
- Modify: `apps/web/src/lib/contentApi.test.ts`
- Create: `apps/web/src/lib/homeApi.ts`
- Create: `apps/web/src/lib/homeApi.test.ts`

**Step 1: Add typed clients from the contract**

Follow existing `apiRequest` conventions:
- No raw `fetch` in pages.
- Keep response types close to API contract names.
- Encode query params safely with `URLSearchParams`.

**Step 2: Test request paths and methods**

For every new client method, assert:
- Path.
- HTTP method.
- Body shape for writes.
- Query string for list/filter methods.

**Step 3: Verify**

Run:

```bash
npm test -- --run apps/web/src/lib/accountApi.test.ts apps/web/src/lib/notificationsApi.test.ts apps/web/src/lib/membershipApi.test.ts apps/web/src/lib/contentApi.test.ts apps/web/src/lib/supportApi.test.ts apps/web/src/lib/homeApi.test.ts
```

Actual:

```bash
cd apps/web
npm test -- --run src/lib/accountApi.test.ts src/lib/notificationsApi.test.ts src/lib/membershipApi.test.ts src/lib/contentApi.test.ts src/lib/supportApi.test.ts src/lib/homeApi.test.ts
npm test -- --run
npm run lint
npm run build
```

Result: PASS. Build emits the existing Vite chunk-size warning for the main bundle.

## Task 8: Wire First Pages to Real APIs

**Files:**
- Modify: `apps/web/src/pages/MembershipPage.tsx`
- Modify: `apps/web/src/pages/MembershipPage.test.tsx`
- Modify: `apps/web/src/pages/MessagesPage.tsx`
- Modify: `apps/web/src/pages/MessagesPage.test.tsx`
- Modify: `apps/web/src/pages/ProfilePage.tsx`
- Modify: `apps/web/src/pages/ProfilePage.test.tsx`
- Modify: `apps/web/src/pages/HelpPage.tsx`
- Modify: `apps/web/src/pages/HelpPage.test.tsx`
- Modify: `apps/web/src/pages/HomePage.tsx`
- Modify: `apps/web/src/pages/HomePage.test.tsx`
- Modify: `apps/web/src/pages/ToolsPage.tsx`
- Modify: `apps/web/src/pages/ToolsPage.test.tsx`
- Modify: `apps/web/src/pages/CommunityPage.tsx`
- Modify: `apps/web/src/pages/CommunityPage.test.tsx`
- Modify: `apps/web/src/pages/CommunityMembersPage.tsx`
- Modify: `apps/web/src/pages/CommunityMembersPage.test.tsx`
- Modify: `apps/web/src/pages/CommunityEnterprisePage.tsx`
- Modify: `apps/web/src/pages/CommunityEnterprisePage.test.tsx`
- Modify: `apps/web/src/pages/InsightsPage.tsx`
- Modify: `apps/web/src/pages/InsightsPage.test.tsx`

**Step 1: MembershipPage**

Status: Done.

Wire:
- Current membership.
- Plans.
- Usage.
- Orders.
- Redemption.
- Checkout manual pending response.

Keep fallback only for explicit empty states, not for backend failures.

Actual:

- Purchase page loads current membership, plans, usage and orders through `membershipApi`.
- Non-current plan actions create manual checkout orders and prepend the pending order to the table.
- Redeem page submits codes through `membershipApi.redeem` and refreshes the account summary.
- Page tests cover API loading, checkout and redeem flows.

**Step 2: MessagesPage**

Status: Done.

Wire:
- List notifications.
- Detail by id.
- Mark read.
- Read-all.

Display backend errors visibly while preserving page structure.

Actual:

- List page loads notifications and summary through `notificationsApi`.
- Message links now use backend notification ids (`/messages/:id`).
- Detail page loads by id and marks unread notifications as read.
- Read-all action calls the backend and updates local unread state.

**Step 3: ProfilePage**

Status: Done.

Wire:
- Profile.
- Preferences.
- Quotas.
- Account content.
- Onboarding completion prompt where relevant.

Password/delete overlays may call APIs but can show unsupported/pending if backend returns planned statuses.

Actual:

- Profile overview loads profile and quota data through `accountApi`.
- Settings page renders profile bindings and onboarding sections from the backend.
- My Content loads account content records from the backend.
- Preferences page reads and saves backend preferences.
- Delete account overlay submits `accountApi.deleteAccount`; password remains UI-only because `PATCH /account/password` is still not implemented.

**Step 4: HelpPage**

Status: Done.

Wire:
- Help topics.
- Help article list.
- Support ticket create/list.

Actual:

- Help topics and help article list load through `contentApi`.
- Existing tickets load through `supportApi.listTickets`.
- Feedback form submits through `supportApi.createTicket` and prepends the created ticket to the list.

**Step 5: HomePage**

Status: Done.

Wire:
- Home summary.
- Notifications summary.
- Account summary.
- Assistant files can keep using `copilotApi.listFiles`.

Actual:

- Signed-in home loads `homeApi.summary`.
- Hero cards, recommendations, recent tasks, notification summary and account summary render from the backend summary payload.
- Account menu and the welcome panel show plan, credit balance and quota warnings from account summary.
- Notification dropdown links to backend-provided action URLs when available.
- Copilot utility UI keeps the existing static/file toggle behavior for this slice.

**Step 6: Content pages**

Status: Done.

Wire:
- `ToolsPage` filters/detail/favorite basics.
- Community pages to extended community config and join requests.
- Insights list/detail/bookmark basics.

Actual:

- `ToolsPage` loads tools with category/search/sort filters, reads tool detail by slug, and calls favorite/unfavorite APIs.
- `CommunityPage`, `CommunityMembersPage`, and `CommunityEnterprisePage` read community config through `contentApi.getCommunityConfig`.
- Member and enterprise community pages submit join requests through `contentApi.joinCommunity`.
- `InsightsPage` reads article lists and detail records from content APIs.
- Article bookmark/unbookmark support was added to backend content APIs, frontend `contentApi`, migrations, and Insights detail UI.

**Step 7: Verify page tests**

Run focused tests first:

```bash
npm test -- --run src/pages/MembershipPage.test.tsx src/pages/MessagesPage.test.tsx src/pages/ProfilePage.test.tsx src/pages/HelpPage.test.tsx src/pages/HomePage.test.tsx src/pages/ToolsPage.test.tsx src/pages/CommunityPage.test.tsx src/pages/CommunityMembersPage.test.tsx src/pages/CommunityEnterprisePage.test.tsx src/pages/InsightsPage.test.tsx
```

Then:

```bash
npm test -- --run
npm run lint
npm run build
```

Expected: PASS.

## Task 9: Full Backend Verification

**Files:**
- No new files unless a failing integration test reveals a missing route registration test.

**Step 1: Run package tests**

Run:

```bash
go test ./internal/account ./internal/notifications ./internal/membership ./internal/content ./internal/support ./internal/home ./internal/platform/httpserver
```

Expected: PASS.

**Step 2: Run full Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

**Step 3: Run migration smoke**

Start local Postgres if needed, then run the API with auto migrations or the repo's migration command. Confirm current version reaches `21`.

Expected:
- `000019_account_notifications`
- `000020_membership_catalog_orders`
- `000021_content_help_support_extensions`

Actual:

- `go test ./internal/account ./internal/notifications ./internal/membership ./internal/content ./internal/support ./internal/home ./internal/platform/httpserver`: PASS.
- `go test ./...`: PASS.
- Migration smoke on local `opcv2-postgres-1`/`opcv2-redis-1`: API auto migration reached `current_version=21` with `applied_count=3`.

## Task 10: Documentation Cleanup

**Files:**
- Modify: `docs/api/backend-api-contract.md`
- Modify: `docs/plans/2026-07-02-page-backend-gap-analysis-and-architecture.md`
- Modify: `docs/plans/2026-07-02-first-backend-slice-implementation.md`

**Step 1: Update API statuses**

After implementation lands:
- Remove `Status: Planned, not implemented` from endpoints that are now wired.
- Keep future-only endpoints marked planned.

**Step 2: Update gap matrix**

Mark first-slice pages as partially or fully backed by APIs:
- Home
- Membership
- Messages
- Profile
- Help
- Tools
- Community
- Insights

**Step 3: Record verification**

Add the final commands and results to the implementation PR or final task summary.

Actual:

- API contract statuses updated: implemented first-slice endpoints are marked implemented; tool recommendations remain planned.
- Page gap matrix updated for Home, Membership, Messages, Profile, Help, Tools, Community, and Insights.
- Verification results recorded in Task 9 and final task summary.

## Commit Strategy

Recommended commits:

1. `docs: plan first backend slice`
2. `feat: add account and notifications APIs`
3. `feat: extend membership catalog APIs`
4. `feat: extend content help and support APIs`
5. `feat: add home summary API`
6. `feat: wire account content pages to backend`

Do not combine all backend domains and frontend rewiring into one commit unless explicitly requested.
