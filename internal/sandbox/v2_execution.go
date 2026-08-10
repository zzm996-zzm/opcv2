package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
	"github.com/zzm/opcv2/internal/membership"
)

type V2ExecutionRepository interface {
	V2Repository
	PrepareV2Run(ctx context.Context, run V2SandboxRun, roles []V2RunRole, inputHash string, routing map[string]any) (V2SandboxRun, error)
	FailV2RunStart(ctx context.Context, userID, runID int64, errorCode string) error
	StopV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error)
	AppendV2Event(ctx context.Context, event V2ProgressEvent, userID int64) (V2ProgressEvent, error)
	ListV2Events(ctx context.Context, userID, runID, afterID int64, limit int) ([]V2ProgressEvent, error)
}

func (s *Service) StartV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	repository, err := s.v2ExecutionRepository()
	if err != nil || s.queue == nil {
		return V2SandboxRun{}, ErrServiceNotReady
	}
	run, err := repository.GetV2Run(ctx, userID, runID)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if run.Status == V2StatusRunning || run.Status == V2StatusDone || run.Status == V2StatusPartial {
		return run, nil
	}
	if run.Status != V2StatusReady {
		return V2SandboxRun{}, ErrV2StartConflict
	}
	roles, err := normalizeV2Roles(run.Roles)
	if err != nil {
		return V2SandboxRun{}, err
	}
	configs, err := repository.ListV2RoleConfigs(ctx)
	if err != nil {
		return V2SandboxRun{}, err
	}
	configByCode := make(map[string]V2RoleConfig, len(configs))
	for _, config := range configs {
		configByCode[config.Code] = config
	}
	inputHash, err := hashV2Input(run)
	if err != nil {
		return V2SandboxRun{}, err
	}
	routing := make(map[string]any, len(roles))
	roleRuns := make([]V2RunRole, 0, len(roles))
	for index, roleCode := range roles {
		config, ok := configByCode[roleCode]
		if !ok {
			return V2SandboxRun{}, ErrV2InvalidRoles
		}
		routing[roleCode] = map[string]any{"route": config.DefaultModelRoute, "prompt_version": config.PromptVersion}
		roleRuns = append(roleRuns, V2RunRole{
			RunID: run.ID, RoleCode: roleCode, Seq: index + 1,
			RoleSessionID: v2RoleSessionID(run, roleCode, inputHash), ModelRoute: config.DefaultModelRoute,
			PromptVersion: config.PromptVersion, Dimensions: append([]string(nil), config.Dimensions...),
			SystemPrompt: config.SystemPrompt, InputHash: inputHash, Status: "queued",
		})
	}
	run.Roles = roles
	run, err = repository.PrepareV2Run(ctx, run, roleRuns, inputHash, routing)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if s.quota != nil {
		_, _ = s.quota.CheckAndConsume(ctx, membership.ConsumeInput{
			UserID: userID, FeatureKey: membership.FeatureSandboxRuns, Amount: 1,
			IdempotencyKey: fmt.Sprintf("sandbox-v2-run-%d-revision-%d", run.ID, run.Revision),
		})
	}
	err = s.queue.Enqueue(ctx, jobs.Job{
		Type: jobs.TypeSandboxV2Run, IdempotencyKey: fmt.Sprintf("sandbox-v2-run-%d-revision-%d", run.ID, run.Revision),
		Payload:  map[string]any{"user_id": userID, "run_id": run.ID, "revision": run.Revision},
		MaxRetry: 1, Timeout: 15 * time.Minute,
	})
	if err != nil {
		_ = repository.FailV2RunStart(ctx, userID, runID, "queue_unavailable")
		return V2SandboxRun{}, err
	}
	return run, nil
}

func (s *Service) StopV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	repository, err := s.v2ExecutionRepository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	run, err := repository.StopV2Run(ctx, userID, runID)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if run.Status == V2StatusPartial {
		if orchestrator, ok := s.repository.(V2OrchestrationRepository); ok && s.generator != nil {
			fullRun, getErr := orchestrator.GetV2Run(ctx, userID, runID)
			if getErr == nil {
				completed, failed := v2RoleResults(fullRun.RunRoles)
				_, _ = orchestrator.AppendV2Event(ctx, V2ProgressEvent{RunID: runID, Event: V2EventReportStart, Payload: map[string]any{"stopped": true}}, userID)
				report, reportSessionID, inputTokens, outputTokens, modelGenerated := s.synthesizeV2Report(ctx, fullRun, completed, failed)
				_ = orchestrator.SaveV2Report(ctx, userID, runID, reportSessionID, report, completed, failed, inputTokens, outputTokens, modelGenerated)
			}
		}
	}
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{
		RunID: runID, Event: V2EventRunDone, Payload: map[string]any{"status": run.Status, "stopped": true}, CreatedAt: s.now(),
	}, userID)
	return repository.GetV2Run(ctx, userID, runID)
}

func (s *Service) ListV2Events(ctx context.Context, userID, runID, afterID int64, limit int) ([]V2ProgressEvent, error) {
	repository, err := s.v2ExecutionRepository()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	return repository.ListV2Events(ctx, userID, runID, afterID, limit)
}

func (s *Service) v2ExecutionRepository() (V2ExecutionRepository, error) {
	repository, ok := s.repository.(V2ExecutionRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func hashV2Input(run V2SandboxRun) (string, error) {
	value := struct {
		Product     V2Product        `json:"product"`
		Context     V2RunContext     `json:"context"`
		Assumptions []string         `json:"assumptions"`
		Roles       []string         `json:"roles"`
		Evidence    []map[string]any `json:"evidence"`
	}{run.Product, run.Context, run.Assumptions, run.Roles, run.EvidencePack}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func v2RoleSessionID(run V2SandboxRun, roleCode, inputHash string) string {
	value := sha256.Sum256([]byte(fmt.Sprintf("%d:%d:%s:%s", run.ID, run.Revision, roleCode, inputHash)))
	return "role_" + hex.EncodeToString(value[:12])
}
