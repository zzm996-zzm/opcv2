import { useEffect, useState } from "react";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { geoApi, type GeoAnalysisRequest, type GeoOverview } from "../lib/geoApi";

const roadmap = [
  ["1", "锁定问题", "从客户搜索、竞品内容和 AI 回答里提取高意向问题"],
  ["2", "生成内容", "围绕选型、对比、案例、问答建立可引用内容阵地"],
  ["3", "优化引用", "补充结构化事实、案例、价格和可信来源"],
  ["4", "承接线索", "把高意向访问导入诊断表单、CRM 和跟进任务"]
] as const;

function GeoAcquisitionPage() {
  const [overview, setOverview] = useState<GeoOverview | null>(null);
  const [analysisRequests, setAnalysisRequests] = useState<GeoAnalysisRequest[]>([]);
  const [loadError, setLoadError] = useState("");
  const [requestLoadError, setRequestLoadError] = useState("");
  const [target, setTarget] = useState("");
  const [submitStatus, setSubmitStatus] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    geoApi
      .overview()
      .then((payload) => {
        if (!active) return;
        setOverview(payload);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setOverview(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取 GEO 数据"));
      });
    geoApi
      .listAnalysisRequests(5)
      .then((payload) => {
        if (!active) return;
        setAnalysisRequests(payload.requests ?? []);
        setRequestLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setAnalysisRequests([]);
        setRequestLoadError(apiErrorMessage(error, "暂时无法读取 GEO 分析请求"));
      });
    return () => {
      active = false;
    };
  }, []);

  const stats = overview?.stats ?? [];
  const engines = overview?.engines ?? [];
  const leadSignals = overview?.lead_signals ?? [];
  const keywords = overview?.keywords ?? [];
  const contentTasks = overview?.content_tasks ?? [];
  const hasOverviewData = stats.length > 0 || engines.length > 0 || leadSignals.length > 0 || keywords.length > 0 || contentTasks.length > 0;

  const submitAnalysisRequest = async () => {
    const normalizedTarget = target.trim();
    if (!normalizedTarget) {
      setSubmitStatus("");
      setSubmitError("请输入 GEO 获客目标");
      return;
    }
    setSubmitting(true);
    setSubmitStatus("");
    setSubmitError("");
    try {
      const request = await geoApi.createAnalysisRequest({ target: normalizedTarget });
      setAnalysisRequests((current) => [request, ...current.filter((item) => item.id !== request.id)].slice(0, 5));
      setSubmitStatus("GEO 分析请求已提交");
      setTarget("");
    } catch (error) {
      setSubmitError(apiErrorMessage(error, "暂时无法提交 GEO 分析请求"));
    } finally {
      setSubmitting(false);
    }
  };

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
            <span className="module-kicker">GEO 模块 · 接口待接入</span>
            <h2>把高意向问题变成持续获客入口</h2>
            <p>系统会拆解客户在 AI 搜索里会问什么、哪些答案已经引用你、哪里还缺可信内容，并生成可执行的内容和线索承接任务。</p>
            {loadError && <p className="form-error" role="alert">{loadError}</p>}
            {!loadError && !hasOverviewData && <p className="form-error" role="status">GEO 后端接口未接入</p>}
            <div className="module-stat-strip">
              {stats.length === 0 ? (
                <>
                  <article>
                    <small>AI引用覆盖</small>
                    <strong>0%</strong>
                  </article>
                  <article>
                    <small>目标关键词</small>
                    <strong>0</strong>
                  </article>
                  <article>
                    <small>待补内容</small>
                    <strong>0</strong>
                  </article>
                  <article>
                    <small>潜在线索数</small>
                    <strong>0</strong>
                  </article>
                </>
              ) : stats.map((item) => (
                <article key={item.key}>
                  <small>{item.label}</small>
                  <strong>{item.value}</strong>
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
              value={target}
              onChange={(event) => setTarget(event.target.value)}
            />
            {submitStatus && <p className="form-success" role="status">{submitStatus}</p>}
            {submitError && <p className="form-error" role="alert">{submitError}</p>}
            <button type="button" onClick={submitAnalysisRequest} disabled={submitting}>
              {submitting ? "提交中..." : "分析 AI 搜索机会"}
            </button>
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
              {engines.length === 0 && <p>暂无 AI 搜索覆盖数据</p>}
              {engines.map((engine) => (
                <article key={engine.name}>
                  <strong>{engine.name}</strong>
                  <span>{engine.coverage_percent}%</span>
                  <i style={{ width: `${engine.coverage_percent}%` }} aria-hidden="true" />
                  <small>{engine.status || "暂无状态"}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="geo-lead-card" aria-label="线索机会">
            <h2>线索机会</h2>
            {leadSignals.length === 0 && <p>暂无线索机会</p>}
            {leadSignals.map((signal) => (
              <article key={signal.title}>
                <strong>{signal.title}</strong>
                <small>{signal.detail}</small>
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
            {keywords.length === 0 && <p>暂无关键词机会</p>}
            {keywords.map((item) => (
              <article key={item.id}>
                <div>
                  <h3>{item.query}</h3>
                  <small>{item.intent || "未标注意图"} · {item.coverage || "暂无覆盖数据"}</small>
                </div>
                <strong>{item.score}</strong>
                <span>{item.action || "暂无动作"}</span>
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
              {contentTasks.length === 0 && <p>暂无内容阵地任务</p>}
              {contentTasks.map((task) => (
                <article key={task.id}>
                  <span>{task.type || "内容任务"}</span>
                  <strong>{task.title}</strong>
                  <em className={task.priority === "高" ? "hot" : ""}>{task.priority || "未标注"}</em>
                  <small>{task.due_at || "未排期"}</small>
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

          <aside className="geo-roadmap-card" aria-label="GEO分析请求">
            <h2>最近分析请求</h2>
            {requestLoadError && <p className="form-error" role="alert">{requestLoadError}</p>}
            {!requestLoadError && analysisRequests.length === 0 && <p>暂无 GEO 分析请求</p>}
            {analysisRequests.map((request) => (
              <article key={request.id}>
                <b>{request.status}</b>
                <span>
                  <strong>{request.target}</strong>
                  <small>{request.created_at ? new Date(request.created_at).toLocaleString("zh-CN") : "暂无提交时间"}</small>
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
