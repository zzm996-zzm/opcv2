package sandbox

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeV2FollowUpRepository struct {
	*fakeV2Repository
	followUps []V2FollowUp
}

func (r *fakeV2FollowUpRepository) CreateV2FollowUp(_ context.Context, followUp V2FollowUp) (V2FollowUp, error) {
	followUp.ID = int64(len(r.followUps) + 1)
	r.followUps = append(r.followUps, followUp)
	return followUp, nil
}

func (r *fakeV2FollowUpRepository) ListV2FollowUps(_ context.Context, userID, runID int64) ([]V2FollowUp, error) {
	if r.run.UserID != userID || r.run.ID != runID {
		return nil, ErrV2RunNotFound
	}
	return append([]V2FollowUp(nil), r.followUps...), nil
}

func TestAskV2RoleUsesOnlyFrozenSelectedRoleOutput(t *testing.T) {
	repository := &fakeV2FollowUpRepository{fakeV2Repository: newFakeV2Repository()}
	repository.run = completedV2Run()
	repository.run.UserID = 42
	repository.run.RunRoles = []V2RunRole{
		{RunID: 99, RoleCode: "investor", RoleSessionID: "investor-session", SystemPrompt: "investor prompt", InputHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Status: "done", Output: &V2RoleOutput{RoleCode: "investor", Content: "关注留存"}},
		{RunID: 99, RoleCode: "customer", RoleSessionID: "customer-session", SystemPrompt: "customer prompt", InputHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Status: "done", Output: &V2RoleOutput{RoleCode: "customer", Content: "关注体验"}},
	}
	generator := &fakeGenerator{content: []byte(`{"answer":"先验证留存和单位经济模型。"}`)}
	service := NewService(repository, generator)
	message, err := service.AskV2Role(context.Background(), AskV2RoleInput{UserID: 42, RunID: 99, RoleCode: "investor", Question: "先验证什么？"})
	if err != nil || message.Answer == "" {
		t.Fatalf("message/error = %+v/%v", message, err)
	}
	if !strings.Contains(generator.request.UserPrompt, "关注留存") || strings.Contains(generator.request.UserPrompt, "关注体验") || strings.Contains(generator.request.UserPrompt, "customer-session") {
		t.Fatalf("prompt leaked another role: %s", generator.request.UserPrompt)
	}
}

func TestAskV2RoleRejectsUnfinishedOrUnknownRole(t *testing.T) {
	repository := &fakeV2FollowUpRepository{fakeV2Repository: newFakeV2Repository()}
	repository.run = completedV2Run()
	repository.run.UserID = 42
	repository.run.RunRoles = []V2RunRole{{RunID: 99, RoleCode: "investor", Status: "failed"}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"answer":"x"}`)})
	for _, role := range []string{"investor", "customer"} {
		if _, err := service.AskV2Role(context.Background(), AskV2RoleInput{UserID: 42, RunID: 99, RoleCode: role, Question: "为什么？"}); !errors.Is(err, ErrV2InvalidRequest) {
			t.Fatalf("role/error = %s/%v", role, err)
		}
	}
}
