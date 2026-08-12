package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

const (
	MatchStatusQueued    = "queued"
	MatchStatusRunning   = "running"
	MatchStatusCompleted = "completed"
	MatchStatusPartial   = "partial"
	MatchStatusFailed    = "failed"
	MatchStatusCanceled  = "canceled"

	MatchStepQueued         = "queued"
	MatchStepAnalyzing      = "analyzing"
	MatchStepRetrievingKB   = "retrieving_kb"
	MatchStepResearchingWeb = "researching_web"
	MatchStepMerging        = "merging"
	MatchStepGenerating     = "generating"
	MatchStepDone           = "done"
	MatchStepPartial        = "partial"
	MatchStepError          = "error"
	MatchStepCanceled       = "canceled"
)

var (
	ErrMatchNotReady        = errors.New("project match is not ready to generate")
	ErrStaleMatchGeneration = errors.New("stale project match generation")
)

type ProjectMatchQueue interface {
	Enqueue(context.Context, jobs.Job) error
}

type MatchProgressEvent struct {
	ID              int64          `json:"id"`
	MatchID         int64          `json:"match_id"`
	Attempt         int            `json:"attempt"`
	Event           string         `json:"event"`
	ProgressPercent int            `json:"progress_percent"`
	Payload         map[string]any `json:"payload,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type MatchGenerationResponse struct {
	MatchID         int64       `json:"match_id"`
	Status          string      `json:"status"`
	Attempt         int         `json:"attempt"`
	ProgressPercent int         `json:"progress_percent"`
	CurrentStep     string      `json:"current_step"`
	Result          MatchResult `json:"result,omitempty"`
	ErrorCode       string      `json:"error_code,omitempty"`
}

type MatchGenerationRepository interface {
	PrepareMatchGeneration(context.Context, int64, int64) (MatchRun, error)
	UpdateMatchGeneration(context.Context, int64, int64, int, string, int, string, string, *MatchResult) (MatchRun, error)
	CancelMatchGeneration(context.Context, int64, int64) (MatchRun, error)
	ListMatchProgressEvents(context.Context, int64, int64, int64) ([]MatchProgressEvent, error)
}

func (s *Service) generationRepository() (MatchGenerationRepository, error) {
	repository, ok := s.repository.(MatchGenerationRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) GenerateProjectMatch(ctx context.Context, userID, matchID int64) (MatchGenerationResponse, error) {
	repository, err := s.generationRepository()
	if err != nil || s.matchQueue == nil {
		return MatchGenerationResponse{}, ErrServiceNotReady
	}
	workflow, err := s.workflowRepository()
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	run, err := workflow.GetMatchRun(ctx, userID, matchID)
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	switch run.Status {
	case MatchStatusQueued, MatchStatusRunning, MatchStatusCompleted, MatchStatusPartial:
		return matchGenerationResponse(run), nil
	case MatchStatusReady, MatchStatusFailed, MatchStatusCanceled:
	default:
		return MatchGenerationResponse{}, ErrMatchNotReady
	}
	run, err = repository.PrepareMatchGeneration(ctx, userID, matchID)
	if err != nil {
		if errors.Is(err, ErrMatchNotReady) {
			current, getErr := workflow.GetMatchRun(ctx, userID, matchID)
			if getErr == nil && (current.Status == MatchStatusQueued || current.Status == MatchStatusRunning || current.Status == MatchStatusCompleted || current.Status == MatchStatusPartial) {
				return matchGenerationResponse(current), nil
			}
		}
		return MatchGenerationResponse{}, err
	}
	if err := s.matchQueue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeProjectMatchGenerate,
		IdempotencyKey: fmt.Sprintf("project-match-%d-attempt-%d", matchID, run.GenerationAttempt),
		Payload:        map[string]any{"user_id": userID, "match_id": matchID, "attempt": run.GenerationAttempt},
		MaxRetry:       1,
		Timeout:        10 * time.Minute,
	}); err != nil {
		failed, _ := repository.UpdateMatchGeneration(ctx, userID, matchID, run.GenerationAttempt, MatchStatusFailed, 100, MatchStepError, "queue_unavailable", nil)
		if failed.ID != 0 {
			return matchGenerationResponse(failed), err
		}
		return MatchGenerationResponse{}, err
	}
	return matchGenerationResponse(run), nil
}

func (s *Service) CancelProjectMatch(ctx context.Context, userID, matchID int64) (MatchGenerationResponse, error) {
	repository, err := s.generationRepository()
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	workflow, err := s.workflowRepository()
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	run, err := workflow.GetMatchRun(ctx, userID, matchID)
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	if run.Status == MatchStatusCanceled {
		return matchGenerationResponse(run), nil
	}
	if run.Status != MatchStatusQueued && run.Status != MatchStatusRunning {
		return MatchGenerationResponse{}, ErrMatchNotReady
	}
	run, err = repository.CancelMatchGeneration(ctx, userID, matchID)
	if err != nil {
		return MatchGenerationResponse{}, err
	}
	return matchGenerationResponse(run), nil
}

func (s *Service) ListProjectMatchProgress(ctx context.Context, userID, matchID, afterID int64) ([]MatchProgressEvent, error) {
	repository, err := s.generationRepository()
	if err != nil {
		return nil, err
	}
	if _, err := s.GetProjectMatch(ctx, userID, matchID); err != nil {
		return nil, err
	}
	return repository.ListMatchProgressEvents(ctx, userID, matchID, afterID)
}

func (s *Service) ProcessProjectMatch(ctx context.Context, userID, matchID int64, attempt int) error {
	repository, err := s.generationRepository()
	if err != nil || s.generator == nil {
		return ErrServiceNotReady
	}
	active, err := s.advanceMatchGeneration(ctx, repository, userID, matchID, attempt, MatchStepAnalyzing, 10)
	if err != nil || !active {
		return err
	}
	workflow, err := s.workflowRepository()
	if err != nil {
		return err
	}
	run, err := workflow.GetMatchRun(ctx, userID, matchID)
	if err != nil {
		return err
	}
	active, err = s.advanceMatchGeneration(ctx, repository, userID, matchID, attempt, MatchStepRetrievingKB, 30)
	if err != nil || !active {
		return err
	}
	bundle, evidenceErr := s.buildMatchEvidence(ctx, run)
	active, err = s.advanceMatchGeneration(ctx, repository, userID, matchID, attempt, MatchStepResearchingWeb, 50)
	if err != nil || !active {
		return err
	}
	if saveErr := s.saveMatchEvidence(ctx, run, bundle); saveErr != nil {
		if errors.Is(saveErr, ErrStaleMatchGeneration) {
			return nil
		}
		return saveErr
	}
	if evidenceErr != nil {
		if errors.Is(evidenceErr, ErrInsufficientEvidence) && len(bundle.Catalog) > 0 {
			partial := catalogResult(bundle.Catalog)
			partial.Status, partial.Evidence, partial.EvidenceStatus = MatchStatusPartial, bundle.Evidence, "insufficient"
			_, updateErr := repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusPartial, 100, MatchStepPartial, "insufficient_evidence", &partial)
			if errors.Is(updateErr, ErrStaleMatchGeneration) {
				return nil
			}
			return updateErr
		}
		_, _ = repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusFailed, 100, MatchStepError, "insufficient_evidence", nil)
		return evidenceErr
	}
	active, err = s.advanceMatchGeneration(ctx, repository, userID, matchID, attempt, MatchStepMerging, 70)
	if err != nil || !active {
		return err
	}
	_, filePrompt, fileErr := s.resolveProjectMatchFiles(ctx, userID, run.InputSnapshot.FileIDs)
	if fileErr != nil {
		return fileErr
	}
	active, err = s.advanceMatchGeneration(ctx, repository, userID, matchID, attempt, MatchStepGenerating, 85)
	if err != nil || !active {
		return err
	}
	profilePrompt := ""
	if len(run.ParsedProfile) > 0 {
		if payload, marshalErr := json.Marshal(run.ParsedProfile); marshalErr == nil {
			profilePrompt = "已确认用户画像：" + string(payload)
		}
	}
	result, err := s.generateMatchFromEvidence(ctx, MatchInput{UserID: userID, Intent: run.Need, FilePrompt: filePrompt}, profilePrompt, bundle.Catalog, bundle.Evidence)
	if err != nil {
		if len(bundle.Catalog) > 0 {
			partial := catalogResult(bundle.Catalog)
			partial.Evidence, partial.EvidenceStatus = bundle.Evidence, "partial"
			_, updateErr := repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusPartial, 100, MatchStepPartial, "generation_degraded", &partial)
			if errors.Is(updateErr, ErrStaleMatchGeneration) {
				return nil
			}
			return updateErr
		}
		_, _ = repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusFailed, 100, MatchStepError, "invalid_ai_result", nil)
		return err
	}
	if bundle.Degraded {
		result.Status, result.EvidenceStatus = MatchStatusPartial, "partial"
		_, err = repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusPartial, 100, MatchStepPartial, "research_partial", &result)
		if errors.Is(err, ErrStaleMatchGeneration) {
			return nil
		}
		return err
	}
	_, err = repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusCompleted, 100, MatchStepDone, "", &result)
	if errors.Is(err, ErrStaleMatchGeneration) {
		return nil
	}
	return err
}

func (s *Service) partialCatalogResult(ctx context.Context) (MatchResult, bool) {
	if s.repository == nil {
		return MatchResult{}, false
	}
	catalog, err := s.repository.ListOpportunities(ctx, OpportunityFilters{Limit: 3})
	if err != nil || len(catalog) == 0 {
		return MatchResult{}, false
	}
	return catalogResult(catalog), true
}

func catalogResult(catalog []Opportunity) MatchResult {
	if len(catalog) > 3 {
		catalog = catalog[:3]
	}
	result := MatchResult{Status: MatchStatusPartial, Projects: make([]ProjectMatch, 0, len(catalog))}
	for index, item := range catalog {
		result.Projects = append(result.Projects, ProjectMatch{Rank: index + 1, OpportunitySlug: item.Slug, Title: item.Title, Tags: append([]string(nil), item.Tags...), Budget: item.BudgetBand, Reasons: []string{"基于已发布项目目录返回的降级候选"}, Risk: "个性化生成暂不可用，请结合项目详情进一步判断"})
	}
	return result
}

func (s *Service) advanceMatchGeneration(ctx context.Context, repository MatchGenerationRepository, userID, matchID int64, attempt int, step string, progress int) (bool, error) {
	workflow, err := s.workflowRepository()
	if err != nil {
		return false, err
	}
	run, err := workflow.GetMatchRun(ctx, userID, matchID)
	if err != nil {
		return false, err
	}
	if run.GenerationAttempt != attempt || run.Status == MatchStatusCanceled || run.Status == MatchStatusCompleted {
		return false, nil
	}
	_, err = repository.UpdateMatchGeneration(ctx, userID, matchID, attempt, MatchStatusRunning, progress, step, "", nil)
	if errors.Is(err, ErrStaleMatchGeneration) {
		return false, nil
	}
	return err == nil, err
}

func matchGenerationResponse(run MatchRun) MatchGenerationResponse {
	return MatchGenerationResponse{MatchID: run.ID, Status: run.Status, Attempt: run.GenerationAttempt, ProgressPercent: run.ProgressPercent, CurrentStep: run.CurrentStep, Result: run.Result, ErrorCode: run.ErrorCode}
}

func isTerminalMatchStatus(status string) bool {
	switch status {
	case MatchStatusCompleted, MatchStatusPartial, MatchStatusFailed, MatchStatusCanceled:
		return true
	default:
		return false
	}
}
