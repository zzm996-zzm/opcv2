import { ArrowLeft, ArrowRight, CalendarDays, Crown, FileText, History, RotateCcw, Search, Sparkles } from "lucide-react";
import { useMemo, useState } from "react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import type { SandboxSession } from "../../lib/sandboxApi";

type SandboxHistoryViewProps = {
  loading?: boolean;
  sessions: SandboxSession[];
};

const pageSize = 7;

function SandboxHistoryView({ loading = false, sessions }: SandboxHistoryViewProps) {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [role, setRole] = useState("all");
  const [dateRange, setDateRange] = useState("all");
  const [page, setPage] = useState(1);
  const roleOptions = useMemo(() => Array.from(new Set(sessions.flatMap((session) => session.roles))).sort(), [sessions]);
  const filtered = useMemo(() => sessions.filter((session) => {
    const text = `${session.product} ${session.goal} ${session.target_users}`.toLowerCase();
    if (query.trim() && !text.includes(query.trim().toLowerCase())) return false;
    if (status !== "all" && session.status !== status) return false;
    if (role !== "all" && !session.roles.includes(role)) return false;
    if (dateRange !== "all") {
      const age = Date.now() - new Date(session.created_at).getTime();
      const days = age / 86_400_000;
      if (dateRange === "7" && days > 7) return false;
      if (dateRange === "30" && days > 30) return false;
      if (dateRange === "90" && days > 90) return false;
    }
    return true;
  }), [dateRange, query, role, sessions, status]);
  const pageCount = Math.max(1, Math.ceil(filtered.length / pageSize));
  const safePage = Math.min(page, pageCount);
  const rows = filtered.slice((safePage - 1) * pageSize, safePage * pageSize);

  function reset() {
    setQuery("");
    setStatus("all");
    setRole("all");
    setDateRange("all");
    setPage(1);
  }

  return (
    <SandboxFrame copilotMode="history">
      <header className="sb-history-heading"><div className="sb-breadcrumb"><Link to="/sandbox">商业沙盘</Link><span>/</span><strong>历史推演</strong><small>查看与管理你过往的商业沙盘推演记录</small></div></header>
      <section className="sb-history-notice">
        <span><History size={22} /></span>
        <div><strong>历史记录管理说明</strong><p>推演记录来自你的真实沙盘会话。可按项目、状态、角色和创建时间筛选。</p></div>
        <Link to="/membership/upgrade"><Crown size={16} />升级会员，保留更多记录</Link>
      </section>
      <section className="sb-history-filters" aria-label="历史推演筛选">
        <label className="sb-search-field"><Search size={16} /><input aria-label="搜索项目名称或关键词" onChange={(event) => { setQuery(event.target.value); setPage(1); }} placeholder="搜索项目名称 / 关键词" value={query} /></label>
        <label><span>状态</span><select aria-label="全部状态" onChange={(event) => { setStatus(event.target.value); setPage(1); }} value={status}><option value="all">全部状态</option><option value="completed">已完成</option><option value="running">进行中</option><option value="queued">等待执行</option><option value="draft">草稿</option><option value="failed">失败</option><option value="canceled">已取消</option></select></label>
        <label><span>参与角色</span><select aria-label="全部角色" onChange={(event) => { setRole(event.target.value); setPage(1); }} value={role}><option value="all">全部角色</option>{roleOptions.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
        <label><CalendarDays size={15} /><select aria-label="选择时间范围" onChange={(event) => { setDateRange(event.target.value); setPage(1); }} value={dateRange}><option value="all">选择时间范围</option><option value="7">最近 7 天</option><option value="30">最近 30 天</option><option value="90">最近 90 天</option></select></label>
        <button onClick={reset} type="button"><RotateCcw size={15} />重置筛选</button>
      </section>
      <section className="sb-history-table">
        <header><span>项目名称 / 描述</span><span>参与角色</span><span>推演时间</span><span>状态</span><span>综合评分</span><span>风险等级</span><span>操作</span></header>
        {loading ? <div className="sb-table-state"><Sparkles size={20} />正在读取推演记录...</div> : rows.length ? rows.map((session) => <HistoryRow key={session.id} session={session} />) : <div className="sb-table-state"><FileText size={22} /><strong>没有符合条件的推演记录</strong><span>调整筛选条件，或开始一轮新的商业沙盘。</span><Link to="/sandbox">开始新推演<ArrowRight size={15} /></Link></div>}
      </section>
      <footer className="sb-history-pagination">
        <span>共 {filtered.length} 条</span>
        <div><button aria-label="上一页" disabled={safePage <= 1} onClick={() => setPage((value) => Math.max(1, value - 1))} type="button"><ArrowLeft size={16} /></button><b>{safePage}</b><button aria-label="下一页" disabled={safePage >= pageCount} onClick={() => setPage((value) => Math.min(pageCount, value + 1))} type="button"><ArrowRight size={16} /></button></div>
        <span>{pageSize} 条 / 页</span>
      </footer>
    </SandboxFrame>
  );
}

function HistoryRow({ session }: { session: SandboxSession }) {
  const report = session.report;
  const score = report?.score ? report.score.toFixed(1) : "--";
  const href = sessionHref(session);
  return <article><div className="sb-history-project"><span aria-hidden="true">{session.product?.slice(0, 1) || "沙"}</span><div><strong>{session.product || session.goal}</strong><small>{session.goal}</small></div></div><div className="sb-history-roles">{session.roles.slice(0, 3).map((item) => <em key={item}>{shortRole(item)}</em>)}{session.roles.length > 3 && <b>+{session.roles.length - 3}</b>}</div><time>{formatDate(session.created_at)}</time><span className={`sb-status-pill is-${session.status}`}>{statusLabel(session.status)}</span><strong className="sb-history-score">{score === "--" ? score : `★ ${score}`}</strong><span className={`sb-risk-badge is-${report?.risk_level || "unknown"}`}>{riskLabel(report?.risk_level)}</span><Link to={href}>{session.status === "completed" ? "查看报告" : "继续处理"}</Link></article>;
}

function sessionHref(session: SandboxSession) {
  if (session.status === "completed") return `/sandbox/sessions/${session.id}/report`;
  if (session.status === "queued" || session.status === "running" || session.status === "failed" || session.status === "canceled") return `/sandbox/run?session=${session.id}`;
  if (session.intake?.status === "questions") return `/sandbox/questions?session=${session.id}`;
  if (!session.roles.length) return `/sandbox/roles?session=${session.id}`;
  return `/sandbox/start?session=${session.id}`;
}

function shortRole(value: string) { return value.replace("视角", "").replace("代理商 / 渠道方", "渠道"); }
function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString("zh-CN", { hour12: false, month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }); }
function statusLabel(value: SandboxSession["status"]) { return ({ draft: "草稿", queued: "等待", running: "进行中", completed: "已完成", failed: "失败", canceled: "已取消" } as const)[value]; }
function riskLabel(value?: string) { return value === "high" ? "较高" : value === "low" ? "较低" : value === "medium" ? "中等" : "待评估"; }

export default SandboxHistoryView;

