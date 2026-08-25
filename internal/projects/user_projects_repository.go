package projects

import "context"

func (r *PostgresRepository) CreateUserProject(ctx context.Context, project UserProject) (UserProject, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_projects
			(user_id, name, description, status, source_type, source_id, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (user_id, idempotency_key) WHERE idempotency_key <> ''
		DO UPDATE SET updated_at = user_projects.updated_at
		RETURNING id, user_id, name, description, status, source_type, source_id, idempotency_key, created_at, updated_at
	`, project.UserID, project.Name, project.Description, project.Status, project.SourceType, project.SourceID, project.IdempotencyKey, project.CreatedAt).Scan(
		&project.ID, &project.UserID, &project.Name, &project.Description, &project.Status, &project.SourceType,
		&project.SourceID, &project.IdempotencyKey, &project.CreatedAt, &project.UpdatedAt,
	)
	return project, err
}

func (r *PostgresRepository) ListUserProjects(ctx context.Context, userID, limit int) ([]UserProject, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, description, status, source_type, source_id, idempotency_key, created_at, updated_at
		FROM user_projects WHERE user_id = $1 AND status <> 'archived'
		ORDER BY created_at DESC, id DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]UserProject, 0)
	for rows.Next() {
		var item UserProject
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.Status, &item.SourceType, &item.SourceID, &item.IdempotencyKey, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
