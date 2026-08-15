package tasks

import (
	"context"
	"fmt"
)

// ListTaskCalendar uses the same ownership and text filters as the list view,
// but expands the date window so the calendar is not limited to the current
// list page.
func (r *PostgresRepository) ListTaskCalendar(ctx context.Context, userID int64, filters CalendarFilters) (CalendarPage, error) {
	const filter = `
		AND ($2 = '' OR status = $2)
		AND ($3 = '' OR project = $3)
		AND ($4 = '' OR priority = $4)
		AND ($5 = '' OR tags ? $5)
		AND ($6 = '' OR title ILIKE '%%' || $6 || '%%' OR description ILIKE '%%' || $6 || '%%' OR assignee ILIKE '%%' || $6 || '%%' OR project ILIKE '%%' || $6 || '%%' OR tags::TEXT ILIKE '%%' || $6 || '%%' OR learning ILIKE '%%' || $6 || '%%')
	`
	selectColumns := `id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, progress, completed_at, version, source_type, source_id, source_title, source_url, created_at, updated_at`
	args := []any{userID, filters.Status, filters.Project, filters.Priority, filters.Tag, filters.Query, filters.From, filters.To}

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND due_at >= $7 AND due_at < $8
		%s
		ORDER BY due_at ASC, id DESC
		LIMIT 500
	`, selectColumns, filter), args...)
	if err != nil {
		return CalendarPage{}, err
	}
	defer rows.Close()
	tasks := make([]Task, 0)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return CalendarPage{}, err
		}
		tasks = append(tasks, item)
	}
	if err := rows.Err(); err != nil {
		return CalendarPage{}, err
	}

	unscheduledRows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL AND due_at IS NULL
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT 500
	`, selectColumns, filter), args[:6]...)
	if err != nil {
		return CalendarPage{}, err
	}
	defer unscheduledRows.Close()
	unscheduled := make([]Task, 0)
	for unscheduledRows.Next() {
		item, err := scanTask(unscheduledRows)
		if err != nil {
			return CalendarPage{}, err
		}
		unscheduled = append(unscheduled, item)
	}
	if err := unscheduledRows.Err(); err != nil {
		return CalendarPage{}, err
	}

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND (due_at IS NULL OR (due_at >= $7 AND due_at < $8))
		%s
	`, filter), args...).Scan(&total); err != nil {
		return CalendarPage{}, err
	}
	return CalendarPage{Tasks: tasks, Unscheduled: unscheduled, Total: total, From: filters.From, To: filters.To}, nil
}
