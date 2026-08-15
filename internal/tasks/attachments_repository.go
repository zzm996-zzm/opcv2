package tasks

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) ListTaskAttachments(ctx context.Context, userID, taskID int64) ([]TaskAttachment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT attachment.id, attachment.task_id, attachment.comment_id, attachment.user_id,
		       attachment.name, attachment.mime_type, attachment.size_bytes, attachment.storage_key,
		       attachment.sha256, attachment.created_at, attachment.deleted_at
		FROM task_attachments AS attachment
		JOIN tasks AS task ON task.id = attachment.task_id
		WHERE task.user_id = $1 AND task.id = $2 AND task.deleted_at IS NULL
		  AND attachment.user_id = $1 AND attachment.deleted_at IS NULL
		ORDER BY attachment.created_at DESC, attachment.id DESC
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TaskAttachment, 0)
	for rows.Next() {
		item, err := scanTaskAttachment(rows)
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

func (r *PostgresRepository) CreateTaskAttachment(ctx context.Context, attachment TaskAttachment) (TaskAttachment, error) {
	created, err := scanTaskAttachment(r.db.QueryRow(ctx, `
		INSERT INTO task_attachments (
			task_id, comment_id, user_id, name, mime_type, size_bytes, storage_key, sha256, created_at
		)
		SELECT task.id, $3, task.user_id, $4, $5, $6, $7, $8, $9
		FROM tasks AS task
		WHERE task.id = $1 AND task.user_id = $2 AND task.deleted_at IS NULL
		  AND ($3::BIGINT IS NULL OR EXISTS (
			SELECT 1 FROM task_comments AS comment
			WHERE comment.id = $3 AND comment.task_id = task.id
			  AND comment.user_id = task.user_id AND comment.deleted_at IS NULL
		  ))
		RETURNING id, task_id, comment_id, user_id, name, mime_type, size_bytes,
		          storage_key, sha256, created_at, deleted_at
	`, attachment.TaskID, attachment.UserID, optionalInt64(attachment.CommentID), attachment.Name,
		attachment.MIMEType, attachment.SizeBytes, attachment.StorageKey, attachment.SHA256, attachment.CreatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskAttachment{}, ErrTaskNotFound
	}
	return created, err
}

func (r *PostgresRepository) GetTaskAttachment(ctx context.Context, userID, taskID, attachmentID int64) (TaskAttachment, error) {
	attachment, err := scanTaskAttachment(r.db.QueryRow(ctx, `
		SELECT attachment.id, attachment.task_id, attachment.comment_id, attachment.user_id,
		       attachment.name, attachment.mime_type, attachment.size_bytes, attachment.storage_key,
		       attachment.sha256, attachment.created_at, attachment.deleted_at
		FROM task_attachments AS attachment
		JOIN tasks AS task ON task.id = attachment.task_id
		WHERE task.user_id = $1 AND task.id = $2 AND task.deleted_at IS NULL
		  AND attachment.id = $3 AND attachment.user_id = $1 AND attachment.deleted_at IS NULL
	`, userID, taskID, attachmentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskAttachment{}, ErrTaskAttachmentNotFound
	}
	return attachment, err
}

func (r *PostgresRepository) DeleteTaskAttachment(ctx context.Context, userID, taskID, attachmentID int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE task_attachments AS attachment
		SET deleted_at = NOW()
		WHERE attachment.id = $3 AND attachment.task_id = $2 AND attachment.user_id = $1
		  AND attachment.deleted_at IS NULL
		  AND EXISTS (
			SELECT 1 FROM tasks AS task
			WHERE task.id = $2 AND task.user_id = $1 AND task.deleted_at IS NULL
		  )
	`, userID, taskID, attachmentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskAttachmentNotFound
	}
	return nil
}

func scanTaskAttachment(scanner taskScanner) (TaskAttachment, error) {
	var attachment TaskAttachment
	var commentID sql.NullInt64
	var deletedAt sql.NullTime
	err := scanner.Scan(
		&attachment.ID,
		&attachment.TaskID,
		&commentID,
		&attachment.UserID,
		&attachment.Name,
		&attachment.MIMEType,
		&attachment.SizeBytes,
		&attachment.StorageKey,
		&attachment.SHA256,
		&attachment.CreatedAt,
		&deletedAt,
	)
	if commentID.Valid {
		attachment.CommentID = &commentID.Int64
	}
	if deletedAt.Valid {
		attachment.DeletedAt = &deletedAt.Time
	}
	return attachment, err
}
