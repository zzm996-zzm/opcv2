package analysis

import "context"

type DevelopmentProvider struct{}

func NewDevelopmentProvider() *DevelopmentProvider {
	return &DevelopmentProvider{}
}

func (p *DevelopmentProvider) GenerateDirection(_ context.Context, input DirectionInput) (DirectionResult, error) {
	cards := []DirectionCard{
		{
			Name:           "本地教培小班陪跑",
			Score:          91,
			Reasons:        []string{"经验匹配度高", "成都本地可先小范围验证", "5万预算足够启动第一批样板"},
			MarketEvidence: "双减后家长更关注个性化与提分陪跑，可先围绕公开家长社群和本地内容平台验证需求。",
			Difficulty:     Difficulty{Level: "中", Notes: []string{"首批客户获取", "合规宣传边界", "交付标准化"}},
			Benchmarks:     []string{"本地数学提分工作室", "小红书家长陪跑账号"},
			Actions:        []string{"写出3个目标家长画像", "发布1篇本地提分案例笔记", "联系10个已有家长资源做访谈"},
			Upsell:         "升级后可直接生成100条本地教培相关企业与社群线索。",
		},
		{
			Name:           "教师副业内容账号",
			Score:          86,
			Reasons:        []string{"启动成本低", "内容可复用经验", "适合兼职持续迭代"},
			MarketEvidence: "教育经验类内容在短视频和图文平台长期有搜索需求，适合用公开内容数据做选题验证。",
			Difficulty:     Difficulty{Level: "低", Notes: []string{"需要稳定发布", "早期变现慢"}},
			Benchmarks:     []string{"教培老师知识付费账号", "升学规划图文账号"},
			Actions:        []string{"拆解10个同类账号标题", "发布3篇经验型内容", "设计一个99元咨询入口"},
			Upsell:         "升级后可拆解对标账号内容结构，并沉淀可联系合作方。",
		},
		{
			Name:           "中小机构招生顾问",
			Score:          82,
			Reasons:        []string{"懂教培业务", "B端付费意愿更明确", "可从本地机构轻量试单"},
			MarketEvidence: "中小教培机构长期需要招生、内容和转化方案，公开企业名录可支持首批名单构建。",
			Difficulty:     Difficulty{Level: "中", Notes: []string{"需要案例背书", "销售周期略长"}},
			Benchmarks:     []string{"机构招生咨询服务", "本地教育营销代运营"},
			Actions:        []string{"整理1页招生诊断清单", "找20家本地机构做免费诊断", "沉淀3个可复用方案模板"},
			Upsell:         "升级后可批量搜索本地教培机构并加入CRM跟进。",
		},
	}
	return DirectionResult{Status: StatusCompleted, Cards: cards}, nil
}
