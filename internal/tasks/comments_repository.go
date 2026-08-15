package tasks

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

type TaskCommentRepository interface {
	ListTaskComments(context.Context, int64, int64, int, int) ([]TaskComment, error)
	CountTaskComments(context.Context, int64, int64) (int, error)
	CreateTaskComment(context.Context, TaskComment) (TaskComment, error)
	UpdateTaskComment(context.Context, int64, int64, int64, string) (TaskComment, error)
	DeleteTaskComment(context.Context, int64, int64, int64) error
}

func (r *PostgresRepository) ListTaskComments(ctx context.Context, userID, taskID int64, limit, offset int) ([]TaskComment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.task_id, c.user_id, c.parent_comment_id, c.content, c.created_at, c.updated_at
		FROM task_comments c
		JOIN tasks t ON t.id = c.task_id AND t.user_id = $1 AND t.deleted_at IS NULL
		WHERE c.user_id = $1 AND c.task_id = $2 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC, c.id DESC
		LIMIT $3 OFFSET $4
	`, userID, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := make([]TaskComment, 0)
	for rows.Next() {
		comment, err := scanTaskComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *PostgresRepository) CountTaskComments(ctx context.Context, userID, taskID int64) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM task_comments c
		JOIN tasks t ON t.id = c.task_id AND t.user_id = $1 AND t.deleted_at IS NULL
		WHERE c.user_id = $1 AND c.task_id = $2 AND c.deleted_at IS NULL
	`, userID, taskID).Scan(&total)
	return total, err
}

func (r *PostgresRepository) CreateTaskComment(ctx context.Context, comment TaskComment) (TaskComment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TaskComment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var row taskScanner
	if comment.ParentCommentID == nil {
		row = tx.QueryRow(ctx, `
			INSERT INTO task_comments (task_id, user_id, parent_comment_id, content)
			SELECT $2, $1, NULL::BIGINT, $3::TEXT
			WHERE EXISTS (
				SELECT 1 FROM tasks t WHERE t.id = $2 AND t.user_id = $1 AND t.deleted_at IS NULL
			)
			RETURNING id, task_id, user_id, parent_comment_id, content, created_at, updated_at
		`, comment.UserID, comment.TaskID, strings.TrimSpace(comment.Content))
	} else {
		row = tx.QueryRow(ctx, `
			INSERT INTO task_comments (task_id, user_id, parent_comment_id, content)
			SELECT $2, $1, $3, $4
			WHERE EXISTS (
				SELECT 1 FROM tasks t WHERE t.id = $2 AND t.user_id = $1 AND t.deleted_at IS NULL
			)
			AND EXISTS (
				SELECT 1 FROM task_comments parent
				WHERE parent.id = $3 AND parent.task_id = $2 AND parent.user_id = $1 AND parent.deleted_at IS NULL
			)
			RETURNING id, task_id, user_id, parent_comment_id, content, created_at, updated_at
		`, comment.UserID, comment.TaskID, *comment.ParentCommentID, strings.TrimSpace(comment.Content))
	}
	created, err := scanTaskComment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskComment{}, ErrTaskCommentNotFound
	}
	if err != nil {
		return TaskComment{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO task_activities (task_id, user_id, action, after_data, metadata)
		VALUES ($1, $2, 'comment_added', jsonb_build_object('comment_id', $3::BIGINT, 'content_length', char_length($4::TEXT)), '{}'::JSONB)
	`, created.TaskID, created.UserID, created.ID, created.Content); err != nil {
		return TaskComment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TaskComment{}, err
	}
	return created, nil
}

func (r *PostgresRepository) UpdateTaskComment(ctx context.Context, userID, taskID, id int64, content string) (TaskComment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TaskComment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanTaskComment(tx.QueryRow(ctx, `
		UPDATE task_comments
		SET content = $4, updated_at = NOW()
		WHERE user_id = $1 AND task_id = $2 AND id = $3 AND deleted_at IS NULL
		  AND EXISTS (SELECT 1 FROM tasks t WHERE t.id = $2 AND t.user_id = $1 AND t.deleted_at IS NULL)
		RETURNING id, task_id, user_id, parent_comment_id, content, created_at, updated_at
	`, userID, taskID, id, strings.TrimSpace(content)))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskComment{}, ErrTaskCommentNotFound
	}
	if err != nil {
		return TaskComment{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO task_activities (task_id, user_id, action, after_data, metadata)
		VALUES ($1, $2, 'comment_updated', jsonb_build_object('comment_id', $3::BIGINT, 'content_length', char_length($4::TEXT)), '{}'::JSONB)
	`, updated.TaskID, updated.UserID, updated.ID, updated.Content); err != nil {
		return TaskComment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TaskComment{}, err
	}
	return updated, nil
}

func (r *PostgresRepository) DeleteTaskComment(ctx context.Context, userID, taskID, id int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var commentTaskID int64
	err = tx.QueryRow(ctx, `
		UPDATE task_comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND task_id = $2 AND id = $3 AND deleted_at IS NULL
		  AND EXISTS (SELECT 1 FROM tasks t WHERE t.id = $2 AND t.user_id = $1 AND t.deleted_at IS NULL)
		RETURNING task_id
	`, userID, taskID, id).Scan(&commentTaskID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTaskCommentNotFound
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO task_activities (task_id, user_id, action, after_data, metadata)
		VALUES ($1, $2, 'comment_deleted', jsonb_build_object('comment_id', $3::BIGINT), '{}'::JSONB)
	`, commentTaskID, userID, id)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	return err
}

func scanTaskComment(scanner taskScanner) (TaskComment, error) {
	var comment TaskComment
	err := scanner.Scan(
		&comment.ID,
		&comment.TaskID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	return comment, err
}
