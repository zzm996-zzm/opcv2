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

	V2StatusDraft      = "draft"
	V2StatusClarifying = "clarifying"
	V2StatusReady      = "ready"
	V2StatusRunning    = "running"
	V2StatusPartial    = "partial"
	V2StatusDone       = "done"
	V2StatusFailed     = "failed"
	V2StatusNoResult   = "ai_no_result"

	V2EventRunQueued   = "run_queued"
	V2EventRoleQueued  = "role_queued"
	V2EventRoleStart   = "role_start"
	V2EventToken       = "token"
	V2EventRoleDone    = "role_done"
	V2EventRoleFailed  = "role_failed"
	V2EventReportStart = "report_start"
	V2EventRunDone     = "run_done"
	V2EventError       = "error"
)

var (
	ErrServiceNotReady  = errors.New("sandbox service is not configured")
	ErrInvalidAIResult  = errors.New("invalid sandbox ai result")
	ErrInvalidSession   = errors.New("invalid sandbox session")
	ErrSessionNotFound  = errors.New("sandbox session not found")
	ErrStaleRun         = errors.New("stale sandbox run")
	ErrInvalidIntake    = errors.New("invalid sandbox intake")
	ErrIntakeIncomplete = errors.New("sandbox intake is incomplete")
	ErrV2RunNotFound    = errors.New("sandbox run not found")
	ErrV2InvalidRequest = errors.New("invalid sandbox v2 request")
	ErrV2InvalidRoles   = errors.New("invalid sandbox roles")
	ErrV2ActiveRun      = errors.New("sandbox user already has an active run")
	ErrV2StartConflict  = errors.New("sandbox run start conflict")
	ErrV2Revision       = errors.New("sandbox run revision conflict")
	ErrV2ExportExpired  = errors.New("sandbox export expired")
	ErrV2InvalidEvent   = errors.New("invalid sandbox analytics event")
)

var V2RoleCodes = []string{"customer", "investor", "competitor", "channel", "supply", "expert", "skeptic", "partner"}

type Role struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Badge       string `json:"badge"`
}

const (
	SandboxEventHomeView     = "sandbox_home_view"
	SandboxEventDraftCreate  = "sandbox_draft_create"
	SandboxEventClarifyRound = "sandbox_clarify_round"
	SandboxEventRolesSelect  = "sandbox_roles_select"
	SandboxEventRunStart     = "sandbox_run_start"
	SandboxEventRoleStart    = "sandbox_role_start"
	SandboxEventRoleDone     = "sandbox_role_done"
	SandboxEventRoleFailed   = "sandbox_role_failed"
	SandboxEventReportView   = "sandbox_report_view"
	SandboxEventReportAction = "sandbox_report_action"
	SandboxEventHistoryView  = "sandbox_history_view"
	SandboxEventQuotaBlock   = "sandbox_quota_block"
)

type SandboxAnalyticsInput struct {
	UserID     int64          `json:"-"`
	EventID    string         `json:"event_id"`
	EventName  string         `json:"event"`
	VisitorKey string         `json:"visitor_key,omitempty"`
	Route      string         `json:"route,omitempty"`
	RunID      int64          `json:"run_id,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type SandboxAnalyticsReceipt struct {
	EventID   string `json:"event_id"`
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
}
type SandboxAnalyticsEvent struct {
	EventID, EventName, VisitorHash, Route string
	UserID                                 *int64
	RunID                                  *int64
	Properties                             map[string]any
	OccurredAt, CreatedAt                  time.Time
}

type AskV2RoleInput struct {
	UserID   int64  `json:"-"`
	RunID    int64  `json:"-"`
	RoleCode string `json:"role_code"`
	Question string `json:"question"`
}
type V2FollowUp struct {
	ID               int64     `json:"id"`
	RunID            int64     `json:"run_id"`
	UserID           int64     `json:"-"`
	RoleCode         string    `json:"role_code"`
	Question         string    `json:"question"`
	Answer           string    `json:"answer"`
	InputContextHash string    `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
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

type V2Product struct {
	Name         string `json:"name,omitempty"`
	PriceCents   int64  `json:"price_cents,omitempty"`
	SellingPoint string `json:"selling_point,omitempty"`
	Cost         string `json:"cost,omitempty"`
	Stage        string `json:"stage,omitempty"`
}

type V2RunContext struct {
	TargetCustomer string `json:"target_customer,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Market         string `json:"market,omitempty"`
	Extra          string `json:"extra,omitempty"`
}

type V2Question struct {
	Key      string   `json:"key"`
	Field    string   `json:"field"`
	Type     string   `json:"type"`
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
	Answer   string   `json:"answer,omitempty"`
	Skipped  bool     `json:"skipped,omitempty"`
}

type V2RoleConfig struct {
	Code              string   `json:"role_code"`
	DisplayName       string   `json:"display_name"`
	Description       string   `json:"description"`
	Dimensions        []string `json:"analysis_dimensions"`
	DefaultSelected   bool     `json:"default_selected"`
	Required          bool     `json:"is_required"`
	DefaultModelRoute string   `json:"default_model_route"`
	PromptVersion     string   `json:"prompt_version"`
	SystemPrompt      string   `json:"-"`
}

type V2RoleOutput struct {
	RoleCode          string             `json:"role_code"`
	Stance            string             `json:"stance"`
	Verdict           string             `json:"verdict"`
	Content           string             `json:"content"`
	DimensionScores   []V2DimensionScore `json:"dimension_scores"`
	KeyFindings       []string           `json:"key_findings"`
	Risks             []V2RoleRisk       `json:"risks"`
	Recommendations   []V2Recommendation `json:"recommendations"`
	QuestionsToVerify []string           `json:"questions_to_validate"`
	Assumptions       []string           `json:"assumptions"`
	KillCriteria      []string           `json:"kill_criteria,omitempty"`
	IsModelGenerated  bool               `json:"is_model_generated"`
}

type V2DimensionScore struct {
	Code        string   `json:"code"`
	Score       int      `json:"score"`
	Basis       string   `json:"basis"`
	Confidence  float64  `json:"confidence"`
	EvidenceRef []string `json:"evidence_refs"`
}

type V2RoleRisk struct {
	Point    string `json:"point"`
	Severity string `json:"severity"`
	Basis    string `json:"basis"`
}

type V2Recommendation struct {
	Action string `json:"action"`
	Why    string `json:"why"`
}

type V2RunRole struct {
	RunID         int64         `json:"run_id"`
	RoleCode      string        `json:"role_code"`
	Seq           int           `json:"seq"`
	RoleSessionID string        `json:"role_session_id"`
	ModelProvider string        `json:"model_provider,omitempty"`
	ModelName     string        `json:"model_name,omitempty"`
	ModelRoute    string        `json:"model_route"`
	PromptVersion string        `json:"prompt_version"`
	SystemPrompt  string        `json:"-"`
	Dimensions    []string      `json:"analysis_dimensions"`
	InputHash     string        `json:"input_context_hash"`
	Status        string        `json:"status"`
	Stance        string        `json:"stance,omitempty"`
	Output        *V2RoleOutput `json:"output,omitempty"`
	InputTokens   int           `json:"input_tokens"`
	OutputTokens  int           `json:"output_tokens"`
	LatencyMS     int           `json:"latency_ms,omitempty"`
	RetryCount    int           `json:"retry_count"`
	ErrorCode     string        `json:"error_code,omitempty"`
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	FinishedAt    *time.Time    `json:"finished_at,omitempty"`
}

type V2SandboxRun struct {
	ID                   int64            `json:"id"`
	UserID               int64            `json:"-"`
	Name                 string           `json:"name,omitempty"`
	Product              V2Product        `json:"product"`
	Context              V2RunContext     `json:"context"`
	Questions            []V2Question     `json:"questions,omitempty"`
	Assumptions          []string         `json:"assumptions,omitempty"`
	Roles                []string         `json:"roles"`
	OrchestrationMode    string           `json:"orchestration_mode"`
	ModelRoutingSnapshot map[string]any   `json:"model_routing_snapshot,omitempty"`
	EvidencePack         []map[string]any `json:"evidence_pack,omitempty"`
	InputContextHash     string           `json:"input_context_hash,omitempty"`
	Completeness         float64          `json:"completeness"`
	Rounds               int              `json:"rounds"`
	Status               string           `json:"status"`
	StartedAt            *time.Time       `json:"started_at,omitempty"`
	FinishedAt           *time.Time       `json:"finished_at,omitempty"`
	Revision             int              `json:"revision"`
	RunRoles             []V2RunRole      `json:"run_roles,omitempty"`
	Report               *V2SandboxReport `json:"report,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	NextQuestions        []V2Question     `json:"next_questions,omitempty"`
	Done                 bool             `json:"done"`
}

type V2SandboxReport struct {
	Summary             string                `json:"summary"`
	Feasibility         V2Feasibility         `json:"feasibility"`
	PurchaseProbability V2PurchaseProbability `json:"purchase_probability"`
	Opportunity         []V2Insight           `json:"opportunity"`
	Risk                []V2ReportRisk        `json:"risk"`
	Advice              []V2Advice            `json:"advice"`
	RoleTakeaways       []V2RoleTakeaway      `json:"role_takeaways"`
	DimensionSummary    []V2DimensionSummary  `json:"dimension_summary"`
	Disagreements       []V2Disagreement      `json:"disagreements"`
	MissingRoles        []string              `json:"missing_roles"`
	Scenarios           map[string]V2Scenario `json:"scenarios"`
	Assumptions         []string              `json:"assumptions"`
	IsModelGenerated    bool                  `json:"is_model_generated"`
}

type V2Feasibility struct {
	Score int    `json:"score"`
	Level string `json:"level"`
	Basis string `json:"basis"`
}
type V2PurchaseProbability struct {
	ValuePct         int    `json:"value_pct"`
	Basis            string `json:"basis"`
	IsModelGenerated bool   `json:"is_model_generated"`
}
type V2Insight struct {
	Point  string `json:"point"`
	Reason string `json:"reason"`
}
type V2ReportRisk struct {
	Point      string `json:"point"`
	Severity   string `json:"severity"`
	Reason     string `json:"reason"`
	Mitigation string `json:"mitigation"`
}
type V2Advice struct {
	Action   string `json:"action"`
	Why      string `json:"why"`
	Priority int    `json:"priority"`
	Effort   string `json:"effort"`
}
type V2RoleTakeaway struct {
	Role            string             `json:"role"`
	Stance          string             `json:"stance"`
	KeyPoints       []string           `json:"key_points"`
	DimensionScores []V2DimensionScore `json:"dimension_scores"`
}
type V2DimensionSummary struct {
	Dimension       string   `json:"dimension"`
	Score           int      `json:"score"`
	Consensus       string   `json:"consensus"`
	SupportingRoles []string `json:"supporting_roles"`
	OpposingRoles   []string `json:"opposing_roles"`
}
type V2Disagreement struct {
	Topic          string       `json:"topic"`
	Views          []V2RoleView `json:"views"`
	DecisionNeeded string       `json:"decision_needed"`
}
type V2RoleView struct {
	Role  string `json:"role"`
	Point string `json:"point"`
}
type V2Scenario struct {
	Desc      string `json:"desc"`
	Condition string `json:"condition"`
}

type V2ProgressEvent struct {
	ID        int64          `json:"id"`
	RunID     int64          `json:"run_id"`
	RoleCode  string         `json:"role,omitempty"`
	Event     string         `json:"event"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type CreateV2RunInput struct {
	UserID  int64        `json:"-"`
	Name    string       `json:"name,omitempty"`
	Product V2Product    `json:"product"`
	Context V2RunContext `json:"context"`
}

type AnswerV2RunInput struct {
	UserID   int64              `json:"-"`
	RunID    int64              `json:"-"`
	Answers  []V2QuestionAnswer `json:"answers,omitempty"`
	Skip     bool               `json:"skip,omitempty"`
	Revision int                `json:"revision,omitempty"`
}

type V2QuestionAnswer struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Skipped bool   `json:"skipped,omitempty"`
}
type SetV2RolesInput struct {
	UserID   int64    `json:"-"`
	RunID    int64    `json:"-"`
	Roles    []string `json:"roles"`
	Revision int      `json:"revision,omitempty"`
}

type RenameV2RunInput struct {
	UserID   int64  `json:"-"`
	RunID    int64  `json:"-"`
	Name     string `json:"name"`
	Revision int    `json:"revision,omitempty"`
}

type V2RunListInput struct {
	UserID  int64
	Page    int
	Limit   int
	Status  string
	Product string
}

type V2RunListResult struct {
	Runs  []V2SandboxRun `json:"runs"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type V2Export struct {
	ID          int64     `json:"id"`
	RunID       int64     `json:"run_id"`
	Format      string    `json:"format"`
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type V2SandboxHome struct {
	Roles        []V2RoleConfig `json:"roles"`
	RecentRuns   []V2SandboxRun `json:"recent_runs"`
	Capabilities []string       `json:"capabilities"`
}

type CreateV2ExportInput struct {
	UserID int64  `json:"-"`
	RunID  int64  `json:"-"`
	Format string `json:"format"`
}

type SandboxTaskHandoffInput struct { UserID int64 `json:"-"`; RunID int64 `json:"-"`; AdviceIndexes []int `json:"advice_indexes"` }
type SandboxTaskHandoffResult struct { Tasks []map[string]any `json:"tasks"` }
type SandboxGrowthHandoff struct { URL string `json:"url"`; PricingCents int64 `json:"pricing_cents,omitempty"`; Channel string `json:"channel,omitempty"` }

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
