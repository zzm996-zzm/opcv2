package sandbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeV2Repository struct {
	*fakeRepository
	run   V2SandboxRun
	roles []V2RoleConfig
}

func newFakeV2Repository() *fakeV2Repository {
	return &fakeV2Repository{fakeRepository: &fakeRepository{}, roles: DefaultV2RoleConfigs()}
}

func (r *fakeV2Repository) ListV2RoleConfigs(context.Context) ([]V2RoleConfig, error) {
	return append([]V2RoleConfig(nil), r.roles...), nil
}

func (r *fakeV2Repository) CreateV2Run(_ context.Context, run V2SandboxRun) (V2SandboxRun, error) {
	run.ID = 99
	r.run = run
	return run, nil
}

func (r *fakeV2Repository) GetV2Run(_ context.Context, userID, runID int64) (V2SandboxRun, error) {
	if r.run.ID != runID || r.run.UserID != userID {
		return V2SandboxRun{}, ErrV2RunNotFound
	}
	return r.run, nil
}

func (r *fakeV2Repository) UpdateV2RunDraft(_ context.Context, run V2SandboxRun, expectedRevision int) (V2SandboxRun, error) {
	if expectedRevision != r.run.Revision {
		return V2SandboxRun{}, ErrV2Revision
	}
	run.Revision++
	r.run = run
	return run, nil
}

func (r *fakeV2Repository) ListV2Runs(_ context.Context, userID int64, _ int) ([]V2SandboxRun, error) {
	if r.run.UserID != userID {
		return []V2SandboxRun{}, nil
	}
	return []V2SandboxRun{r.run}, nil
}

func (r *fakeV2Repository) ListV2RunsFiltered(_ context.Context, input V2RunListInput) ([]V2SandboxRun, error) {
	return r.ListV2Runs(context.Background(), input.UserID, input.Limit)
}

func (r *fakeV2Repository) DeleteV2Run(_ context.Context, userID, runID int64) error {
	if r.run.UserID != userID || r.run.ID != runID {
		return ErrV2RunNotFound
	}
	r.run = V2SandboxRun{}
	return nil
}

func TestV2RoleDictionaryAndDefaults(t *testing.T) {
	configs := DefaultV2RoleConfigs()
	if len(configs) != 8 {
		t.Fatalf("role configs = %d, want 8", len(configs))
	}
	defaults := DefaultV2Roles()
	if len(defaults) != 7 {
		t.Fatalf("default roles = %d, want 7", len(defaults))
	}
	required := false
	for _, role := range configs {
		if role.Code == "skeptic" {
			required = role.Required && role.DefaultSelected
		}
	}
	if !required {
		t.Fatal("skeptic must be required and selected by default")
	}
}

func TestNormalizeV2RolesRequiresThreeAndSkeptic(t *testing.T) {
	valid := []string{"customer", "investor", "skeptic"}
	if roles, err := normalizeV2Roles(valid); err != nil || len(roles) != 3 {
		t.Fatalf("normalize valid roles = %v, %v", roles, err)
	}
	for _, invalid := range [][]string{
		{"customer", "skeptic"},
		{"customer", "investor", "partner"},
		{"customer", "skeptic", "skeptic"},
	} {
		if _, err := normalizeV2Roles(invalid); !errors.Is(err, ErrV2InvalidRoles) {
			t.Fatalf("normalizeV2Roles(%v) error = %v", invalid, err)
		}
	}
}

func TestV2WorkflowCreatesQuestionsAndSkipCompletesClarification(t *testing.T) {
	repository := newFakeV2Repository()
	service := NewService(repository, nil)
	service.now = func() time.Time { return time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC) }
	run, err := service.CreateV2Run(context.Background(), CreateV2RunInput{
		UserID: 42, Product: V2Product{Name: "AI 客服"}, Context: V2RunContext{},
	})
	if err != nil {
		t.Fatalf("CreateV2Run() error = %v", err)
	}
	if run.Status != V2StatusClarifying || len(run.Questions) != 3 || len(run.Roles) != 7 || run.Done {
		t.Fatalf("created run = %+v", run)
	}
	completed, err := service.AnswerV2Run(context.Background(), AnswerV2RunInput{
		UserID: 42, RunID: run.ID, Revision: run.Revision, Skip: true,
	})
	if err != nil {
		t.Fatalf("AnswerV2Run() error = %v", err)
	}
	if completed.Status != V2StatusReady || !completed.Done || len(completed.Assumptions) != 3 {
		t.Fatalf("completed run = %+v", completed)
	}
}

func TestV2InputHashIsStableAndIncludesRoles(t *testing.T) {
	run := V2SandboxRun{Product: V2Product{Name: "AI 客服"}, Context: V2RunContext{TargetCustomer: "门店"}, Roles: []string{"customer", "investor", "skeptic"}}
	first, err := hashV2Input(run)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := hashV2Input(run)
	if first != second || len(first) != 64 {
		t.Fatalf("hashes = %q / %q", first, second)
	}
	run.Roles = append(run.Roles, "partner")
	third, _ := hashV2Input(run)
	if third == first {
		t.Fatal("role changes must alter input hash")
	}
}

func TestV2RoleValidationRequiresDimensionsAndSkepticObjections(t *testing.T) {
	base := `{
		"role_code":"skeptic","stance":"oppose","verdict":"暂不建议扩大投入","content":"需要先验证关键假设。",
		"dimension_scores":[{"code":"fatal_assumption","score":30,"basis":"证据不足","confidence":0.8,"evidence_refs":[]}],
		"risks":[{"point":"风险1","severity":"high","basis":"依据1"},{"point":"风险2","severity":"high","basis":"依据2"},{"point":"风险3","severity":"high","basis":"依据3"}],
		"recommendations":[],"questions_to_validate":[],"assumptions":[],"kill_criteria":["无付费意愿"]
	}`
	if err := validateV2RoleOutput([]byte(base), "skeptic", []string{"fatal_assumption"}); err != nil {
		t.Fatalf("valid skeptic output error = %v", err)
	}
	if err := validateV2RoleOutput([]byte(base), "skeptic", []string{"fatal_assumption", "cash_break"}); !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("missing dimension error = %v", err)
	}
}
