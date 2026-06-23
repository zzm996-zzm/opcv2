import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

type ToolsPageProps = {
  variant?: "library" | "all" | "recommend" | "plan" | "detail";
};

const categories = ["全部", "写作", "绘图", "视频", "营销", "办公", "自动化", "数据分析"] as const;

const hotScenarios = [
  ["内容创作加速", "提升内容生产效率", "write"],
  ["社媒增长利器", "扩大品牌影响力", "bars"],
  ["数据洞察分析", "发现增长机会", "pie"],
  ["自动化省时神器", "解放重复性工作", "robot"]
] as const;

const tools = [
  {
    name: "Notion AI",
    desc: "智能写作助手，帮助你快速总结、起草文档和管理知识。",
    tags: ["写作", "办公", "知识管理"],
    price: "免费试用",
    platform: "Web",
    accent: "notion",
    category: "办公"
  },
  {
    name: "Midjourney",
    desc: "根据文本生成高质量图像，激发无限创意灵感。",
    tags: ["绘图", "设计", "创意"],
    price: "付费",
    platform: "Web",
    accent: "midjourney",
    category: "绘图"
  },
  {
    name: "Runway",
    desc: "AI视频创作平台，轻松生成、编辑和特效处理视频。",
    tags: ["视频", "创作", "剪辑"],
    price: "免费试用",
    platform: "Web",
    accent: "runway",
    category: "视频"
  },
  {
    name: "Jasper",
    desc: "AI文案写作工具，打造高转化率的营销内容。",
    tags: ["写作", "营销", "内容生成"],
    price: "付费",
    platform: "Web",
    accent: "jasper",
    category: "营销"
  },
  {
    name: "Perplexity",
    desc: "基于AI的智能搜索引擎，提供精准可靠的答案与来源。",
    tags: ["搜索", "研究", "信息检索"],
    price: "免费",
    platform: "Web / iOS / Android",
    accent: "perplexity",
    category: "数据分析"
  },
  {
    name: "Gamma",
    desc: "AI生成演示文稿和文档，快速将想法变成精美内容。",
    tags: ["办公", "演示", "文档"],
    price: "免费试用",
    platform: "Web",
    accent: "gamma",
    category: "办公"
  },
  {
    name: "Zapier AI",
    desc: "自动化连接数千款应用，AI助你构建智能工作流。",
    tags: ["自动化", "集成", "效率提升"],
    price: "免费试用",
    platform: "Web",
    accent: "zapier",
    category: "自动化"
  },
  {
    name: "Claude",
    desc: "强大的AI对话助手，擅长理解、分析和创作复杂内容。",
    tags: ["对话", "写作", "分析"],
    price: "付费",
    platform: "Web / iOS",
    accent: "claude",
    category: "写作"
  }
] as const;

const recommendedTools = [
  {
    name: "ChatGPT",
    desc: "智能对话与内容创作助手",
    tags: ["内容创作", "文案撰写", "用户洞察"],
    reason: "擅长市场调研、用户洞察、内容大纲与文案生成，帮助你快速产出高质量营销文案与方案。",
    accent: "chatgpt"
  },
  {
    name: "Canva AI",
    desc: "智能设计与多媒体创作",
    tags: ["设计制作", "社媒运营", "品牌物料"],
    reason: "快速生成海报、社媒素材、演示文稿等视觉内容，内置模板丰富，适合低预算高效制作。",
    accent: "canva"
  },
  {
    name: "Notion AI",
    desc: "智能笔记与知识管理助手",
    tags: ["协作管理", "内容规划", "项目执行"],
    reason: "用于方案规划、内容日历与协作管理，整合信息与任务，帮助团队高效协同落地。",
    accent: "notion"
  }
] as const;

const planSteps = [
  {
    step: "1",
    title: "市场调研",
    tool: "Perplexity",
    why: "实时检索全网权威信息与数据，快速了解新能源市场规模、趋势、竞品与用户痛点。",
    output: "市场洞察报告、竞品分析、用户需求总结与趋势预测。",
    accent: "perplexity"
  },
  {
    step: "2",
    title: "内容策划",
    tool: "Notion AI 或 Claude",
    why: "结构化梳理目标受众与卖点，生成内容框架、脚本大纲与传播策略。",
    output: "短视频脚本、内容大纲、传播策略、关键信息点与分镜脚本。",
    accent: "notion"
  },
  {
    step: "3",
    title: "短视频生成",
    tool: "Runway",
    why: "AI视频生成与编辑能力强，支持文生视频、智能剪辑与特效，快速产出高质量短视频。",
    output: "完整短视频成片、字幕与配乐版本。",
    accent: "runway"
  },
  {
    step: "4",
    title: "营销文案与投放优化",
    tool: "Jasper 或 Gamma",
    why: "生成高转化营销文案与广告创意，沉淀投放报告与策略展示。",
    output: "广告文案、投放素材文案、A/B测试版本、数据可视化报告。",
    accent: "jasper"
  }
] as const;

const adviceCards = [
  ["时间预估", "整体预计 3-5 个工作日", ["市场调研：0.5-1 天", "内容策划：1 天", "短视频生成：1-2 天", "文案与投放优化：0.5-1 天"]],
  ["成本敏感度", "中等", ["整体成本可控，按需使用订阅/按量付费。", "可优先使用免费额度启动。", "组合使用可显著降低人力时间成本。"]],
  ["关键注意事项", "执行前确认", ["明确目标受众与核心卖点。", "准备品牌素材和参数说明。", "投放前建议进行小范围 A/B 测试。"]]
] as const;

function ToolsPage({ variant = "library" }: ToolsPageProps) {
  const withCopilot = variant !== "all";

  return (
    <V4PageShell>
      <section className={`toolhub-page ${withCopilot ? "with-copilot" : "wide"}`} aria-label="工具箱">
        <main className="toolhub-main">
          {variant === "recommend" && <ToolRecommendation />}
          {variant === "plan" && <ToolPlan />}
          {variant === "detail" && <ToolDetail />}
          {(variant === "library" || variant === "all") && <ToolLibrary full={variant === "all"} />}
        </main>
        {withCopilot && <ToolsCopilot variant={variant} />}
      </section>
    </V4PageShell>
  );
}

function ToolLibrary({ full }: { full: boolean }) {
  const visibleTools = full ? tools : tools.slice(0, 6);
  const visibleScenarios = full ? hotScenarios : hotScenarios.slice(0, 3);

  return (
    <>
      <header className="toolhub-title">
        <h1>工具箱</h1>
        <p>浏览全市场 AI 工具，快速找到适合你业务场景的效率工具</p>
      </header>

      <section className="toolhub-search-block" aria-label="工具筛选">
        <label className="toolhub-search">
          <span aria-hidden="true">⌕</span>
          <input aria-label="搜索工具" placeholder="搜索工具名，用途或标签" />
        </label>
        <div className="toolhub-filter-row">
          <div className="toolhub-tabs" role="tablist" aria-label="工具分类">
            {categories.map((category, index) => (
              <button className={index === 0 ? "active" : ""} key={category} role="tab" type="button">
                {category}
              </button>
            ))}
          </div>
          <div className="toolhub-selects" aria-label="筛选条件">
            {["场景", "价格", "平台", "按热度排序"].slice(0, full ? 4 : 2).map((label) => (
              <button key={label} type="button">{label}⌄</button>
            ))}
          </div>
        </div>
      </section>

      <section className="toolhub-hot" aria-label="本周热门工具">
        <header>
          <h2>🔥 本周热门工具</h2>
          <p>结合你的画像推荐适合营销与内容增长的工具</p>
          {full && <Link to="/tools/recommend">查看全部推荐 ›</Link>}
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
        {visibleTools.map((tool) => <ToolCard key={tool.name} tool={tool} />)}
      </section>

      <ToolPagination />
    </>
  );
}

function ToolCard({ tool }: { tool: (typeof tools)[number] }) {
  return (
    <article className="toolhub-card">
      <header>
        <span className={`toolhub-logo ${tool.accent}`} aria-hidden="true" />
        <button aria-label={`收藏${tool.name}`} type="button">♡</button>
      </header>
      <h2>{tool.name}</h2>
      <p>{tool.desc}</p>
      <div className="toolhub-tag-row">
        {tool.tags.map((tag) => <span key={tag}>{tag}</span>)}
      </div>
      <footer>
        <span className={tool.price.includes("付费") ? "paid" : "free"}>{tool.price}</span>
        <small>{tool.platform}</small>
        <Link to="/tools/detail">查看详情 ›</Link>
      </footer>
    </article>
  );
}

function ToolRecommendation() {
  return (
    <>
      <header className="toolhub-title">
        <h1>工具推荐结果</h1>
        <p>根据你的问题与画像，为你匹配适合的 AI 工具</p>
      </header>
      <section className="toolhub-summary">
        <h2>需求摘要</h2>
        <p>目标：为智能客服SaaS产品制定内容营销与获客方案；预算有限；需要文案、设计与协作工具</p>
      </section>
      <section className="toolhub-recommend">
        <h2>为你推荐的 AI 工具（3）</h2>
        <div>
          {recommendedTools.map((tool) => (
            <article key={tool.name}>
              <span className={`toolhub-logo ${tool.accent}`} aria-hidden="true" />
              <h3>{tool.name}</h3>
              <p>{tool.desc}</p>
              <b>免费版可用</b>
              <strong>为什么推荐</strong>
              <small>{tool.reason}</small>
              <strong>适用场景标签</strong>
              <div className="toolhub-tag-row">
                {tool.tags.map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <Link to="/tools/detail">查看详情 ›</Link>
            </article>
          ))}
        </div>
      </section>
      <section className="toolhub-advice">
        <h2>推荐理由与使用建议</h2>
        <ul>
          <li>内容创作（ChatGPT）：用于输出 SEO 文章、推广文案、邮件/脚本，建立内容资产。</li>
          <li>视觉设计（Canva AI）：将文案快速转化为各类营销素材，提升内容传播效率。</li>
          <li>协作管理（Notion AI）：制定内容日历、管理项目进度与资产沉淀，保障执行落地。</li>
          <li>组合使用建议：ChatGPT 产出内容 → Canva AI 制作素材 → Notion AI 管理计划与复盘。</li>
        </ul>
      </section>
      <footer className="toolhub-bottom-actions">
        <Link className="primary" to="/tools/recommendation-plan">生成整套方案</Link>
        <button type="button">存为收藏</button>
      </footer>
    </>
  );
}

function ToolPlan() {
  return (
    <>
      <nav className="toolhub-breadcrumb" aria-label="工具方案路径">
        <Link to="/tools">工具箱</Link><span>/</span><Link to="/tools/recommend">AI找工具</Link><span>/</span><b>整套工具方案</b>
      </nav>
      <header className="toolhub-title compact">
        <h1>整套工具方案 <span>组合能力</span></h1>
        <p>基于您的需求，智能匹配并串联最优工具组合，助力高效完成目标</p>
      </header>
      <section className="toolhub-plan-summary">
        <article><strong>您的原始需求</strong><p>为一款智能手机防跟踪新品制定从市场调研到视频内容生产，再到营销文案与投放优化的完整方案。</p></article>
        <article><strong>我们的目标</strong><p>用最合适的AI工具组合，完成从洞察 → 内容 → 制作 → 投放优化的全流程，提升效率并降低成本。</p></article>
      </section>
      <section className="toolhub-plan-layout">
        <div className="toolhub-flow">
          <h2>推荐执行流程 <span>4步完成</span></h2>
          <div>
            {planSteps.map((item, index) => (
              <article key={item.step}>
                <i>{item.step}</i>
                <h3>{item.title}</h3>
                <small>使用 {item.tool}</small>
                <span className={`toolhub-logo ${item.accent}`} aria-hidden="true" />
                <strong>为什么用它</strong>
                <p>{item.why}</p>
                <strong>产出内容</strong>
                <p>{item.output}</p>
                <Link to="/tools/detail">查看工具详情 ›</Link>
                {index < planSteps.length - 1 && <b aria-hidden="true">→</b>}
              </article>
            ))}
          </div>
        </div>
        <aside className="toolhub-plan-aside">
          <section>
            <h2>方案概览</h2>
            <p>适用对象：消费电子 / 智能硬件品牌、市场部、内容团队、运营团队</p>
            <div className="budget-meter" aria-label="预算友好度"><span /><span /><span /><span /><span className="muted" /></div>
            <p>上手难度：★★★☆☆</p>
            <p>预计周期：3-5 个工作日</p>
          </section>
          <section>
            <h2>下一步操作</h2>
            <button type="button">保存方案</button>
            <button type="button">导出PDF</button>
            <small>保存后可在「任务中心」中查看，并随时使用与执行。</small>
          </section>
        </aside>
      </section>
      <section className="toolhub-execution">
        <h2>执行建议</h2>
        <div>
          {adviceCards.map(([title, badge, points]) => (
            <article key={title}>
              <h3>{title}</h3>
              <strong>{badge}</strong>
              <ul>{points.map((point) => <li key={point}>{point}</li>)}</ul>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

function ToolDetail() {
  return (
    <>
      <nav className="toolhub-breadcrumb" aria-label="工具详情路径">
        <Link to="/tools">工具箱</Link><span>/</span><Link to="/tools/all">绘图</Link><span>/</span><b>Midjourney</b>
      </nav>
      <section className="toolhub-detail-hero">
        <div className="toolhub-detail-copy">
          <span className="toolhub-logo midjourney large" aria-hidden="true" />
          <div>
            <h1>Midjourney</h1>
            <p>专业 AI 图像生成工具</p>
            <div className="toolhub-tag-row">
              {["绘图", "设计", "创意"].map((tag) => <span key={tag}>{tag}</span>)}
            </div>
          </div>
          <div className="toolhub-detail-actions">
            <button type="button">访问官网</button>
            <button type="button">收藏工具</button>
            <Link to="/tools/recommend">让智活 Copilot 评估是否适合我 ›</Link>
          </div>
        </div>
        <figure className="toolhub-detail-visual" aria-label="Midjourney 生成图像预览">
          <span /><span /><span /><span />
        </figure>
      </section>
      <section className="toolhub-detail-panel">
        <h2>工具介绍</h2>
        <div className="toolhub-detail-grid">
          <article>
            <h3>功能简介</h3>
            <p>Midjourney 是一款通过自然语言描述生成高质量图像的 AI 工具，擅长艺术创作、概念设计、插画与视觉探索。</p>
            <h3>适用场景</h3>
            <p>创意构思、概念设计、插画制作、品牌视觉、游戏/影视设定、营销素材等。</p>
            <h3>优点</h3>
            <p>图像质量高、风格多样、生成速度快，创意表现力强，持续迭代更新。</p>
            <h3>注意点</h3>
            <p>需要学习提示词写法以获得更理想的效果；部分复杂场景可能需要多次迭代优化。</p>
          </article>
          <article>
            <h3>使用步骤</h3>
            <p>注册/登录 → 加入 Discord → 输入提示词 → 生成图像 → 优化迭代或下载。</p>
            <h3>入口链接</h3>
            <p><a href="https://www.midjourney.com/">https://www.midjourney.com/</a></p>
            <h3>价格信息</h3>
            <p>订阅制：基础计划 $10/月起，专业计划 $30/月起，企业计划 $60/月起。</p>
            <h3>适合人群</h3>
            <p>设计师、插画师、内容创作者、市场营销人员、产品经理、学生等。</p>
          </article>
        </div>
      </section>
      <section className="toolhub-solve">
        <h2>用它解决什么</h2>
        <p>Midjourney 可以帮助你快速把想法转化为高质量视觉内容，提升创作效率，激发灵感，并在各类场景中发挥重要作用。</p>
        <div>
          {["创意激发", "概念设计", "内容创作", "品牌视觉", "营销素材"].map((item) => (
            <article key={item}>
              <strong>{item}</strong>
              <small>快速生成可参考的视觉方案，辅助产品、场景、角色等设计探索。</small>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

function ToolPagination() {
  return (
    <footer className="toolhub-pagination" aria-label="工具分页">
      <button type="button">‹</button>
      {[1, 2, 3, 4, 5].map((page) => <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>)}
      <span>…</span>
      <button type="button">20</button>
      <button type="button">›</button>
      <small>每页显示 12 条⌄</small>
    </footer>
  );
}

function ToolsCopilot({ variant }: { variant: NonNullable<ToolsPageProps["variant"]> }) {
  const isRecommend = variant === "recommend";
  const isPlan = variant === "plan";
  const isDetail = variant === "detail";

  return (
    <aside className="learning-copilot toolhub-copilot" aria-label="智活 Copilot 工具助手">
      <header>
        <div>
          <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
          <p>你的全球 AI 助手，随时为你提供帮助</p>
        </div>
        <div className="learning-copilot-tools" aria-hidden="true"><span>⚙</span><span>⌄</span></div>
      </header>
      <div className="learning-chat toolhub-chat">
        {isPlan ? (
          <>
            <article className="user"><p>请帮我生成一套从市场调研到视频推广的工具方案</p></article>
            <article><span className="ai-avatar">A</span><p>好的，我已为你生成从市场调研到视频推广的整套工具方案，包含四个环节和执行建议。</p></article>
            <article className="toolhub-copilot-card"><Link to="/tools/recommendation-plan">查看从市场调研到视频推广的工具方案 ›</Link></article>
          </>
        ) : isRecommend ? (
          <>
            <article><span className="ai-avatar">A</span><p>收到！基于你的需求，我为你推荐了 3 款最合适的 AI 工具：ChatGPT、Canva AI、Notion AI。</p></article>
            <article><span className="ai-avatar">A</span><p>这三款工具可覆盖文案、设计与协作方面的主要需求，并兼顾免费可用与低成本策略。</p></article>
          </>
        ) : isDetail ? (
          <>
            <article><span className="ai-avatar">A</span><p>嗨，张婧！今天我能帮你分析竞品、推荐工具或制定落地计划。</p></article>
            <article className="user"><p>试试问问我：“帮我生成新品上市推广方案”或“分析本月竞品动态”</p></article>
            <article><span className="ai-avatar">A</span><p>好的，已为你生成分析报告，包含市场规模、竞争格局和增长要点。</p></article>
          </>
        ) : (
          <>
            <article><span className="ai-avatar">A</span><p>嗨，张婧！今天想聚焦哪个方向？我可以帮你分析机会，推荐工具或制定落地计划。</p></article>
            <article className="user"><p>帮我分析一下智能硬件赛道的市场机会和潜在关键点。</p></article>
            <article><span className="ai-avatar">A</span><p>好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点。</p></article>
          </>
        )}
      </div>
      <nav className="learning-copilot-actions" aria-label="工具助手快捷入口">
        <Link to="/analysis">分析项目机会 <span aria-hidden="true">›</span></Link>
        <Link to="/tools/recommend">推荐工具 <span aria-hidden="true">›</span></Link>
        <Link to="/learning/plan">制定落地计划 <span aria-hidden="true">›</span></Link>
        {isRecommend && <Link to="/tools/recommendation-plan">生成整套方案 <span aria-hidden="true">›</span></Link>}
      </nav>
      <form className="learning-copilot-input">
        <button aria-label="添加附件" type="button">+</button>
        <input aria-label="向工具箱 Copilot 提问" placeholder="询问任何问题..." />
        <button aria-label="发送" type="button">⌁</button>
      </form>
    </aside>
  );
}

export default ToolsPage;
