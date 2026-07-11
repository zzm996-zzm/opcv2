package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type fakeTaskGenerator struct {
	request ai.GenerateJSONRequest
	result  ai.GenerateJSONResult
	err     error
}

func (g *fakeTaskGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	return g.result, g.err
}

func TestServiceGeneratesAndPersistsTaskPlan(t *testing.T) {
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	generator := &fakeTaskGenerator{result: ai.GenerateJSONResult{Content: []byte(`{
		"tasks":[
			{"title":"整理访谈名单","description":"筛选首批目标客户","project":"客户验证","priority":"high","tags":["访谈"],"due_in_days":1,"tools":["CRM"],"learning":"客户访谈"},
			{"title":"完成访谈复盘","description":"归纳高频问题","project":"客户验证","priority":"medium","tags":["复盘"],"due_in_days":3,"tools":["AI助手"],"learning":"需求分析"}
		]
	}`)}}
	service := NewService(repository, WithTaskGenerator(generator))
	service.now = func() time.Time { return now }

	sourceID := int64(99)
	result, err := service.GenerateTasks(context.Background(), GenerateTasksInput{
		UserID: 42, Goal: "验证教培客户需求", SourceType: SourceLearningDiagnosis,
		SourceID: &sourceID, SourceTitle: "企业AI落地能力路径", SourceURL: "/learning/plan",
	})

	if err != nil {
		t.Fatalf("GenerateTasks() error = %v", err)
	}
	if generator.request.Feature != "tasks.generate" || generator.request.UserID != 42 || generator.request.Validate == nil {
		t.Fatalf("generator request = %+v", generator.request)
	}
	if len(result.Tasks) != 2 || len(repository.createdTasks) != 2 {
		t.Fatalf("result/repository = %+v/%+v", result, repository.createdTasks)
	}
	if result.Tasks[0].Status != StatusTodo || result.Tasks[0].UserID != 42 || result.Tasks[0].DueAt == nil || !result.Tasks[0].DueAt.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("first task = %+v", result.Tasks[0])
	}
	if result.Tasks[0].SourceType != SourceLearningDiagnosis || result.Tasks[0].SourceID == nil || *result.Tasks[0].SourceID != 99 || result.Tasks[0].SourceURL != "/learning/plan" {
		t.Fatalf("first task source = %+v", result.Tasks[0])
	}
}

func TestServiceRejectsInvalidGeneratedTaskPlan(t *testing.T) {
	repository := &fakeRepository{}
	generator := &fakeTaskGenerator{result: ai.GenerateJSONResult{Content: []byte(`{"tasks":[]}`)}}
	service := NewService(repository, WithTaskGenerator(generator))

	_, err := service.GenerateTasks(context.Background(), GenerateTasksInput{UserID: 42, Goal: "验证客户需求"})

	if !errors.Is(err, ErrInvalidGeneratedTasks) || len(repository.createdTasks) != 0 {
		t.Fatalf("err/created = %v/%+v", err, repository.createdTasks)
	}
}
