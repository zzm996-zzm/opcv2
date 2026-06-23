import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const sandboxStats = [
  ["当前胜率", "72%"],
  ["关键风险", "4"],
  ["待验证假设", "9"],
  ["预计周期", "21天"]
] as const;

const scenarioSteps = [
  ["1", "需求强度", "验证客户是否愿意为自动化客服付费", "较强"],
  ["2", "交付难度", "评估 MVP、模型接入和行业知识库搭建成本", "中等"],
  ["3", "获客路径", "比较私域、渠道伙伴和行业展会三种冷启动方式", "可行"],
  ["4", "现金流", "用试点价格、交付周期和续费率测算回本速度", "待压测"]
] as const;

const roleViews = [
  {
    role: "CEO",
    focus: "战略取舍",
    verdict: "先做垂直行业样板，不要一开始做通用客服平台。",
    tone: "blue"
  },
  {
    role: "销售负责人",
    focus: "成交路径",
    verdict: "优先锁定已有客服团队、咨询量高且预算明确的企业。",
    tone: "green"
  },
  {
    role: "交付负责人",
    focus: "履约风险",
    verdict: "知识库整理和机器人兜底流程必须标准化，否则毛利会被吞掉。",
    tone: "amber"
  },
  {
    role: "财务顾问",
    focus: "资金效率",
    verdict: "首批试点应控制在 3 个以内，保证现金流回款快于功能扩张。",
    tone: "violet"
  }
] as const;

const riskRows = [
  ["客户愿意为智能客服持续付费", "高", "进行中", "用 10 家目标客户访谈确认预算和决策链"],
  ["交付周期可控制在 14 天内", "中", "待验证", "制作行业模板和首版 SOP 后做一次试交付"],
  ["大模型成本不会侵蚀毛利", "中", "已缓解", "设置消息额度、缓存常见问答、分层调用模型"],
  ["渠道伙伴能带来稳定线索", "低", "待验证", "筛选 3 家客服外包商或 SaaS 服务商试合作"]
] as const;

const timeline = [
  ["09:30", "已完成机会识别", "智能客服系统适合从企业服务切入"],
  ["09:42", "角色推演完成", "CEO、销售、交付、财务四方给出结论"],
  ["10:05", "生成验证任务", "访谈、SOP、报价和试点名单进入任务中心"]
] as const;

function SandboxPage() {
  return (
    <V4PageShell className="sandbox-shell">
      <section className="module-page sandbox-page" aria-label="商业沙盘">
        <div className="page-title-row">
          <div>
            <h1>商业沙盘</h1>
            <p>用多角色推演、风险压测和现金流测算，判断一个项目是否值得启动</p>
          </div>
          <button className="module-primary-action" type="button">开始推演</button>
        </div>

        <section className="module-overview-card sandbox-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">智能客服系统沙盘</span>
            <h2>先把项目放进压力场，再决定要不要投入资源</h2>
            <p>沙盘会从客户需求、交付能力、获客路径和财务回收四个方向推演，提前暴露风险点，并生成下一步验证任务。</p>
            <div className="module-stat-strip">
              {sandboxStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact sandbox-input-card">
            <label htmlFor="sandbox-brief">输入商业假设</label>
            <textarea
              id="sandbox-brief"
              aria-label="输入商业假设"
              placeholder="例如：面向中小企业做智能客服系统，3 周交付，首批找 10 家试点客户..."
            />
            <button type="button">生成沙盘</button>
          </form>
        </section>

        <section className="sandbox-grid">
          <div className="sandbox-main-card">
            <div className="module-section-head">
              <div>
                <h2>推演路径</h2>
                <p>当前按商业可行性优先级排序</p>
              </div>
              <div className="module-chip-row compact">
                {["机会", "交付", "增长", "财务"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>
            <div className="sandbox-step-list">
              {scenarioSteps.map(([number, title, detail, status]) => (
                <article key={title}>
                  <b>{number}</b>
                  <span>
                    <strong>{title}</strong>
                    <small>{detail}</small>
                  </span>
                  <em>{status}</em>
                </article>
              ))}
            </div>
          </div>

          <aside className="sandbox-side-card" aria-label="沙盘结论">
            <h2>本轮建议</h2>
            <strong>小范围试点，先卖行业样板</strong>
            <p>商业机会成立，但不建议直接做通用平台。先选择一个咨询量高、服务流程明确的行业做样板，再复用到第二行业。</p>
            <Link to="/tasks">生成验证任务</Link>
          </aside>
        </section>

        <section className="sandbox-role-section">
          <div className="module-section-head">
            <div>
              <h2>多角色观点</h2>
              <p>让不同角色提前提出反对意见，避免只看增长故事</p>
            </div>
          </div>
          <div className="sandbox-role-grid">
            {roleViews.map((view) => (
              <article className={view.tone} key={view.role}>
                <span>{view.role}</span>
                <small>{view.focus}</small>
                <p>{view.verdict}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="sandbox-lower-grid">
          <div className="sandbox-risk-card">
            <div className="module-section-head">
              <div>
                <h2>风险假设矩阵</h2>
                <p>风险越高，越需要先做低成本验证</p>
              </div>
            </div>
            <div className="sandbox-risk-table">
              {riskRows.map(([assumption, level, status, action]) => (
                <article key={assumption}>
                  <strong>{assumption}</strong>
                  <span className={level === "高" ? "high" : level === "中" ? "mid" : ""}>{level}</span>
                  <em>{status}</em>
                  <small>{action}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="sandbox-timeline-card" aria-label="推演记录">
            <h2>推演记录</h2>
            {timeline.map(([time, title, detail]) => (
              <article key={title}>
                <time>{time}</time>
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

export default SandboxPage;
