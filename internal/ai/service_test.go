package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type fakeRepository struct {
	created   []Run
	completed []completedRun
	failed    []failedRun
	nextID    int64
}

type completedRun struct {
	id     int64
	result RunResult
}

type failedRun struct {
	id      int64
	failure RunFailure
}

func (r *fakeRepository) CreateRun(_ context.Context, run Run) (Run, error) {
	if r.nextID == 0 {
		r.nextID = 100
	}
	run.ID = r.nextID
	r.nextID++
	r.created = append(r.created, run)
	return run, nil
}

func (r *fakeRepository) CompleteRun(_ context.Context, id int64, result RunResult) error {
	r.completed = append(r.completed, completedRun{id: id, result: result})
	return nil
}

func (r *fakeRepository) FailRun(_ context.Context, id int64, failure RunFailure) error {
	r.failed = append(r.failed, failedRun{id: id, failure: failure})
	return nil
}

type fakeProvider struct {
	responses []ProviderResponse
	errs      []error
	requests  []ProviderRequest
}

func (p *fakeProvider) Stream(_ context.Context, request ProviderRequest, onDelta func([]byte) error) (ProviderResponse, error) {
	p.requests = append(p.requests, request)
	index := len(p.requests) - 1
	if index < len(p.errs) && p.errs[index] != nil {
		return ProviderResponse{}, p.errs[index]
	}
	response := ProviderResponse{}
	if index < len(p.responses) {
		response = p.responses[index]
	}
	if len(response.Content) > 0 {
		if err := onDelta(response.Content); err != nil {
			return ProviderResponse{}, err
		}
	}
	return response, nil
}

func (p *fakeProvider) Generate(_ context.Context, request ProviderRequest) (ProviderResponse, error) {
	p.requests = append(p.requests, request)
	index := len(p.requests) - 1
	if index < len(p.errs) && p.errs[index] != nil {
		return ProviderResponse{}, p.errs[index]
	}
	if index < len(p.responses) {
		return p.responses[index], nil
	}
	return ProviderResponse{}, nil
}

type typedPayload struct {
	Name string `json:"name"`
}

func validateTypedPayload(data []byte) error {
	var payload typedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if payload.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func TestServiceGeneratesJSONAndCompletesRun(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		responses: []ProviderResponse{{
			Content:      []byte(`{"name":"AI短视频脚本工作室"}`),
			InputTokens:  12,
			OutputTokens: 34,
		}},
	}
	service := NewService(repository, provider, Config{
		Provider: "development",
		Model:    "dev-model",
	})
	service.now = func() time.Time { return time.Unix(100, 0) }

	result, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
		UserID:         42,
		Feature:        "projects.match",
		PromptVersion:  "project_match_v1",
		SystemPrompt:   "Return JSON only.",
		UserPrompt:     "match projects",
		SchemaName:     "project_match",
		Validate:       validateTypedPayload,
		RepairAttempts: 1,
	})
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if string(result.Content) != `{"name":"AI短视频脚本工作室"}` {
		t.Fatalf("result.Content = %s", result.Content)
	}
	if len(repository.created) != 1 || repository.created[0].Status != StatusPending {
		t.Fatalf("created runs = %+v", repository.created)
	}
	if len(repository.completed) != 1 || repository.completed[0].id != 100 {
		t.Fatalf("completed runs = %+v", repository.completed)
	}
	if repository.completed[0].result.InputTokens != 12 || repository.completed[0].result.OutputTokens != 34 {
		t.Fatalf("completed result = %+v", repository.completed[0].result)
	}
	if len(repository.failed) != 0 {
		t.Fatalf("failed runs = %+v", repository.failed)
	}
}

func TestServiceStreamsTextAndCompletesRun(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{responses: []ProviderResponse{{Content: []byte("实时回答"), InputTokens: 3, OutputTokens: 4}}}
	service := NewService(repository, provider, Config{Provider: "development", Model: "dev-model"})
	var streamed bytes.Buffer

	result, err := service.GenerateTextStream(context.Background(), GenerateTextRequest{
		UserID: 42, Feature: "copilot.chat_stream", PromptVersion: "copilot_chat_stream_v1", UserPrompt: "分析机会",
	}, func(delta []byte) error {
		_, _ = streamed.Write(delta)
		return nil
	})

	if err != nil {
		t.Fatalf("GenerateTextStream() error = %v", err)
	}
	if streamed.String() != "实时回答" || result.Content != "实时回答" {
		t.Fatalf("stream/result = %q/%q", streamed.String(), result.Content)
	}
	if len(repository.completed) != 1 || string(repository.completed[0].result.Response) != "实时回答" {
		t.Fatalf("completed = %+v", repository.completed)
	}
}

func TestServiceRetriesInvalidJSONOnceWithRepairInstruction(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		responses: []ProviderResponse{
			{Content: []byte(`{"title":"missing required name"}`)},
			{Content: []byte(`{"name":"修复后的项目"}`)},
		},
	}
	service := NewService(repository, provider, Config{Provider: "development", Model: "dev-model"})
	service.now = func() time.Time { return time.Unix(200, 0) }

	result, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
		UserID:         42,
		Feature:        "projects.match",
		PromptVersion:  "project_match_v1",
		SystemPrompt:   "Return JSON only.",
		UserPrompt:     "match projects",
		SchemaName:     "project_match",
		Validate:       validateTypedPayload,
		RepairAttempts: 1,
	})
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}
	if string(result.Content) != `{"name":"修复后的项目"}` {
		t.Fatalf("result.Content = %s", result.Content)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("provider requests = %+v", provider.requests)
	}
	if provider.requests[1].Attempt != 1 || provider.requests[1].RepairInstruction == "" {
		t.Fatalf("repair request = %+v", provider.requests[1])
	}
	if len(repository.completed) != 1 || len(repository.failed) != 0 {
		t.Fatalf("completed = %+v failed = %+v", repository.completed, repository.failed)
	}
}

func TestServiceFailsRunAfterInvalidJSONRetry(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		responses: []ProviderResponse{
			{Content: []byte(`{"title":"bad"}`)},
			{Content: []byte(`not-json`)},
		},
	}
	service := NewService(repository, provider, Config{Provider: "development", Model: "dev-model"})
	service.now = func() time.Time { return time.Unix(300, 0) }

	_, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
		UserID:         42,
		Feature:        "projects.match",
		PromptVersion:  "project_match_v1",
		SystemPrompt:   "Return JSON only.",
		UserPrompt:     "match projects",
		SchemaName:     "project_match",
		Validate:       validateTypedPayload,
		RepairAttempts: 1,
	})
	if !errors.Is(err, ErrInvalidModelJSON) {
		t.Fatalf("GenerateJSON() error = %v, want ErrInvalidModelJSON", err)
	}
	if len(repository.completed) != 0 || len(repository.failed) != 1 {
		t.Fatalf("completed = %+v failed = %+v", repository.completed, repository.failed)
	}
	if repository.failed[0].failure.Code != ErrorInvalidModelJSON {
		t.Fatalf("failure = %+v", repository.failed[0].failure)
	}
}

func TestServiceClassifiesProviderTimeout(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		errs: []error{ErrProviderTimeout},
	}
	service := NewService(repository, provider, Config{Provider: "development", Model: "dev-model"})
	service.now = func() time.Time { return time.Unix(400, 0) }

	_, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
		UserID:        42,
		Feature:       "projects.match",
		PromptVersion: "project_match_v1",
		SystemPrompt:  "Return JSON only.",
		UserPrompt:    "match projects",
		SchemaName:    "project_match",
		Validate:      validateTypedPayload,
	})
	if !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("GenerateJSON() error = %v, want ErrProviderTimeout", err)
	}
	if len(repository.failed) != 1 || repository.failed[0].failure.Code != ErrorProviderTimeout {
		t.Fatalf("failed runs = %+v", repository.failed)
	}
}

func TestServiceClassifiesProviderConfigurationFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
		msg  string
	}{
		{name: "authentication", err: ErrProviderAuthentication, code: ErrorProviderAuthentication, msg: "AI provider authentication failed"},
		{name: "permission", err: ErrProviderPermission, code: ErrorProviderPermission, msg: "AI provider permission or quota was denied"},
		{name: "model not found", err: ErrProviderModelNotFound, code: ErrorProviderModelNotFound, msg: "AI provider model was not found"},
		{name: "bad request", err: ErrProviderBadRequest, code: ErrorProviderBadRequest, msg: "AI provider rejected the request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeRepository{}
			provider := &fakeProvider{errs: []error{tt.err}}
			service := NewService(repository, provider, Config{Provider: "openai-compatible", Model: "deepseek"})
			service.now = func() time.Time { return time.Unix(410, 0) }

			_, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
				UserID:        42,
				Feature:       "copilot.model_smoke",
				PromptVersion: "copilot_model_smoke_v1",
				SystemPrompt:  "Return JSON only.",
				UserPrompt:    "ping",
				SchemaName:    "copilot_model_smoke",
				Validate:      validateTypedPayload,
			})
			if !errors.Is(err, tt.err) {
				t.Fatalf("GenerateJSON() error = %v, want %v", err, tt.err)
			}
			if len(repository.failed) != 1 {
				t.Fatalf("failed runs = %+v", repository.failed)
			}
			if repository.failed[0].failure.Code != tt.code || repository.failed[0].failure.Message != tt.msg {
				t.Fatalf("failure = %+v, want code=%s message=%q", repository.failed[0].failure, tt.code, tt.msg)
			}
		})
	}
}

func TestServiceMarksRunFailedWhenContextIsCanceled(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		errs: []error{context.Canceled},
	}
	service := NewService(repository, provider, Config{Provider: "development", Model: "dev-model"})
	service.now = func() time.Time { return time.Unix(450, 0) }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := service.GenerateJSON(ctx, GenerateJSONRequest{
		UserID:        42,
		Feature:       "projects.match",
		PromptVersion: "project_match_v1",
		SystemPrompt:  "Return JSON only.",
		UserPrompt:    "match projects",
		SchemaName:    "project_match",
		Validate:      validateTypedPayload,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GenerateJSON() error = %v, want context.Canceled", err)
	}
	if len(repository.failed) != 1 {
		t.Fatalf("failed runs = %+v, want one failed run", repository.failed)
	}
	if repository.failed[0].failure.Code != ErrorProviderTimeout {
		t.Fatalf("failure = %+v, want provider timeout classification", repository.failed[0].failure)
	}
}

func TestServiceLogsProviderFailureWithoutRawPromptOrSecret(t *testing.T) {
	repository := &fakeRepository{}
	provider := &fakeProvider{
		errs: []error{errors.New("provider rejected api_key=secret-private-key")},
	}
	var buffer bytes.Buffer
	service := NewService(repository, provider, Config{
		Provider: "openai",
		Model:    "gpt-test",
		Logger:   slog.New(slog.NewTextHandler(&buffer, nil)),
	})
	service.now = func() time.Time { return time.Unix(500, 0) }

	_, err := service.GenerateJSON(context.Background(), GenerateJSONRequest{
		UserID:        42,
		Feature:       "projects.match",
		PromptVersion: "project_match_v1",
		SystemPrompt:  "Return JSON only.",
		UserPrompt:    "private customer phone 13800138000",
		SchemaName:    "project_match",
		Validate:      validateTypedPayload,
	})
	if err == nil {
		t.Fatal("GenerateJSON() error = nil, want provider error")
	}

	logs := buffer.String()
	if !strings.Contains(logs, "ai_run_failed") || !strings.Contains(logs, "error_code=internal_error") || !strings.Contains(logs, "provider=openai") {
		t.Fatalf("logs = %s", logs)
	}
	if strings.Contains(logs, "secret-private-key") || strings.Contains(logs, "13800138000") {
		t.Fatalf("logs leaked private data: %s", logs)
	}
}

func TestDevelopmentProviderReturnsFeatureSpecificProjectMatchJSON(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "projects.match"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		Projects []struct {
			Title string `json:"title"`
			Score int    `json:"score"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("project match response is not JSON: %v", err)
	}
	if len(payload.Projects) == 0 || payload.Projects[0].Title == "" || payload.Projects[0].Score == 0 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestDevelopmentProviderReturnsProjectMatchAnalysisJSON(t *testing.T) {
	provider := NewDevelopmentProvider()
	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "projects.match_analysis"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		Completeness float64 `json:"completeness"`
		Questions    []struct {
			ID string `json:"id"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("project match analysis response is not JSON: %v", err)
	}
	if payload.Completeness <= 0 || len(payload.Questions) == 0 || payload.Questions[0].ID == "" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestDevelopmentProviderReturnsFeatureSpecificAnalysisJSON(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "analysis.direction"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		Cards []struct {
			Name string `json:"name"`
		} `json:"cards"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("analysis response is not JSON: %v", err)
	}
	if len(payload.Cards) == 0 || payload.Cards[0].Name == "" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestDevelopmentProviderReturnsCompetitorAnalysisForEveryTarget(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{
		Feature:    "competitor.analysis",
		UserPrompt: `{"targets":["小鹅通","有赞教育"]}`,
	})

	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var result struct {
		Competitors []struct {
			Name string `json:"name"`
		} `json:"competitors"`
	}
	if err := json.Unmarshal(response.Content, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(result.Competitors) != 2 || result.Competitors[0].Name != "小鹅通" || result.Competitors[1].Name != "有赞教育" {
		t.Fatalf("result = %+v", result)
	}
}

func TestDevelopmentProviderReturnsFeatureSpecificSandboxJSON(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "sandbox.run"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		ReportVersion       string `json:"report_version"`
		Score               int    `json:"score"`
		Summary             string `json:"summary"`
		ConsumerProbability int    `json:"consumer_probability"`
		RiskLevel           string `json:"risk_level"`
		RecommendationGrade string `json:"recommendation_grade"`
		Metrics             []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"metrics"`
		RoleSummaries []struct {
			Role string `json:"role"`
			View string `json:"view"`
		} `json:"role_summaries"`
		Risks               []string `json:"risks"`
		NextActions         []string `json:"next_actions"`
		CoreConclusions     []string `json:"core_conclusions"`
		OpportunityAnalysis []struct {
			Title  string   `json:"title"`
			Detail string   `json:"detail"`
			Tags   []string `json:"tags"`
		} `json:"opportunity_analysis"`
		RiskAnalysis []struct {
			Title  string   `json:"title"`
			Detail string   `json:"detail"`
			Tags   []string `json:"tags"`
		} `json:"risk_analysis"`
		ActionPlan []struct {
			Order    int    `json:"order"`
			Title    string `json:"title"`
			Detail   string `json:"detail"`
			Duration string `json:"duration"`
		} `json:"action_plan"`
		GrowthPath []struct {
			Stage  int    `json:"stage"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"growth_path"`
		ValidationMetrics []struct {
			Label             string `json:"label"`
			Current           string `json:"current"`
			Target            string `json:"target"`
			ConfidencePercent int    `json:"confidence_percent"`
		} `json:"validation_metrics"`
		Timeline []struct {
			Title  string `json:"title"`
			Period string `json:"period"`
		} `json:"timeline"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("sandbox response is not JSON: %v", err)
	}
	if payload.ReportVersion == "" || payload.Score == 0 || payload.Summary == "" || payload.ConsumerProbability == 0 || payload.RiskLevel == "" || payload.RecommendationGrade == "" ||
		len(payload.Metrics) == 0 || len(payload.RoleSummaries) == 0 || len(payload.Risks) == 0 || len(payload.NextActions) == 0 ||
		len(payload.CoreConclusions) == 0 || len(payload.OpportunityAnalysis) == 0 || len(payload.RiskAnalysis) == 0 || len(payload.ActionPlan) == 0 ||
		len(payload.GrowthPath) == 0 || len(payload.ValidationMetrics) == 0 || len(payload.Timeline) == 0 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestDevelopmentProviderReturnsFeatureSpecificSandboxIntakeJSON(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "sandbox.intake"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		Goal             string `json:"goal"`
		TargetUsers      string `json:"target_users"`
		Product          string `json:"product"`
		RecognizedFields []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"recognized_fields"`
		Questions []struct {
			Key         string `json:"key"`
			Title       string `json:"title"`
			Hint        string `json:"hint"`
			Placeholder string `json:"placeholder"`
			Required    bool   `json:"required"`
			MaxLength   int    `json:"max_length"`
			Position    int    `json:"position"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("sandbox intake response is not JSON: %v", err)
	}
	if payload.Goal == "" || payload.TargetUsers == "" || payload.Product == "" || len(payload.RecognizedFields) != 3 || len(payload.Questions) != 5 {
		t.Fatalf("payload = %+v", payload)
	}
	seen := make(map[string]struct{}, len(payload.Questions))
	for _, question := range payload.Questions {
		if question.Key == "" || question.Title == "" || question.Hint == "" || question.Placeholder == "" || question.MaxLength < 1 || question.MaxLength > 2000 || question.Position < 1 {
			t.Fatalf("invalid question = %+v", question)
		}
		if _, exists := seen[question.Key]; exists {
			t.Fatalf("duplicate question key = %q", question.Key)
		}
		seen[question.Key] = struct{}{}
	}
}

func TestDevelopmentProviderReturnsFeatureSpecificSandboxFollowUpJSON(t *testing.T) {
	provider := NewDevelopmentProvider()

	response, err := provider.Generate(context.Background(), ProviderRequest{Feature: "sandbox.follow_up"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var payload struct {
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(response.Content, &payload); err != nil {
		t.Fatalf("sandbox follow-up response is not JSON: %v", err)
	}
	if payload.Answer == "" {
		t.Fatalf("payload = %+v", payload)
	}
}
