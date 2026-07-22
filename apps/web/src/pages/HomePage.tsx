import { useCallback, useEffect, useState } from "react";
import { Bell, ChevronDown, ChevronUp, Settings, X } from "lucide-react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import FloatingCopilotOrb from "../components/FloatingCopilotOrb";
import ReferenceShell from "../components/ReferenceShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { authApi } from "../lib/authApi";
import { authSession, useAuthSession } from "../lib/authSession";
import { homeApi, type HomeSummary } from "../lib/homeApi";
import { notificationsApi } from "../lib/notificationsApi";

const heroCards = [
  {
    title: "项目确定及拆解",
    desc: "洞察机会，精准定位，科学拆解",
    href: "/projects",
    art: "cube"
  },
  {
    title: "落地",
    desc: "工具赋能，咨询陪跑，高效执行",
    href: "/tasks",
    art: "ring"
  },
  {
    title: "增长",
    desc: "获客转化，客户运营，持续增长",
    href: "/geo",
    art: "growth"
  }
];

const emptyWorkbenchActions = [
  {
    title: "浏览项目超市",
    desc: "先选择一个想验证的项目方向",
    href: "/projects"
  },
  {
    title: "打开商业沙盘",
    desc: "拆解商业模式、成本和落地路径",
    href: "/sandbox"
  },
  {
    title: "咨询 Copilot",
    desc: "让 AI 帮你生成第一版行动计划",
    href: "/copilot"
  }
];

const guestRecommendations = [
  {
    title: "项目推荐",
    desc: "发现优质项目，把握市场机会",
    meta: "智能匹配",
    cta: "精准推荐",
    href: "/projects",
    priority: "",
    accent: "violet",
    art: "board"
  },
  {
    title: "工具推荐",
    desc: "精选高效工具，提升落地效率",
    meta: "热门工具",
    cta: "场景适配",
    href: "/tools",
    priority: "",
    accent: "blue",
    art: "blocks"
  },
  {
    title: "资讯推荐",
    desc: "洞察行业动态，掌握最新趋势",
    meta: "精选资讯",
    cta: "每日更新",
    href: "/insights",
    priority: "",
    accent: "cyan",
    art: "news"
  }
];

const guestTasks = [
  { title: "完成【AI 智能硬件】项目商业画布", project: "", status: "", priority: "项目拆解", priorityClass: "project", time: "今天 10:00", overdue: false },
  { title: "与咨询顾问沟通落地方案细节", project: "", status: "", priority: "咨询通", priorityClass: "consulting", time: "今天 14:30", overdue: false },
  { title: "处理新客户需求：智能客服系统", project: "", status: "", priority: "CRM 客户管理", priorityClass: "crm", time: "明天 09:30", overdue: false },
  { title: "查看 AI 学习计划：数据分析工程师", project: "", status: "", priority: "AI 教学", priorityClass: "learning", time: "06-25 11:00", overdue: false },
  { title: "分析投放数据并优化获客策略", project: "", status: "", priority: "AI 线索开发", priorityClass: "leads", time: "06-25 15:00", overdue: false }
];

const assistantReplies = [
  {
    from: "assistant",
    text: "嗨，张婧！今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。"
  },
  {
    from: "user",
    text: "帮我分析一下智能客服系统的市场机会和落地关键点。"
  },
  {
    from: "assistant",
    text: "好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点，点击下方查看详情。",
    file: "智能客服系统机会分析报告 PDF · 1.2 MB"
  }
];

const accountLinks: Array<[string, string, string]> = [
  ["个人中心", "/profile", "user"],
  ["账号与资料设置", "/profile/settings", "settings"],
  ["会员与账单", "/membership", "wallet"],
  ["我的内容", "/profile/content", "content"],
  ["偏好设置", "/profile/preferences", "gear"]
];

const models: Array<[string, string, boolean]> = [
  ["Claude opus4.8", "claude", true],
  ["Chatgpt 5.5", "chatgpt", false],
  ["Grok4.3", "grok", false]
];

type HomePageProps = {
  assistantState?: "settings" | "files" | "collapsed";
  menuState?: "notice" | "account";
};

function HomePage({ assistantState, menuState }: HomePageProps) {
  const session = useAuthSession();
  const [summary, setSummary] = useState<HomeSummary | null>(null);
  const [accountOpen, setAccountOpen] = useState(menuState === "account");
  const [noticeOpen, setNoticeOpen] = useState(menuState === "notice");
  const [assistantOpen, setAssistantOpen] = useState(Boolean(session.user) && assistantState !== "collapsed");
  const [assistantMode, setAssistantMode] = useState<"chat" | "settings">(assistantState === "settings" ? "settings" : "chat");
  const [filesOpen, setFilesOpen] = useState(assistantState === "files");
  const [noticeBusy, setNoticeBusy] = useState(false);
  const [noticeError, setNoticeError] = useState("");
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [summaryError, setSummaryError] = useState("");
  const signedIn = Boolean(session.user);
  const nickname = session.user?.nickname || "张婧";
  const visibleHeroCards = summary?.hero_cards.length ? summary.hero_cards.map((card, index) => ({
    title: card.title,
    desc: card.summary ?? heroCards[index % heroCards.length].desc,
    href: card.url || heroCards[index % heroCards.length].href,
    art: heroCards[index % heroCards.length].art
  })) : heroCards;
  const visibleActions = summary ? (
    summary.action_items?.length ? summary.action_items.map((item, index) => ({
      title: item.title,
      desc: item.summary ?? "",
      meta: actionTypeLabel(item.type),
      href: item.url || "/projects",
      cta: item.cta || "去处理",
      priority: item.priority,
      accent: actionAccent(item.type, index),
      art: ["board", "blocks", "news"][index % 3]
    })) : summary.recommendations.map((card, index) => ({
      title: card.title,
      desc: card.summary ?? "",
      meta: "为你推荐",
      href: card.url || "/projects",
      cta: "查看",
      priority: "medium",
      accent: ["violet", "blue", "cyan"][index % 3],
      art: ["board", "blocks", "news"][index % 3]
    }))
  ) : signedIn ? [] : guestRecommendations;
  const visibleTasks = summary ? summary.recent_tasks.map((task) => ({
    title: task.title,
    project: task.project,
    status: statusLabel(task.status),
    priority: priorityText(task.priority),
    priorityClass: task.priority || "medium",
    time: task.due_at ? dueLabel(task.due_at, task.is_overdue) : "暂无截止",
    overdue: task.is_overdue
  })) : signedIn ? [] : guestTasks;
  const visibleMetrics = summary?.metrics.length ? summary.metrics.slice(0, 3) : signedIn ? [
    { label: "进行中任务", value: String(visibleTasks.length), icon: "folder" },
    { label: "待办任务", value: String(visibleTasks.length), icon: "inbox" },
    { label: "今日跟进", value: "-", icon: "trend" }
  ] : [
    { label: "进行中项目", value: "-", icon: "folder" },
    { label: "待办事项", value: "-", icon: "inbox" },
    { label: "本周新增线索", value: "-", icon: "trend" }
  ];
  const inProgressTaskCount = signedIn
    ? visibleTasks.filter((task) => task.status === "进行中").length
    : 3;
  const visibleNotifications = summary ? summary.notification_summary.latest.map((item) => ({
    title: item.title,
    desc: item.summary ?? "",
    time: formatHomeTime(item.created_at),
    icon: notificationIcon(item.type),
    unread: !item.read_at,
    href: item.action_url || "/messages",
    typeLabel: notificationTypeLabel(item.type),
    actionLabel: item.action_label || "查看详情"
  })) : [];
  const notificationTabs = summary ? [
    { label: "全部", count: summary.notification_summary.latest.length },
    ...((summary.notification_summary.by_type ?? []).slice(0, 3).map((item) => ({
      label: notificationTypeLabel(item.type),
      count: item.count
    })))
  ] : [];
  const accountSummary = summary?.account_summary;

  const loadHomeSummary = useCallback(async () => {
    if (!signedIn) return;
    setSummaryLoading(true);
    setSummaryError("");
    try {
      const payload = await homeApi.summary();
      setSummary(payload);
      setNoticeError("");
    } catch (error) {
      setSummary(null);
      setSummaryError(apiErrorMessage(error, "暂时无法同步工作台数据"));
    } finally {
      setSummaryLoading(false);
    }
  }, [signedIn]);

  useEffect(() => {
    if (!signedIn) {
      setSummary(null);
      setSummaryError("");
      setSummaryLoading(false);
      return;
    }
    void loadHomeSummary();
  }, [signedIn, loadHomeSummary]);

  useEffect(() => {
    if (!assistantState) {
      setAssistantOpen(Boolean(session.user));
      return;
    }
    setAssistantOpen(assistantState !== "collapsed" && Boolean(session.user));
    setAssistantMode(assistantState === "settings" ? "settings" : "chat");
    setFilesOpen(assistantState === "files");
  }, [assistantState, session.user]);

  useEffect(() => {
    setNoticeOpen(menuState === "notice");
    setAccountOpen(menuState === "account");
  }, [menuState]);

  async function logout() {
    try {
      await authApi.logout();
    } catch {
      // Local logout must still work when the API is temporarily unavailable.
    } finally {
      authSession.clear();
      setAccountOpen(false);
      setNoticeOpen(false);
      setAssistantOpen(false);
    }
  }

  async function markAllNotificationsRead() {
    if (!summary?.notification_summary.unread || noticeBusy) return;
    setNoticeBusy(true);
    setNoticeError("");
    try {
      await notificationsApi.markAllRead();
      const payload = await homeApi.summary();
      setSummary(payload);
    } catch (error) {
      setNoticeError(apiErrorMessage(error, "暂时无法标记全部已读"));
    } finally {
      setNoticeBusy(false);
    }
  }

  return (
    <ReferenceShell
      accountSlot={signedIn ? (
            <div className="ref-home-account-area">
              <div className="notice-control">
                <button
                  aria-expanded={noticeOpen}
                  aria-label="通知"
                  className="ref-home-notice"
                  onClick={() => {
                    setNoticeOpen((open) => !open);
                    setAccountOpen(false);
                    setAssistantOpen(false);
                  }}
                  type="button"
                >
                  <Bell aria-hidden="true" size={22} strokeWidth={1.8} />
                  <span />
                </button>
                {noticeOpen && (
                  <div className="notice-menu" role="dialog" aria-label="通知下拉框">
                    <div className="notice-head">
                      <h2>通知</h2>
                      <div>
                        <button disabled={!summary?.notification_summary.unread || noticeBusy} onClick={markAllNotificationsRead} type="button">
                          {noticeBusy ? "处理中..." : "全部已读"}
                        </button>
                        <Link to="/messages">查看消息中心</Link>
                        <button aria-label="关闭通知" onClick={() => setNoticeOpen(false)} type="button">×</button>
                      </div>
                    </div>
                    {noticeError && <p className="notice-error" role="alert">{noticeError}</p>}
                    <div className="notice-tabs" role="tablist" aria-label="通知分类">
                      {(notificationTabs.length ? notificationTabs : [{ label: "全部", count: 0 }]).map((tab, index) => (
                        <button className={index === 0 ? "active" : ""} key={tab.label} role="tab" aria-selected={index === 0} type="button">
                          {tab.label}<span>{tab.count}</span>
                        </button>
                      ))}
                    </div>
                    <div className="notice-list">
                      {visibleNotifications.length === 0 ? (
                        <p className="module-empty-state">暂无通知</p>
                      ) : visibleNotifications.map((notice) => (
                          <Link key={`${notice.title}-${notice.time}`} className={`notice-item ${notice.unread ? "unread" : "muted"}`} to={notice.href}>
                            <span className={`notice-icon ${notice.icon}`} aria-hidden="true" />
                            <span>
                              <strong>{notice.title}</strong>
                              <small>{notice.desc}</small>
                              <em>{notice.typeLabel} · {notice.actionLabel}</em>
                            </span>
                            <time>{notice.time}</time>
                          </Link>
                        ))}
                    </div>
                    <Link className="notice-more" to="/messages">查看更多</Link>
                  </div>
                )}
              </div>
              <div className="ref-home-account-control">
                <button
                  aria-expanded={accountOpen}
                  aria-label={`${nickname}的账号菜单`}
                  className="ref-home-account-trigger"
                  onClick={() => {
                    setAccountOpen((open) => !open);
                    setNoticeOpen(false);
                    setAssistantOpen(false);
                  }}
                  type="button"
                >
                  <img alt="" src="/public-components/avatar.jpg" />
                  <span>
                    <strong>{nickname} · 智活AI</strong>
                    <small>企业管理员</small>
                  </span>
                  <ChevronDown aria-hidden="true" size={15} strokeWidth={1.8} />
                </button>
                {accountOpen && (
                  <div className="v4-account-menu ref-home-account-menu" role="dialog" aria-label="头像下拉框">
                    <div className="account-card-head">
                      <img className="v4-avatar large" alt="" src="/public-components/avatar.jpg" />
                      <div>
                        <strong>{nickname}</strong>
                        <p>
                          <small>企业管理员</small>
                          <small className="gold">企业版</small>
                        </p>
                      </div>
                    </div>
                    <div className="account-plan">
                      <span>{accountSummary?.plan_name ?? "暂无会员信息"}</span>
                      <strong>{accountSummary ? `${accountSummary.credit_balance} 积分` : "0 积分"}</strong>
                    </div>
                    {accountSummary?.quota_warnings.map((warning) => (
                      <p className="form-error" key={warning.key}>{warning.message}</p>
                    ))}
                    <div className="account-menu-list">
                      {accountLinks.map(([label, href, icon]) => (
                        <Link key={label} to={href}>
                          <span className={`account-menu-icon ${icon}`} aria-hidden="true" />
                          {label}
                          <b aria-hidden="true">›</b>
                        </Link>
                      ))}
                    </div>
                    <button className="logout-button" onClick={logout} type="button">
                      <span className="account-menu-icon logout" aria-hidden="true" />
                      退出登录
                    </button>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div className="ref-home-guest-actions">
              <Link to="/login">登录 / 注册</Link>
              <Link to="/projects">立即体验</Link>
            </div>
          )}
      className="ref-home-shell"
      mainClassName={`ref-home-layout ${signedIn ? "signed-in" : "signed-out"} ${assistantOpen ? "with-assistant" : "copilot-collapsed"}`}
    >
          <section className="ref-home-dashboard" aria-label="智活AI 工作台">
            <div className="ref-home-welcome">
              <div>
                <h1>
                  {signedIn ? `上午好，${nickname}` : "欢迎来到 智活AI"}
                  {signedIn && <span className="welcome-wave" aria-hidden="true">👋</span>}
                </h1>
                <p>{signedIn ? "专注创造价值的每一步" : "一站式智能项目到增长工作台"}</p>
                {!signedIn && (
                  <small>从项目确定到规模增长，智活AI 助你高效决策、快速落地、持续增长。</small>
                )}
                {signedIn && accountSummary && (
                  <div className="ref-home-account-summary" aria-label="账户权益概览">
                    <span>{accountSummary.plan_name}</span>
                    <span>{accountSummary.credit_balance} 积分</span>
                    {accountSummary.quota_warnings.map((warning) => (
                      <span key={warning.key}>{warning.message}</span>
                    ))}
                  </div>
                )}
                {signedIn && summaryLoading && <p className="ref-home-summary-status" role="status">正在同步工作台数据...</p>}
                {signedIn && summaryError && (
                  <div className="ref-home-summary-error" role="alert">
                    <span>{summaryError}</span>
                    <button onClick={loadHomeSummary} type="button">重新加载</button>
                  </div>
                )}
              </div>
              <div className="ref-home-stats" aria-label="工作台统计">
                {visibleMetrics.map((metric) => (
                  <MetricCard icon={metric.icon || "folder"} key={metric.label} label={metric.label} value={metric.value} />
                ))}
              </div>
            </div>

            <div className="ref-home-domains">
              {visibleHeroCards.map((card) => (
                <Link key={card.title} className="ref-home-domain" to={card.href}>
                  <div>
                    <h2>{card.title}</h2>
                    <p>{card.desc}</p>
                  </div>
                  <span className="ref-home-domain-arrow" aria-hidden="true">→</span>
                  <span className={`ref-home-domain-art ${card.art}`} aria-hidden="true" />
                </Link>
              ))}
            </div>

            <section className="ref-home-recommend" aria-label="为你推荐">
              <div className="ref-home-section-head">
                <h2>为你推荐</h2>
                <Link to="/tasks">查看全部 <span aria-hidden="true">›</span></Link>
              </div>
              <div className="ref-home-recommend-grid">
                {visibleActions.length === 0 ? (
                  <div className="recommend-empty-state" role="status">
                    <span className="recommend-empty-icon" aria-hidden="true" />
                    <div>
                      <h3>暂无待处理行动</h3>
                      <p>当你浏览项目、使用工具或创建任务后，这里会展示接口返回的推荐内容。</p>
                    </div>
                    <Link to="/projects">先去项目超市 <span aria-hidden="true">›</span></Link>
                  </div>
                ) : visibleActions.map((card) => (
                    <Link key={card.title} className="ref-home-recommend-card" to={card.href}>
                      <span className={`recommend-icon ${card.accent}`} aria-hidden="true" />
                      <div>
                        {card.priority && <span className={`action-priority ${card.priority}`}>{priorityLabel(card.priority)}</span>}
                        <h3>{card.title}</h3>
                        <p>{card.desc}</p>
                        <small>{card.meta} · {card.cta}</small>
                      </div>
                      <span className={`mini-art ${card.art}`} aria-hidden="true" />
                      <span className="tiny-arrow" aria-hidden="true">→</span>
                    </Link>
                  ))}
              </div>
            </section>

            <section className="ref-home-tasks" aria-label="我的待办和进行中">
              <div className="ref-home-section-head">
                <div className="task-tabs">
                  <h2>我的待办 / 进行中</h2>
                  <button className="active" type="button">待办 {visibleTasks.length}</button>
                  <button type="button">进行中 {inProgressTaskCount}</button>
                </div>
                <Link to="/tasks">查看全部 <span aria-hidden="true">›</span></Link>
              </div>
              <div className="task-list">
                {visibleTasks.length === 0 ? (
                  <div className="empty-workbench-actions" role="status">
                    <div className="empty-workbench-copy">
                      <h3>暂无待办任务</h3>
                      <p>完成项目浏览、沙盘拆解或 Copilot 咨询后，接口返回的待办会出现在这里。</p>
                    </div>
                    <div className="empty-action-grid">
                      {emptyWorkbenchActions.map((action) => (
                        <Link key={action.href} className="empty-action-card" to={action.href}>
                          <span>{action.title}</span>
                          <small>{action.desc}</small>
                          <b aria-hidden="true">›</b>
                        </Link>
                      ))}
                    </div>
                  </div>
                ) : visibleTasks.map((task) => (
                    <Link key={task.title} className={`task-row ${task.overdue ? "overdue" : ""}`} to="/tasks">
                      <span className="task-check" aria-hidden="true" />
                      <span className="task-title">
                        <strong>{task.title}</strong>
                        {task.project && <small>{task.project} · {task.status}</small>}
                      </span>
                      <span className={`task-tag priority-${task.priorityClass}`}>{task.priority}</span>
                      <time>{task.time}</time>
                    </Link>
                  ))}
              </div>
            </section>
          </section>

          <aside className={`ref-home-copilot ${assistantOpen ? "open" : "closed"}`} aria-label="智活 Copilot">
            {assistantOpen ? (
              <>
                <div className="ref-home-copilot-head">
                  <div>
                    <span className="spark" aria-hidden="true">✦</span>
                    <strong>智活 <b>Copilot</b></strong>
                    <p>你的全球 AI 助手，随时为你提供帮助</p>
                  </div>
                  <div className="ref-home-copilot-head-actions">
                    <button
                      aria-label="打开 Copilot 设置"
                      className={assistantMode === "settings" ? "active" : ""}
                      onClick={() => setAssistantMode((mode) => (mode === "settings" ? "chat" : "settings"))}
                      type="button"
                    >
                      <Settings aria-hidden="true" size={17} strokeWidth={1.8} />
                    </button>
                    <button
                      aria-label="收起智活 Copilot"
                      onClick={() => {
                        setAssistantOpen(false);
                      }}
                      type="button"
                    >
                      <ChevronDown aria-hidden="true" size={18} strokeWidth={1.8} />
                    </button>
                  </div>
                </div>
                {assistantMode === "settings" && (
                  <div className="ref-home-copilot-settings">
                    <div className="ref-home-settings-title">
                      <strong>Copilot 设置</strong>
                      <button aria-label="关闭 Copilot 设置" onClick={() => setAssistantMode("chat")} type="button">
                        <X aria-hidden="true" size={17} strokeWidth={1.8} />
                      </button>
                    </div>
                    <div className="ref-home-model-list" aria-label="模型选择">
                      <span>模型选择</span>
                      {models.map(([model, icon, selected]) => (
                        <button className={selected ? "selected" : ""} key={model} type="button">
                          <i className={`model-icon ${icon}`} aria-hidden="true" />
                          {model}
                        </button>
                      ))}
                    </div>
                    <div className="ref-home-deep-thinking">
                      <div>
                        <strong>深度思考 <small>VIP</small></strong>
                        <span>更深入分析，回复更完整</span>
                      </div>
                      <button aria-label="深度思考开关" className="toggle-on" type="button" />
                    </div>
                  </div>
                )}
                <div className="ref-home-copilot-thread">
                  {assistantReplies.map((message, index) => (
                    <article key={`${message.from}-${index}`} className={message.from}>
                      <span className="ai-avatar">A</span>
                      <div>
                        <p>{message.text}</p>
                        {message.file && <span className="file-chip">{message.file}</span>}
                      </div>
                    </article>
                  ))}
                </div>
                <div className="ref-home-copilot-actions">
                  <Link to="/projects">分析项目机会 <span aria-hidden="true">›</span></Link>
                  <Link to="/tools">推荐工具 <span aria-hidden="true">›</span></Link>
                  <Link to="/tasks">制定落地计划 <span aria-hidden="true">›</span></Link>
                </div>
                <div className="ref-home-copilot-input-stack">
                  {filesOpen && (
                    <div className="attachment-strip">
                      <article>
                        <span className="attachment-thumb" aria-hidden="true" />
                        <button aria-label="移除图片附件" onClick={() => setFilesOpen(false)} type="button">×</button>
                      </article>
                      <article>
                        <span className="pdf-thumb" aria-hidden="true">PDF</span>
                        <span>
                          <strong>市场分析报告.pdf</strong>
                          <small>1.2 MB</small>
                        </span>
                        <button aria-label="移除 PDF 附件" onClick={() => setFilesOpen(false)} type="button">×</button>
                      </article>
                    </div>
                  )}
                  <MiniCopilotForm
                    attachLabel="添加文件"
                    className="ref-home-copilot-input"
                    inputAriaLabel="询问智活 Copilot"
                    onAttach={() => setFilesOpen((open) => !open)}
                    sendIcon="↗"
                  />
                </div>
              </>
            ) : (
              <button
                className="ref-home-copilot-mini"
                onClick={() => {
                  setAssistantOpen(true);
                  setAssistantMode("chat");
                }}
                type="button"
              >
                <span className="mini-logo" aria-hidden="true" />
                <strong>智活 Copilot</strong>
                <ChevronUp className="mini-caret" aria-hidden="true" size={18} strokeWidth={1.8} />
                <span className="ref-home-mini-input" aria-hidden="true">
                  <i>+</i>
                  <small>输入问题，发送后自动展开回复</small>
                  <b>↗</b>
                </span>
              </button>
            )}
          </aside>

          <FloatingCopilotOrb
            className={filesOpen || assistantMode === "settings" ? "home-extended-orb" : ""}
            onActivate={() => setAssistantOpen(true)}
          />

    </ReferenceShell>
  );
}

function MetricCard({ label, value, icon }: { label: string; value: string; icon: string }) {
  return (
    <article className="ref-home-metric">
      <span>
        <small>{label}</small>
        <strong>{value}</strong>
      </span>
      <i className={`metric-icon ${icon}`} aria-hidden="true" />
    </article>
  );
}

function actionTypeLabel(type: string) {
  switch (type) {
    case "membership":
      return "会员权益";
    case "task":
      return "任务中心";
    case "leads":
      return "AI线索";
    case "sandbox":
      return "商业沙盘";
    case "competitor":
      return "竞品数据";
    case "crm":
      return "CRM客户";
    case "project":
      return "项目超市";
    default:
      return "工作台";
  }
}

function actionAccent(type: string, index: number) {
  switch (type) {
    case "task":
    case "crm":
    case "membership":
      return "blue";
    case "leads":
    case "competitor":
      return "cyan";
    default:
      return ["violet", "blue", "cyan"][index % 3];
  }
}

function priorityLabel(priority: string) {
  switch (priority) {
    case "high":
      return "优先处理";
    case "low":
      return "等待中";
    default:
      return "建议处理";
  }
}

function statusLabel(status: string) {
  switch (status) {
    case "todo":
      return "待办";
    case "in_progress":
      return "进行中";
    case "completed":
      return "已完成";
    case "reminder":
      return "提醒";
    default:
      return status || "未分类";
  }
}

function priorityText(priority?: string) {
  switch (priority) {
    case "high":
      return "高优先级";
    case "low":
      return "低优先级";
    default:
      return "中优先级";
  }
}

function dueLabel(value: string, overdue: boolean) {
  const formatted = formatHomeTime(value);
  return overdue ? `已逾期 ${formatted}` : `截止 ${formatted}`;
}

function notificationTypeLabel(type: string) {
  switch (type) {
    case "task":
      return "任务";
    case "analysis":
      return "分析";
    case "lead":
      return "线索";
    case "crm":
      return "CRM";
    case "membership":
      return "会员";
    case "system":
      return "系统";
    default:
      return "通知";
  }
}

function notificationIcon(type: string) {
  switch (type) {
    case "task":
      return "check";
    case "lead":
    case "crm":
      return "target";
    case "analysis":
      return "calc";
    case "membership":
      return "gift";
    default:
      return "system";
  }
}

function formatHomeTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

export default HomePage;
