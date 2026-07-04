import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type ContentTool } from "../lib/contentApi";
import { CdkTopNav } from "./AnalysisPage";

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

const categories = ["精选", "最新", "热门", "收藏"] as const;
const sidebarCategories = ["全部工具", "创业获客", "内容生产", "图片设计", "视频剪辑", "客户管理", "数据分析", "跨境外贸"] as const;

const hotScenarios = [
  ["内容创作加速", "提升内容生产效率", "write"],
  ["社媒增长利器", "扩大品牌影响力", "bars"],
  ["数据洞察分析", "发现增长机会", "pie"],
  ["自动化省时神器", "解放重复性工作", "robot"]
] as const;

function toDisplayTool(tool: ContentTool): DisplayTool {
  return {
    slug: tool.slug,
    name: tool.name,
    desc: tool.description || "AI 工具能力已收录，可进入详情查看适用场景。",
    tags: ["AI工具", tool.category || "已收录", tool.status === "published" ? "公开可见" : "草稿"],
    price: "可用",
    platform: tool.url ? "Web" : "待补充",
    accent: "notion",
    category: tool.category || "办公",
    url: tool.url
  };
}

function ToolsPage({ variant = "library" }: ToolsPageProps) {
  const withCopilot = variant !== "all";

  return (
    <main className="cdk-analysis-page cdk-tools-page">
      <CdkTopNav active="工具箱" />
      <section className={`toolhub-page ${withCopilot ? "with-copilot" : "wide"}`} aria-label="工具箱">
        <main className="toolhub-main">
          {variant === "recommend" && <ToolRecommendation />}
          {variant === "plan" && <ToolPlan />}
          {variant === "detail" && <ToolDetail />}
          {(variant === "library" || variant === "all") && <ToolLibrary full={variant === "all"} />}
        </main>
        {withCopilot && <ToolsCopilot variant={variant} />}
      </section>
    </main>
  );
}

function ToolLibrary({ full }: { full: boolean }) {
  const [apiTools, setApiTools] = useState<DisplayTool[]>([]);
  const [selectedCategory, setSelectedCategory] = useState("全部工具");
  const [selectedTab, setSelectedTab] = useState("精选");
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    setError("");
    contentApi
      .listTools({
        category: selectedCategory === "全部工具" ? undefined : selectedCategory,
        q: search,
        sort: toolSort(selectedTab),
        limit: full ? 100 : 9
      })
      .then((payload) => {
        if (!active) return;
        setApiTools(payload.tools.map(toDisplayTool));
      })
      .catch((error) => {
        if (!active) return;
        setError(apiErrorMessage(error, "暂时无法读取工具库"));
        setApiTools([]);
      });
    return () => {
      active = false;
    };
  }, [full, search, selectedCategory, selectedTab]);

  const visibleTools = full ? apiTools : apiTools.slice(0, 9);
  const visibleScenarios = full ? hotScenarios : hotScenarios.slice(0, 3);

  return (
    <>
      <header className="toolhub-title">
        <h1>工具箱</h1>
        <p>精选全球优质AI工具，助力创业获客与高效增长</p>
      </header>

      <section className="toolhub-search-block" aria-label="工具筛选">
        <label className="toolhub-search">
          <span aria-hidden="true">⌕</span>
          <input
            aria-label="搜索工具"
            onChange={(event) => setSearch(event.target.value)}
            placeholder="搜索工具名，用途或标签"
            value={search}
          />
        </label>
        <div className="toolhub-filter-row">
          <div className="toolhub-tabs" role="tablist" aria-label="工具分类">
            {categories.map((category) => (
              <button
                aria-selected={selectedTab === category}
                className={selectedTab === category ? "active" : ""}
                key={category}
                onClick={() => setSelectedTab(category)}
                role="tab"
                type="button"
              >
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

      <section className="cdk-toolhub-library">
        <aside className="cdk-toolhub-sidebar" aria-label="工具分类">
          {sidebarCategories.map((item) => (
            <button
              className={selectedCategory === item ? "active" : ""}
              key={item}
              onClick={() => setSelectedCategory(item)}
              type="button"
            >
              {item}
            </button>
          ))}
          <div>
            <strong>提交优质工具</strong>
            <p>推荐好工具，帮助更多创业者</p>
            <Link to="/tools/recommend">立即推荐 ›</Link>
          </div>
        </aside>

        <div className={`toolhub-grid ${full ? "full" : ""}`} aria-label="工具列表">
          {error && <p className="form-error" role="alert">{error}</p>}
          {visibleTools.length === 0 ? (
            <div className="module-empty-state" role="status">暂无工具数据</div>
          ) : visibleTools.map((tool) => <ToolCard key={tool.name} tool={tool} />)}
        </div>
      </section>

      <ToolPagination />
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
      const result = favorited
        ? await contentApi.unfavoriteTool(tool.slug)
        : await contentApi.favoriteTool(tool.slug);
      setFavorited(result.favorited);
    } catch (error) {
      setError(apiErrorMessage(error, "收藏失败，请稍后再试"));
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
        <div className="toolhub-tag-row">
          {tool.tags.map((tag) => <span key={tag}>{tag}</span>)}
        </div>
        {error && <small className="form-error" role="alert">{error}</small>}
      </div>
      <button
        className="toolhub-favorite"
        aria-label={`${favorited ? "取消收藏" : "收藏"}${tool.name}`}
        disabled={!tool.slug || pending}
        onClick={toggleFavorite}
        type="button"
      >
        {favorited ? "★" : "♡"}
      </button>
      <footer>
        <span className={tool.price.includes("付费") ? "paid" : "free"}>{tool.price}</span>
        <small>{tool.platform}</small>
        <Link to={tool.slug ? `/tools/detail?tool=${tool.slug}` : "/tools/detail"}>查看详情 ›</Link>
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
      <section className="toolhub-recommend">
        <h2>为你推荐的 AI 工具</h2>
        <div className="module-empty-state" role="status">暂无工具推荐结果</div>
      </section>
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
      <div className="module-empty-state" role="status">暂无工具方案</div>
    </>
  );
}

function ToolDetail() {
  const location = useLocation();
  const toolSlug = new URLSearchParams(location.search).get("tool") || "midjourney";
  const [tool, setTool] = useState<ContentTool | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    setError("");
    contentApi
      .getTool(toolSlug)
      .then((payload) => {
        if (active) setTool(payload);
      })
      .catch((error) => {
        if (!active) return;
        setTool(null);
        setError(apiErrorMessage(error, "暂时无法读取工具详情"));
      });
    return () => {
      active = false;
    };
  }, [toolSlug]);

  if (!tool) {
    return (
      <>
        <nav className="toolhub-breadcrumb" aria-label="工具详情路径">
          <Link to="/tools">工具箱</Link><span>/</span><b>工具详情</b>
        </nav>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="module-empty-state" role="status">暂无工具详情</div>
      </>
    );
  }

  const title = tool.name;
  const description = tool.description || "暂无工具说明";
  const category = tool.category || "未分类";
  const website = tool.url || "";

  return (
    <>
      <nav className="toolhub-breadcrumb" aria-label="工具详情路径">
        <Link to="/tools">工具箱</Link><span>/</span><Link to="/tools/all">{category}</Link><span>/</span><b>{title}</b>
      </nav>
      {error && <p className="form-error" role="alert">{error}</p>}
      <section className="toolhub-detail-hero">
        <div className="toolhub-detail-copy">
          <span className="toolhub-logo midjourney large" aria-hidden="true" />
          <div>
            <h1>{title}</h1>
            <p>{description}</p>
            <div className="toolhub-tag-row">
              {[category, "AI工具", tool?.status === "draft" ? "草稿" : "公开"].map((tag) => <span key={tag}>{tag}</span>)}
            </div>
          </div>
          <div className="toolhub-detail-actions">
            {website ? <a href={website}>访问官网</a> : <span>暂无官网</span>}
            <button type="button">收藏工具</button>
            <Link to="/tools/recommend">让智活 Copilot 评估是否适合我 ›</Link>
          </div>
        </div>
        <figure className="toolhub-detail-visual" aria-label={`${title} 工具能力预览`}>
          <span /><span /><span /><span />
        </figure>
      </section>
      <section className="toolhub-detail-panel">
        <h2>工具介绍</h2>
        <div className="toolhub-detail-grid">
          <article>
            <h3>功能简介</h3>
            <p>{description}</p>
            <h3>适用场景</h3>
            <p>{category}、内容生产、营销素材、增长分析等。</p>
            <h3>优点</h3>
            <p>暂无优点说明</p>
            <h3>注意点</h3>
            <p>暂无注意事项</p>
          </article>
          <article>
            <h3>使用步骤</h3>
            <p>暂无使用步骤</p>
            <h3>入口链接</h3>
            <p>{website ? <a href={website}>{website}</a> : "暂无入口链接"}</p>
            <h3>价格信息</h3>
            <p>暂无价格信息</p>
            <h3>适合人群</h3>
            <p>暂无适合人群说明</p>
          </article>
        </div>
      </section>
      <section className="toolhub-solve">
        <h2>用它解决什么</h2>
        <p>{description}</p>
      </section>
    </>
  );
}

function toolSort(tab: string) {
  if (tab === "最新") return "latest";
  if (tab === "热门") return "hot";
  if (tab === "收藏") return "favorite";
  return "featured";
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
      <header className="toolhub-ai-head">
        <div>
          <strong><span aria-hidden="true">✦</span> AI 工具助手</strong>
          <p>告诉我你想做什么，我会帮你匹配合适工具</p>
        </div>
        <button aria-label="关闭工具助手" type="button">×</button>
      </header>
      <label className="toolhub-ai-input">
        <input aria-label="工具需求输入" placeholder="例如：我想做小红书海报，还想配套文案和数据复盘" />
        <button type="button">开始分析</button>
      </label>
      <div className="learning-chat toolhub-chat">
        {isPlan ? (
          <>
            <article><span className="ai-avatar">A</span><p>暂无工具方案，待推荐接口接入后这里会展示生成结果。</p></article>
          </>
        ) : isRecommend ? (
          <>
            <article><span className="ai-avatar">A</span><p>暂无工具推荐结果，待推荐接口接入后这里会展示匹配工具。</p></article>
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
      <section className="toolhub-ai-results" aria-label="AI 推荐结果">
        <header><h2>AI 推荐结果</h2><button type="button">×</button></header>
        <p className="module-empty-state">暂无AI推荐结果</p>
      </section>
    </aside>
  );
}

export default ToolsPage;
