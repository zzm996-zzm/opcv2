import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type ContentArticle } from "../lib/contentApi";

type InsightsPageProps = {
  variant?: "list" | "detail" | "fileAnalysis";
};

type DisplayArticle = {
  slug: string;
  title: string;
  summary: string;
  source: string;
  time: string;
  tags: readonly string[];
  image: string;
};

const imagePool = [
  "/insights/customer-service.jpg",
  "/insights/overseas.jpg",
  "/insights/marketing.jpg",
  "/insights/llm.jpg",
  "/insights/investment.jpg",
  "/insights/digital.jpg"
] as const;

const referenceArticles: ContentArticle[] = [
  {
    id: 101,
    slug: "smart-customer-service-growth",
    title: "企业智能客服落地实践：从成本中心到增长引擎",
    summary: "越来越多企业通过智能客服系统实现服务效率提升与客户体验升级。本文从部署策略、场景选择、效果评估等方面，分享从成本中心到增长引擎的实践路径。",
    body: "智能客服正在从单纯的效率工具升级为企业增长引擎。随着大模型、全渠道融合与数据闭环能力提升，企业可以在服务过程中持续识别客户需求、沉淀业务洞察，并把服务触点转化为新的增长机会。",
    status: "published",
    source_name: "虎嗅智库",
    source_url: "https://www.huxiu.com/",
    author: "量子位智库团队",
    category: "AI",
    tags: ["AI客服", "客户体验", "降本增效"],
    citations: [],
    published_at: "2024-06-25T10:30:00+08:00",
    created_at: "2024-06-25T10:30:00+08:00",
    updated_at: "2024-06-25T10:30:00+08:00"
  },
  {
    id: 102,
    slug: "southeast-asia-saas",
    title: "出海新机遇：东南亚SaaS市场洞察与机会分析",
    summary: "东南亚SaaS市场进入高速增长期，本报告从市场规模、核心国家、行业机会及竞争格局等维度，解读出海企业的增长机会。",
    status: "published",
    source_name: "亿欧智库",
    category: "商业",
    tags: ["出海", "SaaS", "东南亚"],
    citations: [],
    published_at: "2024-06-25T09:15:00+08:00",
    created_at: "2024-06-25T09:15:00+08:00",
    updated_at: "2024-06-25T09:15:00+08:00"
  },
  {
    id: 103,
    slug: "ai-automated-marketing",
    title: "AI驱动的自动化营销：提升转化率的实战方法论",
    summary: "结合AI算法与自动化工具，实现用户洞察、内容生成、渠道触达与效果优化的闭环，全面提升营销ROI。",
    status: "published",
    source_name: "增长黑盒",
    category: "案例",
    tags: ["自动化营销", "AI应用", "转化率"],
    citations: [],
    published_at: "2024-06-24T16:45:00+08:00",
    created_at: "2024-06-24T16:45:00+08:00",
    updated_at: "2024-06-24T16:45:00+08:00"
  },
  {
    id: 104,
    slug: "enterprise-llm-roadmap",
    title: "大模型在企业级应用的落地路径与挑战",
    summary: "从技术选型、数据治理、场景落地到组织变革，系统梳理大模型在企业端落地的关键步骤与常见挑战。",
    status: "published",
    source_name: "InfoQ",
    category: "行业",
    tags: ["大模型", "企业应用", "技术趋势"],
    citations: [],
    published_at: "2024-06-24T14:20:00+08:00",
    created_at: "2024-06-24T14:20:00+08:00",
    updated_at: "2024-06-24T14:20:00+08:00"
  },
  {
    id: 105,
    slug: "ai-investment-2024",
    title: "2024上半年AI投融资盘点：热门赛道与资本动向",
    summary: "回顾2024年上半年全球AI领域投融资情况，聚焦热门赛道、代表性企业与资本趋势，为创业者与投资人提供参考。",
    status: "published",
    source_name: "投资界",
    category: "商业",
    tags: ["投融资", "AI", "资本市场"],
    citations: [],
    published_at: "2024-06-24T11:05:00+08:00",
    created_at: "2024-06-24T11:05:00+08:00",
    updated_at: "2024-06-24T11:05:00+08:00"
  },
  {
    id: 106,
    slug: "digital-process-cases",
    title: "企业数字化转型加速：AI+业务流程重塑案例集",
    summary: "精选制造、零售、金融等行业的AI+流程重塑案例，解析企业如何通过智能化实现效率提升与业务创新。",
    status: "published",
    source_name: "艾瑞咨询",
    category: "案例",
    tags: ["数字化转型", "流程自动化", "行业案例"],
    citations: [],
    published_at: "2024-06-24T09:40:00+08:00",
    created_at: "2024-06-24T09:40:00+08:00",
    updated_at: "2024-06-24T09:40:00+08:00"
  }
];

function toDisplayArticle(article: ContentArticle, index: number): DisplayArticle {
  return {
    slug: article.slug,
    title: article.title,
    summary: article.summary || "资讯正文已收录，进入详情查看完整内容。",
    source: article.source_name || "量子位智库",
    time: formatInsightTime(article.source_published_at || article.published_at || article.created_at),
    tags: article.tags?.length ? article.tags : [article.category || "AI"],
    image: imagePool[index % imagePool.length]
  };
}

function InsightsPage({ variant = "list" }: InsightsPageProps) {
  return (
    <V4PageShell className="insights-v4-shell" showCopilotMini={false}>
      <section className={`insights-page insights-reference-page ${variant}`} aria-label="咨询通">
        <main className="insights-main">
          {variant === "detail" ? <InsightDetail /> : <InsightList />}
        </main>
        <InsightsCopilot variant={variant} />
      </section>
    </V4PageShell>
  );
}

function InsightList() {
  const [apiArticles, setApiArticles] = useState<ContentArticle[]>([]);
  const [category, setCategory] = useState("全部");
  const [search, setSearch] = useState("");

  useEffect(() => {
    let active = true;
    contentApi
      .listArticles({ limit: 20 })
      .then((payload) => {
        if (active) setApiArticles(payload.articles);
      })
      .catch(() => {
        if (active) setApiArticles([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const sourceArticles = apiArticles.length > 0 ? apiArticles : referenceArticles;
  const visibleArticles = useMemo(() => {
    const keyword = search.trim().toLowerCase();
    return sourceArticles.filter((article) => {
      const matchesCategory = category === "全部" || article.category === category || article.tags.includes(category);
      const matchesSearch = !keyword || `${article.title}${article.summary || ""}${article.tags.join("")}`.toLowerCase().includes(keyword);
      return matchesCategory && matchesSearch;
    }).map(toDisplayArticle);
  }, [category, search, sourceArticles]);

  const focusArticles = referenceArticles.slice(0, 3);

  return (
    <>
      <header className="insights-title">
        <h1>咨询通</h1>
        <p>浏览商业 / AI / 行业资讯，快速获取与你业务相关的动态与洞察</p>
      </header>

      <section className="insights-toolbar" aria-label="资讯筛选">
        <label className="insights-search">
          <span aria-hidden="true">⌕</span>
          <input
            aria-label="搜索资讯目录"
            onChange={(event) => setSearch(event.target.value)}
            placeholder="搜索资讯关键词，如：AI客服、出海、SaaS"
            value={search}
          />
        </label>
        <div className="insights-tabs" role="tablist" aria-label="资讯分类">
          {["全部", "商业", "AI", "行业", "政策", "案例"].map((item) => (
            <button
              aria-selected={category === item}
              className={category === item ? "active" : ""}
              key={item}
              onClick={() => setCategory(item)}
              role="tab"
              type="button"
            >
              {item}
            </button>
          ))}
        </div>
      </section>

      <section className="insights-focus" aria-label="今日关注">
        <div className="insights-focus-label">
          <h2>今日关注</h2>
          <small><span aria-hidden="true">🔥</span> 更新 12 条</small>
        </div>
        {focusArticles.map((article, index) => (
          <article key={article.slug}>
            <div>
              <strong>{index === 0 ? "AI客服行业市场规模持续增长" : index === 1 ? "出海企业布局东南亚加速" : "AI营销自动化趋势观察"}</strong>
              <p>{index === 0 ? "2024年中国智能客服市场规模预计突破120亿元，年复合增长率达28%。" : index === 1 ? "东南亚数字经济规模突破3000亿美元，SaaS与AI应用需求旺盛。" : "生成式AI与自动化营销深度融合，助力企业降本增效与精细化运营。"}</p>
            </div>
            <span className={`focus-art focus-${index + 1}`} aria-hidden="true" />
          </article>
        ))}
        <Link to="/insights/file-analysis">查看全部专题 <span aria-hidden="true">›</span></Link>
      </section>

      <section className="insight-list-card" aria-label="资讯列表">
        {visibleArticles.length === 0 ? (
          <div className="insights-empty-state" role="status">
            <span className="insights-empty-visual" aria-hidden="true" />
            <div><h2>没有找到相关资讯</h2><p>换一个关键词或分类试试。</p></div>
          </div>
        ) : visibleArticles.map((article) => (
          <article className="insight-row" key={article.slug}>
            <img alt="" className="insight-thumb" src={article.image} />
            <div className="insight-row-copy">
              <h2>{article.title}</h2>
              <p>{article.summary}</p>
              <footer>
                <strong>◉ {article.source}</strong>
                <time>{article.time}</time>
                {article.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}
              </footer>
            </div>
            <button aria-label={`收藏${article.title}`} className="insight-star" type="button">☆</button>
            <Link className="insight-detail-link" to={`/insights/detail?article=${article.slug}`}>查看详情</Link>
          </article>
        ))}
      </section>

      <nav className="insights-pagination" aria-label="资讯分页">
        <button aria-label="上一页" type="button">‹</button>
        {[1, 2, 3, 4, 5].map((page) => <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>)}
        <span>…</span><button type="button">20</button><button aria-label="下一页" type="button">›</button>
        <small>每页显示&nbsp;&nbsp; 12 &nbsp;条⌄</small>
      </nav>
    </>
  );
}

function InsightDetail() {
  const location = useLocation();
  const articleSlug = new URLSearchParams(location.search).get("article") || referenceArticles[0].slug;
  const fallbackArticle = referenceArticles.find((item) => item.slug === articleSlug) || referenceArticles[0];
  const [article, setArticle] = useState<ContentArticle>(fallbackArticle);
  const [bookmarked, setBookmarked] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi.getArticle(articleSlug).then((payload) => {
      if (active) setArticle(payload);
    }).catch(() => {
      if (active) setArticle(fallbackArticle);
    });
    return () => { active = false; };
  }, [articleSlug, fallbackArticle]);

  async function toggleBookmark() {
    setError("");
    try {
      const result = bookmarked
        ? await contentApi.unbookmarkArticle(article.slug)
        : await contentApi.bookmarkArticle(article.slug);
      setBookmarked(result.bookmarked);
    } catch (caught) {
      setBookmarked((value) => !value);
      setError(apiErrorMessage(caught, "收藏状态已在本地更新"));
    }
  }

  const publishedAt = formatInsightTime(article.source_published_at || article.published_at || article.created_at);
  const isReference = article.slug === referenceArticles[0].slug;

  return (
    <article className="insight-detail-page">
      <nav aria-label="资讯路径"><Link to="/insights">咨询通</Link><span>/</span><Link to="/insights">AI资讯</Link><span>/</span><b>资讯详情</b></nav>
      <header>
        <h1>{isReference ? "企业智能客服进入规模化落地阶段：从效率工具走向增长引擎" : article.title}</h1>
        <div className="insight-meta-row">
          <span>来源：{article.source_name || "量子位智库"}</span>
          <time>发布时间：{publishedAt}</time>
          <span>作者：{article.author || "量子位智库团队"}</span>
        </div>
        <div className="insight-tag-row">
          {(article.tags?.length ? article.tags : [article.category || "AI"]).slice(0, 5).map((tag) => <span key={tag}>{tag}</span>)}
        </div>
        <div className="insight-detail-actions">
          <Link className="primary" to={`/insights/file-analysis?article=${article.slug}`}><span aria-hidden="true">✦</span> 分析我能学到什么</Link>
          <button aria-label={bookmarked ? "取消收藏资讯" : "收藏资讯"} onClick={toggleBookmark} type="button">{bookmarked ? "★ 已收藏" : "☆ 收藏"}</button>
          {article.source_url ? <a href={article.source_url}>原始来源 ↗</a> : <button type="button">原始来源 ↗</button>}
        </div>
        {error && <p className="sr-only" role="status">{error}</p>}
      </header>

      <div className="insight-article-layout">
        <div className="insight-article">
          <img alt="智能客服行业资讯配图" className="insight-article-cover" src="/insights/detail-hero.jpg" />
          <section>
            <h2>市场背景</h2>
            <p>{article.body || "在大模型技术加速成熟、企业数字化转型纵深推进的背景下，智能客服正从早期的“降本工具”加速演进为“增长伙伴”。2023年中国智能客服市场规模超过152亿元，同比增长28.7%，预计2025年将突破230亿元。头部企业的实践表明，智能客服在提升服务效率的同时，正在客户体验、销售转化与运营优化等方面创造显著的业务价值。"}</p>
            <h2>关键趋势</h2>
            <ul>
              <li><b>大模型驱动体验跃迁：</b>大模型的语义理解、生成与知识推理能力，显著提升复杂问题解决率与用户满意度。</li>
              <li><b>全渠道融合成为标配：</b>打通在线客服、电话、社媒、邮件等触点，构建一致的客户服务旅程。</li>
              <li><b>从被动响应到主动服务：</b>基于用户行为与意图识别，主动触达、预警与推荐，推动服务向更高价值转化。</li>
              <li><b>数据闭环与知识进化：</b>构建“数据-知识-模型-应用”的闭环体系，知识自学习与持续优化成为核心竞争力。</li>
            </ul>
            <h2>企业落地启示</h2>
            <p>智能客服的价值不再局限于降本，更体现在对客户体验的重塑与业务增长的驱动。领先企业将智能客服与CRM、营销自动化、工单系统等深度集成，实现服务即洞察、服务即增长的业务闭环。</p>
            <h2>行动建议</h2>
            <ol>
              <li>优先梳理高频高价值场景，从“问题解决率”和“客户满意度”双指标驱动落地。</li>
              <li>构建企业专属知识中台，沉淀高质量知识资产，持续优化模型效果。</li>
              <li>打通服务与营销链路，将服务过程转为营销增量，通过AI服务中心实现增长。</li>
              <li>建立数据监测与效果评估体系，持续迭代与规模化复制成功经验。</li>
            </ol>
          </section>
        </div>

        <aside className="insight-summary-card">
          <h2>资讯摘要</h2>
          <section><strong>核心观点</strong><p>{article.summary || "智能客服已从效率工具升级为增长引擎，通过大模型、全渠道融合与数据闭环，驱动客户体验提升与业务增长。"}</p></section>
          <section><strong>适合谁关注</strong><div>{["客户服务负责人", "数字化转型负责人", "产品负责人", "增长与运营负责人", "IT与数据团队"].map((tag) => <span key={tag}>{tag}</span>)}</div></section>
          <section><strong>相关主题</strong><div>{["大模型应用", "客户体验管理", "数字化运营", "智能营销", "数据智能"].map((tag) => <span key={tag}>{tag}</span>)}</div></section>
          {article.citations.length > 0 && <section><strong>已保存引用</strong><div>{article.citations.map((citation) => <a href={citation.source_url} key={citation.id}>{citation.source_name}：{citation.label}</a>)}</div></section>}
        </aside>
      </div>

      <section className="insight-related">
        <h2>相关推荐</h2>
        <div>{referenceArticles.slice(1, 4).map((item, index) => <Link key={item.slug} to={`/insights/detail?article=${item.slug}`}><img alt="" src={imagePool[index + 1]} /><strong>{item.title}</strong><small>{formatInsightTime(item.published_at || item.created_at)}</small><b>{item.tags[0]}</b></Link>)}</div>
      </section>
    </article>
  );
}

function InsightsCopilot({ variant }: { variant: NonNullable<InsightsPageProps["variant"]> }) {
  const isDetail = variant === "detail";
  const isFileAnalysis = variant === "fileAnalysis";
  return (
    <aside className={`insights-copilot ${isFileAnalysis ? "analysis-open" : ""}`} aria-label="智活 Copilot 咨询助手">
      <header className="insights-ai-head">
        <div><strong><span aria-hidden="true">✦</span> 智活 Copilot</strong><p>你的全球 AI 助手，随时为你提供帮助</p></div>
        <div className="insights-ai-tools"><button aria-label="设置" type="button">⚙</button><button aria-label="收起" type="button">⌄</button></div>
      </header>
      {isFileAnalysis ? <FileAnalysisChat /> : isDetail ? <DetailCopilot /> : <ListCopilot />}
      {!isFileAnalysis && (
        <nav className="insights-copilot-actions" aria-label="咨询助手快捷入口">
          {(isDetail ? ["总结这篇资讯要点", "提炼行业启示", "推荐相关工具/报告"] : ["追踪 AI 客服行业资讯", "总结今天的重点动态", "推荐相关报告与工具"]).map((item) => <Link key={item} to="/insights/file-analysis"><span aria-hidden="true">▣</span>{item}<b aria-hidden="true">›</b></Link>)}
        </nav>
      )}
      {!isFileAnalysis && <form className="insights-chat-composer" onSubmit={(event) => event.preventDefault()}><button aria-label="添加附件" type="button">＋</button><input aria-label="咨询通提问" placeholder="询问任何问题..." /><button aria-label="发送问题" type="submit">➤</button></form>}
    </aside>
  );
}

function ListCopilot() {
  return <div className="insights-chat"><article><span className="ai-avatar">A</span><p>嗨，张婧！<br />今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</p></article><article className="user"><p>帮我追踪 AI 客服行业资讯<br />并总结今日重点动态。</p></article><article><span className="ai-avatar">A</span><div><p>好的，已为你生成今日重点资讯摘要，包含趋势、机会与行动建议，点击下方查看详情。</p><span className="learning-file-chip">今日 AI 客服行业资讯摘要<small>PDF · 1.7 MB</small></span></div></article></div>;
}

function DetailCopilot() {
  return <div className="insights-chat"><article><span className="ai-avatar">A</span><p>您好，我是智活 Copilot。<br />您可以问我这篇资讯、分析内容并提炼优质出处引用。</p></article><article><span className="ai-avatar">A</span><div><p>您可以这样问（与资讯相关）</p><ul><li>行业趋势、市场动态</li><li>企业动态、投融资信息</li><li>政策法规、行业标准</li><li>技术发展、产品对比</li></ul></div></article><article><span className="ai-avatar">A</span><div><p>例如：</p><ul><li>2024 年智能客服行业的最新趋势</li><li>国内 AI 客服领域的头部企业</li><li>智能客服在金融行业的落地案例</li></ul></div></article></div>;
}

function FileAnalysisChat() {
  const location = useLocation();
  const requestedArticle = new URLSearchParams(location.search).get("article") || referenceArticles[0].slug;
  const [articles, setArticles] = useState<ContentArticle[]>(referenceArticles);
  const [selectedArticle, setSelectedArticle] = useState(requestedArticle);
  const [question, setQuestion] = useState("");
  const [answer, setAnswer] = useState<Awaited<ReturnType<typeof contentApi.answerInsightQuestion>> | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi.listArticles({ limit: 20 }).then((payload) => {
      if (!active || payload.articles.length === 0) return;
      setArticles(payload.articles);
      setSelectedArticle(payload.articles.some((article) => article.slug === requestedArticle) ? requestedArticle : payload.articles[0].slug);
    }).catch(() => undefined);
    return () => { active = false; };
  }, [requestedArticle]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedArticle || !question.trim() || pending) return;
    setPending(true);
    setError("");
    try {
      setAnswer(await contentApi.answerInsightQuestion(question, [selectedArticle]));
    } catch (caught) {
      setAnswer(null);
      setError(apiErrorMessage(caught, "暂时无法生成带引用回答"));
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="insights-file-chat">
      <div className="insights-chat"><article><span className="ai-avatar">A</span><p>请帮我分析 2024 年企业 AI 客服的市场趋势、代表企业和落地机会。</p></article><article><span className="ai-avatar">A</span><div><p>为你分析如下：</p><h3>市场趋势</h3><p>2024 年企业 AI 客服市场持续向智能化、全渠道融合和场景化演进，大模型与 RAG 技术提升了复杂问题解决率。</p><h3>代表企业</h3><p>国内阿里云、腾讯云、百度智能云、华为云，以及 Salesforce、Intercom、Ada、Zendesk 等持续布局智能客服。</p><h3>落地机会</h3><p>重点机会在于大模型能力升级、企业知识中台建设、行业垂直方案和客服营销一体化。</p><h3>建议动作</h3><p>建议企业从高频场景切入，优化知识库与自动化流程，并建立可量化的服务与增长指标。</p></div></article></div>
      {answer && <article className="insights-live-answer"><span className="ai-avatar">A</span><div><p>{answer.answer}</p>{answer.citations.length > 0 && <section className="insights-reference-panel"><h2>出处引用</h2><div>{answer.citations.map((citation) => <a href={citation.source_url} key={citation.id}><b>{citation.source_name}</b><span>{citation.label}</span><small>{citation.excerpt}</small></a>)}</div></section>}</div></article>}
      <section className="insights-static-sources"><h2>出处引用（点击查看原文）</h2><div>{[["艾瑞咨询", "《2024年中国智能客服行业研究报告》", "2024-06-18"], ["IDC", "《中国AI应用市场（2024）预测》", "2024-05-22"], ["赛迪顾问", "《2025中国企业AI应用白皮书》", "2024-06-12"], ["Gartner", "Cool Vendors in Customer Service and Support, 2024", "2024-07-15"]].map(([source, title, date]) => <a href="#sources" key={title}><b>{source}</b><span>{title}</span><small>{date}</small></a>)}</div></section>
      <form className="insights-chat-composer" onSubmit={submit}>
        <select aria-label="问答资讯" className="sr-only" onChange={(event) => setSelectedArticle(event.target.value)} value={selectedArticle}>{articles.map((article) => <option key={article.slug} value={article.slug}>{article.title}</option>)}</select>
        <input aria-label="资讯问答问题" onChange={(event) => setQuestion(event.target.value)} placeholder="继续提问，获取更精准的资讯..." value={question} />
        <button aria-label="生成带引用回答" disabled={!question.trim() || pending} type="submit">{pending ? "…" : "➤"}</button>
      </form>
      {error && <p className="form-error" role="alert">{error}</p>}
    </div>
  );
}

function formatInsightTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", hour12: false }).replace(/\//g, "-");
}

export default InsightsPage;
