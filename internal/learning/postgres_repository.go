package learning

import (
	"context"
	"database/sql"
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

func (r *PostgresRepository) ListCourseMaterials(ctx context.Context, courseSlug string) ([]CourseMaterial, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, course_slug, title, material_type, content_url, position, downloadable, created_at, updated_at
		FROM learning_course_materials
		WHERE course_slug = $1
		ORDER BY position ASC, id ASC
	`, courseSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var materials []CourseMaterial
	for rows.Next() {
		var material CourseMaterial
		if err := rows.Scan(&material.ID, &material.CourseSlug, &material.Title, &material.MaterialType, &material.ContentURL, &material.Position, &material.Downloadable, &material.CreatedAt, &material.UpdatedAt); err != nil {
			return nil, err
		}
		materials = append(materials, material)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return materials, nil
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

func (r *PostgresRepository) GetProgress(ctx context.Context, userID int64, courseSlug string) (Progress, error) {
	progress, err := scanProgress(r.db.QueryRow(ctx, `
		SELECT lp.id, lp.user_id, lp.course_slug, lc.title, lp.percent, lp.last_lesson, lp.recommended_action, lp.updated_at
		FROM learning_progress lp
		JOIN learning_courses lc ON lc.slug = lp.course_slug
		WHERE lp.user_id = $1 AND lp.course_slug = $2
	`, userID, courseSlug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Progress{}, ErrProgressNotFound
	}
	return progress, err
}

func (r *PostgresRepository) UpsertProgress(ctx context.Context, progress Progress) (Progress, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO learning_progress (user_id, course_slug, percent, last_lesson, recommended_action, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, course_slug) DO UPDATE SET
			percent = EXCLUDED.percent,
			last_lesson = EXCLUDED.last_lesson,
			recommended_action = EXCLUDED.recommended_action,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, course_slug, percent, last_lesson, recommended_action, updated_at
	`, progress.UserID, progress.CourseSlug, progress.Percent, progress.LastLesson, progress.RecommendedAction, progress.UpdatedAt).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseSlug,
		&progress.Percent,
		&progress.LastLesson,
		&progress.RecommendedAction,
		&progress.UpdatedAt,
	)
	return progress, err
}

func scanProgress(scanner courseScanner) (Progress, error) {
	var progress Progress
	err := scanner.Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseSlug,
		&progress.CourseTitle,
		&progress.Percent,
		&progress.LastLesson,
		&progress.RecommendedAction,
		&progress.UpdatedAt,
	)
	return progress, err
}

func (r *PostgresRepository) CreateDiagnosis(ctx context.Context, diagnosis Diagnosis) (Diagnosis, error) {
	focusAbilities, err := json.Marshal(diagnosis.FocusAbilities)
	if err != nil {
		return Diagnosis{}, err
	}
	dimensions, err := json.Marshal(diagnosis.Dimensions)
	if err != nil {
		return Diagnosis{}, err
	}
	recommendations, err := json.Marshal(diagnosis.Recommendations)
	if err != nil {
		return Diagnosis{}, err
	}
	answers, err := json.Marshal(diagnosis.Answers)
	if err != nil {
		return Diagnosis{}, err
	}
	assumptions, err := json.Marshal(diagnosis.Assumptions)
	if err != nil {
		return Diagnosis{}, err
	}
	evidenceSources, err := json.Marshal(diagnosis.EvidenceSources)
	if err != nil {
		return Diagnosis{}, err
	}
	gapsSnapshot, err := json.Marshal(diagnosis.GapsSnapshot)
	if err != nil {
		return Diagnosis{}, err
	}
	recommendationsSnapshot, err := json.Marshal(diagnosis.RecommendationsSnapshot)
	if err != nil {
		return Diagnosis{}, err
	}
	planSnapshot, err := json.Marshal(diagnosis.PlanSnapshot)
	if err != nil {
		return Diagnosis{}, err
	}
	reportSnapshot, err := json.Marshal(diagnosis.ReportSnapshot)
	if err != nil {
		return Diagnosis{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO learning_diagnoses (
			user_id, goal, project, focus_abilities, weekly_time, bottleneck, answers,
			status, overall_score, dimensions, recommendations, basis, disclaimer,
			assumptions, evidence_sources, gaps_snapshot, recommendations_snapshot,
			plan_snapshot, report_snapshot, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $20)
		RETURNING id, updated_at
	`, diagnosis.UserID, diagnosis.Goal, diagnosis.Project, focusAbilities, diagnosis.WeeklyTime, diagnosis.Bottleneck, answers,
		diagnosis.Status, diagnosis.OverallScore, dimensions, recommendations, diagnosis.Basis, diagnosis.Disclaimer,
		assumptions, evidenceSources, gapsSnapshot, recommendationsSnapshot, planSnapshot, reportSnapshot, diagnosis.CreatedAt,
	).Scan(&diagnosis.ID, &diagnosis.UpdatedAt)
	return diagnosis, err
}

func (r *PostgresRepository) GetDiagnosis(ctx context.Context, userID, id int64) (Diagnosis, error) {
	diagnosis, err := scanDiagnosis(r.db.QueryRow(ctx, `
		SELECT id, user_id, goal, project, focus_abilities, weekly_time, bottleneck, answers,
			status, overall_score, dimensions, recommendations, basis, disclaimer, assumptions,
			evidence_sources, gaps_snapshot, recommendations_snapshot, plan_snapshot, report_snapshot,
			created_at, updated_at
		FROM learning_diagnoses
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Diagnosis{}, ErrDiagnosisNotFound
	}
	return diagnosis, err
}

func (r *PostgresRepository) LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error) {
	diagnosis, err := scanDiagnosis(r.db.QueryRow(ctx, `
		SELECT id, user_id, goal, project, focus_abilities, weekly_time, bottleneck, answers,
			status, overall_score, dimensions, recommendations, basis, disclaimer, assumptions,
			evidence_sources, gaps_snapshot, recommendations_snapshot, plan_snapshot, report_snapshot,
			created_at, updated_at
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

func (r *PostgresRepository) ListPlanItems(ctx context.Context, userID, diagnosisID int64) ([]PlanItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, diagnosis_id, stage_number, title, completed, completed_at, updated_at
		FROM learning_plan_items
		WHERE user_id = $1 AND diagnosis_id = $2
		ORDER BY stage_number ASC
	`, userID, diagnosisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PlanItem
	for rows.Next() {
		item, err := scanPlanItem(rows)
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

func (r *PostgresRepository) UpsertPlanItem(ctx context.Context, item PlanItem) (PlanItem, error) {
	updated, err := scanPlanItem(r.db.QueryRow(ctx, `
		INSERT INTO learning_plan_items (user_id, diagnosis_id, stage_number, title, completed, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id, diagnosis_id, stage_number) DO UPDATE SET
			title = EXCLUDED.title,
			completed = EXCLUDED.completed,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, diagnosis_id, stage_number, title, completed, completed_at, updated_at
	`, item.UserID, item.DiagnosisID, item.StageNumber, item.Title, item.Completed, item.CompletedAt, item.UpdatedAt))
	return updated, err
}

func scanPlanItem(scanner courseScanner) (PlanItem, error) {
	var item PlanItem
	var completedAt sql.NullTime
	err := scanner.Scan(&item.ID, &item.UserID, &item.DiagnosisID, &item.StageNumber, &item.Title, &item.Completed, &completedAt, &item.UpdatedAt)
	if err != nil {
		return PlanItem{}, err
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	return item, nil
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
	var focusAbilities []byte
	var answers []byte
	var dimensions []byte
	var recommendations []byte
	var assumptions []byte
	var evidenceSources []byte
	var gapsSnapshot []byte
	var recommendationsSnapshot []byte
	var planSnapshot []byte
	var reportSnapshot []byte
	if err := scanner.Scan(
		&diagnosis.ID,
		&diagnosis.UserID,
		&diagnosis.Goal,
		&diagnosis.Project,
		&focusAbilities,
		&diagnosis.WeeklyTime,
		&diagnosis.Bottleneck,
		&answers,
		&diagnosis.Status,
		&diagnosis.OverallScore,
		&dimensions,
		&recommendations,
		&diagnosis.Basis,
		&diagnosis.Disclaimer,
		&assumptions,
		&evidenceSources,
		&gapsSnapshot,
		&recommendationsSnapshot,
		&planSnapshot,
		&reportSnapshot,
		&diagnosis.CreatedAt,
		&diagnosis.UpdatedAt,
	); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(focusAbilities, &diagnosis.FocusAbilities); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(answers, &diagnosis.Answers); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(dimensions, &diagnosis.Dimensions); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(recommendations, &diagnosis.Recommendations); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(assumptions, &diagnosis.Assumptions); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(evidenceSources, &diagnosis.EvidenceSources); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(gapsSnapshot, &diagnosis.GapsSnapshot); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(recommendationsSnapshot, &diagnosis.RecommendationsSnapshot); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(planSnapshot, &diagnosis.PlanSnapshot); err != nil {
		return Diagnosis{}, err
	}
	if err := json.Unmarshal(reportSnapshot, &diagnosis.ReportSnapshot); err != nil {
		return Diagnosis{}, err
	}
	return diagnosis, nil
}
