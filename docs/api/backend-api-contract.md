# Backend API Contract

This document tracks backend endpoints for the new web modules.

Sections without an explicit status are implemented and wired. Sections marked
`Planned` are accepted contract drafts for upcoming implementation and must not
be treated as available until the matching backend handlers and tests land.

## Common Rules

- Base path: `/api/v1`
- Protected endpoints require `Authorization: Bearer <access_token>`.
- Public endpoints do not require an access token.
- Request and response bodies are JSON.
- Error response shape is stable:

```json
{
  "error": "invalid_request"
}
```

- List endpoints that accept `limit` default to `20`.
- `limit <= 0` or non-numeric `limit` returns `400 {"error":"invalid_limit"}`.
- `limit > 100` is capped to `100`.
- Empty list responses return empty arrays, not `null`.

## Error Codes

| Code | Status | Meaning |
| --- | ---: | --- |
| `invalid_request` | 400 | JSON body is invalid or required fields are missing/invalid. |
| `invalid_limit` | 400 | Query `limit` is non-numeric or less than 1. |
| `invalid_*_id` | 400 | Path id is non-numeric or less than 1. |
| `*_not_found` | 404 | Resource does not exist or does not belong to the authenticated user. |
| `service_not_ready` | 500 | Backend dependency is not configured. |
| `internal_error` | 500 | Unclassified server error. |
| `invalid_ai_result` | 500 | AI output failed backend validation. |
| `quota_exceeded` | 402 | Authenticated user has exhausted the configured membership quota for this action. |
| `quota_not_configured` | 500 | Backend quota config is missing for a gated action. |

## Sandbox

All sandbox endpoints are protected.

### Create Session

`POST /api/v1/sandbox/sessions`

Request:

```json
{
  "goal": "验证 AI 客服项目",
  "target_users": "本地教培机构",
  "product": "AI 客服工具",
  "roles": ["用户", "投资人"]
}
```

Validation:

- `goal`, `target_users`, and `product` must be non-empty after trimming.
- `roles` must contain at least one non-empty role.
- Client-supplied `user_id` is ignored.

Response `200`: `SandboxSession`

```json
{
  "id": 99,
  "user_id": 42,
  "goal": "验证 AI 客服项目",
  "target_users": "本地教培机构",
  "product": "AI 客服工具",
  "roles": ["用户", "投资人"],
  "status": "draft",
  "report": {
    "score": 0,
    "summary": "",
    "metrics": [],
    "role_summaries": [],
    "risks": [],
    "next_actions": []
  },
  "created_at": "2026-06-30T10:00:00Z",
  "updated_at": "2026-06-30T10:00:00Z"
}
```

### Run Session

`POST /api/v1/sandbox/sessions/{id}/run`

Consumes membership quota key `sandbox_runs`. Defaults seeded by migrations:
free users get 1 run/month, pro users get 20 runs/month.

Response `200`: `SandboxSession` with `status: "completed"` and populated `report`.

Errors:

- `400 invalid_session_id`
- `404 session_not_found`
- `402 quota_exceeded`
- `500 invalid_ai_result`
- `500 quota_not_configured`

### List Sessions

`GET /api/v1/sandbox/sessions?limit=20`

Response:

```json
{
  "sessions": []
}
```

### Get Session

`GET /api/v1/sandbox/sessions/{id}`

Response `200`: `SandboxSession`

Errors:

- `400 invalid_session_id`
- `404 session_not_found`

## Projects

All project endpoints are protected.

Project match status values:

- `needs_input`
- `completed`

### Create Project Match

`POST /api/v1/projects/matches`

Request:

```json
{
  "intent": "我擅长内容创作，预算3万以内，每周能投入20小时，希望做一人公司线上项目",
  "answers": [
    { "key": "budget", "value": "1-3万" }
  ]
}
```

Response `200`: `ProjectMatchResult`

Notes:

- If required context is missing, response status is `needs_input` and
  `questions` contains follow-up questions.
- If context is sufficient, response status is `completed` and `projects`
  contains ranked matches.

### List Project Matches

`GET /api/v1/projects/matches?limit=20`

Response `200`:

```json
{
  "matches": []
}
```

### Get Project Match

`GET /api/v1/projects/matches/{id}`

Response `200`: `ProjectMatchSession`

Errors:

- `400 invalid_match_id`
- `404 match_not_found`

### Favorite Project Match

`POST /api/v1/projects/matches/{id}/favorite`

Response `200`: `ProjectFavorite`

Errors:

- `400 invalid_match_id`
- `404 match_not_found`

## Tasks

All task endpoints are protected.

Task enums:

- `status`: `todo`, `in_progress`, `completed`, `reminder`
- `priority`: `low`, `medium`, `high`

### Create Task

`POST /api/v1/tasks`

Request:

```json
{
  "title": "整理客户名单",
  "project": "AI线索开发",
  "priority": "high",
  "due_at": "2026-06-30T12:00:00Z",
  "tools": ["CRM"],
  "learning": "线索评分"
}
```

Validation:

- `title` and `project` must be non-empty after trimming.
- `priority` must be one of the task priority enum values.
- Client-supplied `user_id` is ignored.

Response `200`: `Task`

### List Tasks

`GET /api/v1/tasks?status=in_progress&project=商业沙盘&q=接口&limit=20`

Query:

- `status` optional task status enum.
- `project` optional exact project name.
- `q` optional keyword matched against title, project and learning field.
- `limit` optional, capped at 100.

Response:

```json
{
  "tasks": []
}
```

### Task Stats

`GET /api/v1/tasks/stats`

Response `200`:

```json
{
  "total": 12,
  "todo": 3,
  "in_progress": 5,
  "completed": 2,
  "reminder": 2,
  "overdue": 1
}
```

### Get Task

`GET /api/v1/tasks/{id}`

Response `200`: `Task`

Errors:

- `400 invalid_task_id`
- `404 task_not_found`

### Update Task

`PATCH /api/v1/tasks/{id}`

Request fields are optional:

```json
{
  "title": "整理客户名单",
  "project": "AI线索开发",
  "status": "completed",
  "priority": "high",
  "due_at": "2026-06-30T12:00:00Z",
  "tools": ["CRM"],
  "learning": "线索评分"
}
```

Validation:

- If present, `title` and `project` must be non-empty after trimming.
- If present, `status` and `priority` must match their enum values.

Response `200`: `Task`

## Dashboard

Dashboard endpoints are protected.

### Summary

`GET /api/v1/dashboard/summary`

Response:

```json
{
  "metrics": [{ "label": "本月线索", "value": "128", "change": "+18%" }],
  "projects": [{ "name": "AI线索开发", "value": "42", "leads": "128", "stage": "增长中" }],
  "trend": [{ "label": "周一", "value": 12 }],
  "pipeline": [{ "stage": "qualified", "count": "24", "percent": "38%" }],
  "alerts": [{ "title": "待跟进客户", "detail": "有客户需要今天跟进" }],
  "actions": [{ "time": "09:30", "title": "跟进星桥教育集团" }]
}
```

## CRM

All CRM endpoints are protected.

Customer stages:

- `new`
- `contacted`
- `qualified`
- `proposal`
- `won`
- `lost`

### List Customers

`GET /api/v1/crm/customers?stage=contacted&q=启明星&limit=20`

Query:

- `stage` optional customer stage.
- `q` optional keyword matched against name, phone, email and website.
- `limit` optional, capped at 100 by service defaults.

Response `200`:

```json
{
  "customers": []
}
```

### Get Customer

`GET /api/v1/crm/customers/{id}`

Errors:

- `400 invalid_customer_id`
- `404 customer_not_found`

### Update Customer

`PATCH /api/v1/crm/customers/{id}`

Request fields are optional, but at least one field must be present:

```json
{
  "name": "成都启明星教育",
  "phone": "028-12345678",
  "email": "hello@example.com",
  "website": "https://example.com"
}
```

Validation:

- `name`, if present, must be non-empty after trimming.
- Contact fields are trimmed and may be set to empty strings.

Response `200`: `CrmCustomer`

Errors:

- `400 invalid_customer_id`
- `400 invalid_crm_input`
- `404 customer_not_found`

### List Customer Activities

`GET /api/v1/crm/customers/{id}/activities?limit=20`

Response `200`:

```json
{
  "activities": []
}
```

Errors:

- `400 invalid_customer_id`
- `404 customer_not_found`

### Import Lead

`POST /api/v1/crm/customers/import-lead`

Request:

```json
{
  "lead_result_id": 99,
  "name": "成都启明星教育",
  "phone": "028-12345678",
  "email": "hello@example.com",
  "website": "https://example.com"
}
```

Validation:

- `lead_result_id` must be positive.
- `name` must be non-empty after trimming.
- Client-supplied `user_id` is ignored.

Response `200`: `CrmCustomer`

### Update Stage

`POST /api/v1/crm/customers/{id}/stage`

Request:

```json
{
  "stage": "contacted",
  "note": "电话已接通"
}
```

Response `200`: `CrmCustomer`

### Record Follow-up

`POST /api/v1/crm/customers/{id}/follow-ups`

Request:

```json
{
  "note": "已发送资料",
  "next_follow_up_at": "2026-07-03T14:00:00Z"
}
```

Response `200`: `CrmFollowUp`

### List Follow-ups

`GET /api/v1/crm/follow-ups?customer_id=100&limit=20`

Query:

- `customer_id` optional.
- `limit` optional.

Response `200`:

```json
{
  "follow_ups": []
}
```

### List Due Customers

`GET /api/v1/crm/customers/due?limit=20`

Response `200`:

```json
{
  "customers": []
}
```

### Pipeline Stats

`GET /api/v1/crm/pipeline-stats`

Response `200`:

```json
{
  "total": 12,
  "new": 2,
  "contacted": 3,
  "qualified": 2,
  "proposal": 1,
  "won": 3,
  "lost": 1,
  "due_today": 4
}
```

### Generate Follow-up Copy

`POST /api/v1/crm/customers/{id}/follow-up-copy`

Request:

```json
{
  "goal": "推进方案会"
}
```

Response `200`:

```json
{
  "subject": "智能客服升级方案跟进",
  "body": "您好，我们已根据贵司场景整理了下一步方案。",
  "channel": "wechat"
}
```

## Growth

All growth endpoints are protected.

### Create Model

`POST /api/v1/growth/models`

Request:

```json
{
  "name": "标准方案",
  "monthly_visits": 24000,
  "lead_rate": 0.068,
  "deal_rate": 0.14,
  "average_order": 820,
  "acquisition_cost": 42,
  "delivery_cost": 51000
}
```

Validation:

- `name` must be non-empty after trimming.
- `monthly_visits` and `average_order` must be greater than 0.
- `lead_rate` and `deal_rate` must be between `0` and `1`.
- `acquisition_cost` and `delivery_cost` must be greater than or equal to 0.

Response `200`: `GrowthModel`

### List Models

`GET /api/v1/growth/models?limit=20`

Response:

```json
{
  "models": []
}
```

### Get Model

`GET /api/v1/growth/models/{id}`

Errors:

- `400 invalid_model_id`
- `404 model_not_found`

### Get Model Scenarios

`GET /api/v1/growth/models/{id}/scenarios`

Response:

```json
{
  "model_id": 99,
  "model_name": "标准方案",
  "scenarios": [
    {
      "name": "标准方案",
      "revenue": 186960,
      "cost": 119544,
      "margin": 0.36,
      "highlight": "当前推荐：投放验证关键词，私域承接高意向线索。"
    }
  ],
  "generated_at": "2026-06-30T08:30:00Z"
}
```

### Get Model Forecast

`GET /api/v1/growth/models/{id}/forecast`

Response:

```json
{
  "model_id": 99,
  "model_name": "标准方案",
  "months": [
    {
      "month": "第3月",
      "revenue": 186960,
      "phase": "稳定投放",
      "progress_percent": 60
    }
  ],
  "generated_at": "2026-06-30T08:30:00Z"
}
```

### Get Model Recommendations

`GET /api/v1/growth/models/{id}/recommendations`

Response:

```json
{
  "model_id": 99,
  "model_name": "标准方案",
  "headline": "先提成交率，再扩大预算",
  "summary": "当前模型每月预计产生 1632 条线索、228 个成交，优先提升转化质量再放大渠道预算。",
  "cost_items": [
    {
      "name": "投放预算",
      "amount": 68544,
      "detail": "搜索词和信息流测试"
    }
  ],
  "action_items": ["优先优化线索到成交转化率，比单纯买流量更划算。"],
  "generated_at": "2026-06-30T08:30:00Z"
}
```

## Leads

All lead endpoints are protected.

Lead task status values:

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`
- `refunded`

### Create Lead Task

`POST /api/v1/leads/tasks`

Request:

```json
{
  "query": "成都 教培 私域转化",
  "idempotency_key": "lead-task-20260703-001"
}
```

Validation:

- `query` must be non-empty after trimming.
- `idempotency_key` must be non-empty.
- Client-supplied `user_id` is ignored.

Response `200`: `LeadTask`

### List Lead Tasks

`GET /api/v1/leads/tasks?limit=20`

Response:

```json
{
  "tasks": []
}
```

### Get Lead Task

`GET /api/v1/leads/tasks/{id}`

Response:

```json
{
  "task": {
    "id": 99,
    "user_id": 42,
    "query": "成都 教培 私域转化",
    "status": "succeeded",
    "idempotency_key": "lead-task-20260703-001",
    "credit_cost": 1,
    "created_at": "2026-06-25T12:00:00Z",
    "updated_at": "2026-06-25T12:05:00Z"
  },
  "progress_percent": 100,
  "message": "线索采集已完成，可以查看结果并导入 CRM。",
  "results_count": 12
}
```

Errors:

- `400 invalid_task_id`
- `404 task_not_found`

### List Lead Results

`GET /api/v1/leads/tasks/{id}/results?limit=20`

Response:

```json
{
  "results": [
    {
      "id": 7,
      "task_id": 99,
      "name": "成都启明星教育",
      "phone": "028-12345678",
      "website": "https://example.com",
      "evidence": [
        {
          "type": "website",
          "title": "官网出现暑期招生咨询入口",
          "url": "https://example.com"
        }
      ],
      "created_at": "2026-06-25T12:05:00Z"
    }
  ]
}
```

Errors:

- `400 invalid_task_id`
- `404 task_not_found`

## Competitor

All competitor endpoints are protected.

Scan statuses:

- `queued`
- `running`
- `succeeded`
- `failed`

### Create Scan

`POST /api/v1/competitor/scans`

Request:

```json
{
  "targets": ["小鹅通", "有赞教育"],
  "focus": "价格、案例、招聘和 AI 功能"
}
```

Validation:

- `targets` must contain at least one non-empty value after trimming.
- `focus` must be non-empty after trimming.
- Client-supplied `user_id` is ignored.

Response `200`: `CompetitorScan`

```json
{
  "id": 99,
  "user_id": 42,
  "targets": ["小鹅通", "有赞教育"],
  "focus": "价格、案例、招聘和 AI 功能",
  "status": "queued",
  "progress_percent": 0,
  "current_step": "queued",
  "competitors": [],
  "conclusions": [],
  "evidence_sources": [],
  "created_at": "2026-07-06T10:00:00Z",
  "updated_at": "2026-07-06T10:00:00Z"
}
```

Consumes membership quota key `competitor_scans`. Defaults seeded by migrations:
free users get 5 scans/month, pro users get 200 scans/month.

Notes:

- Creating a scan creates a queued script task and enqueues a `competitor.scan` background job. It must not fabricate competitor cards or AI conclusions.
- The worker marks scans `running` while processing, then writes scanner results and marks `succeeded`, or marks `failed` with `error_message`.
- Successful scanner results can include `evidence_sources` with `source_type`, `title`, `url`, `summary`, and `captured_at`.
- `OPCV2_COMPETITOR_SCANNER_PROVIDER=development` enables the development scanner for local/demo use. The default is empty, and production rejects the development scanner.
- Real script account execution remains a separate provider implementation step.

Errors:

- `402 quota_exceeded`
- `500 quota_not_configured`

### List Scans

`GET /api/v1/competitor/scans?limit=20`

Response:

```json
{
  "scans": []
}
```

### Get Scan

`GET /api/v1/competitor/scans/{id}`

Errors:

- `400 invalid_scan_id`
- `404 scan_not_found`

### Retry Scan

`POST /api/v1/competitor/scans/{id}/retry`

Requeues an existing scan for the authenticated user. The service resets the
scan to `queued`, sets `progress_percent` to `0`, clears `error_message`, and
enqueues a `competitor.scan` background job.

Response `200`: `CompetitorScan`

Errors:

- `400 invalid_scan_id`
- `404 scan_not_found`
- `500 service_not_ready`

### Add Scan Competitor To Watchlist

`POST /api/v1/competitor/scans/{id}/watchlist`

Adds one competitor from an owned scan result into dynamic monitoring.

Request:

```json
{
  "competitor_name": "增长雷达"
}
```

Response `200`: `CompetitorWatchItem`

Errors:

- `400 invalid_scan_id`
- `400 invalid_watch_item`
- `404 scan_not_found`

### Monitoring

`GET /api/v1/competitor/monitoring?limit=20`

Response:

```json
{
  "watchlist": [],
  "events": []
}
```

### Create Monitoring Watch Item

`POST /api/v1/competitor/monitoring/watchlist`

Creates a monitored competitor entry for the authenticated user. This first
slice stores the object in `competitor_watchlist`; script scheduling and event
generation remain worker/provider follow-up work.

Request:

```json
{
  "name": "增长雷达",
  "category": "商业情报",
  "channels": ["官网 / 价格页", "招聘动态"]
}
```

Response `200`: `CompetitorWatchItem`

`CompetitorWatchItem` includes `id`, `name`, `category`, `status`, `threat`,
`last_seen_at`, `channels`, and `signal`.

Errors:

- `400 invalid_watch_item`
- `500 service_not_ready`

### Delete Monitoring Watch Item

`DELETE /api/v1/competitor/monitoring/watchlist/{id}`

Deletes a monitored competitor entry owned by the authenticated user.

Response:

```json
{
  "deleted": true
}
```

Errors:

- `400 invalid_watch_item`
- `404 watch_item_not_found`

### Start Watch Item Scan

`POST /api/v1/competitor/monitoring/watchlist/{id}/scan`

Creates a queued full-data scan from a monitored competitor entry owned by the
authenticated user. The scan target is the watch item name, and the default
focus is `价格、招聘、内容和产品变化`.

Response `200`: `CompetitorScan`

Errors:

- `400 invalid_watch_item`
- `402 quota_exceeded`
- `404 watch_item_not_found`
- `500 service_not_ready`

## Learning

Course endpoints are public. Progress and diagnosis endpoints are protected.

### List Courses

`GET /api/v1/learning/courses?category=实战&limit=20`

Response:

```json
{
  "courses": []
}
```

### Get Course

`GET /api/v1/learning/courses/{slug}`

Errors:

- `404 course_not_found`

### List Progress

`GET /api/v1/learning/progress`

Protected.

Response:

```json
{
  "progress": []
}
```

### Create Diagnosis

`POST /api/v1/learning/diagnoses`

Protected.

Request:

```json
{
  "goal": "提升AI能力",
  "project": "智能客服"
}
```

Validation:

- `goal` and `project` must be non-empty after trimming.
- Client-supplied `user_id` is ignored.

Response `200`: `LearningDiagnosis`

### Latest Diagnosis

`GET /api/v1/learning/diagnoses/latest`

Protected.

Response `200`: `LearningDiagnosis`

Errors:

- `404 diagnosis_not_found`

### Latest Diagnosis Gaps

`GET /api/v1/learning/diagnoses/latest/gaps`

Protected. Derived from the latest completed diagnosis; no separate table.

Response `200`:

```json
{
  "diagnosis_id": 99,
  "goal": "提升AI能力",
  "project": "智能客服",
  "overall_score": 72,
  "gaps": [
    {
      "name": "数据分析能力",
      "current": 64,
      "target": 86,
      "gap": 22,
      "priority": "high",
      "summary": "需要加强漏斗和转化分析。",
      "evidence": "诊断显示数据分析能力当前为64分，目标差距22分。",
      "recommended": "优先补齐数据分析能力。"
    }
  ],
  "evidence": [],
  "generated_at": "2026-07-03T10:00:00Z"
}
```

Errors:

- `404 diagnosis_not_found`

### Latest Diagnosis Recommendations

`GET /api/v1/learning/diagnoses/latest/recommendations`

Protected. Derived from the latest completed diagnosis.

Response `200` includes `focus`, `recommendations`, and suggested learning
`methods`.

Errors:

- `404 diagnosis_not_found`

### Latest Diagnosis Plan

`GET /api/v1/learning/diagnoses/latest/plan`

Protected. Derived from the latest completed diagnosis.

Response `200` includes `title`, `description`, `recommendations`, `stages`,
`estimated_hours`, and `weekly_suggestion`.

Errors:

- `404 diagnosis_not_found`

### Latest Diagnosis Report

`GET /api/v1/learning/diagnoses/latest/report`

Protected. Derived from the latest completed diagnosis.

Response `200` includes `overall_score`, `dimensions`, `priority_gaps`,
`recommendations`, and `evidence`.

Errors:

- `404 diagnosis_not_found`

## Copilot

All Copilot endpoints are protected. This version is non-streaming. File support stores text content and can inject selected references into Copilot prompts; binary parsing, object storage, embeddings, and full RAG retrieval are not included yet.

### Create Thread

`POST /api/v1/copilot/threads`

Request:

```json
{
  "title": "智能客服机会分析",
  "mode": "chat",
  "model": "gpt-4o"
}
```

Validation:

- `title` may be blank; the backend defaults it to `新会话`.
- `mode` defaults to `chat`.
- Client-supplied `user_id` is ignored.

Response `200`: `CopilotThread`

### List Threads

`GET /api/v1/copilot/threads?limit=20`

Response:

```json
{
  "threads": []
}
```

### List Models

`GET /api/v1/copilot/models`

Response:

```json
{
  "models": [
    {
      "name": "DeepSeek",
      "value": "deepseek",
      "provider": "openai-compatible",
      "is_default": true
    }
  ]
}
```

Notes:

- `value` is the model alias clients send in message and compare requests.
- Models are derived from backend AI configuration, not hard-coded in the client.
- When backend models are configured, unknown aliases return `400 invalid_request`.

### Smoke Test Model

`POST /api/v1/copilot/models/smoke`

Request:

```json
{
  "model": "deepseek",
  "prompt": "ping"
}
```

Validation:

- `model` is a configured model alias from `GET /api/v1/copilot/models`.
- Empty `model` uses the configured default model.
- Empty `prompt` defaults to `ping`.

Response `200`:

```json
{
  "ok": true,
  "model": "deepseek",
  "reply": "pong",
  "input_tokens": 1,
  "output_tokens": 1
}
```

Notes:

- This endpoint runs through the same `GenerateJSON` route as chat and compare, so it verifies alias routing, provider configuration, JSON validation, and AI run logging.
- The response never includes provider secrets, API keys, base URLs, or raw provider metadata.

Errors:

- `400 invalid_request`
- `500 service_not_ready`
- `500 invalid_ai_result`

### List AI Runs

`GET /api/v1/copilot/ai-runs?limit=20`

Response `200`:

```json
{
  "runs": [
    {
      "id": 11,
      "user_id": 42,
      "feature": "copilot.model_smoke",
      "prompt_version": "copilot_model_smoke_v1",
      "provider": "openai-compatible",
      "model": "deepseek",
      "status": "failed",
      "error_code": "provider_unavailable",
      "error_message": "AI provider is unavailable",
      "input_tokens": 12,
      "output_tokens": 0,
      "latency_ms": 2080,
      "created_at": "2026-07-01T10:25:00Z",
      "updated_at": "2026-07-01T10:25:02Z"
    }
  ]
}
```

Notes:

- Only returns runs for the authenticated user and Copilot features (`copilot.%`).
- `limit` defaults to `20` and is capped at `100`.
- The response intentionally excludes stored prompt `request`, model `response`, API keys, base URLs, and raw provider metadata.
- Use this after smoke tests or chat/compare failures to distinguish route, provider, JSON validation, timeout, and rate-limit issues.

Stable AI run `error_code` values:

| Code | Typical cause |
| --- | --- |
| `provider_authentication_failed` | API key is missing, invalid, or rejected by the provider. |
| `provider_permission_denied` | The key lacks permission, quota, account access, or billing access for the route. |
| `provider_model_not_found` | The configured provider model name is wrong or unavailable to the account. |
| `provider_bad_request` | Provider rejected request shape, JSON mode, model capability, or other request parameters. |
| `provider_rate_limited` | Provider returned a rate-limit response. |
| `provider_timeout` | Provider request timed out or the client context was canceled. |
| `provider_unavailable` | Provider returned a server/upstream/network availability failure. |
| `invalid_model_json` | Provider responded, but the body failed backend JSON/schema validation. |
| `internal_error` | Unclassified generation error. |

Errors:

- `400 invalid_limit`
- `500 service_not_ready`

### Get Thread

`GET /api/v1/copilot/threads/{id}`

Errors:

- `400 invalid_thread_id`
- `404 thread_not_found`

### Rename Thread

`PATCH /api/v1/copilot/threads/{id}`

Request:

```json
{
  "title": "新的会话标题"
}
```

Validation:

- `title` must be non-empty after trimming.

Response `200`: `CopilotThread`

### Archive Thread

`DELETE /api/v1/copilot/threads/{id}`

Response `204`.

### List Messages

`GET /api/v1/copilot/threads/{id}/messages?limit=50`

Response:

```json
{
  "messages": [
    {
      "metadata": {
        "kind": "compare_answer"
      }
    }
  ]
}
```

Message metadata:

- `metadata.kind = compare_question` marks the user question that started a comparison.
- `metadata.kind = compare_answer` marks an individual model answer in a comparison.
- `metadata.kind = compare_summary` marks the synthesized comparison summary.
- Older messages may have empty metadata; clients should tolerate missing `metadata.kind`.

### Send Message

`POST /api/v1/copilot/threads/{id}/messages`

Request:

```json
{
  "content": "帮我分析智能客服市场机会",
  "model": "gpt-4o",
  "reference_ids": [17]
}
```

Validation:

- `content` must be non-empty after trimming.
- `reference_ids` is optional; up to five positive, unique IDs are used.
- Referenced files must belong to the authenticated user.

Response `200`:

```json
{
  "user_message": {},
  "assistant_message": {}
}
```

Errors:

- `400 invalid_thread_id`
- `400 invalid_request`
- `404 thread_not_found`
- `500 invalid_ai_result`

### Compare Message

`POST /api/v1/copilot/threads/{id}/compare`

Request:

```json
{
  "content": "对比分析这个项目机会",
  "models": ["deepseek", "gpt-main"]
}
```

Validation:

- `content` must be non-empty after trimming.
- `models` is a list of model aliases from `GET /api/v1/copilot/models`.
- Empty `models` defaults to the configured default model.
- Duplicate models are ignored.
- More than 4 models returns `400 invalid_request`.

Response `200`:

```json
{
  "user_message": {
    "metadata": {
      "kind": "compare_question"
    }
  },
  "answers": [
    {
      "model": "deepseek",
      "assistant_message": {
        "metadata": {
          "kind": "compare_answer"
        }
      }
    }
  ]
}
```

Errors:

- `400 invalid_thread_id`
- `400 invalid_request`
- `404 thread_not_found`
- `500 invalid_ai_result`

### Summarize Comparison

`POST /api/v1/copilot/threads/{id}/compare/summary`

Request:

```json
{
  "content": "对比分析这个项目机会",
  "model": "deepseek",
  "answers": [
    {
      "model": "deepseek",
      "assistant_message": {
        "content": "先做低成本验证。"
      }
    }
  ]
}
```

Validation:

- `content` must be non-empty after trimming.
- `model` is the model alias used to generate the summary. Empty `model` defaults to the thread/default model.
- At least one answer must include non-empty `assistant_message.content`.

Response `200`:

```json
{
  "summary_message": {
    "metadata": {
      "kind": "compare_summary"
    }
  }
}
```

Errors:

- `400 invalid_thread_id`
- `400 invalid_request`
- `404 thread_not_found`
- `500 invalid_ai_result`

### List Memories

`GET /api/v1/copilot/memories?limit=50`

Response:

```json
{
  "memories": []
}
```

### Save Memory

`POST /api/v1/copilot/memories`

Request:

```json
{
  "key": "industry",
  "value": "教培",
  "confidence": 0.9,
  "source": "manual"
}
```

Validation:

- `key` and `value` must be non-empty after trimming.
- `confidence` defaults to `1` and is capped to `1`.

Response `200`: `CopilotMemory`

### Delete Memory

`DELETE /api/v1/copilot/memories/{id}`

Response `204`.

Errors:

- `400 invalid_memory_id`
- `404 memory_not_found`

### List Files

`GET /api/v1/copilot/files?limit=50`

Response:

```json
{
  "files": [
    {
      "id": 17,
      "user_id": 42,
      "name": "智能客服竞品功能对比表.txt",
      "mime_type": "text/plain",
      "size_bytes": 64,
      "content": "小鹅通：私域工具强；有赞教育：交易能力强。",
      "created_at": "2026-07-02T09:00:00Z",
      "updated_at": "2026-07-02T09:00:00Z"
    }
  ]
}
```

### Save File

`POST /api/v1/copilot/files`

Request:

```json
{
  "name": "客户访谈纪要.txt",
  "mime_type": "text/plain",
  "content": "客户最关注响应速度和私域转化。"
}
```

Validation:

- `name` and `content` must be non-empty after trimming.
- `mime_type` defaults to `text/plain`.
- `content` is capped at 120,000 bytes.

Response `200`: `CopilotFile`

## Planned Next Slice Contracts

Status: Mixed. Notifications, membership extensions, content/help/support
extensions, and home aggregation are implemented. Account is partially
implemented except password update.

These contracts come from `docs/plans/2026-07-02-page-backend-gap-analysis-and-architecture.md`.
They cover the first low-risk backend slice: account, notifications, membership
catalog/usage, content/help, and home aggregation.

## Account

Status: Partially implemented. All endpoints in this section are implemented
except `PATCH /api/v1/account/password`, which still needs auth-service password
verification/update wiring.

All account endpoints are protected.

### Get Profile

`GET /api/v1/account/profile`

Response `200`:

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

### Update Profile

`PATCH /api/v1/account/profile`

Request fields are optional:

```json
{
  "nickname": "张晨",
  "email": "founder@example.com",
  "wechat": "zhangchen",
  "company": "智活AI科技有限公司",
  "industry": "企业服务",
  "role": "创始人"
}
```

Validation:

- If present, string fields are trimmed.
- `email`, if present and non-empty, must be a valid email shape.
- Client-supplied `id`, `phone`, `created_at`, and `updated_at` are ignored.

Response `200`: same shape as Get Profile.

### Get Profile Context

`GET /api/v1/account/profile-context`

Returns the normalized six-group user profile context used by AI workflows.
It merges basic account/profile fields with saved onboarding sections. Missing
groups are returned as empty maps so callers can rely on stable keys.

Group keys are fixed:

- `identity`
- `business`
- `products`
- `resources`
- `goals`
- `preferences`

Response `200`:

```json
{
  "user_id": 42,
  "completed": true,
  "groups": [
    {
      "key": "identity",
      "title": "基本身份",
      "fields": {
        "nickname": "张晨",
        "role": "创始人",
        "industry": "企业服务"
      }
    },
    {
      "key": "business",
      "title": "我的业务/公司",
      "fields": {
        "company": "智活AI科技有限公司",
        "industry": "企业服务",
        "stage": "启动"
      }
    },
    {
      "key": "products",
      "title": "我的产品",
      "fields": {}
    },
    {
      "key": "resources",
      "title": "能力与资源",
      "fields": {}
    },
    {
      "key": "goals",
      "title": "目标与诉求",
      "fields": {}
    },
    {
      "key": "preferences",
      "title": "偏好",
      "fields": {}
    }
  ]
}
```

### Get Onboarding

`GET /api/v1/account/onboarding`

Response `200`:

```json
{
  "completed": false,
  "sections": [
    {
      "key": "identity",
      "title": "基本身份",
      "fields": {
        "role": "创始人"
      }
    }
  ]
}
```

### Save Onboarding

`PUT /api/v1/account/onboarding`

Request:

```json
{
  "sections": [
    {
      "key": "identity",
      "fields": {
        "role": "创始人"
      }
    }
  ]
}
```

Validation:

- `sections` must be present.
- Section keys and field keys must be non-empty after trimming.

Response `200`: same shape as Get Onboarding.

### Complete Onboarding

`POST /api/v1/account/onboarding/complete`

Response `200`:

```json
{
  "completed": true
}
```

### Update Password

`PATCH /api/v1/account/password`

Request:

```json
{
  "current_password": "old-password",
  "new_password": "new-password"
}
```

Validation:

- `current_password` and `new_password` must be non-empty.
- Password verification and storage stay owned by the auth service.

Response `204`.

### Get Preferences

`GET /api/v1/account/preferences`

Response `200`:

```json
{
  "notifications_enabled": true,
  "default_model": "deepseek",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

### Update Preferences

`PATCH /api/v1/account/preferences`

Request fields are optional:

```json
{
  "notifications_enabled": true,
  "default_model": "deepseek",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

Response `200`: same shape as Get Preferences.

### Get Quotas

`GET /api/v1/account/quotas`

Response `200`:

```json
{
  "quotas": [
    {
      "key": "ai_tokens",
      "label": "AI智算额度",
      "used": 120,
      "limit": 1000,
      "unit": "次/月",
      "reset_at": "2026-08-01T00:00:00Z"
    }
  ]
}
```

### List Account Content

`GET /api/v1/account/content?limit=20`

Response `200`:

```json
{
  "items": [
    {
      "id": "analysis:99",
      "type": "analysis",
      "title": "智能客服系统项目匹配",
      "summary": "91分",
      "url": "/analysis/sessions/99",
      "created_at": "2026-07-02T10:00:00Z"
    }
  ]
}
```

### Delete Account

`DELETE /api/v1/account`

Response `202`:

```json
{
  "status": "pending_deletion"
}
```

## Notifications

Status: Implemented.

All notification endpoints are protected.

Notification enums:

- `type`: `system`, `task`, `analysis`, `lead`, `crm`, `membership`
- `status`: `all`, `unread`, `read`

### List Notifications

`GET /api/v1/notifications?type=task&status=unread&limit=20`

Response `200`:

```json
{
  "notifications": [
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
  ]
}
```

### Get Notification

`GET /api/v1/notifications/{id}`

Errors:

- `400 invalid_notification_id`
- `404 notification_not_found`

### Mark Notification Read

`PATCH /api/v1/notifications/{id}/read`

Response `200`: `Notification`

### Mark All Notifications Read

`POST /api/v1/notifications/read-all`

Response `200`:

```json
{
  "updated": 12
}
```

### Notification Summary

`GET /api/v1/notifications/summary`

Response `200`:

```json
{
  "unread": 3,
  "by_type": [
    { "type": "task", "count": 2 },
    { "type": "crm", "count": 1 }
  ],
  "latest": []
}
```

## Membership Extensions

Status: Implemented.

The existing `GET /api/v1/membership/me` and
`POST /api/v1/redemptions/redeem` endpoints remain implemented. The endpoints
below extend membership for the membership and profile pages.

Server-side gated actions use the membership usage ledger. The current gated
keys are:

- `sandbox_runs`
- `competitor_scans`

Usage is reset monthly at the first day of the next month. Service methods use
idempotency keys so retried actions do not double-charge quota.

### List Plans

`GET /api/v1/membership/plans`

Response `200`:

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
        {
          "key": "lead_tasks",
          "label": "AI线索任务",
          "limit": 30,
          "unit": "次/月"
        }
      ],
      "recommended": true
    }
  ]
}
```

### Get Membership Usage

`GET /api/v1/membership/usage`

Response `200`:

```json
{
  "usage": [
    {
      "key": "lead_tasks",
      "label": "AI线索任务",
      "used": 8,
      "limit": 30,
      "unit": "次/月",
      "reset_at": "2026-08-01T00:00:00Z"
    }
  ]
}
```

### List Orders

`GET /api/v1/membership/orders?limit=20`

Response `200`:

```json
{
  "orders": [
    {
      "id": 12,
      "order_no": "ZS-20250531-0012",
      "plan_code": "pro",
      "amount_cents": 6900,
      "status": "paid",
      "created_at": "2026-07-02T10:00:00Z",
      "paid_at": "2026-07-02T10:01:00Z"
    }
  ]
}
```

### Create Checkout Order

`POST /api/v1/membership/checkout`

Request:

```json
{
  "plan_code": "pro",
  "billing_cycle": "month"
}
```

Validation:

- `plan_code` must match an active plan.
- `billing_cycle` must match the selected plan.

Response `200`:

```json
{
  "order": {
    "id": 13,
    "order_no": "ZS-20260702-0013",
    "plan_code": "pro",
    "amount_cents": 6900,
    "status": "pending"
  },
  "payment": {
    "mode": "manual",
    "message": "请联系顾问完成开通"
  }
}
```

## Content, Community, Help, and Support Extensions

Status: Implemented for content tools, article bookmarks, community join,
help, and support ticket endpoints. Tool recommendation endpoints remain
planned and are marked below.

Existing public content endpoints remain implemented. The endpoints below extend
content-driven pages and add help/support.

### List Tools With Filters

`GET /api/v1/content/tools?category=创业获客&q=canva&sort=hot&limit=20`

Query:

- `category` optional.
- `q` optional search keyword.
- `sort` optional: `featured`, `latest`, `hot`, `favorite`.

Response `200`:

```json
{
  "tools": []
}
```

### Get Tool

`GET /api/v1/content/tools/{slug}`

Errors:

- `404 tool_not_found`

### Favorite Tool

`POST /api/v1/content/tools/{slug}/favorite`

Response `200`:

```json
{
  "slug": "canva-ai",
  "favorited": true
}
```

### Unfavorite Tool

`DELETE /api/v1/content/tools/{slug}/favorite`

Response `200`:

```json
{
  "slug": "canva-ai",
  "favorited": false
}
```

### Create Tool Recommendation

Status: Planned, not implemented.

`POST /api/v1/tools/recommendations`

Request:

```json
{
  "need": "我想做小红书海报，还想配套文案和数据复盘"
}
```

Validation:

- `need` must be non-empty after trimming.

Response `200`:

```json
{
  "id": 31,
  "summary": "需求摘要",
  "tools": [],
  "created_at": "2026-07-02T10:00:00Z"
}
```

### Get Tool Recommendation Plan

Status: Planned, not implemented.

`GET /api/v1/tools/recommendation-plans/{id}`

Errors:

- `400 invalid_recommendation_id`
- `404 recommendation_not_found`

### Article Bookmark

`POST /api/v1/content/articles/{slug}/bookmark`

Response `200`:

```json
{
  "slug": "ai-customer-service",
  "bookmarked": true
}
```

`DELETE /api/v1/content/articles/{slug}/bookmark`

Response `200`:

```json
{
  "slug": "ai-customer-service",
  "bookmarked": false
}
```

### Join Community Request

`POST /api/v1/community/join-requests`

Request:

```json
{
  "community": "members",
  "contact": "13800138000",
  "note": "希望加入创业成长互助社区"
}
```

Validation:

- `community` must be `members` or `enterprise`.
- `contact` must be non-empty after trimming.

Response `200`:

```json
{
  "id": 9,
  "status": "submitted"
}
```

### List Help Topics

`GET /api/v1/help/topics`

Response `200`:

```json
{
  "topics": [
    { "key": "account", "name": "账号与安全" }
  ]
}
```

### List Help Articles

`GET /api/v1/help/articles?topic=account&q=登录&limit=20`

Response `200`:

```json
{
  "articles": [
    {
      "slug": "login-help",
      "title": "如何登录账号",
      "topic": "account",
      "summary": "登录常见问题"
    }
  ]
}
```

### Get Help Article

`GET /api/v1/help/articles/{slug}`

Errors:

- `404 help_article_not_found`

### Create Support Ticket

`POST /api/v1/support/tickets`

Request:

```json
{
  "topic": "套餐与额度",
  "title": "额度没有更新",
  "body": "我兑换后额度仍未变化。"
}
```

Validation:

- `title` and `body` must be non-empty after trimming.

Response `200`:

```json
{
  "id": 22,
  "status": "open",
  "created_at": "2026-07-02T10:00:00Z"
}
```

### List Support Tickets

`GET /api/v1/support/tickets?limit=20`

Response `200`:

```json
{
  "tickets": []
}
```

## Home Aggregation

Status: Implemented.

### Home Summary

`GET /api/v1/home/summary`

Response `200`:

```json
{
  "hero_cards": [],
  "recommendations": [],
  "recent_tasks": [],
  "notification_summary": {
    "unread": 3,
    "latest": []
  },
  "account_summary": {
    "plan_name": "会员版",
    "quota_warnings": []
  }
}
```

Notes:

- Home summary is an aggregation endpoint, not a source of truth.
- It should tolerate partial source-domain failures by returning safe empty
  arrays for failed optional sections.

## GEO Acquisition

Status: Implemented for Postgres-backed overview reads, analysis request
submission, recent request listing/detail, and worker status processing. Full
analyzer execution, evidence generation, result freshness, CRM handoff, and
admin/write APIs for overview rows are still missing.

All GEO endpoints are protected.

GEO analysis request statuses:

- `queued`
- `running`
- `succeeded`
- `failed`

### GEO Overview

`GET /api/v1/geo/overview`

Response `200`:

```json
{
  "stats": [],
  "engines": [],
  "lead_signals": [],
  "keywords": [],
  "content_tasks": []
}
```

Optional populated item shapes:

```json
{
  "stats": [{ "key": "coverage", "label": "覆盖指标", "value": "12%" }],
  "engines": [{ "name": "AI Search", "coverage_percent": 22, "status": "待优化" }],
  "lead_signals": [{ "title": "线索信号", "detail": "来自后端的线索信号" }],
  "keywords": [{
    "id": 1,
    "query": "后端关键词",
    "intent": "研究",
    "coverage": "待覆盖",
    "score": 71,
    "action": "补充内容页"
  }],
  "content_tasks": [{
    "id": 2,
    "type": "内容页",
    "title": "后端返回的内容任务",
    "priority": "高",
    "due_at": "2026-07-08"
  }]
}
```

### Create GEO Analysis Request

`POST /api/v1/geo/analysis-requests`

Request:

```json
{
  "target": "面向制造业的 AI 质检工具"
}
```

Validation:

- `target` must be non-empty after trimming.
- Client-supplied `user_id` and `status` are ignored.
- Creating a request does not create mock overview rows. The request is stored
  with `status: "queued"` and enqueues a `geo.analysis` background job.
- The current worker marks the request `running`, then `failed` with
  `error_message: "analyzer_not_configured"` when no GEO analyzer is configured.
  It must not synthesize overview rows.

Response `200`:

```json
{
  "id": 7,
  "user_id": 42,
  "target": "面向制造业的 AI 质检工具",
  "status": "queued",
  "created_at": "2026-07-04T08:00:00Z",
  "updated_at": "2026-07-04T08:00:00Z"
}
```

Errors:

- `400 invalid_request`
- `500 service_not_ready`

Worker-visible failed request example:

```json
{
  "id": 7,
  "user_id": 42,
  "target": "面向制造业的 AI 质检工具",
  "status": "failed",
  "error_message": "analyzer_not_configured",
  "created_at": "2026-07-04T08:00:00Z",
  "updated_at": "2026-07-04T08:01:00Z"
}
```

### List GEO Analysis Requests

`GET /api/v1/geo/analysis-requests?limit=20`

Response `200`:

```json
{
  "requests": [
    {
      "id": 7,
      "user_id": 42,
      "target": "面向制造业的 AI 质检工具",
      "status": "queued",
      "created_at": "2026-07-04T08:00:00Z",
      "updated_at": "2026-07-04T08:00:00Z"
    }
  ]
}
```

Notes:

- Only returns requests owned by the authenticated user.
- `limit` defaults to `20` and is capped to `100`.
- Empty list responses return `{"requests":[]}`.

Errors:

- `400 invalid_limit`

### Get GEO Analysis Request

`GET /api/v1/geo/analysis-requests/{id}`

Response `200`:

```json
{
  "id": 7,
  "user_id": 42,
  "target": "面向制造业的 AI 质检工具",
  "status": "queued",
  "created_at": "2026-07-04T08:00:00Z",
  "updated_at": "2026-07-04T08:00:00Z"
}
```

Notes:

- Only returns requests owned by the authenticated user.
- This endpoint does not include generated evidence yet; it only returns the
  stored request metadata.
- `error_message` is optional and appears when the request has a failure reason.

Errors:

- `400 invalid_analysis_request_id`
- `404 analysis_request_not_found`

## Enterprise Companion

Status: Implemented as a Postgres-backed overview plus diagnosis request
workflow. It supports user-scoped metrics, plans, delivery board items,
milestones, cases, diagnosis request creation/listing/status updates, and CRM
handoff for completed delivery. Milestone write/admin APIs and payment handoff
are still missing.

All enterprise endpoints are protected.

### Enterprise Overview

`GET /api/v1/enterprise/overview`

Response `200`:

```json
{
  "stats": [],
  "plans": [],
  "delivery_board": [],
  "milestones": [],
  "cases": []
}
```

### Diagnosis Requests

`GET /api/v1/enterprise/diagnosis-requests?limit=5`

Response `200`:

```json
{
  "requests": [{
    "id": 8,
    "user_id": 42,
    "need": "30人销售团队需要AI获客陪跑",
    "status": "submitted",
    "created_at": "2026-07-07T10:30:00Z",
    "updated_at": "2026-07-07T10:30:00Z"
  }]
}
```

`POST /api/v1/enterprise/diagnosis-requests`

Request:

```json
{
  "need": "30人销售团队需要AI获客陪跑"
}
```

`PATCH /api/v1/enterprise/diagnosis-requests/{id}`

Request:

```json
{
  "status": "completed"
}
```

Allowed `status` values:

- `follow_up_created`
- `in_delivery`
- `completed`

### CRM Handoff

`POST /api/v1/enterprise/diagnosis-requests/{id}/crm-customer`

Protected. Only completed diagnosis requests can be handed off. The endpoint is
idempotent by `enterprise_diagnosis_request:{id}` and creates or returns a CRM
customer with `source = enterprise` and `stage = won`.

Response `200`:

```json
{
  "id": 100,
  "user_id": 42,
  "import_key": "enterprise_diagnosis_request:8",
  "name": "30人销售团队需要AI获客陪跑",
  "stage": "won",
  "source": "enterprise"
}
```

Optional populated item shapes:

```json
{
  "stats": [{ "key": "companies", "label": "服务企业数", "value": "3" }],
  "plans": [{
    "id": 1,
    "title": "陪跑方案",
    "audience": "增长团队",
    "price_label": "待报价",
    "focus": ["诊断", "训练"],
    "result": "完成系统上线"
  }],
  "delivery_board": [{ "stage": "诊断中", "count": 1, "detail": "交付阶段" }],
  "milestones": [{ "time_label": "第1周", "title": "里程碑", "detail": "完成诊断" }],
  "cases": [{ "id": 7, "company": "企业案例", "result": "完成落地复盘" }]
}
```
