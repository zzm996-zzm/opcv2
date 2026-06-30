package learning

import (
	"errors"
	"time"
)

var (
	ErrServiceNotReady   = errors.New("learning service is not configured")
	ErrCourseNotFound    = errors.New("learning course not found")
	ErrDiagnosisNotFound = errors.New("learning diagnosis not found")
)

const DiagnosisCompleted = "completed"

type CourseFilter struct {
	Category string
	Limit    int
}

type Course struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Level       string    `json:"level"`
	Hours       int       `json:"hours"`
	Learners    int       `json:"learners"`
	PriceLabel  string    `json:"price_label"`
	Tags        []string  `json:"tags"`
	Outline     []string  `json:"outline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Progress struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	CourseSlug        string    `json:"course_slug"`
	CourseTitle       string    `json:"course_title"`
	Percent           int       `json:"percent"`
	LastLesson         string    `json:"last_lesson"`
	RecommendedAction string    `json:"recommended_action"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Dimension struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Gap     int    `json:"gap"`
	Summary string `json:"summary"`
}

type Diagnosis struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	Goal            string      `json:"goal"`
	Project         string      `json:"project"`
	Status          string      `json:"status"`
	OverallScore    int         `json:"overall_score"`
	Dimensions      []Dimension `json:"dimensions"`
	Recommendations []string    `json:"recommendations"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CreateDiagnosisInput struct {
	UserID  int64  `json:"-"`
	Goal    string `json:"goal"`
	Project string `json:"project"`
}
