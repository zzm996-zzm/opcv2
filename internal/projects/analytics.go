package projects

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

const (
	ProjectEventHomeView         = "project_home_view"
	ProjectEventSearch           = "project_search"
	ProjectEventFilterApply      = "project_filter_apply"
	ProjectEventCardClick        = "project_card_click"
	ProjectEventDetailView       = "project_detail_view"
	ProjectEventTabSwitch        = "project_tab_switch"
	ProjectEventLockView         = "project_lock_view"
	ProjectEventUnlockClick      = "project_unlock_click"
	ProjectEventDiagnoseSubmit   = "project_diagnose_submit"
	ProjectEventDiagnoseResult   = "project_diagnose_result"
	ProjectEventBannerSubmit     = "project_banner_submit"
	ProjectEventMatchFileUpload  = "match_file_upload"
	ProjectEventMatchParseResult = "match_file_parse_result"
	ProjectEventMatchStart       = "match_start"
	ProjectEventMatchAnswer      = "match_answer"
	ProjectEventMatchResearch    = "match_research"
	ProjectEventMatchResult      = "match_result"
	ProjectEventExploreView      = "project_explore_view"
	ProjectEventExploreSource    = "project_explore_source_click"
	ProjectEventCaseView         = "project_case_view"
	ProjectEventCaseSource       = "project_case_source_click"
	ProjectEventMatchExport      = "match_export_click"
	ProjectEventCompareAdd       = "project_compare_add"
	ProjectEventCompareView      = "project_compare_view"
)

var projectEventProperties = map[string]map[string]bool{
	ProjectEventHomeView:         propertySet("entry_from"),
	ProjectEventSearch:           propertySet("keyword", "result_count"),
	ProjectEventFilterApply:      propertySet("filters"),
	ProjectEventCardClick:        propertySet("project_id", "position", "list_type"),
	ProjectEventDetailView:       propertySet("project_id", "tab"),
	ProjectEventTabSwitch:        propertySet("project_id", "from", "to"),
	ProjectEventLockView:         propertySet("project_id", "block"),
	ProjectEventUnlockClick:      propertySet("project_id", "tab"),
	ProjectEventDiagnoseSubmit:   propertySet("project_id"),
	ProjectEventDiagnoseResult:   propertySet("project_id", "fit_score", "verdict"),
	ProjectEventBannerSubmit:     propertySet("query", "parsed_filters", "result_count"),
	ProjectEventMatchFileUpload:  propertySet("match_id", "file_type", "size_bytes"),
	ProjectEventMatchParseResult: propertySet("match_id", "file_type", "size_bytes", "status", "error_code"),
	ProjectEventMatchStart:       propertySet("match_id"),
	ProjectEventMatchAnswer:      propertySet("match_id", "rounds", "completeness"),
	ProjectEventMatchResearch:    propertySet("match_id", "kb_sufficiency", "web_trigger_reason", "source_count"),
	ProjectEventMatchResult:      propertySet("match_id", "rounds", "completeness", "kb_sufficiency", "web_trigger_reason", "source_count", "result_count"),
	ProjectEventExploreView:      propertySet("opportunity_id", "group", "query"),
	ProjectEventExploreSource:    propertySet("opportunity_id", "group", "query", "source_id"),
	ProjectEventCaseView:         propertySet("case_id", "case_type"),
	ProjectEventCaseSource:       propertySet("case_id", "case_type", "source_id", "field"),
	ProjectEventMatchExport:      propertySet("match_id", "plan", "blocked"),
	ProjectEventCompareAdd:       propertySet("project_ids"),
	ProjectEventCompareView:      propertySet("project_ids"),
}

type AnalyticsEventInput struct {
	EventID    string         `json:"event_id"`
	EventName  string         `json:"event_name"`
	VisitorKey string         `json:"visitor_key"`
	Route      string         `json:"route"`
	RefModule  string         `json:"ref_module,omitempty"`
	Properties map[string]any `json:"properties"`
}

type AnalyticsEvent struct {
	ID          int64          `json:"id"`
	EventID     string         `json:"event_id"`
	EventName   string         `json:"event_name"`
	UserID      *int64         `json:"-"`
	VisitorHash string         `json:"-"`
	Route       string         `json:"route"`
	RefModule   string         `json:"ref_module,omitempty"`
	ProjectID   *int64         `json:"project_id,omitempty"`
	Properties  map[string]any `json:"properties"`
	OccurredAt  time.Time      `json:"occurred_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

type AnalyticsReceipt struct {
	EventID   string `json:"event_id"`
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
}

type ProjectAnalyticsRepository interface {
	RecordProjectEvent(context.Context, AnalyticsEvent) (bool, error)
	RecalculateProjectHeat(context.Context, int64, time.Time) error
}

type ProjectAnalyticsApplication interface {
	RecordProjectEvent(context.Context, AnalyticsEventInput) (AnalyticsReceipt, error)
}

func propertySet(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func (s *Service) analyticsRepository() (ProjectAnalyticsRepository, error) {
	repository, ok := s.repository.(ProjectAnalyticsRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) RecordProjectEvent(ctx context.Context, input AnalyticsEventInput) (AnalyticsReceipt, error) {
	repository, err := s.analyticsRepository()
	if err != nil {
		return AnalyticsReceipt{}, err
	}
	event, err := s.normalizeProjectEvent(input)
	if err != nil {
		return AnalyticsReceipt{}, err
	}
	recorded, err := repository.RecordProjectEvent(ctx, event)
	if err != nil {
		return AnalyticsReceipt{}, err
	}
	if recorded && event.EventName == ProjectEventDetailView && event.ProjectID != nil && s.matchQueue != nil {
		_ = s.matchQueue.Enqueue(ctx, jobs.Job{
			Type: jobs.TypeProjectHeatAggregate, IdempotencyKey: "project-heat-event-" + event.EventID,
			Payload: map[string]any{"project_id": *event.ProjectID}, MaxRetry: 3, Timeout: time.Minute,
		})
	}
	return AnalyticsReceipt{EventID: event.EventID, Accepted: true, Duplicate: !recorded}, nil
}

func (s *Service) normalizeProjectEvent(input AnalyticsEventInput) (AnalyticsEvent, error) {
	input.EventID = strings.TrimSpace(input.EventID)
	input.EventName = strings.TrimSpace(input.EventName)
	input.VisitorKey = strings.TrimSpace(input.VisitorKey)
	input.Route = strings.TrimSpace(input.Route)
	input.RefModule = strings.TrimSpace(input.RefModule)
	allowed, ok := projectEventProperties[input.EventName]
	if !ok || len(input.EventID) < 8 || len(input.EventID) > 128 || len(input.VisitorKey) < 8 || len(input.VisitorKey) > 256 || input.Route == "" || len(input.Route) > 500 || len(input.RefModule) > 100 {
		return AnalyticsEvent{}, ErrInvalidEvent
	}
	if input.Properties == nil {
		input.Properties = map[string]any{}
	}
	for key, value := range input.Properties {
		if !allowed[key] || key == "user_id" || key == "heat" || !validAnalyticsProperty(value, 0) {
			return AnalyticsEvent{}, ErrInvalidEvent
		}
	}
	var projectID *int64
	for _, key := range []string{"project_id", "opportunity_id"} {
		if value, exists := input.Properties[key]; exists {
			parsed, valid := analyticsInt64(value)
			if !valid || parsed <= 0 {
				return AnalyticsEvent{}, ErrInvalidEvent
			}
			projectID = &parsed
			break
		}
	}
	hash := sha256.Sum256([]byte(input.VisitorKey))
	now := s.now().UTC()
	return AnalyticsEvent{
		EventID: input.EventID, EventName: input.EventName, VisitorHash: hex.EncodeToString(hash[:]),
		Route: input.Route, RefModule: input.RefModule, ProjectID: projectID,
		Properties: input.Properties, OccurredAt: now, CreatedAt: now,
	}, nil
}

func validAnalyticsProperty(value any, depth int) bool {
	if depth > 2 {
		return false
	}
	switch typed := value.(type) {
	case nil, bool, float64, json.Number:
		return true
	case string:
		return len(typed) <= 2000
	case []any:
		if len(typed) > 20 {
			return false
		}
		for _, item := range typed {
			if !validAnalyticsProperty(item, depth+1) {
				return false
			}
		}
		return true
	case map[string]any:
		if len(typed) > 20 {
			return false
		}
		for key, item := range typed {
			if len(key) > 100 || !validAnalyticsProperty(item, depth+1) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func analyticsInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		parsed := int64(typed)
		return parsed, float64(parsed) == typed
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func (s *Service) ProcessProjectHeat(ctx context.Context, projectID int64) error {
	repository, err := s.analyticsRepository()
	if err != nil {
		return err
	}
	if projectID <= 0 {
		return ErrInvalidEvent
	}
	return repository.RecalculateProjectHeat(ctx, projectID, s.now().Add(-30*24*time.Hour))
}

func (s *Service) enqueueProjectHeat(ctx context.Context, projectID int64) {
	if projectID <= 0 || s.matchQueue == nil {
		return
	}
	key := fmt.Sprintf("project-heat-%d-%d", projectID, s.now().UnixNano())
	_ = s.matchQueue.Enqueue(ctx, jobs.Job{Type: jobs.TypeProjectHeatAggregate, IdempotencyKey: key, Payload: map[string]any{"project_id": projectID}, MaxRetry: 3, Timeout: time.Minute})
}
