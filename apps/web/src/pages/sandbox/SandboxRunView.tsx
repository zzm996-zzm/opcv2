import { AlertTriangle, CheckCircle2, CircleDashed, FileText, MessageCircleQuestion, PauseCircle, RefreshCw, Send } from "lucide-react";
import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import SandboxFrame from "../../components/sandbox/SandboxFrame";
import { sandboxApi, type SandboxFollowUp, type SandboxRun } from "../../lib/sandboxApi";

type Props = { run: SandboxRun; onStop: () => Promise<void>; onRetryRole: (role: string) => Promise<void> };
const roleNames: Record<string, string> = { customer: "目标客户", investor: "投资人", competitor: "竞争对手", channel: "渠道方", supply: "供应链运营", expert: "行业专家", skeptic: "悲观者", partner: "合伙人" };

export default function SandboxRunView({ run, onStop, onRetryRole }: Props) {
  const rows = useMemo(() => [...(run.run_roles ?? [])].sort((a, b) => a.seq - b.seq), [run.run_roles]);
  const [selected, setSelected] = useState(rows[0]?.role_code ?? run.roles[0] ?? "customer"); const [question, setQuestion] = useState(""); const [followUps, setFollowUps] = useState<SandboxFollowUp[]>([]); const [busy, setBusy] = useState(""); const [error, setError] = useState("");
  const completed = rows.filter((role) => role.status === "done").length; const active = run.status === "running"; const terminal = ["done", "partial", "failed", "ai_no_result"].includes(run.status); const activeRole = rows.find((role) => role.status === "running"); const queued = rows.filter((role) => role.status === "queued").length; const total = rows.length || run.roles.length; const progress = total ? completed / total * 100 : 0;
  async function action(key: string, work: () => Promise<void>) { if (busy) return; setBusy(key); setError(""); try { await work(); } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "操作失败，请稍后重试。"); } finally { setBusy(""); } }
  async function ask() { const value = question.trim(); if (!value) return; await action("ask", async () => { const message = await sandboxApi.askRole(run.id, { role_code: selected, question: value }); setFollowUps((current) => [...current, message]); setQuestion(""); }); }
  return <SandboxFrame copilotMode="run" copilotProgress={rows.length ? completed / rows.length * 100 : 0} copilotProject={run.product.name} copilotRunID={run.id}>
    <header className="sb-run-heading"><div><small>模型推演</small><h1>{run.name || run.product.name}</h1><p>{active ? "各角色正在相互隔离的会话中分析" : terminal ? "本轮角色分析已结束" : "等待开始"}</p></div><strong>{statusLabel(run.status)}</strong></header>
    <section className="sb-run-role-strip" aria-live="polite"><header><strong>已完成 {completed}/{total} 角色</strong><span>{activeRole ? `${roleNames[activeRole.role_code] ?? activeRole.role_code} 正在分析` : active && queued ? `任务已提交，${queued} 个角色正在排队` : run.status === "partial" ? "部分完成，可基于现有结果生成报告" : "进度来自服务端角色状态"}</span></header><i className="sb-run-progress-track" aria-label={`当前完成进度 ${Math.round(progress)}%`}><span className={active && completed === 0 ? "is-indeterminate" : ""} style={{ width: `${progress}%` }} /></i><div className="sb-run-role-tabs">{run.roles.map((code) => <button className={selected === code ? "is-active" : ""} key={code} onClick={() => setSelected(code)} type="button">{roleNames[code] ?? code}</button>)}</div></section>
    <div className="sb-run-grid"><section className="sb-run-dialog">{rows.map((role) => <article className={`sb-run-message is-${role.status}`} id={`role-${role.role_code}`} key={role.role_code}><span className="sb-run-status-icon" aria-hidden="true">{role.status === "done" ? <CheckCircle2 size={18} /> : role.status === "failed" ? <AlertTriangle size={18} /> : <CircleDashed size={18} />}</span><div><header><strong>{roleNames[role.role_code] ?? role.role_code}</strong>{role.output?.stance ? <em className={`is-${role.output.stance}`}>{stanceLabel(role.output.stance)}</em> : null}<small>{roleStatusLabel(role.status)}</small></header><p>{role.output?.content || (role.status === "failed" ? `生成失败：${roleErrorMessage(role.error_code)}` : "等待角色输出")}</p>{role.status === "failed" ? <button disabled={Boolean(busy)} onClick={() => void action(`retry-${role.role_code}`, () => onRetryRole(role.role_code))} type="button"><RefreshCw size={15} />重试该角色</button> : null}</div></article>)}{!rows.length ? <div className="sb-state-card"><CircleDashed size={22} /><div><strong>角色执行快照尚未生成</strong><p>请返回确认页启动本轮推演。</p></div></div> : null}<footer className="sb-run-actions">{terminal && completed > 0 ? <Link className="is-primary" to={`/sandbox-runs/${run.id}/report`}><FileText size={16} />查看推演报告</Link> : null}{active ? <button className="sb-cancel-button" disabled={Boolean(busy)} onClick={() => void action("stop", onStop)} type="button"><PauseCircle size={16} />停止推演</button> : null}</footer></section>
      <aside className="sb-run-side-panel"><section><header><MessageCircleQuestion size={17} /><strong>推演后追问</strong></header><p className="sb-muted-copy">只使用所选角色的冻结配置和本轮输出，不读取其他角色观点。</p><textarea aria-label="追加追问" disabled={!terminal || completed === 0 || Boolean(busy)} onChange={(event) => setQuestion(event.target.value)} placeholder="选择角色后输入问题" value={question} /><button disabled={!question.trim() || Boolean(busy)} onClick={() => void ask()} type="button">{busy === "ask" ? "发送中..." : "发送追问"}<Send size={15} /></button></section></aside></div>
    {followUps.length ? <section className="sb-message-history" aria-label="角色追问记录">{followUps.map((item) => <article key={item.id}><small>{roleNames[item.role_code] ?? item.role_code}</small><p><b>{item.question}</b>{item.answer}</p></article>)}</section> : null}{error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
  </SandboxFrame>;
}
function statusLabel(value: SandboxRun["status"]) { return ({ draft: "草稿", clarifying: "补充信息", ready: "准备就绪", running: "推演中", partial: "部分完成", done: "已完成", failed: "推演失败", ai_no_result: "无有效结果" } as const)[value]; }
function roleStatusLabel(value: string) { return ({ pending: "等待", queued: "排队", running: "分析中", done: "已完成", failed: "失败", cancelled: "已停止" } as Record<string, string>)[value] ?? value; }
function roleErrorMessage(value?: string) {
  return ({
    provider_timeout: "模型响应超时，请稍后重试。",
    provider_rate_limited: "模型请求过于频繁，请稍后重试。",
    provider_authentication_failed: "模型服务鉴权失败，请联系管理员检查配置。",
    provider_permission_denied: "模型服务额度或权限不足，请联系管理员。",
    provider_model_not_found: "当前模型不可用，请联系管理员检查模型配置。",
    provider_bad_request: "模型请求参数无效，请重试。",
    provider_unavailable: "模型服务暂时不可用，请稍后重试。",
    invalid_model_json: "模型返回格式不正确，请重试。",
    internal_error: "系统处理失败，请稍后重试。",
    cancelled: "本角色已停止。",
    run_timeout: "本轮推演超时，请重试。",
  } as Record<string, string>)[value ?? ""] ?? "暂时无法生成，请稍后重试。";
}
function stanceLabel(value: string) { return value === "support" ? "支持" : value === "oppose" ? "反对" : "中立"; }
