import { Link } from "react-router-dom";

import { CdkTopNav } from "./AnalysisPage";

type InsightsPageProps = {
  variant?: "list" | "detail" | "fileAnalysis";
};

const focusItems = [
  ["AI 客服行业市场规模持续增长", "2024 年中国智能客服市场规模预计突破 120 亿元，年复合增长率达 28%。", "chart"],
  ["出海企业布局东南亚加速", "东南亚数字经济规模突破 3000 亿美元，SaaS 与 AI 应用需求旺盛。", "globe"],
  ["AI 营销自动化趋势观察", "生成式 AI 与自动化营销深度融合，助力企业降本增效与精细化运营。", "cube"]
] as const;

const articles = [
  {
    title: "企业智能客服落地实践：从成本中心到增长引擎",
    summary: "越来越多企业通过智能客服系统实现服务效率提升与客户体验升级。本文从部署策略、场景选择、效果评估等方面，分享企业的实践路径与关键启示。",
    source: "虎嗅智库",
    time: "2024-06-25 10:30",
    tags: ["AI客服", "客户体验", "降本增效"],
    visual: "bot"
  },
  {
    title: "出海新机遇：东南亚SaaS市场洞察与机会分析",
    summary: "东南亚 SaaS 市场进入高速增长期，本报告从市场规模、核心国家、行业机会及竞争格局等维度，解读出海企业的增长机会。",
    source: "亿欧智库",
    time: "2024-06-25 09:15",
    tags: ["出海", "SaaS", "东南亚"],
    visual: "map"
  },
  {
    title: "AI驱动的自动化营销：提升转化率的实战方法论",
    summary: "结合 AI 算法与自动化工具，实现用户洞察、内容生成、渠道触达与效果优化的闭环，全面提升营销 ROI。",
    source: "增长黑盒",
    time: "2024-06-24 16:45",
    tags: ["自动化营销", "AI应用", "转化率"],
    visual: "target"
  },
  {
    title: "大模型在企业级应用的落地路径与挑战",
    summary: "从技术选型、数据治理、场景落地到组织变革，系统梳理大模型在企业落地的关键步骤与常见挑战。",
    source: "InfoQ",
    time: "2024-06-24 14:20",
    tags: ["大模型", "企业应用", "技术趋势"],
    visual: "chip"
  },
  {
    title: "2024上半年AI投融资盘点：热门赛道与资本动向",
    summary: "回顾 2024 年上半年全球 AI 领域投融资情况，聚焦热门赛道、代表性企业与资本趋势，为创业者与投资人提供参考。",
    source: "投资界",
    time: "2024-06-24 11:05",
    tags: ["投融资", "AI", "资本市场"],
    visual: "growth"
  },
  {
    title: "企业数字化转型加速：AI+业务流程重塑案例集",
    summary: "精选制造、零售、金融等行业的 AI+ 流程重塑案例，解析企业如何通过智能化实现效率提升与业务创新。",
    source: "钛媒体",
    time: "2024-06-24 09:40",
    tags: ["数字化转型", "流程自动化", "行业案例"],
    visual: "cloud"
  }
] as const;

const relatedCards = [
  ["大模型时代智能客服的技术演进与应用实践", "量子位智库", "2024-06-18 14:22", "大模型应用"],
  ["从降本到增长：智能客服如何驱动业务增长", "艾瑞咨询", "2024-06-10 10:30", "增长"],
  ["智能客服落地指南：从0到1构建企业专属方案", "智活AI研究院", "2024-06-05 09:45", "落地方法"]
] as const;

const references = [
  ["艾瑞咨询", "《2024年中国智能客服行业研究报告》", "2024-04-18"],
  ["IDC", "《中国AI应用市场（2024）预测》", "2024-05-22"],
  ["赛迪顾问", "《2024中国企业AI应用白皮书》", "2024-06-12"],
  ["Gartner", "Cool Vendors in Customer Service and Support, 2024", "2024-07-15"]
] as const;

function InsightsPage({ variant = "list" }: InsightsPageProps) {
  return (
    <main className="cdk-analysis-page cdk-insights-page">
      <CdkTopNav active="咨询通" />
      <section className={`insights-page ${variant === "detail" ? "detail" : ""}`} aria-label="咨询通">
        <main className="insights-main">
          {variant === "detail" ? <InsightDetail /> : <InsightList compact={variant === "fileAnalysis"} />}
        </main>
        <InsightsCopilot variant={variant} />
      </section>
    </main>
  );
}

function InsightList({ compact }: { compact: boolean }) {
  return (
    <>
      <header className="insights-title">
        <span className="insights-title-icon" aria-hidden="true">•••</span>
        <div>
          <h1>咨询通</h1>
          <p>AI创业资讯、案例、政策、经营趋势，一站掌握，助你决策快人一步</p>
          <div className="insights-benefits">
            {["AI筛选高价值内容", "行业主题实时追踪", "每日更新，节省信息搜集时间"].map((item) => (
              <span key={item}>✓ {item}</span>
            ))}
          </div>
        </div>
      </header>

      <section className="insights-filter-bar" aria-label="资讯筛选">
        <div className="insights-tabs" role="tablist" aria-label="资讯分类">
          {["精选", "融资", "获客案例", "工具更新", "政策风险", "行业趋势"].map((category, index) => (
            <button className={index === 0 ? "active" : ""} key={category} role="tab" type="button">
              {category}
            </button>
          ))}
        </div>
      </section>

      <section className="insights-focus" aria-label="今日关注">
        <div>
          <h2>今日关注</h2>
          <small>🔥 更新 12 条</small>
        </div>
        {focusItems.map(([title, desc, visual]) => (
          <article key={title}>
            <span className={`focus-art ${visual}`} aria-hidden="true" />
            <strong>{title}</strong>
            <p>{desc}</p>
          </article>
        ))}
        <Link to="/insights/file-analysis">查看全部专题 ›</Link>
      </section>

      <section className={`insight-list-card ${compact ? "compact" : ""}`} aria-label="资讯列表">
        {articles.map((article) => (
          <article className="insight-row" key={article.title}>
            <span className={`insight-thumb ${article.visual}`} aria-hidden="true" />
            <div className="insight-row-copy">
              <h2>{article.title}</h2>
              <p>{article.summary}</p>
              <footer>
                <strong>{article.source}</strong>
                <time>{article.time}</time>
                {article.tags.map((tag) => <span key={tag}>{tag}</span>)}
              </footer>
            </div>
            <button aria-label={`收藏${article.title}`} className="insight-star" type="button">☆</button>
            <Link className="insight-detail-link" to="/insights/detail">查看详情</Link>
          </article>
        ))}
      </section>

      <footer className="insights-pagination" aria-label="资讯分页">
        <button type="button">‹</button>
        {[1, 2, 3, 4, 5].map((page) => (
          <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>
        ))}
        <span>…</span>
        <button type="button">20</button>
        <button type="button">›</button>
        <small>每页显示 12 条⌄</small>
      </footer>
    </>
  );
}

function InsightDetail() {
  return (
    <article className="insight-detail-page">
      <nav aria-label="资讯路径">
        <Link to="/insights">咨询通</Link>
        <span>/</span>
        <Link to="/insights">AI资讯</Link>
        <span>/</span>
        <b>资讯详情</b>
      </nav>
      <header>
        <h1>企业智能客服进入规模化落地阶段：从效率工具走向增长引擎</h1>
        <div className="insight-meta-row">
          <span>来源：量子位智库</span>
          <time>发布时间：2024-06-25 09:15</time>
          <span>作者：量子位智库团队</span>
        </div>
        <div className="insight-tag-row">
          {["智能客服", "大模型应用", "客户体验", "降本增效", "增长引擎"].map((tag) => <span key={tag}>{tag}</span>)}
        </div>
        <div className="insight-detail-actions">
          <Link className="primary" to="/insights/file-analysis">分析我能学到什么</Link>
          <button type="button">☆ 收藏</button>
          <button type="button">原始来源 ↗</button>
        </div>
      </header>
      <div className="insight-article-layout">
        <div className="insight-article">
          <span className="insight-article-cover bot" aria-hidden="true" />
          <section>
            <h2>市场背景</h2>
            <p>在大模型技术加速成熟、企业数字化转型纵深推进的背景下，智能客服正从早期的“降本工具”加速演进为“增长触点”。2023年中国智能客服市场规模超过百亿元，同期增长显著，预计未来将保持较高增速。</p>
          </section>
          <section>
            <h2>关键趋势</h2>
            <ul>
              <li>大模型驱动体验跃迁：多轮理解、生成式回复与知识推理能力提升服务质量。</li>
              <li>全渠道融合成为标配：打通在线客服、电话、社媒、邮件等触点，构建一致客户服务旅程。</li>
              <li>从被动响应到主动服务：基于用户行为与意图识别，提前触达并推动服务向增值转化。</li>
              <li>数据闭环与知识进化：对话数据沉淀为企业知识资产，持续提升业务洞察能力。</li>
            </ul>
          </section>
          <section>
            <h2>企业落地启示</h2>
            <p>智能客服的价值不再局限于降本，更体现在对客户体验的重塑与业务增长的驱动。领先企业将智能客服与 CRM、营销自动化、工单系统等深度集成，实现客户洞察、服务响应与业务转化的协同。</p>
          </section>
          <section>
            <h2>行动建议</h2>
            <ol>
              <li>优先梳理高频高价值场景，从“问题解决率”和“客户满意度”双指标驱动落地。</li>
              <li>构建企业专属知识中台，沉淀高质量知识资产，持续优化模型效果。</li>
              <li>打通服务与销售链路，将客服数据参与销售提醒、商机识别和客户分层。</li>
              <li>建立数据监测与效果评估体系，持续迭代与规模化复制成功经验。</li>
            </ol>
          </section>
        </div>

        <aside className="insight-summary-card">
          <h2>资讯摘要</h2>
          <section>
            <strong>核心观点</strong>
            <p>智能客服已从效率工具升级为增长引擎，通过大模型、全渠道融合与数据闭环，驱动客户体验提升与业务增长。</p>
          </section>
          <section>
            <strong>适合谁关注</strong>
            <div>
              {["客户服务负责人", "数字化转型负责人", "产品负责人", "增长与运营负责人", "IT与数据团队"].map((tag) => <span key={tag}>{tag}</span>)}
            </div>
          </section>
          <section>
            <strong>相关主题</strong>
            <div>
              {["大模型应用", "客户体验管理", "数字化运营", "智能营销", "数据智能"].map((tag) => <span key={tag}>{tag}</span>)}
            </div>
          </section>
        </aside>
      </div>

      <section className="insight-related" aria-label="相关推荐">
        <h2>相关推荐</h2>
        <div>
          {relatedCards.map(([title, source, time, tag], index) => (
            <Link key={title} to="/insights/detail">
              <span className={`insight-thumb small ${articles[index].visual}`} aria-hidden="true" />
              <strong>{title}</strong>
              <small>{source} · {time}</small>
              <b>{tag}</b>
            </Link>
          ))}
        </div>
      </section>
    </article>
  );
}

function InsightsCopilot({ variant }: { variant: NonNullable<InsightsPageProps["variant"]> }) {
  const isFileAnalysis = variant === "fileAnalysis";
  const isDetail = variant === "detail";

  return (
    <aside className={`learning-copilot insights-copilot ${isFileAnalysis ? "analysis-open" : ""}`} aria-label="智活 Copilot 咨询助手">
      <header className="insights-ai-head">
        <span className="insights-bot-art" aria-hidden="true" />
        <div>
          <strong><span aria-hidden="true">✦</span> AI咨询通助手 <b>✦</b></strong>
          <p>有什么想了解的咨询？输入关键词，AI为你推荐相关文章</p>
        </div>
      </header>

      <label className="insights-side-search">
        <input aria-label="搜索资讯" placeholder="输入关键词，例如：AI获客、政策补贴、内容营销..." />
        <button type="button">⌕</button>
      </label>

      {isFileAnalysis ? <FileAnalysisChat /> : isDetail ? <DetailChat /> : <ListChat />}

      <nav className="learning-copilot-actions" aria-label="咨询助手快捷入口">
        <Link to="/insights/file-analysis">总结这篇资讯重点 <span aria-hidden="true">›</span></Link>
        <Link to="/insights/detail">提炼行业启示 <span aria-hidden="true">›</span></Link>
        <Link to="/tools/recommend">推荐相关工具/报告 <span aria-hidden="true">›</span></Link>
      </nav>

      <form className="learning-copilot-input">
        <button aria-label="添加附件" type="button">+</button>
        <input aria-label="向咨询通 Copilot 提问" placeholder={isFileAnalysis ? "继续提问，获取更精准的资讯..." : "询问任何问题..."} />
        <button aria-label="发送" type="button">⌁</button>
      </form>

      <section className="insights-question-card" aria-label="可复用选题">
        <header><h2>可复用选题</h2><button type="button">换一批</button></header>
        {["AI创业如何选择第一个落地场景？", "AI产品冷启动的3种低成本获客策略", "2024年AI创业的政策红利有哪些？"].map((topic, index) => (
          <p key={topic}><span>{topic}</span><small>热度 {index === 0 ? 86 : index === 1 ? 74 : 68}</small></p>
        ))}
      </section>

      <section className="insights-topic-card" aria-label="社群讨论话题">
        <header><h2>社群讨论话题</h2><Link to="/community">去社群 ›</Link></header>
        {["你在用哪些AI工具提升团队效率？", "AI创业者如何构建自己的护城河？", "最近有哪些值得关注的AI融资事件？"].map((topic, index) => (
          <p key={topic}><span>{topic}</span><small>{[128, 96, 73][index]}条讨论</small></p>
        ))}
      </section>
    </aside>
  );
}

function ListChat() {
  return (
    <div className="learning-chat insights-chat">
      <article>
        <span className="ai-avatar">A</span>
        <p>嗨，张婧！<br />今天想聚焦哪个方向？我可以帮你分析机会，推荐工具或制定落地计划。</p>
      </article>
      <article className="user">
        <p>帮我追踪 AI 客服行业资讯并总结今日重点动态。</p>
      </article>
      <article>
        <span className="ai-avatar">A</span>
        <div>
          <p>好的，已为你生成今日重点资讯摘要，包含趋势、机会与行动建议。</p>
          <span className="learning-file-chip">今日 AI 客服行业资讯摘要<small>PDF · 1.2 MB</small></span>
        </div>
      </article>
    </div>
  );
}

function DetailChat() {
  return (
    <div className="learning-chat insights-chat">
      <article>
        <span className="ai-avatar">A</span>
        <p>您好，我是智活 Copilot。您可以问我这篇资讯、分析内容并提供权威出处引用。</p>
      </article>
      <article>
        <span className="ai-avatar">A</span>
        <div>
          <p>您可以这样问（与资讯相关）</p>
          <ul>
            <li>行业趋势、市场动态</li>
            <li>企业动态、投融资信息</li>
            <li>政策法规、行业标准</li>
            <li>技术发展、产品对比</li>
          </ul>
        </div>
      </article>
      <article>
        <span className="ai-avatar">A</span>
        <div>
          <p>例如：</p>
          <ul>
            <li>2024 年智能客服行业的最新趋势</li>
            <li>国内 AI 客服领域的头部企业</li>
            <li>AI 客服在金融行业的落地案例</li>
          </ul>
        </div>
      </article>
    </div>
  );
}

function FileAnalysisChat() {
  return (
    <div className="learning-chat insights-chat">
      <article className="user">
        <p>请帮我分析 2024 年企业 AI 客服的市场趋势、代表企业和落地机会。</p>
      </article>
      <article>
        <span className="ai-avatar">A</span>
        <div className="analysis-answer">
          <p>为您分析如下：</p>
          <strong>市场趋势</strong>
          <p>2024 年企业 AI 客服市场持续向智能化、全渠道融合和场景化落地演进。大模型与 RAG 技术提升了复杂问题解决率，多模态交互增强体验。</p>
          <strong>代表企业</strong>
          <p>国内：阿里云、腾讯云、百度智能云、华为云；垂直厂商：Salesforce、Intercom、Ada、Zendesk 等在售前售后链路赋能方面能力突出。</p>
          <strong>落地机会</strong>
          <p>重点机会在于服务自动化、知识库平台、行业专属问答、客户数据分析与销售转化联动。</p>
          <strong>建议动作</strong>
          <p>建议企业从高频场景切入，优先梳理高价值知识库与自动化流程，并同步建立 ROI 与客户体验指标。</p>
        </div>
      </article>
      <section className="insights-reference-panel" aria-label="出处引用">
        <h2>出处引用（点击查看原文）</h2>
        <div>
          {references.map(([source, title, time]) => (
            <Link key={title} to="/insights/detail">
              <b>{source}</b>
              <span>{title}</span>
              <small>{time}</small>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

export default InsightsPage;
