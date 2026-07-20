import { AlertTriangle, CircleDashed, ExternalLink, MessageCircleQuestion, PauseCircle, RefreshCw, Send, ShieldAlert, Sparkles, TrendingUp } from "lucide-react";
import { useMemo, useState } from "react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import type { SandboxMessage, SandboxSession } from "../../lib/sandboxApi";

type SandboxRunViewProps = {
  loading?: boolean;
  messages: SandboxMessage[];
  onAsk: (role: string, question: string) => Promise<void>;
  onCancel: () => Promise<void>;
  onRetry: () => Promise<void>;
  session: SandboxSession;
};

function SandboxRunView({ loading = false, messages, onAsk, onCancel, onRetry, session }: SandboxRunViewProps) {
  const [role, setRole] = useState(session.roles[0] ?? "");
  const [question, setQuestion] = useState("");
  const [sending, setSending] = useState(false);
  const [actioning, setActioning] = useState(false);
  const [error, setError] = useState("");
  const completed = session.status === "completed";
  const active = session.status === "queued" || session.status === "running";
  const statusText = statusLabel(session.status, session.current_step);
  const roleRows = useMemo(() => session.roles.map((label, index) => {
    const summaries = session.report?.role_summaries ?? [];
    const summary = summaries.find((item) => item.role === label);
    return { label, summary, index };
  }), [session.report?.role_summaries, session.roles]);

  async function ask() {
    const value = question.trim();
    if (!completed || !role || !value || sending) return;
    setSending(true);
    setError("");
    try {
      await onAsk(role, value);
      setQuestion("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "追问发送失败，请稍后重试。");
    } finally {
      setSending(false);
    }
  }

  async function runAction(action: () => Promise<void>) {
    if (actioning) return;
    setActioning(true);
    setError("");
    try { await action(); } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "操作失败，请稍后重试。"); } finally { setActioning(false); }
  }

  return (
    <SandboxFrame className={loading ? "sb-is-loading" : ""} copilotMode="run" copilotProgress={session.progress_percent} copilotProject={session.intake?.initial_idea || session.goal}>
      <header className="sb-run-heading">
        <div><h1>{session.product || session.goal}</h1><div><span className={`sb-status-pill is-${session.status}`}>{statusText}</span><span className="sb-run-round">第 {Math.max(session.run_attempt, 1)} 轮推演</span></div></div>
        <small>{session.progress_percent}% · {session.current_step || "准备中"}</small>
      </header>
      <section className="sb-run-role-summary">
        <header><strong>本轮参与角色（{session.roles.length}）</strong><span>{active ? "AI 正在逐个分析角色视角" : completed ? "已完成角色观点汇总" : "等待重新启动"}</span></header>
        <div className="sb-run-role-tabs">
          {session.roles.map((label) => <button className={label === role ? "is-active" : ""} key={label} onClick={() => setRole(label)} type="button">{label}</button>)}
        </div>
      </section>
      <div className="sb-run-grid">
        <section className="sb-run-dialog">
          {active && <div className="sb-run-progress-card"><div className="sb-run-progress-orb"><Sparkles size={30} /></div><div><strong>{session.status === "queued" ? "已进入推演队列" : "AI 正在生成角色观点"}</strong><p>{session.current_step === "generating_report" ? "正在汇总机会、风险与验证动作，请稍候。" : "会话已保存，页面会持续同步最新进度。"}</p></div><b>{session.progress_percent}%</b></div>}
          {completed && roleRows.map(({ label, summary, index }) => (
            <article className="sb-run-message" key={label}>
              <time>{String(index + 1).padStart(2, "0")}</time>
              <strong>{label}</strong>
              <p>{summary?.view ?? "该角色观点已生成，进入报告查看完整分析。"}</p>
            </article>
          ))}
          {session.status === "failed" && <div className="sb-state-card is-error"><AlertTriangle size={22} /><div><strong>本轮推演没有完成</strong><p>{session.error_message || "服务暂时无法生成报告，请重试。"}</p></div></div>}
          {session.status === "canceled" && <div className="sb-state-card"><PauseCircle size={22} /><div><strong>本轮推演已取消</strong><p>可以保留当前配置并重新启动一轮推演。</p></div></div>}
          {!active && !completed && session.status !== "failed" && session.status !== "canceled" && <div className="sb-state-card"><CircleDashed size={22} /><div><strong>等待开始推演</strong><p>返回准备页确认角色和设置后开始。</p></div></div>}
          {completed && (
            <footer className="sb-run-actions">
              <Link className="is-primary" to={`/sandbox/sessions/${session.id}/report`}>生成推演报告<ExternalLink size={16} /></Link>
              <button className="is-secondary" disabled={actioning} onClick={() => void runAction(onRetry)} type="button"><RefreshCw size={16} />重新推演</button>
            </footer>
          )}
          {active && <button className="sb-cancel-button" disabled={actioning} onClick={() => void runAction(onCancel)} type="button"><PauseCircle size={16} />取消本轮推演</button>}
          {(session.status === "failed" || session.status === "canceled") && <button className="sb-retry-button" disabled={actioning} onClick={() => void runAction(onRetry)} type="button"><RefreshCw size={16} />重新推演</button>}
        </section>
        <aside className="sb-run-side-panel">
          <section><header><Sparkles size={17} /><strong>调整变量</strong></header><label><span>当前推演深度</span><b>{session.settings?.depth === "deep" ? "深度" : "标准"}</b></label><label><span>输出内容</span><b>{session.settings?.generate_outline ? "结构化大纲" : "精简报告"}</b></label></section>
          <section><header><TrendingUp size={17} /><strong>当前识别的机会</strong></header>{(session.report?.opportunity_analysis ?? []).slice(0, 3).map((item) => <p className="sb-opportunity" key={item.title}>{item.title}</p>)}{!session.report?.opportunity_analysis?.length && <p className="sb-muted-copy">完成推演后显示机会识别。</p>}</section>
          <section><header><ShieldAlert size={17} /><strong>当前识别的风险</strong></header>{(session.report?.risk_analysis ?? []).slice(0, 3).map((item) => <p className="sb-risk" key={item.title}>{item.title}</p>)}{!session.report?.risk_analysis?.length && <p className="sb-muted-copy">完成推演后显示风险识别。</p>}</section>
          <section className="sb-follow-up"><header><MessageCircleQuestion size={17} /><strong>追加追问</strong></header><textarea aria-label="追加追问" disabled={!completed || sending} onChange={(event) => setQuestion(event.target.value)} placeholder={completed ? "输入问题，进一步追问任意角色..." : "推演完成后可继续追问"} value={question} /><button disabled={!completed || sending || !question.trim()} onClick={() => void ask()} type="button">{sending ? "发送中..." : "发送追问"}<Send size={15} /></button></section>
        </aside>
      </div>
      {messages.length > 0 && <section className="sb-message-history" aria-label="角色追问记录">{messages.map((message) => <article key={message.id}><small>{message.role}</small><p><b>{message.question}</b>{message.answer}</p></article>)}</section>}
      {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
    </SandboxFrame>
  );
}

function statusLabel(status: SandboxSession["status"], step: string) {
  if (status === "queued") return "等待执行";
  if (status === "running") return step === "storing_report" ? "正在整理报告" : "推演中";
  if (status === "completed") return "已完成";
  if (status === "failed") return "推演失败";
  if (status === "canceled") return "已取消";
  return "草稿";
}

export default SandboxRunView;
