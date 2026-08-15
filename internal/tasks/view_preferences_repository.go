package tasks

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) GetTaskViewPreference(ctx context.Context, userID int64, view string) (TaskViewPreference, error) {
	var preference TaskViewPreference
	var raw []byte
	err := r.db.QueryRow(ctx, `
		INSERT INTO task_view_preferences (user_id, view_type)
		VALUES ($1, $2)
		ON CONFLICT (user_id, view_type) DO NOTHING
		RETURNING view_type, columns, updated_at
	`, userID, view).Scan(&preference.View, &raw, &preference.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = r.db.QueryRow(ctx, `
			SELECT view_type, columns, updated_at
			FROM task_view_preferences
			WHERE user_id = $1 AND view_type = $2
		`, userID, view).Scan(&preference.View, &raw, &preference.UpdatedAt)
	}
	if err != nil {
		return TaskViewPreference{}, err
	}
	if err := json.Unmarshal(raw, &preference.Columns); err != nil {
		return TaskViewPreference{}, err
	}
	return preference, nil
}

func (r *PostgresRepository) SaveTaskViewPreference(ctx context.Context, userID int64, view string, columns []string) (TaskViewPreference, error) {
	raw, err := marshalTaskListColumns(columns)
	if err != nil {
		return TaskViewPreference{}, err
	}
	var preference TaskViewPreference
	var returned []byte
	err = r.db.QueryRow(ctx, `
		INSERT INTO task_view_preferences (user_id, view_type, columns, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, view_type) DO UPDATE SET
			columns = EXCLUDED.columns,
			updated_at = NOW()
		RETURNING view_type, columns, updated_at
	`, userID, view, raw).Scan(&preference.View, &returned, &preference.UpdatedAt)
	if err != nil {
		return TaskViewPreference{}, err
	}
	if err := json.Unmarshal(returned, &preference.Columns); err != nil {
		return TaskViewPreference{}, err
	}
	return preference, nil
}
