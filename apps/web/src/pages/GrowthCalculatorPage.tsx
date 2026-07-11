import { useEffect, useState } from "react";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { tasksApi } from "../lib/tasksApi";
import {
  growthApi,
  type GrowthDraft,
  type GrowthForecast,
  type GrowthModel,
  type GrowthRecommendations,
  type GrowthScenarios,
  type GrowthSnapshot
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
  const [draft, setDraft] = useState<GrowthDraft | null>(null);
  const [businessInput, setBusinessInput] = useState("");
  const [draftAnswers, setDraftAnswers] = useState<Record<string, string>>({});
  const [draftError, setDraftError] = useState("");
  const [snapshots, setSnapshots] = useState<GrowthSnapshot[]>([]);
  const [selectedSnapshotID, setSelectedSnapshotID] = useState("");
  const [syncingTasks, setSyncingTasks] = useState(false);
  const [taskSyncMessage, setTaskSyncMessage] = useState("");
  const [taskSyncError, setTaskSyncError] = useState("");

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
        const snapshotPayload = await growthApi.listSnapshots(model.id).catch(() => ({ snapshots: [] }));
        if (!active) return;
        setSnapshots(snapshotPayload.snapshots);
      } catch (error) {
        if (!active) return;
        setLatestModel(null);
        setScenarioView(null);
        setForecastView(null);
        setRecommendationView(null);
        setSnapshots([]);
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

  async function loadModelViews(model: GrowthModel) {
    const [scenariosPayload, forecastPayload, recommendationsPayload] = await Promise.all([
      growthApi.modelScenarios(model.id).catch(() => null),
      growthApi.modelForecast(model.id).catch(() => null),
      growthApi.modelRecommendations(model.id).catch(() => null)
    ]);
    setScenarioView(scenariosPayload);
    setForecastView(forecastPayload);
    setRecommendationView(recommendationsPayload);
  }

  async function startDraft() {
    if (isSaving || !businessInput.trim()) return;
    setIsSaving(true);
    setDraftError("");
    try {
      const nextDraft = await growthApi.createDraft(businessInput.trim());
      setDraft(nextDraft);
      setDraftAnswers({});
    } catch (error) {
      setDraftError(apiErrorMessage(error, "暂时无法分析业务信息"));
    } finally {
      setIsSaving(false);
    }
  }

  async function submitAnswers() {
    if (!draft || isSaving) return;
    const answers = Object.fromEntries(draft.questions.map((question) => [question.key, Number(draftAnswers[question.key])]));
    if (Object.values(answers).some((value) => !Number.isFinite(value))) {
      setDraftError("请补齐所有测算信息");
      return;
    }
    setIsSaving(true);
    setDraftError("");
    try {
      const nextDraft = await growthApi.answerDraft(draft.id, answers);
      setDraft(nextDraft);
      setDraftAnswers({});
    } catch (error) {
      setDraftError(apiErrorMessage(error, "暂时无法保存补充信息"));
    } finally {
      setIsSaving(false);
    }
  }

  async function calculateDraft() {
    if (!draft || draft.status !== "ready" || isSaving) return;
    if (isSaving) return;
    setIsSaving(true);
    setDraftError("");
    try {
      const calculation = await growthApi.calculateDraft(draft.id);
      setDraft(calculation.draft);
      setLatestModel(calculation.model);
      await loadModelViews(calculation.model);
      if (calculation.snapshot) {
        setSnapshots((current) => [calculation.snapshot, ...current.filter((item) => item.id !== calculation.snapshot.id)]);
      }
    } catch (error) {
      setDraftError(apiErrorMessage(error, "暂时无法生成测算结果"));
    } finally {
      setIsSaving(false);
    }
  }

  function resetDraft() {
    setDraft(null);
    setBusinessInput("");
    setDraftAnswers({});
    setDraftError("");
  }

  function restoreSnapshot(snapshotID: string) {
    setSelectedSnapshotID(snapshotID);
    const snapshot = snapshots.find((item) => item.id === Number(snapshotID));
    if (!snapshot) return;
    setLatestModel({
      id: snapshot.model_id,
      user_id: snapshot.user_id,
      name: snapshot.model_name,
      assumptions: snapshot.assumptions,
      result: snapshot.result,
      created_at: snapshot.created_at,
      updated_at: snapshot.created_at
    });
    setScenarioView(snapshot.scenarios);
    setForecastView(snapshot.forecast);
    setRecommendationView(snapshot.recommendations);
    setTaskSyncMessage("");
    setTaskSyncError("");
  }

  async function syncToTaskCenter() {
    if (!latestModel || !recommendationView || syncingTasks) return;
    setSyncingTasks(true);
    setTaskSyncMessage("");
    setTaskSyncError("");
    try {
      const result = await tasksApi.generateTasks(
        `执行${recommendationView.model_name}的增长优化：${recommendationView.action_items.join("；")}`,
        {
          sourceType: "growth_model",
          sourceId: recommendationView.model_id,
          sourceTitle: recommendationView.model_name,
          sourceUrl: "/growth-calculator"
        }
      );
      setTaskSyncMessage(`已创建 ${result.tasks.length} 个增长任务`);
    } catch {
      setTaskSyncError("同步任务失败，请稍后重试");
    } finally {
      setSyncingTasks(false);
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
          <div className="growth-page-actions">
            {snapshots.length > 0 ? (
              <label>
                <span>历史测算</span>
                <select aria-label="历史测算" onChange={(event) => restoreSnapshot(event.target.value)} value={selectedSnapshotID}>
                  <option value="">当前结果</option>
                  {snapshots.map((snapshot) => (
                    <option key={snapshot.id} value={snapshot.id}>{new Date(snapshot.created_at).toLocaleString("zh-CN")}</option>
                  ))}
                </select>
              </label>
            ) : null}
            <button className="module-primary-action" disabled={isSaving} onClick={resetDraft} type="button">
              新建测算
            </button>
          </div>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}

        <section className="module-overview-card growth-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">{modelName}</span>
            {latestModel ? <span className="growth-model-label">模型测算</span> : null}
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

          <form className="growth-input-card" aria-label="增长测算输入" onSubmit={(event) => event.preventDefault()}>
            <strong>业务测算输入</strong>
            <label className="growth-intake-field">
              <span>业务与增长问题</span>
              <textarea
                aria-label="业务与增长问题"
                disabled={Boolean(draft)}
                onChange={(event) => setBusinessInput(event.target.value)}
                placeholder="描述业务、获客方式、成本、客单价，以及你最想确认的增长问题"
                value={businessInput}
              />
            </label>
            {!draft ? (
              <button className="growth-draft-button" disabled={isSaving || !businessInput.trim()} onClick={() => void startDraft()} type="button">
                {isSaving ? "分析中..." : "开始测算"}
              </button>
            ) : null}
            {draft?.status === "needs_input" ? (
              <div className="growth-question-list">
                <p className="growth-draft-status">还需要 {draft.questions.length} 项信息</p>
                {draft.questions.map((question) => (
                  <label className="growth-question-row" key={question.key}>
                    <span>{question.label}</span>
                    <div>
                      <input
                        aria-label={question.label}
                        min={question.min}
                        max={question.max}
                        onChange={(event) => setDraftAnswers((current) => ({ ...current, [question.key]: event.target.value }))}
                        type="number"
                        value={draftAnswers[question.key] ?? ""}
                      />
                      <small>{question.unit}</small>
                    </div>
                  </label>
                ))}
                <button className="growth-draft-button" disabled={isSaving} onClick={() => void submitAnswers()} type="button">
                  {isSaving ? "提交中..." : "提交补充信息"}
                </button>
              </div>
            ) : null}
            {draft?.status === "ready" ? (
              <div className="growth-ready-state">
                <span>测算参数已补齐</span>
                <button className="growth-draft-button" disabled={isSaving} onClick={() => void calculateDraft()} type="button">
                  {isSaving ? "生成中..." : "生成测算结果"}
                </button>
              </div>
            ) : null}
            {draftError ? <p className="form-error" role="alert">{draftError}</p> : null}
            {latestModel ? (
              <div className="growth-assumption-list">
                <strong>核心假设 · 模型测算</strong>
                {visibleAssumptions.map(([label, value, helper]) => (
                  <label className="growth-assumption-row" key={label}>
                    <span>{label}</span>
                    <input aria-label={label} readOnly value={value} />
                    <small>{helper}</small>
                  </label>
                ))}
              </div>
            ) : null}
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
            <button disabled={!latestModel || !recommendationView || syncingTasks} onClick={() => void syncToTaskCenter()} type="button">
              {syncingTasks ? "生成中..." : "生成增长任务"}
            </button>
            {taskSyncMessage ? <p className="form-success" role="status">{taskSyncMessage}</p> : null}
            {taskSyncError ? <p className="form-error" role="alert">{taskSyncError}</p> : null}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default GrowthCalculatorPage;
