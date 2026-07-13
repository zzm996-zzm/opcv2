import { useEffect, useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type ContentArticle } from "../lib/contentApi";
import { CdkTopNav } from "./AnalysisPage";

type InsightsPageProps = {
  variant?: "list" | "detail" | "fileAnalysis";
};

const articleVisuals = ["bot", "map", "target", "chip", "growth", "cloud"] as const;

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
    source: article.source_name || "来源待补充",
    time: formatInsightTime(article.source_published_at || article.published_at || article.created_at),
    tags: article.tags?.length ? article.tags : [article.category || "未分类"],
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
  const [category, setCategory] = useState("");
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi
      .listArticles({ category: category || undefined, q: search, limit: compact ? 6 : 20 })
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
  }, [category, compact, search]);

  const visibleArticles = apiArticles;
  const focusArticles = visibleArticles.slice(0, 3);

  return (
    <>
      <header className="insights-title">
        <span className="insights-title-icon" aria-hidden="true">•••</span>
        <div>
          <h1>咨询通</h1>
          <p>查看后台发布的资讯、来源和引用，并基于已保存证据提问</p>
          <div className="insights-benefits">
            {["只展示已发布内容", "详情保留原始来源", "AI回答返回引用"].map((item) => (
              <span key={item}>✓ {item}</span>
            ))}
          </div>
        </div>
      </header>

      <section className="insights-filter-bar" aria-label="资讯筛选">
        <label className="insights-search"><span>搜索</span><input aria-label="搜索资讯目录" onChange={(event) => setSearch(event.target.value)} value={search} /></label>
        <div className="insights-tabs" role="tablist" aria-label="资讯分类">
          {["精选", "融资", "获客案例", "工具更新", "政策风险", "行业趋势"].map((item) => (
            <button className={(item === "精选" ? category === "" : category === item) ? "active" : ""} key={item} onClick={() => setCategory(item === "精选" ? "" : item)} role="tab" type="button">
              {item}
            </button>
          ))}
        </div>
      </section>

      <section className="insights-focus" aria-label="最新发布">
        <div>
          <h2>最新发布</h2>
          <small>当前 {visibleArticles.length} 条</small>
        </div>
        {focusArticles.length === 0 ? (
          <div className="insights-focus-empty" role="status">
            <strong>暂无已发布资讯</strong>
            <span>内容由后台发布后会在这里展示。</span>
          </div>
        ) : focusArticles.map((article) => (
            <article key={article.slug}>
              <span className={`focus-art ${article.visual === "map" ? "globe" : article.visual === "target" ? "cube" : "chart"}`} aria-hidden="true" />
              <strong>{article.title}</strong>
              <p>{article.summary}</p>
            </article>
          ))}
        <Link to="/insights/file-analysis">基于已发布资讯提问 ›</Link>
      </section>

      <section className={`insight-list-card ${compact ? "compact" : ""}`} aria-label="资讯列表">
        {error && <p className="form-error" role="alert">{error}</p>}
        {visibleArticles.length === 0 ? (
          <div className="insights-empty-state" role="status">
            <span className="insights-empty-visual" aria-hidden="true" />
            <div>
              <h2>暂无资讯数据</h2>
              <p>接口当前没有返回已发布资讯。请等待内容管理员完成入库和发布。</p>
            </div>
            <div className="insights-empty-actions">
              <Link to="/insights/file-analysis">基于已发布资讯提问</Link>
              <Link to="/tools">查看相关工具</Link>
            </div>
          </div>
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

    </>
  );
}

function InsightDetail() {
  const location = useLocation();
  const articleSlug = new URLSearchParams(location.search).get("article") || "";
  const [article, setArticle] = useState<ContentArticle | null>(null);
  const [bookmarked, setBookmarked] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!articleSlug) {
      setArticle(null);
      setError("");
      return;
    }
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
            <Link className="primary" to="/insights/file-analysis">选择资讯后提问</Link>
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
  const publishedAt = formatInsightTime(article.source_published_at || article.published_at || article.created_at);

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
          <span>来源：{article.source_name || "待补充"}</span>
          <time>发布时间：{publishedAt}</time>
          <span>作者：{article.author || "待补充"}</span>
        </div>
        <div className="insight-tag-row">
          {(article.tags?.length ? article.tags : [article.category || "未分类"]).map((tag) => <span key={tag}>{tag}</span>)}
        </div>
        <div className="insight-detail-actions">
          <Link className="primary" to={`/insights/file-analysis?article=${article.slug}`}>基于这篇资讯提问</Link>
          <button aria-label={bookmarked ? "取消收藏资讯" : "收藏资讯"} onClick={toggleBookmark} type="button">
            {bookmarked ? "★ 已收藏" : "☆ 收藏"}
          </button>
          {article.source_url && <a href={article.source_url}>原始来源 ↗</a>}
        </div>
        {error && <p className="form-error" role="alert">{error}</p>}
      </header>
      <div className="insight-article-layout">
        <div className="insight-article">
          <span className="insight-article-cover bot" aria-hidden="true" />
          <section>
            <h2>正文</h2>
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
            <strong>已保存引用</strong>
            <div>
              {article.citations?.length ? article.citations.map((citation) => (
                <a href={citation.source_url} key={citation.id}>{citation.source_name}：{citation.label}</a>
              )) : <span>暂无结构化引用</span>}
            </div>
          </section>
        </aside>
      </div>
    </article>
  );
}

function InsightsCopilot({ variant }: { variant: NonNullable<InsightsPageProps["variant"]> }) {
  const isFileAnalysis = variant === "fileAnalysis";
  return (
    <aside className={`learning-copilot insights-copilot ${isFileAnalysis ? "analysis-open" : ""}`} aria-label="智活 Copilot 咨询助手">
      <header className="insights-ai-head">
        <span className="insights-bot-art" aria-hidden="true" />
        <div>
          <strong><span aria-hidden="true">✦</span> 资讯来源助手</strong>
          <p>问答只使用你选择的已发布资讯及其后台保存引用。</p>
        </div>
      </header>
      {isFileAnalysis ? (
        <FileAnalysisChat />
      ) : (
        <div className="learning-chat insights-chat">
          <article><span className="ai-avatar">A</span><p>打开资讯详情可核对原始来源；进入“基于资讯提问”可获得带引用的回答。</p></article>
        </div>
      )}
      <nav className="learning-copilot-actions" aria-label="咨询助手快捷入口">
        <Link to="/insights">浏览已发布资讯 <span aria-hidden="true">›</span></Link>
        <Link to="/insights/file-analysis">基于资讯提问 <span aria-hidden="true">›</span></Link>
      </nav>
    </aside>
  );
}

function FileAnalysisChat() {
  const location = useLocation();
  const requestedArticle = new URLSearchParams(location.search).get("article") || "";
  const [articles, setArticles] = useState<ContentArticle[]>([]);
  const [selectedArticle, setSelectedArticle] = useState(requestedArticle);
  const [question, setQuestion] = useState("");
  const [answer, setAnswer] = useState<Awaited<ReturnType<typeof contentApi.answerInsightQuestion>> | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi.listArticles({ limit: 20 }).then((payload) => {
      if (!active) return;
      setArticles(payload.articles);
      const requestedExists = payload.articles.some((article) => article.slug === requestedArticle);
      setSelectedArticle(requestedExists ? requestedArticle : payload.articles[0]?.slug || "");
    }).catch((error) => {
      if (active) setError(apiErrorMessage(error, "暂时无法读取可提问资讯"));
    });
    return () => { active = false; };
  }, [requestedArticle]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedArticle || !question.trim() || pending) return;
    setPending(true);
    setError("");
    try {
      setAnswer(await contentApi.answerInsightQuestion(question, [selectedArticle]));
    } catch (error) {
      setAnswer(null);
      setError(apiErrorMessage(error, "暂时无法生成带引用回答"));
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="learning-chat insights-chat">
      <form onSubmit={submit}>
        <label>选择资讯
          <select aria-label="问答资讯" onChange={(event) => setSelectedArticle(event.target.value)} value={selectedArticle}>
            <option value="">请选择已发布资讯</option>
            {articles.map((article) => <option key={article.slug} value={article.slug}>{article.title}</option>)}
          </select>
        </label>
        <label>问题
          <textarea aria-label="资讯问答问题" onChange={(event) => setQuestion(event.target.value)} value={question} />
        </label>
        <button disabled={!selectedArticle || !question.trim() || pending} type="submit">{pending ? "回答中..." : "生成带引用回答"}</button>
      </form>
      {articles.length === 0 && !error && <p role="status">暂无带来源的已发布资讯，当前不能生成回答。</p>}
      {error && <p className="form-error" role="alert">{error}</p>}
      {answer && (
        <article className="analysis-answer">
          <span className="ai-avatar">A</span>
          <div>
            <p>{answer.answer}</p>
            <small>{answer.disclaimer}</small>
            {answer.assumptions.length > 0 && <ul>{answer.assumptions.map((item) => <li key={item}>{item}</li>)}</ul>}
            <section className="insights-reference-panel" aria-label="出处引用">
              <h2>出处引用</h2>
              <div>{answer.citations.map((citation) => (
                <a href={citation.source_url} key={citation.id}><b>{citation.source_name}</b><span>{citation.label}</span><small>{citation.excerpt}</small></a>
              ))}</div>
            </section>
          </div>
        </article>
      )}
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
