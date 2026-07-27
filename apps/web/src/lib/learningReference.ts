import type {
  LearningCourse,
  LearningCourseMaterial,
  LearningDiagnosis,
  LearningGaps,
  LearningPlan,
  LearningProgress,
  LearningRecommendations,
  LearningReport
} from "./learningApi";

const timestamp = "2024-05-20T10:30:00+08:00";

const courseCoverByCategory: Record<string, string> = {
  入门: "/learning/course-entry.jpg",
  实战: "/learning/course-practice.jpg",
  行业: "/learning/course-industry.jpg",
  工具: "/learning/course-tools.jpg"
};

export function learningCourseCover(category: string) {
  return courseCoverByCategory[category] ?? courseCoverByCategory.实战;
}

export const referenceCourses: LearningCourse[] = [
  [1, "ai-basics", "AI基础入门：从0到1了解AI", "快速掌握 AI 核心概念与应用场景", "入门", "入门", 18, 12400, "免费", ["AI基础", "认知入门"], ["AI基础认知", "AI能力地图与应用场景", "常用AI工具入门"]],
  [2, "prompt-engineering", "提示词工程实战", "掌握高质量提示词设计与优化技巧", "实战", "实战", 24, 8700, "会员免费", ["提示词", "实战"], ["提示词基本结构", "复杂场景提示词设计", "多轮对话策略"]],
  [3, "customer-service-cases", "智能客服应用案例解析", "打造高效客户服务，提升满意度", "行业", "实战", 20, 6300, "会员免费", ["智能客服", "案例"], ["客服场景拆解", "知识库设计", "服务效果优化"]],
  [4, "ai-market-analysis", "AI行业分析方法", "用 AI 洞察市场趋势与竞争格局", "工具", "实战", 22, 9100, "会员免费", ["行业分析", "市场洞察", "数据分析"], ["行业分析概述", "行业数据获取与清洗", "市场规模与增长趋势分析", "行业驱动因素与机会洞察", "竞争格局与头部企业分析", "用户需求与细分市场分析", "行业分析报告撰写与呈现"]],
  [5, "llm-boundaries", "大模型原理与能力边界", "理解大模型底层原理与适用边界", "入门", "进阶", 16, 5600, "免费", ["大模型", "原理"], ["模型原理", "能力边界", "风险与合规"]],
  [6, "ai-content", "AI内容创作实战", "高效生成文案、脚本、图片与内容", "实战", "实战", 22, 11300, "会员免费", ["内容创作", "实战"], ["内容策略", "AI文案", "视觉内容"]],
  [7, "ai-ecommerce", "AI赋能电商运营", "提升转化效率与用户价值", "行业", "进阶", 20, 6800, "会员免费", ["电商", "运营"], ["选品分析", "营销素材", "用户运营"]],
  [8, "python-ai", "Python + AI实战入门", "用 Python 调用 AI 服务完成自动化", "工具", "入门", 24, 7400, "会员免费", ["Python", "AI工具"], ["Python基础", "API调用", "自动化项目"]],
  [9, "data-visualization", "数据分析与可视化实战", "从数据洞察到可视化呈现", "实战", "实战", 21, 8900, "会员免费", ["数据分析", "可视化"], ["数据清洗", "洞察分析", "可视化报告"]],
  [10, "ai-finance", "AI在金融行业的应用", "风险、投研与智能服务创新", "行业", "进阶", 18, 4900, "会员免费", ["金融", "行业"], ["行业场景", "风控模型", "智能服务"]],
  [11, "automation", "AI自动化办公实战", "高效处理文档、表格与邮件", "工具", "实战", 19, 6100, "会员免费", ["自动化", "办公"], ["文档自动化", "表格处理", "工作流"]],
  [12, "rag-knowledge", "RAG知识库构建实战", "构建企业专属知识库与问答系统", "工具", "进阶", 24, 5500, "会员免费", ["RAG", "知识库"], ["知识切片", "检索增强", "效果评估"]],
  [13, "ai-office", "AI自动化办公实战", "高效处理文档、表格与邮件", "工具", "实战", 19, 6100, "会员免费", ["自动化", "办公"], ["文档自动化", "表格处理", "邮件工作流"]],
  [14, "healthcare-ai", "AI医疗行业应用案例", "影像识别、辅助诊断与管理", "行业", "进阶", 17, 4300, "会员免费", ["医疗", "行业"], ["医疗场景", "案例拆解", "合规边界"]],
  [15, "customer-insight", "用户画像与精准营销", "构建用户画像，提升营销效果", "实战", "实战", 20, 6700, "会员免费", ["用户画像", "营销"], ["标签体系", "画像构建", "精准触达"]]
].map(([id, slug, title, description, category, level, hours, learners, price_label, tags, outline]) => ({
  id: id as number,
  slug: slug as string,
  title: title as string,
  description: description as string,
  category: category as string,
  level: level as string,
  hours: hours as number,
  learners: learners as number,
  price_label: price_label as string,
  tags: tags as string[],
  outline: outline as string[],
  created_at: timestamp,
  updated_at: timestamp
}));

export function referenceCourse(slug: string) {
  return referenceCourses.find((course) => course.slug === slug) ?? referenceCourses[3];
}

export const referenceMaterials: LearningCourseMaterial[] = [
  [1, "行业生命周期分析框架.pptx", "PPTX", "/learning/materials/industry-life-cycle.pptx"],
  [2, "行业发展阶段判断案例.pdf", "PDF", "/learning/materials/industry-stage.pdf"],
  [3, "行业关键指标参考模板.xlsx", "XLSX", "/learning/materials/industry-metrics.xlsx"]
].map(([id, title, material_type, content_url], index) => ({
  id: id as number,
  course_slug: "ai-market-analysis",
  title: title as string,
  material_type: material_type as string,
  content_url: content_url as string,
  position: index + 1,
  downloadable: true,
  created_at: timestamp,
  updated_at: timestamp
}));

export const referenceProgress: LearningProgress[] = [
  [1, "ai-basics", "AI基础入门", 68, "第6章 提示词工程实战", "继续学习", "2025-05-16T10:24:00+08:00"],
  [2, "prompt-engineering", "提示词工程实战", 42, "第7章 复杂场景提示词设计", "继续学习", "2025-05-15T15:48:00+08:00"],
  [3, "customer-service-cases", "智能客服应用案例", 55, "第4章 客服话术优化", "继续学习", "2025-05-14T09:12:00+08:00"],
  [4, "ai-market-analysis", "AI行业分析方法", 30, "第5章 竞品分析模型实战", "继续学习", "2025-05-13T20:31:00+08:00"]
].map(([id, course_slug, course_title, percent, last_lesson, recommended_action, updated_at]) => ({
  id: id as number,
  user_id: 7,
  course_slug: course_slug as string,
  course_title: course_title as string,
  percent: percent as number,
  last_lesson: last_lesson as string,
  recommended_action: recommended_action as string,
  updated_at: updated_at as string
}));

const evidenceSources = [
  { type: "project", label: "项目超市：智能客服与市场分析", captured_at: timestamp },
  { type: "tasks", label: "任务中心：最近完成任务 5 条", captured_at: timestamp },
  { type: "tools", label: "工具箱：常用工具 8 个", captured_at: timestamp },
  { type: "profile", label: "用户画像：时间投入偏好", captured_at: timestamp }
];

const dimensions = [
  { name: "AI基础理解", score: 70, gap: 15, summary: "基础概念掌握较稳" },
  { name: "提示词工程实战", score: 55, gap: 23, summary: "需要补齐复杂场景设计能力" },
  { name: "行业分析方法", score: 60, gap: 18, summary: "具备框架基础，需加强实战" },
  { name: "智能客服案例拆解", score: 45, gap: 35, summary: "当前最优先补齐方向" },
  { name: "数据洞察能力", score: 58, gap: 18, summary: "数据分析与洞察能力中等" }
];

const provenance = {
  basis: "model_assessment",
  disclaimer: "诊断结果为模型评估，不代表标准化考试成绩或能力认证。",
  assumptions: ["基于当前项目、任务和工具使用情况", "目标方向为智能客服与市场分析能力提升"],
  evidence_sources: evidenceSources
};

export const referenceDiagnosis: LearningDiagnosis = {
  ...provenance,
  id: 99,
  user_id: 7,
  goal: "提升智能客服与市场分析能力",
  project: "智能客服产品升级项目",
  focus_abilities: ["智能客服案例拆解", "提示词工程实战", "数据洞察能力"],
  weekly_time: "每周 6-8 小时",
  bottleneck: "缺少系统方法与行业案例积累",
  answers: [],
  status: "completed",
  overall_score: 62,
  dimensions,
  recommendations: ["优先补齐智能客服案例拆解", "完成提示词工程实战", "建立数据洞察分析模板"],
  created_at: timestamp,
  updated_at: timestamp
};

const gapItems = [
  { name: "智能客服案例拆解", current: 45, target: 80, gap: 35, priority: "差距较大", summary: "服务流程理解与方案拆解能力短板明显", evidence: "项目超市与任务中心记录", recommended: "完成智能客服案例课程与实战任务" },
  { name: "提示词工程实战", current: 55, target: 78, gap: 23, priority: "差距较大", summary: "复杂场景提示词优化经验不足", evidence: "工具箱使用频率与任务记录", recommended: "完成提示词工程进阶课程" },
  { name: "数据洞察能力", current: 58, target: 76, gap: 18, priority: "需提升", summary: "数据清洗与洞察提炼能力需加强", evidence: "项目分析任务记录", recommended: "完成数据分析与可视化实战" }
];

export const referenceGaps: LearningGaps = {
  ...provenance,
  diagnosis_id: 99,
  goal: referenceDiagnosis.goal,
  project: referenceDiagnosis.project,
  overall_score: 62,
  gaps: gapItems,
  evidence: evidenceSources.map((item) => item.label),
  generated_at: timestamp
};

export const referenceRecommendations: LearningRecommendations = {
  ...provenance,
  diagnosis_id: 99,
  goal: referenceDiagnosis.goal,
  project: referenceDiagnosis.project,
  focus: gapItems.map((item) => ({ name: item.name, priority: item.priority, summary: item.summary })),
  recommendations: ["智能客服应用案例解析", "提示词工程实战", "AI行业分析方法", "数据分析与可视化实战"],
  methods: [
    { title: "建议每周学习节奏", value: "每周 6-8 小时", detail: "建议每周学习 2-3 次" },
    { title: "预计完成周期", value: "3-4 周", detail: "约 24-32 小时学习量" },
    { title: "建议学习顺序", value: "案例拆解 → 提示词 → 行业分析", detail: "最后完成数据洞察实战" },
    { title: "学习目标产出", value: "3 项业务成果", detail: "案例报告、分析报告与优化方案" }
  ],
  generated_at: timestamp
};

export const referencePlan: LearningPlan = {
  ...provenance,
  diagnosis_id: 99,
  title: "系统学习路径",
  description: "基于你的项目方向与能力诊断结果，为你量身定制的学习路径，助你高效掌握智能客服与市场分析相关能力。",
  recommendations: ["按能力差距顺序学习", "每周安排 2-3 次学习", "每阶段完成一个可验证成果"],
  stages: [
    { number: 1, title: "AI基础认知", status: "completed", courses: ["AI基础入门", "AI能力地图与应用场景"], duration: "4.5 小时", goal: "建立AI思维与基础能力", milestone: "完成AI基础测验" },
    { number: 2, title: "提示词与工具实操", status: "in_progress", courses: ["提示词工程实战", "AI工具实战指南"], duration: "6.5 小时", goal: "掌握提示词技巧与实战工具", milestone: "完成提示词实战任务" },
    { number: 3, title: "行业分析方法", status: "pending", courses: ["AI行业分析方法", "竞品全盘数据破解实战"], duration: "7.0 小时", goal: "学会用AI做市场与竞品分析", milestone: "提交行业分析报告" },
    { number: 4, title: "智能客服应用案例", status: "pending", courses: ["智能客服应用案例", "客服数据分析与优化"], duration: "6.0 小时", goal: "落地客服场景与优化运营", milestone: "完成客服方案优化项目" }
  ],
  items: [{ id: 1, user_id: 7, diagnosis_id: 99, stage_number: 1, title: "AI基础认知", completed: true, completed_at: timestamp, updated_at: timestamp }],
  estimated_hours: 24,
  weekly_suggestion: "每周 6-8 小时",
  generated_at: timestamp
};

export const referenceReport: LearningReport = {
  ...provenance,
  diagnosis_id: 99,
  goal: referenceDiagnosis.goal,
  project: referenceDiagnosis.project,
  overall_score: 62,
  dimensions,
  priority_gaps: gapItems,
  recommendations: referenceDiagnosis.recommendations,
  evidence: evidenceSources.map((item) => item.label),
  generated_at: timestamp
};
