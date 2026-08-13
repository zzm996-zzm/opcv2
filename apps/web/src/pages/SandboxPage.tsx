import { FileWarning } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import SandboxFrame from "../components/sandbox/SandboxFrame";
import SandboxQuotaDialog from "../components/sandbox/SandboxQuotaDialog";
import { apiErrorMessage } from "../lib/apiErrors";
import { sandboxApi, type SandboxRole, type SandboxRun } from "../lib/sandboxApi";
import SandboxHistoryView from "./sandbox/SandboxHistoryView";
import SandboxHomeView from "./sandbox/SandboxHomeView";
import SandboxQuestionsView from "./sandbox/SandboxQuestionsView";
import SandboxReportView from "./sandbox/SandboxReportView";
import SandboxRolesView from "./sandbox/SandboxRolesView";
import SandboxRunView from "./sandbox/SandboxRunView";
import SandboxStartView from "./sandbox/SandboxStartView";
import SandboxSummaryView from "./sandbox/SandboxSummaryView";

type SandboxVariant = "home" | "new" | "run" | "report" | "history" | "quota";

export default function SandboxPage({ variant = "home" }: { variant?: SandboxVariant }) {
  const navigate = useNavigate(); const location = useLocation(); const params = useParams();
  const query = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const rawID = params.runId ?? query.get("run") ?? query.get("session"); const runID = rawID && /^\d+$/.test(rawID) ? Number(rawID) : null;
  const defaultStep = location.pathname === "/sandbox/start" ? 4 : 1;
  const step = Math.max(1, Math.min(4, Number(query.get("step") || defaultStep)));
  const [run, setRun] = useState<SandboxRun | null>(null); const [roles, setRoles] = useState<SandboxRole[]>([]); const [error, setError] = useState(""); const [loading, setLoading] = useState(Boolean(runID));
  const lastEventID = useRef(0); const streamAbort = useRef<AbortController | null>(null);
  const accept = useCallback((next: SandboxRun) => { setRun((current) => current && current.id === next.id && next.revision < current.revision ? current : next); return next; }, []);

  useEffect(() => {
    if (!runID) { setLoading(false); return; }
    let active = true; setLoading(true); setError("");
    void sandboxApi.getRun(runID).then((payload) => { if (active) accept(payload); }).catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "沙盘推演加载失败，请返回历史记录重试。")); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [accept, runID]);

  useEffect(() => {
    if (variant !== "new" || !runID || step < 3) return;
    let active = true; void sandboxApi.listRoles().then((payload) => { if (active) setRoles(payload.roles ?? []); }).catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "角色加载失败，请稍后重试。")); });
    return () => { active = false; };
  }, [runID, step, variant]);

  useEffect(() => {
    if (variant !== "run" || !runID || !run || !["running", "ready"].includes(run.status)) return;
    streamAbort.current?.abort(); const controller = new AbortController(); streamAbort.current = controller; let active = true;
    void sandboxApi.streamRun(runID, lastEventID.current, (event) => { lastEventID.current = event.id; if (active && ["run_done", "error", "role_done", "role_failed"].includes(event.event)) void sandboxApi.getRun(runID).then(accept).catch(() => undefined); }, controller.signal).catch(() => undefined);
    return () => { active = false; controller.abort(); };
  }, [accept, run, runID, variant]);

  useEffect(() => () => streamAbort.current?.abort(), []);

  async function createRun(initialIdea: string) {
    const created = await sandboxApi.createRun({ name: initialIdea.slice(0, 80), product: { name: initialIdea }, context: { extra: initialIdea } });
    accept(created); navigate(`/sandbox/new?run=${created.id}&step=${created.status === "clarifying" ? 2 : 3}`); return created;
  }
  async function answer(input: { revision: number; answers?: { key: string; value: string; skipped?: boolean }[]; skip?: boolean }) { if (!run) throw new Error("当前沙盘不存在"); return accept(await sandboxApi.answerRun(run.id, input)); }
  async function saveRoles(selected: string[]) { if (!run) throw new Error("当前沙盘不存在"); const updated = accept(await sandboxApi.setRoles(run.id, { revision: run.revision, roles: selected })); navigate(`/sandbox/new?run=${run.id}&step=4`); return updated; }
  async function start() { if (!run) return; const started = accept(await sandboxApi.startRun(run.id)); navigate(`/sandbox-runs/${started.id}`); }
  async function stop() { if (run) accept(await sandboxApi.stopRun(run.id)); }
  async function retryRole(role: string) { if (run) accept(await sandboxApi.retryRole(run.id, role)); }

  if (variant === "home") return <SandboxHomeView onCreate={createRun} />;
  if (variant === "new" && !runID) return <SandboxHomeView onCreate={createRun} />;
  if (variant === "history") return <SandboxHistoryView />;
  if (variant === "quota") return <><SandboxHomeView onCreate={createRun} /><SandboxQuotaDialog limit={-1} used={0} onClose={() => navigate("/sandbox")} /></>;
  if (loading) return <LoadingSession />;
  if (!runID || !run) return <MissingSession message={error || "没有找到当前沙盘推演。"} />;
  if (variant === "report") return <SandboxReportView run={run} loading={loading} />;
  if (variant === "run") return <SandboxRunView run={run} onStop={stop} onRetryRole={retryRole} />;
  if (step === 2 && (run.status === "clarifying" || run.next_questions?.length)) return <SandboxQuestionsView onAnswer={answer} onFinished={(updated) => { accept(updated); navigate(`/sandbox/new?run=${updated.id}&step=${updated.status === "ready" ? 3 : 2}`); }} run={run} />;
  if (step === 2) return <SandboxSummaryView run={run} />;
  if (step === 3) return <SandboxRolesView onSave={saveRoles} roles={roles} run={run} />;
  if (step === 4) return <SandboxStartView onStart={start} roles={roles} run={run} />;
  return <SandboxSummaryView run={run} />;
}

function LoadingSession() { return <SandboxFrame copilotMode="home"><section className="sb-empty-panel"><span className="sb-loading-spinner" /><h1>正在加载商业沙盘</h1><p>正在同步服务端运行快照，请稍候。</p></section></SandboxFrame>; }
function MissingSession({ message }: { message: string }) { return <SandboxFrame copilotMode="home"><section className="sb-empty-panel"><FileWarning size={30} /><h1>当前页面缺少沙盘会话</h1><p>{message}</p><div><Link to="/sandbox">返回首页</Link><Link to="/sandbox/new">创建新推演</Link></div></section></SandboxFrame>; }
