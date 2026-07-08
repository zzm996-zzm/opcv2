import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { competitorApi, type CompetitorScan } from "../lib/competitorApi";
import { membershipApi, type MembershipUsageItem } from "../lib/membershipApi";
import { quotaKeys, quotaSummary } from "../lib/quotaUsage";
import { tasksApi } from "../lib/tasksApi";

const emptyDataStats = [
  ["采集完成", "0%"],
  ["竞品数", "0"],
  ["高威胁", "0"],
  ["AI结论", "0"]
] as const;

const dataSources = [
  ["官网页面", "产品、价格、案例、更新日志", "已采集"],
  ["招聘动态", "岗位、团队扩张、重点能力", "采集中"],
  ["内容矩阵", "公众号、视频号、SEO 页面", "已采集"],
  ["投放素材", "关键词、落地页、转化钩子", "排队中"]
] as const;

const taskFlow = [
  ["1", "输入竞品", "域名、品牌名、关键词或截图"],
  ["2", "脚本采集", "官网、招聘、内容、投放与价格页"],
  ["3", "AI 清洗", "去重、归类、识别异常变化"],
  ["4", "生成结论", "输出威胁等级和应对动作"]
] as const;

const defaultScanTargets = ["小鹅通", "有赞教育", "企微管家"];
const defaultScanFocus = "价格、案例、招聘和 AI 功能";

function parseScanPlanInput(rawInput: string) {
  const [targetPart = "", ...focusParts] = rawInput.trim().split(/[；;]/);
  const targets = targetPart
    .split(/[,\uFF0C、/|\s]+/)
    .map((item) => item.trim())
    .filter(Boolean);
  const focus = focusParts
    .join("；")
    .replace(/^(重点)?关注[:：]?/, "")
    .trim();
  return {
    targets,
    focus: focus || defaultScanFocus
  };
}

function scanStatusCopy(scan: CompetitorScan | null) {
  if (!scan) {
    return { label: "未开始", detail: "启动一次采集任务后，这里会显示脚本队列与处理进度。", progress: 0 };
  }
  if (scan.status === "queued") {
    return { label: "排队中", detail: "脚本任务已排队，等待采集账号执行。", progress: scan.progress_percent };
  }
  if (scan.status === "running") {
    return { label: "采集中", detail: "脚本正在采集公开数据，完成后会生成 AI 破解结论。", progress: scan.progress_percent };
  }
  if (scan.status === "failed") {
    return { label: "采集失败", detail: scan.error_message || "脚本采集失败，请稍后重试或联系运营检查账号池。", progress: scan.progress_percent };
  }
  return { label: "已完成", detail: "采集与 AI 分析已完成，可以查看竞品画像和破解结论。", progress: 100 };
}

function shouldPollScan(scan: CompetitorScan | null) {
  return scan?.status === "queued" || scan?.status === "running";
}

function CompetitorDataPage() {
  const [latestScan, setLatestScan] = useState<CompetitorScan | null>(null);
  const [loadError, setLoadError] = useState("");
  const [isScanning, setIsScanning] = useState(false);
  const [isRetrying, setIsRetrying] = useState(false);
  const [watchlistMessage, setWatchlistMessage] = useState("");
  const [addingWatchCompetitor, setAddingWatchCompetitor] = useState("");
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [taskMessage, setTaskMessage] = useState("");
  const [scanPlanInput, setScanPlanInput] = useState("");
  const [scanPlanError, setScanPlanError] = useState("");
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const scanQuota = quotaSummary(usage, quotaKeys.competitorScans, "竞品全盘数据破解");

  useEffect(() => {
    let active = true;
    void competitorApi.listScans(20)
      .then((payload) => {
        if (!active) return;
        setLatestScan(payload.scans[0] ?? null);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setLatestScan(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取竞品采集数据"));
      });
    void membershipApi.usage()
      .then((payload) => {
        if (active) setUsage(payload.usage ?? []);
      })
      .catch(() => {
        if (active) setUsage([]);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!latestScan || !shouldPollScan(latestScan)) return;
    const scanToPoll = latestScan;
    let active = true;
    const timer = window.setInterval(() => {
      void competitorApi
        .getScan(scanToPoll.id)
        .then((scan) => {
          if (active) {
            setLatestScan(scan);
          }
        })
        .catch(() => {
          // Keep the last visible status; the next interval can recover.
        });
    }, 3000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [latestScan?.id, latestScan?.status]);

  const statusCopy = scanStatusCopy(latestScan);
  const visibleStats = latestScan ? [
    ["采集完成", `${statusCopy.progress}%`],
    ["竞品数", String(latestScan.competitors.length)],
    ["高威胁", String(latestScan.competitors.filter((item) => item.risk === "强").length)],
    ["AI结论", String(latestScan.conclusions.length)]
  ] as const : emptyDataStats;
  const visibleCompetitors = latestScan?.competitors ?? [];
  const visibleConclusions = latestScan?.conclusions.map((item) => [item.title, item.detail] as const) ?? [];
  const visibleEvidenceSources = latestScan?.evidence_sources ?? [];
  const primaryConclusion = latestScan?.conclusions[0];

  async function startScan() {
    if (isScanning) return;
    if (scanQuota.blocked) {
      setLoadError("本月竞品全盘数据破解额度已用完，请升级套餐或等待下月重置。");
      return;
    }
    setIsScanning(true);
    try {
      const scan = await competitorApi.createScan({
        targets: defaultScanTargets,
        focus: defaultScanFocus
      });
      setLatestScan(scan);
      const usagePayload = await membershipApi.usage().catch(() => null);
      if (usagePayload) setUsage(usagePayload.usage ?? []);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法启动采集任务"));
    } finally {
      setIsScanning(false);
    }
  }

  async function retryScan() {
    if (isRetrying || !latestScan) return;
    setIsRetrying(true);
    try {
      const scan = await competitorApi.retryScan(latestScan.id);
      setLatestScan(scan);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法重新采集"));
    } finally {
      setIsRetrying(false);
    }
  }

  async function addCompetitorToWatchlist(name: string) {
    if (!latestScan || addingWatchCompetitor) return;
    setAddingWatchCompetitor(name);
    setWatchlistMessage("");
    try {
      const item = await competitorApi.addScanCompetitorToWatchlist(latestScan.id, name);
      setLoadError("");
      setWatchlistMessage(`已将${item.name}加入动态监测。`);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法加入动态监测"));
    } finally {
      setAddingWatchCompetitor("");
    }
  }

  async function createCounterTask() {
    if (isCreatingTask || !primaryConclusion) return;
    setIsCreatingTask(true);
    setTaskMessage("");
    try {
      const task = await tasksApi.createTask({
        title: `反击：${primaryConclusion.title}`,
        project: "竞品动态监测",
        priority: "high",
        tools: ["竞品全盘数据破解", "任务中心"],
        learning: primaryConclusion.detail
      });
      setLoadError("");
      setTaskMessage(`已生成反击任务：${task.title}`);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法生成反击任务"));
    } finally {
      setIsCreatingTask(false);
    }
  }

  async function submitScanPlan() {
    if (isScanning) return;
    if (scanQuota.blocked) {
      setScanPlanError("本月竞品全盘数据破解额度已用完，请升级套餐或等待下月重置。");
      return;
    }
    const input = parseScanPlanInput(scanPlanInput);
    if (input.targets.length === 0) {
      setScanPlanError("请输入至少一个竞品或关键词");
      return;
    }
    setIsScanning(true);
    setScanPlanError("");
    try {
      const scan = await competitorApi.createScan(input);
      setLatestScan(scan);
      setLoadError("");
      const usagePayload = await membershipApi.usage().catch(() => null);
      if (usagePayload) setUsage(usagePayload.usage ?? []);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法生成采集计划"));
    } finally {
      setIsScanning(false);
    }
  }

  return (
    <V4PageShell className="competitor-data-shell">
      <section className="module-page competitor-data-page" aria-label="竞品全盘数据破解">
        <div className="page-title-row">
          <div>
            <h1>竞品全盘数据破解</h1>
            <p>发起脚本代查，自动采集竞品公开数据，再交给 AI 提炼威胁、机会和反击动作</p>
          </div>
          <div className={scanQuota.blocked ? "module-quota-inline depleted" : "module-quota-inline"}>
            <small>{scanQuota.label}</small>
            <strong>{scanQuota.value}</strong>
            {scanQuota.blocked ? <Link to="/membership">升级套餐</Link> : <span>{scanQuota.unit}</span>}
          </div>
          <button className="module-primary-action" disabled={isScanning || scanQuota.blocked} onClick={() => void startScan()} type="button">
            {isScanning ? "采集中..." : "启动采集任务"}
          </button>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
        {watchlistMessage ? <p className="form-success" role="status">{watchlistMessage}</p> : null}
        {taskMessage ? <p className="form-success" role="status">{taskMessage}</p> : null}

        <section className="module-overview-card competitor-data-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">{latestScan ? latestScan.targets.join(" / ") : "暂无扫描任务"}</span>
            <h2>把分散的公开信号合成一张可行动的竞品地图</h2>
            <p>系统会从官网、招聘、内容、价格、案例和投放素材中提取变化，识别竞品正在抢什么客户、推什么能力、用什么话术。</p>
            <div className="module-empty-state" role="status">
              <strong>{statusCopy.label}</strong>
              <span>{statusCopy.detail}</span>
              {latestScan?.status === "failed" ? (
                <button className="competitor-retry-button" disabled={isRetrying} onClick={() => void retryScan()} type="button">
                  {isRetrying ? "提交中..." : "重新采集"}
                </button>
              ) : null}
            </div>
            <div className="module-stat-strip">
              {visibleStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact competitor-data-input" onSubmit={(event) => {
            event.preventDefault();
            void submitScanPlan();
          }}>
            <label htmlFor="competitor-target">输入竞品或关键词</label>
            <textarea
              id="competitor-target"
              aria-label="输入竞品或关键词"
              onChange={(event) => setScanPlanInput(event.target.value)}
              placeholder="例如：小鹅通、有赞教育、企微管家；重点关注价格、案例、招聘和 AI 功能..."
              value={scanPlanInput}
            />
            {scanPlanError ? <small className="form-error" role="alert">{scanPlanError}</small> : null}
            <button disabled={isScanning || scanQuota.blocked} type="submit">{isScanning ? "生成中..." : "生成采集计划"}</button>
          </form>
        </section>

        <section className="competitor-data-grid">
          <div className="competitor-data-main">
            <div className="module-section-head">
              <div>
                <h2>采集任务流</h2>
                <p>脚本代查会拆成可追踪节点，方便部署后接入真实队列</p>
              </div>
            </div>
            <div className="competitor-flow">
              {taskFlow.map(([step, title, detail]) => (
                <article key={title}>
                  <b>{step}</b>
                  <strong>{title}</strong>
                  <small>{detail}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="competitor-data-side" aria-label="数据源状态">
            <h2>数据源状态</h2>
            {dataSources.map(([source, detail, status]) => (
              <article key={source}>
                <span>
                  <strong>{source}</strong>
                  <small>{detail}</small>
                </span>
                <em>{status}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="competitor-card-section">
          <div className="module-section-head">
            <div>
              <h2>重点竞品画像</h2>
              <p>AI 根据公开变化生成威胁评分和关键动作</p>
            </div>
            <div className="module-chip-row compact">
              {["全部", "高威胁", "价格变化", "招聘扩张"].map((view, index) => (
                <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
              ))}
            </div>
          </div>
          <div className="competitor-card-grid">
            {visibleCompetitors.length === 0 ? (
              <div className="module-empty-state" role="status">暂无竞品画像</div>
            ) : visibleCompetitors.map((item) => (
              <article key={item.name}>
                <header>
                  <div>
                    <h2>{item.name}</h2>
                    <p>{item.category}</p>
                  </div>
                  <strong>{String(item.score)}</strong>
                </header>
                <p>{item.signal}</p>
                <div className="tool-tags">
                  {item.tags.map((tag) => <span key={tag}>{tag}</span>)}
                </div>
                <footer>
                  <span className={item.risk === "强" ? "high" : ""}>威胁 {item.risk}</span>
                  <button
                    aria-label={`加入监测 ${item.name}`}
                    disabled={addingWatchCompetitor === item.name}
                    onClick={() => void addCompetitorToWatchlist(item.name)}
                    type="button"
                  >
                    {addingWatchCompetitor === item.name ? "加入中" : "加入监测"}
                  </button>
                </footer>
              </article>
            ))}
          </div>
        </section>

        <section className="competitor-lower-grid">
          <div className="competitor-conclusion-card">
            <div className="module-section-head">
              <div>
                <h2>AI 破解结论</h2>
                <p>从数据变化里抽出能指导打法的判断</p>
              </div>
            </div>
            <div className="competitor-conclusion-list">
              {visibleConclusions.length === 0 ? (
                <div className="module-empty-state" role="status">暂无破解结论</div>
              ) : visibleConclusions.map(([title, detail]) => (
                <article key={title}>
                  <strong>{title}</strong>
                  <p>{detail}</p>
                </article>
              ))}
            </div>
            {visibleEvidenceSources.length > 0 ? (
              <div className="competitor-evidence-list">
                <h2>证据来源</h2>
                {visibleEvidenceSources.slice(0, 4).map((source) => (
                  <article key={`${source.source_type}-${source.url}-${source.title}`}>
                    <span>
                      <strong>{source.title}</strong>
                      <small>{source.source_type}</small>
                    </span>
                    <p>{source.summary}</p>
                    <a href={source.url} rel="noreferrer" target="_blank">打开来源</a>
                  </article>
                ))}
              </div>
            ) : null}
          </div>

          <aside className="competitor-action-card" aria-label="建议动作">
            <h2>建议动作</h2>
            <strong>{visibleConclusions.length > 0 ? "根据破解结论生成反击任务" : "暂无建议动作"}</strong>
            <p>{visibleConclusions[0]?.[1] ?? "完成一次竞品采集后，这里会显示后端生成的反击建议。"}</p>
            <button disabled={!primaryConclusion || isCreatingTask} onClick={() => void createCounterTask()} type="button">
              {isCreatingTask ? "生成中..." : "生成反击任务"}
            </button>
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default CompetitorDataPage;
