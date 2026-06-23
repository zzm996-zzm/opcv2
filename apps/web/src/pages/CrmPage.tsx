import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const crmStats = [
  ["客户总数", "146"],
  ["本月新增", "28"],
  ["跟进中商机", "¥128万"],
  ["预计成交", "¥42万"]
] as const;

const customers = [
  {
    name: "星桥教育集团",
    owner: "张婧",
    value: "¥36万",
    stage: "方案演示",
    health: "高意向",
    next: "今天 14:00 发送智能客服实施方案",
    tags: ["连锁教育", "企微转化", "预算明确"]
  },
  {
    name: "橙果职业培训",
    owner: "林航",
    value: "¥18万",
    stage: "需求确认",
    health: "可推进",
    next: "明天 10:30 预约增长负责人复盘",
    tags: ["成人教育", "招生咨询", "多校区"]
  },
  {
    name: "领航企业内训",
    owner: "王宁",
    value: "¥24万",
    stage: "报价评估",
    health: "待决策",
    next: "06-25 补充企业内训行业案例",
    tags: ["B端服务", "销售扩编", "内容增长"]
  }
] as const;

const pipelineStages = [
  ["新线索", "46", "18%"],
  ["需求确认", "32", "28%"],
  ["方案演示", "18", "44%"],
  ["报价谈判", "9", "64%"],
  ["已成交", "6", "100%"]
] as const;

const followBoard = [
  ["今日必须处理", "12", "高意向客户超过 24 小时未跟进"],
  ["待补材料", "7", "报价、案例、ROI 测算需要同步"],
  ["需要协同", "5", "交付、产品或老板介入判断"]
] as const;

const activities = [
  ["09:40", "星桥教育集团", "客户查看了智能客服报价单，建议追加 ROI 测算"],
  ["11:20", "橙果职业培训", "线索评分提升到 84，新增校区扩张信号"],
  ["14:00", "领航企业内训", "待发送行业案例，当前处于报价评估阶段"]
] as const;

function CrmPage() {
  return (
    <V4PageShell className="crm-shell">
      <section className="module-page crm-page" aria-label="CRM客户管理">
        <div className="page-title-row">
          <div>
            <h1>CRM客户管理</h1>
            <p>统一管理从 AI 线索开发进入的客户、商机阶段、跟进动作和成交预测</p>
          </div>
          <button className="module-primary-action" type="button">新建客户</button>
        </div>

        <section className="module-overview-card crm-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">客户资产 · 商机推进</span>
            <h2>把线索、跟进、报价和成交进度放进一张经营视图</h2>
            <p>CRM 会承接 AI 线索开发的企业信息，记录每次触达、下一步动作、负责人和预计金额，减少客户跟进断层。</p>
            <div className="module-stat-strip">
              {crmStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact crm-quick-card">
            <label htmlFor="crm-note">快速记录跟进</label>
            <textarea
              id="crm-note"
              aria-label="快速记录跟进"
              placeholder="例如：星桥教育集团今天确认预算，希望本周看到智能客服部署周期和报价..."
            />
            <button type="button">AI整理为客户记录</button>
          </form>
        </section>

        <section className="crm-workbench">
          <div className="crm-customer-card">
            <div className="module-section-head">
              <div>
                <h2>客户列表</h2>
                <p>按商机金额、阶段和跟进时效排序，优先处理最可能成交的客户</p>
              </div>
              <div className="module-chip-row compact">
                {["全部", "高意向", "本周跟进", "待报价"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>

            <div className="crm-customer-list">
              {customers.map((customer) => (
                <article key={customer.name}>
                  <header>
                    <div>
                      <h3>{customer.name}</h3>
                      <small>负责人 {customer.owner} · {customer.stage}</small>
                    </div>
                    <strong>{customer.value}</strong>
                  </header>
                  <p>{customer.next}</p>
                  <div className="tool-tags">
                    {customer.tags.map((tag) => <span key={tag}>{tag}</span>)}
                  </div>
                  <footer>
                    <span className={customer.health === "高意向" ? "hot" : ""}>{customer.health}</span>
                    <Link to="/leads">查看来源</Link>
                  </footer>
                </article>
              ))}
            </div>
          </div>

          <aside className="crm-follow-card" aria-label="跟进看板">
            <h2>跟进看板</h2>
            {followBoard.map(([title, count, detail]) => (
              <article key={title}>
                <span>
                  <strong>{title}</strong>
                  <small>{detail}</small>
                </span>
                <em>{count}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="crm-lower-grid">
          <div className="crm-pipeline-card">
            <div className="module-section-head">
              <div>
                <h2>成交漏斗</h2>
                <p>查看客户从新线索到成交的阶段分布和推进转化率</p>
              </div>
            </div>
            <div className="crm-pipeline">
              {pipelineStages.map(([stage, count, percent]) => (
                <article key={stage}>
                  <span>
                    <strong>{stage}</strong>
                    <small>{count} 家客户</small>
                  </span>
                  <b style={{ width: percent }} aria-hidden="true" />
                  <em>{percent}</em>
                </article>
              ))}
            </div>
          </div>

          <aside className="crm-activity-card" aria-label="客户动态">
            <h2>客户动态</h2>
            {activities.map(([time, company, detail]) => (
              <article key={`${time}-${company}`}>
                <time>{time}</time>
                <strong>{company}</strong>
                <small>{detail}</small>
              </article>
            ))}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default CrmPage;
