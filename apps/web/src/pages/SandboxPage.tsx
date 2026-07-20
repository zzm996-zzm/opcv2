import { FileWarning } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";

import SandboxFrame from "../components/sandbox/SandboxFrame";
import SandboxQuotaDialog from "../components/sandbox/SandboxQuotaDialog";
import { apiErrorMessage } from "../lib/apiErrors";
import { membershipApi, type MembershipUsageItem } from "../lib/membershipApi";
import { quotaKeys } from "../lib/quotaUsage";
import { sandboxApi, type SandboxMessage, type SandboxOptions, type SandboxRunSettings, type SandboxSession } from "../lib/sandboxApi";
import SandboxHistoryView from "./sandbox/SandboxHistoryView";
import SandboxHomeView from "./sandbox/SandboxHomeView";
import SandboxQuestionsView from "./sandbox/SandboxQuestionsView";
import SandboxReportView from "./sandbox/SandboxReportView";
import SandboxRolesView from "./sandbox/SandboxRolesView";
import SandboxRunView from "./sandbox/SandboxRunView";
import SandboxStartView from "./sandbox/SandboxStartView";
import SandboxSummaryView from "./sandbox/SandboxSummaryView";

type SandboxVariant = "home" | "setup" | "roles" | "start" | "questions" | "run" | "report" | "history" | "quota";

type SandboxPageProps = {
  variant?: SandboxVariant;
};

function SandboxPage({ variant = "home" }: SandboxPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();
  const query = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const rawSessionID = params.sessionId ?? query.get("session");
  const sessionID = rawSessionID && Number.isFinite(Number(rawSessionID)) ? Number(rawSessionID) : null;
  const exampleKey = query.get("example");
  const routedSession = useMemo(() => (location.state as { sandboxSession?: SandboxSession } | null)?.sandboxSession ?? null, [location.state]);
  const [session, setSession] = useState<SandboxSession | null>(routedSession);
  const [sessions, setSessions] = useState<SandboxSession[]>([]);
  const [examples, setExamples] = useState<SandboxSession[]>([]);
  const [options, setOptions] = useState<SandboxOptions | null>(null);
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const [messages, setMessages] = useState<SandboxMessage[]>([]);
  const [loadingSession, setLoadingSession] = useState(Boolean(sessionID && !routedSession));
  const [loadingList, setLoadingList] = useState(variant === "history");
  const [error, setError] = useState("");
  const sandboxUsage = usage.find((item) => item.key === quotaKeys.sandboxRuns);

  const acceptSession = useCallback((incoming: SandboxSession) => {
    setSession((current) => chooseNewestSession(current, incoming));
    setSessions((current) => [incoming, ...current.filter((item) => item.id !== incoming.id)]);
    return incoming;
  }, []);

  useEffect(() => {
    if (!sessionID) {
      setLoadingSession(false);
      return;
    }
    let active = true;
    if (!routedSession || routedSession.id !== sessionID) setLoadingSession(true);
    setError("");
    void sandboxApi.getSession(sessionID)
      .then((payload) => { if (active) acceptSession(payload); })
      .catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "沙盘会话加载失败，请返回历史记录重试。")); })
      .finally(() => { if (active) setLoadingSession(false); });
    return () => { active = false; };
  }, [acceptSession, routedSession, sessionID]);

  useEffect(() => {
    if (variant !== "roles" && variant !== "start") return;
    let active = true;
    void sandboxApi.getOptions()
      .then((payload) => { if (active) setOptions(payload); })
      .catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "沙盘选项加载失败，请稍后重试。")); });
    return () => { active = false; };
  }, [variant]);

  useEffect(() => {
    if (variant !== "history") return;
    let active = true;
    setLoadingList(true);
    void Promise.all([sandboxApi.listSessions(100), sandboxApi.listExamples()])
      .then(([payload, examplePayload]) => { if (active) { setSessions(payload.sessions ?? []); setExamples(examplePayload.sessions ?? []); } })
      .catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "历史推演加载失败，请稍后重试。")); })
      .finally(() => { if (active) setLoadingList(false); });
    return () => { active = false; };
  }, [variant]);

  useEffect(() => {
    if (variant !== "report" || !exampleKey) return;
    let active = true;
    setLoadingSession(true);
    void sandboxApi.listExamples()
      .then((payload) => {
        if (!active) return;
        const selected = (payload.sessions ?? []).find((item) => item.example_key === exampleKey);
        setExamples(payload.sessions ?? []);
        if (selected) acceptSession(selected);
        else setError("没有找到这条示例推演报告。");
      })
      .catch((requestError) => { if (active) setError(apiErrorMessage(requestError, "示例报告加载失败，请稍后重试。")); })
      .finally(() => { if (active) setLoadingSession(false); });
    return () => { active = false; };
  }, [acceptSession, exampleKey, variant]);

  useEffect(() => {
    if (variant !== "start" && variant !== "quota") return;
    let active = true;
    void membershipApi.usage()
      .then((payload) => { if (active) setUsage(payload.usage ?? []); })
      .catch(() => { if (active) setUsage([]); });
    return () => { active = false; };
  }, [variant]);

  useEffect(() => {
    if (variant !== "run" || !sessionID) return;
    let active = true;
    void sandboxApi.listMessages(sessionID)
      .then((payload) => { if (active) setMessages(payload.messages ?? []); })
      .catch(() => { if (active) setMessages([]); });
    return () => { active = false; };
  }, [sessionID, session?.status, variant]);

  useEffect(() => {
    if (variant !== "run" || !sessionID || !session || (session.status !== "queued" && session.status !== "running")) return;
    let active = true;
    const timer = window.setInterval(() => {
      void sandboxApi.getStatus(sessionID).then((payload) => { if (active) acceptSession(payload); }).catch(() => undefined);
    }, 2000);
    return () => { active = false; window.clearInterval(timer); };
  }, [acceptSession, session, sessionID, variant]);

  async function createIntake(initialIdea: string) {
    const created = await sandboxApi.createIntake(initialIdea);
    acceptSession(created);
    navigate(`/sandbox/questions?session=${created.id}`, { state: { sandboxSession: created } });
    return created;
  }

  async function answerQuestion(questionKey: string, answer: string, skipped: boolean) {
    if (!sessionID) throw new Error("当前沙盘会话不存在。");
    const updated = await sandboxApi.answerIntakeQuestion(sessionID, questionKey, { answer, skipped });
    return acceptSession(updated);
  }

  async function completeIntake() {
    if (!sessionID) throw new Error("当前沙盘会话不存在。");
    const updated = await sandboxApi.completeIntake(sessionID);
    return acceptSession(updated);
  }

  async function saveRoles(roles: string[]) {
    if (!sessionID) throw new Error("当前沙盘会话不存在。");
    const updated = await sandboxApi.updateDraft(sessionID, { roles });
    acceptSession(updated);
    navigate(`/sandbox/start?session=${sessionID}`, { state: { sandboxSession: updated } });
    return updated;
  }

  async function startSandbox(settings: SandboxRunSettings) {
    if (!sessionID) throw new Error("当前沙盘会话不存在。");
    const configured = await sandboxApi.updateDraft(sessionID, { settings });
    acceptSession(configured);
    const queued = await sandboxApi.runSession(sessionID);
    acceptSession(queued);
    const usagePayload = await membershipApi.usage().catch(() => null);
    if (usagePayload) setUsage(usagePayload.usage ?? []);
    navigate(`/sandbox/run?session=${sessionID}`, { state: { sandboxSession: queued } });
  }

  async function cancelSandbox() {
    if (!sessionID) return;
    acceptSession(await sandboxApi.cancelSession(sessionID));
  }

  async function retrySandbox() {
    if (!sessionID) return;
    acceptSession(await sandboxApi.retrySession(sessionID));
  }

  async function askRole(role: string, questionText: string) {
    if (!sessionID) return;
    const message = await sandboxApi.askRole(sessionID, { role, question: questionText });
    setMessages((current) => [...current, message]);
  }

  if (variant === "home") return <SandboxHomeView onCreate={createIntake} />;
  if (variant === "history") return <><SandboxHistoryView examples={examples} loading={loadingList} sessions={sessions} />{error ? <GlobalError message={error} /> : null}</>;
  if (variant === "quota") return <><SandboxHomeView onCreate={createIntake} /><SandboxQuotaDialog limit={sandboxUsage?.limit ?? 1} onClose={() => navigate("/sandbox")} used={sandboxUsage?.used ?? sandboxUsage?.limit ?? 1} /></>;

  if (variant === "report" && exampleKey) return <SandboxReportView loading={loadingSession} session={session?.example_key === exampleKey ? session : null} />;
  if (!sessionID) return <MissingSession />;
  if (loadingSession && !session) return <LoadingSession />;
  if (!session || session.id !== sessionID) return <MissingSession message={error || "没有找到当前沙盘会话。"} />;

  if (variant === "questions") return <SandboxQuestionsView onComplete={completeIntake} onFinished={(updated) => navigate(`/sandbox/setup?session=${updated.id}`, { state: { sandboxSession: updated } })} onSave={answerQuestion} session={session} />;
  if (variant === "setup") return <SandboxSummaryView session={session} />;
  if (variant === "roles") return options ? <SandboxRolesView onSave={saveRoles} roles={options.roles} session={session} /> : <LoadingSession message={error || "正在加载推演角色..."} />;
  if (variant === "start") return options ? <SandboxStartView onStart={startSandbox} options={options} session={session} usage={sandboxUsage} /> : <LoadingSession message={error || "正在加载推演设置..."} />;
  if (variant === "run") return <SandboxRunView loading={loadingSession} messages={messages} onAsk={askRole} onCancel={cancelSandbox} onRetry={retrySandbox} session={session} />;
  if (variant === "report") return <SandboxReportView loading={loadingSession} session={session} />;
  return <MissingSession />;
}

function chooseNewestSession(current: SandboxSession | null, incoming: SandboxSession) {
  if (!current || current.id !== incoming.id) return incoming;
  if ((incoming.run_attempt ?? 0) < (current.run_attempt ?? 0)) return current;
  if ((incoming.run_attempt ?? 0) > (current.run_attempt ?? 0)) return incoming;
  const ranks: Record<SandboxSession["status"], number> = { draft: 0, queued: 1, running: 2, completed: 3, failed: 3, canceled: 3 };
  if (ranks[incoming.status] < ranks[current.status]) return current;
  const incomingTime = Date.parse(incoming.updated_at || "");
  const currentTime = Date.parse(current.updated_at || "");
  if (Number.isFinite(incomingTime) && Number.isFinite(currentTime) && incomingTime < currentTime) return current;
  return incoming;
}

function LoadingSession({ message = "正在加载商业沙盘..." }: { message?: string }) {
  return <SandboxFrame copilotMode="home"><section className="sb-empty-panel"><span className="sb-loading-spinner" /><h1>{message}</h1><p>正在同步服务端会话，请稍候。</p></section></SandboxFrame>;
}

function MissingSession({ message = "请先从商业沙盘首页创建推演。" }: { message?: string }) {
  return <SandboxFrame copilotMode="home"><section className="sb-empty-panel"><FileWarning size={30} /><h1>当前页面缺少沙盘会话</h1><p>{message}</p><Link to="/sandbox">返回商业沙盘首页</Link></section></SandboxFrame>;
}

function GlobalError({ message }: { message: string }) {
  return <div className="sb-global-error" role="alert">{message}</div>;
}

export default SandboxPage;
