import { useEffect, useState, type MouseEvent, type ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { sandboxApi, type SandboxSession } from "../lib/sandboxApi";

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

const setupFacts = [
  ["项目设想", "做一款面向上班族的 AI 低卡代餐奶昔，通过 AI 用户规划每天的营养和饮食计划。"],
  ["所属行业", "人工智能 / 健康轻食"],
  ["目标用户", "上班族"],
  ["定价区间", "30元 / 杯"],
  ["销售场景", "写字楼附近线下门店"]
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

const conversationRows = [
  ["14:32", "用户视角", "中小企业普遍面临知识分散、检索效率低、沉淀难的问题，经理/员工对快速找资料、复用知识、降低重复劳动需求强。"],
  ["14:33", "投资人视角", "目标市场规模可观，国内中小企业超 4000 万家，SaaS 渗透率持续提升。ROI 关注点：客户留存与扩张、单位经济模型、产品壁垒。"],
  ["14:34", "竞争对手", "替代方案包括通用网盘、搜索、企业微信/钉钉文档、Notion 等。差异点在于行业知识图谱构建、AI 检索准确率。"],
  ["14:35", "运营策略", "获客：内容营销 + 渠道伙伴 + 行业社群；留存：知识沉淀可视化、使用激励、角色化权限提升粘性。"],
  ["14:35", "增长路径", "阶段 1 聚焦 10 个细分行业，阶段 2 渠道赋能，阶段 3 向中大型客户延伸。"],
  ["14:37", "风险研判", "合规风险、模型幻觉、客户迁移成本和续费口碑是主要压力点。"],
  ["14:38", "智活 Copilot 总结", "本轮对话建议先聚焦中等及以上风险项目，完成用户验证与 MVP 试点后再扩张。"]
] as const;

const reportMetrics = [
  ["综合可行性", "83", "/100", "可行性较高，建议推进验证"],
  ["消费概率", "68", "%", "市场接受度中高，具备增长潜力"],
  ["风险等级", "中等", "", "存在可控风险，需重点关注3项"],
  ["推荐优先级", "A", "级", "建议优先投入验证资源"]
] as const;

const roleSummaries = [
  ["用户视角", "操作简单，快速检索与沉淀知识问题直接，价值感明显。"],
  ["投资人视角", "赛道空间大，建议关注单位经济模型与商业化闭环。"],
  ["竞争对手", "同类产品通用性强，行业化与落地服务是差异点。"],
  ["运营策略", "建议先聚焦细分行业与核心场景，建立标杆案例。"],
  ["增长路径", "出口型与渠道驱动为主，逐步拓展生态合作。"],
  ["风险研判", "重点关注数据合规、客户教育成本与模型成本。"]
] as const;

const nextActions = [
  ["用户验证", "录入访谈10+目标客户，验证核心痛点与付费意愿。"],
  ["MVP测试", "开发核心功能 MVP，进行小范围可用性测试。"],
  ["定价试验", "设计2-3套定价方案，开展小规模价格测试。"],
  ["渠道试点", "对接3-5个渠道伙伴，试点推广与合作模式。"],
  ["合规检查", "梳理数据安全与合规要求，完成必要认证准备。"]
] as const;

const historyRows = [
  ["AI智能客服SaaS平台", "面向中小企业的智能客服解决方案", "用户 投资人 运营 +2", "2024-05-20 14:32", "已完成", "8.6", "中等"],
  ["跨境电商供应链协同平台", "一站式跨境供应链协同与管理平台", "用户 运营 供应链 +1", "2024-05-18 09:16", "已完成", "7.9", "较高"],
  ["AI个性化学习助手", "基于AI的个性化学习与辅导工具", "用户 教育专家 投资人 +1", "2024-05-16 16:45", "已完成", "8.1", "中等"],
  ["社区团购O2O平台", "本地社区团购与即时配送服务", "用户 运营 投资人 +1", "2024-05-15 11:20", "进行中", "7.2", "较高"],
  ["健康管理小程序", "个人健康数据管理与健康建议服务", "用户 医生 运营 +1", "2024-05-14 10:08", "草稿", "6.4", "中等"],
  ["企业数据分析平台", "中小企业数据可视化与分析平台", "用户 数据专家 投资人 +1", "2024-05-12 15:33", "已完成", "8.3", "中等"],
  ["智能硬件IoT解决方案", "智能家居硬件及物联网平台方案", "用户 工程师 投资人 +2", "2024-05-10 09:50", "已完成", "7.6", "较高"]
] as const;

const selectedRoleIndexes = new Set([0, 1, 3, 4]);

function SandboxPage({ variant = "home" }: SandboxPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [sessions, setSessions] = useState<SandboxSession[]>([]);
  const [isStarting, setIsStarting] = useState(false);
  const [startError, setStartError] = useState("");

  useEffect(() => {
    let active = true;
    sandboxApi
      .listSessions(10)
      .then((payload) => {
        if (active) setSessions(payload.sessions);
      })
      .catch(() => {
        if (active) setSessions([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const routedSession = (location.state as { sandboxSession?: SandboxSession } | null)?.sandboxSession ?? null;
  const reportSession = routedSession ?? sessions.find((session) => session.status === "completed" && session.report) ?? sessions.find((session) => session.report) ?? null;

  async function startSandbox(event: MouseEvent<HTMLAnchorElement>) {
    event.preventDefault();
    if (isStarting) return;
    setIsStarting(true);
    setStartError("");
    try {
      const draft = await sandboxApi.createSession({
        goal: "验证 AI 低卡代餐奶昔",
        targetUsers: "上班族",
        product: "AI 低卡代餐奶昔",
        roles: [roles[0][0], roles[1][0], roles[3][0], roles[4][0]]
      });
      const completed = await sandboxApi.runSession(draft.id);
      setSessions((current) => [completed, ...current.filter((session) => session.id !== completed.id)]);
      navigate("/sandbox/report", { state: { sandboxSession: completed } });
    } catch (error) {
      setStartError(apiErrorMessage(error, "推演启动失败，请稍后重试。"));
    } finally {
      setIsStarting(false);
    }
  }

  return (
    <V4PageShell className="sandbox-shell" showCopilotMini={false}>
      <section className={`sandbox-v2 sandbox-${variant}`} aria-label="商业沙盘">
        {variant === "home" && <SandboxHome />}
        {variant === "setup" && <SetupPage />}
        {variant === "roles" && <RolesPage />}
        {variant === "start" && <StartPage isStarting={isStarting} onStart={startSandbox} startError={startError} />}
        {variant === "questions" && <QuestionsPage />}
        {variant === "run" && <RunPage />}
        {variant === "report" && <ReportPage session={reportSession} />}
        {variant === "history" && <HistoryPage sessions={sessions} />}
        {variant === "quota" && (
          <>
            <SandboxHome />
            <QuotaModal />
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
          <strong>你已输入的信息</strong>
          <p>我想做一款面向上班族的低卡代餐奶昔，通过 AI 帮用户规划每天的营养和饮食计划，按月订阅收费，大概定价在 30 元左右。</p>
          <button type="button">编辑</button>
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
        <Link className="sandbox-primary-wide" to="/sandbox/roles">确认以上信息，选择推演角色 →</Link>
      </section>
    </SandboxWorkLayout>
  );
}

function RolesPage() {
  return (
    <SandboxWorkLayout mode="roles">
      <SandboxStepper active={2} />
      <section className="sandbox-work-card roles-work">
        <h2>你希望从谁的视角进行推演？</h2>
        <p>选择一个或多个角色，AI 将基于该角色的立场对话并给出建议。</p>
        <div className="role-select-grid">
          {roles.map(([title, detail, badge, kind], index) => (
            <article className={selectedRoleIndexes.has(index) ? "selected" : ""} key={title}>
              <i className={`role-avatar ${kind}`} />
              <b>{selectedRoleIndexes.has(index) ? "✓" : ""}</b>
              <strong>{title}</strong>
              <small>{detail}</small>
              <span>{badge}</span>
            </article>
          ))}
        </div>
        <aside className="role-help-band">
          <strong>为什么选择多个角色？</strong>
          <p>多角色视角能帮助你更全面地发现机会、识别风险，获得更立体的推演结果。</p>
        </aside>
        <footer>
          <Link to="/sandbox/setup">上一步</Link>
          <Link to="/sandbox/start">确认角色，进入下一步 →</Link>
        </footer>
      </section>
    </SandboxWorkLayout>
  );
}

function StartPage({
  isStarting,
  onStart,
  startError
}: {
  isStarting: boolean;
  onStart: (event: MouseEvent<HTMLAnchorElement>) => void;
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
              <button type="button">编辑</button>
            </header>
            {setupFacts.map(([label, value]) => (
              <article key={label}>
                <b>{label}</b>
                <span>{value}</span>
              </article>
            ))}
            <Link to="/sandbox/setup">查看完整信息 ›</Link>
          </section>
          <section>
            <h2>参与推演角色 <small>已选 4 个</small></h2>
            <div className="start-role-grid">
              {[roles[0], roles[1], roles[3], roles[4]].map(([title, detail, , kind]) => (
                <article key={title}>
                  <i className={`role-avatar ${kind}`} />
                  <b>✓</b>
                  <strong>{title}</strong>
                  <small>{detail}</small>
                </article>
              ))}
            </div>
          </section>
        </div>
        <footer>
          <div>
            <strong>推演深度</strong>
            <span>标准（推荐）</span>
          </div>
          <div>
            <strong>输出风格</strong>
            <span>结构化报告（推荐）</span>
          </div>
          <label>
            <input defaultChecked type="checkbox" />
            生成推演大纲
          </label>
          <Link aria-disabled={isStarting} onClick={onStart} to="/sandbox/run">{isStarting ? "推演启动中..." : "开始推演 🚀"}</Link>
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

function RunPage() {
  return (
    <SandboxWorkLayout mode="run">
      <section className="run-work">
        <header>
          <div>
            <h1>AI 驱动中小企业知识管理平台</h1>
            <span>推演中</span>
            <span>标准深度</span>
          </div>
          <small>本轮推演 · 第 1 轮</small>
        </header>
        <nav className="run-role-tabs">
          {["用户视角", "投资人视角", "代理商/渠道方", "竞争对手", "运营策略", "增长路径", "风险研判"].map((item) => (
            <button key={item} type="button">{item}</button>
          ))}
        </nav>
        <div className="run-grid">
          <section className="run-dialog">
            {conversationRows.map(([time, role, text]) => (
              <article key={`${time}-${role}`}>
                <time>{time}</time>
                <strong>{role}</strong>
                <p>{text}</p>
              </article>
            ))}
            <footer>
              <Link to="/sandbox/run">继续推演</Link>
              <button type="button">调整变量后重跑</button>
              <Link to="/sandbox/report">生成推演报告</Link>
            </footer>
          </section>
          <aside className="run-variable-panel">
            <h2>调整变量</h2>
            {["定价策略", "获客渠道", "服务模式", "目标客群"].map((item) => (
              <label key={item}>
                <span>{item}</span>
                <select aria-label={item}>
                  <option>{item === "定价策略" ? "中档订阅（39 元/人/月）" : item === "获客渠道" ? "内容营销 + 渠道伙伴" : item === "服务模式" ? "SaaS 标准版" : "10-200 人规模的中小企业"}</option>
                </select>
              </label>
            ))}
            <h2>追加追问</h2>
            <textarea aria-label="追加追问" placeholder="输入你的问题，进一步追问任意角色..." />
            <button type="button">发送追问</button>
          </aside>
        </div>
      </section>
    </SandboxWorkLayout>
  );
}

function formatSessionTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { hour12: false });
}

function reportMetricProgress(value: string, index: number) {
  const numeric = Number.parseFloat(value);
  if (Number.isFinite(numeric)) return Math.max(0, Math.min(100, numeric));
  return index === 2 ? 55 : 92;
}

function ReportPage({ session }: { session: SandboxSession | null }) {
  const report = session?.report;
  const visibleTitle = session?.product || "AI 驱动中小企业知识管理平台";
  const visibleTime = session ? formatSessionTime(session.updated_at) : "2025-05-20 14:32";
  const visibleRolesCount = session?.roles.length ?? 6;
  const visibleMetrics = report?.metrics.length
    ? report.metrics.map((metric, index) => [
      metric.label,
      metric.value,
      index === 0 ? "/100" : "",
      "来自本次沙盘推演报告"
    ] as const)
    : reportMetrics;
  const visibleConclusions = report ? [report.summary, ...report.risks.map((risk) => `风险提示：${risk}`)] : [
    "市场需求明确且增长潜力大，中小企业知识管理数字化痛点显著。",
    "AI 驱动方案可有效提升效率并降低成本。",
    "风险主要集中在数据安全合规、客户教育成本与付费转化路径。"
  ];
  const visibleRoleSummaries = report?.role_summaries.length
    ? report.role_summaries.map((item) => [item.role, item.view] as const)
    : roleSummaries;
  const visibleNextActions = report?.next_actions.length
    ? report.next_actions.map((item, index) => [`行动 ${index + 1}`, item] as const)
    : nextActions;

  return (
    <SandboxWorkLayout mode="report">
      <section className="report-work">
        <header>
          <div>
            <p>商业沙盘 / 历史推演</p>
            <h1>{visibleTitle}</h1>
            <small>推演时间：{visibleTime}　参与角色数：{visibleRolesCount}　报告版本：V1.0</small>
          </div>
          <button type="button">导出报告</button>
        </header>
        <div className="report-metrics">
          {visibleMetrics.map(([label, value, suffix, detail], index) => (
            <article key={label}>
              <small>{label}</small>
              <strong>{value}<span>{suffix}</span></strong>
              <p>{detail}</p>
              <i style={{ width: `${reportMetricProgress(value, index)}%` }} />
            </article>
          ))}
        </div>
        <section className="report-summary">
          <h2>核心结论</h2>
          <ul>
            {visibleConclusions.map((item) => <li key={item}>{item}</li>)}
          </ul>
        </section>
        <div className="report-two-col">
          <article>
            <h2>机会分析</h2>
            <p>中小企业数字化渗透率持续提升，知识管理场景刚需明显，现有竞品价格偏高且通用性强，轻量化 SaaS 存在市场空白。</p>
          </article>
          <article>
            <h2>风险分析</h2>
            <p>涉及企业敏感数据，合规与安全要求高；中小企业付费意愿和迁移成本需要通过样板客户验证。</p>
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

function sessionRisk(session: SandboxSession) {
  if (!session.report) return "中等";
  if (session.report.score >= 85) return "较低";
  if (session.report.score >= 70) return "中等";
  return "较高";
}

function toHistoryRow(session: SandboxSession) {
  return [
    session.product || session.goal,
    session.goal,
    session.roles.join(" ") || "未选择角色",
    formatSessionTime(session.updated_at),
    session.status === "completed" ? "已完成" : "草稿",
    session.report ? (session.report.score / 10).toFixed(1) : "-",
    sessionRisk(session)
  ] as const;
}

function HistoryPage({ sessions }: { sessions: SandboxSession[] }) {
  const visibleRows = sessions.length ? sessions.map(toHistoryRow) : historyRows;

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
        <aside className="history-notice">
          <strong>历史记录管理说明</strong>
          <span>普通版仅保存最近 3 条历史，会员版可长期保存更多记录。</span>
          <button type="button">升级会员，解锁更多记录</button>
        </aside>
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
            <span>风险等级</span>
            <span>操作</span>
          </div>
          {visibleRows.map(([title, detail, role, time, status, score, risk]) => (
            <article key={title}>
              <strong>{title}<small>{detail}</small></strong>
              <span>{role}</span>
              <time>{time}</time>
              <em>{status}</em>
              <b>★ {score}</b>
              <i>{risk}</i>
              <Link to="/sandbox/report">查看报告</Link>
            </article>
          ))}
        </div>
      </section>
    </SandboxWorkLayout>
  );
}

function QuotaModal() {
  return (
    <div className="sandbox-modal-scrim">
      <section className="quota-modal" role="dialog" aria-label="本月沙盘次数已用尽">
        <button aria-label="关闭" type="button">×</button>
        <div className="quota-lock" />
        <h2>本月沙盘次数已用尽</h2>
        <p>您已用完普通版每月可用的沙盘推演次数。</p>
        <div className="quota-compare">
          <article><small>普通版</small><strong>1 次/月</strong></article>
          <b>VS</b>
          <article><small>会员版</small><strong>20 次/月</strong></article>
        </div>
        <p>本月已使用：1 / 1 次，重置时间：2025-06-01</p>
        <h3>升级会员版，立即享受更多权益</h3>
        <div className="quota-benefits">
          {["解锁更多推演次数", "保存更多历史记录", "导出完整推演报告", "优先体验高级分析能力"].map((item) => <span key={item}>{item}</span>)}
        </div>
        <footer>
          <Link to="/membership">升级套餐</Link>
          <button type="button">联系客服</button>
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
        {mode === "home" ? "我有一个智能宠物新品的想法，帮我从多角度分析市场并评估机会与潜在风险。" : "我想做一款面向上班族的低卡代餐奶昔，请帮我判断机会。"}
      </div>
      <div className="sandbox-chat">
        {resultMode
          ? "已为你梳理当前推演的关键建议，优先验证中等及以上风险项目。"
          : "好的，我将基于你提供的信息和选择的角色进行多维度推演分析。"}
      </div>
      <nav>
        {(resultMode
          ? ["查看完整建议动作", "生成PPT报告", "导出Excel数据", "分享报告链接"]
          : ["开始多角色推演", "查看推演思路", "生成推演大纲", "历史推演记录"]
        ).map((item) => (
          <Link key={item} to={item.includes("历史") ? "/sandbox/history" : item.includes("开始") ? "/sandbox/setup" : "#"}>{item}</Link>
        ))}
      </nav>
      <label>
        <span>＋</span>
        <input aria-label="向沙盘 Copilot 提问" placeholder="询问任何问题..." />
        <b>↗</b>
      </label>
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
