import { ArrowLeft, BarChart3, CircleAlert, Clock3, Download, FileCheck2, Flag, ShieldAlert, Sparkles, Target, TrendingUp, Users } from "lucide-react";
import { Link } from "react-router-dom";
import type { ReactNode } from "react";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import type { SandboxReport, SandboxSession } from "../../lib/sandboxApi";

type SandboxReportViewProps = {
  loading?: boolean;
  session: SandboxSession | null;
};

function SandboxReportView({ loading = false, session }: SandboxReportViewProps) {
  const report = session?.report;

  if (!session || session.status !== "completed" || !report?.score) {
    return (
      <SandboxFrame copilotMode="report">
        <section className="sb-empty-panel sb-report-empty">
          <FileCheck2 size={32} />
          <h1>{loading ? "正在加载推演报告" : "暂无可查看的推演报告"}</h1>
          <p>{loading ? "正在从服务端读取当前沙盘会话。" : "完成一轮推演后，结构化报告会显示在这里。"}</p>
          <Link to="/sandbox">返回商业沙盘首页<ArrowLeft size={16} /></Link>
        </section>
      </SandboxFrame>
    );
  }

  const currentSession = session;

  function exportReport() {
    const blob = new Blob([JSON.stringify({ session: currentSession, report }, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${currentSession.product || "商业沙盘推演报告"}.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  return (
    <SandboxFrame copilotMode="report" copilotProject={session.intake?.initial_idea || session.goal}>
      <header className="sb-report-heading">
        <div><div className="sb-breadcrumb"><Link to="/sandbox/history">商业沙盘</Link><span>/</span><strong>推演报告</strong></div><h1>{session.product || session.goal}</h1><p>推演时间：{formatDate(session.updated_at)} <span>·</span> 参与角色数：{session.roles.length} <span>·</span> 报告版本：{report.report_version || "V1.0"}</p></div>
        <button className="sb-export-button" onClick={exportReport} type="button"><Download size={16} />导出报告</button>
      </header>
      <section className="sb-report-metrics">
        <MetricCard icon={<Target size={21} />} label="综合可行性" value={`${report.score}`} suffix="/100" tone="purple" detail="建议先进入小范围验证" />
        <MetricCard icon={<CircleAlert size={21} />} label="消费概率" value={`${report.consumer_probability ?? "-"}`} suffix="%" tone="blue" detail="基于当前输入条件推演" />
        <MetricCard icon={<ShieldAlert size={21} />} label="风险等级" value={riskLabel(report.risk_level)} suffix="" tone="orange" detail={`${report.risks.length} 项风险需要关注`} />
        <MetricCard icon={<Sparkles size={21} />} label="推荐优先级" value={report.recommendation_grade || "-"} suffix="" tone="green" detail="建议优先投入验证资源" />
      </section>
      <section className="sb-report-conclusion">
        <header><Sparkles size={18} /><h2>核心结论</h2></header>
        <p>{report.summary}</p>
        <ul>{(report.core_conclusions?.length ? report.core_conclusions : report.next_actions).slice(0, 4).map((item) => <li key={item}>{item}</li>)}</ul>
      </section>
      <div className="sb-report-two-col">
        <InsightSection icon={<TrendingUp size={18} />} title="机会分析" tone="opportunity" items={report.opportunity_analysis} fallback={report.next_actions} />
        <InsightSection icon={<ShieldAlert size={18} />} title="风险分析" tone="risk" items={report.risk_analysis} fallback={report.risks} />
      </div>
      <section className="sb-report-section">
        <header><Users size={18} /><h2>角色观点摘要</h2></header>
        <div className="sb-report-role-grid">{report.role_summaries.map((item) => <article key={item.role}><strong>{item.role}</strong><p>{item.view}</p></article>)}</div>
      </section>
      <section className="sb-report-section">
        <header><Flag size={18} /><h2>建议的下一步动作</h2></header>
        <div className="sb-action-plan">{(report.action_plan?.length ? report.action_plan : report.next_actions.map((title, index) => ({ order: index + 1, title, detail: "根据当前推演结果安排验证。", duration: "待定" }))).map((item) => <article key={`${item.order}-${item.title}`}><b>{item.order}</b><div><strong>{item.title}</strong><p>{item.detail}</p></div><small>{item.duration}</small></article>)}</div>
      </section>
      <div className="sb-report-bottom-grid">
        <section className="sb-report-section"><header><BarChart3 size={18} /><h2>关键验证指标</h2></header><div className="sb-validation-list">{(report.validation_metrics ?? report.metrics.map((metric) => ({ label: metric.label, current: "待确认", target: metric.value, confidence_percent: 50 }))).map((metric) => <article key={metric.label}><div><strong>{metric.label}</strong><span>{metric.confidence_percent}%</span></div><p><span>{metric.current}</span><b>目标：{metric.target}</b></p><i><span style={{ width: `${metric.confidence_percent}%` }} /></i></article>)}</div></section>
        <section className="sb-report-section"><header><Clock3 size={18} /><h2>建议时间线</h2></header><ol className="sb-report-timeline">{(report.timeline ?? []).map((item) => <li key={item.title}><span /><div><strong>{item.title}</strong><small>{item.period}</small></div></li>)}</ol>{!report.timeline?.length && <p className="sb-muted-copy">报告未返回时间线，请先完成关键假设验证。</p>}</section>
      </div>
      <footer className="sb-report-footer"><Link to={`/sandbox/run?session=${session.id}`}><ArrowLeft size={16} />返回本轮推演</Link><Link className="is-primary" to="/sandbox/history">查看推演记录</Link></footer>
      {report.disclaimer ? <p className="sb-report-disclaimer">{report.disclaimer}</p> : null}
    </SandboxFrame>
  );
}

function MetricCard({ detail, icon, label, suffix, tone, value }: { detail: string; icon: ReactNode; label: string; suffix: string; tone: string; value: string }) {
  return <article className={`sb-metric-card is-${tone}`}><span>{icon}</span><small>{label}</small><strong>{value}<em>{suffix}</em></strong><p>{detail}</p><i /></article>;
}

function InsightSection({ fallback, icon, items, title, tone }: { fallback: string[]; icon: ReactNode; items?: SandboxReport["opportunity_analysis"]; title: string; tone: string }) {
  return <section className={`sb-report-section sb-insight-section is-${tone}`}><header>{icon}<h2>{title}</h2></header><div>{items?.length ? items.map((item) => <article key={item.title}><strong>{item.title}</strong><p>{item.detail}</p><small>{item.tags.join(" · ")}</small></article>) : fallback.slice(0, 3).map((item) => <article key={item}><strong>{item}</strong><p>请在真实验证中补充证据和边界条件。</p></article>)}</div></section>;
}

function formatDate(value: string) {
  if (!value) return "--";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString("zh-CN", { hour12: false });
}

function riskLabel(value?: string) {
  if (value === "high") return "高风险";
  if (value === "low") return "低风险";
  if (value === "medium") return "中等";
  return value || "待评估";
}

export default SandboxReportView;
