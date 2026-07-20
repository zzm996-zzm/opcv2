package sandbox

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesSession(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sandbox_sessions (user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"验证 AI 客服项目",
			"本地教培机构",
			"AI 客服工具",
			[]byte(`["用户","投资人"]`),
			StatusDraft,
			0,
			StatusDraft,
			"",
			0,
			[]byte(`{}`),
			[]byte(`{"depth":"standard","output_style":"structured_report","generate_outline":true,"variables":{}}`),
			[]byte(`{}`),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	session, err := repository.CreateSession(context.Background(), Session{
		UserID:      42,
		Goal:        "验证 AI 客服项目",
		TargetUsers: "本地教培机构",
		Product:     "AI 客服工具",
		Roles:       []string{"用户", "投资人"},
		Status:      StatusDraft,
		CurrentStep: StatusDraft,
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID != 99 {
		t.Fatalf("session.ID = %d, want 99", session.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesOwnedSessionResult(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 100, current_step = $1, error_message = '', report = $2, updated_at = NOW()
		WHERE user_id = $3 AND id = $4 AND run_attempt = $5 AND status = $6
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).
		WithArgs(
			StatusCompleted,
			pgxmock.AnyArg(),
			int64(42),
			int64(99),
			1,
			StatusRunning,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "goal", "target_users", "product", "roles", "status", "progress_percent", "current_step", "error_message", "run_attempt", "intake", "settings", "report", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"验证 AI 客服项目",
			"本地教培机构",
			"AI 客服工具",
			[]byte(`["用户","投资人"]`),
			StatusCompleted,
			100,
			StatusCompleted,
			"",
			1,
			[]byte(`{}`),
			[]byte(`{}`),
			[]byte(`{"score":83,"summary":"可以验证","metrics":[{"label":"市场吸引力","value":"8.4"}],"role_summaries":[{"role":"用户","view":"关注效率"}],"risks":["客户教育成本高"],"next_actions":["访谈客户"]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	session, err := repository.UpdateSessionResult(context.Background(), 42, 99, 1, Report{
		Score:         83,
		Summary:       "可以验证",
		Metrics:       []Metric{{Label: "市场吸引力", Value: "8.4"}},
		RoleSummaries: []RoleSummary{{Role: "用户", View: "关注效率"}},
		Risks:         []string{"客户教育成本高"},
		NextActions:   []string{"访谈客户"},
	})
	if err != nil {
		t.Fatalf("UpdateSessionResult() error = %v", err)
	}
	if session.Status != StatusCompleted || session.Report.Score != 83 {
		t.Fatalf("session = %+v", session)
	}
	if session.Intake.Status != IntakeStatusReady || session.Intake.Questions == nil || session.Settings.Depth != RunDepthStandard || session.Settings.Variables == nil {
		t.Fatalf("legacy defaults = %+v/%+v", session.Intake, session.Settings)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryPreparesAndUpdatesOwnedSessionRun(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 13, 13, 0, 0, 0, time.UTC)
	columns := []string{"id", "user_id", "goal", "target_users", "product", "roles", "status", "progress_percent", "current_step", "error_message", "run_attempt", "intake", "settings", "report", "created_at", "updated_at"}
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 0, current_step = $1, error_message = '',
		    run_attempt = run_attempt + 1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND status IN ($4, $5, $6)
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).WithArgs(StatusQueued, int64(42), int64(99), StatusDraft, StatusFailed, StatusCanceled).
		WillReturnRows(pgxmock.NewRows(columns).AddRow(int64(99), int64(42), "验证", "客户", "产品", []byte(`["用户"]`), StatusQueued, 0, StatusQueued, "", 1, []byte(`{}`), []byte(`{}`), []byte(`{}`), now, now))
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = $2, current_step = $3, error_message = $4, updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND run_attempt = $7
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).WithArgs(StatusRunning, 20, "generating_report", "", int64(42), int64(99), 1).
		WillReturnRows(pgxmock.NewRows(columns).AddRow(int64(99), int64(42), "验证", "客户", "产品", []byte(`["用户"]`), StatusRunning, 20, "generating_report", "", 1, []byte(`{}`), []byte(`{}`), []byte(`{}`), now, now))

	repository := NewPostgresRepository(db)
	prepared, err := repository.PrepareSessionRun(context.Background(), 42, 99)
	if err != nil || prepared.RunAttempt != 1 || prepared.Status != StatusQueued {
		t.Fatalf("prepared = %+v err=%v", prepared, err)
	}
	running, err := repository.UpdateSessionProgress(context.Background(), 42, 99, 1, StatusRunning, 20, "generating_report", "")
	if err != nil || running.ProgressPercent != 20 {
		t.Fatalf("running = %+v err=%v", running, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesOwnedDraft(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 11, 18, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET goal = $1, target_users = $2, product = $3, roles = $4, updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND status = $7
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).WithArgs("验证项目", "连锁门店", "AI运营平台", []byte(`["用户视角"]`), int64(42), int64(99), StatusDraft).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "goal", "target_users", "product", "roles", "status", "progress_percent", "current_step", "error_message", "run_attempt", "intake", "settings", "report", "created_at", "updated_at"}).AddRow(
			int64(99), int64(42), "验证项目", "连锁门店", "AI运营平台", []byte(`["用户视角"]`), StatusDraft, 0, StatusDraft, "", 0, []byte(`{}`), []byte(`{}`), []byte(`{}`), now, now,
		))

	repository := NewPostgresRepository(db)
	session, err := repository.UpdateSessionDraft(context.Background(), 42, 99, DraftUpdate{
		Goal: stringPointer("验证项目"), TargetUsers: stringPointer("连锁门店"), Product: stringPointer("AI运营平台"), Roles: &[]string{"用户视角"},
	})
	if err != nil || session.Product != "AI运营平台" {
		t.Fatalf("session = %+v err = %v", session, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func stringPointer(value string) *string { return &value }

func TestPostgresRepositoryPersistsIntakeAndSettings(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	columns := []string{"id", "user_id", "goal", "target_users", "product", "roles", "status", "progress_percent", "current_step", "error_message", "run_attempt", "intake", "settings", "report", "created_at", "updated_at"}
	intakeJSON := []byte(`{"status":"questions","initial_idea":"门店助手","recognized_fields":[],"questions":[{"key":"pain","title":"痛点","hint":"","placeholder":"","required":true,"max_length":1000,"position":1,"answer":"回复慢","skipped":false}],"answered_count":1,"total_questions":1}`)
	settingsJSON := []byte(`{"depth":"deep","output_style":"concise_report","generate_outline":false,"variables":{"budget":"3万以内"}}`)

	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET intake = $1, current_step = $2, updated_at = NOW()
		WHERE user_id = $3 AND id = $4 AND status = $5
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).WithArgs(intakeJSON, IntakeStatusQuestions, int64(42), int64(99), StatusDraft).
		WillReturnRows(pgxmock.NewRows(columns).AddRow(
			int64(99), int64(42), "验证", "门店", "助手", []byte(`[]`), StatusDraft, 0, IntakeStatusQuestions, "", 0,
			intakeJSON, []byte(`{}`), []byte(`{}`), now, now,
		))
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE sandbox_sessions
		SET settings = $1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND status = $4
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`)).WithArgs(settingsJSON, int64(42), int64(99), StatusDraft).
		WillReturnRows(pgxmock.NewRows(columns).AddRow(
			int64(99), int64(42), "验证", "门店", "助手", []byte(`[]`), StatusDraft, 0, IntakeStatusQuestions, "", 0,
			intakeJSON, settingsJSON, []byte(`{}`), now, now,
		))

	repository := NewPostgresRepository(db)
	intake := Intake{
		Status: IntakeStatusQuestions, InitialIdea: "门店助手", RecognizedFields: []RecognizedField{},
		Questions:     []IntakeQuestion{{Key: "pain", Title: "痛点", Required: true, MaxLength: 1000, Position: 1, Answer: "回复慢"}},
		AnsweredCount: 1, TotalQuestions: 1,
	}
	updated, err := repository.UpdateSessionIntake(context.Background(), 42, 99, intake)
	if err != nil || updated.Intake.AnsweredCount != 1 || updated.Settings.Depth != RunDepthStandard {
		t.Fatalf("intake session = %+v err=%v", updated, err)
	}
	settings := RunSettings{Depth: RunDepthDeep, OutputStyle: OutputStyleConcise, Variables: map[string]string{"budget": "3万以内"}}
	updated, err = repository.UpdateSessionSettings(context.Background(), 42, 99, settings)
	if err != nil || updated.Settings.Depth != RunDepthDeep || updated.Settings.OutputStyle != OutputStyleConcise {
		t.Fatalf("settings session = %+v err=%v", updated, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAndListsMessages(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 11, 18, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sandbox_messages (session_id, user_id, role, question, answer, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)).WithArgs(int64(99), int64(42), "投资人视角", "关注什么", "关注留存", now).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, session_id, user_id, role, question, answer, created_at
		FROM sandbox_messages
		WHERE user_id = $1 AND session_id = $2
		ORDER BY created_at ASC, id ASC
	`)).WithArgs(int64(42), int64(99)).WillReturnRows(pgxmock.NewRows([]string{"id", "session_id", "user_id", "role", "question", "answer", "created_at"}).AddRow(
		int64(1), int64(99), int64(42), "投资人视角", "关注什么", "关注留存", now,
	))

	repository := NewPostgresRepository(db)
	created, err := repository.CreateMessage(context.Background(), Message{SessionID: 99, UserID: 42, Role: "投资人视角", Question: "关注什么", Answer: "关注留存", CreatedAt: now})
	if err != nil || created.ID != 1 {
		t.Fatalf("created = %+v err = %v", created, err)
	}
	messages, err := repository.ListMessages(context.Background(), 42, 99)
	if err != nil || len(messages) != 1 || messages[0].Answer != "关注留存" {
		t.Fatalf("messages = %+v err = %v", messages, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
