import { type FormEvent, useEffect, useState } from "react";
import { useLocation } from "react-router-dom";

import UnifiedCopilotPanel from "../components/UnifiedCopilotPanel";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { competitorApi, type CompetitorEvent, type CompetitorWatchItem } from "../lib/competitorApi";
import { tasksApi } from "../lib/tasksApi";

const emptyMonitoringStats = [
  ["关注对象", "0"],
  ["已记录动态", "0"],
  ["待处理提醒", "0"],
  ["已生成任务", "0"]
] as const;

const observationDimensions = [
  ["官网 / 价格页", "记录价格与权益变化", "关注中"],
  ["招聘动态", "记录组织与岗位信号", "关注中"],
  ["内容发布", "记录主题与营销动作", "关注中"],
  ["产品动态", "记录产品与定位变化", "关注中"]
] as const;

const alertRules = [
  ["价格变化", "归类套餐、权益、免费试用和低价入口调整"],
  ["招聘扩张", "归类销售、增长和 AI 产品岗位相关的变化"],
  ["内容爆发", "归类公众号、视频号和 SEO 页面标题变化"],
  ["产品转向", "归类定位、场景和客群相关的信息"]
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
    ...item,
    status: item.status === "监测中" ? "已关注" : item.status,
    signal: item.signal === "已创建监测规则，等待首次巡检。" ? "已添加关注对象，等待补充动态。" : item.signal,
    lastSeen: formatLastSeen(item.last_seen_at)
  };
}

function toTimelineRow(item: CompetitorEvent) {
  return [formatEventTime(item.occurred_at), item.company, item.title, item.detail, item.level] as const;
}

function CompetitorMonitoringPage() {
  const location = useLocation();
  const [watchlist, setWatchlist] = useState<CompetitorWatchItem[]>([]);
  const [events, setEvents] = useState<CompetitorEvent[]>([]);
  const [loadError, setLoadError] = useState("");
  const [showWatchForm, setShowWatchForm] = useState(false);
  const [watchName, setWatchName] = useState("");
  const [watchCategory, setWatchCategory] = useState("");
  const [formError, setFormError] = useState("");
  const [isSavingWatchItem, setIsSavingWatchItem] = useState(false);
  const [deletingWatchItemId, setDeletingWatchItemId] = useState<number | null>(null);
  const [startingScanItemId, setStartingScanItemId] = useState<number | null>(null);
  const [scanLaunchMessage, setScanLaunchMessage] = useState("");
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [taskMessage, setTaskMessage] = useState("");
  const [generatedTaskCount, setGeneratedTaskCount] = useState(0);

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
    ["关注对象", String(watchlist.length)],
    ["已记录动态", String(events.length)],
    ["待处理提醒", String(highRiskCount)],
    ["已生成任务", String(generatedTaskCount)]
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

  async function startWatchItemScan(item: CompetitorWatchItem) {
    if (!item.id || startingScanItemId) return;
    setStartingScanItemId(item.id);
    setScanLaunchMessage("");
    try {
      await competitorApi.startWatchItemScan(item.id);
      setLoadError("");
      setScanLaunchMessage(`已创建 ${item.name} 的分析任务，正在排队处理。`);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法创建分析任务"));
    } finally {
      setStartingScanItemId(null);
    }
  }

  async function createAlertTask() {
    if (!firstAlert || isCreatingTask) return;
    setIsCreatingTask(true);
    setTaskMessage("");
    try {
      const task = await tasksApi.createTask({
        title: `竞品动态跟进：${firstAlert[1]} ${firstAlert[2]}`,
        project: "竞品动态监测",
        priority: firstAlert[4] === "强" ? "high" : "medium",
        tools: ["竞品动态监测", "任务中心"],
        learning: firstAlert[3],
        sourceType: "competitor_monitoring",
        sourceTitle: `竞品监测：${firstAlert[1]} ${firstAlert[2]}`,
        sourceUrl: "/competitor-monitoring"
      });
      setLoadError("");
      setGeneratedTaskCount((count) => count + 1);
      setTaskMessage(`已生成跟进任务：${task.title}`);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法生成跟进任务"));
    } finally {
      setIsCreatingTask(false);
    }
  }

  return (
    <V4PageShell className="competitor-monitoring-shell">
      <section className="module-page competitor-monitoring-page" aria-label="竞品动态监测">
        <div className="competitor-monitoring-content">
          <div className="page-title-row">
          <div>
            <h1>竞品动态监测</h1>
            <p>集中记录竞品的价格、招聘、内容和产品变化，由 AI 帮你梳理影响并形成下一步动作</p>
          </div>
          <button className="module-primary-action" onClick={() => setShowWatchForm((visible) => !visible)} type="button">新增监测对象</button>
          </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
        {scanLaunchMessage ? <p className="form-success" role="status">{scanLaunchMessage}</p> : null}
        {taskMessage ? <p className="form-success" role="status">{taskMessage}</p> : null}
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
            <span className="module-kicker">测试版 · 竞品观察</span>
            <h2>把零散的竞品信息整理成可执行的判断</h2>
            <p>基于已添加的关注对象和已记录动态，生成分析建议并整理下一步动作。</p>
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
              <small>待处理提醒</small>
            </div>
            <p>{firstAlert ? `${firstAlert[1]}：${firstAlert[2]}` : "暂无待处理提醒，添加关注对象或记录动态后可在这里查看优先事项。"}</p>
          </div>
        </section>

        <section className="monitoring-grid">
          <div className="monitoring-main-card">
            <div className="module-section-head">
              <div>
                <h2>关注对象</h2>
                <p>查看已添加的关注对象，并基于已记录动态安排后续分析</p>
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
                    <div className="monitoring-row-actions">
                      <button
                        aria-label={`分析 ${item.name}`}
                        className="monitoring-scan-button"
                        disabled={startingScanItemId === item.id}
                        onClick={() => void startWatchItemScan(item)}
                        type="button"
                      >
                        {startingScanItemId === item.id ? "创建中" : "创建分析"}
                      </button>
                      <button
                        aria-label={`移除 ${item.name}`}
                        className="monitoring-remove-button"
                        disabled={deletingWatchItemId === item.id}
                        onClick={() => void deleteWatchItem(item)}
                        type="button"
                      >
                        {deletingWatchItemId === item.id ? "移除中" : "移除"}
                      </button>
                    </div>
                  ) : null}
                  <p>{item.signal}</p>
                  <div className="tool-tags">
                    {item.channels.map((channel) => <span key={channel}>{channel}</span>)}
                  </div>
                </article>
              ))}
            </div>
          </div>

          <aside className="monitoring-side-card" aria-label="关注维度">
            <h2>关注维度</h2>
            {observationDimensions.map(([source, description, status]) => (
              <article key={source}>
                <span>
                  <strong>{source}</strong>
                  <small>{description}</small>
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
            <h2>优先处理</h2>
            <strong>{firstAlert ? `先处理${firstAlert[1]}动态` : "暂无待处理动态"}</strong>
            <p>{firstAlert ? firstAlert[3] : "后端返回监测事件后，这里会展示需要优先处理的预警动作。"}</p>
            <div>
              {firstAlert ? <span>{firstAlert[2]}</span> : <span>暂无跟进任务</span>}
            </div>
            <button disabled={!firstAlert || isCreatingTask} onClick={() => void createAlertTask()} type="button">
              {isCreatingTask ? "生成中..." : "生成跟进任务"}
            </button>
          </aside>
        </section>

        <section className="monitoring-rule-section">
          <div className="module-section-head">
            <div>
              <h2>预警规则</h2>
              <p>测试版先按固定规则归类已记录动态，方便 AI 解读优先级和建议动作</p>
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
        </div>
        <UnifiedCopilotPanel
          activeFilters={{ module: "monitoring", view: "home" }}
          ariaLabel="竞品动态监测 Copilot"
          className="monitoring-copilot-panel"
          currentView={`${location.pathname}${location.search}`}
          inputAriaLabel="向竞品动态监测 Copilot 提问"
          response={firstAlert ? "我会结合待处理事项、关注对象和已记录动态，分析影响并给出应对动作。" : "关注对象和动态会在这里形成上下文；有数据后我可以帮你排序优先级和安排动作。"}
          userPrompt={firstAlert ? `分析待处理事项：${firstAlert[1]} ${firstAlert[2]}` : "分析当前关注对象和动态"}
        />
      </section>
    </V4PageShell>
  );
}

export default CompetitorMonitoringPage;
