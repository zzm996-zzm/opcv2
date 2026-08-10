package sandbox

import (
	"context"
	"strings"
)

const v2MaxClarificationRounds = 3

type V2Repository interface {
	ListV2RoleConfigs(ctx context.Context) ([]V2RoleConfig, error)
	CreateV2Run(ctx context.Context, run V2SandboxRun) (V2SandboxRun, error)
	GetV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error)
	UpdateV2RunDraft(ctx context.Context, run V2SandboxRun, expectedRevision int) (V2SandboxRun, error)
	ListV2Runs(ctx context.Context, userID int64, limit int) ([]V2SandboxRun, error)
	DeleteV2Run(ctx context.Context, userID, runID int64) error
}

func DefaultV2RoleConfigs() []V2RoleConfig {
	return []V2RoleConfig{
		{Code: "customer", DisplayName: "目标客户", Description: "判断真实购买者会不会买", Dimensions: []string{"pain", "usage", "buyer", "alternatives", "trigger", "price", "decision", "trust", "friction", "retention"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "investor", DisplayName: "投资人", Description: "判断投入与规模化价值", Dimensions: []string{"demand", "market", "revenue", "unit_economics", "scale", "moat", "team", "capital", "cash_burn", "return"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "competitor", DisplayName: "竞争对手", Description: "寻找竞争软肋和反击路径", Dimensions: []string{"substitutes", "differentiation", "copyability", "price_war", "channel", "brand", "switching_cost", "speed", "weakness", "defense"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "channel", DisplayName: "渠道方", Description: "判断是否愿意销售", Dimensions: []string{"fit", "margin", "sales_cycle", "training", "inventory", "payment", "after_sales", "conflict", "promotion", "priority"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "supply", DisplayName: "供应链运营", Description: "判断能否稳定交付", Dimensions: []string{"supply", "cost", "capacity", "delivery", "quality", "inventory", "logistics", "service", "sop", "cash", "bottleneck"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "expert", DisplayName: "行业专家", Description: "用行业规律校正判断", Dimensions: []string{"lifecycle", "drivers", "trend", "maturity", "policy", "compliance", "benchmark", "seasonality", "window", "risk"}, DefaultSelected: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "skeptic", DisplayName: "悲观者", Description: "主动推翻乐观假设", Dimensions: []string{"fatal_assumption", "evidence_gap", "refusal", "worst_case", "cash_break", "execution", "bias", "compliance", "abuse", "kill_criteria"}, DefaultSelected: true, Required: true, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
		{Code: "partner", DisplayName: "合伙人", Description: "判断分工与合作风险", Dimensions: []string{"founder_fit", "complement", "resources", "roles", "decision", "equity", "commitment", "conflict", "milestone", "exit"}, DefaultModelRoute: "role_default", PromptVersion: "sandbox_role_v1"},
	}
}

func DefaultV2Roles() []string {
	roles := make([]string, 0, 7)
	for _, role := range DefaultV2RoleConfigs() {
		if role.DefaultSelected {
			roles = append(roles, role.Code)
		}
	}
	return roles
}

func (s *Service) ListV2Roles(ctx context.Context) ([]V2RoleConfig, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return nil, err
	}
	return repository.ListV2RoleConfigs(ctx)
}

func (s *Service) CreateV2Run(ctx context.Context, input CreateV2RunInput) (V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Product = normalizeV2Product(input.Product)
	input.Context = normalizeV2Context(input.Context)
	if input.UserID <= 0 || input.Product.Name == "" {
		return V2SandboxRun{}, ErrV2InvalidRequest
	}
	if input.Name == "" {
		input.Name = input.Product.Name
	}
	now := s.now()
	run := V2SandboxRun{
		UserID: input.UserID, Name: input.Name, Product: input.Product, Context: input.Context,
		Roles: DefaultV2Roles(), OrchestrationMode: "isolated_sessions", ModelRoutingSnapshot: map[string]any{},
		EvidencePack: []map[string]any{}, Revision: 1, CreatedAt: now, UpdatedAt: now,
	}
	run.Completeness = v2Completeness(run)
	run.Questions = buildV2Questions(run, nil)
	if run.Completeness >= 0.8 {
		run.Status = V2StatusReady
	} else {
		run.Status = V2StatusClarifying
	}
	return repository.CreateV2Run(ctx, run)
}

func (s *Service) AnswerV2Run(ctx context.Context, input AnswerV2RunInput) (V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	if input.UserID <= 0 || input.RunID <= 0 || (!input.Skip && len(input.Answers) == 0) || len(input.Answers) > 3 {
		return V2SandboxRun{}, ErrV2InvalidRequest
	}
	run, err := repository.GetV2Run(ctx, input.UserID, input.RunID)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if input.Revision > 0 && input.Revision != run.Revision {
		return V2SandboxRun{}, ErrV2Revision
	}
	if run.Status != V2StatusClarifying && run.Status != V2StatusDraft {
		return V2SandboxRun{}, ErrV2InvalidRequest
	}
	answered := make(map[string]bool, len(input.Answers))
	for _, answer := range input.Answers {
		key := strings.TrimSpace(answer.Key)
		value := strings.TrimSpace(answer.Value)
		if key == "" || (value == "" && !answer.Skipped) || answered[key] {
			return V2SandboxRun{}, ErrV2InvalidRequest
		}
		if !applyV2Answer(&run, key, value) {
			return V2SandboxRun{}, ErrV2InvalidRequest
		}
		answered[key] = true
		markV2Question(&run, key, value, answer.Skipped)
		if answer.Skipped {
			run.Assumptions = appendUniqueString(run.Assumptions, "用户跳过澄清项："+key)
		}
	}
	if input.Skip {
		for i := range run.Questions {
			if strings.TrimSpace(run.Questions[i].Answer) == "" && !run.Questions[i].Skipped {
				run.Questions[i].Skipped = true
				run.Assumptions = appendUniqueString(run.Assumptions, "用户跳过澄清项："+run.Questions[i].Key)
			}
		}
	}
	run.Rounds++
	run.Completeness = v2Completeness(run)
	if input.Skip || run.Completeness >= 0.8 || run.Rounds >= v2MaxClarificationRounds {
		run.Status = V2StatusReady
	} else {
		run.Status = V2StatusClarifying
		run.Questions = buildV2Questions(run, answered)
	}
	return repository.UpdateV2RunDraft(ctx, run, run.Revision)
}

func (s *Service) SetV2Roles(ctx context.Context, input SetV2RolesInput) (V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	roles, err := normalizeV2Roles(input.Roles)
	if err != nil {
		return V2SandboxRun{}, err
	}
	run, err := repository.GetV2Run(ctx, input.UserID, input.RunID)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if input.Revision > 0 && input.Revision != run.Revision {
		return V2SandboxRun{}, ErrV2Revision
	}
	if run.Status == V2StatusRunning || run.Status == V2StatusDone || run.Status == V2StatusPartial {
		return V2SandboxRun{}, ErrV2InvalidRequest
	}
	run.Roles = roles
	return repository.UpdateV2RunDraft(ctx, run, run.Revision)
}

func (s *Service) GetV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	return repository.GetV2Run(ctx, userID, runID)
}

func (s *Service) ListV2Runs(ctx context.Context, userID int64, limit int) ([]V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return repository.ListV2Runs(ctx, userID, limit)
}

func (s *Service) DeleteV2Run(ctx context.Context, userID, runID int64) error {
	repository, err := s.v2Repository()
	if err != nil {
		return err
	}
	run, err := repository.GetV2Run(ctx, userID, runID)
	if err != nil {
		return err
	}
	if run.Status == V2StatusRunning {
		return ErrV2InvalidRequest
	}
	return repository.DeleteV2Run(ctx, userID, runID)
}

func (s *Service) v2Repository() (V2Repository, error) {
	repository, ok := s.repository.(V2Repository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func normalizeV2Roles(input []string) ([]string, error) {
	if len(input) == 0 || len(input) > len(V2RoleCodes) {
		return nil, ErrV2InvalidRoles
	}
	allowed := make(map[string]bool, len(V2RoleCodes))
	for _, role := range V2RoleCodes {
		allowed[role] = true
	}
	seen := make(map[string]bool, len(input))
	roles := make([]string, 0, len(input))
	for _, value := range input {
		role := strings.TrimSpace(value)
		if !allowed[role] || seen[role] {
			return nil, ErrV2InvalidRoles
		}
		seen[role] = true
		roles = append(roles, role)
	}
	if !seen["skeptic"] {
		return nil, ErrV2InvalidRoles
	}
	return roles, nil
}

func normalizeV2Product(product V2Product) V2Product {
	product.Name = strings.TrimSpace(product.Name)
	product.SellingPoint = strings.TrimSpace(product.SellingPoint)
	product.Cost = strings.TrimSpace(product.Cost)
	product.Stage = strings.TrimSpace(product.Stage)
	return product
}

func normalizeV2Context(runContext V2RunContext) V2RunContext {
	runContext.TargetCustomer = strings.TrimSpace(runContext.TargetCustomer)
	runContext.Channel = strings.TrimSpace(runContext.Channel)
	runContext.Market = strings.TrimSpace(runContext.Market)
	runContext.Extra = strings.TrimSpace(runContext.Extra)
	return runContext
}

func v2Completeness(run V2SandboxRun) float64 {
	values := []bool{
		run.Product.Name != "",
		run.Product.SellingPoint != "",
		run.Context.TargetCustomer != "",
		run.Context.Channel != "" || run.Context.Market != "",
		run.Product.PriceCents > 0 || run.Product.Cost != "" || run.Product.Stage != "",
	}
	complete := 0
	for _, value := range values {
		if value {
			complete++
		}
	}
	return float64(complete) / float64(len(values))
}

func buildV2Questions(run V2SandboxRun, exclude map[string]bool) []V2Question {
	candidates := []V2Question{
		{Key: "selling_point", Field: "product.selling_point", Type: "text", Question: "这个产品解决的核心问题和主要卖点是什么？", Required: true},
		{Key: "target_customer", Field: "context.target_customer", Type: "text", Question: "谁是最明确的目标客户和实际付费者？", Required: true},
		{Key: "channel", Field: "context.channel", Type: "text", Question: "计划通过什么渠道触达和成交？", Required: true},
		{Key: "market", Field: "context.market", Type: "text", Question: "目标市场或经营区域是什么？", Required: false},
		{Key: "stage", Field: "product.stage", Type: "text", Question: "项目当前处于什么阶段？", Required: false},
		{Key: "cost", Field: "product.cost", Type: "text", Question: "当前已知的成本或预算约束是什么？", Required: false},
		{Key: "price", Field: "product.price_cents", Type: "number", Question: "预期售价是多少元？", Required: false},
	}
	questions := make([]V2Question, 0, 3)
	for _, question := range candidates {
		if exclude[question.Key] || v2FieldKnown(run, question.Key) {
			continue
		}
		questions = append(questions, question)
		if len(questions) == 3 {
			break
		}
	}
	return questions
}

func v2FieldKnown(run V2SandboxRun, key string) bool {
	switch key {
	case "selling_point":
		return run.Product.SellingPoint != ""
	case "target_customer":
		return run.Context.TargetCustomer != ""
	case "channel":
		return run.Context.Channel != ""
	case "market":
		return run.Context.Market != ""
	case "stage":
		return run.Product.Stage != ""
	case "cost":
		return run.Product.Cost != ""
	case "price":
		return run.Product.PriceCents > 0
	default:
		return false
	}
}

func applyV2Answer(run *V2SandboxRun, key, value string) bool {
	switch key {
	case "selling_point":
		run.Product.SellingPoint = value
	case "target_customer":
		run.Context.TargetCustomer = value
	case "channel":
		run.Context.Channel = value
	case "market":
		run.Context.Market = value
	case "stage":
		run.Product.Stage = value
	case "cost":
		run.Product.Cost = value
	case "price":
		if value == "" {
			return true
		}
		var cents int64
		for _, char := range value {
			if char < '0' || char > '9' {
				return false
			}
			cents = cents*10 + int64(char-'0')
		}
		run.Product.PriceCents = cents * 100
	default:
		return false
	}
	return true
}

func markV2Question(run *V2SandboxRun, key, value string, skipped bool) {
	for i := range run.Questions {
		if run.Questions[i].Key == key {
			run.Questions[i].Answer = value
			run.Questions[i].Skipped = skipped
			return
		}
	}
}

func appendUniqueString(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}
