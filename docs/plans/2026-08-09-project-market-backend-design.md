# Project Market Backend Design

## Context

The project market PRD defines four distinct products over one evidence base: a public project catalog, offline-produced opportunity exploration, evidence-verified case studies, and an online personalized matching workflow. The current repository already has a Go/Gin modular monolith, PostgreSQL, Redis/Asynq, model generation, and an early `internal/projects` package. The existing implementation must remain usable while the richer evidence and orchestration model is added.

The PRD SQL is conceptual rather than executable in this repository. It assumes UUID user IDs and pre-existing `projects` and `match_runs` tables, while the application uses BIGINT IDs, `project_opportunities`, and `project_match_sessions`. All new migrations therefore use the repository's BIGINT convention and extend the existing project tables instead of creating a second incompatible identity system.

## Chosen Approach

Keep a modular monolith and evolve the current project domain in place. Public catalog reads and authenticated user actions will be registered separately. Existing `/projects/opportunities`, `/projects/cases`, and `/projects/matches` endpoints remain available as compatibility routes while the PRD routes are introduced. The first implementation slice establishes executable schema, feature configuration, public catalog APIs, source validation, and stable response contracts.

Two alternatives were rejected. Replacing all current project tables would simplify naming but would break the implemented frontend and discard useful tests. Building separate microservices for research, files, and matching would isolate workloads but adds deployment and operational cost before the workflows are proven. The modular-monolith approach retains transaction boundaries and testability while allowing later extraction behind explicit interfaces.

## Domain Boundaries

- `projects/catalog`: home aggregation, filtering, details, favorites, comparisons, views, and heat.
- `projects/knowledge`: startup failures, rebuild plans, localizations, dictionaries, and knowledge chunks.
- `projects/evidence`: field-level source records, web sources, claim mappings, quality, and conflicts.
- `projects/match`: input snapshots, clarification rounds, retrieval, research orchestration, results, and history.
- `projects/files`: ownership, upload lifecycle, scanning, extraction, expiry, and deletion.
- `projects/research`: search plans, opened-page evidence, canonicalization, quality scoring, and refresh.
- `projects/publishing`: opportunity and case production states, review, publication, staleness, and rollback.

These boundaries initially remain packages within the same Go module. Interfaces isolate model providers, embedding providers, object storage, document parsing, malware scanning, web search, page fetching, and job progress publishing.

## Data Design

Existing `project_opportunities` remains the canonical public project record. It is extended with list/detail fields, publication metadata, references to source failure/rebuild records, and the paywall-compatible response flags. Structured detail stays in JSONB for the five fixed tabs, while facts and sources are normalized so publication rules can be enforced in SQL and service code.

New content tables use BIGSERIAL/BIGINT foreign keys: `project_dict_items`, `startup_failures`, `startup_failure_sources`, `startup_failure_antipatterns`, `rebuild_plans`, `rebuild_localizations`, `web_research_jobs`, `web_sources`, `opportunity_items`, `opportunity_item_sources`, `project_case_sources`, `project_views`, `project_favorites`, `project_compare_items`, `project_import_batches`, and `content_corrections`.

The existing `project_match_sessions` is evolved into the match run rather than creating a parallel `match_runs` table. It gains input snapshots, parsed profile, assumptions, completeness, clarification round, knowledge sufficiency, research job reference, progress status, idempotency key, cancellation timestamp, and structured result. A later migration adds `project_match_files` and `kb_chunks` after object storage and the embedding model are selected.

## Request And Processing Flow

Public browsing reads only published records and never invokes a model. Banner queries are parsed into filters and query terms, then search published opportunity items. Authenticated favorites and comparisons use optional user context without making public catalog access require a token.

Matching is a persisted state machine. Create stores an immutable input snapshot and returns structured analysis plus up to three questions. Answers append immutable answer events and re-run completeness. Generate enqueues an idempotent Asynq job. The worker applies hard filters, performs full-text/vector retrieval, calculates knowledge sufficiency, optionally creates a web research job, merges evidence, validates claim references, and writes a complete or partial result. Progress is persisted and published through Redis; SSE reads the persisted snapshot first and then subscribes to updates, enabling reconnects.

Offline opportunity and case production uses the same research and evidence components but separate job types and publication rules. User browsing never triggers per-card generation.

## Security And Failure Handling

Only `http` and `https` evidence URLs are returned. `loot-drop.io` cannot qualify as an independently verified source. Model-generated content is marked at the data boundary. Published facts require at least one valid claim-level source; sensitive negative cases require review. Retrieved pages and uploaded files are untrusted data and cannot issue tool instructions.

File endpoints validate extension, declared MIME, detected MIME, ownership, per-file size, total size, and expiration. Storage keys are never exposed directly. Matching and generation commands require `Idempotency-Key`. Cancellation is cooperative and prevents later stages from starting. Network failure returns `partial` when local evidence remains usable; absence of both evidence types returns the business state `ai_no_result` with HTTP 200.

The configured performance numbers are objectives, not correctness assumptions. Timeouts return durable partial state, while background work may finish and update the snapshot. No user file content enters shared caches.

## Testing Strategy

Each slice starts with service and HTTP tests. Repository tests cover published-only reads, source eligibility, pagination, ownership, idempotency, and state transitions. Migration tests run every migration against PostgreSQL and verify rollback where supported. Contract tests protect both compatibility routes and new PRD routes. Worker tests cover cancellation and partial degradation without external providers.

The first slice is complete when public project browsing works without authentication, protected actions still require authentication, config returns paywall disabled, migrations match BIGINT identities, existing project endpoints remain compatible, and all Go tests pass.
