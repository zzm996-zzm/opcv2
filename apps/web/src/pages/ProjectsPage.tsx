import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const categories = ["全部", "AI应用", "内容电商", "企业服务", "本地生活", "智能硬件"];

const projectStats = [
  ["项目货架", "128"],
  ["可免费浏览", "36"],
  ["本周新增", "12"]
] as const;

const projects = [
  {
    title: "智能客服系统",
    desc: "面向中小企业的客服自动化项目，适合已有行业资源或交付经验的团队。",
    budget: "3-8万",
    level: "中等",
    tags: ["SaaS", "企业服务", "AI应用"],
    path: ["验证痛点", "搭建 MVP", "行业模板", "渠道试销"],
    fit: "适合有 B 端销售或客服交付经验的团队"
  },
  {
    title: "AI 短视频代运营",
    desc: "用脚本、剪辑和投放工具提升内容生产效率，面向本地商家和小品牌。",
    budget: "1-3万",
    level: "较低",
    tags: ["短视频", "营销", "服务"],
    path: ["选细分行业", "搭建素材库", "跑样板客户", "复用 SOP"],
    fit: "适合内容、电商或私域运营背景"
  },
  {
    title: "AI 智能硬件陪跑",
    desc: "围绕硬件产品定义、供应链验证与渠道冷启动提供 AI 辅助方案。",
    budget: "10万+",
    level: "较高",
    tags: ["硬件", "供应链", "品牌"],
    path: ["定义场景", "验证样机", "测算渠道", "小批量试销"],
    fit: "适合有产品、供应链或行业客户资源的人"
  }
];

function ProjectsPage() {
  return (
    <V4PageShell>
      <section className="module-page projects-page" aria-label="项目超市">
        <div className="page-title-row">
          <div>
            <h1>项目超市</h1>
            <p>像逛货架一样找项目，再用 AI 匹配你的能力、预算与起步路径</p>
          </div>
          <Link className="module-primary-action" to="/tasks">生成落地任务</Link>
        </div>

        <section className="module-overview-card projects-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">项目货架</span>
            <h2>从项目摘要开始筛选，再进入 AI 匹配和落地任务</h2>
            <p>按行业、预算、难度快速浏览项目，先看清楚适配度，再决定是否进入完整拆解。</p>
            <div className="module-stat-strip">
              {projectStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact">
            <label htmlFor="project-match">AI 匹配项目</label>
            <textarea id="project-match" aria-label="描述你的项目条件" placeholder="输入你的优势、预算、目标..." />
            <button type="button">开始 AI 匹配</button>
          </form>
        </section>

        <div className="module-filter-card">
          <div className="module-search">
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索项目" placeholder="搜索项目名 / 行业 / 标签" />
          </div>
          <div className="module-chip-row" aria-label="项目分类">
            {categories.map((category, index) => (
              <button className={index === 0 ? "active" : ""} key={category} type="button">{category}</button>
            ))}
          </div>
        </div>

        <section className="module-list-card">
          <div className="module-section-head">
            <div>
              <h2>项目货架</h2>
              <p>免费浏览摘要，完整拆解可进入详情页</p>
            </div>
            <Link to="/projects">查看全部 ›</Link>
          </div>
          <div className="project-grid">
          {projects.map((project) => (
            <article className="project-card" key={project.title}>
              <header>
                <div>
                  <h2>{project.title}</h2>
                  <p>{project.desc}</p>
                </div>
                <button aria-label={`收藏${project.title}`} type="button">☆</button>
              </header>
              <div className="project-meta">
                <span>预算 {project.budget}</span>
                <span>难度 {project.level}</span>
              </div>
              <div className="tool-tags">
                {project.tags.map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <div className="project-path" aria-label={`${project.title}起步路径`}>
                {project.path.map((step, index) => (
                  <span key={step}><b>{index + 1}</b>{step}</span>
                ))}
              </div>
              <footer>
                <small>{project.fit}</small>
                <Link to="/projects">查看详情</Link>
              </footer>
            </article>
          ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

export default ProjectsPage;
