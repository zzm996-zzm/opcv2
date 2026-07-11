package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, task Task) (Task, error) {
	return createTask(ctx, r.db, task)
}

type taskWriter interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func createTask(ctx context.Context, writer taskWriter, task Task) (Task, error) {
	tags, err := json.Marshal(task.Tags)
	if err != nil {
		return Task{}, err
	}
	tools, err := json.Marshal(task.Tools)
	if err != nil {
		return Task{}, err
	}
	err = writer.QueryRow(ctx, `
		INSERT INTO tasks (user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16)
		RETURNING id, created_at, updated_at
	`,
		task.UserID,
		task.Title,
		task.Description,
		task.Assignee,
		task.Project,
		task.Status,
		task.Priority,
		tags,
		task.DueAt,
		tools,
		task.Learning,
		task.SourceType,
		task.SourceID,
		task.SourceTitle,
		task.SourceURL,
		task.CreatedAt,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	return task, err
}

func (r *PostgresRepository) CreateTasks(ctx context.Context, tasks []Task) ([]Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	created := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		item, err := createTask(ctx, tx, task)
		if err != nil {
			return nil, err
		}
		created = append(created, item)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *PostgresRepository) ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR project = $3)
		  AND ($4 = '' OR priority = $4)
		  AND ($5 = '' OR tags ? $5)
		  AND ($6 = '' OR title ILIKE '%' || $6 || '%' OR description ILIKE '%' || $6 || '%' OR assignee ILIKE '%' || $6 || '%' OR project ILIKE '%' || $6 || '%' OR tags::TEXT ILIKE '%' || $6 || '%' OR learning ILIKE '%' || $6 || '%')
		ORDER BY created_at DESC
		LIMIT $7
		OFFSET $8
	`, userID, filters.Status, filters.Project, filters.Priority, filters.Tag, filters.Query, filters.Limit, filters.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *PostgresRepository) CountTasks(ctx context.Context, userID int64, filters ListFilters) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR project = $3)
		  AND ($4 = '' OR priority = $4)
		  AND ($5 = '' OR tags ? $5)
		  AND ($6 = '' OR title ILIKE '%' || $6 || '%' OR description ILIKE '%' || $6 || '%' OR assignee ILIKE '%' || $6 || '%' OR project ILIKE '%' || $6 || '%' OR tags::TEXT ILIKE '%' || $6 || '%' OR learning ILIKE '%' || $6 || '%')
	`, userID, filters.Status, filters.Project, filters.Priority, filters.Tag, filters.Query).Scan(&total)
	return total, err
}

func (r *PostgresRepository) ListTaskProjects(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT project
		FROM tasks
		WHERE user_id = $1 AND project <> ''
		ORDER BY project
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []string
	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *PostgresRepository) ListTaskTags(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT tag.value
		FROM tasks
		CROSS JOIN LATERAL jsonb_array_elements_text(tags) AS tag(value)
		WHERE user_id = $1
		ORDER BY tag.value
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *PostgresRepository) TaskStats(ctx context.Context, userID int64, now time.Time) (Stats, error) {
	var stats Stats
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)::INT,
			COUNT(*) FILTER (WHERE status = 'todo')::INT,
			COUNT(*) FILTER (WHERE status = 'in_progress')::INT,
			COUNT(*) FILTER (WHERE status = 'completed')::INT,
			COUNT(*) FILTER (WHERE status = 'reminder')::INT,
			COUNT(*) FILTER (WHERE due_at IS NOT NULL AND due_at < $2 AND status <> 'completed')::INT
		FROM tasks
		WHERE user_id = $1
	`, userID, now).Scan(
		&stats.Total,
		&stats.Todo,
		&stats.InProgress,
		&stats.Completed,
		&stats.Reminder,
		&stats.Overdue,
	)
	return stats, err
}

func (r *PostgresRepository) GetTask(ctx context.Context, userID, id int64) (Task, error) {
	task, err := scanTask(r.db.QueryRow(ctx, `
		SELECT id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	return task, err
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error) {
	var tags any
	if update.Tags != nil {
		payload, err := json.Marshal(*update.Tags)
		if err != nil {
			return Task{}, err
		}
		tags = payload
	}
	var tools any
	if update.Tools != nil {
		payload, err := json.Marshal(*update.Tools)
		if err != nil {
			return Task{}, err
		}
		tools = payload
	}
	task, err := scanTask(r.db.QueryRow(ctx, `
		UPDATE tasks
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    assignee = COALESCE($3, assignee),
		    project = COALESCE($4, project),
		    status = COALESCE($5, status),
		    priority = COALESCE($6, priority),
		    tags = COALESCE($7, tags),
		    due_at = CASE WHEN $9 THEN NULL ELSE COALESCE($8, due_at) END,
		    tools = COALESCE($10, tools),
		    learning = COALESCE($11, learning),
		    updated_at = NOW()
		WHERE user_id = $12 AND id = $13
		RETURNING id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
	`,
		optionalString(update.Title),
		optionalString(update.Description),
		optionalString(update.Assignee),
		optionalString(update.Project),
		optionalString(update.Status),
		optionalString(update.Priority),
		tags,
		optionalTime(update.DueAt),
		update.ClearDueAt,
		tools,
		optionalString(update.Learning),
		userID,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	return task, err
}

func (r *PostgresRepository) DeleteTask(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM tasks
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return value
}

func optionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (Task, error) {
	var task Task
	var tags []byte
	var tools []byte
	if err := scanner.Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Assignee,
		&task.Project,
		&task.Status,
		&task.Priority,
		&tags,
		&task.DueAt,
		&tools,
		&task.Learning,
		&task.SourceType,
		&task.SourceID,
		&task.SourceTitle,
		&task.SourceURL,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal(tags, &task.Tags); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal(tools, &task.Tools); err != nil {
		return Task{}, err
	}
	return task, nil
}
