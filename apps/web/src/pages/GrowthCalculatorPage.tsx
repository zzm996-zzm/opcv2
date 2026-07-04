import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import {
  growthApi,
  type GrowthForecast,
  type GrowthModel,
  type GrowthRecommendations,
  type GrowthScenarios
} from "../lib/growthApi";

const emptyGrowthStats = [
  ["预计月收入", "¥0"],
  ["获客成本", "¥0"],
  ["回本周期", "0天"],
  ["净利润率", "0.0%"]
] as const;

const emptyAssumptions = [
  ["月访问量", "0", "暂无后端测算模型"],
  ["线索转化率", "0.0%", "暂无后端测算模型"],
  ["成交转化率", "0.0%", "暂无后端测算模型"],
  ["平均客单价", "¥0", "暂无后端测算模型"]
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

function funnelForModel(model: GrowthModel | null) {
  if (!model) return [];
  const visits = model.assumptions.monthly_visits;
  const leads = model.result.leads;
  const deals = model.result.deals;
  return [
    { label: "访问", value: formatNumber(visits), percent: 100, note: "后端模型月访问量" },
    { label: "线索", value: formatNumber(leads), percent: Math.max(Math.round(model.assumptions.lead_rate * 100), 1), note: "访问到留资或预约咨询" },
    { label: "成交", value: formatNumber(deals), percent: Math.max(Math.round(model.assumptions.lead_rate * model.assumptions.deal_rate * 100), 1), note: "线索到首单客户" }
  ];
}

function scenariosForView(view: GrowthScenarios | null) {
  if (!view) return [];
  return view.scenarios.map((scenario) => ({
    name: scenario.name,
    revenue: formatCurrency(scenario.revenue),
    cost: formatCurrency(scenario.cost),
    margin: formatPercent(scenario.margin),
    highlight: scenario.highlight
  }));
}

function forecastForView(view: GrowthForecast | null) {
  if (!view) return [];
  return view.months.map((month) => [
    month.month,
    formatCurrency(month.revenue),
    month.phase,
    month.progress_percent
  ] as const);
}

function costItemsForView(view: GrowthRecommendations | null) {
  if (!view) return [];
  return view.cost_items.map((item) => [item.name, formatCurrency(item.amount), item.detail] as const);
}

function GrowthCalculatorPage() {
  const [latestModel, setLatestModel] = useState<GrowthModel | null>(null);
  const [scenarioView, setScenarioView] = useState<GrowthScenarios | null>(null);
  const [forecastView, setForecastView] = useState<GrowthForecast | null>(null);
  const [recommendationView, setRecommendationView] = useState<GrowthRecommendations | null>(null);
  const [loadError, setLoadError] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let active = true;
    async function loadModels() {
      try {
        const payload = await growthApi.listModels();
        if (!active) return;
        const model = payload.models[0] ?? null;
        setLatestModel(model);
        setLoadError("");
        if (!model) {
          setScenarioView(null);
          setForecastView(null);
          setRecommendationView(null);
          return;
        }
        const [scenariosPayload, forecastPayload, recommendationsPayload] = await Promise.all([
          growthApi.modelScenarios(model.id).catch(() => null),
          growthApi.modelForecast(model.id).catch(() => null),
          growthApi.modelRecommendations(model.id).catch(() => null)
        ]);
        if (!active) return;
        setScenarioView(scenariosPayload);
        setForecastView(forecastPayload);
        setRecommendationView(recommendationsPayload);
      } catch (error) {
        if (!active) return;
        setLatestModel(null);
        setScenarioView(null);
        setForecastView(null);
        setRecommendationView(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取测算模型"));
      }
    }
    void loadModels();
    return () => {
      active = false;
    };
  }, []);

  const visibleStats = latestModel ? statsForModel(latestModel) : emptyGrowthStats;
  const visibleAssumptions = latestModel ? assumptionsForModel(latestModel) : emptyAssumptions;
  const visibleFunnel = funnelForModel(latestModel);
  const visibleScenarios = scenariosForView(scenarioView);
  const visibleForecast = forecastForView(forecastView);
  const visibleCostItems = costItemsForView(recommendationView);
  const visibleActionItems = recommendationView?.action_items ?? [];
  const recommendationHeadline = recommendationView?.headline ?? "暂无测算建议";
  const recommendationSummary = recommendationView?.summary ?? "保存或选择一个测算模型后，这里会展示后端生成的增长建议。";
  const modelName = latestModel?.name ?? "暂无测算模型";

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
      const [scenariosPayload, forecastPayload, recommendationsPayload] = await Promise.all([
        growthApi.modelScenarios(model.id).catch(() => null),
        growthApi.modelForecast(model.id).catch(() => null),
        growthApi.modelRecommendations(model.id).catch(() => null)
      ]);
      setScenarioView(scenariosPayload);
      setForecastView(forecastPayload);
      setRecommendationView(recommendationsPayload);
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
              {visibleFunnel.length === 0 ? (
                <div className="module-empty-state" role="status">暂无漏斗数据</div>
              ) : visibleFunnel.map((step) => (
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
            {visibleCostItems.length === 0 ? (
              <p className="module-empty-state">暂无成本结构</p>
            ) : visibleCostItems.map(([label, value, detail]) => (
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
            {visibleScenarios.length === 0 ? (
              <div className="module-empty-state" role="status">暂无情景对比</div>
            ) : visibleScenarios.map((scenario) => (
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
              {visibleForecast.length === 0 ? (
                <div className="module-empty-state" role="status">暂无收入预测</div>
              ) : visibleForecast.map(([month, revenue, phase, height]) => (
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
            <strong>{recommendationHeadline}</strong>
            <p>{recommendationSummary}</p>
            <div>
              {visibleActionItems.length === 0 ? <span>暂无行动建议</span> : visibleActionItems.map((item) => <span key={item}>{item}</span>)}
            </div>
            <Link to="/tasks">生成增长任务</Link>
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default GrowthCalculatorPage;
