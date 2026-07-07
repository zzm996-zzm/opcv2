import { type FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { competitorApi, type CompetitorEvent, type CompetitorWatchItem } from "../lib/competitorApi";

const emptyMonitoringStats = [
  ["监测中竞品", "0"],
  ["今日新增动态", "0"],
  ["高风险预警", "0"],
  ["已生成任务", "0"]
] as const;

const channelHealth = [
  ["官网 / 价格页", "每 6 小时", "正常"],
  ["招聘动态", "每 12 小时", "正常"],
  ["内容矩阵", "每 3 小时", "密集"],
  ["投放素材", "每日", "排队"]
] as const;

const alertRules = [
  ["价格变化", "捕捉套餐、权益、免费试用和低价入口调整"],
  ["招聘扩张", "识别销售、增长、AI 产品岗位的异常增长"],
  ["内容爆发", "监测公众号、视频号、SEO 页面标题变化"],
  ["产品转向", "从产品页和案例页判断定位、场景和客群变化"]
] as const;

const defaultWatchChannels = ["官网 / 价格页", "招聘动态"] as const;

function formatEventTime(value: string) {
  return new Date(value).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function formatLastSeen(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function toTrackedCompetitor(item: CompetitorWatchItem) {
  return {
    id: item.id,
    name: item.name,
    category: item.category,
    status: item.status,
    threat: item.threat,
    lastSeen: formatLastSeen(item.last_seen_at),
    channels: item.channels,
    signal: item.signal
  };
}

function toTimelineRow(item: CompetitorEvent) {
  return [formatEventTime(item.occurred_at), item.company, item.title, item.detail, item.level] as const;
}

function CompetitorMonitoringPage() {
  const [watchlist, setWatchlist] = useState<CompetitorWatchItem[]>([]);
  const [events, setEvents] = useState<CompetitorEvent[]>([]);
  const [loadError, setLoadError] = useState("");
  const [showWatchForm, setShowWatchForm] = useState(false);
  const [watchName, setWatchName] = useState("");
  const [watchCategory, setWatchCategory] = useState("");
  const [formError, setFormError] = useState("");
  const [isSavingWatchItem, setIsSavingWatchItem] = useState(false);
  const [deletingWatchItemId, setDeletingWatchItemId] = useState<number | null>(null);

  useEffect(() => {
    let active = true;
    competitorApi
      .getMonitoring(20)
      .then((payload) => {
        if (!active) return;
        setWatchlist(payload.watchlist);
        setEvents(payload.events);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setWatchlist([]);
        setEvents([]);
        setLoadError(apiErrorMessage(error, "暂时无法读取竞品监测数据"));
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleCompetitors = watchlist.map(toTrackedCompetitor);
  const visibleTimeline = events.map(toTimelineRow);
  const highRiskCount = watchlist.filter((item) => item.threat === "强").length + events.filter((item) => item.level === "强").length;
  const visibleStats = watchlist.length > 0 || events.length > 0 ? [
    ["监测中竞品", String(watchlist.length)],
    ["今日新增动态", String(events.length)],
    ["高风险预警", String(highRiskCount)],
    ["已生成任务", "0"]
  ] as const : emptyMonitoringStats;
  const firstAlert = visibleTimeline.find(([, , , , level]) => level === "强") ?? visibleTimeline[0] ?? null;

  async function createWatchItem(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSavingWatchItem) return;
    setIsSavingWatchItem(true);
    setFormError("");
    try {
      const item = await competitorApi.createWatchItem({
        name: watchName,
        category: watchCategory,
        channels: [...defaultWatchChannels]
      });
      setWatchlist((current) => [item, ...current]);
      setWatchName("");
      setWatchCategory("");
      setShowWatchForm(false);
    } catch (error) {
      setFormError(apiErrorMessage(error, "暂时无法新增监测对象"));
    } finally {
      setIsSavingWatchItem(false);
    }
  }

  async function deleteWatchItem(item: CompetitorWatchItem) {
    if (!item.id || deletingWatchItemId) return;
    setDeletingWatchItemId(item.id);
    try {
      await competitorApi.deleteWatchItem(item.id);
      setWatchlist((current) => current.filter((watchItem) => watchItem.id !== item.id));
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法移除监测对象"));
    } finally {
      setDeletingWatchItemId(null);
    }
  }

  return (
    <V4PageShell className="competitor-monitoring-shell">
      <section className="module-page competitor-monitoring-page" aria-label="竞品动态监测">
        <div className="page-title-row">
          <div>
            <h1>竞品动态监测</h1>
            <p>持续盯住竞品的价格、招聘、内容、投放和产品页变化，把异常信号自动沉淀成反击动作</p>
          </div>
          <button className="module-primary-action" onClick={() => setShowWatchForm((visible) => !visible)} type="button">新增监测对象</button>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
        {showWatchForm ? (
          <form className="monitoring-watch-form" onSubmit={(event) => void createWatchItem(event)}>
            <label>
              <span>监测对象名称</span>
              <input aria-label="监测对象名称" onChange={(event) => setWatchName(event.target.value)} placeholder="例如：增长雷达" value={watchName} />
            </label>
            <label>
              <span>对象分类</span>
              <input aria-label="对象分类" onChange={(event) => setWatchCategory(event.target.value)} placeholder="例如：商业情报" value={watchCategory} />
            </label>
            <div className="monitoring-watch-channels" aria-label="默认监测渠道">
              {defaultWatchChannels.map((channel) => <span key={channel}>{channel}</span>)}
            </div>
            <button disabled={isSavingWatchItem} type="submit">{isSavingWatchItem ? "保存中..." : "保存监测对象"}</button>
            {formError ? <p className="form-error" role="alert">{formError}</p> : null}
          </form>
        ) : null}

        <section className="module-overview-card competitor-monitoring-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">实时竞争雷达</span>
            <h2>从“偶尔看看竞品”升级成持续预警系统</h2>
            <p>监测规则会自动巡检竞品公开页面和内容渠道，识别高频变化、定位漂移与获客动作，并把可执行建议推送到任务中心。</p>
            <div className="module-stat-strip">
              {visibleStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <div className="monitoring-radar-card" aria-label="竞品动态雷达">
            <div className="monitoring-radar">
              <span className="radar-sweep" aria-hidden="true" />
              <i className="dot hot" aria-hidden="true" />
              <i className="dot warm" aria-hidden="true" />
              <i className="dot cool" aria-hidden="true" />
              <strong>{highRiskCount}</strong>
              <small>高风险预警</small>
            </div>
            <p>{firstAlert ? `${firstAlert[1]}：${firstAlert[2]}` : "暂无高风险预警，新增监测对象后这里会展示最新异常信号。"}</p>
          </div>
        </section>

        <section className="monitoring-grid">
          <div className="monitoring-main-card">
            <div className="module-section-head">
              <div>
                <h2>监测中竞品</h2>
                <p>按威胁等级和最近变化排序，点击后可进入全盘数据破解</p>
              </div>
              <div className="module-chip-row compact">
                {["全部", "强预警", "价格变化", "招聘扩张"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>

            <div className="monitoring-competitor-list">
              {visibleCompetitors.length === 0 ? (
                <div className="module-empty-state" role="status">暂无监测对象</div>
              ) : visibleCompetitors.map((item) => (
                <article key={item.name}>
                  <span className={`monitoring-pulse ${item.threat === "强" ? "hot" : ""}`} aria-hidden="true" />
                  <div>
                    <h3>{item.name}</h3>
                    <small>{item.category} · {item.lastSeen}</small>
                  </div>
                  <em className={item.threat === "强" ? "hot" : ""}>威胁 {item.threat}</em>
                  <span className="monitoring-state">{item.status}</span>
                  {item.id ? (
                    <button
                      aria-label={`移除 ${item.name}`}
                      className="monitoring-remove-button"
                      disabled={deletingWatchItemId === item.id}
                      onClick={() => void deleteWatchItem(item)}
                      type="button"
                    >
                      {deletingWatchItemId === item.id ? "移除中" : "移除"}
                    </button>
                  ) : null}
                  <p>{item.signal}</p>
                  <div className="tool-tags">
                    {item.channels.map((channel) => <span key={channel}>{channel}</span>)}
                  </div>
                </article>
              ))}
            </div>
          </div>

          <aside className="monitoring-side-card" aria-label="监测频率">
            <h2>监测频率</h2>
            {channelHealth.map(([source, frequency, status]) => (
              <article key={source}>
                <span>
                  <strong>{source}</strong>
                  <small>{frequency}</small>
                </span>
                <em>{status}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="monitoring-lower-grid">
          <div className="monitoring-timeline-card">
            <div className="module-section-head">
              <div>
                <h2>动态时间线</h2>
                <p>把散落的变化按时间串起来，方便判断竞品动作是否连续</p>
              </div>
            </div>
            <div className="monitoring-timeline">
              {visibleTimeline.length === 0 ? (
                <div className="module-empty-state" role="status">暂无监测动态</div>
              ) : visibleTimeline.map(([time, company, title, detail, level]) => (
                <article key={`${time}-${title}`}>
                  <time>{time}</time>
                  <div>
                    <strong>{company}</strong>
                    <h3>{title}</h3>
                    <p>{detail}</p>
                  </div>
                  <em className={level === "强" ? "hot" : ""}>{level}</em>
                </article>
              ))}
            </div>
          </div>

          <aside className="monitoring-action-card" aria-label="预警动作">
            <h2>今日预警</h2>
            <strong>{firstAlert ? `先处理${firstAlert[1]}动态` : "暂无今日预警"}</strong>
            <p>{firstAlert ? firstAlert[3] : "后端返回监测事件后，这里会展示需要优先处理的预警动作。"}</p>
            <div>
              {firstAlert ? <span>{firstAlert[2]}</span> : <span>暂无反击任务</span>}
            </div>
            <Link to="/tasks">生成反击任务</Link>
          </aside>
        </section>

        <section className="monitoring-rule-section">
          <div className="module-section-head">
            <div>
              <h2>预警规则</h2>
              <p>第一版先用固定规则承接，后续可接入真实爬虫、队列和模型评分</p>
            </div>
          </div>
          <div className="monitoring-rule-grid">
            {alertRules.map(([title, detail]) => (
              <article key={title}>
                <strong>{title}</strong>
                <p>{detail}</p>
              </article>
            ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

export default CompetitorMonitoringPage;
