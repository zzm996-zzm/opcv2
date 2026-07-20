package sandbox

import (
	"errors"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"

	IntakeStatusQuestions = "questions"
	IntakeStatusReady     = "ready"

	RunDepthStandard = "standard"
	RunDepthDeep     = "deep"

	OutputStyleStructured = "structured_report"
	OutputStyleConcise    = "concise_report"
)

var (
	ErrServiceNotReady  = errors.New("sandbox service is not configured")
	ErrInvalidAIResult  = errors.New("invalid sandbox ai result")
	ErrInvalidSession   = errors.New("invalid sandbox session")
	ErrSessionNotFound  = errors.New("sandbox session not found")
	ErrStaleRun         = errors.New("stale sandbox run")
	ErrInvalidIntake    = errors.New("invalid sandbox intake")
	ErrIntakeIncomplete = errors.New("sandbox intake is incomplete")
)

type Role struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Badge       string `json:"badge"`
}

type CreateInput struct {
	UserID      int64    `json:"-"`
	Goal        string   `json:"goal"`
	TargetUsers string   `json:"target_users"`
	Product     string   `json:"product"`
	Roles       []string `json:"roles"`
}

type DraftUpdate struct {
	Goal        *string      `json:"goal,omitempty"`
	TargetUsers *string      `json:"target_users,omitempty"`
	Product     *string      `json:"product,omitempty"`
	Roles       *[]string    `json:"roles,omitempty"`
	Settings    *RunSettings `json:"settings,omitempty"`
}

type AskRoleInput struct {
	UserID    int64  `json:"-"`
	SessionID int64  `json:"-"`
	Role      string `json:"role"`
	Question  string `json:"question"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

func DefaultRoles() []Role {
	return []Role{
		{Key: "user", Label: "用户视角", Description: "评估产品体验与价值", Badge: "推荐优先"},
		{Key: "investor", Label: "投资人视角", Description: "评估市场潜力与回报", Badge: "热门选择"},
		{Key: "channel", Label: "代理商 / 渠道方视角", Description: "评估项目落地可行性", Badge: "渠道必选"},
		{Key: "competitor", Label: "竞争对手视角", Description: "评估竞争格局与策略", Badge: "深度分析"},
		{Key: "operator", Label: "运营视角", Description: "评估执行与增长策略", Badge: "运营必选"},
	}
}

func DefaultSystemPerspectives() []Role {
	return []Role{
		{Key: "growth", Label: "增长路径", Description: "规划获客与规模化路径", Badge: "系统分析"},
		{Key: "risk", Label: "风险研判", Description: "识别合规、交付与经营风险", Badge: "系统分析"},
	}
}

type SettingOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Recommended bool   `json:"recommended,omitempty"`
}

type RunSettings struct {
	Depth           string            `json:"depth"`
	OutputStyle     string            `json:"output_style"`
	GenerateOutline bool              `json:"generate_outline"`
	Variables       map[string]string `json:"variables"`
}

func DefaultRunSettings() RunSettings {
	return RunSettings{
		Depth:           RunDepthStandard,
		OutputStyle:     OutputStyleStructured,
		GenerateOutline: true,
		Variables:       map[string]string{},
	}
}

type Options struct {
	Roles              []Role          `json:"roles"`
	SystemPerspectives []Role          `json:"system_perspectives"`
	Depths             []SettingOption `json:"depths"`
	OutputStyles       []SettingOption `json:"output_styles"`
	Defaults           RunSettings     `json:"defaults"`
}

func DefaultOptions() Options {
	return Options{
		Roles:              DefaultRoles(),
		SystemPerspectives: DefaultSystemPerspectives(),
		Depths: []SettingOption{
			{Value: RunDepthStandard, Label: "标准", Description: "覆盖核心机会、风险与行动建议", Recommended: true},
			{Value: RunDepthDeep, Label: "深度", Description: "增加验证指标、增长路径与建议时间线"},
		},
		OutputStyles: []SettingOption{
			{Value: OutputStyleStructured, Label: "结构化报告", Description: "按结论、机会、风险与行动计划组织", Recommended: true},
			{Value: OutputStyleConcise, Label: "精简摘要", Description: "突出关键结论与下一步行动"},
		},
		Defaults: DefaultRunSettings(),
	}
}

type RecognizedField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type IntakeQuestion struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Hint        string `json:"hint"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	MaxLength   int    `json:"max_length"`
	Position    int    `json:"position"`
	Answer      string `json:"answer,omitempty"`
	Skipped     bool   `json:"skipped"`
}

type Intake struct {
	Status           string            `json:"status"`
	InitialIdea      string            `json:"initial_idea"`
	RecognizedFields []RecognizedField `json:"recognized_fields"`
	Questions        []IntakeQuestion  `json:"questions"`
	AnsweredCount    int               `json:"answered_count"`
	TotalQuestions   int               `json:"total_questions"`
}

func DefaultReadyIntake() Intake {
	return Intake{
		Status:           IntakeStatusReady,
		RecognizedFields: []RecognizedField{},
		Questions:        []IntakeQuestion{},
	}
}

type IntakeCreateInput struct {
	UserID      int64  `json:"-"`
	InitialIdea string `json:"initial_idea"`
}

type IntakeAnswerInput struct {
	UserID      int64  `json:"-"`
	SessionID   int64  `json:"-"`
	QuestionKey string `json:"-"`
	Answer      string `json:"answer"`
	Skipped     bool   `json:"skipped"`
}

type Session struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	Goal            string      `json:"goal"`
	TargetUsers     string      `json:"target_users"`
	Product         string      `json:"product"`
	Roles           []string    `json:"roles"`
	Status          string      `json:"status"`
	ProgressPercent int         `json:"progress_percent"`
	CurrentStep     string      `json:"current_step"`
	ErrorMessage    string      `json:"error_message,omitempty"`
	RunAttempt      int         `json:"run_attempt"`
	Intake          Intake      `json:"intake"`
	Settings        RunSettings `json:"settings"`
	Report          Report      `json:"report,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	IsExample       bool        `json:"is_example,omitempty"`
	ExampleKey      string      `json:"example_key,omitempty"`
}

type Report struct {
	Score               int                `json:"score"`
	Summary             string             `json:"summary"`
	Basis               string             `json:"basis"`
	Disclaimer          string             `json:"disclaimer"`
	Assumptions         []string           `json:"assumptions"`
	EvidenceSources     []ReportEvidence   `json:"evidence_sources"`
	Metrics             []Metric           `json:"metrics"`
	RoleSummaries       []RoleSummary      `json:"role_summaries"`
	Risks               []string           `json:"risks"`
	NextActions         []string           `json:"next_actions"`
	ReportVersion       string             `json:"report_version,omitempty"`
	ConsumerProbability int                `json:"consumer_probability,omitempty"`
	RiskLevel           string             `json:"risk_level,omitempty"`
	RecommendationGrade string             `json:"recommendation_grade,omitempty"`
	CoreConclusions     []string           `json:"core_conclusions,omitempty"`
	OpportunityAnalysis []ReportInsight    `json:"opportunity_analysis,omitempty"`
	RiskAnalysis        []ReportInsight    `json:"risk_analysis,omitempty"`
	ActionPlan          []ActionPlanItem   `json:"action_plan,omitempty"`
	GrowthPath          []GrowthPathItem   `json:"growth_path,omitempty"`
	ValidationMetrics   []ValidationMetric `json:"validation_metrics,omitempty"`
	Timeline            []TimelineItem     `json:"timeline,omitempty"`
}

type ReportInsight struct {
	Title  string   `json:"title"`
	Detail string   `json:"detail"`
	Tags   []string `json:"tags"`
}

type ActionPlanItem struct {
	Order    int    `json:"order"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Duration string `json:"duration"`
}

type GrowthPathItem struct {
	Stage  int    `json:"stage"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type ValidationMetric struct {
	Label             string `json:"label"`
	Current           string `json:"current"`
	Target            string `json:"target"`
	ConfidencePercent int    `json:"confidence_percent"`
}

type TimelineItem struct {
	Title  string `json:"title"`
	Period string `json:"period"`
}

type ReportEvidence struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	CapturedAt time.Time `json:"captured_at"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type RoleSummary struct {
	Role string `json:"role"`
	View string `json:"view"`
}
