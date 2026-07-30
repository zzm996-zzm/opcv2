import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  CheckCircle2,
  ChevronRight,
  Clock3,
  FileText,
  MessageSquareReply,
  Plus,
  Search,
  Send,
  Ticket,
  X
} from "lucide-react";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type HelpTopic } from "../lib/contentApi";
import { supportApi, type SupportTicket } from "../lib/supportApi";

type TicketStatus = "pending" | "processing" | "replied" | "closed";
type TicketFilter = "all" | TicketStatus;

const referenceTopics: HelpTopic[] = [
  { key: "account", name: "账号与安全" },
  { key: "membership", name: "套餐与额度" },
  { key: "projects", name: "项目超市" },
  { key: "sandbox", name: "商业沙盘" },
  { key: "data", name: "数据破解" },
  { key: "growth", name: "增长测算" }
];

const referenceTickets: SupportTicket[] = [
  {
    id: 20418,
    topic: "技术支持",
    title: "AI线索任务结果无法导入 CRM",
    body: "线索任务已经完成，但点击加入 CRM 后没有生成客户记录。",
    status: "replied",
    created_at: "2026-07-26T15:20:00+08:00",
    updated_at: "2026-07-26T15:40:00+08:00"
  },
  {
    id: 20390,
    topic: "套餐与额度",
    title: "会员续费后额度未到账",
    body: "会员方案已续费，页面中的 AI 任务额度仍未更新。",
    status: "processing",
    created_at: "2026-07-24T09:10:00+08:00",
    updated_at: "2026-07-24T09:30:00+08:00"
  },
  {
    id: 20355,
    topic: "售前咨询",
    title: "咨询导出报告是否支持自定义模板",
    body: "希望确认企业版报告是否可以配置品牌、章节和导出模板。",
    status: "closed",
    created_at: "2026-07-19T10:22:00+08:00",
    updated_at: "2026-07-19T10:22:00+08:00"
  }
];

const statusFilters: ReadonlyArray<{ key: TicketFilter; label: string }> = [
  { key: "all", label: "全部" },
  { key: "processing", label: "处理中" },
  { key: "replied", label: "已回复" },
  { key: "pending", label: "待处理" },
  { key: "closed", label: "已关闭" }
];

const statusLabels: Record<TicketStatus, string> = {
  pending: "待处理",
  processing: "处理中",
  replied: "已回复",
  closed: "已关闭"
};

function normalizeTicketStatus(value: string): TicketStatus {
  const status = value.trim().toLowerCase();
  if (["processing", "in_progress", "处理中"].includes(status)) return "processing";
  if (["replied", "answered", "已回复"].includes(status)) return "replied";
  if (["closed", "resolved", "已关闭", "已解决"].includes(status)) return "closed";
  return "pending";
}

function ticketNumber(ticket: SupportTicket) {
  return `TK-${String(ticket.id).padStart(5, "0")}`;
}

function formatTicketTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function HelpPage() {
  const [topics, setTopics] = useState<HelpTopic[]>([]);
  const [tickets, setTickets] = useState<SupportTicket[]>([]);
  const [activeFilter, setActiveFilter] = useState<TicketFilter>("all");
  const [query, setQuery] = useState("");
  const [topic, setTopic] = useState("");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [notice, setNotice] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [loadError, setLoadError] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [composerOpen, setComposerOpen] = useState(false);
  const [selectedTicket, setSelectedTicket] = useState<SupportTicket | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all([contentApi.listHelpTopics(), supportApi.listTickets(20)])
      .then(([topicsPayload, ticketsPayload]) => {
        if (!active) return;
        const visibleTopics = topicsPayload.topics.length ? topicsPayload.topics : referenceTopics;
        setTopics(visibleTopics);
        setTickets(ticketsPayload.tickets.length ? ticketsPayload.tickets : referenceTickets);
        setTopic(visibleTopics[0]?.name || "");
        setLoadError("");
      })
      .catch(() => {
        if (!active) return;
        setTopics(referenceTopics);
        setTickets(referenceTickets);
        setTopic(referenceTopics[0].name);
        setLoadError("暂时无法读取最新工单，当前展示示例内容。稍后可刷新重试。");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const counts = useMemo(() => {
    const next = { total: tickets.length, pending: 0, processing: 0, replied: 0, closed: 0 };
    tickets.forEach((ticket) => {
      next[normalizeTicketStatus(ticket.status)] += 1;
    });
    return next;
  }, [tickets]);

  const visibleTickets = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase("zh-CN");
    return tickets.filter((ticket) => {
      if (activeFilter !== "all" && normalizeTicketStatus(ticket.status) !== activeFilter) return false;
      if (!normalizedQuery) return true;
      return [ticketNumber(ticket), String(ticket.id), ticket.title, ticket.topic ?? ""]
        .some((value) => value.toLocaleLowerCase("zh-CN").includes(normalizedQuery));
    });
  }, [activeFilter, query, tickets]);

  const stats = [
    { label: "全部工单", value: counts.total, tone: "all", icon: Ticket },
    { label: "处理中", value: counts.processing, tone: "processing", icon: Clock3 },
    { label: "已回复", value: counts.replied, tone: "replied", icon: MessageSquareReply },
    { label: "已关闭", value: counts.closed, tone: "closed", icon: CheckCircle2 }
  ] as const;

  function openComposer() {
    setNotice("");
    setSubmitError("");
    setComposerOpen(true);
  }

  async function createTicket(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!topic || !title.trim() || !body.trim() || submitting) return;
    setNotice("");
    setSubmitError("");
    setSubmitting(true);
    try {
      const ticket = await supportApi.createTicket({ topic, title: title.trim(), body: body.trim() });
      setTickets((current) => [ticket, ...current]);
      setTitle("");
      setBody("");
      setComposerOpen(false);
      setActiveFilter("all");
      setQuery("");
      setNotice(`工单 ${ticketNumber(ticket)} 已提交，我们会尽快处理。`);
    } catch (error) {
      setSubmitError(apiErrorMessage(error, "工单提交失败，请稍后重试"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <V4PageShell className="public-help-shell ticket-center-shell">
      <section className="support-ticket-page" aria-label="工单中心">
        <section className="support-overview">
          <header>
            <div>
              <h1>工单中心</h1>
              <p>遇到技术、额度或使用问题？提交工单，我们的团队会尽快为你处理。</p>
            </div>
            <button className="support-create-button" onClick={openComposer} type="button">
              <Plus size={19} aria-hidden="true" />
              提交工单
            </button>
          </header>

          <div className="support-stat-strip" aria-label="工单统计">
            {stats.map(({ icon: Icon, label, tone, value }) => (
              <article key={label}>
                <span className={`support-stat-icon ${tone}`} aria-hidden="true"><Icon size={21} /></span>
                <strong>{value}</strong>
                <small>{label}</small>
              </article>
            ))}
          </div>
        </section>

        {loadError && <p className="support-inline-message error" role="alert">{loadError}</p>}
        {notice && <p className="support-inline-message success" role="status">{notice}</p>}

        <div className="support-toolbar">
          <div className="support-tabs" role="tablist" aria-label="工单状态">
            {statusFilters.map((filter) => (
              <button
                aria-selected={activeFilter === filter.key}
                className={activeFilter === filter.key ? "active" : ""}
                key={filter.key}
                onClick={() => setActiveFilter(filter.key)}
                role="tab"
                type="button"
              >
                {filter.label}
              </button>
            ))}
          </div>
          <label className="support-search">
            <Search size={19} aria-hidden="true" />
            <input
              aria-label="搜索工单"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="搜索工单号或标题"
              value={query}
            />
          </label>
        </div>

        <div className="support-ticket-list" aria-live="polite">
          {loading && <div className="support-empty-state">正在读取工单...</div>}
          {!loading && visibleTickets.length === 0 && (
            <div className="support-empty-state">
              <Ticket size={28} aria-hidden="true" />
              <strong>没有符合条件的工单</strong>
              <span>可以调整筛选条件或提交一个新工单。</span>
            </div>
          )}
          {!loading && visibleTickets.map((ticket) => {
            const viewStatus = normalizeTicketStatus(ticket.status);
            return (
              <button className="support-ticket-row" key={ticket.id} onClick={() => setSelectedTicket(ticket)} type="button">
                <span className="support-ticket-copy">
                  <span className="support-ticket-tags">
                    <b className={`status-${viewStatus}`}><i aria-hidden="true" />{statusLabels[viewStatus]}</b>
                    {ticket.topic && <em>{ticket.topic}</em>}
                  </span>
                  <strong>{ticket.title}</strong>
                  <span className="support-ticket-meta">
                    <span>{ticketNumber(ticket)}</span>
                    {ticket.topic && <span>分类：{ticket.topic}</span>}
                    <span>最后更新：{formatTicketTime(ticket.updated_at || ticket.created_at)}</span>
                  </span>
                </span>
                <ChevronRight size={22} aria-hidden="true" />
              </button>
            );
          })}
        </div>
      </section>

      {composerOpen && (
        <div className="support-modal-backdrop" role="presentation">
          <section aria-label="提交工单" aria-modal="true" className="support-modal" role="dialog">
            <header>
              <span className="support-modal-icon" aria-hidden="true"><FileText size={22} /></span>
              <div><h2>提交工单</h2><p>请提供足够的信息，便于支持团队快速定位问题。</p></div>
              <button aria-label="关闭提交工单" className="support-icon-button" onClick={() => setComposerOpen(false)} title="关闭" type="button"><X size={20} /></button>
            </header>
            <form onSubmit={(event) => void createTicket(event)}>
              <label>
                <span>问题类型</span>
                <select aria-label="问题类型" onChange={(event) => setTopic(event.target.value)} value={topic}>
                  {topics.map((item) => <option key={item.key} value={item.name}>{item.name}</option>)}
                </select>
              </label>
              <label>
                <span>问题标题</span>
                <input aria-label="问题标题" maxLength={120} onChange={(event) => setTitle(event.target.value)} placeholder="请用一句话概括问题" value={title} />
              </label>
              <label>
                <span>问题描述</span>
                <textarea aria-label="问题描述" maxLength={500} onChange={(event) => setBody(event.target.value)} placeholder="请描述复现步骤、实际结果和期望结果..." value={body} />
                <small>{body.length}/500</small>
              </label>
              {submitError && <p className="support-form-error" role="alert">{submitError}</p>}
              <footer>
                <button className="support-cancel-button" onClick={() => setComposerOpen(false)} type="button">取消</button>
                <button className="support-submit-button" disabled={!topic || !title.trim() || !body.trim() || submitting} type="submit">
                  <Send size={17} aria-hidden="true" />
                  {submitting ? "提交中..." : "提交工单"}
                </button>
              </footer>
            </form>
          </section>
        </div>
      )}

      {selectedTicket && (
        <div className="support-modal-backdrop" role="presentation">
          <section aria-label="工单详情" aria-modal="true" className="support-modal support-detail-modal" role="dialog">
            <header>
              <span className="support-modal-icon detail" aria-hidden="true"><Ticket size={22} /></span>
              <div><small>{ticketNumber(selectedTicket)}</small><h2>{selectedTicket.title}</h2></div>
              <button aria-label="关闭工单详情" className="support-icon-button" onClick={() => setSelectedTicket(null)} title="关闭" type="button"><X size={20} /></button>
            </header>
            <div className="support-detail-summary">
              <span><small>状态</small><b className={`status-${normalizeTicketStatus(selectedTicket.status)}`}>{statusLabels[normalizeTicketStatus(selectedTicket.status)]}</b></span>
              <span><small>问题类型</small><strong>{selectedTicket.topic || "未分类"}</strong></span>
              <span><small>提交时间</small><strong>{formatTicketTime(selectedTicket.created_at)}</strong></span>
            </div>
            <article><h3>问题描述</h3><p>{selectedTicket.body || "该工单未提供详细描述。"}</p></article>
            <footer><button className="support-submit-button" onClick={() => setSelectedTicket(null)} type="button">关闭</button></footer>
          </section>
        </div>
      )}
    </V4PageShell>
  );
}

export default HelpPage;
