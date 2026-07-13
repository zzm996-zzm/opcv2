package ai

import "context"

type DevelopmentProvider struct {
	Response []byte
}

func NewDevelopmentProvider() *DevelopmentProvider {
	return &DevelopmentProvider{}
}

func (p *DevelopmentProvider) Generate(_ context.Context, request ProviderRequest) (ProviderResponse, error) {
	response := p.responseFor(request.Feature)
	return ProviderResponse{Content: response}, nil
}

func (p *DevelopmentProvider) Stream(_ context.Context, request ProviderRequest, onDelta func([]byte) error) (ProviderResponse, error) {
	content := p.responseFor(request.Feature)
	if err := onDelta(content); err != nil {
		return ProviderResponse{}, err
	}
	return ProviderResponse{Content: content}, nil
}

func (p *DevelopmentProvider) responseFor(feature string) []byte {
	if len(p.Response) > 0 {
		return p.Response
	}
	switch feature {
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
	case "sandbox.run":
		return []byte(`{
			"score":83,
			"summary":"建议先做小范围客户验证。当前方案具备明确场景和可解释价值，但需要优先验证付费意愿、数据安全顾虑和交付成本。",
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
			"next_actions":["访谈10个目标客户，确认高频咨询问题","做出一个教培场景演示样板","定义首月试点价格和成功指标"]
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
