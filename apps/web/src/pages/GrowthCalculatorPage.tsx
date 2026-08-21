import { useEffect, useState } from "react";
import { Download } from "lucide-react";
import { Link, useLocation, useSearchParams } from "react-router-dom";
import UnifiedCopilotPanel from "../components/UnifiedCopilotPanel";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { tasksApi, type Task, type TaskAIDraft } from "../lib/tasksApi";
import {
  growthApi,
  type GrowthActionPlan,
  type GrowthDraft,
  type GrowthForecast,
  type GrowthInputs,
  type GrowthModel,
  type GrowthRecommendations,
  type GrowthRisks,
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

function inputFieldValue(field: GrowthInputs["fields"][number]) {
  if (field.key.endsWith("rate")) return formatPercent(field.value);
  if (field.key === "monthly_visits") return formatNumber(field.value);
  return formatCurrency(field.value);
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

function modelIDFromQuery(value: string | null) {
  const id = Number(value);
  return Number.isInteger(id) && id > 0 ? id : null;
}

function riskLevelLabel(level: "low" | "medium" | "high") {
  return level === "high" ? "高风险" : level === "medium" ? "中风险" : "低风险";
}

function GrowthCalculatorPage() {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const requestedModelID = modelIDFromQuery(searchParams.get("model_id"));
  const [latestModel, setLatestModel] = useState<GrowthModel | null>(null);
  const [scenarioView, setScenarioView] = useState<GrowthScenarios | null>(null);
  const [forecastView, setForecastView] = useState<GrowthForecast | null>(null);
  const [recommendationView, setRecommendationView] = useState<GrowthRecommendations | null>(null);
  const [riskView, setRiskView] = useState<GrowthRisks | null>(null);
  const [actionPlanView, setActionPlanView] = useState<GrowthActionPlan | null>(null);
  const [inputsView, setInputsView] = useState<GrowthInputs | null>(null);
  const [loadError, setLoadError] = useState("");
  const [isLoading, setIsLoading] = useState(true);
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
  const [taskDraft, setTaskDraft] = useState<TaskAIDraft | null>(null);
  const [createdGrowthTasks, setCreatedGrowthTasks] = useState<Task[]>([]);
  const [forecastViewMode, setForecastViewMode] = useState("5个月");
  const [forecastMonths, setForecastMonths] = useState(5);
  const [isExporting, setIsExporting] = useState(false);
  const [exportMessage, setExportMessage] = useState("");
  const [exportError, setExportError] = useState("");

  useEffect(() => {
    let active = true;
    async function loadModels() {
      try {
        let model: GrowthModel | null = null;
        if (requestedModelID) {
          model = await growthApi.getModel(requestedModelID);
        } else {
          const payload = await growthApi.listModels();
          model = payload.models[0] ?? null;
        }
        if (!active) return;
        setLatestModel(model);
        setLoadError("");
        if (!model) {
          setScenarioView(null);
          setForecastView(null);
          setRecommendationView(null);
          setRiskView(null);
          setActionPlanView(null);
          setInputsView(null);
          return;
        }
        const [scenariosPayload, forecastPayload, recommendationsPayload, risksPayload, actionPlanPayload, inputsPayload] = await Promise.all([
          growthApi.modelScenarios(model.id).catch(() => null),
          growthApi.modelForecast(model.id, forecastMonths).catch(() => null),
          growthApi.modelRecommendations(model.id).catch(() => null),
          growthApi.modelRisks(model.id).catch(() => null),
          growthApi.modelActionPlan(model.id).catch(() => null),
          growthApi.modelInputs(model.id).catch(() => null)
        ]);
        if (!active) return;
        setScenarioView(scenariosPayload);
        setForecastView(forecastPayload);
        setRecommendationView(recommendationsPayload);
        setRiskView(risksPayload);
        setActionPlanView(actionPlanPayload);
        setInputsView(inputsPayload);
        const snapshotPayload = await growthApi.listSnapshots(model.id).catch(() => ({ snapshots: [] }));
        if (!active) return;
        setSnapshots(snapshotPayload.snapshots);
      } catch (error) {
        if (!active) return;
        setLatestModel(null);
        setScenarioView(null);
        setForecastView(null);
        setRecommendationView(null);
        setRiskView(null);
        setActionPlanView(null);
        setInputsView(null);
        setSnapshots([]);
        setLoadError(apiErrorMessage(error, "暂时无法读取测算模型"));
      } finally {
        if (active) setIsLoading(false);
      }
    }
    void loadModels();
    return () => {
      active = false;
    };
  }, [forecastMonths, requestedModelID]);

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
    const [scenariosPayload, forecastPayload, recommendationsPayload, risksPayload, actionPlanPayload, inputsPayload] = await Promise.all([
      growthApi.modelScenarios(model.id).catch(() => null),
      growthApi.modelForecast(model.id, forecastMonths).catch(() => null),
      growthApi.modelRecommendations(model.id).catch(() => null),
      growthApi.modelRisks(model.id).catch(() => null),
      growthApi.modelActionPlan(model.id).catch(() => null),
      growthApi.modelInputs(model.id).catch(() => null)
    ]);
    setScenarioView(scenariosPayload);
    setForecastView(forecastPayload);
    setRecommendationView(recommendationsPayload);
    setRiskView(risksPayload);
    setActionPlanView(actionPlanPayload);
    setInputsView(inputsPayload);
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
    setTaskDraft(null);
    setCreatedGrowthTasks([]);
    setTaskSyncMessage("");
    setTaskSyncError("");
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
    setRiskView(null);
    setActionPlanView(null);
    setInputsView(null);
    setTaskSyncMessage("");
    setTaskSyncError("");
  }

  async function syncToTaskCenter() {
    if (!latestModel || !recommendationView || syncingTasks) return;
    setSyncingTasks(true);
    setTaskSyncMessage("");
    setTaskSyncError("");
    try {
      const actionPlanContext = actionPlanView?.phases.flatMap((phase) => phase.items.map((item) =>
        `[${phase.name}][${item.id}] ${item.title}：${item.detail} 目标：${item.target_metric} 预期：${item.expected_result}`
      )).join("；") ?? "";
      const result = await tasksApi.generateTasks(
        `执行${recommendationView.model_name}的增长优化：${recommendationView.action_items.join("；")}。90 天阶段计划：${actionPlanContext}`,
        {
          sourceType: "growth_model",
          sourceId: recommendationView.model_id,
          sourceTitle: recommendationView.model_name,
          sourceUrl: "/growth-calculator"
        }
      );
      setTaskDraft(result.draft);
      setTaskSyncMessage(`已生成 ${result.tasks.length} 条任务草稿，请确认后创建`);
    } catch (error) {
      setTaskSyncError(apiErrorMessage(error, "同步任务失败，请稍后重试"));
    } finally {
      setSyncingTasks(false);
    }
  }

  async function adoptTaskDraft() {
    if (!taskDraft || syncingTasks) return;
    setSyncingTasks(true);
    setTaskSyncError("");
    try {
      const result = await tasksApi.adoptTaskAIDraft(
        taskDraft.id,
        taskDraft.tasks.map((task, draftIndex) => ({
          draftIndex,
          title: task.title,
          description: task.description,
          assignee: task.assignee,
          project: task.project,
          priority: task.priority,
          tags: Array.from(new Set([...(task.tags ?? []), "growth-model", "growth-plan"])),
          dueAt: task.due_at,
          tools: task.tools,
          learning: task.learning
        })),
        `growth-model-${latestModel?.id ?? 0}-task-draft-${taskDraft.id}`
      );
      setTaskDraft(null);
      setCreatedGrowthTasks(result.tasks ?? []);
      setTaskSyncMessage(`已创建 ${result.tasks.length} 个增长任务`);
    } catch (error) {
      setTaskSyncError(apiErrorMessage(error, "创建增长任务失败，请稍后重试"));
    } finally {
      setSyncingTasks(false);
    }
  }

  async function exportReport() {
    if (!latestModel || isExporting) return;
    setIsExporting(true);
    setExportMessage("");
    setExportError("");
    try {
      const blob = await growthApi.exportModel(latestModel.id);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = `growth-report-${latestModel.id}.json`;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
      setExportMessage("报告已导出");
    } catch (error) {
      setExportError(apiErrorMessage(error, "报告导出失败，请稍后重试"));
    } finally {
      setIsExporting(false);
    }
  }

  function closeTaskDraft() {
    setTaskDraft(null);
    setTaskSyncMessage("任务草稿已取消，尚未创建正式任务");
  }

  const growthCopilotFilters = {
    module: "growth",
    view: location.pathname.endsWith("/report") ? "report" : location.pathname.endsWith("/questions") ? "questions" : "home",
    ...(latestModel ? { model_id: String(latestModel.id) } : {})
  };

  return (
    <V4PageShell className="growth-calculator-shell">
      <section className="module-page growth-calculator-page" aria-label="增长测算">
        <div className="growth-calculator-content">
          {taskDraft ? (
          <div className="task-modal-backdrop">
            <section aria-label="增长任务草稿预览" aria-modal="true" className="task-ai-draft-dialog" role="dialog">
              <header>
                <span>任务草稿</span>
                <h2>确认后再进入任务中心</h2>
                <p>共 {taskDraft.tasks.length} 条建议，可取消或一次性确认创建。</p>
              </header>
              <div className="task-ai-draft-list">
                {taskDraft.tasks.map((task, index) => (
                  <article className="task-ai-draft-item selected" key={`${index}-${task.title}`}>
                    <div className="task-ai-draft-fields">
                      <strong>{task.title}</strong>
                      <span>{task.project} · {task.assignee || "待指定负责人"}</span>
                      <p>{task.description || "暂无补充说明"}</p>
                    </div>
                  </article>
                ))}
              </div>
              <footer className="task-ai-draft-footer">
                <button disabled={syncingTasks} onClick={closeTaskDraft} type="button">取消</button>
                <button className="primary" disabled={syncingTasks} onClick={() => void adoptTaskDraft()} type="button">
                  {syncingTasks ? "创建中..." : "确认创建任务"}
                </button>
              </footer>
            </section>
          </div>
        ) : null}
        <div className="page-title-row">
          <div>
            <h1>增长测算</h1>
            <p>用访问量、转化率、客单价、获客成本和交付成本，提前算清楚增长动作的收入和利润边界</p>
          </div>
          <div className="growth-page-actions">
            <Link to="/growth-calculator/history">测算历史</Link>
            <button className="growth-export-button" disabled={!latestModel || isExporting} onClick={() => void exportReport()} type="button">
              <Download aria-hidden="true" size={16} />
              {isExporting ? "导出中..." : "导出报告"}
            </button>
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
        {exportMessage ? <p className="form-success" role="status">{exportMessage}</p> : null}
        {exportError ? <p className="form-error" role="alert">{exportError}</p> : null}
        {isLoading ? <p className="module-loading" role="status">正在读取测算数据...</p> : null}
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
                maxLength={2000}
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

        <section className="growth-input-trace-section" aria-label="输入来源与可信度">
          <div className="module-section-head">
            <div>
              <h2>输入来源与可信度</h2>
              <p>来源和可信度独立于预测结果，避免把规则提取误读为预测准确率。</p>
            </div>
            {inputsView ? <strong className="growth-input-completeness">信息完整度 {inputsView.completeness_percent}%</strong> : null}
          </div>
          {inputsView?.fields?.length ? (
            <div className="growth-input-trace-grid">
              {inputsView.fields.map((field) => (
                <article key={field.key}>
                  <header><strong>{field.label}</strong><span>{inputFieldValue(field)} / {field.unit}</span></header>
                  <dl>
                    <div><dt>来源</dt><dd>{field.source}</dd></div>
                    <div><dt>可信度</dt><dd>{field.confidence}</dd></div>
                    <div><dt>用户确认</dt><dd>{field.confirmed_by_user ? "是" : "否"}</dd></div>
                  </dl>
                </article>
              ))}
            </div>
          ) : <p className="module-empty-state" role="status">暂无输入来源信息</p>}
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

        <section className="growth-risk-section" aria-label="关键风险诊断">
          <div className="module-section-head">
            <div>
              <h2>关键风险诊断</h2>
              <p>风险等级由当前模型规则计算，AI 只负责解释原因和建议。</p>
            </div>
            {riskView ? <strong className={`growth-risk-level ${riskView.overall_level}`}>整体风险：{riskLevelLabel(riskView.overall_level)}</strong> : null}
          </div>
          {riskView?.risks?.length ? (
            <div className="growth-risk-grid">
              {riskView.risks.map((risk) => (
                <article key={risk.key}>
                  <header>
                    <strong>{risk.name}</strong>
                    <span className={`growth-risk-level ${risk.level}`}>{riskLevelLabel(risk.level)}</span>
                  </header>
                  <p>{risk.reason}</p>
                  <dl>
                    <div><dt>当前值</dt><dd>{risk.current_value}</dd></div>
                    <div><dt>规则阈值</dt><dd>{risk.threshold}</dd></div>
                  </dl>
                  <small>{risk.suggestion}</small>
                </article>
              ))}
            </div>
          ) : <p className="module-empty-state" role="status">暂无风险诊断</p>}
        </section>

        <section className="growth-action-plan-section" aria-label="90 天行动计划">
          <div className="module-section-head">
            <div>
              <h2>90 天行动计划</h2>
              <p>按阶段拆解目标、责任角色和目标指标，执行前仍需在任务草稿中确认。</p>
            </div>
          </div>
          {actionPlanView?.phases?.length ? (
            <div className="growth-action-plan-grid">
              {actionPlanView.phases.map((phase) => (
                <article key={phase.key}>
                  <header>
                    <strong>{phase.name}</strong>
                    <span>{phase.goal}</span>
                  </header>
                  <div>
                    {phase.items.map((item) => (
                      <section key={item.id}>
                        <h3>{item.title}</h3>
                        <p>{item.detail}</p>
                        <dl>
                          <div><dt>责任角色</dt><dd>{item.owner_role}</dd></div>
                          <div><dt>目标指标</dt><dd>{item.target_metric}</dd></div>
                          <div><dt>预期结果</dt><dd>{item.expected_result}</dd></div>
                        </dl>
                      </section>
                    ))}
                  </div>
                </article>
              ))}
            </div>
          ) : <p className="module-empty-state" role="status">暂无行动计划</p>}
        </section>

        <section className="growth-lower-grid">
          <div className="growth-forecast-card">
            <div className="module-section-head">
              <div>
                <h2>收入预测</h2>
                <p>按月展示模型推演结果，方便评估节奏和资源缺口</p>
              </div>
              <div className="module-chip-row compact" aria-label="预测周期">
                {[{ label: "3个月", months: 3 }, { label: "5个月", months: 5 }, { label: "12个月", months: 12 }].map((view) => (
                  <button
                    aria-pressed={forecastViewMode === view.label}
                    className={forecastViewMode === view.label ? "active" : ""}
                    key={view.label}
                    onClick={() => {
                      setForecastViewMode(view.label);
                      setForecastMonths(view.months);
                    }}
                    type="button"
                  >
                    {view.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="growth-chart" aria-label={`${forecastMonths}个月收入预测`}>
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
              {syncingTasks ? "生成中..." : "生成任务草稿"}
            </button>
            {taskSyncMessage ? <p className="form-success" role="status">{taskSyncMessage}</p> : null}
            {taskSyncError ? <p className="form-error" role="alert">{taskSyncError}</p> : null}
            {createdGrowthTasks.length > 0 ? (
              <div className="growth-linked-tasks" aria-label="已关联增长任务">
                <strong>已关联任务</strong>
                {createdGrowthTasks.map((task) => (
                  <Link key={task.id} to="/tasks">#{task.id} {task.title}</Link>
                ))}
              </div>
            ) : null}
          </aside>
        </section>
        </div>
        <UnifiedCopilotPanel
          activeFilters={growthCopilotFilters}
          ariaLabel="增长测算 Copilot"
          className="growth-copilot-panel"
          currentView={`${location.pathname}${location.search}`}
          inputAriaLabel="向增长测算 Copilot 提问"
          response={latestModel ? "我会结合当前测算模型的收入、成本、转化和风险，帮你判断优先行动。" : "请先描述业务并完成测算，我会结合结果解释关键假设、风险和下一步动作。"}
          userPrompt={latestModel ? `分析当前测算：${latestModel.name}` : "帮我梳理当前增长测算需要补齐的信息"}
        />
      </section>
    </V4PageShell>
  );
}

export default GrowthCalculatorPage;
