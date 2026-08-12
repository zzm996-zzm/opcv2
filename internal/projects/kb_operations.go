package projects

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

const (
	OperationsStatusQueued  = "queued"
	OperationsStatusRunning = "running"
	OperationsStatusReady   = "ready"
	OperationsStatusFailed  = "failed"
)

type KBReindexJob struct {
	ID            int64      `json:"id"`
	TaskKey       string     `json:"task_key"`
	RequestedBy   *int64     `json:"requested_by,omitempty"`
	SourceVersion int        `json:"source_version"`
	TargetVersion int        `json:"target_version"`
	Status        string     `json:"status"`
	DocumentCount int        `json:"document_count"`
	ErrorCode     string     `json:"error_code,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
	created       bool
}

type ProjectAIAnswer struct {
	ID                int64          `json:"id"`
	UserID            *int64         `json:"user_id,omitempty"`
	Entry             string         `json:"entry"`
	RawInput          string         `json:"raw_input,omitempty"`
	ParsedIntent      map[string]any `json:"parsed_intent"`
	CitedChunkIDs     []string       `json:"cited_chunk_ids"`
	CitedWebSourceIDs []string       `json:"cited_web_source_ids"`
	KBSufficiency     *float64       `json:"kb_sufficiency,omitempty"`
	Result            any            `json:"result,omitempty"`
	ModelName         string         `json:"model_name,omitempty"`
	PromptVersion     string         `json:"prompt_version,omitempty"`
	KBVersion         int            `json:"kb_version"`
	CacheKey          string         `json:"cache_key,omitempty"`
	LatencyMS         int            `json:"latency_ms"`
	Status            string         `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
}

type ContentProductionRun struct {
	ID            int64      `json:"id"`
	TaskKey       string     `json:"task_key"`
	ScheduleDate  string     `json:"schedule_date"`
	Status        string     `json:"status"`
	PlannedCount  int        `json:"planned_count"`
	ProducedCount int        `json:"produced_count"`
	ErrorCode     string     `json:"error_code,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
	created       bool
}

type ProjectKBRepository interface {
	GetActiveProjectKBVersion(context.Context) (int, error)
	CreateProjectKBReindex(context.Context, int64, string, time.Time) (KBReindexJob, error)
	GetProjectKBReindex(context.Context, int64) (KBReindexJob, error)
	MarkProjectKBReindexFailed(context.Context, int64, string, time.Time) error
	RebuildProjectKB(context.Context, int64, time.Time) (KBReindexJob, error)
	ListProjectAIAnswers(context.Context, int) ([]ProjectAIAnswer, error)
	SaveProjectAIAnswer(context.Context, ProjectAIAnswer) error
	CreateContentProductionRun(context.Context, string, string, int, time.Time) (ContentProductionRun, error)
	ProduceProjectContent(context.Context, int64, time.Time) (ContentProductionRun, error)
	MarkContentProductionFailed(context.Context, int64, string, time.Time) error
}

type ProjectKBAdminApplication interface {
	RequestProjectKBReindex(context.Context, int64) (KBReindexJob, error)
	GetProjectKBReindex(context.Context, int64, int64) (KBReindexJob, error)
	ListProjectAIAnswers(context.Context, int64, int) ([]ProjectAIAnswer, error)
}

func (s *Service) kbRepository() (ProjectKBRepository, error) {
	repository, ok := s.repository.(ProjectKBRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) RequestProjectKBReindex(ctx context.Context, adminUserID int64) (KBReindexJob, error) {
	if _, err := s.requireProjectAdmin(ctx, adminUserID); err != nil {
		return KBReindexJob{}, err
	}
	repository, err := s.kbRepository()
	if err != nil || s.matchQueue == nil {
		return KBReindexJob{}, ErrServiceNotReady
	}
	version, err := repository.GetActiveProjectKBVersion(ctx)
	if err != nil {
		return KBReindexJob{}, err
	}
	taskKey := fmt.Sprintf("project-kb-reindex-v%d", version+1)
	job, err := repository.CreateProjectKBReindex(ctx, adminUserID, taskKey, s.now())
	if err != nil {
		return KBReindexJob{}, err
	}
	if !job.created || job.Status != OperationsStatusQueued {
		return job, nil
	}
	if err := s.matchQueue.Enqueue(ctx, jobs.Job{Type: jobs.TypeProjectKBReindex, IdempotencyKey: taskKey, Payload: map[string]any{"reindex_id": job.ID}, MaxRetry: 1, Timeout: 15 * time.Minute}); err != nil {
		_ = repository.MarkProjectKBReindexFailed(ctx, job.ID, "queue_unavailable", s.now())
		job.Status, job.ErrorCode = OperationsStatusFailed, "queue_unavailable"
		return job, err
	}
	return job, nil
}

func (s *Service) GetProjectKBReindex(ctx context.Context, adminUserID, id int64) (KBReindexJob, error) {
	if _, err := s.requireProjectAdmin(ctx, adminUserID); err != nil {
		return KBReindexJob{}, err
	}
	repository, err := s.kbRepository()
	if err != nil {
		return KBReindexJob{}, err
	}
	return repository.GetProjectKBReindex(ctx, id)
}

func (s *Service) ListProjectAIAnswers(ctx context.Context, adminUserID int64, limit int) ([]ProjectAIAnswer, error) {
	if _, err := s.requireProjectAdmin(ctx, adminUserID); err != nil {
		return nil, err
	}
	repository, err := s.kbRepository()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items, err := repository.ListProjectAIAnswers(ctx, limit)
	if items == nil {
		items = []ProjectAIAnswer{}
	}
	return items, err
}

func (s *Service) ProcessProjectKBReindex(ctx context.Context, id int64) error {
	repository, err := s.kbRepository()
	if err != nil {
		return err
	}
	_, err = repository.RebuildProjectKB(ctx, id, s.now())
	if err != nil {
		_ = repository.MarkProjectKBReindexFailed(ctx, id, "reindex_failed", s.now())
	}
	return err
}

func (s *Service) ActiveProjectKBVersion(ctx context.Context) int {
	repository, err := s.kbRepository()
	if err != nil {
		return 1
	}
	version, err := repository.GetActiveProjectKBVersion(ctx)
	if err != nil || version <= 0 {
		return 1
	}
	return version
}

func (s *Service) recordMatchAIAnswer(ctx context.Context, run MatchRun, bundle matchEvidenceBundle, result MatchResult, status string, started time.Time) {
	repository, err := s.kbRepository()
	if err != nil {
		return
	}
	chunks := make([]string, 0)
	web := make([]string, 0)
	for _, evidence := range bundle.Evidence {
		if evidence.SourceType == "knowledge_base" && evidence.SourceID != "" {
			chunks = append(chunks, evidence.SourceID)
		}
		if evidence.SourceType == "web" && evidence.URL != "" {
			web = append(web, evidence.URL)
		}
	}
	sufficiency := bundle.Sufficiency
	userID := run.UserID
	answer := ProjectAIAnswer{
		UserID: &userID, Entry: "match", RawInput: run.Need, ParsedIntent: run.ParsedProfile,
		CitedChunkIDs: chunks, CitedWebSourceIDs: web, KBSufficiency: &sufficiency, Result: result,
		PromptVersion: "project_match_v2", KBVersion: max(1, bundle.KBVersion),
		CacheKey:  fmt.Sprintf("project-match:%d:%d", run.ID, run.GenerationAttempt),
		LatencyMS: int(s.now().Sub(started).Milliseconds()), Status: status, CreatedAt: s.now(),
	}
	_ = repository.SaveProjectAIAnswer(ctx, answer)
}

func (s *Service) ScheduleContentProduction(ctx context.Context, at time.Time) (ContentProductionRun, error) {
	repository, err := s.kbRepository()
	if err != nil || s.matchQueue == nil {
		return ContentProductionRun{}, ErrServiceNotReady
	}
	date := at.In(projectOperationsLocation()).Format("2006-01-02")
	key := "project-content-batch-" + date
	run, err := repository.CreateContentProductionRun(ctx, key, date, 30, s.now())
	if err != nil {
		return ContentProductionRun{}, err
	}
	if !run.created || run.Status != OperationsStatusQueued {
		return run, nil
	}
	if err := s.matchQueue.Enqueue(ctx, jobs.Job{Type: jobs.TypeProjectContentBatch, IdempotencyKey: key, Payload: map[string]any{"run_id": run.ID}, MaxRetry: 1, Timeout: 10 * time.Minute}); err != nil {
		return run, err
	}
	return run, nil
}

func (s *Service) ProcessProjectContentBatch(ctx context.Context, id int64) error {
	repository, err := s.kbRepository()
	if err != nil {
		return err
	}
	_, err = repository.ProduceProjectContent(ctx, id, s.now())
	if err != nil {
		_ = repository.MarkContentProductionFailed(ctx, id, "production_failed", s.now())
	}
	return err
}

func projectOperationsLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*60*60)
	}
	return location
}

func RunProjectContentScheduler(ctx context.Context, service interface {
	ScheduleContentProduction(context.Context, time.Time) (ContentProductionRun, error)
}, interval time.Duration, now func() time.Time, onError func(error)) {
	if interval <= 0 {
		interval = time.Hour
	}
	if now == nil {
		now = time.Now
	}
	run := func() {
		current := now().In(projectOperationsLocation())
		if (current.Weekday() != time.Tuesday && current.Weekday() != time.Friday) || current.Hour() != 2 {
			return
		}
		if _, err := service.ScheduleContentProduction(ctx, current); err != nil && onError != nil {
			onError(err)
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func kbVersionString(version int) string { return strconv.Itoa(max(1, version)) }
func normalizeAIAnswerStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "ok", "partial", "ai_no_result", "degraded", "failed":
		return strings.TrimSpace(value)
	default:
		return "failed"
	}
}
