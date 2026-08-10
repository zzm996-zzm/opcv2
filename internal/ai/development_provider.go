package ai

import (
	"context"
	"encoding/json"
	"strings"
)

type DevelopmentProvider struct {
	Response []byte
}

func NewDevelopmentProvider() *DevelopmentProvider {
	return &DevelopmentProvider{}
}

func (p *DevelopmentProvider) Generate(_ context.Context, request ProviderRequest) (ProviderResponse, error) {
	response := p.responseFor(request)
	return ProviderResponse{Content: response}, nil
}

func (p *DevelopmentProvider) Stream(_ context.Context, request ProviderRequest, onDelta func([]byte) error) (ProviderResponse, error) {
	content := p.responseFor(request)
	if err := onDelta(content); err != nil {
		return ProviderResponse{}, err
	}
	return ProviderResponse{Content: content}, nil
}

func (p *DevelopmentProvider) responseFor(request ProviderRequest) []byte {
	if len(p.Response) > 0 {
		return p.Response
	}
	switch request.Feature {
	case "copilot.chat_stream":
		return []byte("我会先明确目标客户，再验证最高频痛点，最后做一个低成本样板。")
	case "analysis.direction":
		return []byte(`{
			"status":"completed",
			"cards":[
				{
					"name":"本地教培小班陪跑",
					"score":91,
					"reasons":["经验匹配度高","成都本地可先小范围验证","5万预算足够启动第一批样板"],
					"market_evidence":"双减后家长更关注个性化与提分陪跑，可先围绕公开家长社群和本地内容平台验证需求。",
					"difficulty":{"level":"中","notes":["首批客户获取","合规宣传边界","交付标准化"]},
					"benchmarks":["本地数学提分工作室","小红书家长陪跑账号"],
					"actions":["写出3个目标家长画像","发布1篇本地提分案例笔记","联系10个已有家长资源做访谈"],
					"upsell":"升级后可直接生成100条本地教培相关企业与社群线索。"
				}
			]
		}`)
	case "projects.match":
		return []byte(`{
			"status":"completed",
			"projects":[
				{
					"rank":1,
					"title":"AI短视频脚本工作室",
					"score":94,
					"tags":["内容创作","低成本启动","高需求"],
					"budget":"¥1,000 - ¥3,000",
					"reasons":["能力匹配","启动成本低","市场需求明确"],
					"risk":"同质化竞争较多，需打造差异化定位"
				},
				{
					"rank":2,
					"title":"Excel自动化报表定制",
					"score":82,
					"tags":["办公效率","刚需服务","稳定复购"],
					"budget":"¥1,500 - ¥4,000",
					"reasons":["交付标准化","复购机会多","适合轻量试跑"],
					"risk":"需控制需求边界和交付周期"
				}
			]
		}`)
	case "projects.match_analysis":
		if strings.Contains(request.UserPrompt, `"budget_band"`) && strings.Contains(request.UserPrompt, `"time_per_week"`) && strings.Contains(request.UserPrompt, `"risk_preference"`) {
			return []byte(`{
				"analysis_summary":"信息完整，可以开始生成匹配结果。",
				"parsed_profile":{"skills":["内容创作"],"team_size":1},
				"field_sources":{"budget_band":[{"type":"answer","locator":"budget"}],"time_per_week":[{"type":"answer","locator":"time"}],"risk_preference":[{"type":"answer","locator":"risk"}]},
				"completeness":0.9,
				"missing_fields":[],
				"questions":[]
			}`)
		}
		return []byte(`{
			"analysis_summary":"已识别项目方向和个人能力，仍需确认预算、时间投入和风险偏好。",
			"parsed_profile":{"skills":["内容创作"],"team_size":1},
			"field_sources":{"skills":[{"type":"text","locator":"need"}],"team_size":[{"type":"inference","locator":"need"}]},
			"completeness":0.55,
			"missing_fields":["budget_band","time_per_week","risk_preference"],
			"questions":[
				{"id":"budget","field":"budget_band","type":"single","question":"你的启动预算范围是多少？","options":["0-5k","5k-2w","2w以上"],"required":true,"reason":"预算决定可行的项目范围"},
				{"id":"time","field":"time_per_week","type":"single","question":"每周可以投入多少时间？","options":["5小时以内","5-20小时","20小时以上"],"required":true,"reason":"时间投入影响项目复杂度"},
				{"id":"risk","field":"risk_preference","type":"single","question":"你的风险偏好是什么？","options":["低","中","高"],"required":true,"reason":"风险偏好影响匹配排序"}
			]
		}`)
	case "projects.diagnose":
		return []byte(`{
			"fit_score":78,
			"verdict":"recommended",
			"reasons":["项目启动成本可控","方向与用户目标一致"],
			"prerequisites":["明确首批目标客户","完成最小可行方案"],
			"next_steps":["访谈3位目标用户","完成一次付费意向验证"]
		}`)
	case "sandbox.intake":
		return []byte(`{
			"goal":"验证面向本地教培机构的 AI 客服与企微转化助手是否值得投入开发和推广",
			"target_users":"拥有 30-200 人团队、存在高频招生咨询与私域转化需求的本地教培机构",
			"product":"接入企业微信的 AI 客服与线索转化助手，为教培机构提供自动答疑、意向识别和销售跟进建议",
			"recognized_fields":[
				{"key":"goal","label":"推演目标","value":"验证面向本地教培机构的 AI 客服与企微转化助手是否值得投入开发和推广"},
				{"key":"target_users","label":"目标用户","value":"拥有 30-200 人团队、存在高频招生咨询与私域转化需求的本地教培机构"},
				{"key":"product","label":"产品方案","value":"接入企业微信的 AI 客服与线索转化助手，为教培机构提供自动答疑、意向识别和销售跟进建议"}
			],
			"questions":[
				{"key":"customer_pain","title":"目标客户目前最急需解决的问题是什么？","hint":"描述最常出现、影响成交或交付的具体问题。","placeholder":"例如：招生旺季咨询量大，销售无法及时跟进高意向家长。","required":true,"max_length":1000,"position":1},
				{"key":"current_solution","title":"客户现在如何解决这个问题？","hint":"说明现有流程、工具或人工方式，以及主要不足。","placeholder":"例如：由课程顾问轮班回复企微，靠表格记录意向。","required":true,"max_length":1000,"position":2},
				{"key":"value_proposition","title":"你的方案能为客户带来什么可衡量的价值？","hint":"优先填写效率、收入、成本或体验方面的指标。","placeholder":"例如：首次响应时间降到 1 分钟内，销售有效跟进率提升 30%。","required":true,"max_length":1200,"position":3},
				{"key":"business_model","title":"你准备如何收费并获得第一批客户？","hint":"描述价格、销售渠道和首批验证范围。","placeholder":"例如：按门店收取月费，通过已有教培客户资源完成 3 家试点。","required":false,"max_length":1200,"position":4},
				{"key":"success_criteria","title":"这次推演最希望验证哪些关键结果？","hint":"列出决定继续、调整或停止项目的判断标准。","placeholder":"例如：客户愿意付费、数据合规可控、单店交付成本在预算内。","required":true,"max_length":1500,"position":5}
			]
		}`)
	case "sandbox.follow_up":
		return []byte(`{
			"answer":"从当前推演结果看，我会优先关注付费客户留存、单店交付成本和获客回收周期。建议先用 3 家教培机构做 4 周试点，并用真实的响应时长、有效线索率和续费意愿决定是否扩大投入。"
		}`)
	case "sandbox.role":
		return sandboxV2RoleDevelopmentResponse(request.UserPrompt)
	case "sandbox.report":
		return []byte(`{
			"summary":"当前方案具备明确使用场景，建议用小范围付费试点验证需求强度、交付成本和复购意愿。",
			"feasibility":{"score":72,"level":"mid","basis":"多数角色认可场景价值，但对获客和交付成本仍有保留。"},
			"purchase_probability":{"value_pct":63,"basis":"基于角色模拟，不代表真实市场统计。","is_model_generated":true},
			"opportunity":[{"point":"高频流程可被标准化","reason":"目标客户存在重复且可量化的业务动作。"}],
			"risk":[{"point":"早期定制需求过多","severity":"high","reason":"不同客户流程存在差异。","mitigation":"限定首期场景和配置边界。"}],
			"advice":[{"action":"完成3家付费试点","why":"用真实成交和交付数据替代主观判断。","priority":1,"effort":"medium"}],
			"role_takeaways":[],
			"dimension_summary":[],
			"disagreements":[{"topic":"是否立即扩大投入","views":[{"role":"investor","point":"先验证单位经济模型"},{"role":"customer","point":"先证明使用价值"}],"decision_needed":"达到试点续费标准后再扩大投入"}],
			"missing_roles":[],
			"scenarios":{"base":{"desc":"小范围试点后逐步复制","condition":"3家试点中至少2家续费"}},
			"assumptions":["当前输入能够代表首批目标客户"],
			"is_model_generated":true
		}`)
	case "sandbox.run":
		return []byte(`{
			"report_version":"sandbox_report_v3",
			"score":83,
			"summary":"建议先做小范围客户验证。当前方案具备明确场景和可解释价值，但需要优先验证付费意愿、数据安全顾虑和交付成本。",
			"consumer_probability":72,
			"risk_level":"medium",
			"recommendation_grade":"A-",
			"assumptions":["目标客户存在高频咨询场景","客户愿意为效率提升付费","试点期间可以合规使用必要数据"],
			"metrics":[
				{"label":"市场吸引力","value":"8.4"},
				{"label":"落地难度","value":"中等"},
				{"label":"回本周期","value":"6-10周"}
			],
			"role_summaries":[
				{"role":"用户","view":"更关注响应效率、隐私安全和是否能接入现有企微流程。"},
				{"role":"投资人","view":"会重点观察客户获取成本、续费率和交付是否足够标准化。"},
				{"role":"增长顾问","view":"建议用教培机构的高频咨询场景切入，先做一个可复制样板。"}
			],
			"risks":["客户教育成本偏高","敏感数据合规要求高","早期交付容易被定制需求拖慢"],
			"next_actions":["访谈10个目标客户，确认高频咨询问题","做出一个教培场景演示样板","定义首月试点价格和成功指标"],
			"core_conclusions":[
				"教培机构高频咨询与企微跟进是一个适合小范围验证的明确切入场景。",
				"产品价值必须通过响应效率、有效线索率和续费意愿三类真实指标验证。",
				"在合规和交付标准化尚未验证前，不建议直接扩大获客投入。"
			],
			"opportunity_analysis":[
				{"title":"旺季咨询自动分流","detail":"招生旺季咨询集中，自动答疑和意向识别可以缩短首次响应时间并减少销售漏跟。","tags":["高频场景","效率提升"]},
				{"title":"企微跟进标准化","detail":"将咨询摘要和下一步建议直接同步给课程顾问，有机会形成可复制的门店交付流程。","tags":["私域转化","标准化"]}
			],
			"risk_analysis":[
				{"title":"数据合规风险","detail":"家长和学生信息属于敏感业务数据，试点前必须明确最小采集范围、授权方式和留存规则。","tags":["合规","数据安全"]},
				{"title":"定制交付风险","detail":"不同机构的话术和流程差异可能抬高实施成本，需要限制首期能力边界。","tags":["交付成本","范围控制"]}
			],
			"action_plan":[
				{"order":1,"title":"完成需求访谈","detail":"访谈 10 家目标机构，确认高频问题、现有处理方式和付费意愿。","duration":"1 周"},
				{"order":2,"title":"交付试点样板","detail":"选择 3 家机构接入最小可用版本，并建立统一配置和验收清单。","duration":"2 周"},
				{"order":3,"title":"复盘试点指标","detail":"对比响应时长、有效线索率、人工投入和续费意愿，决定继续、调整或停止。","duration":"1 周"}
			],
			"growth_path":[
				{"stage":1,"title":"单场景验证","detail":"聚焦招生咨询和企微跟进，跑通 3 家付费试点。"},
				{"stage":2,"title":"门店复制","detail":"沉淀行业知识库、配置模板和交付手册，验证跨门店复制效率。"},
				{"stage":3,"title":"区域扩张","detail":"通过渠道合作拓展同类机构，并根据真实留存数据控制获客投入。"}
			],
			"validation_metrics":[
				{"label":"首次响应时间","current":"约 15 分钟","target":"1 分钟内","confidence_percent":78},
				{"label":"有效线索跟进率","current":"约 55%","target":"80% 以上","confidence_percent":68},
				{"label":"试点续费意愿","current":"尚未验证","target":"3 家中至少 2 家愿意续费","confidence_percent":55}
			],
			"timeline":[
				{"title":"客户访谈与需求确认","period":"第 1 周"},
				{"title":"最小版本配置与试点运行","period":"第 2-3 周"},
				{"title":"数据复盘与投入决策","period":"第 4 周"}
			]
		}`)
	case "learning.diagnosis":
		return []byte(`{
			"overall_score":71,
			"dimensions":[
				{"name":"目标拆解能力","score":68,"gap":18,"summary":"已说明学习目标，但仍需用具体交付物检验拆解质量。"},
				{"name":"数据洞察能力","score":62,"gap":24,"summary":"当前输入未提供可验证的数据分析作品或标准化测评结果。"},
				{"name":"执行复盘能力","score":74,"gap":14,"summary":"具备持续投入意愿，实际复盘质量仍需通过学习记录验证。"}
			],
			"recommendations":[
				"先选择一个与目标项目直接相关的课程模块并完成可交付练习。",
				"为每次练习保留输入、输出和复盘记录，再根据真实表现重新诊断。",
				"优先补齐差距最大的能力，避免同时铺开过多学习主题。"
			],
			"assumptions":[
				"用户提交的目标、项目和能力自述能够代表当前学习需求。",
				"当前没有标准化考试或外部能力认证结果可供校准。"
			]
		}`)
	case "copilot.chat":
		return []byte(`{
			"reply":"我会先基于你当前目标拆成三步：明确目标客户、验证最高频痛点、做一个低成本样板。第一步建议先访谈 10 个目标客户，记录他们现在怎么处理咨询、跟进和转化。",
			"memory_candidates":[
				{"key":"preferred_style","value":"直接给执行清单","confidence":0.82,"source":"copilot"}
			]
		}`)
	case "tasks.generate":
		return []byte(`{
			"tasks":[
				{"title":"明确目标客户范围","description":"整理行业、区域和客户规模标准，形成首批筛选条件。","project":"目标落地","priority":"high","tags":["客户验证"],"due_in_days":1,"tools":["CRM"],"learning":"客户画像"},
				{"title":"完成首批客户访谈","description":"联系10位目标客户并记录高频问题、现有解决方式和付费意愿。","project":"目标落地","priority":"high","tags":["客户验证","访谈"],"due_in_days":3,"tools":["CRM","AI助手"],"learning":"客户访谈"},
				{"title":"整理验证结论和下一步","description":"汇总访谈证据，确定保留、调整或停止的关键假设。","project":"目标落地","priority":"medium","tags":["复盘"],"due_in_days":5,"tools":["AI助手"],"learning":"需求分析"}
			]
		}`)
	default:
		return []byte(`{"ok":true}`)
	}
}

func sandboxV2RoleDevelopmentResponse(prompt string) []byte {
	var input struct {
		RoleCode   string   `json:"role_code"`
		Dimensions []string `json:"analysis_dimensions"`
	}
	if err := json.Unmarshal([]byte(prompt), &input); err != nil || input.RoleCode == "" {
		return []byte(`{"role_code":"unknown","stance":"neutral","verdict":"需要补充信息","content":"当前输入不足。","dimension_scores":[]}`)
	}
	scores := make([]map[string]any, 0, len(input.Dimensions))
	for index, dimension := range input.Dimensions {
		scores = append(scores, map[string]any{
			"code": dimension, "score": 62 + index%12, "basis": "基于当前项目输入的开发环境模拟判断",
			"confidence": 0.62, "evidence_refs": []string{},
		})
	}
	risks := []map[string]any{{"point": "缺少真实付费样本", "severity": "high", "basis": "当前结论主要来自用户输入"}}
	if input.RoleCode == "skeptic" {
		risks = append(risks,
			map[string]any{"point": "获客成本可能超过毛利", "severity": "high", "basis": "尚无真实渠道数据"},
			map[string]any{"point": "定制交付可能无法规模化", "severity": "high", "basis": "尚无标准交付样本"},
		)
	}
	response := map[string]any{
		"role_code": input.RoleCode, "stance": "neutral", "verdict": "建议先做小范围验证",
		"content":               "当前方案具备可验证的使用场景，但付费意愿、获客成本和交付边界仍需要真实样本支持。",
		"dimension_scores":      scores,
		"key_findings":          []string{"存在明确业务场景", "关键经营指标尚未验证"},
		"risks":                 risks,
		"recommendations":       []map[string]any{{"action": "完成3家付费试点", "why": "验证需求和交付成本"}},
		"questions_to_validate": []string{"客户是否愿意按目标价格持续付费？"},
		"assumptions":           []string{"用户输入的目标客户描述准确"},
		"kill_criteria":         []string{"连续10次有效访谈均无付费意愿"},
		"is_model_generated":    true,
	}
	data, _ := json.Marshal(response)
	return data
}
