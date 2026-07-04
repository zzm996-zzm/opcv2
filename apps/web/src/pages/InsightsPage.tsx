import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type ContentArticle } from "../lib/contentApi";
import { CdkTopNav } from "./AnalysisPage";

type InsightsPageProps = {
  variant?: "list" | "detail" | "fileAnalysis";
};

const articleVisuals = ["bot", "map", "target", "chip", "growth", "cloud"] as const;

const references = [
  ["艾瑞咨询", "《2024年中国智能客服行业研究报告》", "2024-04-18"],
  ["IDC", "《中国AI应用市场（2024）预测》", "2024-05-22"],
  ["赛迪顾问", "《2024中国企业AI应用白皮书》", "2024-06-12"],
  ["Gartner", "Cool Vendors in Customer Service and Support, 2024", "2024-07-15"]
] as const;

type DisplayArticle = {
  slug: string;
  title: string;
  summary: string;
  source: string;
  time: string;
  tags: readonly string[];
  visual: string;
};

function toDisplayArticle(article: ContentArticle, index: number): DisplayArticle {
  return {
    slug: article.slug,
    title: article.title,
    summary: article.summary || "资讯正文已收录，进入详情查看完整内容。",
    source: "智活AI研究院",
    time: formatInsightTime(article.published_at || article.created_at),
    tags: ["AI创业", "增长洞察", article.status === "published" ? "已发布" : "草稿"],
    visual: articleVisuals[index % articleVisuals.length]
  };
}

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
  const [apiArticles, setApiArticles] = useState<DisplayArticle[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi
      .listArticles()
      .then((payload) => {
        if (!active) return;
        setApiArticles(payload.articles.map(toDisplayArticle));
      })
      .catch((error) => {
        if (!active) return;
        setError(apiErrorMessage(error, "暂时无法读取资讯列表"));
        setApiArticles([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleArticles = apiArticles;

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
          <small>更新 {visibleArticles.length} 条</small>
        </div>
        <div className="module-empty-state" role="status">暂无今日关注</div>
        <Link to="/insights/file-analysis">查看全部专题 ›</Link>
      </section>

      <section className={`insight-list-card ${compact ? "compact" : ""}`} aria-label="资讯列表">
        {error && <p className="form-error" role="alert">{error}</p>}
        {visibleArticles.length === 0 ? (
          <div className="module-empty-state" role="status">暂无资讯数据</div>
        ) : visibleArticles.map((article) => (
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
              <Link className="insight-detail-link" to={`/insights/detail?article=${article.slug}`}>查看详情</Link>
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
  const location = useLocation();
  const articleSlug = new URLSearchParams(location.search).get("article") || "ai-customer-service";
  const [article, setArticle] = useState<ContentArticle | null>(null);
  const [bookmarked, setBookmarked] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    setError("");
    contentApi
      .getArticle(articleSlug)
      .then((payload) => {
        if (active) setArticle(payload);
      })
      .catch((error) => {
        if (!active) return;
        setArticle(null);
        setError(apiErrorMessage(error, "暂时无法读取资讯详情"));
      });
    return () => {
      active = false;
    };
  }, [articleSlug]);

  async function toggleBookmark() {
    setError("");
    try {
      const result = bookmarked
        ? await contentApi.unbookmarkArticle(articleSlug)
        : await contentApi.bookmarkArticle(articleSlug);
      setBookmarked(result.bookmarked);
    } catch (error) {
      setError(apiErrorMessage(error, "收藏失败，请稍后再试"));
    }
  }

  if (!article) {
    return (
      <article className="insight-detail-page">
        <nav aria-label="资讯路径">
          <Link to="/insights">咨询通</Link>
          <span>/</span>
          <b>资讯详情</b>
        </nav>
        <header>
          <div className="insight-detail-actions">
            <Link className="primary" to="/insights/file-analysis">分析我能学到什么</Link>
          </div>
          {error && <p className="form-error" role="alert">{error}</p>}
        </header>
        <div className="module-empty-state" role="status">暂无资讯详情</div>
      </article>
    );
  }

  const title = article.title;
  const summary = article.summary || "暂无资讯摘要";
  const body = article.body || "暂无资讯正文";
  const publishedAt = formatInsightTime(article.published_at || article.created_at);

  return (
    <article className="insight-detail-page">
      <nav aria-label="资讯路径">
        <Link to="/insights">咨询通</Link>
        <span>/</span>
        <Link to="/insights">AI资讯</Link>
        <span>/</span>
        <b>{title}</b>
      </nav>
      <header>
        <h1>{title}</h1>
        <div className="insight-meta-row">
          <span>来源：智活AI研究院</span>
          <time>发布时间：{publishedAt}</time>
          <span>作者：量子位智库团队</span>
        </div>
        <div className="insight-tag-row">
          {["智能客服", "大模型应用", "客户体验", "降本增效", "增长引擎"].map((tag) => <span key={tag}>{tag}</span>)}
        </div>
        <div className="insight-detail-actions">
          <Link className="primary" to="/insights/file-analysis">分析我能学到什么</Link>
          <button aria-label={bookmarked ? "取消收藏资讯" : "收藏资讯"} onClick={toggleBookmark} type="button">
            {bookmarked ? "★ 已收藏" : "☆ 收藏"}
          </button>
          <button type="button">原始来源 ↗</button>
        </div>
        {error && <p className="form-error" role="alert">{error}</p>}
      </header>
      <div className="insight-article-layout">
        <div className="insight-article">
          <span className="insight-article-cover bot" aria-hidden="true" />
          <section>
            <h2>市场背景</h2>
            <p>{body}</p>
          </section>
        </div>

        <aside className="insight-summary-card">
          <h2>资讯摘要</h2>
          <section>
            <strong>核心观点</strong>
            <p>{summary}</p>
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

function formatInsightTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  }).replace(/\//g, "-");
}

export default InsightsPage;
