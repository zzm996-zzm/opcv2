UPDATE project_opportunities
SET sections = $detail$
[
  {
    "key": "path",
    "title": "成功路径",
    "body": "从一个明确行业切入，用标准化脚本服务完成首批验证，再逐步扩展为月度内容服务。",
    "items": [
      "第1-2周：选择一个熟悉行业，访谈10位潜在客户",
      "第3-4周：制作3套样板脚本并完成首个付费验证",
      "第2个月：沉淀需求表、脚本模板和交付检查表",
      "第3个月起：增加复购套餐和转介绍机制"
    ],
    "blocks": [
      {
        "type": "profile",
        "title": "项目关键指标",
        "columns": 4,
        "items": [
          {"title": "启动预算", "value": "0.8-3 万元", "detail": "轻资产启动，主要用于工具、样板与获客", "tone": "primary"},
          {"title": "预计回本周期", "value": "2-4 个月", "detail": "演示估算，需通过真实订单验证", "tone": "positive"},
          {"title": "难度", "value": "中等", "detail": "内容策划与行业理解是关键", "tone": "warning", "progress": 60},
          {"title": "适合人群", "value": "自由职业者 / 内容团队", "detail": "有内容表达能力并愿意持续服务客户", "tone": "neutral"}
        ]
      },
      {
        "type": "overview",
        "title": "1. 项目概览",
        "subtitle": "用 AI 工具配合人工创意，为企业和个人 IP 提供短视频脚本与内容策划。",
        "columns": 3,
        "items": [
          {"title": "核心价值", "value": "提效 + 稳定交付", "detail": "把选题、脚本与修改过程沉淀为可复用流程", "tone": "primary"},
          {"title": "主要变现", "value": "单条 / 月度套餐", "detail": "从单次脚本验证逐步转向持续内容服务", "tone": "positive"},
          {"title": "交付周期", "value": "1-3 天", "detail": "需求确认后完成初稿并保留人工复核", "tone": "neutral"}
        ]
      },
      {
        "type": "steps",
        "title": "2. 崛起路径：从 0 到 1 再到增长",
        "columns": 3,
        "items": [
          {"title": "起步阶段", "value": "0-1 个月", "detail": "选择细分行业，访谈客户并制作样板脚本", "meta": "目标：完成首个付费验证", "tone": "primary", "tags": ["定位", "样板"]},
          {"title": "验证阶段", "value": "1-3 个月", "detail": "沉淀需求表、脚本模板和交付检查表", "meta": "目标：形成稳定交付", "tone": "primary", "tags": ["流程", "复盘"]},
          {"title": "增长阶段", "value": "3 个月以上", "detail": "推出月度套餐并建立转介绍机制", "meta": "目标：提升复购", "tone": "positive", "tags": ["复购", "转介绍"]}
        ]
      },
      {
        "type": "cards",
        "title": "3. 真实案例",
        "columns": 2,
        "items": [
          {"title": "三个样板脚本拿到首个客户", "value": "成功案例", "detail": "聚焦餐饮行业，先展示样板再报价，完成首个付费交付。", "meta": "来源：演示案例来源", "tone": "positive", "tags": ["细分定位", "首单"]},
          {"title": "需求边界不清导致连续返工", "value": "失败教训", "detail": "未约定修改次数和验收口径，交付周期与毛利失控。", "meta": "来源：演示案例来源", "tone": "danger", "tags": ["交付边界", "返工"]}
        ]
      },
      {
        "type": "checklist",
        "title": "4. 资源能力清单",
        "columns": 4,
        "items": [
          {"title": "资金", "value": "0.8-3 万元", "detail": "软件订阅、样板制作与基础获客", "tone": "neutral", "tags": ["自有资金", "外部资金"]},
          {"title": "内容能力", "value": "策划与写作", "detail": "理解行业需求并完成脚本表达", "tone": "primary", "tags": ["已有经验", "需要提升"]},
          {"title": "剪辑 / 工具", "value": "AI 工具链", "detail": "会使用写作、素材与协作工具", "tone": "primary", "tags": ["已掌握", "继续学习"]},
          {"title": "获客渠道", "value": "样板 + 转介绍", "detail": "从熟悉行业和已有关系开始验证", "tone": "positive", "tags": ["已有渠道", "需要拓展"]}
        ]
      },
      {
        "type": "cards",
        "title": "5. 拓展玩法",
        "columns": 3,
        "items": [
          {"title": "企业品牌脚本工作坊", "value": "¥3,000-¥20,000 / 项目", "detail": "把脚本服务延伸为企业内容团队培训", "tone": "primary"},
          {"title": "个人 IP 孵化陪跑", "value": "¥2,000-¥8,000 / 月", "detail": "提供选题、脚本与复盘的周期服务", "tone": "positive"},
          {"title": "行业垂直脚本库订阅", "value": "¥299-¥999 / 月", "detail": "沉淀可持续更新的行业模板与案例", "tone": "neutral"}
        ]
      },
      {
        "type": "sources",
        "title": "6. 数据来源与出处",
        "columns": 3,
        "items": [
          {"title": "项目目录演示数据", "detail": "预算、难度与资源要求来自已发布项目目录", "meta": "项目超市", "tone": "neutral"},
          {"title": "成功与失败案例", "detail": "方法、经验与风险来自关联案例记录", "meta": "真实案例库", "tone": "neutral"},
          {"title": "用户匹配记录", "detail": "适配判断需结合用户自己的目标与资源", "meta": "AI 匹配", "tone": "neutral"}
        ]
      }
    ]
  },
  {
    "key": "data",
    "title": "当前数据",
    "body": "以下为结构化演示目录数据，用于还原页面状态，不代表真实市场预测。",
    "items": [
      "近12个月相关搜索热度演示值：+128%",
      "常见单次报价区间：800-5,000元",
      "演示平均交付周期：1-3天",
      "演示复购率：35.7%"
    ],
    "blocks": [
      {
        "type": "chart",
        "title": "1. 市场需求与搜索热度",
        "subtitle": "短视频脚本相关关键词热度（近 12 个月演示值）",
        "series": [
          {"title": "6月", "value": "18", "progress": 18},
          {"title": "7月", "value": "24", "progress": 24},
          {"title": "8月", "value": "31", "progress": 31},
          {"title": "9月", "value": "37", "progress": 37},
          {"title": "10月", "value": "32", "progress": 32},
          {"title": "11月", "value": "43", "progress": 43},
          {"title": "12月", "value": "47", "progress": 47},
          {"title": "1月", "value": "58", "progress": 58},
          {"title": "2月", "value": "72", "progress": 72},
          {"title": "3月", "value": "75", "progress": 75},
          {"title": "4月", "value": "86", "progress": 86},
          {"title": "5月", "value": "100", "meta": "+128%", "tone": "positive", "progress": 100}
        ]
      },
      {
        "type": "metrics",
        "title": "2. 客户画像与订单结构",
        "columns": 4,
        "items": [
          {"title": "企业品牌方", "value": "42%", "detail": "品牌宣传与产品种草", "tone": "primary", "progress": 42},
          {"title": "个体 IP", "value": "28%", "detail": "知识博主与个人账号", "tone": "primary", "progress": 28},
          {"title": "本地商家", "value": "18%", "detail": "探店与团购内容", "tone": "neutral", "progress": 18},
          {"title": "MCN / 代运营", "value": "12%", "detail": "批量脚本与矩阵支持", "tone": "neutral", "progress": 12}
        ]
      },
      {
        "type": "metrics",
        "title": "3. 报价与交付效率",
        "columns": 4,
        "items": [
          {"title": "常见报价区间", "value": "¥800-¥5,000 / 条", "detail": "随行业、复杂度和服务范围变化"},
          {"title": "平均交付周期", "value": "1-3 天", "detail": "演示口径，不含客户确认时间"},
          {"title": "平均修改次数", "value": "1.6 次", "detail": "明确需求表后可减少返工"},
          {"title": "平均毛利率", "value": "68%", "detail": "未计入获客与管理成本", "tone": "positive", "progress": 68}
        ]
      },
      {
        "type": "channels",
        "title": "4. 渠道表现",
        "subtitle": "询盘占比演示值",
        "columns": 4,
        "items": [
          {"title": "小红书", "value": "38%", "progress": 38},
          {"title": "抖音", "value": "32%", "progress": 32},
          {"title": "视频号", "value": "18%", "progress": 18},
          {"title": "私域转介绍", "value": "12%", "progress": 12}
        ]
      },
      {
        "type": "funnel",
        "title": "5. 转化漏斗",
        "columns": 3,
        "items": [
          {"title": "咨询人数", "value": "1,248", "detail": "进入需求沟通", "progress": 100},
          {"title": "脚本下单数", "value": "286", "meta": "转化率 22.9%", "progress": 22.9},
          {"title": "复购客户数", "value": "102", "meta": "复购率 35.7%", "progress": 8.2}
        ]
      },
      {
        "type": "metrics",
        "title": "6. 本月关键指标",
        "columns": 5,
        "items": [
          {"title": "订单数", "value": "286", "meta": "环比 +18.7%", "tone": "positive"},
          {"title": "询盘转化率", "value": "22.9%", "meta": "环比 +8.6%", "tone": "positive"},
          {"title": "复购率", "value": "35.7%", "meta": "环比 +6.2%", "tone": "positive"},
          {"title": "客户满意度", "value": "4.8 / 5", "meta": "环比 +0.3", "tone": "positive"},
          {"title": "平均客单价", "value": "¥2,180", "meta": "环比 +11.4%", "tone": "positive"}
        ]
      }
    ]
  },
  {
    "key": "swot",
    "title": "优劣势",
    "body": "该项目启动成本低、交付快，但需要持续建立差异化和客户信任。",
    "items": [
      "优势：轻资产、标准化空间大、可按月复购",
      "优势：AI工具可明显缩短初稿时间",
      "短板：同质化竞争明显，案例质量决定成交率",
      "短板：需求边界不清时容易频繁返工"
    ],
    "blocks": [
      {
        "type": "cards",
        "title": "核心优势",
        "columns": 3,
        "items": [
          {"title": "交付快", "detail": "单条脚本通常 1-3 天可交付", "tone": "positive"},
          {"title": "低启动成本", "detail": "主要投入为工具订阅与样板制作", "tone": "positive"},
          {"title": "适合轻资产创业", "detail": "个人或小团队可逐步验证", "tone": "positive"},
          {"title": "可标准化服务", "detail": "需求表、模板和检查表可复用", "tone": "positive"},
          {"title": "内容生产效率高", "detail": "AI 可辅助选题与初稿生成", "tone": "positive"},
          {"title": "易做案例积累", "detail": "样板与前后对比便于展示价值", "tone": "positive"}
        ]
      },
      {
        "type": "cards",
        "title": "主要短板",
        "columns": 3,
        "items": [
          {"title": "审美与表达门槛", "detail": "AI 初稿仍需人工判断和重写", "tone": "danger"},
          {"title": "客户信任建立慢", "detail": "需要持续积累行业案例", "tone": "danger"},
          {"title": "前期获客不稳定", "detail": "没有渠道时线索波动较大", "tone": "danger"},
          {"title": "同质化竞争", "detail": "低价与模板化内容较多", "tone": "danger"},
          {"title": "修改返工风险", "detail": "边界不清会侵蚀项目毛利", "tone": "danger"},
          {"title": "规模化管理要求", "detail": "扩团队后需明确 SOP 与质检", "tone": "danger"}
        ]
      },
      {
        "type": "cards",
        "title": "适合与不适合的人",
        "columns": 2,
        "items": [
          {"title": "适合的人", "detail": "对内容创作敏感、能持续学习 AI 工具、具备沟通与服务意识", "tone": "positive", "tags": ["内容创作", "客户沟通", "持续学习"]},
          {"title": "不适合的人", "detail": "期待短期暴利、不愿投入沟通、不能承受前期获客波动", "tone": "danger", "tags": ["短期暴利", "拒绝复盘", "零沟通"]}
        ]
      },
      {
        "type": "matrix",
        "title": "优劣势矩阵（机会-风险）",
        "columns": 2,
        "items": [
          {"title": "机会区", "detail": "AI 提效、短视频需求持续增长、行业垂直化", "tone": "positive"},
          {"title": "警惕区", "detail": "同质化竞争、平台规则变化、口碑积累慢", "tone": "warning"},
          {"title": "优化区", "detail": "完善流程、明确边界、加强团队协作", "tone": "primary"},
          {"title": "风险区", "detail": "无限修改、低价竞争、需求定位模糊", "tone": "danger"}
        ]
      },
      {
        "type": "overview",
        "title": "结论建议",
        "items": [
          {"title": "总体判断", "value": "轻资产、低门槛、需求强", "detail": "适合个人或小团队起步，但需要依靠差异化与标准化建立长期优势。", "tone": "primary"},
          {"title": "关键门槛能力", "detail": "内容策划、脚本结构、行业理解、客户沟通与持续获客", "tags": ["内容策划", "行业理解", "沟通", "获客"]}
        ]
      }
    ]
  },
  {
    "key": "learning",
    "title": "可学经验",
    "body": "先用小范围真实交付验证方法，再把有效动作固化成模板。",
    "items": [
      "先聚焦一个细分行业，不做全行业通用服务",
      "建立标准需求表和三档服务包",
      "每次交付记录修改原因并更新检查表",
      "用案例前后对比展示价值而非承诺流量"
    ],
    "blocks": [
      {
        "type": "steps",
        "title": "1. 可复制的方法论",
        "columns": 4,
        "items": [
          {"title": "明确细分定位", "detail": "聚焦 1-2 种内容场景和熟悉行业", "tone": "primary"},
          {"title": "建立标准服务包", "detail": "固定脚本类型、交付物与修改次数", "tone": "primary"},
          {"title": "用案例驱动成交", "detail": "用真实样板展示价值和边界", "tone": "positive"},
          {"title": "用流程提升效率", "detail": "用 SOP、提示词库和工具链减少返工", "tone": "positive"}
        ]
      },
      {
        "type": "cards",
        "title": "2. 关键经验总结",
        "columns": 5,
        "items": [
          {"title": "先做小而明确的场景", "detail": "避免一开始覆盖全部行业"},
          {"title": "用样稿降低成交门槛", "detail": "先让客户看到可验证的质量"},
          {"title": "用复盘沉淀模板", "detail": "每次修改都更新需求表和检查表"},
          {"title": "通过反馈优化沟通", "detail": "把问题转化为标准问法"},
          {"title": "建立内容素材库", "detail": "积累选题、金句和结构模板"}
        ]
      },
      {
        "type": "cards",
        "title": "3. 从案例中学什么",
        "columns": 3,
        "items": [
          {"title": "个人 IP 拍摄", "value": "案例 1", "detail": "用样板脚本降低首次合作决策成本", "tone": "positive"},
          {"title": "产品测评单", "value": "案例 2", "detail": "固定交付结构后减少反复修改", "tone": "positive"},
          {"title": "知识付费类", "value": "案例 3", "detail": "持续积累素材库后提升复购", "tone": "positive"}
        ]
      },
      {
        "type": "actions",
        "title": "4. 你现在就能推进的动作",
        "columns": 2,
        "items": [
          {"title": "7 天行动清单", "detail": "选择一个细分行业；访谈 3 位潜在客户；完成 3 套样板；联系首批客户", "meta": "7 DAY", "tone": "primary", "tags": ["定位", "访谈", "样板", "触达"]},
          {"title": "30 天行动清单", "detail": "完成 2-3 个项目；优化服务包；持续输出案例；建立获客渠道", "meta": "30 DAY", "tone": "positive", "tags": ["交付", "服务包", "案例", "渠道"]}
        ]
      },
      {
        "type": "cards",
        "title": "5. 适合迁移到哪些项目",
        "columns": 4,
        "items": [
          {"title": "企业知识 IP", "detail": "帮助企业打造创始人或员工 IP"},
          {"title": "短视频代运营", "detail": "从脚本延伸到拍摄、剪辑和发布"},
          {"title": "脚本咨询", "detail": "为品牌提供团队培训与策略咨询"},
          {"title": "内容策划服务", "detail": "扩展到公众号、直播和种草笔记"}
        ]
      }
    ]
  },
  {
    "key": "avoid",
    "title": "要避免的行为",
    "body": "主要风险来自定位模糊、低价竞争和缺少交付边界。",
    "items": [
      "一开始覆盖过多行业和平台",
      "只用低价吸引客户，忽略服务边界",
      "没有书面确认需求和修改次数",
      "把演示指标当作真实收益承诺"
    ],
    "blocks": [
      {
        "type": "overview",
        "title": "1. 风险总览",
        "columns": 4,
        "items": [
          {"title": "高频踩坑", "value": "6 大常见误区", "tone": "danger"},
          {"title": "业务风险", "value": "3 类典型场景", "tone": "danger"},
          {"title": "影响严重", "value": "可能导致交付失败", "tone": "warning"},
          {"title": "可控可防", "value": "提前建立流程与边界", "tone": "positive"}
        ]
      },
      {
        "type": "cards",
        "title": "2. 常见误区",
        "columns": 6,
        "items": [
          {"title": "一开始什么都接", "detail": "没有聚焦行业和场景", "tone": "danger"},
          {"title": "只拼低价", "detail": "价格低但没有服务边界", "tone": "danger"},
          {"title": "没有标准交付流程", "detail": "文档、验收和修改次数不明确", "tone": "danger"},
          {"title": "忽视客户反馈", "detail": "不复盘导致问题重复发生", "tone": "danger"},
          {"title": "过度承诺效果", "detail": "承诺无法控制的流量结果", "tone": "danger"},
          {"title": "案例和定位不清晰", "detail": "客户无法快速建立信任", "tone": "danger"}
        ]
      },
      {
        "type": "cards",
        "title": "3. 风险场景",
        "columns": 3,
        "items": [
          {"title": "返工过多", "detail": "需求不清、脚本质量不足，交付周期被拉长", "meta": "影响：利润与满意度下降", "tone": "danger"},
          {"title": "需求失控", "detail": "范围无边界，客户不断增加要求", "meta": "影响：团队效率降低", "tone": "warning"},
          {"title": "获客渠道单一", "detail": "依赖单一平台或关系渠道", "meta": "影响：收入波动", "tone": "warning"}
        ]
      },
      {
        "type": "checklist",
        "title": "4. 避坑建议",
        "columns": 5,
        "items": [
          {"title": "先限定服务边界", "detail": "在合同中写清交付范围与修改次数", "tone": "positive"},
          {"title": "统一报价模板", "detail": "按复杂度和服务范围分档", "tone": "positive"},
          {"title": "建立交付清单", "detail": "明确 SOP、文件格式和验收口径", "tone": "positive"},
          {"title": "用试单验证客户", "detail": "小单验证需求真实性与配合度", "tone": "positive"},
          {"title": "沉淀案例资产", "detail": "用真实前后对比建立长期信任", "tone": "positive"}
        ]
      },
      {
        "type": "channels",
        "title": "5. 风险优先级",
        "columns": 1,
        "items": [
          {"title": "需求失控 / 范围无限", "value": "高概率 · 高影响", "tone": "danger", "progress": 95},
          {"title": "返工过多 / 质量不稳定", "value": "高概率 · 中影响", "tone": "danger", "progress": 82},
          {"title": "只拼低价 / 利润被压缩", "value": "中概率 · 中影响", "tone": "warning", "progress": 64},
          {"title": "获客渠道单一", "value": "中概率 · 中影响", "tone": "warning", "progress": 58},
          {"title": "案例与定位不清晰", "value": "低概率 · 中影响", "tone": "neutral", "progress": 38}
        ]
      },
      {
        "type": "cards",
        "title": "6. 建议先建立的底层能力",
        "columns": 4,
        "items": [
          {"title": "沟通能力", "detail": "确认需求、澄清边界并减少返工"},
          {"title": "项目管理", "detail": "进度把控、协同与按时交付"},
          {"title": "内容审核", "detail": "逻辑结构、表达和质量把控"},
          {"title": "基础数据复盘", "detail": "复盘关键指标并持续优化"}
        ]
      }
    ]
  }
]
$detail$::jsonb,
    updated_at = NOW()
WHERE slug = 'ai-short-video-studio';

UPDATE project_opportunities
SET sections = jsonb_build_array(
        jsonb_build_object(
            'key', 'path',
            'title', '成功路径',
            'body', COALESCE(sections -> 0 ->> 'body', summary),
            'items', COALESCE(sections -> 0 -> 'items', '[]'::jsonb),
            'blocks', jsonb_build_array(
                jsonb_build_object(
                    'type', 'profile',
                    'title', '项目关键指标',
                    'columns', 4,
                    'items', jsonb_build_array(
                        jsonb_build_object('title', '启动预算', 'value', budget_band, 'detail', '来自已发布项目目录', 'tone', 'primary'),
                        jsonb_build_object('title', '预计回本周期', 'value', '待真实订单验证', 'detail', '目录暂未提供可验证回本数据', 'tone', 'neutral'),
                        jsonb_build_object('title', '难度', 'value', difficulty, 'detail', '结合资源要求综合判断', 'tone', 'warning'),
                        jsonb_build_object('title', '适合人群', 'value', industry, 'detail', summary, 'tone', 'neutral', 'tags', tags)
                    )
                ),
                jsonb_build_object(
                    'type', 'steps',
                    'title', '最小验证路径',
                    'columns', 3,
                    'items', COALESCE((
                        SELECT jsonb_agg(jsonb_build_object('title', step, 'tone', 'primary'))
                        FROM jsonb_array_elements_text(COALESCE(sections -> 0 -> 'items', '[]'::jsonb)) AS source(step)
                    ), '[]'::jsonb)
                )
            )
        ),
        jsonb_build_object(
            'key', 'data',
            'title', '当前数据',
            'body', '当前仅展示已发布目录中的可验证字段，更多数据需通过真实交付持续补充。',
            'items', jsonb_build_array(budget_band, difficulty, industry),
            'blocks', jsonb_build_array(jsonb_build_object(
                'type', 'metrics',
                'title', '目录基础指标',
                'columns', 3,
                'items', jsonb_build_array(
                    jsonb_build_object('title', '启动预算', 'value', budget_band, 'tone', 'neutral'),
                    jsonb_build_object('title', '项目难度', 'value', difficulty, 'tone', 'neutral'),
                    jsonb_build_object('title', '所属行业', 'value', industry, 'tone', 'primary')
                )
            ))
        ),
        jsonb_build_object(
            'key', 'swot',
            'title', '优劣势',
            'body', '结合项目简介、资源要求和真实案例核对优势、短板与适配边界。',
            'items', jsonb_build_array(summary),
            'blocks', jsonb_build_array(jsonb_build_object(
                'type', 'matrix',
                'title', '基础判断',
                'columns', 2,
                'items', jsonb_build_array(
                    jsonb_build_object('title', '可利用优势', 'detail', summary, 'tone', 'positive'),
                    jsonb_build_object('title', '启动门槛', 'value', jsonb_array_length(resource_requirements) || ' 项资源要求', 'detail', '进入项目前逐项确认资源可获得性。', 'tone', 'warning')
                )
            ))
        ),
        jsonb_build_object(
            'key', 'learning',
            'title', '可学经验',
            'body', '从最小可交付版本开始，用真实反馈沉淀模板、流程和复盘方法。',
            'items', COALESCE(sections -> 0 -> 'items', '[]'::jsonb),
            'blocks', jsonb_build_array(jsonb_build_object(
                'type', 'actions',
                'title', '可立即执行的学习动作',
                'items', jsonb_build_array(jsonb_build_object(
                    'title', '完成一次最小验证',
                    'detail', COALESCE(sections -> 0 ->> 'body', summary),
                    'tone', 'primary',
                    'tags', tags
                ))
            ))
        ),
        jsonb_build_object(
            'key', 'avoid',
            'title', '要避免的行为',
            'body', '不要把目录演示信息当作收益承诺；启动前需确认预算、资源、交付边界和证据来源。',
            'items', jsonb_build_array('忽略资源门槛', '未经验证承诺结果', '缺少书面交付边界'),
            'blocks', jsonb_build_array(jsonb_build_object(
                'type', 'checklist',
                'title', '启动前避坑清单',
                'columns', 3,
                'items', jsonb_build_array(
                    jsonb_build_object('title', '核对预算', 'value', budget_band, 'tone', 'warning'),
                    jsonb_build_object('title', '核对资源', 'value', jsonb_array_length(resource_requirements) || ' 项', 'tone', 'warning'),
                    jsonb_build_object('title', '核对来源', 'detail', '优先查看关联案例与公开出处。', 'tone', 'positive')
                )
            ))
        )
    ),
    updated_at = NOW()
WHERE slug IN (
    'excel-automation-reporting',
    'ai-resume-optimization',
    'local-pet-care-subscription',
    'knowledge-course-production',
    'local-ai-sales-consulting',
    'niche-travel-guide',
    'data-dashboard-service'
);
