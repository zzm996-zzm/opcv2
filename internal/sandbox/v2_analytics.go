package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

var sandboxEventProperties = map[string]map[string]bool{
	SandboxEventHomeView: {}, SandboxEventDraftCreate: {"product": true}, SandboxEventClarifyRound: {"round": true, "completeness": true},
	SandboxEventRolesSelect: {"roles": true, "count": true}, SandboxEventRunStart: {"role_count": true},
	SandboxEventRoleStart:  {"role_code": true, "role_session_id": true, "model_route": true, "prompt_version": true},
	SandboxEventRoleDone:   {"role_code": true, "role_session_id": true, "dimension_coverage": true, "latency_ms": true, "tokens": true},
	SandboxEventRoleFailed: {"role_code": true, "role_session_id": true, "error_code": true}, SandboxEventReportView: {"status": true},
	SandboxEventReportAction: {"action": true}, SandboxEventHistoryView: {"count": true}, SandboxEventQuotaBlock: {"feature": true, "used": true, "limit": true},
}

func (s *Service) recordV2RoleLifecycle(ctx context.Context, eventName string, run V2SandboxRun, role V2RunRole, properties map[string]any) {
	properties["role_code"] = role.RoleCode
	properties["role_session_id"] = role.RoleSessionID
	if eventName == SandboxEventRoleStart {
		properties["model_route"] = role.ModelRoute
		properties["prompt_version"] = role.PromptVersion
	}
	_, _ = s.RecordSandboxEvent(ctx, SandboxAnalyticsInput{
		UserID: run.UserID, RunID: run.ID, EventID: fmt.Sprintf("sandbox-%s-%d-%s-%d", eventName, run.ID, role.RoleCode, role.RetryCount),
		EventName: eventName, VisitorKey: fmt.Sprintf("user:%d", run.UserID), Route: fmt.Sprintf("/sandbox-runs/%d", run.ID), Properties: properties,
	})
}

type SandboxAnalyticsRepository interface {
	RecordSandboxEvent(context.Context, SandboxAnalyticsEvent) (bool, error)
}

func (s *Service) RecordSandboxEvent(ctx context.Context, input SandboxAnalyticsInput) (SandboxAnalyticsReceipt, error) {
	repository, ok := s.repository.(SandboxAnalyticsRepository)
	if !ok {
		return SandboxAnalyticsReceipt{}, ErrServiceNotReady
	}
	event, err := s.normalizeSandboxEvent(input)
	if err != nil {
		return SandboxAnalyticsReceipt{}, err
	}
	if event.RunID != nil {
		runs, runErr := s.v2Repository()
		if runErr != nil {
			return SandboxAnalyticsReceipt{}, runErr
		}
		if _, runErr = runs.GetV2Run(ctx, input.UserID, *event.RunID); runErr != nil {
			return SandboxAnalyticsReceipt{}, runErr
		}
	}
	recorded, err := repository.RecordSandboxEvent(ctx, event)
	if err != nil {
		return SandboxAnalyticsReceipt{}, err
	}
	return SandboxAnalyticsReceipt{EventID: event.EventID, Accepted: true, Duplicate: !recorded}, nil
}

func (s *Service) normalizeSandboxEvent(input SandboxAnalyticsInput) (SandboxAnalyticsEvent, error) {
	input.EventID = strings.TrimSpace(input.EventID)
	input.EventName = strings.TrimSpace(input.EventName)
	input.VisitorKey = strings.TrimSpace(input.VisitorKey)
	input.Route = strings.TrimSpace(input.Route)
	allowed, exists := sandboxEventProperties[input.EventName]
	if input.UserID <= 0 || len(input.EventID) < 8 || len(input.EventID) > 128 || input.Route == "" || len(input.Route) > 500 || !exists {
		return SandboxAnalyticsEvent{}, ErrV2InvalidEvent
	}
	if input.Properties == nil {
		input.Properties = map[string]any{}
	}
	for key, value := range input.Properties {
		if !allowed[key] || key == "user_id" || !validSandboxAnalyticsProperty(value, 0) {
			return SandboxAnalyticsEvent{}, ErrV2InvalidEvent
		}
	}
	var runID *int64
	if input.RunID > 0 {
		runID = &input.RunID
	}
	visitor := input.VisitorKey
	if visitor == "" {
		visitor = "user:" + formatInt(input.UserID)
	}
	hash := sha256.Sum256([]byte(visitor))
	now := s.now().UTC()
	userID := input.UserID
	return SandboxAnalyticsEvent{EventID: input.EventID, EventName: input.EventName, UserID: &userID, RunID: runID, VisitorHash: hex.EncodeToString(hash[:]), Route: input.Route, Properties: input.Properties, OccurredAt: now, CreatedAt: now}, nil
}

func validSandboxAnalyticsProperty(value any, depth int) bool {
	if depth > 2 {
		return false
	}
	switch typed := value.(type) {
	case nil, bool, float64, json.Number, int, int64:
		return true
	case string:
		return len(typed) <= 2000
	case []any:
		if len(typed) > 20 {
			return false
		}
		for _, item := range typed {
			if !validSandboxAnalyticsProperty(item, depth+1) {
				return false
			}
		}
		return true
	case map[string]any:
		if len(typed) > 20 {
			return false
		}
		for key, item := range typed {
			if len(key) > 100 || !validSandboxAnalyticsProperty(item, depth+1) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
