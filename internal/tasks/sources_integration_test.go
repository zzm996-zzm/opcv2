package tasks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryPersistsTaskSourceIntegration(t *testing.T) {
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
	if err := db.QueryRow(ctx, `
		INSERT INTO users (nickname, phone, agreement_accepted_at)
		VALUES ('来源集成测试', $1, NOW())
		RETURNING id
	`, fmt.Sprintf("%d", time.Now().UnixNano())).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	sourceID := int64(11)
	repository := NewPostgresRepository(db)
	created, err := repository.CreateTask(ctx, Task{
		UserID: userID, Title: "反击竞品更新", Project: "竞品动态监测", Status: StatusTodo, Priority: PriorityHigh,
		Tags: []string{}, Tools: []string{}, SourceType: SourceCompetitorScan, SourceID: &sourceID,
		SourceTitle: "竞品扫描：商业沙盘竞品", SourceURL: "/competitor-data", CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	loaded, err := repository.GetTask(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if loaded.SourceType != SourceCompetitorScan || loaded.SourceID == nil || *loaded.SourceID != sourceID || loaded.SourceTitle != "竞品扫描：商业沙盘竞品" || loaded.SourceURL != "/competitor-data" {
		t.Fatalf("loaded source = %+v", loaded)
	}
}
