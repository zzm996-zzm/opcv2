package tasks

import (
	"context"
	"encoding/json"
)

func (r *PostgresRepository) ListTaskActivities(ctx context.Context, userID, taskID int64, limit, offset int) ([]TaskActivity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT activity.id, activity.task_id, activity.user_id, activity.action,
		       activity.before_data, activity.after_data, activity.metadata, activity.created_at
		FROM task_activities AS activity
		JOIN tasks AS task ON task.id = activity.task_id AND task.user_id = $1 AND task.deleted_at IS NULL
		WHERE activity.task_id = $2
		ORDER BY activity.created_at DESC, activity.id DESC
		LIMIT $3 OFFSET $4
	`, userID, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activities := make([]TaskActivity, 0)
	for rows.Next() {
		activity, err := scanTaskActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *PostgresRepository) CountTaskActivities(ctx context.Context, userID, taskID int64) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM task_activities AS activity
		JOIN tasks AS task ON task.id = activity.task_id AND task.user_id = $1 AND task.deleted_at IS NULL
		WHERE activity.task_id = $2
	`, userID, taskID).Scan(&total)
	return total, err
}

func scanTaskActivity(scanner taskScanner) (TaskActivity, error) {
	var activity TaskActivity
	var beforeData []byte
	var afterData []byte
	var metadata []byte
	if err := scanner.Scan(
		&activity.ID,
		&activity.TaskID,
		&activity.UserID,
		&activity.Action,
		&beforeData,
		&afterData,
		&metadata,
		&activity.CreatedAt,
	); err != nil {
		return TaskActivity{}, err
	}
	if err := json.Unmarshal(beforeData, &activity.BeforeData); err != nil {
		return TaskActivity{}, err
	}
	if err := json.Unmarshal(afterData, &activity.AfterData); err != nil {
		return TaskActivity{}, err
	}
	if err := json.Unmarshal(metadata, &activity.Metadata); err != nil {
		return TaskActivity{}, err
	}
	return activity, nil
}
