import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

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

function buildTabs(notifications: NotificationItem[], summary: NotificationSummary | null) {
  const counts = new Map(summary?.by_type.map((item) => [item.type, item.count]) ?? []);
  return [
    ["全部", notifications.length],
    ["任务", counts.get("task") ?? notifications.filter((item) => item.type === "task").length],
    ["系统", counts.get("system") ?? notifications.filter((item) => item.type === "system").length],
    ["营销", counts.get("marketing") ?? notifications.filter((item) => item.type === "marketing").length]
  ] as const;
}

function MessageListPage() {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [summary, setSummary] = useState<NotificationSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("");
  const [markingAllRead, setMarkingAllRead] = useState(false);

  useEffect(() => {
    let active = true;
    Promise.all([
      notificationsApi.list({ limit: 20 }),
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
  }, []);

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

  const tabs = buildTabs(notifications, summary);

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
          {tabs.map(([label, count], index) => (
            <button key={label} className={index === 0 ? "active" : ""} role="tab" aria-selected={index === 0} type="button">
              {label} <span>{count}</span>
            </button>
          ))}
        </div>

        <div className="message-list">
          {loading && <p>正在读取消息...</p>}
          {!loading && !error && notifications.length === 0 && <p>暂无消息</p>}
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
  const [message, setMessage] = useState<NotificationItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

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
          void notificationsApi.markRead(id);
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

  return (
    <section className="message-page message-detail-page" aria-label="消息详情">
      <div className="detail-topline">
        <Link to="/messages">← 返回</Link>
        <div>
          <button type="button"><span aria-hidden="true">▱</span> 标记已读</button>
          <button className="danger" type="button"><span aria-hidden="true">⌫</span> 删除</button>
        </div>
      </div>

      {loading && <p>正在读取消息详情...</p>}
      {error && <p className="form-error" role="alert">{error}</p>}
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
