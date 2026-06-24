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

func (p *DevelopmentProvider) responseFor(feature string) []byte {
	if len(p.Response) > 0 {
		return p.Response
	}
	switch feature {
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
	default:
		return []byte(`{"ok":true}`)
	}
}
