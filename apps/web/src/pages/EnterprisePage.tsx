import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const enterpriseStats = [
  ["服务企业", "38"],
  ["平均周期", "12周"],
  ["交付任务", "164"],
  ["续约率", "72%"]
] as const;

const plans = [
  {
    title: "增长团队训练营",
    audience: "适合 10-50 人销售/运营团队",
    price: "¥12万起",
    focus: ["AI工具流搭建", "线索开发 SOP", "周度经营复盘"],
    result: "4 周内搭好从获客到 CRM 的标准动作"
  },
  {
    title: "AI获客陪跑",
    audience: "适合正在验证新项目的业务团队",
    price: "¥18万起",
    focus: ["GEO内容矩阵", "竞品监测", "高意向线索池"],
    result: "8 周内跑通目标客户画像和首批成交机会"
  },
  {
    title: "企业定制系统",
    audience: "适合需要私有流程和数据看板的企业",
    price: "定制报价",
    focus: ["需求诊断", "系统集成", "交付培训"],
    result: "12 周内完成流程定制、数据接入和团队上线"
  }
] as const;

const deliveryBoard = [
  ["诊断中", "4", "业务访谈、数据梳理、流程盘点"],
  ["方案中", "7", "项目路径、ROI 测算、资源排期"],
  ["交付中", "12", "工具配置、团队训练、周度复盘"],
  ["复盘中", "5", "指标验收、续约判断、二期规划"]
] as const;

const milestones = [
  ["第1周", "企业诊断", "厘清业务目标、客户画像、组织分工和数据现状"],
  ["第2-4周", "流程搭建", "上线 AI 线索开发、CRM 跟进、任务中心和经营仪表盘"],
  ["第5-8周", "团队陪跑", "周度复盘关键客户、竞品动态、GEO 内容与成交动作"],
  ["第9-12周", "验收迭代", "沉淀 SOP、训练团队负责人、规划下一阶段增长实验"]
] as const;

const cases = [
  ["连锁教育集团", "客服响应效率提升 43%，新增可跟进商机 86 个"],
  ["职业培训机构", "8 周跑通 GEO 获客内容，试点校区转化率提升 18%"],
  ["企业内训服务商", "完成销售流程标准化，报价周期从 5 天缩短到 2 天"]
] as const;

function EnterprisePage() {
  return (
    <V4PageShell className="enterprise-shell">
      <section className="module-page enterprise-page" aria-label="企业定制化陪跑">
        <div className="page-title-row">
          <div>
            <h1>企业定制化陪跑</h1>
            <p>面向企业团队提供诊断、方案、系统搭建、训练和复盘的一体化增长陪跑</p>
          </div>
          <button className="module-primary-action" type="button">预约企业诊断</button>
        </div>

        <section className="module-overview-card enterprise-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">企业服务 · 定制交付</span>
            <h2>从业务问题到团队上线，陪企业把 AI 增长流程真正跑起来</h2>
            <p>通过企业诊断、工具配置、实战陪跑和交付验收，把项目超市、GEO 获客、AI 线索开发和 CRM 组合成企业自己的增长系统。</p>
            <div className="module-stat-strip">
              {enterpriseStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact enterprise-diagnosis-card">
            <label htmlFor="enterprise-need">描述企业需求</label>
            <textarea
              id="enterprise-need"
              aria-label="描述企业需求"
              placeholder="例如：30人销售团队，希望用 AI 提升线索开发、客户跟进和经营复盘效率..."
            />
            <button type="button">生成诊断提纲</button>
          </form>
        </section>

        <section className="enterprise-workbench">
          <div className="enterprise-plan-card">
            <div className="module-section-head">
              <div>
                <h2>陪跑方案</h2>
                <p>按企业团队规模、增长目标和系统复杂度选择交付模式</p>
              </div>
              <div className="module-chip-row compact">
                {["标准", "获客", "系统", "定制"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>

            <div className="enterprise-plan-grid">
              {plans.map((plan) => (
                <article key={plan.title}>
                  <header>
                    <div>
                      <h3>{plan.title}</h3>
                      <small>{plan.audience}</small>
                    </div>
                    <strong>{plan.price}</strong>
                  </header>
                  <p>{plan.result}</p>
                  <div className="tool-tags">
                    {plan.focus.map((item) => <span key={item}>{item}</span>)}
                  </div>
                  <Link to="/crm">沉淀到CRM</Link>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-delivery-card" aria-label="交付看板">
            <h2>交付看板</h2>
            {deliveryBoard.map(([stage, count, detail]) => (
              <article key={stage}>
                <span>
                  <strong>{stage}</strong>
                  <small>{detail}</small>
                </span>
                <em>{count}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="enterprise-lower-grid">
          <div className="enterprise-milestone-card">
            <div className="module-section-head">
              <div>
                <h2>陪跑里程碑</h2>
                <p>把企业服务拆成可验收、可复盘、可续约的交付节奏</p>
              </div>
            </div>
            <div className="enterprise-milestones">
              {milestones.map(([time, title, detail]) => (
                <article key={time}>
                  <b>{time}</b>
                  <strong>{title}</strong>
                  <small>{detail}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-case-card" aria-label="企业案例">
            <h2>企业案例</h2>
            {cases.map(([company, result]) => (
              <article key={company}>
                <strong>{company}</strong>
                <small>{result}</small>
              </article>
            ))}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default EnterprisePage;
