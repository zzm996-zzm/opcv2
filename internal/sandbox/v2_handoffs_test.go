package sandbox

import (
	"context"
	"testing"

	"github.com/zzm/opcv2/internal/tasks"
)

type fakeSandboxTaskCreator struct{ input tasks.BatchCreateInput }

func (c *fakeSandboxTaskCreator) CreateTasks(_ context.Context, input tasks.BatchCreateInput) ([]tasks.Task, error) {
	c.input = input
	created := make([]tasks.Task, len(input.Tasks))
	for i, item := range input.Tasks {
		created[i] = tasks.Task{ID: int64(i + 1), Title: item.Title, SourceType: item.SourceType, SourceID: item.SourceID}
	}
	return created, nil
}

func TestSandboxReportCreatesSelectedAdviceTasks(t *testing.T) {
	repository := newFakeV2Repository()
	repository.run = completedV2Run()
	repository.run.UserID = 42
	repository.run.Name = "门店 AI 沙盘"
	repository.run.Product.Name = "AI运营"
	creator := &fakeSandboxTaskCreator{}
	service := NewService(repository, nil, WithTaskCreator(creator))
	result, err := service.CreateSandboxTasks(context.Background(), SandboxTaskHandoffInput{UserID: 42, RunID: 99, AdviceIndexes: []int{0}})
	if err != nil || len(result.Tasks) != 1 {
		t.Fatalf("result/error=%+v/%v", result, err)
	}
	item := creator.input.Tasks[0]
	if creator.input.UserID != 42 || item.SourceType != tasks.SourceSandboxSession || item.SourceID == nil || *item.SourceID != 99 || item.SourceURL != "/sandbox-runs/99/report" {
		t.Fatalf("input=%+v", creator.input)
	}
	if item.IdempotencyKey != "sandbox:99:advice:0" || result.Tasks[0]["source_id"] != int64(99) {
		t.Fatalf("idempotency/result=%q/%+v", item.IdempotencyKey, result)
	}
}

func TestSandboxReportRejectsInvalidAdviceSelection(t *testing.T) {
	repository := newFakeV2Repository()
	repository.run = completedV2Run()
	repository.run.UserID = 42
	service := NewService(repository, nil, WithTaskCreator(&fakeSandboxTaskCreator{}))
	for _, indexes := range [][]int{{}, {-1}, {999}, {0, 0}} {
		if _, err := service.CreateSandboxTasks(context.Background(), SandboxTaskHandoffInput{UserID: 42, RunID: 99, AdviceIndexes: indexes}); err != ErrV2InvalidRequest {
			t.Fatalf("indexes/error=%v/%v", indexes, err)
		}
	}
}

func TestSandboxGrowthHandoffCarriesValidatedPricingAndChannel(t *testing.T) {
	repository := newFakeV2Repository()
	repository.run = completedV2Run()
	repository.run.UserID = 42
	repository.run.Name = "门店 AI"
	repository.run.Product.PriceCents = 3900
	repository.run.Context.Channel = "行业伙伴"
	service := NewService(repository, nil)
	result, err := service.CreateSandboxGrowthHandoff(context.Background(), 42, 99)
	if err != nil {
		t.Fatal(err)
	}
	if result.PricingCents != 3900 || result.Channel != "行业伙伴" || result.URL != "/growth-calculator?channel=%E8%A1%8C%E4%B8%9A%E4%BC%99%E4%BC%B4&name=%E9%97%A8%E5%BA%97+AI&price_cents=3900&sandbox_run=99" {
		t.Fatalf("result=%+v", result)
	}
}
