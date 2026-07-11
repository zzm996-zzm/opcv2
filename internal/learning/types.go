package learning

import (
	"errors"
	"time"
)

var (
	ErrServiceNotReady   = errors.New("learning service is not configured")
	ErrCourseNotFound    = errors.New("learning course not found")
	ErrProgressNotFound  = errors.New("learning progress not found")
	ErrInvalidProgress   = errors.New("invalid learning progress")
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
	LastLesson        string    `json:"last_lesson"`
	RecommendedAction string    `json:"recommended_action"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UpdateProgressInput struct {
	UserID            int64  `json:"-"`
	CourseSlug        string `json:"-"`
	Percent           int    `json:"percent"`
	LastLesson        string `json:"last_lesson"`
	RecommendedAction string `json:"recommended_action"`
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
	FocusAbilities  []string    `json:"focus_abilities"`
	WeeklyTime      string      `json:"weekly_time"`
	Bottleneck      string      `json:"bottleneck"`
	Status          string      `json:"status"`
	OverallScore    int         `json:"overall_score"`
	Dimensions      []Dimension `json:"dimensions"`
	Recommendations []string    `json:"recommendations"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CreateDiagnosisInput struct {
	UserID         int64    `json:"-"`
	Goal           string   `json:"goal"`
	Project        string   `json:"project"`
	FocusAbilities []string `json:"focus_abilities"`
	WeeklyTime     string   `json:"weekly_time"`
	Bottleneck     string   `json:"bottleneck"`
}

type GapItem struct {
	Name        string `json:"name"`
	Current     int    `json:"current"`
	Target      int    `json:"target"`
	Gap         int    `json:"gap"`
	Priority    string `json:"priority"`
	Summary     string `json:"summary"`
	Evidence    string `json:"evidence"`
	Recommended string `json:"recommended"`
}

type DiagnosisGaps struct {
	DiagnosisID  int64     `json:"diagnosis_id"`
	Goal         string    `json:"goal"`
	Project      string    `json:"project"`
	OverallScore int       `json:"overall_score"`
	Gaps         []GapItem `json:"gaps"`
	Evidence     []string  `json:"evidence"`
	GeneratedAt  time.Time `json:"generated_at"`
}

type RecommendationFocus struct {
	Name     string `json:"name"`
	Priority string `json:"priority"`
	Summary  string `json:"summary"`
}

type LearningMethod struct {
	Title  string `json:"title"`
	Value  string `json:"value"`
	Detail string `json:"detail"`
}

type DiagnosisRecommendations struct {
	DiagnosisID     int64                 `json:"diagnosis_id"`
	Goal            string                `json:"goal"`
	Project         string                `json:"project"`
	Focus           []RecommendationFocus `json:"focus"`
	Recommendations []string              `json:"recommendations"`
	Methods         []LearningMethod      `json:"methods"`
	GeneratedAt     time.Time             `json:"generated_at"`
}

type PlanStage struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Courses   []string `json:"courses"`
	Duration  string   `json:"duration"`
	Goal      string   `json:"goal"`
	Milestone string   `json:"milestone"`
}

type DiagnosisPlan struct {
	DiagnosisID      int64       `json:"diagnosis_id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Recommendations  []string    `json:"recommendations"`
	Stages           []PlanStage `json:"stages"`
	EstimatedHours   int         `json:"estimated_hours"`
	WeeklySuggestion string      `json:"weekly_suggestion"`
	GeneratedAt      time.Time   `json:"generated_at"`
}

type DiagnosisReport struct {
	DiagnosisID     int64       `json:"diagnosis_id"`
	Goal            string      `json:"goal"`
	Project         string      `json:"project"`
	OverallScore    int         `json:"overall_score"`
	Dimensions      []Dimension `json:"dimensions"`
	PriorityGaps    []GapItem   `json:"priority_gaps"`
	Recommendations []string    `json:"recommendations"`
	Evidence        []string    `json:"evidence"`
	GeneratedAt     time.Time   `json:"generated_at"`
}
