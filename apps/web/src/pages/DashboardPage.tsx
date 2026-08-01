import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import FeatureLockedPanel from "../components/FeatureLockedPanel";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { dashboardApi, type DashboardSummary } from "../lib/dashboardApi";
import { membershipApi, type FeatureAccess } from "../lib/membershipApi";

function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loadError, setLoadError] = useState("");
  const [featureAccess, setFeatureAccess] = useState<FeatureAccess | null>(null);
  const [featureAccessError, setFeatureAccessError] = useState("");
  const [summaryLoading, setSummaryLoading] = useState(false);

  useEffect(() => {
    let active = true;
    membershipApi
      .featureAccess(["dashboard"])
      .then((payload) => {
        if (!active) return;
        setFeatureAccess(payload.features[0] ?? null);
        setFeatureAccessError("");
      })
      .catch((error) => {
        if (!active) return;
        setFeatureAccess(null);
        setFeatureAccessError(apiErrorMessage(error, "暂时无法读取功能开通状态"));
      });
    return () => {
      active = false;
    };
  }, []);

  const canUseWorkflow = featureAccess?.allow_workflow === true;

  useEffect(() => {
    if (!canUseWorkflow) return;
    let active = true;
    setSummaryLoading(true);
    dashboardApi
      .getSummary()
      .then((payload) => {
        if (!active) return;
        setSummary(payload);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setSummary(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取仪表盘数据"));
      })
      .finally(() => {
        if (active) setSummaryLoading(false);
      });
    return () => {
      active = false;
    };
  }, [canUseWorkflow]);

  const visibleMetrics = summary?.metrics.map((item) => [item.label, item.value, item.change] as const) ?? [];
  const visibleProjects = summary?.projects.map((item) => [item.name, item.value, item.leads, item.stage] as const) ?? [];
  const visibleTrend = summary?.trend.map((item) => [item.label, item.value] as const) ?? [];
  const visiblePipeline = summary?.pipeline.map((item) => [item.stage, item.count, item.percent] as const) ?? [];
  const visibleAlerts = summary?.alerts.map((item) => [item.title, item.detail] as const) ?? [];
  const visibleTasks = summary?.actions.map((item) => [item.time, item.title] as const) ?? [];

  if (!canUseWorkflow) {
    if (featureAccess) return <FeatureLockedPanel feature={featureAccess} variant="dashboard" />;
    return (
      <V4PageShell className="dashboard-shell" showCopilotMini={false}>
        <DashboardSkeleton error={featureAccessError} />
      </V4PageShell>
    );
  }

  return (
    <V4PageShell className="dashboard-shell">
      <section className="module-page dashboard-page" aria-label="仪表盘">
        <div className="page-title-row">
          <div>
            <h1>仪表盘</h1>
            <p>汇总项目收入、线索、任务、CRM 和预警，帮你判断今天该优先推进什么</p>
          </div>
          <button className="module-primary-action" type="button">生成经营周报</button>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}

        <section className="dashboard-metric-section">
          <div className="module-section-head">
            <div>
              <h2>经营指标</h2>
              <p>按本月口径汇总收入、线索和执行效率</p>
            </div>
          </div>
          <div className="dashboard-metric-grid">
            {summaryLoading ? <DashboardMetricSkeleton /> : visibleMetrics.length === 0 ? (
              <div className="module-empty-state" role="status">暂无经营指标</div>
            ) : visibleMetrics.map(([label, value, change]) => (
              <article key={label}>
                <small>{label}</small>
                <strong>{value}</strong>
                <span className={change.startsWith("-") ? "down" : ""}>{change}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="dashboard-grid">
          <div className="dashboard-trend-card">
            <div className="module-section-head">
              <div>
                <h2>增长趋势</h2>
                <p>最近 7 天线索与成交机会变化</p>
              </div>
              <div className="module-chip-row compact">
                {["7天", "30天", "季度"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>
            <div className="dashboard-trend-chart" aria-label="增长趋势图">
              {summaryLoading ? <DashboardChartSkeleton /> : visibleTrend.length === 0 ? (
                <div className="module-empty-state" role="status">暂无增长趋势</div>
              ) : visibleTrend.map(([day, height]) => (
                <article key={day}>
                  <span style={{ height: `${height}%` }} aria-hidden="true" />
                  <small>{day}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="dashboard-alert-card" aria-label="经营预警">
            <h2>经营预警</h2>
            {summaryLoading ? <DashboardListSkeleton /> : visibleAlerts.length === 0 ? (
              <p className="module-empty-state">暂无经营预警</p>
            ) : visibleAlerts.map(([title, detail]) => (
              <article key={title}>
                <strong>{title}</strong>
                <small>{detail}</small>
              </article>
            ))}
          </aside>
        </section>

        <section className="dashboard-lower-grid">
          <div className="dashboard-project-card">
            <div className="module-section-head">
              <div>
                <h2>项目机会</h2>
                <p>把项目收入、线索和当前阶段放在同一张表里</p>
              </div>
            </div>
            <div className="dashboard-project-list">
              {summaryLoading ? <DashboardListSkeleton /> : visibleProjects.length === 0 ? (
                <div className="module-empty-state" role="status">暂无项目机会</div>
              ) : visibleProjects.map(([name, value, leads, stage]) => (
                <article key={name}>
                  <strong>{name}</strong>
                  <span>{value}</span>
                  <small>{leads}</small>
                  <em>{stage}</em>
                </article>
              ))}
            </div>
          </div>

          <aside className="dashboard-pipeline-card" aria-label="销售漏斗">
            <h2>销售漏斗</h2>
            {summaryLoading ? <DashboardListSkeleton /> : visiblePipeline.length === 0 ? (
              <p className="module-empty-state">暂无销售漏斗</p>
            ) : visiblePipeline.map(([stage, count, percent]) => (
              <article key={stage}>
                <span>
                  <strong>{stage}</strong>
                  <small>{count} 个</small>
                </span>
                <em>{percent}</em>
              </article>
            ))}
            <Link to="/crm">查看CRM</Link>
          </aside>
        </section>

        <section className="dashboard-task-section">
          <div className="module-section-head">
            <div>
              <h2>今日行动</h2>
              <p>从任务中心、线索开发和竞品监测同步过来的关键动作</p>
            </div>
            <Link to="/tasks">查看全部任务</Link>
          </div>
          <div className="dashboard-task-list">
            {summaryLoading ? <DashboardListSkeleton /> : visibleTasks.length === 0 ? (
              <div className="module-empty-state" role="status">暂无今日行动</div>
            ) : visibleTasks.map(([time, action]) => (
              <article key={`${time}-${action}`}>
                <time>{time}</time>
                <strong>{action}</strong>
              </article>
            ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

function DashboardSkeleton({ error }: { error: string }) {
  return <section aria-busy="true" aria-label="仪表盘加载中" className="module-page dashboard-page dashboard-skeleton">
    <div className="page-title-row"><div><h1>仪表盘</h1><p>{error || "正在读取仪表盘能力与经营数据..."}</p></div><span className="dashboard-skeleton-button" /></div>
    {error ? <p className="form-error" role="alert">{error}</p> : null}
    <section className="dashboard-metric-section"><div className="module-section-head"><div><h2>经营指标</h2><p>按本月口径汇总收入、线索和执行效率</p></div></div><div className="dashboard-metric-grid"><DashboardMetricSkeleton /></div></section>
    <section className="dashboard-grid"><div className="dashboard-trend-card"><div className="module-section-head"><div><h2>增长趋势</h2><p>最近 7 天线索与成交机会变化</p></div></div><DashboardChartSkeleton /></div><aside className="dashboard-alert-card"><h2>经营预警</h2><DashboardListSkeleton /></aside></section>
  </section>;
}

function DashboardMetricSkeleton() {
  return <>{["收入", "线索", "转化", "利润"].map((label) => <article className="skeleton-block" key={label}><small /><strong /><span /></article>)}</>;
}

function DashboardChartSkeleton() {
  return <div aria-hidden="true" className="dashboard-chart-skeleton">{[42, 64, 50, 78, 58, 88, 70].map((height, index) => <i key={index} style={{ height: `${height}%` }} />)}</div>;
}

function DashboardListSkeleton() {
  return <div aria-hidden="true" className="dashboard-list-skeleton"><i /><i /><i /></div>;
}

export default DashboardPage;
