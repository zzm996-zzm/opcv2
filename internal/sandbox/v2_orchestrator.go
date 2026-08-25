package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type V2OrchestrationRepository interface {
	V2ExecutionRepository
	ClaimV2Role(ctx context.Context, userID, runID int64, roleCode string) error
	CompleteV2Role(ctx context.Context, userID, runID int64, roleCode string, output V2RoleOutput, inputTokens, outputTokens, latencyMS int) error
	FailV2Role(ctx context.Context, userID, runID int64, roleCode, errorCode string) error
	SaveV2Report(ctx context.Context, userID, runID int64, reportSessionID string, report V2SandboxReport, completedRoles, failedRoles []string, inputTokens, outputTokens int, modelGenerated bool) error
	CompleteV2Run(ctx context.Context, userID, runID int64, status string) error
	GetV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error)
}

type v2StreamingGenerator interface {
	GenerateTextStream(ctx context.Context, request ai.GenerateTextRequest, onDelta func([]byte) error) (ai.GenerateTextResult, error)
}

func (s *Service) GetV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error) {
	repository, ok := s.repository.(interface {
		GetV2Report(context.Context, int64, int64) (V2SandboxReport, error)
	})
	if !ok {
		return V2SandboxReport{}, ErrServiceNotReady
	}
	return repository.GetV2Report(ctx, userID, runID)
}

func (s *Service) GenerateV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error) {
	repository, ok := s.repository.(V2OrchestrationRepository)
	if !ok || repository == nil || s.generator == nil {
		return V2SandboxReport{}, ErrServiceNotReady
	}
	if report, err := repository.GetV2Report(ctx, userID, runID); err == nil {
		return report, nil
	}
	run, err := repository.GetV2Run(ctx, userID, runID)
	if err != nil {
		return V2SandboxReport{}, err
	}
	if run.Status != V2StatusDone && run.Status != V2StatusPartial {
		return V2SandboxReport{}, ErrV2InvalidRequest
	}
	completed, failed := v2RoleResults(run.RunRoles)
	if len(completed) == 0 {
		return V2SandboxReport{}, ErrV2InvalidRequest
	}
	report, reportSessionID, inputTokens, outputTokens, modelGenerated := s.synthesizeV2Report(ctx, run, completed, failed)
	if err := repository.SaveV2Report(ctx, userID, runID, reportSessionID, report, completed, failed, inputTokens, outputTokens, modelGenerated); err != nil {
		return V2SandboxReport{}, err
	}
	return report, nil
}

func (s *Service) ProcessV2Run(ctx context.Context, userID, runID int64, revision int) error {
	repository, ok := s.repository.(V2OrchestrationRepository)
	if !ok || repository == nil || s.generator == nil {
		return ErrServiceNotReady
	}
	runContext, cancelRun := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelRun()
	run, err := repository.GetV2Run(runContext, userID, runID)
	if err != nil {
		return err
	}
	if run.Status != V2StatusRunning || run.Revision != revision {
		return nil
	}
	configs, err := repository.ListV2RoleConfigs(runContext)
	if err != nil {
		return err
	}
	configByCode := make(map[string]V2RoleConfig, len(configs))
	for _, config := range configs {
		configByCode[config.Code] = config
	}

	semaphore := make(chan struct{}, 3)
	var group sync.WaitGroup
	var tokenUsage atomic.Int64
	for _, role := range run.RunRoles {
		role := role
		config, exists := configByCode[role.RoleCode]
		if !exists {
			_ = repository.FailV2Role(ctx, userID, runID, role.RoleCode, "role_config_missing")
			continue
		}
		group.Add(1)
		go func() {
			defer group.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-runContext.Done():
				return
			}
			if tokenUsage.Load() >= 40000 {
				_ = repository.FailV2Role(ctx, run.UserID, run.ID, role.RoleCode, "token_limit")
				_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleFailed, Payload: map[string]any{"code": "token_limit"}}, run.UserID)
				return
			}
			tokenUsage.Add(int64(s.processV2Role(runContext, repository, run, role, config)))
		}()
	}
	group.Wait()
	if runContext.Err() != nil {
		for _, role := range run.RunRoles {
			_ = repository.FailV2Role(ctx, userID, runID, role.RoleCode, "run_timeout")
		}
	}
	return s.finishV2Run(ctx, repository, userID, runID)
}

func (s *Service) ProcessV2RoleRetry(ctx context.Context, userID, runID int64, revision int, roleCode string) error {
	repository, ok := s.repository.(V2OrchestrationRepository)
	if !ok || repository == nil || s.generator == nil {
		return ErrServiceNotReady
	}
	run, err := repository.GetV2Run(ctx, userID, runID)
	if err != nil || run.Status != V2StatusRunning || run.Revision != revision {
		return err
	}
	var role V2RunRole
	found := false
	for _, candidate := range run.RunRoles {
		if candidate.RoleCode == roleCode {
			role = candidate
			found = true
			break
		}
	}
	if !found {
		return ErrV2InvalidRoles
	}
	configs, err := repository.ListV2RoleConfigs(ctx)
	if err != nil {
		return err
	}
	var config V2RoleConfig
	configFound := false
	for _, candidate := range configs {
		if candidate.Code == roleCode {
			config = candidate
			configFound = true
			break
		}
	}
	if !configFound {
		return ErrV2InvalidRoles
	}
	s.processV2Role(ctx, repository, run, role, config)
	return s.finishV2Run(ctx, repository, userID, runID)
}

func (s *Service) finishV2Run(ctx context.Context, repository V2OrchestrationRepository, userID, runID int64) error {
	run, err := repository.GetV2Run(ctx, userID, runID)
	if err != nil || run.Status != V2StatusRunning {
		return err
	}
	completed, failed := v2RoleResults(run.RunRoles)
	if len(completed) == 0 {
		if err := repository.CompleteV2Run(ctx, userID, runID, V2StatusNoResult); err != nil {
			return err
		}
		_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: runID, Event: V2EventError, Payload: map[string]any{"code": "all_roles_failed"}}, userID)
		return nil
	}
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: runID, Event: V2EventReportStart, Payload: map[string]any{"completed_roles": completed}}, userID)
	report, reportSessionID, inputTokens, outputTokens, modelGenerated := s.synthesizeV2Report(ctx, run, completed, failed)
	if err := repository.SaveV2Report(ctx, userID, runID, reportSessionID, report, completed, failed, inputTokens, outputTokens, modelGenerated); err != nil {
		return err
	}
	status := V2StatusDone
	if len(failed) > 0 {
		status = V2StatusPartial
	}
	if err := repository.CompleteV2Run(ctx, userID, runID, status); err != nil {
		return err
	}
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: runID, Event: V2EventRunDone, Payload: map[string]any{"status": status, "completed_roles": completed, "failed_roles": failed}}, userID)
	return nil
}

func (s *Service) processV2Role(ctx context.Context, repository V2OrchestrationRepository, run V2SandboxRun, role V2RunRole, config V2RoleConfig) int {
	roleContext, cancelRole := context.WithTimeout(ctx, time.Minute)
	defer cancelRole()
	if err := repository.ClaimV2Role(ctx, run.UserID, run.ID, role.RoleCode); err != nil {
		return 0
	}
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleStart, Payload: map[string]any{"session_id": role.RoleSessionID}}, run.UserID)
	s.recordV2RoleLifecycle(ctx, SandboxEventRoleStart, run, role, map[string]any{})
	prompt, err := json.Marshal(map[string]any{
		"role_code": role.RoleCode, "role_session_id": role.RoleSessionID,
		"analysis_dimensions": role.Dimensions, "product": run.Product, "context": run.Context,
		"assumptions": run.Assumptions, "evidence_pack": run.EvidencePack,
	})
	if err != nil {
		_ = repository.FailV2Role(ctx, run.UserID, run.ID, role.RoleCode, "prompt_encode_failed")
		return 0
	}
	startedAt := s.now()
	systemPrompt := role.SystemPrompt + "\n你是商业沙盘中的" + config.DisplayName + "。只根据当前项目输入独立分析，不得假设或引用其他角色的输出。" + v2RoleOutputSchemaInstruction(role.RoleCode, role.Dimensions)
	var content []byte
	var inputTokens, outputTokens int
	if streaming, ok := s.generator.(v2StreamingGenerator); ok {
		streamResult, streamErr := streaming.GenerateTextStream(roleContext, ai.GenerateTextRequest{
			UserID: run.UserID, Feature: "sandbox.role", PromptVersion: role.PromptVersion,
			SystemPrompt: systemPrompt, UserPrompt: string(prompt),
		}, func(delta []byte) error {
			_, eventErr := repository.AppendV2Event(roleContext, V2ProgressEvent{
				RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventToken,
				Payload: map[string]any{"delta": string(delta)},
			}, run.UserID)
			return eventErr
		})
		if streamErr == nil && validateV2RoleOutput([]byte(streamResult.Content), role.RoleCode, role.Dimensions) == nil {
			content = []byte(streamResult.Content)
			inputTokens = streamResult.InputTokens
			outputTokens = streamResult.OutputTokens
		}
	}
	if len(content) == 0 {
		result, err := s.generator.GenerateJSON(roleContext, ai.GenerateJSONRequest{
			UserID: run.UserID, Feature: "sandbox.role", PromptVersion: role.PromptVersion,
			SystemPrompt: systemPrompt, UserPrompt: string(prompt), SchemaName: "sandbox_v2_role_output", RepairAttempts: 1,
			Validate: func(data []byte) error { return validateV2RoleOutput(data, role.RoleCode, role.Dimensions) },
		})
		if err != nil {
			_ = repository.FailV2Role(ctx, run.UserID, run.ID, role.RoleCode, safeV2ErrorCode(err))
			_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleFailed, Payload: map[string]any{"code": safeV2ErrorCode(err)}}, run.UserID)
			s.recordV2RoleLifecycle(ctx, SandboxEventRoleFailed, run, role, map[string]any{"error_code": safeV2ErrorCode(err)})
			return 0
		}
		content = result.Content
		inputTokens = result.InputTokens
		outputTokens = result.OutputTokens
	}
	var output V2RoleOutput
	if err := json.Unmarshal(content, &output); err != nil {
		_ = repository.FailV2Role(ctx, run.UserID, run.ID, role.RoleCode, "invalid_json")
		return 0
	}
	output.IsModelGenerated = true
	latency := int(s.now().Sub(startedAt).Milliseconds())
	if err := repository.CompleteV2Role(ctx, run.UserID, run.ID, role.RoleCode, output, inputTokens, outputTokens, latency); err != nil {
		return inputTokens + outputTokens
	}
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventToken, Payload: map[string]any{"input_tokens": inputTokens, "output_tokens": outputTokens}}, run.UserID)
	_, _ = repository.AppendV2Event(ctx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleDone, Payload: map[string]any{"stance": output.Stance}}, run.UserID)
	s.recordV2RoleLifecycle(ctx, SandboxEventRoleDone, run, role, map[string]any{"dimension_coverage": 1.0, "latency_ms": latency, "tokens": inputTokens + outputTokens})
	return inputTokens + outputTokens
}

func validateV2RoleOutput(data []byte, roleCode string, dimensions []string) error {
	var output V2RoleOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return err
	}
	if output.RoleCode != roleCode || strings.TrimSpace(output.Verdict) == "" || strings.TrimSpace(output.Content) == "" {
		return ErrInvalidAIResult
	}
	if output.Stance != "support" && output.Stance != "neutral" && output.Stance != "oppose" {
		return ErrInvalidAIResult
	}
	scores := make(map[string]bool, len(output.DimensionScores))
	for _, score := range output.DimensionScores {
		if score.Score < 0 || score.Score > 100 || score.Confidence < 0 || score.Confidence > 1 || strings.TrimSpace(score.Basis) == "" {
			return ErrInvalidAIResult
		}
		scores[score.Code] = true
	}
	for _, dimension := range dimensions {
		if !scores[dimension] {
			return ErrInvalidAIResult
		}
	}
	if roleCode == "skeptic" && (len(output.Risks) < 3 || len(output.KillCriteria) == 0) {
		return ErrInvalidAIResult
	}
	return nil
}

func v2RoleResults(roles []V2RunRole) (completed, failed []string) {
	for _, role := range roles {
		switch role.Status {
		case "done":
			completed = append(completed, role.RoleCode)
		case "failed", "cancelled":
			failed = append(failed, role.RoleCode)
		}
	}
	return completed, failed
}

func (s *Service) synthesizeV2Report(ctx context.Context, run V2SandboxRun, completed, failed []string) (V2SandboxReport, string, int, int, bool) {
	reportSessionID := v2ReportSessionID(run)
	roleOutputs := make([]V2RoleOutput, 0, len(completed))
	for _, role := range run.RunRoles {
		if role.Output != nil {
			roleOutputs = append(roleOutputs, *role.Output)
		}
	}
	prompt, _ := json.Marshal(map[string]any{
		"report_session_id": reportSessionID, "product": run.Product, "context": run.Context,
		"assumptions": run.Assumptions, "role_outputs": roleOutputs, "failed_roles": failed,
	})
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID: run.UserID, Feature: "sandbox.report", PromptVersion: "sandbox_report_v1",
		SystemPrompt: "你是独立的商业沙盘报告汇总器。综合角色 JSON，明确分歧、缺失角色和假设，返回结构化 JSON。",
		UserPrompt:   string(prompt), SchemaName: "sandbox_v2_report", RepairAttempts: 1,
		Validate: validateV2Report,
	})
	if err == nil {
		var report V2SandboxReport
		if json.Unmarshal(result.Content, &report) == nil {
			report.MissingRoles = append([]string(nil), failed...)
			report.Assumptions = appendUniqueStrings(report.Assumptions, run.Assumptions...)
			report = completeV2Report(report, roleOutputs)
			report.IsModelGenerated = true
			return report, reportSessionID, result.InputTokens, result.OutputTokens, true
		}
	}
	return fallbackV2Report(run, roleOutputs, failed), reportSessionID, 0, 0, false
}

func completeV2Report(report V2SandboxReport, outputs []V2RoleOutput) V2SandboxReport {
	if report.PurchaseProbability.Basis == "" {
		report.PurchaseProbability.Basis = "基于已完成角色的模型推演，不代表真实市场统计。"
	}
	report.PurchaseProbability.IsModelGenerated = true
	if len(report.RoleTakeaways) == 0 {
		for _, output := range outputs {
			report.RoleTakeaways = append(report.RoleTakeaways, V2RoleTakeaway{Role: output.RoleCode, Stance: output.Stance, KeyPoints: output.KeyFindings, DimensionScores: output.DimensionScores})
		}
	}
	if len(report.DimensionSummary) == 0 {
		for _, output := range outputs {
			for _, score := range output.DimensionScores {
				report.DimensionSummary = append(report.DimensionSummary, V2DimensionSummary{Dimension: score.Code, Score: score.Score, Consensus: score.Basis, SupportingRoles: []string{output.RoleCode}})
			}
		}
	}
	if len(report.Scenarios) == 0 {
		report.Scenarios = map[string]V2Scenario{}
	}
	if _, ok := report.Scenarios["optimistic"]; !ok {
		report.Scenarios["optimistic"] = V2Scenario{Desc: "试点指标超出目标后复制到更多客户", Condition: "试点续费率和单位经济模型达标"}
	}
	if _, ok := report.Scenarios["base"]; !ok {
		report.Scenarios["base"] = V2Scenario{Desc: "先完成小范围试点再决定投入", Condition: "获得可复现的付费样本"}
	}
	if _, ok := report.Scenarios["pessimistic"]; !ok {
		report.Scenarios["pessimistic"] = V2Scenario{Desc: "需求或交付成本不成立，停止扩大投入", Condition: "连续验证未达到停止条件"}
	}
	for i := range report.Risk {
		if report.Risk[i].Mitigation == "" {
			report.Risk[i].Mitigation = "在下一轮试点中验证并设置停止条件。"
		}
	}
	for i := range report.Advice {
		if report.Advice[i].Effort == "medium" {
			report.Advice[i].Effort = "mid"
		}
	}
	if report.Opportunity == nil {
		report.Opportunity = []V2Insight{}
	}
	if report.Risk == nil {
		report.Risk = []V2ReportRisk{}
	}
	if report.Advice == nil {
		report.Advice = []V2Advice{}
	}
	if report.RoleTakeaways == nil {
		report.RoleTakeaways = []V2RoleTakeaway{}
	}
	if report.DimensionSummary == nil {
		report.DimensionSummary = []V2DimensionSummary{}
	}
	if report.Disagreements == nil {
		report.Disagreements = []V2Disagreement{}
	}
	if report.MissingRoles == nil {
		report.MissingRoles = []string{}
	}
	if report.Assumptions == nil {
		report.Assumptions = []string{}
	}
	for index := range report.RoleTakeaways {
		if report.RoleTakeaways[index].KeyPoints == nil {
			report.RoleTakeaways[index].KeyPoints = []string{}
		}
		if report.RoleTakeaways[index].DimensionScores == nil {
			report.RoleTakeaways[index].DimensionScores = []V2DimensionScore{}
		}
	}
	for index := range report.DimensionSummary {
		if report.DimensionSummary[index].SupportingRoles == nil {
			report.DimensionSummary[index].SupportingRoles = []string{}
		}
		if report.DimensionSummary[index].OpposingRoles == nil {
			report.DimensionSummary[index].OpposingRoles = []string{}
		}
	}
	for index := range report.Disagreements {
		if report.Disagreements[index].Views == nil {
			report.Disagreements[index].Views = []V2RoleView{}
		}
	}
	return report
}

func validateV2Report(data []byte) error {
	var report V2SandboxReport
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	if strings.TrimSpace(report.Summary) == "" || report.Feasibility.Score < 0 || report.Feasibility.Score > 100 || report.PurchaseProbability.ValuePct < 0 || report.PurchaseProbability.ValuePct > 100 {
		return ErrInvalidAIResult
	}
	return nil
}

func fallbackV2Report(run V2SandboxRun, outputs []V2RoleOutput, failed []string) V2SandboxReport {
	stances := map[string]int{"support": 0, "neutral": 0, "oppose": 0}
	takeaways := make([]V2RoleTakeaway, 0, len(outputs))
	total, count := 0, 0
	for _, output := range outputs {
		stances[output.Stance]++
		takeaways = append(takeaways, V2RoleTakeaway{Role: output.RoleCode, Stance: output.Stance, KeyPoints: output.KeyFindings, DimensionScores: output.DimensionScores})
		for _, score := range output.DimensionScores {
			total += score.Score
			count++
		}
	}
	score := 50
	if count > 0 {
		score = total / count
	}
	level := "mid"
	if score >= 75 {
		level = "high"
	} else if score < 50 {
		level = "low"
	}
	return V2SandboxReport{
		Summary:             fmt.Sprintf("%s 已完成 %d 个角色的独立推演，请优先验证报告中的关键假设。", run.Product.Name, len(outputs)),
		Feasibility:         V2Feasibility{Score: score, Level: level, Basis: "基于已完成角色维度评分的算术平均值"},
		PurchaseProbability: V2PurchaseProbability{ValuePct: score, Basis: "缺少真实成交样本，暂以角色评分作为模拟值", IsModelGenerated: false},
		RoleTakeaways:       takeaways, MissingRoles: append([]string(nil), failed...), Scenarios: map[string]V2Scenario{},
		Assumptions: appendUniqueStrings(nil, run.Assumptions...), IsModelGenerated: false,
	}
}

func v2ReportSessionID(run V2SandboxRun) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("report:%d:%d:%s", run.ID, run.Revision, run.InputContextHash)))
	return "report_" + hex.EncodeToString(hash[:12])
}

func appendUniqueStrings(values []string, additions ...string) []string {
	for _, value := range additions {
		values = appendUniqueString(values, value)
	}
	return values
}

func safeV2ErrorCode(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, ai.ErrProviderTimeout):
		return ai.ErrorProviderTimeout
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, ai.ErrProviderRateLimited):
		return ai.ErrorProviderRateLimited
	case errors.Is(err, ai.ErrProviderAuthentication):
		return ai.ErrorProviderAuthentication
	case errors.Is(err, ai.ErrProviderPermission):
		return ai.ErrorProviderPermission
	case errors.Is(err, ai.ErrProviderModelNotFound):
		return ai.ErrorProviderModelNotFound
	case errors.Is(err, ai.ErrProviderBadRequest):
		return ai.ErrorProviderBadRequest
	case errors.Is(err, ai.ErrProviderUnavailable):
		return ai.ErrorProviderUnavailable
	case errors.Is(err, ai.ErrInvalidModelJSON), errors.Is(err, ErrInvalidAIResult):
		return ai.ErrorInvalidModelJSON
	default:
		return ai.ErrorInternal
	}
}

func v2RoleOutputSchemaInstruction(roleCode string, dimensions []string) string {
	dimensionJSON := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionJSON = append(dimensionJSON, fmt.Sprintf("{\"code\":%q,\"score\":整数0-100,\"basis\":\"评分依据\",\"confidence\":0到1之间的小数,\"evidence_refs\":[]}", dimension))
	}
	extra := ""
	if roleCode == "skeptic" {
		extra = " risks 至少包含 3 项，kill_criteria 至少包含 1 项。"
	}
	return fmt.Sprintf(" 必须只返回一个 JSON 对象，不要 Markdown 或解释文字。字段必须包括 role_code=%q、stance（support/neutral/oppose）、verdict、content、dimension_scores、key_findings、risks、recommendations、questions_to_validate、assumptions、kill_criteria、is_model_generated。dimension_scores 必须逐项包含以下维度：[%s]。每项 risks 使用 {point,severity,basis}，recommendations 使用 {action,why}。%s", roleCode, strings.Join(dimensionJSON, ","), extra)
}
