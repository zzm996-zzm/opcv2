package sandbox

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryRecordsSandboxEventIdempotently(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	userID, runID := int64(42), int64(99)
	db.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sandbox_analytics_events (event_id,event_name,user_id,visitor_hash,run_id,route,properties,occurred_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (event_id) DO NOTHING RETURNING id`)).
		WithArgs("sandbox-event-001", SandboxEventReportView, &userID, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", &runID, "/sandbox-runs/99/report", map[string]any{"status": "done"}, now, now).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
	recorded, err := NewPostgresRepository(db).RecordSandboxEvent(context.Background(), SandboxAnalyticsEvent{EventID: "sandbox-event-001", EventName: SandboxEventReportView, UserID: &userID, RunID: &runID, VisitorHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Route: "/sandbox-runs/99/report", Properties: map[string]any{"status": "done"}, OccurredAt: now, CreatedAt: now})
	if err != nil || !recorded {
		t.Fatalf("recorded/error=%v/%v", recorded, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesOwnedV2FollowUp(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	db.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sandbox_run_follow_ups (run_id,user_id,role_code,question,answer,input_context_hash,created_at) SELECT s.id,s.user_id,r.role_code,$4,$5,$6,$7 FROM sandbox_sessions s JOIN sandbox_run_roles r ON r.run_id=s.id AND r.role_code=$3 AND r.status='done' WHERE s.id=$1 AND s.user_id=$2 AND s.sandbox_version=2 RETURNING id,created_at`)).
		WithArgs(int64(99), int64(42), "investor", "先验证什么？", "验证留存", hash, now).WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(1), now))
	item, err := NewPostgresRepository(db).CreateV2FollowUp(context.Background(), V2FollowUp{RunID: 99, UserID: 42, RoleCode: "investor", Question: "先验证什么？", Answer: "验证留存", InputContextHash: hash, CreatedAt: now})
	if err != nil || item.ID != 1 {
		t.Fatalf("item/error=%+v/%v", item, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
