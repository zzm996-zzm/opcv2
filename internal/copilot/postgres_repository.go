package copilot

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zzm/opcv2/internal/ai"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateThread(ctx context.Context, thread Thread) (Thread, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO copilot_threads (user_id, title, mode, model, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, created_at, updated_at
	`, thread.UserID, thread.Title, thread.Mode, thread.Model, thread.CreatedAt).
		Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt)
	return thread, err
}

func (r *PostgresRepository) ListThreads(ctx context.Context, userID int64, limit int) ([]Thread, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, title, mode, model, archived_at, created_at, updated_at
		FROM copilot_threads
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []Thread
	for rows.Next() {
		thread, err := scanThread(rows)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, rows.Err()
}

func (r *PostgresRepository) GetThread(ctx context.Context, userID, id int64) (Thread, error) {
	thread, err := scanThread(r.db.QueryRow(ctx, `
		SELECT id, user_id, title, mode, model, archived_at, created_at, updated_at
		FROM copilot_threads
		WHERE user_id = $1 AND id = $2 AND archived_at IS NULL
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, ErrThreadNotFound
	}
	return thread, err
}

func (r *PostgresRepository) UpdateThreadTitle(ctx context.Context, userID, id int64, title string) (Thread, error) {
	thread, err := scanThread(r.db.QueryRow(ctx, `
		UPDATE copilot_threads
		SET title = $1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND archived_at IS NULL
		RETURNING id, user_id, title, mode, model, archived_at, created_at, updated_at
	`, title, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, ErrThreadNotFound
	}
	return thread, err
}

func (r *PostgresRepository) ArchiveThread(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE copilot_threads
		SET archived_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND id = $2 AND archived_at IS NULL
	`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrThreadNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, message Message) (Message, error) {
	metadata := message.Metadata
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	metadataBytes := []byte(metadata)
	err := r.db.QueryRow(ctx, `
		INSERT INTO copilot_messages (user_id, thread_id, role, content, status, model, error_code, input_tokens, output_tokens, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`,
		message.UserID,
		message.ThreadID,
		message.Role,
		message.Content,
		message.Status,
		message.Model,
		message.ErrorCode,
		message.InputTokens,
		message.OutputTokens,
		metadataBytes,
		message.CreatedAt,
	).Scan(&message.ID, &message.CreatedAt)
	message.Metadata = metadata
	if err != nil {
		return Message{}, err
	}
	_, _ = r.db.Exec(ctx, `UPDATE copilot_threads SET updated_at = NOW() WHERE user_id = $1 AND id = $2`, message.UserID, message.ThreadID)
	return message, nil
}

func (r *PostgresRepository) ListMessages(ctx context.Context, userID, threadID int64, limit int) ([]Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, thread_id, role, content, status, model, error_code, input_tokens, output_tokens, metadata, created_at
		FROM copilot_messages
		WHERE user_id = $1 AND thread_id = $2
		ORDER BY created_at ASC
		LIMIT $3
	`, userID, threadID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *PostgresRepository) ListMemories(ctx context.Context, userID int64, limit int) ([]Memory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, key, value, confidence, source, created_at, updated_at
		FROM copilot_memories
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		memory, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	return memories, rows.Err()
}

func (r *PostgresRepository) UpsertMemory(ctx context.Context, memory Memory) (Memory, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO copilot_memories (user_id, key, value, confidence, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, key)
		DO UPDATE SET value = EXCLUDED.value, confidence = EXCLUDED.confidence, source = EXCLUDED.source, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`, memory.UserID, memory.Key, memory.Value, memory.Confidence, memory.Source, memory.CreatedAt).
		Scan(&memory.ID, &memory.CreatedAt, &memory.UpdatedAt)
	return memory, err
}

func (r *PostgresRepository) DeleteMemory(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM copilot_memories WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMemoryNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateFile(ctx context.Context, file File) (File, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO copilot_files (user_id, name, mime_type, size_bytes, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, created_at, updated_at
	`, file.UserID, file.Name, file.MimeType, file.SizeBytes, file.Content, file.CreatedAt).
		Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)
	return file, err
}

func (r *PostgresRepository) ListFiles(ctx context.Context, userID int64, limit int) ([]File, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, mime_type, size_bytes, content, created_at, updated_at
		FROM copilot_files
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (r *PostgresRepository) GetFilesByIDs(ctx context.Context, userID int64, ids []int64) ([]File, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, mime_type, size_bytes, content, created_at, updated_at
		FROM copilot_files
		WHERE user_id = $1 AND id = ANY($2)
		ORDER BY updated_at DESC
	`, userID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (r *PostgresRepository) ListRuns(ctx context.Context, userID int64, featurePrefix string, limit int) ([]ai.Run, error) {
	rows, err := r.db.Query(ctx, `
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
	`, userID, featurePrefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []ai.Run{}
	for rows.Next() {
		var run ai.Run
		if err := rows.Scan(
			&run.ID,
			&run.UserID,
			&run.Feature,
			&run.PromptVersion,
			&run.Provider,
			&run.Model,
			&run.Status,
			&run.ErrorCode,
			&run.ErrorMessage,
			&run.InputTokens,
			&run.OutputTokens,
			&run.LatencyMS,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanThread(scanner scanner) (Thread, error) {
	var thread Thread
	if err := scanner.Scan(
		&thread.ID,
		&thread.UserID,
		&thread.Title,
		&thread.Mode,
		&thread.Model,
		&thread.ArchivedAt,
		&thread.CreatedAt,
		&thread.UpdatedAt,
	); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func scanMessage(scanner scanner) (Message, error) {
	var message Message
	if err := scanner.Scan(
		&message.ID,
		&message.UserID,
		&message.ThreadID,
		&message.Role,
		&message.Content,
		&message.Status,
		&message.Model,
		&message.ErrorCode,
		&message.InputTokens,
		&message.OutputTokens,
		&message.Metadata,
		&message.CreatedAt,
	); err != nil {
		return Message{}, err
	}
	if !json.Valid(message.Metadata) {
		message.Metadata = []byte(`{}`)
	}
	return message, nil
}

func scanMemory(scanner scanner) (Memory, error) {
	var memory Memory
	if err := scanner.Scan(
		&memory.ID,
		&memory.UserID,
		&memory.Key,
		&memory.Value,
		&memory.Confidence,
		&memory.Source,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func scanFile(scanner scanner) (File, error) {
	var file File
	if err := scanner.Scan(
		&file.ID,
		&file.UserID,
		&file.Name,
		&file.MimeType,
		&file.SizeBytes,
		&file.Content,
		&file.CreatedAt,
		&file.UpdatedAt,
	); err != nil {
		return File{}, err
	}
	return file, nil
}
