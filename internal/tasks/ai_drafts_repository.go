package tasks

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) CreateTaskAIDraft(ctx context.Context, draft TaskAIDraft) (TaskAIDraft, error) {
	if draft.Tasks == nil {
		draft.Tasks = []Task{}
	}
	if draft.AdoptedTaskIDs == nil {
		draft.AdoptedTaskIDs = []int64{}
	}
	tasks, err := json.Marshal(draft.Tasks)
	if err != nil {
		return TaskAIDraft{}, err
	}
	adoptedTaskIDs, err := json.Marshal(draft.AdoptedTaskIDs)
	if err != nil {
		return TaskAIDraft{}, err
	}
	if draft.Status == "" {
		draft.Status = TaskAIDraftStatusDraft
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO task_ai_drafts (user_id, goal, source_type, source_id, source_title, source_url, tasks, status, adopted_task_ids, created_at, updated_at, adopted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11)
		RETURNING id, created_at, updated_at, adopted_at
	`, draft.UserID, draft.Goal, draft.SourceType, draft.SourceID, draft.SourceTitle, draft.SourceURL, tasks, draft.Status, adoptedTaskIDs, draft.CreatedAt, draft.AdoptedAt).
		Scan(&draft.ID, &draft.CreatedAt, &draft.UpdatedAt, &draft.AdoptedAt)
	return draft, err
}

func (r *PostgresRepository) GetTaskAIDraft(ctx context.Context, userID, id int64) (TaskAIDraft, error) {
	draft, err := scanTaskAIDraft(r.db.QueryRow(ctx, `
		SELECT id, user_id, goal, source_type, source_id, source_title, source_url, tasks, status, adopted_task_ids, created_at, updated_at, adopted_at
		FROM task_ai_drafts
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskAIDraft{}, ErrTaskAIDraftNotFound
	}
	return draft, err
}

func (r *PostgresRepository) AdoptTaskAIDraft(ctx context.Context, userID, id int64, tasks []Task) ([]Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	draft, err := scanTaskAIDraft(tx.QueryRow(ctx, `
		SELECT id, user_id, goal, source_type, source_id, source_title, source_url, tasks, status, adopted_task_ids, created_at, updated_at, adopted_at
		FROM task_ai_drafts
		WHERE user_id = $1 AND id = $2
		FOR UPDATE
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskAIDraftNotFound
	}
	if err != nil {
		return nil, err
	}
	if draft.Status == TaskAIDraftStatusAdopted {
		return loadAdoptedTasks(ctx, tx, userID, draft.AdoptedTaskIDs)
	}

	created := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		item, err := createTask(ctx, tx, task)
		if err != nil {
			return nil, err
		}
		created = append(created, item)
	}
	adoptedTaskIDs := make([]int64, 0, len(created))
	for _, task := range created {
		adoptedTaskIDs = append(adoptedTaskIDs, task.ID)
	}
	payload, err := json.Marshal(adoptedTaskIDs)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE task_ai_drafts
		SET status = 'adopted', adopted_task_ids = $3, adopted_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND id = $2
	`, userID, id, payload); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

func loadAdoptedTasks(ctx context.Context, tx pgx.Tx, userID int64, ids []int64) ([]Task, error) {
	if len(ids) == 0 {
		return nil, ErrTaskAIDraftAlreadyAdopted
	}
	rows, err := tx.Query(ctx, `
		SELECT id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, progress, completed_at, version, source_type, source_id, source_title, source_url, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND id = ANY($2::bigint[])
		ORDER BY array_position($2::bigint[], id)
	`, userID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	loaded := make([]Task, 0, len(ids))
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(loaded) != len(ids) {
		return nil, ErrTaskAIDraftAlreadyAdopted
	}
	return loaded, nil
}

func scanTaskAIDraft(scanner taskScanner) (TaskAIDraft, error) {
	var draft TaskAIDraft
	var tasks []byte
	var adoptedTaskIDs []byte
	if err := scanner.Scan(
		&draft.ID,
		&draft.UserID,
		&draft.Goal,
		&draft.SourceType,
		&draft.SourceID,
		&draft.SourceTitle,
		&draft.SourceURL,
		&tasks,
		&draft.Status,
		&adoptedTaskIDs,
		&draft.CreatedAt,
		&draft.UpdatedAt,
		&draft.AdoptedAt,
	); err != nil {
		return TaskAIDraft{}, err
	}
	if err := json.Unmarshal(tasks, &draft.Tasks); err != nil {
		return TaskAIDraft{}, err
	}
	if err := json.Unmarshal(adoptedTaskIDs, &draft.AdoptedTaskIDs); err != nil {
		return TaskAIDraft{}, err
	}
	return draft, nil
}
