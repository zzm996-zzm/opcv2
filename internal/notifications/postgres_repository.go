package notifications

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type postgresDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListNotifications(ctx context.Context, userID int64, filters ListFilters) ([]Notification, error) {
	const query = `
		SELECT id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		  AND ($2 = '' OR type = $2)
		  AND ($3 = 'all' OR ($3 = 'unread' AND read_at IS NULL) OR ($3 = 'read' AND read_at IS NOT NULL))
		ORDER BY created_at DESC
		LIMIT $4
	`
	rows, err := r.db.Query(ctx, query, userID, filters.Type, filters.Status, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (r *PostgresRepository) GetNotification(ctx context.Context, userID, id int64) (Notification, error) {
	const query = `
		SELECT id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
		FROM notifications
		WHERE user_id = $1 AND id = $2
	`
	row, err := scanNotification(r.db.QueryRow(ctx, query, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, ErrNotificationNotFound
	}
	return row, err
}

func (r *PostgresRepository) MarkRead(ctx context.Context, userID, id int64) (Notification, error) {
	const query = `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
	`
	row, err := scanNotification(r.db.QueryRow(ctx, query, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, ErrNotificationNotFound
	}
	return row, err
}

func (r *PostgresRepository) MarkAllRead(ctx context.Context, userID int64) (int, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE user_id = $1 AND read_at IS NULL
	`, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *PostgresRepository) Summary(ctx context.Context, userID int64) (Summary, error) {
	var summary Summary
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
	`, userID).Scan(&summary.Unread); err != nil {
		return Summary{}, err
	}
	typeRows, err := r.db.Query(ctx, `
		SELECT type, COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
		GROUP BY type
		ORDER BY type
	`, userID)
	if err != nil {
		return Summary{}, err
	}
	defer typeRows.Close()
	for typeRows.Next() {
		var item TypeCount
		if err := typeRows.Scan(&item.Type, &item.Count); err != nil {
			return Summary{}, err
		}
		summary.ByType = append(summary.ByType, item)
	}
	if err := typeRows.Err(); err != nil {
		return Summary{}, err
	}
	latestRows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return Summary{}, err
	}
	defer latestRows.Close()
	latest, err := scanNotifications(latestRows)
	if err != nil {
		return Summary{}, err
	}
	summary.Latest = httpapi.EnsureSlice(latest)
	summary.ByType = httpapi.EnsureSlice(summary.ByType)
	return summary, nil
}

type notificationRow interface {
	Scan(dest ...any) error
}

func scanNotification(row notificationRow) (Notification, error) {
	var notification Notification
	var sourceID sql.NullInt64
	var readAt sql.NullTime
	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Summary,
		&notification.Body,
		&notification.SourceType,
		&sourceID,
		&notification.ActionLabel,
		&notification.ActionURL,
		&readAt,
		&notification.CreatedAt,
	)
	if sourceID.Valid {
		notification.SourceID = &sourceID.Int64
	}
	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	return notification, err
}

func scanNotifications(rows pgx.Rows) ([]Notification, error) {
	var notifications []Notification
	for rows.Next() {
		row, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return httpapi.EnsureSlice(notifications), nil
}
