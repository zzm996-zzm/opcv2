package learning

import (
	"context"
	"encoding/json"
	"errors"

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

func (r *PostgresRepository) ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error) {
	query := `
		SELECT id, slug, title, description, category, level, hours, learners, price_label, tags, outline, created_at, updated_at
		FROM learning_courses
		WHERE ($1 = '' OR category = $1)
		ORDER BY learners DESC, id ASC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, filter.Category, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *PostgresRepository) GetCourse(ctx context.Context, slug string) (Course, error) {
	course, err := scanCourse(r.db.QueryRow(ctx, `
		SELECT id, slug, title, description, category, level, hours, learners, price_label, tags, outline, created_at, updated_at
		FROM learning_courses
		WHERE slug = $1
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Course{}, ErrCourseNotFound
	}
	return course, err
}

func (r *PostgresRepository) ListProgress(ctx context.Context, userID int64) ([]Progress, error) {
	rows, err := r.db.Query(ctx, `
		SELECT lp.id, lp.user_id, lp.course_slug, lc.title, lp.percent, lp.last_lesson, lp.recommended_action, lp.updated_at
		FROM learning_progress lp
		JOIN learning_courses lc ON lc.slug = lp.course_slug
		WHERE lp.user_id = $1
		ORDER BY lp.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []Progress
	for rows.Next() {
		var item Progress
		if err := rows.Scan(&item.ID, &item.UserID, &item.CourseSlug, &item.CourseTitle, &item.Percent, &item.LastLesson, &item.RecommendedAction, &item.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return progress, nil
}

func (r *PostgresRepository) CreateDiagnosis(ctx context.Context, diagnosis Diagnosis) (Diagnosis, error) {
	dimensions, err := json.Marshal(diagnosis.Dimensions)
	if err != nil {
		return Diagnosis{}, err
	}
	recommendations, err := json.Marshal(diagnosis.Recommendations)
	if err != nil {
		return Diagnosis{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO learning_diagnoses (user_id, goal, project, status, overall_score, dimensions, recommendations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, updated_at
	`, diagnosis.UserID, diagnosis.Goal, diagnosis.Project, diagnosis.Status, diagnosis.OverallScore, dimensions, recommendations, diagnosis.CreatedAt).Scan(&diagnosis.ID, &diagnosis.UpdatedAt)
	return diagnosis, err
}

func (r *PostgresRepository) LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error) {
	diagnosis, err := scanDiagnosis(r.db.QueryRow(ctx, `
		SELECT id, user_id, goal, project, status, overall_score, dimensions, recommendations, created_at, updated_at
		FROM learning_diagnoses
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Diagnosis{}, ErrDiagnosisNotFound
	}
	return diagnosis, err
}

type courseScanner interface {
	Scan(dest ...any) error
}

func scanCourse(scanner courseScanner) (Course, error) {
	var course Course
	var tags []byte
	var outline []byte
	if err := scanner.Scan(
		&course.ID,
		&course.Slug,
		&course.Title,
		&course.Description,
		&course.Category,
		&course.Level,
		&course.Hours,
		&course.Learners,
		&course.PriceLabel,
		&tags,
		&outline,
		&course.CreatedAt,
		&course.UpdatedAt,
	); err != nil {
		return Course{}, err
	}
	if err := json.Unmarshal(tags, &course.Tags); err != nil {
		return Course{}, err
	}
	if err := json.Unmarshal(outline, &course.Outline); err != nil {
		return Course{}, err
	}
	return course, nil
}

type diagnosisScanner interface {
	Scan(dest ...any) error
}

func scanDiagnosis(scanner diagnosisScanner) (Diagnosis, error) {
	var diagnosis Diagnosis
	var dimensions []byte
	var recommendations []byte
	if err := scanner.Scan(
		&diagnosis.ID,
		&diagnosis.UserID,
		&diagnosis.Goal,
		&diagnosis.Project,
		&diagnosis.Status,
		&diagnosis.OverallScore,
		&dimensions,
		&recommendations,
		&diagnosis.CreatedAt,
		&diagnosis.UpdatedAt,
	); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(dimensions, &diagnosis.Dimensions); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(recommendations, &diagnosis.Recommendations); err != nil {
		return Diagnosis{}, err
	}
	return diagnosis, nil
}
