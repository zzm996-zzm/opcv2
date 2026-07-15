import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type ContentTool } from "../lib/contentApi";

type ToolsPageProps = {
  variant?: "library" | "all" | "recommend" | "plan" | "detail";
};

type DisplayTool = {
  slug?: string;
  name: string;
  desc: string;
  tags: readonly string[];
  price: string;
  platform: string;
  accent: string;
  category: string;
  url?: string;
  favorited?: boolean;
};

const categories = ["全部", "写作", "绘图", "视频", "营销", "办公", "自动化", "数据分析"] as const;

const hotScenarios = [
  ["内容创作加速", "提升内容生产效率", "write"],
  ["社媒增长利器", "扩大品牌影响力", "bars"],
  ["数据洞察分析", "发现增长机会", "pie"],
  ["自动化省时神器", "解放重复性工作", "robot"]
] as const;

const referenceTools: DisplayTool[] = [
  { slug: "notion-ai", name: "Notion AI", desc: "智能写作助手，帮助你快速总结、起草文档和管理知识。", tags: ["写作", "办公", "知识管理"], price: "免费试用", platform: "Web", accent: "notion", category: "写作" },
  { slug: "midjourney", name: "Midjourney", desc: "根据文本生成高质量图像，激发无限创意灵感。", tags: ["绘图", "设计", "创意"], price: "付费", platform: "Web", accent: "midjourney", category: "绘图" },
  { slug: "runway", name: "Runway", desc: "AI视频创作平台，轻松生成、编辑和特效处理视频。", tags: ["视频", "创作", "剪辑"], price: "免费试用", platform: "Web", accent: "runway", category: "视频" },
  { slug: "jasper", name: "Jasper", desc: "AI文案写作工具，打造高转化率的营销内容。", tags: ["写作", "营销", "内容生成"], price: "付费", platform: "Web", accent: "jasper", category: "营销" },
  { slug: "perplexity", name: "Perplexity", desc: "基于AI的智能搜索引擎，提供精准可靠的答案与来源。", tags: ["搜索", "研究", "信息检索"], price: "免费", platform: "Web / iOS / Android", accent: "perplexity", category: "数据分析" },
  { slug: "gamma", name: "Gamma", desc: "AI生成演示文稿和文档，快速将想法变成精美内容。", tags: ["办公", "演示", "文档"], price: "免费试用", platform: "Web", accent: "gamma", category: "办公" },
  { slug: "zapier-ai", name: "Zapier AI", desc: "自动化连接数千款应用，AI助你构建智能工作流。", tags: ["自动化", "集成", "效率提升"], price: "免费试用", platform: "Web", accent: "zapier", category: "自动化" },
  { slug: "claude", name: "Claude", desc: "强大的AI对话助手，擅长理解、分析和创作复杂内容。", tags: ["对话", "写作", "分析"], price: "付费", platform: "Web / iOS", accent: "claude", category: "写作" }
];

const referenceRecommendations = [
  { ...referenceTools[0], name: "ChatGPT", slug: "chatgpt", accent: "chatgpt", desc: "智能对话与内容创作助手", tags: ["内容创作", "文案撰写", "用户洞察"] },
  { ...referenceTools[1], name: "Canva AI", slug: "canva-ai", accent: "canva", desc: "智能设计与多媒体创作", tags: ["设计制作", "社媒运营", "品牌物料"] },
  { ...referenceTools[0], tags: ["协作管理", "内容规划", "项目执行"] }
];

const planSteps = [
  { title: "市场调研", tools: ["Perplexity"], accents: ["perplexity"], why: "实时检索全网权威信息与数据，快速了解竞品定位、趋势及用户痛点。", output: "市场洞察报告、竞品分析、用户需求总结与趋势预测。" },
  { title: "内容策划", tools: ["Notion AI", "Claude"], accents: ["notion", "claude"], why: "结构化梳理营销要点与创意，生成内容框架、脚本大纲与传播策略。", output: "内容大纲、传播矩阵、关键信息点与分组思路。" },
  { title: "短视频生成", tools: ["Runway"], accents: ["runway"], why: "AI视频生成与编辑能力强，支持文生视频、智能剪辑与特效。", output: "竖版短视频成片、字幕与配乐版本。" },
  { title: "营销文案与投放优化", tools: ["Jasper", "Gamma"], accents: ["jasper", "gamma"], why: "快速生成营销文案与广告创意，形成可直接投放的完整素材。", output: "广告文案、投放素材文档、A/B测试版本及优化建议。" }
] as const;

function toDisplayTool(tool: ContentTool): DisplayTool {
  return {
    slug: tool.slug,
    name: tool.name,
    desc: tool.description || "AI 工具能力已收录，可进入详情查看适用场景。",
    tags: tool.tags?.length ? tool.tags : [tool.category || "未分类"],
    price: tool.price_label || "价格待补充",
    platform: tool.platforms?.join(" / ") || "平台待补充",
    accent: accentForTool(tool),
    category: tool.category || "办公",
    url: tool.url
  };
}

function accentForTool(tool: Pick<ContentTool, "slug" | "name">) {
  const value = `${tool.slug} ${tool.name}`.toLowerCase();
  return ["notion", "midjourney", "runway", "jasper", "perplexity", "gamma", "zapier", "claude", "chatgpt", "canva"]
    .find((accent) => value.includes(accent)) || "notion";
}

function ToolsPage({ variant = "library" }: ToolsPageProps) {
  const withCopilot = variant !== "all";

  return (
    <V4PageShell className={`tools-shell tools-shell-${variant}`} showCopilotMini={false}>
      <section className={`toolhub-page ${withCopilot ? "with-copilot" : "wide"}`} aria-label="工具箱">
        <div className="toolhub-main">
          {variant === "recommend" && <ToolRecommendation />}
          {variant === "plan" && <ToolPlan />}
          {variant === "detail" && <ToolDetail />}
          {(variant === "library" || variant === "all") && <ToolLibrary full={variant === "all"} />}
        </div>
        {withCopilot && <ToolsCopilot variant={variant} />}
      </section>
    </V4PageShell>
  );
}

function ToolLibrary({ full }: { full: boolean }) {
  const [apiTools, setApiTools] = useState<DisplayTool[]>([]);
  const [selectedCategory, setSelectedCategory] = useState("全部");
  const [search, setSearch] = useState("");
  const [, setError] = useState("");

  useEffect(() => {
    let active = true;
    setError("");
    contentApi
      .listTools({ category: selectedCategory === "全部" ? undefined : selectedCategory, q: search, sort: "hot", limit: full ? 100 : 9 })
      .then((payload) => {
        if (active) setApiTools(payload.tools.map(toDisplayTool));
      })
      .catch((caught) => {
        if (!active) return;
        setError(apiErrorMessage(caught, "暂时无法同步最新工具，当前展示精选目录"));
        setApiTools([]);
      });
    return () => { active = false; };
  }, [full, search, selectedCategory]);

  const fallbackTools = useMemo(() => referenceTools.filter((tool) => {
    const matchesCategory = selectedCategory === "全部" || tool.category === selectedCategory || tool.tags.includes(selectedCategory);
    const query = search.trim().toLowerCase();
    return matchesCategory && (!query || `${tool.name} ${tool.desc} ${tool.tags.join(" ")}`.toLowerCase().includes(query));
  }), [search, selectedCategory]);
  const availableTools = apiTools.length > 0 ? apiTools : fallbackTools;
  const visibleTools = full ? availableTools : availableTools.slice(0, 6);
  const visibleScenarios = full ? hotScenarios : hotScenarios.slice(0, 3);

  return (
    <>
      <header className="toolhub-title">
        <h1>工具箱</h1>
        <p>浏览全市场 AI 工具，快速找到适合你业务场景的效率工具</p>
      </header>

      <section className="toolhub-search-block" aria-label="工具筛选">
        <label className="toolhub-search">
          <span aria-hidden="true" />
          <input aria-label="搜索工具" onChange={(event) => setSearch(event.target.value)} placeholder="搜索工具名，用途或标签" value={search} />
        </label>
        <div className="toolhub-filter-row">
          <div className="toolhub-tabs" role="tablist" aria-label="工具分类">
            {categories.map((category) => (
              <button aria-selected={selectedCategory === category} className={selectedCategory === category ? "active" : ""} key={category} onClick={() => setSelectedCategory(category)} role="tab" type="button">
                {category}
              </button>
            ))}
          </div>
          <div className="toolhub-selects" aria-label="筛选条件">
            {["场景", "价格", "平台", "按热度排序"].slice(0, full ? 4 : 2).map((label) => <button key={label} type="button">{label}<span aria-hidden="true">⌄</span></button>)}
          </div>
        </div>
      </section>

      <section className="toolhub-hot" aria-label="本周热门工具">
        <header>
          <h2><span aria-hidden="true">●</span> 本周热门工具</h2>
          <p>结合你的画像推荐适合营销与内容增长的工具</p>
          {full && <Link to="/tools/recommend">查看全部推荐 <span aria-hidden="true">›</span></Link>}
        </header>
        <div className="toolhub-hot-grid">
          {visibleScenarios.map(([title, desc, icon]) => (
            <article key={title}>
              <span className={`toolhub-mini-art ${icon}`} aria-hidden="true" />
              <strong>{title}</strong>
              <small>{desc}</small>
            </article>
          ))}
        </div>
      </section>

      <section className={`toolhub-grid ${full ? "full" : ""}`} aria-label="工具列表">
        {visibleTools.length === 0 ? <div className="cdk-toolhub-empty" role="status">没有匹配的工具，换个关键词或分类试试</div> : visibleTools.map((tool) => <ToolCard key={tool.slug || tool.name} tool={tool} />)}
      </section>

      <nav className="toolhub-pagination" aria-label="工具目录分页">
        <button aria-label="上一页" type="button">‹</button>
        {[1, 2, 3, 4, 5].map((page) => <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>)}
        <span>…</span><button type="button">20</button><button aria-label="下一页" type="button">›</button>
        <small>每页显示 12 条⌄</small>
      </nav>
    </>
  );
}

function ToolCard({ tool }: { tool: DisplayTool }) {
  const [favorited, setFavorited] = useState(Boolean(tool.favorited));
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function toggleFavorite() {
    if (!tool.slug || pending) return;
    setPending(true);
    setError("");
    try {
      const result = favorited ? await contentApi.unfavoriteTool(tool.slug) : await contentApi.favoriteTool(tool.slug);
      setFavorited(result.favorited);
    } catch (caught) {
      setError(apiErrorMessage(caught, "收藏失败，请稍后再试"));
    } finally {
      setPending(false);
    }
  }

  return (
    <article className="toolhub-card">
      <span className={`toolhub-logo ${tool.accent}`} aria-hidden="true" />
      <div className="toolhub-card-copy">
        <h2>{tool.name}</h2>
        <p>{tool.desc}</p>
        <div className="toolhub-tag-row">{tool.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}</div>
        {error && <small className="form-error" role="alert">{error}</small>}
      </div>
      <button className="toolhub-favorite" aria-label={`${favorited ? "取消收藏" : "收藏"}${tool.name}`} disabled={!tool.slug || pending} onClick={toggleFavorite} type="button">{favorited ? "★" : "♡"}</button>
      <footer>
        <span className={tool.price.includes("付费") ? "paid" : "free"}>{tool.price}</span>
        <small>{tool.platform}</small>
        <Link to={tool.slug ? `/tools/detail?tool=${tool.slug}` : "/tools/detail"}>查看详情 <span aria-hidden="true">›</span></Link>
      </footer>
    </article>
  );
}

function ToolRecommendation() {
  const [goal, setGoal] = useState("为智能客服 SaaS 产品制定内容营销与获客方案");
  const [scenario, setScenario] = useState("预算有限，需要文案、设计与协作工具");
  const [tools, setTools] = useState<DisplayTool[]>(referenceRecommendations);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!goal.trim() || !scenario.trim() || pending) return;
    setPending(true);
    setError("");
    try {
      const result = await contentApi.recommendTools({ goal, scenario, limit: 3 });
      setTools(result.tools.length ? result.tools.map(toDisplayTool) : referenceRecommendations);
    } catch (caught) {
      setError(apiErrorMessage(caught, "暂时无法更新匹配，当前展示示例方案"));
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="toolhub-recommend" onSubmit={submit}>
      <header className="toolhub-title">
        <h1>工具推荐结果</h1>
        <p>根据你的问题与画像，为你匹配适合的 AI 工具</p>
      </header>
      <section className="toolhub-summary">
        <strong><span aria-hidden="true">▣</span> 需求摘要</strong>
        <label>目标：<input aria-label="工具匹配目标" onChange={(event) => setGoal(event.target.value)} value={goal} /></label>
        <label>场景：<input aria-label="工具使用场景" onChange={(event) => setScenario(event.target.value)} value={scenario} /></label>
      </section>
      <section className="toolhub-recommend-results">
        <h2>为你推荐的 AI 工具 <span>（{tools.length}）</span></h2>
        <div>{tools.slice(0, 3).map((tool, index) => <RecommendationCard key={tool.slug || tool.name} index={index} tool={tool} />)}</div>
      </section>
      <section className="toolhub-advice">
        <h2><span aria-hidden="true">⌘</span> 推荐理由与使用建议</h2>
        <ul>
          <li>内容创作（ChatGPT）：用于输出 SEO 文章、推广文案、邮件/脚本，建立内容资产。</li>
          <li>视觉设计（Canva AI）：将文案快速转化为各类营销素材，提升内容传播效率。</li>
          <li>协作管理（Notion AI）：制定内容日历、管理项目进度与资产沉淀，保障执行落地。</li>
          <li>组合使用建议：ChatGPT 产出内容 → Canva AI 制作素材 → Notion AI 管理计划与复盘。</li>
        </ul>
      </section>
      {error && <p className="toolhub-sync-note" role="status">{error}</p>}
      <div className="toolhub-bottom-actions">
        <Link className="primary" to="/tools/recommendation-plan"><span aria-hidden="true">✦</span> 生成整套方案</Link>
        <button type="button"><span aria-hidden="true">♡</span> 存为收藏</button>
        <button className="toolhub-refresh" disabled={pending} type="submit">{pending ? "匹配中..." : "换一换"}</button>
      </div>
    </form>
  );
}

function RecommendationCard({ index, tool }: { index: number; tool: DisplayTool }) {
  const why = ["擅长市场调研、用户洞察、内容大纲与文案生成，帮助你快速产出高质量营销文案与方案。", "快速生成海报、社媒素材、演示文稿等视觉内容，内置模板丰富，适合低预算高效创作。", "用于方案规划、内容日历与协作管理，整合信息与任务，帮助团队高效协同落地。"][index] || tool.desc;
  return (
    <article>
      <header><span className={`toolhub-logo ${tool.accent}`} aria-hidden="true" /><div><h3>{tool.name}</h3><p>{tool.desc}</p><b>免费版可用</b></div></header>
      <strong>为什么推荐</strong><p>{why}</p>
      <strong>适用场景标签</strong><div className="toolhub-tag-row">{tool.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}</div>
      <Link to={`/tools/detail?tool=${tool.slug || ""}`}>查看详情 <span aria-hidden="true">›</span></Link>
    </article>
  );
}

function ToolPlan() {
  return (
    <>
      <nav className="toolhub-breadcrumb" aria-label="工具方案路径"><Link to="/tools">工具箱</Link><span>/</span><Link to="/tools/recommend">AI找工具</Link><span>/</span><b>整套工具方案</b></nav>
      <header className="toolhub-title compact"><h1>整套工具方案 <span>组合能力</span></h1><p>基于您的需求，智能匹配并串联最优工具组合，助力高效完成目标</p></header>
      <section className="toolhub-plan-summary">
        <article><strong><span aria-hidden="true">ϟ</span> 您的原始需求</strong><p>为一款智能手机新品制定从市场调研到视频内容生产，再到营销文案与投放优化的完整方案。</p></article>
        <article><strong><span aria-hidden="true">◎</span> 我们的目标</strong><p>用最合适的AI工具组合，完成从洞察 → 内容 → 制作 → 投放优化的全流程，提升效率并降低成本。</p></article>
      </section>
      <div className="toolhub-plan-layout">
        <section className="toolhub-flow">
          <h2>推荐执行流程 <span>（4步完成）</span></h2>
          <div>{planSteps.map((step, index) => <PlanStep index={index} key={step.title} step={step} />)}</div>
        </section>
        <aside className="toolhub-plan-aside">
          <section><h2><span aria-hidden="true">▤</span> 方案概览</h2><strong>适用对象</strong><p>消费电子 / 智能科技品牌<br />市场部 / 内容团队 / 运营团队</p><strong>预算友好度</strong><div className="budget-meter">{[0, 1, 2, 3, 4].map((item) => <span className={item > 2 ? "muted" : ""} key={item}>$</span>)}</div><strong>上手难度</strong><div className="difficulty-stars">★★★<span>★★</span></div><strong>预计周期</strong><p>3 - 5 个工作日</p></section>
          <section><h2><span aria-hidden="true">ϟ</span> 下一步操作</h2><button type="button">▮ 保存方案</button><button type="button">▣ 导出PDF</button><small>方案将保存在「任务中心」中，请排期、共建协作使用。</small></section>
        </aside>
      </div>
      <section className="toolhub-execution">
        <h2>执行建议</h2>
        <div>
          <article><h3>◷ 时间预估</h3><p>整体预计 3-5 个工作日</p><ul><li>市场调研：0.5-1 天</li><li>内容规划：1 天</li><li>短视频生成：1-2 天</li><li>文案与投放优化：0.5-1 天</li></ul></article>
          <article><h3>♧ 成本敏感度 <strong>中等</strong></h3><ul><li>整体成本可控，按需使用订阅/按量付费。</li><li>可优先使用免费额度起步。</li><li>核心环节可根据质量要求提升效率。</li></ul></article>
          <article><h3>♢ 关键注意事项</h3><ul><li>市场目标要充分表达，保证内容一致性。</li><li>准备品牌素材与核心卖点。</li><li>投放前建议进行小范围 A/B 测试。</li></ul></article>
        </div>
      </section>
    </>
  );
}

function PlanStep({ index, step }: { index: number; step: typeof planSteps[number] }) {
  return (
    <article>
      <i>{index + 1}</i><h3>{step.title}</h3><small>使用 {step.tools.join(" 及 ")}</small>
      <div className="toolhub-plan-tools">{step.tools.map((tool, toolIndex) => <div key={tool}><span className={`toolhub-logo ${step.accents[toolIndex]}`} aria-hidden="true" /><strong>{tool}</strong></div>)}</div>
      <strong>为什么用它</strong><p>{step.why}</p><strong>产出内容</strong><p>{step.output}</p><Link to="/tools/detail">查看工具详情 <span aria-hidden="true">›</span></Link>
      {index < planSteps.length - 1 && <b aria-hidden="true">→</b>}
    </article>
  );
}

function ToolDetail() {
  const location = useLocation();
  const toolSlug = new URLSearchParams(location.search).get("tool") || "";
  const [tool, setTool] = useState<ContentTool | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!toolSlug) { setTool(null); setError(""); return; }
    let active = true;
    contentApi.getTool(toolSlug).then((payload) => { if (active) setTool(payload); }).catch((caught) => {
      if (!active) return;
      setTool(null); setError(apiErrorMessage(caught, "暂时无法读取最新详情，当前展示示例内容"));
    });
    return () => { active = false; };
  }, [toolSlug]);

  const display = tool || {
    name: toolSlug && toolSlug !== "midjourney" ? "工具详情" : "Midjourney",
    description: toolSlug && toolSlug !== "midjourney" ? "该工具详情暂未同步。" : "专业 AI 图像生成工具",
    category: "绘图",
    tags: ["绘图", "设计", "创意"],
    url: "https://www.midjourney.com/",
    use_cases: ["创意构思", "概念设计", "插画制作", "品牌视觉", "游戏影视设定", "营销素材"],
    features: ["图像质量高", "风格多样", "生成速度快", "创意表现力强", "持续迭代更新"],
    limitations: ["需要学习提示词写法以获得更理想的效果", "部分复杂场景可能需要多次迭代优化"],
    platforms: ["Discord", "Web"],
    price_label: "基础计划 $10/月起，专业计划 $30/月起，企业计划 $60/月起"
  } as ContentTool;
  const title = display.name;
  const description = display.description || "暂无工具说明";
  const category = display.category || "未分类";
  const website = display.url || "";

  return (
    <>
      <nav className="toolhub-breadcrumb" aria-label="工具详情路径"><Link to="/tools">工具箱</Link><span>/</span><Link to="/tools/all">{category}</Link><span>/</span><b>{title}</b></nav>
      {error && <p className="toolhub-sync-note" role="status">{error}</p>}
      <section className="toolhub-detail-hero">
        <div className="toolhub-detail-copy">
          <span className={`toolhub-logo ${accentForTool(display)} large`} aria-hidden="true" />
          <div><h1>{title}</h1><p>{description}</p><div className="toolhub-tag-row">{(display.tags?.length ? display.tags : [category]).slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}</div></div>
          <div className="toolhub-detail-actions">{website ? <a href={website} rel="noreferrer" target="_blank"><span aria-hidden="true">↗</span> 访问官网</a> : <span>暂无官网</span>}<button type="button"><span aria-hidden="true">☆</span> 收藏工具</button><Link to="/tools/recommend"><span aria-hidden="true">⌾</span> 让智活 Copilot 评估是否适合我 <span aria-hidden="true">›</span></Link></div>
        </div>
        <figure className="toolhub-detail-visual" aria-label={`${title} 工具能力预览`}><span /><span /><span /><span /></figure>
      </section>
      <section className="toolhub-detail-panel">
        <h2>工具介绍</h2>
        <div className="toolhub-detail-grid">
          <article><h3>▣ 功能简介</h3><p>{description}。Midjourney 是一款通过自然语言描述生成高质量图像的 AI 工具，擅长艺术创作、概念设计、插画与视觉探索。</p><h3>▧ 适用场景</h3><p>{display.use_cases?.length ? display.use_cases.join("、") : "暂无适用场景说明"}</p><h3>☆ 优点</h3><p>{display.features?.length ? display.features.join("、") : "暂无功能说明"}</p><h3>♢ 注意点</h3><p>{display.limitations?.length ? display.limitations.join("；") : "暂无注意事项"}</p></article>
          <article><h3>⌂ 使用步骤</h3><p>注册/登录 → 加入 Discord → 输入提示词 → 生成图像 → 优化迭代或下载</p><h3>⌁ 入口链接</h3><p>{website ? <a href={website}>{website}</a> : "暂无入口链接"}</p><h3>♢ 价格信息</h3><p>{display.price_label || "暂无价格信息"}</p><h3>♙ 适合人群</h3><p>设计师、插画师、内容创作者、市场营销人员、产品经理、学生等。</p></article>
        </div>
      </section>
      <section className="toolhub-solve">
        <h2>用它解决什么</h2><p>Midjourney 可以帮助你快速将想法转化为高质量视觉内容，提升创作效率、激发灵感，并在各类场景中发挥重要作用。</p>
        <div>{[["创意激发", "通过关键词生成多样视觉方案，突破思维局限。"], ["概念设计", "快速生成概念图像，辅助产品、场景与角色设计探索。"], ["内容创作", "为文章、视频、社交媒体等内容创作提供高质量配图。"], ["品牌视觉", "生成品牌相关视觉素材，助力品牌视觉统一与营销传播。"], ["营销素材", "快速制作海报、广告、活动图等营销素材，提升效率。"]].map(([name, copy]) => <article key={name}><strong>{name}</strong><small>{copy}</small></article>)}</div>
      </section>
    </>
  );
}

function ToolsCopilot({ variant }: { variant: NonNullable<ToolsPageProps["variant"]> }) {
  const content = variant === "recommend" ? "收到！基于你的需求，我为你推荐了 3 款最合适的 AI 工具。" : variant === "plan" ? "好的，我已为你生成从市场调研到视频推广的整套工具方案。" : "嗨，张博！今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。";
  return (
    <aside className="learning-copilot toolhub-copilot" aria-label="智活 Copilot 工具助手">
      <header className="toolhub-ai-head"><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>你的全球 AI 助手，随时为你提供帮助</p></div><span aria-hidden="true">⚙ ⌄</span></header>
      <div className="learning-chat toolhub-chat"><article><span className="ai-avatar">A</span><div><small>嗨，张博！</small><p>{content}</p></div></article><article className="toolhub-user-bubble"><p>{variant === "plan" ? "请帮我生成一套从市场调研到视频推广的工具方案" : "帮我分析一下智能硬件赛道的市场机会和潜在关键点。"}</p></article>{variant !== "library" && <article><span className="ai-avatar">A</span><p>这套组合覆盖调研、内容、视觉与协作，并兼顾可用性与低成本。</p></article>}</div>
      <nav className="learning-copilot-actions" aria-label="工具助手快捷入口"><Link to="/analysis">♧ 分析项目机会 <span aria-hidden="true">›</span></Link><Link to="/tools/recommend">▣ 推荐工具 <span aria-hidden="true">›</span></Link><Link to="/tools/recommendation-plan">♙ 制定落地计划 <span aria-hidden="true">›</span></Link></nav>
      <MiniCopilotForm className="learning-copilot-input" inputAriaLabel="向工具箱 Copilot 提问" />
    </aside>
  );
}

export default ToolsPage;
