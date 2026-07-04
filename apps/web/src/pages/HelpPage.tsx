import { FormEvent, useEffect, useState } from "react";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type HelpArticle, type HelpTopic } from "../lib/contentApi";
import { supportApi, type SupportTicket } from "../lib/supportApi";

function HelpPage() {
  const [topics, setTopics] = useState<HelpTopic[]>([]);
  const [articles, setArticles] = useState<HelpArticle[]>([]);
  const [apiTickets, setApiTickets] = useState<SupportTicket[]>([]);
  const [topic, setTopic] = useState("");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [status, setStatus] = useState("");
  const [loadError, setLoadError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    Promise.all([
      contentApi.listHelpTopics(),
      contentApi.listHelpArticles({ limit: 10 }),
      supportApi.listTickets(20)
    ])
      .then(([topicsPayload, articlesPayload, ticketsPayload]) => {
        if (!active) return;
        setTopics(topicsPayload.topics);
        setArticles(articlesPayload.articles);
        setApiTickets(ticketsPayload.tickets);
        setTopic((current) => current || topicsPayload.topics[0]?.name || "");
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setLoadError(apiErrorMessage(error, "暂时无法读取帮助数据"));
      });
    return () => {
      active = false;
    };
  }, []);

  async function createTicket(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("");
    setSubmitting(true);
    try {
      const ticket = await supportApi.createTicket({ topic, title, body });
      setApiTickets((current) => [ticket, ...current]);
      setTitle("");
      setBody("");
      setStatus("反馈已提交");
    } catch (error) {
      setStatus(apiErrorMessage(error, "反馈提交失败"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <V4PageShell>
      <section className="help-page" aria-label="帮助与反馈">
        <div className="page-title-row">
          <div>
            <h1>帮助与反馈</h1>
            <p>自助答疑、问题反馈与人工咨询入口</p>
          </div>
        </div>
        {loadError && <p className="form-error" role="alert">{loadError}</p>}
        {status && <p className="form-success" role="status">{status}</p>}

        <div className="help-grid">
          <section className="help-card help-docs">
            <div className="help-title">
              <span className="help-art doc" aria-hidden="true" />
              <div>
                <h2>1. 自助答疑 / 帮助文档</h2>
                <p>搜索常见问题或浏览帮助文档，快速找到解决方案</p>
              </div>
            </div>
            <label className="help-search">
              <span aria-hidden="true">⌕</span>
              <input aria-label="搜索帮助文档" placeholder="搜索帮助文档，如“如何创建项目”" />
            </label>
            <div className="help-topic-list">
              {topics.length === 0 && <p>暂无帮助分类</p>}
              {topics.map((item) => (
                <button key={item.key} type="button">
                  <span className="account-menu-icon help" aria-hidden="true" />
                  {item.name}
                  <b aria-hidden="true">›</b>
                </button>
              ))}
            </div>
            <div className="ticket-list" aria-label="帮助文章">
              {articles.length === 0 && <p>暂无帮助文章</p>}
              {articles.map((article) => (
                <article key={article.slug}>
                  <strong>{article.title}</strong>
                  <span>{article.summary}</span>
                </article>
              ))}
            </div>
            <a href="/help">查看全部帮助文档 ›</a>
          </section>

          <form className="help-card feedback-form" onSubmit={(event) => void createTicket(event)}>
            <div className="help-title">
              <span className="help-art form" aria-hidden="true" />
              <div>
                <h2>2. 意见反馈 / 工单</h2>
                <p>请详细描述您的问题或建议，我们将尽快处理</p>
              </div>
            </div>
            <label>
              <span>问题类型</span>
              <select aria-label="问题类型" onChange={(event) => setTopic(event.target.value)} value={topic}>
                {topics.length === 0 && <option value="">暂无可选分类</option>}
                {topics.map((item) => <option key={item.key} value={item.name}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>问题标题</span>
              <input aria-label="问题标题" onChange={(event) => setTitle(event.target.value)} placeholder="请用一句话概括问题" value={title} />
            </label>
            <label>
              <span>问题描述</span>
              <textarea aria-label="问题描述" onChange={(event) => setBody(event.target.value)} placeholder="请详细描述您遇到的问题、建议或期望..." value={body} />
              <small>{body.length}/500</small>
            </label>
            <label>
              <span>附件上传（选填）</span>
              <div className="upload-box">
                <b aria-hidden="true">☁</b>
                点击或拖拽文件到此处上传
                <small>支持 jpg、png、pdf、doc、docx，单个文件不超过 10MB</small>
              </div>
            </label>
            <button className="submit-feedback" disabled={submitting} type="submit">{submitting ? "提交中..." : "提交反馈"}</button>
          </form>

          <section className="help-card ticket-card">
            <div className="module-section-head">
              <div>
                <h2>3. 反馈记录 / 工单状态</h2>
              </div>
              <a href="/help">查看全部 ›</a>
            </div>
            <div className="ticket-list">
              {apiTickets.length === 0 && <p>暂无反馈记录</p>}
              {apiTickets.map((ticket) => (
                <article key={ticket.id}>
                  <strong>{ticket.title}</strong>
                  <span>工单号：#{ticket.id}</span>
                  <span>提交时间：{ticket.created_at}</span>
                  <b className={ticket.status === "open" || ticket.status === "已提交" ? "submitted" : ticket.status === "处理中" ? "processing" : ""}>{ticket.status === "open" ? "已提交" : ticket.status}</b>
                </article>
              ))}
            </div>
          </section>
        </div>

        <section className="wechat-service-card">
          <span className="service-headset" aria-hidden="true" />
          <div>
            <h2>4. 联系企业微信（人工支持 / 升级咨询）</h2>
            <p>如需人工协助、产品演示或升级咨询，请添加企业微信，我们的团队将为您提供专业支持。</p>
          </div>
          <div className="qr-box" aria-label="企业微信二维码" />
          <div>
            <strong>企业微信在线服务</strong>
            <span>人工答疑 · 快速响应</span>
            <span>产品演示 · 方案咨询</span>
            <span>版本升级 · 定制服务</span>
          </div>
          <button type="button">联系企业微信</button>
        </section>
      </section>
    </V4PageShell>
  );
}

export default HelpPage;
