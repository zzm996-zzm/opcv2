package projects

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryPersistsProjectMatchGenerationIntegration(t *testing.T) {
	databaseURL := os.Getenv("OPCV2_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OPCV2_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer db.Close()
	var userID int64
	if err := db.QueryRow(ctx, `INSERT INTO users (nickname, phone, agreement_accepted_at) VALUES ('项目匹配集成测试', $1, NOW()) RETURNING id`, fmt.Sprintf("%d", time.Now().UnixNano())).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	repository := NewPostgresRepository(db)
	now := time.Now().UTC()
	run, created, err := repository.CreateMatchRun(ctx, MatchRun{
		UserID: userID, WorkflowVersion: 2, Need: "做内容服务", IdempotencyKey: fmt.Sprintf("integration-%d", now.UnixNano()),
		Status: MatchStatusReady, InputSnapshot: MatchInputSnapshot{Need: "做内容服务"}, ParsedProfile: map[string]any{"team_size": 1},
		FieldSources: map[string][]MatchFieldSource{"team_size": {{Type: "profile_patch", Locator: "team_size"}}},
		Completeness: 0.9, Revision: 1, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil || !created {
		t.Fatalf("CreateMatchRun() created/error = %v/%v", created, err)
	}
	queued, err := repository.PrepareMatchGeneration(ctx, userID, run.ID)
	if err != nil {
		t.Fatalf("PrepareMatchGeneration() error = %v", err)
	}
	if queued.Status != MatchStatusQueued || queued.GenerationAttempt != 1 {
		t.Fatalf("queued = %+v", queued)
	}
	running, err := repository.UpdateMatchGeneration(ctx, userID, run.ID, 1, MatchStatusRunning, 30, MatchStepRetrievingKB, "", nil)
	if err != nil {
		t.Fatalf("UpdateMatchGeneration() error = %v", err)
	}
	if running.ProgressPercent != 30 || running.CurrentStep != MatchStepRetrievingKB {
		t.Fatalf("running = %+v", running)
	}
	canceled, err := repository.CancelMatchGeneration(ctx, userID, run.ID)
	if err != nil {
		t.Fatalf("CancelMatchGeneration() error = %v", err)
	}
	if canceled.Status != MatchStatusCanceled || canceled.CanceledAt == nil {
		t.Fatalf("canceled = %+v", canceled)
	}
	events, err := repository.ListMatchProgressEvents(ctx, userID, run.ID, 0)
	if err != nil {
		t.Fatalf("ListMatchProgressEvents() error = %v", err)
	}
	if len(events) != 3 || events[0].Event != MatchStepQueued || events[2].Event != MatchStepCanceled {
		t.Fatalf("events = %+v", events)
	}
}
