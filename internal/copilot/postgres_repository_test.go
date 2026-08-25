package copilot

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/zzm/opcv2/internal/ai"
)

func TestPostgresRepositoryCreatesThread(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO copilot_threads (user_id, title, mode, model, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(int64(42), "智能客服机会分析", ModeChat, "gpt-test", now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(99), now, now))

	repository := NewPostgresRepository(db)
	thread, err := repository.CreateThread(context.Background(), Thread{
		UserID:    42,
		Title:     "智能客服机会分析",
		Mode:      ModeChat,
		Model:     "gpt-test",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.ID != 99 {
		t.Fatalf("thread.ID = %d, want 99", thread.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesMessage(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO copilot_messages (user_id, thread_id, role, content, status, model, error_code, input_tokens, output_tokens, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`)).
		WithArgs(
			int64(42),
			int64(99),
			RoleAssistant,
			"建议先做小范围验证。",
			MessageStatusCompleted,
			"gpt-test",
			"",
			12,
			24,
			[]byte(`{}`),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(7), now))
	db.ExpectExec(regexp.QuoteMeta(`UPDATE copilot_threads SET updated_at = NOW() WHERE user_id = $1 AND id = $2`)).
		WithArgs(int64(42), int64(99)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repository := NewPostgresRepository(db)
	message, err := repository.CreateMessage(context.Background(), Message{
		UserID:       42,
		ThreadID:     99,
		Role:         RoleAssistant,
		Content:      "建议先做小范围验证。",
		Status:       MessageStatusCompleted,
		Model:        "gpt-test",
		InputTokens:  12,
		OutputTokens: 24,
		CreatedAt:    now,
	})
	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	if message.ID != 7 {
		t.Fatalf("message.ID = %d, want 7", message.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsMessagesWithoutFailedPlaceholders(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, thread_id, role, content, status, model, error_code, input_tokens, output_tokens, metadata, created_at
		FROM (
			SELECT id, user_id, thread_id, role, content, status, model, error_code, input_tokens, output_tokens, metadata, created_at
			FROM copilot_messages
			WHERE user_id = $1 AND thread_id = $2 AND status <> 'failed'
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		) AS recent_messages
		ORDER BY created_at ASC, id ASC
	`)).
		WithArgs(int64(42), int64(99), 50).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "thread_id", "role", "content", "status", "model", "error_code",
			"input_tokens", "output_tokens", "metadata", "created_at",
		}).AddRow(
			int64(7), int64(42), int64(99), RoleAssistant, "正常回复", MessageStatusCompleted,
			"deepseek", "", 12, 24, []byte(`{}`), now,
		))

	repository := NewPostgresRepository(db)
	messages, err := repository.ListMessages(context.Background(), 42, 99, 50)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].Content != "正常回复" {
		t.Fatalf("messages = %+v", messages)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpsertsMemory(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO copilot_memories (user_id, key, value, confidence, source, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id, key)
		DO UPDATE SET value = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.value ELSE EXCLUDED.value END,
			confidence = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.confidence ELSE EXCLUDED.confidence END,
			source = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.source ELSE EXCLUDED.source END,
			status = CASE WHEN copilot_memories.status = 'active' THEN copilot_memories.status ELSE EXCLUDED.status END,
			updated_at = NOW()
		RETURNING id, status, created_at, updated_at
	`)).
		WithArgs(int64(42), "industry", "教培", 0.9, "manual", MemoryStatusActive, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "status", "created_at", "updated_at"}).AddRow(int64(7), MemoryStatusActive, now, now))

	repository := NewPostgresRepository(db)
	memory, err := repository.UpsertMemory(context.Background(), Memory{
		UserID:     42,
		Key:        "industry",
		Value:      "教培",
		Confidence: 0.9,
		Source:     "manual",
		Status:     MemoryStatusActive,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpsertMemory() error = %v", err)
	}
	if memory.ID != 7 {
		t.Fatalf("memory.ID = %d, want 7", memory.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpsertsPendingMemoryWithoutOverwritingActiveFields(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO copilot_memories (user_id, key, value, confidence, source, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id, key)
		DO UPDATE SET value = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.value ELSE EXCLUDED.value END,
			confidence = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.confidence ELSE EXCLUDED.confidence END,
			source = CASE WHEN copilot_memories.status = 'active' AND EXCLUDED.status = 'pending' THEN copilot_memories.source ELSE EXCLUDED.source END,
			status = CASE WHEN copilot_memories.status = 'active' THEN copilot_memories.status ELSE EXCLUDED.status END,
			updated_at = NOW()
		RETURNING id, status, created_at, updated_at
	`)).
		WithArgs(int64(42), "industry", "AI 候选", 0.6, "copilot", MemoryStatusPending, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "status", "created_at", "updated_at"}).AddRow(int64(7), MemoryStatusActive, now, now))

	repository := NewPostgresRepository(db)
	memory, err := repository.UpsertMemory(context.Background(), Memory{
		UserID:     42,
		Key:        "industry",
		Value:      "AI 候选",
		Confidence: 0.6,
		Source:     "copilot",
		Status:     MemoryStatusPending,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpsertMemory() error = %v", err)
	}
	if memory.Status != MemoryStatusActive {
		t.Fatalf("memory.Status = %q, want %q", memory.Status, MemoryStatusActive)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesMemory(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE copilot_memories
		SET key = CASE WHEN $3 = '' THEN key ELSE $3 END,
			value = CASE WHEN $4 = '' THEN value ELSE $4 END,
			status = CASE WHEN $5 = '' THEN status ELSE $5 END,
			updated_at = NOW()
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, key, value, confidence, source, status, created_at, updated_at
	`)).WithArgs(int64(42), int64(7), "", "", MemoryStatusInactive).WillReturnRows(
		pgxmock.NewRows([]string{"id", "user_id", "key", "value", "confidence", "source", "status", "created_at", "updated_at"}).AddRow(int64(7), int64(42), "industry", "教培", 0.9, "manual", MemoryStatusInactive, now, now),
	)
	repository := NewPostgresRepository(db)
	memory, err := repository.UpdateMemory(context.Background(), MemoryUpdateInput{UserID: 42, ID: 7, Status: MemoryStatusInactive})
	if err != nil || memory.Status != MemoryStatusInactive {
		t.Fatalf("UpdateMemory() = %+v, error = %v", memory, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAndListsFiles(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO copilot_files (user_id, name, mime_type, size_bytes, content, status, source, sha256, extracted_chars, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(int64(42), "竞品对比.txt", "text/plain", 48, "小鹅通：私域工具强。", FileStatusReady, FileSourceUpload, "abc123", 11, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(17), now, now))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, name, mime_type, size_bytes, '' AS content, status, source, sha256, extracted_chars, error_code, created_at, updated_at
		FROM copilot_files
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "mime_type", "size_bytes", "content", "status", "source", "sha256", "extracted_chars", "error_code", "created_at", "updated_at"}).
			AddRow(int64(17), int64(42), "竞品对比.txt", "text/plain", 48, "", FileStatusReady, FileSourceUpload, "abc123", 11, "", now, now))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, name, mime_type, size_bytes, content, status, source, sha256, extracted_chars, error_code, created_at, updated_at
		FROM copilot_files
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(17)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "mime_type", "size_bytes", "content", "status", "source", "sha256", "extracted_chars", "error_code", "created_at", "updated_at"}).
			AddRow(int64(17), int64(42), "竞品对比.txt", "text/plain", 48, "小鹅通：私域工具强。", FileStatusReady, FileSourceUpload, "abc123", 11, "", now, now))
	db.ExpectExec(regexp.QuoteMeta(`DELETE FROM copilot_files WHERE user_id = $1 AND id = $2`)).
		WithArgs(int64(42), int64(17)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	file, err := repository.CreateFile(context.Background(), File{
		UserID:         42,
		Name:           "竞品对比.txt",
		MimeType:       "text/plain",
		SizeBytes:      48,
		Content:        "小鹅通：私域工具强。",
		Status:         FileStatusReady,
		Source:         FileSourceUpload,
		SHA256:         "abc123",
		ExtractedChars: 11,
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if file.ID != 17 {
		t.Fatalf("file.ID = %d, want 17", file.ID)
	}

	files, err := repository.ListFiles(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 1 || files[0].ID != 17 || files[0].Content != "" || files[0].Status != FileStatusReady {
		t.Fatalf("files = %+v", files)
	}
	detail, err := repository.GetFile(context.Background(), 42, 17)
	if err != nil || detail.Content == "" {
		t.Fatalf("GetFile() = %+v, %v", detail, err)
	}
	if err := repository.DeleteFile(context.Background(), 42, 17); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsAIRuns(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 7, 1, 10, 25, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Second)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			id,
			user_id,
			feature,
			prompt_version,
			provider,
			model,
			status,
			error_code,
			error_message,
			input_tokens,
			output_tokens,
			latency_ms,
			created_at,
			updated_at
		FROM ai_runs
		WHERE user_id = $1
		  AND feature LIKE $2
		ORDER BY created_at DESC
		LIMIT $3
	`)).
		WithArgs(int64(42), "copilot.%", 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id",
			"user_id",
			"feature",
			"prompt_version",
			"provider",
			"model",
			"status",
			"error_code",
			"error_message",
			"input_tokens",
			"output_tokens",
			"latency_ms",
			"created_at",
			"updated_at",
		}).AddRow(
			int64(11),
			int64(42),
			"copilot.model_smoke",
			"copilot_model_smoke_v1",
			"openai-compatible",
			"deepseek",
			ai.StatusFailed,
			"provider_unavailable",
			"AI provider is unavailable",
			12,
			0,
			2080,
			createdAt,
			updatedAt,
		))

	repository := NewPostgresRepository(db)
	runs, err := repository.ListRuns(context.Background(), 42, "copilot.", 20)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs) != 1 || runs[0].ID != 11 || runs[0].Feature != "copilot.model_smoke" {
		t.Fatalf("runs = %+v", runs)
	}
	if len(runs[0].Request) != 0 || len(runs[0].Response) != 0 {
		t.Fatalf("request/response should not be loaded, run = %+v", runs[0])
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
