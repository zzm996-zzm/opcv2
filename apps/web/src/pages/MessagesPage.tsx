import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import {
  notificationsApi,
  type NotificationItem,
  type NotificationSummary
} from "../lib/notificationsApi";

const typeLabels: Record<string, string> = {
  task: "任务",
  system: "系统",
  analysis: "分析",
  lead: "线索",
  crm: "CRM",
  membership: "会员",
  marketing: "营销"
};

const iconByType: Record<string, string> = {
  task: "task",
  system: "system",
  analysis: "report",
  lead: "chart",
  crm: "done",
  membership: "gift",
  marketing: "news"
};

const messageTypes = ["task", "system", "analysis", "lead", "crm", "membership"] as const;

function MessagesPage() {
  const { messageId } = useParams();

  return (
    <V4PageShell>
      {messageId ? <MessageDetailPage messageId={messageId} /> : <MessageListPage />}
    </V4PageShell>
  );
}

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function unread(message: NotificationItem) {
  return !message.read_at;
}

function buildTabs(summary: NotificationSummary | null) {
  const counts = new Map(summary?.by_type.map((item) => [item.type, item.count]) ?? []);
  return [
    { type: "", label: "全部", count: summary?.unread ?? 0 },
    ...messageTypes.map((type) => ({ type, label: typeLabels[type], count: counts.get(type) ?? 0 }))
  ] as const;
}

function MessageListPage() {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [summary, setSummary] = useState<NotificationSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("");
  const [markingAllRead, setMarkingAllRead] = useState(false);
  const [activeType, setActiveType] = useState("");

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    setStatus("");
    setNotifications([]);
    Promise.all([
      notificationsApi.list({ ...(activeType ? { type: activeType } : {}), limit: 20 }),
      notificationsApi.summary()
    ])
      .then(([listPayload, summaryPayload]) => {
        if (!active) return;
        setNotifications(listPayload.notifications);
        setSummary(summaryPayload);
        setError("");
      })
      .catch((err) => {
        if (!active) return;
        setNotifications([]);
        setSummary(null);
        setError(apiErrorMessage(err, "暂时无法读取消息"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [activeType]);

  async function markAllRead() {
    if (markingAllRead || !summary?.unread) return;
    setMarkingAllRead(true);
    setStatus("");
    try {
      const result = await notificationsApi.markAllRead();
      const readAt = new Date().toISOString();
      setNotifications((current) => current.map((item) => ({ ...item, read_at: item.read_at ?? readAt })));
      setSummary((current) => current ? {
        ...current,
        unread: 0,
        by_type: current.by_type.map((item) => ({ ...item, count: 0 })),
        latest: current.latest.map((item) => ({ ...item, read_at: item.read_at ?? readAt }))
      } : current);
      setStatus(`已标记 ${result.updated} 条消息为已读`);
    } catch (err) {
      setStatus(apiErrorMessage(err, "暂时无法标记已读"));
    } finally {
      setMarkingAllRead(false);
    }
  }

  const tabs = buildTabs(summary);

  return (
    <section className="message-page message-list-page" aria-label="消息中心">
      <div className="message-card">
        <div className="message-card-head">
          <div>
            <h1>消息中心</h1>
            <p>查看与你相关的所有通知和消息</p>
          </div>
          <button disabled={!summary?.unread || markingAllRead} onClick={() => void markAllRead()} type="button">
            {markingAllRead ? "处理中..." : "全部已读"}
          </button>
        </div>

        {error && <p className="form-error" role="alert">{error}</p>}
        {status && <p className="form-success" role="status">{status}</p>}

        <div className="message-tabs" role="tablist" aria-label="消息分类">
          {tabs.map((tab) => (
            <button key={tab.type || "all"} className={activeType === tab.type ? "active" : ""} onClick={() => setActiveType(tab.type)} role="tab" aria-selected={activeType === tab.type} type="button">
              {tab.label} <span>{tab.count}</span>
            </button>
          ))}
        </div>

        <div className="message-list">
          {loading && <p>正在读取消息...</p>}
          {!loading && !error && notifications.length === 0 && <p>{activeType ? `暂无${typeLabels[activeType] ?? "此类"}消息` : "暂无消息"}</p>}
          {notifications.map((message) => (
            <Link key={message.id} className={`message-row ${unread(message) ? "unread" : ""}`} to={`/messages/${message.id}`}>
              <span className="message-unread-dot" aria-hidden="true" />
              <span className={`message-square-icon ${iconByType[message.type] ?? "system"}`} aria-hidden="true" />
              <span className="message-copy">
                <strong>{message.title}</strong>
                <small>{message.summary}</small>
              </span>
              <time>{formatTime(message.created_at)}</time>
              <span aria-hidden="true" className="message-arrow">›</span>
            </Link>
          ))}
        </div>

        <footer className="message-pagination" aria-label="消息分页">
          <button type="button" aria-label="上一页">‹</button>
          <button className="active" type="button">1</button>
          <button type="button" aria-label="下一页">›</button>
          <span>20条/页</span>
        </footer>
      </div>
    </section>
  );
}

function MessageDetailPage({ messageId }: { messageId: string }) {
  const navigate = useNavigate();
  const [message, setMessage] = useState<NotificationItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [markingRead, setMarkingRead] = useState(false);
  const [readStatus, setReadStatus] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    let active = true;
    const id = Number(messageId);
    if (!Number.isInteger(id) || id <= 0) {
      setError("消息 ID 不正确");
      setLoading(false);
      return () => {
        active = false;
      };
    }

    notificationsApi
      .get(id)
      .then((payload) => {
        if (!active) return;
        setMessage(payload);
        setError("");
        if (!payload.read_at) {
          setMarkingRead(true);
          notificationsApi
            .markRead(id)
            .then((updated) => {
              if (!active) return;
              setMessage(updated);
              setReadStatus("已标记为已读");
            })
            .catch((err) => {
              if (active) setReadStatus(apiErrorMessage(err, "暂时无法标记已读"));
            })
            .finally(() => {
              if (active) setMarkingRead(false);
            });
        }
      })
      .catch((err) => {
        if (!active) return;
        setMessage(null);
        setError(apiErrorMessage(err, "暂时无法读取消息详情"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [messageId]);

  async function markMessageRead() {
    const id = Number(messageId);
    if (!message || message.read_at || markingRead || !Number.isInteger(id) || id <= 0) return;
    setMarkingRead(true);
    setReadStatus("");
    try {
      const updated = await notificationsApi.markRead(id);
      setMessage(updated);
      setReadStatus("已标记为已读");
    } catch (err) {
      setReadStatus(apiErrorMessage(err, "暂时无法标记已读"));
    } finally {
      setMarkingRead(false);
    }
  }

  async function deleteMessage() {
    const id = Number(messageId);
    if (!message || deleting || !Number.isInteger(id) || id <= 0) return;
    if (!confirmDelete) {
      setConfirmDelete(true);
      return;
    }
    setDeleting(true);
    setError("");
    try {
      await notificationsApi.delete(id);
      navigate("/messages", { replace: true });
    } catch (err) {
      setError(apiErrorMessage(err, "暂时无法删除消息"));
      setConfirmDelete(false);
    } finally {
      setDeleting(false);
    }
  }

  return (
    <section className="message-page message-detail-page" aria-label="消息详情">
      <div className="detail-topline">
        <Link to="/messages">← 返回</Link>
        <div>
          <button disabled={!message || Boolean(message.read_at) || markingRead} onClick={() => void markMessageRead()} type="button">
            <span aria-hidden="true">▱</span> {markingRead ? "处理中..." : message?.read_at ? "已读" : "标记已读"}
          </button>
          <button className={`danger ${confirmDelete ? "confirm" : ""}`} disabled={!message || deleting || markingRead} onClick={() => void deleteMessage()} type="button">
            <span aria-hidden="true">⌫</span> {deleting ? "删除中..." : confirmDelete ? "确认删除" : "删除消息"}
          </button>
        </div>
      </div>

      {loading && <p>正在读取消息详情...</p>}
      {error && <p className="form-error" role="alert">{error}</p>}
      {readStatus && <p className={message?.read_at ? "form-success" : "form-error"} role={message?.read_at ? "status" : "alert"}>{readStatus}</p>}
      {message && <MessageDetail message={message} />}
    </section>
  );
}

type MessageDetailProps = {
  message: NotificationItem;
};

function MessageDetail({ message }: MessageDetailProps) {
  const category = typeLabels[message.type] ?? message.type;
  const actionLabel = message.action_label || "查看详情";
  const actionURL = message.action_url || "/messages";

  return (
    <article className="message-detail-card">
      <header>
        <div>
          <h1>{message.title}</h1>
          <p>智活AI团队 <span>|</span> {formatTime(message.created_at)}</p>
        </div>
        <span>{category}</span>
      </header>

      <div className="detail-content-box">
        <p><strong>您好！</strong></p>
        <p>{message.body || message.summary}</p>
        <div className="detail-actions">
          <Link to={actionURL}>{actionLabel}</Link>
          <Link to="/projects">去项目超市</Link>
        </div>
        <hr />
        <h2>相关操作</h2>
        <section className="related-actions" aria-label="相关操作">
          {[
            ["分享消息", "将消息分享给团队成员", "share"],
            ["查看来源", "打开关联业务页面", "download"],
            ["联系 Copilot", "针对消息内容继续追问", "refresh"]
          ].map(([title, desc, icon]) => (
            <button key={title} type="button">
              <span className={`related-icon ${icon}`} aria-hidden="true" />
              <strong>{title}</strong>
              <small>{desc}</small>
            </button>
          ))}
        </section>
        <p>如果你对消息内容有任何疑问，欢迎随时联系 <Link to="/copilot">智活 Copilot</Link> 助手为你解答。</p>
      </div>

      <footer className="detail-footer">
        <Link to="/messages">← 返回消息中心</Link>
      </footer>
    </article>
  );
}

export default MessagesPage;
