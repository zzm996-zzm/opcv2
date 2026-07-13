import { useEffect, useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
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

const categories = ["精选", "最新", "热门"] as const;
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
    tags: tool.tags?.length ? tool.tags : [tool.category || "未分类"],
    price: tool.price_label || "价格待补充",
    platform: tool.platforms?.join(" / ") || "平台待补充",
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

  const visibleTools = full ? apiTools : apiTools.slice(0, 6);
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

      <section className="toolhub-hot" aria-label="常见工具场景">
        <header>
          <h2>常见工具场景</h2>
          <p>按目标场景检索已发布的工具目录</p>
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
            <strong>按需求找工具</strong>
            <p>输入目标与场景，从已发布目录中匹配</p>
            <Link to="/tools/recommend">开始匹配 ›</Link>
          </div>
        </aside>

        <div className={`toolhub-grid ${full ? "full" : ""}`} aria-label="工具列表">
          {error && <p className="form-error" role="alert">{error}</p>}
          {visibleTools.length === 0 ? (
            <div className="cdk-toolhub-empty" role="status">没有匹配的工具，换个关键词或分类试试</div>
          ) : visibleTools.map((tool) => <ToolCard key={tool.name} tool={tool} />)}
        </div>
      </section>

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
  const [goal, setGoal] = useState("");
  const [scenario, setScenario] = useState("");
  const [tools, setTools] = useState<DisplayTool[]>([]);
  const [submitted, setSubmitted] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!goal.trim() || !scenario.trim() || pending) return;
    setPending(true);
    setError("");
    try {
      const result = await contentApi.recommendTools({ goal, scenario, limit: 6 });
      setTools(result.tools.map(toDisplayTool));
      setSubmitted(true);
    } catch (error) {
      setTools([]);
      setSubmitted(true);
      setError(apiErrorMessage(error, "暂时无法匹配工具"));
    } finally {
      setPending(false);
    }
  }

  return (
    <>
      <header className="toolhub-title">
        <h1>工具推荐结果</h1>
        <p>根据你明确提交的目标和场景，从已发布工具目录中匹配</p>
      </header>
      <form className="toolhub-recommend" onSubmit={submit}>
        <h2>为你推荐的 AI 工具</h2>
        <label>目标<input aria-label="工具匹配目标" onChange={(event) => setGoal(event.target.value)} value={goal} /></label>
        <label>使用场景<input aria-label="工具使用场景" onChange={(event) => setScenario(event.target.value)} value={scenario} /></label>
        <button disabled={pending || !goal.trim() || !scenario.trim()} type="submit">{pending ? "匹配中..." : "匹配目录工具"}</button>
        {error && <p className="form-error" role="alert">{error}</p>}
        {submitted && tools.length === 0 ? (
          <div className="module-empty-state" role="status">已发布目录中暂无匹配工具</div>
        ) : tools.length > 0 ? (
          <div className="toolhub-grid" aria-label="工具推荐结果">{tools.map((tool) => <ToolCard key={tool.slug} tool={tool} />)}</div>
        ) : <p role="status">提交目标和场景后，仅展示已发布目录中的匹配结果。</p>}
      </form>
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
  const toolSlug = new URLSearchParams(location.search).get("tool") || "";
  const [tool, setTool] = useState<ContentTool | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!toolSlug) {
      setTool(null);
      setError("");
      return;
    }
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
              {(tool.tags?.length ? tool.tags : [category]).map((tag) => <span key={tag}>{tag}</span>)}
            </div>
          </div>
          <div className="toolhub-detail-actions">
            {website ? <a href={website}>访问官网</a> : <span>暂无官网</span>}
            <Link to="/tools/recommend">按需求匹配目录工具 ›</Link>
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
            <p>{tool.use_cases?.length ? tool.use_cases.join("、") : "暂无适用场景说明"}</p>
            <h3>主要功能</h3>
            <p>{tool.features?.length ? tool.features.join("、") : "暂无功能说明"}</p>
            <h3>注意点</h3>
            <p>{tool.limitations?.length ? tool.limitations.join("、") : "暂无注意事项"}</p>
          </article>
          <article>
            <h3>使用步骤</h3>
            <p>暂无使用步骤</p>
            <h3>入口链接</h3>
            <p>{website ? <a href={website}>{website}</a> : "暂无入口链接"}</p>
            <h3>价格信息</h3>
            <p>{tool.price_label || "暂无价格信息"}</p>
            <h3>平台</h3>
            <p>{tool.platforms?.length ? tool.platforms.join("、") : "暂无平台信息"}</p>
            <h3>数据来源</h3>
            <p>{tool.source_url ? <a href={tool.source_url}>{tool.provider_name || tool.source_url}</a> : "暂无来源链接"}</p>
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

function ToolsCopilot({ variant }: { variant: NonNullable<ToolsPageProps["variant"]> }) {
  return (
    <aside className="learning-copilot toolhub-copilot" aria-label="智活 Copilot 工具助手">
      <header className="toolhub-ai-head">
        <div>
          <strong><span aria-hidden="true">✦</span> 工具目录助手</strong>
          <p>目录匹配只使用你提交的目标、场景与后台发布的工具资料。</p>
        </div>
      </header>
      <div className="learning-chat toolhub-chat">
        <article>
          <span className="ai-avatar">A</span>
          <p>{variant === "recommend" ? "请在左侧提交目标和使用场景。匹配结果不会补充目录中不存在的工具或能力。" : "可先浏览工具目录，或进入按需求匹配页提交明确条件。"}</p>
        </article>
      </div>
      <nav className="learning-copilot-actions" aria-label="工具助手快捷入口">
        <Link to="/tools">浏览工具目录 <span aria-hidden="true">›</span></Link>
        <Link to="/tools/recommend">按需求匹配 <span aria-hidden="true">›</span></Link>
      </nav>
      <MiniCopilotForm className="learning-copilot-input" inputAriaLabel="向工具箱 Copilot 提问" />
    </aside>
  );
}

export default ToolsPage;
