import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const geoStats = [
  ["AI引用覆盖", "42%"],
  ["目标关键词", "36"],
  ["待补内容", "14"],
  ["潜在线索", "128"]
] as const;

const keywords = [
  {
    query: "智能客服系统怎么选",
    intent: "选型对比",
    coverage: "已被 3 个 AI 回答引用",
    score: "86",
    action: "补充价格和实施周期"
  },
  {
    query: "教育培训私域转化工具",
    intent: "场景方案",
    coverage: "引用不足",
    score: "62",
    action: "新增行业案例页"
  },
  {
    query: "AI客服和企微怎么结合",
    intent: "方案理解",
    coverage: "未进入推荐答案",
    score: "48",
    action: "生成问答型内容"
  }
] as const;

const engines = [
  ["ChatGPT", "38%", "需要补案例"],
  ["豆包", "51%", "表现稳定"],
  ["Kimi", "27%", "引用不足"],
  ["通义千问", "45%", "可继续提升"]
] as const;

const contentTasks = [
  ["选型页", "智能客服系统怎么选？从成本、场景和集成能力拆解", "高", "今天"],
  ["对比页", "AI客服 + 企微 SCRM 与传统客服系统差异", "高", "明天"],
  ["案例页", "教育培训机构如何用 AI 客服提高私域转化", "中", "06-25"],
  ["问答页", "AI客服上线前需要准备哪些数据和流程？", "中", "06-26"]
] as const;

const leadSignals = [
  ["高意向问题", "“智能客服系统怎么选” 下游搜索量上升 18%"],
  ["竞品空位", "3 个竞品没有覆盖教育培训私域场景"],
  ["内容缺口", "价格、实施周期、企微联动是 AI 回答中的弱项"],
  ["线索入口", "建议在选型页加入诊断表单和顾问预约入口"]
] as const;

const roadmap = [
  ["1", "锁定问题", "从客户搜索、竞品内容和 AI 回答里提取高意向问题"],
  ["2", "生成内容", "围绕选型、对比、案例、问答建立可引用内容阵地"],
  ["3", "优化引用", "补充结构化事实、案例、价格和可信来源"],
  ["4", "承接线索", "把高意向访问导入诊断表单、CRM 和跟进任务"]
] as const;

function GeoAcquisitionPage() {
  return (
    <V4PageShell className="geo-acquisition-shell">
      <section className="module-page geo-acquisition-page" aria-label="GEO获客">
        <div className="page-title-row">
          <div>
            <h1>GEO获客</h1>
            <p>围绕 AI 搜索、答案引用和高意向问题建立内容阵地，让客户在提问时更容易看到你</p>
          </div>
          <button className="module-primary-action" type="button">生成GEO方案</button>
        </div>

        <section className="module-overview-card geo-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">智能客服系统 · AI 搜索收录</span>
            <h2>把高意向问题变成持续获客入口</h2>
            <p>系统会拆解客户在 AI 搜索里会问什么、哪些答案已经引用你、哪里还缺可信内容，并生成可执行的内容和线索承接任务。</p>
            <div className="module-stat-strip">
              {geoStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact geo-query-card">
            <label htmlFor="geo-target">输入产品 / 客群 / 场景</label>
            <textarea
              id="geo-target"
              aria-label="输入GEO获客目标"
              placeholder="例如：面向教育培训机构的智能客服系统，希望覆盖选型、价格、企微联动和私域转化问题..."
            />
            <button type="button">分析 AI 搜索机会</button>
          </form>
        </section>

        <section className="geo-workbench">
          <div className="geo-coverage-card">
            <div className="module-section-head">
              <div>
                <h2>AI 搜索覆盖</h2>
                <p>追踪不同 AI 回答中是否出现品牌、案例和关键卖点</p>
              </div>
            </div>
            <div className="geo-engine-grid">
              {engines.map(([name, percent, status]) => (
                <article key={name}>
                  <strong>{name}</strong>
                  <span>{percent}</span>
                  <i style={{ width: percent }} aria-hidden="true" />
                  <small>{status}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="geo-lead-card" aria-label="线索机会">
            <h2>线索机会</h2>
            {leadSignals.map(([title, detail]) => (
              <article key={title}>
                <strong>{title}</strong>
                <small>{detail}</small>
              </article>
            ))}
          </aside>
        </section>

        <section className="geo-keyword-section">
          <div className="module-section-head">
            <div>
              <h2>关键词机会池</h2>
              <p>按客户提问意图排序，优先补最容易转化的内容缺口</p>
            </div>
            <div className="module-chip-row compact">
              {["全部", "选型", "价格", "场景", "竞品对比"].map((view, index) => (
                <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
              ))}
            </div>
          </div>
          <div className="geo-keyword-list">
            {keywords.map((item) => (
              <article key={item.query}>
                <div>
                  <h3>{item.query}</h3>
                  <small>{item.intent} · {item.coverage}</small>
                </div>
                <strong>{item.score}</strong>
                <span>{item.action}</span>
                <Link to="/tasks">生成内容任务</Link>
              </article>
            ))}
          </div>
        </section>

        <section className="geo-lower-grid">
          <div className="geo-content-card">
            <div className="module-section-head">
              <div>
                <h2>内容阵地任务</h2>
                <p>把 AI 能引用的内容拆成页面、标题、优先级和交付日期</p>
              </div>
            </div>
            <div className="geo-task-table">
              {contentTasks.map(([type, title, priority, due]) => (
                <article key={title}>
                  <span>{type}</span>
                  <strong>{title}</strong>
                  <em className={priority === "高" ? "hot" : ""}>{priority}</em>
                  <small>{due}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="geo-roadmap-card" aria-label="GEO执行路径">
            <h2>执行路径</h2>
            {roadmap.map(([step, title, detail]) => (
              <article key={step}>
                <b>{step}</b>
                <span>
                  <strong>{title}</strong>
                  <small>{detail}</small>
                </span>
              </article>
            ))}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default GeoAcquisitionPage;
