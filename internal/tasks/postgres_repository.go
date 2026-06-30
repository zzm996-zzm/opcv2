package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, task Task) (Task, error) {
	tools, err := json.Marshal(task.Tools)
	if err != nil {
		return Task{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO tasks (user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
	`,
		task.UserID,
		task.Title,
		task.Project,
		task.Status,
		task.Priority,
		task.DueAt,
		tools,
		task.Learning,
		task.CreatedAt,
	).Scan(&task.ID)
	return task, err
}

func (r *PostgresRepository) ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
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

func (r *PostgresRepository) GetTask(ctx context.Context, userID, id int64) (Task, error) {
	task, err := scanTask(r.db.QueryRow(ctx, `
		SELECT id, user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	return task, err
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error) {
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
		    project = COALESCE($2, project),
		    status = COALESCE($3, status),
		    priority = COALESCE($4, priority),
		    due_at = COALESCE($5, due_at),
		    tools = COALESCE($6, tools),
		    learning = COALESCE($7, learning),
		    updated_at = NOW()
		WHERE user_id = $8 AND id = $9
		RETURNING id, user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at
	`,
		optionalString(update.Title),
		optionalString(update.Project),
		optionalString(update.Status),
		optionalString(update.Priority),
		optionalTime(update.DueAt),
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
	var tools []byte
	if err := scanner.Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Project,
		&task.Status,
		&task.Priority,
		&task.DueAt,
		&tools,
		&task.Learning,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return Task{}, err
	}
	if err := json.Unmarshal(tools, &task.Tools); err != nil {
		return Task{}, err
	}
	return task, nil
}
