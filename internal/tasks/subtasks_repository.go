package tasks

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) ListSubtasks(ctx context.Context, userID, taskID int64) ([]Subtask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.task_id, s.user_id, s.parent_subtask_id, s.title, s.assignee, s.due_at, s.completed, s.created_at, s.updated_at
		FROM task_subtasks s
		JOIN tasks t ON t.id = s.task_id AND t.user_id = $1 AND t.deleted_at IS NULL
		WHERE s.user_id = $1 AND s.task_id = $2
		ORDER BY s.created_at, s.id
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Subtask, 0)
	for rows.Next() {
		item, err := scanSubtask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) CreateSubtask(ctx context.Context, item Subtask) (Subtask, error) {
	created, err := scanSubtask(r.db.QueryRow(ctx, `
		INSERT INTO task_subtasks (task_id, user_id, parent_subtask_id, title, assignee, due_at, created_at, updated_at)
		SELECT t.id, t.user_id, $6, $3, $4, $5, $7, $7
		FROM tasks t
		WHERE t.id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL
		  AND ($6::BIGINT IS NULL OR EXISTS (
		      SELECT 1
		      FROM task_subtasks parent
		      WHERE parent.id = $6 AND parent.task_id = t.id AND parent.user_id = t.user_id
		  ))
		RETURNING id, task_id, user_id, parent_subtask_id, title, assignee, due_at, completed, created_at, updated_at
	`, item.TaskID, item.UserID, item.Title, item.Assignee, optionalTime(item.DueAt), optionalInt64(item.ParentSubtaskID), item.CreatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		if item.ParentSubtaskID != nil {
			return Subtask{}, ErrSubtaskNotFound
		}
		return Subtask{}, ErrTaskNotFound
	}
	return created, err
}

func (r *PostgresRepository) UpdateSubtask(ctx context.Context, userID, taskID, id int64, update SubtaskUpdate) (Subtask, error) {
	item, err := scanSubtask(r.db.QueryRow(ctx, `
		UPDATE task_subtasks
		SET title = COALESCE($1, title),
		    assignee = COALESCE($2, assignee),
		    due_at = CASE WHEN $4 THEN NULL ELSE COALESCE($3, due_at) END,
		    completed = COALESCE($5, completed),
		    updated_at = NOW()
		WHERE user_id = $6 AND task_id = $7 AND id = $8
		  AND EXISTS (
		      SELECT 1 FROM tasks t
		      WHERE t.id = $7 AND t.user_id = $6 AND t.deleted_at IS NULL
		  )
		RETURNING id, task_id, user_id, parent_subtask_id, title, assignee, due_at, completed, created_at, updated_at
	`, optionalString(update.Title), optionalString(update.Assignee), optionalTime(update.DueAt), update.ClearDueAt, update.Completed, userID, taskID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Subtask{}, ErrSubtaskNotFound
	}
	return item, err
}

func (r *PostgresRepository) DeleteSubtask(ctx context.Context, userID, taskID, id int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM task_subtasks
		WHERE user_id = $1 AND task_id = $2 AND id = $3
		  AND EXISTS (
		      SELECT 1 FROM tasks t
		      WHERE t.id = $2 AND t.user_id = $1 AND t.deleted_at IS NULL
		  )
	`, userID, taskID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSubtaskNotFound
	}
	return nil
}

func scanSubtask(scanner taskScanner) (Subtask, error) {
	var item Subtask
	var parentSubtaskID sql.NullInt64
	err := scanner.Scan(
		&item.ID,
		&item.TaskID,
		&item.UserID,
		&parentSubtaskID,
		&item.Title,
		&item.Assignee,
		&item.DueAt,
		&item.Completed,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if parentSubtaskID.Valid {
		item.ParentSubtaskID = &parentSubtaskID.Int64
	}
	return item, err
}

func optionalInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
