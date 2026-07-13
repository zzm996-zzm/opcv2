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
	ErrInvalidDiagnosis  = errors.New("invalid learning diagnosis")
	ErrInvalidAIResult   = errors.New("invalid learning ai result")
	ErrDiagnosisNotFound = errors.New("learning diagnosis not found")
	ErrInvalidPlanItem   = errors.New("invalid learning plan item")
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

type CourseMaterial struct {
	ID           int64     `json:"id"`
	CourseSlug   string    `json:"course_slug"`
	Title        string    `json:"title"`
	MaterialType string    `json:"material_type"`
	ContentURL   string    `json:"content_url"`
	Position     int       `json:"position"`
	Downloadable bool      `json:"downloadable"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

type AssessmentAnswer struct {
	Key      string `json:"key"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type DiagnosisEvidenceSource struct {
	Type       string    `json:"type"`
	Label      string    `json:"label"`
	CapturedAt time.Time `json:"captured_at"`
}

type Diagnosis struct {
	ID                      int64                     `json:"id"`
	UserID                  int64                     `json:"user_id"`
	Goal                    string                    `json:"goal"`
	Project                 string                    `json:"project"`
	FocusAbilities          []string                  `json:"focus_abilities"`
	WeeklyTime              string                    `json:"weekly_time"`
	Bottleneck              string                    `json:"bottleneck"`
	Answers                 []AssessmentAnswer        `json:"answers"`
	Status                  string                    `json:"status"`
	OverallScore            int                       `json:"overall_score"`
	Dimensions              []Dimension               `json:"dimensions"`
	Recommendations         []string                  `json:"recommendations"`
	Basis                   string                    `json:"basis"`
	Disclaimer              string                    `json:"disclaimer"`
	Assumptions             []string                  `json:"assumptions"`
	EvidenceSources         []DiagnosisEvidenceSource `json:"evidence_sources"`
	CreatedAt               time.Time                 `json:"created_at"`
	UpdatedAt               time.Time                 `json:"updated_at"`
	GapsSnapshot            DiagnosisGaps             `json:"-"`
	RecommendationsSnapshot DiagnosisRecommendations  `json:"-"`
	PlanSnapshot            DiagnosisPlan             `json:"-"`
	ReportSnapshot          DiagnosisReport           `json:"-"`
}

type CreateDiagnosisInput struct {
	UserID         int64              `json:"-"`
	Goal           string             `json:"goal"`
	Project        string             `json:"project"`
	FocusAbilities []string           `json:"focus_abilities"`
	WeeklyTime     string             `json:"weekly_time"`
	Bottleneck     string             `json:"bottleneck"`
	Answers        []AssessmentAnswer `json:"answers"`
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
	DiagnosisID     int64                     `json:"diagnosis_id"`
	Goal            string                    `json:"goal"`
	Project         string                    `json:"project"`
	OverallScore    int                       `json:"overall_score"`
	Gaps            []GapItem                 `json:"gaps"`
	Evidence        []string                  `json:"evidence"`
	Basis           string                    `json:"basis"`
	Disclaimer      string                    `json:"disclaimer"`
	Assumptions     []string                  `json:"assumptions"`
	EvidenceSources []DiagnosisEvidenceSource `json:"evidence_sources"`
	GeneratedAt     time.Time                 `json:"generated_at"`
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
	DiagnosisID     int64                     `json:"diagnosis_id"`
	Goal            string                    `json:"goal"`
	Project         string                    `json:"project"`
	Focus           []RecommendationFocus     `json:"focus"`
	Recommendations []string                  `json:"recommendations"`
	Methods         []LearningMethod          `json:"methods"`
	Basis           string                    `json:"basis"`
	Disclaimer      string                    `json:"disclaimer"`
	Assumptions     []string                  `json:"assumptions"`
	EvidenceSources []DiagnosisEvidenceSource `json:"evidence_sources"`
	GeneratedAt     time.Time                 `json:"generated_at"`
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

type PlanItem struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	DiagnosisID int64      `json:"diagnosis_id"`
	StageNumber int        `json:"stage_number"`
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type UpdatePlanItemInput struct {
	UserID      int64 `json:"-"`
	DiagnosisID int64 `json:"-"`
	StageNumber int   `json:"-"`
	Completed   bool  `json:"completed"`
}

type DiagnosisPlan struct {
	DiagnosisID      int64                     `json:"diagnosis_id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	Recommendations  []string                  `json:"recommendations"`
	Stages           []PlanStage               `json:"stages"`
	Items            []PlanItem                `json:"items"`
	EstimatedHours   int                       `json:"estimated_hours"`
	WeeklySuggestion string                    `json:"weekly_suggestion"`
	Basis            string                    `json:"basis"`
	Disclaimer       string                    `json:"disclaimer"`
	Assumptions      []string                  `json:"assumptions"`
	EvidenceSources  []DiagnosisEvidenceSource `json:"evidence_sources"`
	GeneratedAt      time.Time                 `json:"generated_at"`
}

type DiagnosisReport struct {
	DiagnosisID     int64                     `json:"diagnosis_id"`
	Goal            string                    `json:"goal"`
	Project         string                    `json:"project"`
	OverallScore    int                       `json:"overall_score"`
	Dimensions      []Dimension               `json:"dimensions"`
	PriorityGaps    []GapItem                 `json:"priority_gaps"`
	Recommendations []string                  `json:"recommendations"`
	Evidence        []string                  `json:"evidence"`
	Basis           string                    `json:"basis"`
	Disclaimer      string                    `json:"disclaimer"`
	Assumptions     []string                  `json:"assumptions"`
	EvidenceSources []DiagnosisEvidenceSource `json:"evidence_sources"`
	GeneratedAt     time.Time                 `json:"generated_at"`
}
