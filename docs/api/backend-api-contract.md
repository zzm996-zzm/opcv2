# Backend API Contract

This document tracks the backend endpoints currently wired for the new web modules.

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

Response `200`: `SandboxSession` with `status: "completed"` and populated `report`.

Errors:

- `400 invalid_session_id`
- `404 session_not_found`
- `500 invalid_ai_result`

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

`GET /api/v1/tasks?limit=20`

Response:

```json
{
  "tasks": []
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

## Competitor

All competitor endpoints are protected.

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

### Monitoring

`GET /api/v1/competitor/monitoring?limit=20`

Response:

```json
{
  "watchlist": [],
  "events": []
}
```

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

## Copilot

All Copilot endpoints are protected. This first version is non-streaming and does not include RAG or file upload.

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
  "model": "gpt-4o"
}
```

Validation:

- `content` must be non-empty after trimming.

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
