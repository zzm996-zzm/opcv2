import { useEffect, useState, type MouseEvent, type ReactNode } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { membershipApi, type MembershipUsageItem } from "../lib/membershipApi";
import { quotaKeys, quotaSummary } from "../lib/quotaUsage";
import { sandboxApi, type SandboxMessage, type SandboxRole, type SandboxSession } from "../lib/sandboxApi";

type SandboxVariant =
  | "home"
  | "setup"
  | "roles"
  | "start"
  | "questions"
  | "run"
  | "report"
  | "history"
  | "quota";

type SandboxPageProps = {
  variant?: SandboxVariant;
};

const sandboxPillars = [
  ["多角色推演", "五大核心视角全面洞察"],
  ["机会与风险识别", "推演发现关键影响因素"],
  ["科学决策支持", "数据驱动更优决策"],
  ["推演历史沉淀", "复盘迭代持续优化"]
] as const;

const roles = [
  ["用户视角", "评估产品体验与价值", "推荐优先", "user"],
  ["投资人视角", "评估市场潜力与回报", "热门选择", "investor"],
  ["代理商 / 渠道方视角", "评估项目落地可行性", "渠道必选", "channel"],
  ["竞争对手视角", "评估竞争格局与策略", "深度分析", "competitor"],
  ["运营视角", "评估执行与增长策略", "运营必选", "operator"]
] as const;

const recognizedInfo = [
  ["产品方向", "低卡代餐奶昔"],
  ["目标人群", "上班族"],
  ["定价区间", "30元/杯"],
  ["销售场景", "写字楼附近线下门店"]
] as const;

const questionCards = [
  ["目标用户画像是怎样的？", "如年龄、职业、使用场景、核心痛点", "已回答"],
  ["产品核心功能有哪些？", "列出主要功能点", "已回答"],
  ["产品差异化优势是什么？", "相比现有竞品的独特价值", "已回答"],
  ["主要竞争对手有哪些？", "已知的直接/间接竞品", "可跳过"],
  ["运营与获客计划？", "你打算如何触达和转化用户", "可跳过"],
  ["预算与资源情况？", "团队规模、预算范围、技术资源等", "可跳过"]
] as const;

type SandboxHistoryRow = {
  title: string;
  detail: string;
  role: string;
  time: string;
  status: string;
  score: string;
  risk: string;
  href: string;
};

function SandboxPage({ variant = "home" }: SandboxPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { sessionId } = useParams();
  const [sessions, setSessions] = useState<SandboxSession[]>([]);
  const [selectedSession, setSelectedSession] = useState<SandboxSession | null>(null);
  const [sessionLoading, setSessionLoading] = useState(false);
  const [sessionError, setSessionError] = useState("");
  const [roleCatalog, setRoleCatalog] = useState<SandboxRole[]>([]);
  const [isStarting, setIsStarting] = useState(false);
  const [startError, setStartError] = useState("");
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const sandboxQuota = quotaSummary(usage, quotaKeys.sandboxRuns, "商业沙盘");

  useEffect(() => {
    let active = true;
    void sandboxApi.listSessions(10)
      .then((payload) => {
        if (active) setSessions(payload.sessions);
      })
      .catch(() => {
        if (active) setSessions([]);
      });
    void membershipApi.usage()
      .then((payload) => {
        if (active) setUsage(payload.usage ?? []);
      })
      .catch(() => {
        if (active) setUsage([]);
      });
    void sandboxApi.listRoles()
      .then((payload) => {
        if (active) setRoleCatalog(payload.roles);
      })
      .catch(() => {
        if (active) setRoleCatalog([]);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const querySessionId = new URLSearchParams(location.search).get("session");
    const rawSessionId = sessionId ?? querySessionId;
    if (!rawSessionId) return;
    const id = Number(rawSessionId);
    if (!Number.isFinite(id) || id <= 0) return;
    let active = true;
    setSessionLoading(true);
    setSessionError("");
    sandboxApi
      .getSession(id)
      .then((payload) => {
        if (active) setSelectedSession(payload);
      })
      .catch(() => {
        if (active) {
          setSelectedSession(null);
          setSessionError("沙盘会话加载失败，请返回历史记录重试。");
        }
      })
      .finally(() => {
        if (active) setSessionLoading(false);
      });
    return () => {
      active = false;
    };
  }, [location.search, sessionId]);

  const pollingSessionId = selectedSession && (selectedSession.status === "queued" || selectedSession.status === "running")
    ? selectedSession.id
    : null;

  useEffect(() => {
    if (!pollingSessionId) return;
    let active = true;
    const timer = window.setInterval(() => {
      void sandboxApi.getStatus(pollingSessionId).then((session) => {
        if (!active) return;
        setSelectedSession(session);
        setSessions((current) => [session, ...current.filter((item) => item.id !== session.id)]);
      }).catch(() => {
        // Preserve the last known progress and allow the next poll to recover.
      });
    }, 2000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [pollingSessionId]);

  const routedSession = (location.state as { sandboxSession?: SandboxSession } | null)?.sandboxSession ?? null;
  const reportSession = selectedSession ?? routedSession ?? sessions.find((session) => session.status === "completed" && session.report) ?? sessions.find((session) => session.report) ?? null;

  async function startSandbox(event: MouseEvent<HTMLAnchorElement>) {
    event.preventDefault();
    if (isStarting) return;
    if (sandboxQuota.blocked) {
      setStartError("本月商业沙盘次数已用完，请升级套餐或等待下月重置。");
      return;
    }
    setIsStarting(true);
    setStartError("");
    try {
      if (!selectedSession) {
        setStartError("请先完成推演配置和角色选择。");
        return;
      }
      const queued = await sandboxApi.runSession(selectedSession.id);
      setSelectedSession(queued);
      setSessions((current) => [queued, ...current.filter((session) => session.id !== queued.id)]);
      const usagePayload = await membershipApi.usage().catch(() => null);
      if (usagePayload) setUsage(usagePayload.usage ?? []);
      navigate(`/sandbox/run?session=${queued.id}`, { state: { sandboxSession: queued } });
    } catch (error) {
      setStartError(apiErrorMessage(error, "推演启动失败，请稍后重试。"));
    } finally {
      setIsStarting(false);
    }
  }

  async function cancelSandbox() {
    const session = selectedSession ?? routedSession;
    if (!session) return;
    try {
      const canceled = await sandboxApi.cancelSession(session.id);
      setSelectedSession(canceled);
      setSessions((current) => [canceled, ...current.filter((item) => item.id !== canceled.id)]);
      setSessionError("");
    } catch (error) {
      setSessionError(apiErrorMessage(error, "暂时无法取消推演。"));
    }
  }

  async function retrySandbox() {
    const session = selectedSession ?? routedSession;
    if (!session) return;
    try {
      const queued = await sandboxApi.retrySession(session.id);
      setSelectedSession(queued);
      setSessions((current) => [queued, ...current.filter((item) => item.id !== queued.id)]);
      setSessionError("");
    } catch (error) {
      setSessionError(apiErrorMessage(error, "暂时无法重新推演。"));
    }
  }

  return (
    <V4PageShell className="sandbox-shell" showCopilotMini={false}>
      <section className={`sandbox-v2 sandbox-${variant}`} aria-label="商业沙盘">
        {variant === "home" && <SandboxHome />}
        {variant === "setup" && <SetupPage />}
        {variant === "roles" && <RolesPage roleCatalog={roleCatalog} session={selectedSession} />}
        {variant === "start" && <StartPage isStarting={isStarting} onStart={startSandbox} quota={sandboxQuota} session={selectedSession} startError={startError} />}
        {variant === "questions" && <QuestionsPage />}
        {variant === "run" && <RunPage loading={sessionLoading} loadError={sessionError} onCancel={cancelSandbox} onRetry={retrySandbox} session={selectedSession ?? routedSession} />}
        {variant === "report" && <ReportPage loading={sessionLoading} loadError={sessionError} session={reportSession} />}
        {variant === "history" && <HistoryPage sessions={sessions} />}
        {variant === "quota" && (
          <>
            <SandboxHome />
            <QuotaModal quota={sandboxQuota} />
          </>
        )}
      </section>
    </V4PageShell>
  );
}

function SandboxHome() {
  return (
    <div className="sandbox-home-layout">
      <main>
        <section className="sandbox-hero-v2">
          <div className="sandbox-hero-copy">
            <h1>商业沙盘</h1>
            <h2>多角色模拟未来，推演不同视角，判断项目机会与风险</h2>
            <p>从用户、竞争、运营、增长到风险，多维度模拟真实世界的商业逻辑，助力科学决策。</p>
            <div className="sandbox-start-card">
              <span aria-hidden="true">⌕</span>
              <p>请描述你要推演的项目 / 情况，并补充行业、目标用户、当前阶段、关键约束，以及你最想判断的问题</p>
              <Link to="/sandbox/setup">开始推演 →</Link>
              <Link to="/sandbox/history">沙盘历史</Link>
            </div>
          </div>
          <SandboxOrbit />
        </section>
        <section className="sandbox-pillar-strip">
          {sandboxPillars.map(([title, detail], index) => (
            <article key={title}>
              <b>{index + 1}</b>
              <strong>{title}</strong>
              <small>{detail}</small>
            </article>
          ))}
        </section>
      </main>
      <SandboxCopilot mode="home" />
    </div>
  );
}

function SetupPage() {
  const navigate = useNavigate();
  const [goal, setGoal] = useState("验证面向上班族的 AI 低卡代餐产品是否值得启动");
  const [targetUsers, setTargetUsers] = useState("上班族");
  const [product, setProduct] = useState("AI 低卡代餐奶昔");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");

  async function saveDraft() {
    if (saving) return;
    setSaving(true);
    setSaveError("");
    try {
      const session = await sandboxApi.createSession({ goal, targetUsers, product, roles: [] });
      navigate(`/sandbox/roles?session=${session.id}`);
    } catch (error) {
      setSaveError(apiErrorMessage(error, "沙盘配置保存失败，请稍后重试。"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <SandboxWorkLayout mode="setup">
      <SandboxStepper active={1} />
      <section className="sandbox-work-card setup-work">
        <header>
          <div>
            <h2>推演发起配置</h2>
            <small>AI 智能补充中</small>
          </div>
          <button type="button">重新输入</button>
        </header>
        <p>已识别你提供的信息，AI 正在分析并为你生成需要补充的关键信息，帮助构建更精准的沙盘模型。</p>
        <article className="setup-input-preview">
          <label><strong>推演目标</strong><textarea aria-label="推演目标" onChange={(event) => setGoal(event.target.value)} value={goal} /></label>
          <label><strong>目标用户</strong><input aria-label="目标用户" onChange={(event) => setTargetUsers(event.target.value)} value={targetUsers} /></label>
          <label><strong>产品或方案</strong><input aria-label="产品或方案" onChange={(event) => setProduct(event.target.value)} value={product} /></label>
        </article>
        <h3>为了更准确地推演，我们还需要了解以下关键信息</h3>
        <div className="setup-question-grid">
          {questionCards.map(([title, detail, state]) => (
            <article key={title}>
              <strong>{title}</strong>
              <small>{detail}</small>
              <span>{state}</span>
            </article>
          ))}
        </div>
        <button className="sandbox-primary-wide" disabled={saving || !goal.trim() || !targetUsers.trim() || !product.trim()} onClick={saveDraft} type="button">
          {saving ? "保存中..." : "保存并选择角色 →"}
        </button>
        {saveError ? <p role="alert">{saveError}</p> : null}
      </section>
    </SandboxWorkLayout>
  );
}

function RolesPage({ roleCatalog, session }: { roleCatalog: SandboxRole[]; session: SandboxSession | null }) {
  const navigate = useNavigate();
  const visibleRoles = roleCatalog.length ? roleCatalog : roles.map(([label, description, badge, key]) => ({ key, label, description, badge }));
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");

  useEffect(() => {
    if (session) setSelectedRoles(session.roles);
  }, [session]);

  function toggleRole(label: string) {
    setSelectedRoles((current) => current.includes(label) ? current.filter((role) => role !== label) : [...current, label]);
  }

  async function saveRoles() {
    if (!session || selectedRoles.length === 0 || saving) return;
    setSaving(true);
    setSaveError("");
    try {
      await sandboxApi.updateDraft(session.id, { roles: selectedRoles });
      navigate(`/sandbox/start?session=${session.id}`);
    } catch (error) {
      setSaveError(apiErrorMessage(error, "角色保存失败，请稍后重试。"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <SandboxWorkLayout mode="roles">
      <SandboxStepper active={2} />
      <section className="sandbox-work-card roles-work">
        <h2>你希望从谁的视角进行推演？</h2>
        <p>选择一个或多个角色，AI 将基于该角色的立场对话并给出建议。</p>
        <div className="role-select-grid">
          {visibleRoles.map((role) => (
            <button className={selectedRoles.includes(role.label) ? "selected" : ""} key={role.key} onClick={() => toggleRole(role.label)} type="button">
              <i className={`role-avatar ${role.key}`} />
              <b>{selectedRoles.includes(role.label) ? "✓" : ""}</b>
              <strong>{role.label}</strong>
              <small>{role.description}</small>
              <span>{role.badge}</span>
            </button>
          ))}
        </div>
        <aside className="role-help-band">
          <strong>为什么选择多个角色？</strong>
          <p>多角色视角能帮助你更全面地发现机会、识别风险，获得更立体的推演结果。</p>
        </aside>
        <footer>
          <Link to="/sandbox/setup">上一步</Link>
          <button disabled={!session || selectedRoles.length === 0 || saving} onClick={saveRoles} type="button">{saving ? "保存中..." : "确认角色，进入下一步 →"}</button>
        </footer>
        {!session ? <p role="alert">请先从推演配置页创建沙盘草稿。</p> : null}
        {saveError ? <p role="alert">{saveError}</p> : null}
      </section>
    </SandboxWorkLayout>
  );
}

function StartPage({
  isStarting,
  onStart,
  quota,
  session,
  startError
}: {
  isStarting: boolean;
  onStart: (event: MouseEvent<HTMLAnchorElement>) => void;
  quota: ReturnType<typeof quotaSummary>;
  session: SandboxSession | null;
  startError: string;
}) {
  return (
    <SandboxWorkLayout mode="start">
      <SandboxStepper active={3} />
      <section className="sandbox-work-card start-work">
        <div className="start-columns">
          <section>
            <header>
              <h2>推演设置信息</h2>
            </header>
            {[
              ["推演目标", session?.goal ?? "尚未配置"],
              ["目标用户", session?.target_users ?? "尚未配置"],
              ["产品 / 方案", session?.product ?? "尚未配置"]
            ].map(([label, value]) => (
              <article key={label}>
                <b>{label}</b>
                <span>{value}</span>
              </article>
            ))}
            <Link to="/sandbox/setup">查看完整信息 ›</Link>
          </section>
          <section>
            <h2>参与推演角色 <small>已选 {session?.roles.length ?? 0} 个</small></h2>
            <div className="start-role-grid">
              {(session?.roles ?? []).map((title) => (
                <article key={title}>
                  <i className="role-avatar user" />
                  <b>✓</b>
                  <strong>{title}</strong>
                  <small>参与本轮多角色推演</small>
                </article>
              ))}
            </div>
          </section>
        </div>
        <footer>
          <div>
            <strong>当前流程</strong>
            <span>多角色情景推演</span>
          </div>
          <div>
            <strong>输出内容</strong>
            <span>结构化模型推演报告</span>
          </div>
          <div className={quota.blocked ? "module-quota-inline depleted" : "module-quota-inline"}>
            <small>{quota.label}</small>
            <strong>{quota.value}</strong>
            {quota.blocked ? <Link to="/membership">升级套餐</Link> : <span>{quota.unit}</span>}
          </div>
          <Link aria-disabled={isStarting || quota.blocked} onClick={onStart} to="/sandbox/run">{isStarting ? "推演启动中..." : "开始推演 🚀"}</Link>
        </footer>
        {startError ? <p role="alert">{startError}</p> : null}
      </section>
    </SandboxWorkLayout>
  );
}

function QuestionsPage() {
  return (
    <SandboxWorkLayout mode="questions">
      <SandboxStepper active={1} />
      <section className="sandbox-work-card ask-work">
        <header>
          <div>
            <h2>推演发起配置</h2>
            <small>AI 动态提问中</small>
          </div>
          <button type="button">编辑</button>
        </header>
        <p>已识别你提供的初步想法，AI 正在分析并为你生成最关键的补充问题。</p>
        <div className="recognized-strip">
          {recognizedInfo.map(([label, value]) => (
            <article key={label}>
              <b>✓</b>
              <small>{label}</small>
              <strong>{value}</strong>
            </article>
          ))}
        </div>
        <section className="question-focus-card">
          <p>第 1 题 / 共 5 题</p>
          <h2>你的目标用户更具体是哪些上班族？</h2>
          <small>例如：白领女性 / 健身人群 / 加班人群 / 通勤族等，越具体越有助于推演。</small>
          <textarea aria-label="补充目标用户" placeholder="请详细描述你的目标用户特征，如：年龄段、性别、职业类型、生活习惯、核心需求等..." />
          <footer>
            <button type="button">下一题 →</button>
            <Link to="/sandbox/roles">暂时不清楚，留空继续</Link>
          </footer>
        </section>
      </section>
    </SandboxWorkLayout>
  );
}

function RunPage({
  session,
  loading,
  loadError,
  onCancel,
  onRetry
}: {
  session: SandboxSession | null;
  loading: boolean;
  loadError: string;
  onCancel: () => Promise<void>;
  onRetry: () => Promise<void>;
}) {
  const [messages, setMessages] = useState<SandboxMessage[]>([]);
  const [selectedRole, setSelectedRole] = useState("");
  const [question, setQuestion] = useState("");
  const [sending, setSending] = useState(false);
  const [messageError, setMessageError] = useState("");

  useEffect(() => {
    if (!session) return;
    setSelectedRole((current) => current || session.roles[0] || "");
    let active = true;
    sandboxApi.listMessages(session.id)
      .then((payload) => {
        if (active) setMessages(payload.messages);
      })
      .catch(() => {
        if (active) setMessages([]);
      });
    return () => {
      active = false;
    };
  }, [session]);

  async function sendQuestion() {
    if (!session || !selectedRole || !question.trim() || sending) return;
    setSending(true);
    setMessageError("");
    try {
      const message = await sandboxApi.askRole(session.id, { role: selectedRole, question: question.trim() });
      setMessages((current) => [...current, message]);
      setQuestion("");
    } catch (error) {
      setMessageError(apiErrorMessage(error, "追问失败，请稍后重试。"));
    } finally {
      setSending(false);
    }
  }

  const initialRows = session?.report?.role_summaries.map((summary) => ["本轮", summary.role, summary.view] as const) ?? [];

  if (loading) return <SandboxState title="正在加载沙盘会话..." />;
  if (!session) return <SandboxState title={loadError || "未找到可继续的沙盘会话"} />;

  return (
    <SandboxWorkLayout mode="run">
      <section className="run-work">
        <header>
          <div>
            <h1>{session?.product ?? "商业沙盘推演"}</h1>
            <span>{sandboxStatusLabel(session.status)}</span>
          </div>
          <small>第 {session.run_attempt || 1} 轮 · {session.progress_percent || 0}% · {session.current_step || session.status}</small>
        </header>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
        <nav className="run-role-tabs">
          {(session?.roles ?? []).map((item) => (
            <button className={selectedRole === item ? "active" : ""} key={item} onClick={() => setSelectedRole(item)} type="button">{item}</button>
          ))}
        </nav>
        <div className="run-grid">
          <section className="run-dialog">
            {initialRows.length === 0 ? <div className="module-empty-state" role="status">{session.status === "failed" ? session.error_message || "推演失败，可重新运行。" : session.status === "canceled" ? "推演已取消。" : "报告生成中，完成后将显示各角色推演结论。"}</div> : null}
            {initialRows.map(([time, role, text]) => (
              <article key={`${time}-${role}`}>
                <time>{time}</time>
                <strong>{role}</strong>
                <p>{text}</p>
              </article>
            ))}
            {messages.map((message) => (
              <article key={message.id}>
                <time>{new Date(message.created_at).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false })}</time>
                <strong>{message.role}</strong>
                <p><b>问：{message.question}</b><br />{message.answer}</p>
              </article>
            ))}
            <footer>
              {(session.status === "queued" || session.status === "running") ? <button onClick={() => void onCancel()} type="button">取消推演</button> : null}
              {(session.status === "failed" || session.status === "canceled") ? <button onClick={() => void onRetry()} type="button">重新推演</button> : null}
              {session.status === "completed" ? <Link to={`/sandbox/sessions/${session.id}/report`}>查看推演报告</Link> : null}
            </footer>
          </section>
          <aside className="run-variable-panel">
            <h2>追加追问</h2>
            {session.status !== "completed" ? <p>推演完成后可向已选角色追加追问。</p> : null}
            <textarea aria-label="追加追问" onChange={(event) => setQuestion(event.target.value)} placeholder="输入你的问题，进一步追问任意角色..." value={question} />
            <button disabled={session.status !== "completed" || !selectedRole || !question.trim() || sending} onClick={sendQuestion} type="button">{sending ? "发送中..." : "发送追问"}</button>
            {messageError ? <p role="alert">{messageError}</p> : null}
          </aside>
        </div>
      </section>
    </SandboxWorkLayout>
  );
}

function sandboxStatusLabel(status: SandboxSession["status"]) {
  if (status === "queued") return "等待执行";
  if (status === "running") return "推演中";
  if (status === "completed") return "推演完成";
  if (status === "failed") return "推演失败";
  if (status === "canceled") return "已取消";
  return "草稿";
}

function formatSessionTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { hour12: false });
}

function ReportPage({ session, loading, loadError }: { session: SandboxSession | null; loading: boolean; loadError: string }) {
  if (loading) return <SandboxState title="正在加载推演报告..." />;
  if (!session?.report) return <SandboxState title={loadError || "暂无可查看的推演报告"} />;
  const report = session.report;
  const reportAssumptions = report.assumptions ?? ["历史报告未记录结构化假设，使用前需重新核验输入条件。"];
  const reportEvidence = report.evidence_sources ?? [];
  const reportDisclaimer = report.disclaimer ?? "本报告为历史 AI 模型推演记录，不代表真实市场统计、收益承诺或已验证事实。";
  const visibleTitle = session.product;
  const visibleTime = formatSessionTime(session.updated_at);
  const visibleRolesCount = session.roles.length;
  const visibleMetrics = report.metrics.map((metric) => [
      metric.label,
      metric.value,
      "来自本次沙盘推演报告"
    ] as const);
  const visibleConclusions = [report.summary, ...report.risks.map((risk) => `风险提示：${risk}`)];
  const visibleRoleSummaries = report.role_summaries.map((item) => [item.role, item.view] as const);
  const visibleNextActions = report.next_actions.map((item, index) => [`行动 ${index + 1}`, item] as const);

  return (
    <SandboxWorkLayout mode="report">
      <section className="report-work">
        <header>
          <div>
            <p>商业沙盘 / 历史推演</p>
            <h1>{visibleTitle}</h1>
            <small>推演时间：{visibleTime} 参与角色数：{visibleRolesCount} 报告版本：V2.0 · 模型推演</small>
          </div>
        </header>
        <div className="report-metrics">
          {visibleMetrics.map(([label, value, detail]) => (
            <article key={label}>
              <small>{label}</small>
              <strong>{value}</strong>
              <p>{detail}</p>
            </article>
          ))}
        </div>
        <section className="report-summary">
          <h2>核心结论</h2>
          <p>{reportDisclaimer}</p>
          <ul>
            {visibleConclusions.map((item) => <li key={item}>{item}</li>)}
          </ul>
        </section>
        <div className="report-two-col">
          <article>
            <h2>关键假设</h2>
            <ul>{reportAssumptions.map((item) => <li key={item}>{item}</li>)}</ul>
          </article>
          <article>
            <h2>验证证据</h2>
            {reportEvidence.length === 0 ? <p>本次推演未接入外部验证证据，所有结论需通过访谈、实验或可信数据源复核。</p> : (
              <ul>{reportEvidence.map((item) => <li key={`${item.title}-${item.url}`}><a href={item.url} rel="noreferrer" target="_blank">{item.title}</a></li>)}</ul>
            )}
          </article>
        </div>
        <section className="role-summary-grid">
          {visibleRoleSummaries.map(([title, detail]) => (
            <article key={title}>
              <strong>{title}</strong>
              <small>{detail}</small>
            </article>
          ))}
        </section>
        <section className="next-action-row">
          {visibleNextActions.map(([title, detail], index) => (
            <article key={title}>
              <b>{index + 1}</b>
              <strong>{title}</strong>
              <small>{detail}</small>
            </article>
          ))}
        </section>
      </section>
    </SandboxWorkLayout>
  );
}

function toHistoryRow(session: SandboxSession) {
  return {
    title: session.product || session.goal,
    detail: session.goal,
    role: session.roles.join(" ") || "未选择角色",
    time: formatSessionTime(session.updated_at),
    status: sandboxStatusLabel(session.status),
    score: session.report ? `${session.report.score}/100` : "-",
    risk: session.report ? `${session.report.risks.length} 项` : "-",
    href: `/sandbox/sessions/${session.id}/report`
  };
}

function HistoryPage({ sessions }: { sessions: SandboxSession[] }) {
  const visibleRows: SandboxHistoryRow[] = sessions.map(toHistoryRow);

  return (
    <SandboxWorkLayout mode="history">
      <section className="history-work">
        <header>
          <div>
            <p>商业沙盘 / 历史推演</p>
            <h1>历史推演</h1>
            <small>查看与管理你过往的商业沙盘推演记录</small>
          </div>
        </header>
        <div className="history-filter-row">
          <input aria-label="搜索项目名称或关键词" placeholder="搜索项目名称 / 关键词" />
          <select aria-label="状态"><option>全部状态</option></select>
          <select aria-label="参与角色数"><option>全部角色</option></select>
          <select aria-label="创建时间"><option>选择时间范围</option></select>
          <button type="button">重置筛选</button>
        </div>
        <div className="history-table">
          <div className="history-table-head">
            <span>项目名称 / 描述</span>
            <span>参与角色</span>
            <span>推演时间</span>
            <span>状态</span>
            <span>综合评分</span>
            <span>风险项</span>
            <span>操作</span>
          </div>
          {visibleRows.map(({ title, detail, role, time, status, score, risk, href }) => (
            <article key={title}>
              <strong>{title}<small>{detail}</small></strong>
              <span>{role}</span>
              <time>{time}</time>
              <em>{status}</em>
              <b>★ {score}</b>
              <i>{risk}</i>
              <Link to={href}>查看报告</Link>
            </article>
          ))}
          {visibleRows.length === 0 ? <p className="sandbox-empty-state">暂无推演记录，完成首次配置后会显示在这里。</p> : null}
        </div>
      </section>
    </SandboxWorkLayout>
  );
}

function SandboxState({ title }: { title: string }) {
  return (
    <SandboxWorkLayout mode="report">
      <section className="sandbox-work-card sandbox-empty-state" role="status">
        <h1>{title}</h1>
        <Link to="/sandbox/history">返回历史推演</Link>
        <Link to="/sandbox/setup">发起新推演</Link>
      </section>
    </SandboxWorkLayout>
  );
}

function QuotaModal({ quota }: { quota: ReturnType<typeof quotaSummary> }) {
  const resetAt = quota.item?.reset_at
    ? new Date(quota.item.reset_at).toLocaleString("zh-CN", { hour12: false })
    : "以账户额度页为准";
  return (
    <div className="sandbox-modal-scrim">
      <section className="quota-modal" role="dialog" aria-label="本月沙盘次数已用尽">
        <button aria-label="关闭" type="button">×</button>
        <div className="quota-lock" />
        <h2>本月沙盘次数已用尽</h2>
        {quota.item ? <p>本周期已使用：{quota.item.used} / {quota.item.limit} {quota.item.unit}，重置时间：{resetAt}</p> : <p>额度信息加载中，请前往套餐页查看当前账户的实际额度。</p>}
        <footer>
          <Link to="/membership">查看套餐与额度</Link>
        </footer>
        <Link to="/sandbox">稍后再说</Link>
      </section>
    </div>
  );
}

function SandboxWorkLayout({ children, mode }: { children: ReactNode; mode: SandboxVariant }) {
  return (
    <div className={`sandbox-work-layout sandbox-work-${mode}`}>
      <main>
        <header className="sandbox-work-title">
          <Link aria-label="返回商业沙盘首页" to="/sandbox">←</Link>
          <strong>商业沙盘</strong>
          <span>多角色模拟未来，判断项目机会与风险</span>
        </header>
        {children}
      </main>
      <SandboxCopilot mode={mode} />
    </div>
  );
}

function SandboxStepper({ active }: { active: 1 | 2 | 3 }) {
  const steps = ["智能补充信息", "选择推演角色", "确认并开始推演"] as const;
  return (
    <nav className="sandbox-stepper" aria-label="商业沙盘步骤">
      {steps.map((step, index) => (
        <article className={index + 1 <= active ? "active" : ""} key={step}>
          <b>{index + 1 < active ? "✓" : index + 1}</b>
          <span>{step}</span>
        </article>
      ))}
    </nav>
  );
}

function SandboxCopilot({ mode }: { mode: SandboxVariant }) {
  const resultMode = mode === "run" || mode === "report" || mode === "history";
  const roleMode = mode === "roles";
  return (
    <aside className="sandbox-copilot" aria-label="智活 Copilot">
      <header>
        <span>✦</span>
        <div>
          <strong>智活 Copilot</strong>
          <small>你的全能 AI 助手，随时为你提供帮助</small>
        </div>
        <button type="button">⌃</button>
      </header>
      <div className="sandbox-chat mine">
        {mode === "home"
          ? "如何使用商业沙盘？"
          : roleMode
            ? "我想做一款面向上班族的低卡代餐奶昔"
            : "这些推演结果可以直接作为市场事实吗？"}
      </div>
      <div className="sandbox-chat">
        {roleMode
          ? "已理解你的初步想法。为了更精准地推演，请先选择一个或多个角色，AI 将从该角色的立场与你对话并给出建议。"
          : resultMode
          ? "不可以。沙盘结果是基于输入条件的 AI 情景推演，关键假设和结论仍需通过访谈、实验或可信数据验证。"
          : "先填写目标用户、产品方案和推演目标，再选择角色。系统会生成模型推演报告，并明确标注假设与证据边界。"}
      </div>
      {roleMode ? (
        <div className="sandbox-copilot-pager" aria-label="角色建议页码">
          <button type="button" aria-label="上一页">‹</button>
          <strong>1 / 5</strong>
          <button type="button" aria-label="下一页">›</button>
        </div>
      ) : <nav>
        {(resultMode
          ? [["开始新推演", "/sandbox/setup"], ["历史推演记录", "/sandbox/history"]]
          : [
              ["开始多角色推演", "/sandbox/setup"],
              ["查看推演思路", "/sandbox/setup"],
              ["生成推演大纲", "/sandbox/questions"],
              ["历史推演记录", "/sandbox/history"]
            ]
        ).map((item) => (
          <Link key={item[0]} to={item[1]}>{item[0]}</Link>
        ))}
      </nav>}
      <MiniCopilotForm className="sandbox-copilot-input" inputAriaLabel="向沙盘 Copilot 提问" attachIcon="＋" sendIcon="↗" />
    </aside>
  );
}

function SandboxOrbit() {
  return (
    <div className="sandbox-orbit" aria-hidden="true">
      {["用户视角", "竞争对手", "运营策略", "风险研判", "增长路径"].map((item) => <span key={item}>{item}</span>)}
      <strong />
      <i />
    </div>
  );
}

export default SandboxPage;
