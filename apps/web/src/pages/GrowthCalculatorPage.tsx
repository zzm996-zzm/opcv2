import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { growthApi, type GrowthModel } from "../lib/growthApi";

const growthStats = [
  ["预计月收入", "¥18.6万"],
  ["获客成本", "¥42"],
  ["回本周期", "17天"],
  ["净利润率", "38%"]
] as const;

const assumptions = [
  ["月访问量", "24,000", "来自内容、社群和投放入口"],
  ["线索转化率", "6.8%", "访问到留资或预约咨询"],
  ["成交转化率", "14%", "线索到首单客户"],
  ["平均客单价", "¥820", "课程、服务包或 SaaS 首购"]
] as const;

const funnelSteps = [
  { label: "曝光", value: "180,000", percent: 100, note: "内容矩阵 + 搜索收录" },
  { label: "访问", value: "24,000", percent: 62, note: "落地页与工具入口" },
  { label: "线索", value: "1,632", percent: 34, note: "表单 / 私域 / 预约" },
  { label: "成交", value: "228", percent: 18, note: "顾问跟进与限时权益" }
] as const;

const scenarios = [
  {
    name: "保守方案",
    revenue: "¥9.8万",
    cost: "¥2.7万",
    margin: "27%",
    highlight: "适合冷启动：少投放，多依赖内容和社群转化。"
  },
  {
    name: "标准方案",
    revenue: "¥18.6万",
    cost: "¥5.1万",
    margin: "38%",
    highlight: "当前推荐：投放验证关键词，私域承接高意向线索。"
  },
  {
    name: "进攻方案",
    revenue: "¥31.4万",
    cost: "¥10.8万",
    margin: "34%",
    highlight: "适合预算充足：快速放量，但需要客服和交付能力同步扩容。"
  }
] as const;

const monthlyForecast = [
  ["第1月", "¥8.4万", "验证渠道", 28],
  ["第2月", "¥13.9万", "优化转化", 44],
  ["第3月", "¥18.6万", "稳定投放", 60],
  ["第4月", "¥24.8万", "扩大渠道", 78],
  ["第5月", "¥29.2万", "复购加成", 90]
] as const;

const costItems = [
  ["内容生产", "¥12,000", "短视频、文章、案例页"],
  ["投放预算", "¥26,000", "搜索词和信息流测试"],
  ["工具订阅", "¥3,200", "线索、CRM、自动化工具"],
  ["交付人力", "¥9,800", "顾问跟进与客户成功"]
] as const;

const actionItems = [
  "把客单价从 ¥820 提升到 ¥980，利润率可增加 6 个点",
  "优先优化线索到成交转化率，比单纯买流量更划算",
  "把高意向线索同步到 CRM，并设置 24 小时跟进提醒"
] as const;

function formatCurrency(value: number) {
  return `¥${new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value)}`;
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

function formatPercent(value: number) {
  return `${(value * 100).toFixed(1)}%`;
}

function statsForModel(model: GrowthModel) {
  return [
    ["预计月收入", formatCurrency(model.result.monthly_revenue)],
    ["获客成本", formatCurrency(model.assumptions.acquisition_cost)],
    ["回本周期", `${model.result.payback_days}天`],
    ["净利润率", formatPercent(model.result.net_margin)]
  ] as const;
}

function assumptionsForModel(model: GrowthModel) {
  return [
    ["月访问量", formatNumber(model.assumptions.monthly_visits), "来自内容、社群和投放入口"],
    ["线索转化率", formatPercent(model.assumptions.lead_rate), "访问到留资或预约咨询"],
    ["成交转化率", formatPercent(model.assumptions.deal_rate), "线索到首单客户"],
    ["平均客单价", formatCurrency(model.assumptions.average_order), "课程、服务包或 SaaS 首购"]
  ] as const;
}

function GrowthCalculatorPage() {
  const [latestModel, setLatestModel] = useState<GrowthModel | null>(null);
  const [loadError, setLoadError] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let active = true;
    growthApi
      .listModels()
      .then((payload) => {
        if (!active) return;
        setLatestModel(payload.models[0] ?? null);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setLatestModel(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取测算模型"));
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleStats = latestModel ? statsForModel(latestModel) : growthStats;
  const visibleAssumptions = latestModel ? assumptionsForModel(latestModel) : assumptions;
  const modelName = latestModel?.name ?? "智能客服系统 · 标准方案";

  async function saveModel() {
    if (isSaving) return;
    setIsSaving(true);
    try {
      const model = await growthApi.createModel({
        name: "智能客服系统 · 标准方案",
        monthlyVisits: 24000,
        leadRate: 0.068,
        dealRate: 0.14,
        averageOrder: 820,
        acquisitionCost: 42,
        deliveryCost: 260
      });
      setLatestModel(model);
    } catch {
      // Keep the current static/model values visible; error presentation can be centralized later.
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <V4PageShell className="growth-calculator-shell">
      <section className="module-page growth-calculator-page" aria-label="增长测算">
        <div className="page-title-row">
          <div>
            <h1>增长测算</h1>
            <p>用访问量、转化率、客单价、获客成本和交付成本，提前算清楚增长动作的收入和利润边界</p>
          </div>
          <button className="module-primary-action" disabled={isSaving} onClick={() => void saveModel()} type="button">
            {isSaving ? "保存中..." : "保存测算模型"}
          </button>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}

        <section className="module-overview-card growth-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">{modelName}</span>
            <h2>先算清每一块钱能带来多少真实增长</h2>
            <p>当前模型以内容获客、私域承接和顾问成交为核心路径，帮助你判断该加预算、调价格，还是先优化转化率。</p>
            <div className="module-stat-strip">
              {visibleStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="growth-input-card" aria-label="核心测算假设">
            <strong>核心假设</strong>
            {visibleAssumptions.map(([label, value, helper]) => (
              <label key={label}>
                <span>{label}</span>
                <input aria-label={label} readOnly value={value} />
                <small>{helper}</small>
              </label>
            ))}
          </form>
        </section>

        <section className="growth-workbench">
          <div className="growth-funnel-card">
            <div className="module-section-head">
              <div>
                <h2>增长漏斗</h2>
                <p>从曝光到成交拆出每一步损耗，定位最值得优化的位置</p>
              </div>
            </div>
            <div className="growth-funnel">
              {funnelSteps.map((step) => (
                <article key={step.label}>
                  <div>
                    <strong>{step.label}</strong>
                    <small>{step.note}</small>
                  </div>
                  <span>{step.value}</span>
                  <i style={{ width: `${step.percent}%` }} aria-hidden="true" />
                </article>
              ))}
            </div>
          </div>

          <aside className="growth-cost-card" aria-label="成本结构">
            <h2>成本结构</h2>
            {costItems.map(([label, value, detail]) => (
              <article key={label}>
                <span>
                  <strong>{label}</strong>
                  <small>{detail}</small>
                </span>
                <em>{value}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="growth-scenario-section">
          <div className="module-section-head">
            <div>
              <h2>情景对比</h2>
              <p>同一套漏斗假设下，对比预算强度和转化效率对收入的影响</p>
            </div>
            <div className="module-chip-row compact">
              {["月度", "季度", "半年"].map((view, index) => (
                <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
              ))}
            </div>
          </div>
          <div className="growth-scenario-grid">
            {scenarios.map((scenario) => (
              <article key={scenario.name}>
                <h3>{scenario.name}</h3>
                <strong>{scenario.revenue}</strong>
                <div>
                  <span>成本 {scenario.cost}</span>
                  <span>利润率 {scenario.margin}</span>
                </div>
                <p>{scenario.highlight}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="growth-lower-grid">
          <div className="growth-forecast-card">
            <div className="module-section-head">
              <div>
                <h2>收入预测</h2>
                <p>按月展示模型推演结果，方便评估节奏和资源缺口</p>
              </div>
            </div>
            <div className="growth-chart" aria-label="5个月收入预测">
              {monthlyForecast.map(([month, revenue, phase, height]) => (
                <article key={month}>
                  <div>
                    <span style={{ height: `${height}%` }} aria-hidden="true" />
                  </div>
                  <strong>{revenue}</strong>
                  <small>{month} · {phase}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="growth-action-card" aria-label="测算建议">
            <h2>AI 测算建议</h2>
            <strong>先提成交率，再扩大预算</strong>
            <p>当前模型里，成交转化率每提升 2 个点，比访问量增加 20% 更能改善利润。</p>
            <div>
              {actionItems.map((item) => <span key={item}>{item}</span>)}
            </div>
            <Link to="/tasks">生成增长任务</Link>
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default GrowthCalculatorPage;
