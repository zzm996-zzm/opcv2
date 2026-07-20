package sandbox

import (
	"strconv"
	"time"
)

func ExampleSessions() []Session {
	base := time.Date(2026, 5, 20, 14, 32, 0, 0, time.FixedZone("CST", 8*60*60))
	items := []struct {
		key, product, goal, target, risk string
		score, hourOffset                int
		roles                            []string
	}{
		{"ai-customer-service", "AI智能客服SaaS平台", "验证中小企业智能客服产品的商业化方案", "10-200 人规模的中小企业", "medium", 86, 0, []string{"用户视角", "投资人视角", "运营视角"}},
		{"cross-border-supply-chain", "跨境电商供应链协同平台", "评估跨境供应链协同产品的落地机会", "跨境电商品牌与供应商", "high", 79, -53, []string{"用户视角", "运营视角", "代理商 / 渠道方视角"}},
		{"ai-learning-assistant", "AI个性化学习助手", "验证 AI 学习助手的付费需求和留存", "职业教育与成人学习用户", "medium", 81, -94, []string{"用户视角", "投资人视角", "竞争对手视角"}},
		{"community-commerce", "社区团购O2O平台", "评估社区团购平台的单点盈利模型", "社区家庭用户与团长", "high", 72, -123, []string{"用户视角", "运营视角", "竞争对手视角"}},
		{"health-management", "健康管理小程序", "验证个人健康数据管理与咨询服务", "关注慢病管理的个人用户", "medium", 64, -148, []string{"用户视角", "运营视角"}},
		{"enterprise-analytics", "企业数据分析平台", "评估中小企业数据分析产品的采购意愿", "中小企业管理者", "medium", 83, -196, []string{"用户视角", "投资人视角", "代理商 / 渠道方视角"}},
		{"iot-solution", "智能硬件IoT解决方案", "验证智能硬件行业解决方案的渠道模式", "制造业与园区客户", "high", 76, -244, []string{"用户视角", "投资人视角", "运营视角", "竞争对手视角"}},
	}

	sessions := make([]Session, 0, len(items))
	for index, item := range items {
		createdAt := base.Add(time.Duration(item.hourOffset) * time.Hour)
		sessions = append(sessions, Session{
			ID:              int64(900001 + index),
			Goal:            item.goal,
			TargetUsers:     item.target,
			Product:         item.product,
			Roles:           item.roles,
			Status:          StatusCompleted,
			ProgressPercent: 100,
			CurrentStep:     StatusCompleted,
			RunAttempt:      1,
			Intake: Intake{
				Status: IntakeStatusReady, InitialIdea: item.goal,
				RecognizedFields: []RecognizedField{
					{Key: "goal", Label: "推演目标", Value: item.goal},
					{Key: "target_users", Label: "目标用户", Value: item.target},
					{Key: "product", Label: "产品方案", Value: item.product},
				},
				Questions: []IntakeQuestion{},
			},
			Settings:   DefaultRunSettings(),
			Report:     exampleReport(item.product, item.score, item.risk, item.roles),
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt.Add(8 * time.Minute),
			IsExample:  true,
			ExampleKey: item.key,
		})
	}
	return sessions
}

func exampleReport(product string, score int, risk string, roles []string) Report {
	roleSummaries := make([]RoleSummary, 0, len(roles))
	for _, role := range roles {
		roleSummaries = append(roleSummaries, RoleSummary{Role: role, View: role + "认为应先用真实客户试点验证核心假设，再决定扩大投入。"})
	}
	return Report{
		Score: score, Summary: product + "具备明确应用场景，但仍需验证真实付费意愿、交付成本和持续使用率。",
		Basis: "model_simulation", Disclaimer: "本记录为产品演示数据，仅用于展示商业沙盘界面和流程，不代表真实市场结论。",
		Assumptions: []string{"目标用户存在高频业务需求", "产品能够在可控成本内完成交付", "演示记录未接入真实市场数据"}, EvidenceSources: []ReportEvidence{},
		Metrics: []Metric{{Label: "综合可行性", Value: strconv.Itoa(score)}, {Label: "市场吸引力", Value: "8.1"}, {Label: "执行可控性", Value: "7.4"}}, RoleSummaries: roleSummaries,
		Risks: []string{"真实付费意愿尚未验证", "早期定制交付可能推高成本", "需要明确数据合规边界"}, NextActions: []string{"访谈 10 个目标客户", "完成 3 家最小范围试点", "复盘留存、成本与续费意愿"},
		ReportVersion: "sandbox_demo_v1", ConsumerProbability: min(score-10, 88), RiskLevel: risk, RecommendationGrade: "A-",
		CoreConclusions:     []string{"优先验证高影响、高不确定性的商业假设", "使用真实试点数据决定是否继续投入", "在扩大获客前先沉淀标准交付流程"},
		OpportunityAnalysis: []ReportInsight{{Title: "明确的效率提升场景", Detail: "目标客户存在可被量化的效率与转化问题，适合用小范围试点验证。", Tags: []string{"高频场景", "效率提升"}}, {Title: "标准化复制机会", Detail: "如果首批试点能沉淀配置模板和验收指标，具备跨客户复制空间。", Tags: []string{"标准化", "规模化"}}},
		RiskAnalysis:        []ReportInsight{{Title: "付费与留存风险", Detail: "体验价值不等于持续付费，需要观察真实使用频率与续费行为。", Tags: []string{"付费", "留存"}}, {Title: "交付范围风险", Detail: "过多定制需求会侵蚀毛利，应在试点前固定能力边界。", Tags: []string{"交付", "成本"}}},
		ActionPlan:          []ActionPlanItem{{Order: 1, Title: "客户访谈", Detail: "访谈 10 个目标客户并记录当前流程和购买条件。", Duration: "1 周"}, {Order: 2, Title: "最小试点", Detail: "选择 3 个客户运行最小版本并采集真实指标。", Duration: "2 周"}, {Order: 3, Title: "投入决策", Detail: "根据留存、成本与续费意愿决定继续、调整或停止。", Duration: "1 周"}},
		GrowthPath:          []GrowthPathItem{{Stage: 1, Title: "单场景验证", Detail: "验证核心需求与付费意愿"}, {Stage: 2, Title: "标准交付", Detail: "沉淀配置模板和交付清单"}, {Stage: 3, Title: "渠道复制", Detail: "基于真实留存扩大获客"}},
		ValidationMetrics:   []ValidationMetric{{Label: "试点激活率", Current: "待验证", Target: "80% 以上", ConfidencePercent: 62}, {Label: "客户续费意愿", Current: "待验证", Target: "3 家中至少 2 家", ConfidencePercent: 55}, {Label: "单客户交付周期", Current: "待验证", Target: "5 个工作日内", ConfidencePercent: 68}},
		Timeline:            []TimelineItem{{Title: "客户访谈与需求确认", Period: "第 1 周"}, {Title: "最小版本试点运行", Period: "第 2-3 周"}, {Title: "数据复盘与投入决策", Period: "第 4 周"}},
	}
}
