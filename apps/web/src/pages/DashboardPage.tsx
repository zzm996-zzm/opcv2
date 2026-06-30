import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { dashboardApi, type DashboardSummary } from "../lib/dashboardApi";

const metrics = [
  ["本月收入", "¥18.6万", "+22%"],
  ["新增线索", "328", "+41%"],
  ["成交客户", "46", "+12%"],
  ["待办任务", "17", "-8%"]
] as const;

const projects = [
  ["智能客服系统", "¥86万", "线索 128", "进行中"],
  ["AI短视频代运营", "¥42万", "线索 64", "验证中"],
  ["企业内训工具包", "¥28万", "线索 39", "待推进"]
] as const;

const trend = [
  ["周一", 34],
  ["周二", 46],
  ["周三", 58],
  ["周四", 52],
  ["周五", 73],
  ["周六", 64],
  ["周日", 88]
] as const;

const pipeline = [
  ["线索", "328", "100%"],
  ["已触达", "184", "56%"],
  ["已演示", "72", "22%"],
  ["成交", "46", "14%"]
] as const;

const alerts = [
  ["竞品价格页变化", "小鹅通新增 AI 助教套餐，建议更新对比话术"],
  ["线索跟进延迟", "12 条高意向线索超过 24 小时未触达"],
  ["内容任务阻塞", "GEO 选型页缺少价格和实施周期模块"]
] as const;

const tasks = [
  ["今天 14:00", "跟进星桥教育集团演示邀约"],
  ["今天 18:00", "完成智能客服系统报价模板"],
  ["明天 10:30", "复盘本周 GEO 关键词覆盖"]
] as const;

function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let active = true;
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
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleMetrics = summary?.metrics.length ? summary.metrics.map((item) => [item.label, item.value, item.change] as const) : metrics;
  const visibleProjects = summary?.projects.length ? summary.projects.map((item) => [item.name, item.value, item.leads, item.stage] as const) : projects;
  const visibleTrend = summary?.trend.length ? summary.trend.map((item) => [item.label, item.value] as const) : trend;
  const visiblePipeline = summary?.pipeline.length ? summary.pipeline.map((item) => [item.stage, item.count, item.percent] as const) : pipeline;
  const visibleAlerts = summary?.alerts.length ? summary.alerts.map((item) => [item.title, item.detail] as const) : alerts;
  const visibleTasks = summary?.actions.length ? summary.actions.map((item) => [item.time, item.title] as const) : tasks;

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
            {visibleMetrics.map(([label, value, change]) => (
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
              {visibleTrend.map(([day, height]) => (
                <article key={day}>
                  <span style={{ height: `${height}%` }} aria-hidden="true" />
                  <small>{day}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="dashboard-alert-card" aria-label="经营预警">
            <h2>经营预警</h2>
            {visibleAlerts.map(([title, detail]) => (
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
              {visibleProjects.map(([name, value, leads, stage]) => (
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
            {visiblePipeline.map(([stage, count, percent]) => (
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
            {visibleTasks.map(([time, action]) => (
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

export default DashboardPage;
