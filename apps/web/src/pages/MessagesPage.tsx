import { Link, useParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const messageTabs = [
  ["全部", 12],
  ["任务", 3],
  ["系统", 4],
  ["营销", 2]
] as const;

const messages = [
  {
    id: "task-ai-hardware",
    title: "任务提醒：AI 智能硬件项目拆解完成",
    desc: "你发起的「AI 智能硬件」项目拆解已完成，点击查看拆解报告。",
    category: "任务",
    time: "今天 09:30",
    icon: "task",
    unread: true
  },
  {
    id: "market-report",
    title: "项目匹配推荐",
    desc: "根据你的画像，我们为你推荐了 3 个高潜力项目，快去看看吧！",
    category: "任务",
    time: "今天 08:45",
    icon: "report",
    unread: true
  },
  {
    id: "data-report",
    title: "数据报告生成完成",
    desc: "「小红书美妆行业数据报告」已生成，可前往数据获取查看。",
    category: "系统",
    time: "昨天 18:10",
    icon: "system",
    unread: true
  },
  {
    id: "news",
    title: "资讯更新",
    desc: "行业周报「AI 行业动态」第 8 期新内容，点击查看详情。",
    category: "营销",
    time: "昨天 16:20",
    icon: "news",
    unread: false
  },
  {
    id: "maintenance",
    title: "系统通知",
    desc: "系统将于 6 月 19 日 02:00-04:00 进行例行维护，期间部分功能可能受限。",
    category: "系统",
    time: "06-18 21:00",
    icon: "settings",
    unread: false
  },
  {
    id: "member",
    title: "会员权益更新",
    desc: "尊享会员权益已更新，新增 AI 学习计划专属模板，快去学习吧！",
    category: "营销",
    time: "06-18 10:30",
    icon: "gift",
    unread: false
  },
  {
    id: "weekly",
    title: "周报生成提醒",
    desc: "你的「竞品全盘数据周报」已生成，点击查看本周关键词象。",
    category: "系统",
    time: "06-17 09:00",
    icon: "chart",
    unread: false
  },
  {
    id: "task-done",
    title: "任务执行完成",
    desc: "你创建的任务「竞品监测：洗发日记」执行完成，查看结果。",
    category: "任务",
    time: "06-16 14:30",
    icon: "done",
    unread: false
  }
] as const;

function MessagesPage() {
  const { messageId } = useParams();
  const activeMessage = messages.find((message) => message.id === messageId) || messages[0];

  return (
    <V4PageShell>
      {messageId ? <MessageDetail message={activeMessage} /> : <MessageList />}
    </V4PageShell>
  );
}

function MessageList() {
  return (
    <section className="message-page message-list-page" aria-label="消息中心">
      <div className="message-card">
        <div className="message-card-head">
          <div>
            <h1>消息中心</h1>
            <p>查看与你相关的所有通知和消息</p>
          </div>
          <button type="button">全部已读</button>
        </div>

        <div className="message-tabs" role="tablist" aria-label="消息分类">
          {messageTabs.map(([label, count], index) => (
            <button key={label} className={index === 0 ? "active" : ""} role="tab" aria-selected={index === 0} type="button">
              {label} <span>{count}</span>
            </button>
          ))}
        </div>

        <div className="message-list">
          {messages.map((message) => (
            <Link key={message.id} className={`message-row ${message.unread ? "unread" : ""}`} to={`/messages/${message.id}`}>
              <span className="message-unread-dot" aria-hidden="true" />
              <span className={`message-square-icon ${message.icon}`} aria-hidden="true" />
              <span className="message-copy">
                <strong>{message.title}</strong>
                <small>{message.desc}</small>
              </span>
              <time>{message.time}</time>
              <span aria-hidden="true" className="message-arrow">›</span>
            </Link>
          ))}
        </div>

        <footer className="message-pagination" aria-label="消息分页">
          <button type="button" aria-label="上一页">‹</button>
          <button className="active" type="button">1</button>
          <button type="button">2</button>
          <button type="button">3</button>
          <button type="button" aria-label="下一页">›</button>
          <span>10条/页⌄</span>
        </footer>
      </div>
    </section>
  );
}

type MessageDetailProps = {
  message: (typeof messages)[number];
};

function MessageDetail({ message }: MessageDetailProps) {
  return (
    <section className="message-page message-detail-page" aria-label="消息详情">
      <div className="detail-topline">
        <Link to="/messages">← 返回</Link>
        <div>
          <button type="button"><span aria-hidden="true">▱</span> 标记未读</button>
          <button className="danger" type="button"><span aria-hidden="true">⌫</span> 删除</button>
        </div>
      </div>

      <article className="message-detail-card">
        <header>
          <div>
            <h1>{message.title}</h1>
            <p>智活AI团队 <span>|</span> {message.time}</p>
          </div>
          <span>{message.category}</span>
        </header>

        <div className="detail-content-box">
          <p><strong>张婧，您好！</strong></p>
          <p>你发起的「AI 智能硬件」项目拆解已完成，拆解报告已生成，包含以下内容：</p>
          <div className="detail-checklist" aria-label="拆解完成项">
            {["行业趋势分析", "市场规模与增长预测", "竞品全景分析", "用户画像与需求洞察", "商业模式与盈利分析", "营销策略建议"].map((item) => (
              <span key={item}>
                <i aria-hidden="true">✓</i>
                {item}
              </span>
            ))}
          </div>
          <p>点击下方按钮即可查看完整拆解报告，或在「项目超市」中找到该项目。</p>
          <div className="detail-actions">
            <Link to="/analysis">查看拆解报告</Link>
            <Link to="/projects">去项目超市</Link>
          </div>
          <hr />
          <h2>相关操作</h2>
          <section className="related-actions" aria-label="相关操作">
            {[
              ["分享报告", "将报告分享给团队成员", "share"],
              ["下载报告", "下载 PDF 格式报告", "download"],
              ["重新拆解", "基于最新数据重新分析", "refresh"]
            ].map(([title, desc, icon]) => (
              <button key={title} type="button">
                <span className={`related-icon ${icon}`} aria-hidden="true" />
                <strong>{title}</strong>
                <small>{desc}</small>
              </button>
            ))}
          </section>
          <p>如果你对报告有任何疑问，欢迎随时联系 <Link to="/copilot">智活 Copilot</Link> 助手为你解答。</p>
        </div>

        <footer className="detail-footer">
          <Link to="/messages/news">← 上一条</Link>
          <Link to="/messages/market-report">下一条 →</Link>
        </footer>
      </article>
    </section>
  );
}

export default MessagesPage;
