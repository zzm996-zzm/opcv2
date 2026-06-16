import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const toolCategories = ["全部", "写作", "绘图", "视频", "营销", "数据", "自动化"];

const toolStats = [
  ["工具收录", "260"],
  ["免费工具", "84"],
  ["方案模板", "18"]
] as const;

const tools = [
  {
    name: "ChatGPT",
    desc: "通用对话、方案生成、文案与代码辅助",
    scene: "用 ChatGPT 快速生成项目调研提纲与任务拆解",
    tags: ["通用助手", "写作", "分析"],
    price: "免费 / 付费",
    platform: "Web / App",
    accent: "blue"
  },
  {
    name: "Midjourney",
    desc: "高质量商业视觉、产品概念图与海报生成",
    scene: "用 Midjourney 生成首批投放素材方向",
    tags: ["绘图", "品牌视觉", "海报"],
    price: "付费",
    platform: "Discord / Web",
    accent: "violet"
  },
  {
    name: "Notion AI",
    desc: "知识库、项目文档和团队协作文档整理",
    scene: "用 Notion AI 沉淀项目资料和 SOP",
    tags: ["知识库", "协作", "文档"],
    price: "付费",
    platform: "Web",
    accent: "cyan"
  },
  {
    name: "剪映专业版",
    desc: "短视频剪辑、字幕、模板和营销视频生产",
    scene: "用剪映批量产出产品讲解短视频",
    tags: ["视频", "内容营销", "字幕"],
    price: "免费 / 付费",
    platform: "Desktop / App",
    accent: "green"
  }
];

function ToolsPage() {
  return (
    <V4PageShell>
      <section className="module-page tools-page" aria-label="工具箱">
        <div className="page-title-row">
          <div>
            <h1>工具箱</h1>
            <p>浏览 AI 工具库，并按场景找到解决方案</p>
          </div>
          <Link className="module-primary-action" to="/tasks">存为任务</Link>
        </div>

        <section className="module-overview-card tools-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">工具库</span>
            <h2>按场景浏览工具，再把工具组合沉淀成任务</h2>
            <p>保留工具库的浏览效率，突出“用什么解决什么”的场景化判断。</p>
            <div className="module-stat-strip">
              {toolStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact">
            <label htmlFor="tool-match">AI 找工具 / 方案</label>
            <textarea id="tool-match" aria-label="描述你的工具需求" placeholder="输入业务卡点或想完成的任务..." />
            <button type="button">生成工具方案</button>
          </form>
        </section>

        <div className="module-filter-card">
          <div className="module-search">
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索工具" placeholder="搜索工具名 / 用途 / 标签" />
          </div>
          <div className="module-chip-row" aria-label="工具分类">
            {toolCategories.map((category, index) => (
              <button className={index === 0 ? "active" : ""} key={category} type="button">{category}</button>
            ))}
          </div>
        </div>

        <section className="module-list-card">
          <div className="module-section-head">
            <div>
              <h2>推荐工具</h2>
              <p>先看用途和场景，再进入详情页</p>
            </div>
            <Link to="/tools">查看全部 ›</Link>
          </div>
          <div className="tool-grid">
          {tools.map((tool) => (
            <article className="tool-card" key={tool.name}>
              <div className={`tool-logo ${tool.accent}`} aria-hidden="true">{tool.name.slice(0, 1)}</div>
              <div>
                <h2>{tool.name}</h2>
                <p>{tool.desc}</p>
              </div>
              <div className="tool-tags">
                {tool.tags.map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <div className="tool-scene">
                <strong>用 X 解决 Y</strong>
                <small>{tool.scene}</small>
              </div>
              <footer>
                <span>{tool.price}</span>
                <span>{tool.platform}</span>
                <Link to="/tools">查看详情</Link>
              </footer>
            </article>
          ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

export default ToolsPage;
