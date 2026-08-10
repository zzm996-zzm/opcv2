# Commercial Sandbox V1.2 Backend Design

## Scope

Implement the PRD V1.2 backend inside the existing Go `sandbox` module while preserving the current `/api/v1/sandbox/sessions` compatibility APIs. The new contract uses `/api/v1/sandbox-runs` and adds eight stable role codes, bounded clarification, isolated role model calls, persisted progress events, SSE replay, an independent report synthesis call, history management, and export metadata.

The first production slice does not add multi-user collaboration, editable role personas, live user intervention during a run, public sharing, or multi-round role debate. Paywall enforcement remains disabled. Usage is counted, but quota errors never block a run.

## Architecture

`sandbox_sessions` remains the canonical run table so existing IDs, ownership checks, frontend history, and worker payloads remain valid. A migration adds V2 context, completeness, assumptions, orchestration snapshots, lifecycle timestamps, and terminal statuses. Normalized child tables own role configuration, per-run isolated role execution snapshots, replayable events, synthesized reports, and exports.

The service exposes a V2 application interface in parallel with the existing interface. Draft creation and clarification freeze a structured run context. Starting a run validates 3-8 active roles including required `skeptic`, rejects a second concurrently active run for the same user, prepares deterministic role records once, records usage without blocking, and enqueues the existing Asynq sandbox job with a deterministic task ID.

The worker loads the frozen context and role snapshots. It executes at most three isolated role calls concurrently. Each request contains shared facts plus exactly one role profile and never contains another role's output. Results are schema-validated and dimension coverage is checked. Each role retries once, then becomes `failed`. Events are persisted before being returned by SSE. A separate report session consumes only completed structured role outputs and the failed role list. Partial runs preserve successful outputs and identify missing roles.

## Security And Reliability

- Every read and mutation filters by authenticated `user_id`.
- Start is idempotent and role records are created once per run attempt.
- `skeptic` is required; roles are limited to 3-8 active codes.
- A user may have at most one queued/running V2 run.
- Per-role timeout is 60 seconds; total job timeout remains 300 seconds.
- Each role has a distinct session ID and the report uses another distinct ID.
- Model chain-of-thought is never stored. Stored audit data is limited to prompts/version identifiers, dimensions, input hash, normalized output, tokens, duration, retry count, and errors.
- SSE supports `Last-Event-ID`; terminal events stop the stream.
- User stop preserves completed roles and permits a partial report.

## Compatibility

The current draft/intake, legacy worker behavior, post-run role questions, and existing report fields continue to work. New V2 DTOs do not replace `Session` or `Report`; adapters return V2 snapshots while legacy handlers keep their old payloads. New tables use BIGINT identities to match the repository rather than the UUID examples in the PRD.

## Verification

Unit tests cover role rules, clarification bounds, ownership, concurrent-run rejection, idempotent start, isolated role prompts, retry/partial behavior, report validation, SSE replay, stop, history, and export expiry. Repository tests cover SQL ownership and event ordering. Final verification is `go test ./...`, `make lint`, and `make build`.
