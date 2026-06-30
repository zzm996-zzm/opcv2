import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { competitorApi, type CompetitorScan } from "../lib/competitorApi";

const dataStats = [
  ["采集完成", "86%"],
  ["数据源", "7"],
  ["异常信号", "12"],
  ["AI结论", "5"]
] as const;

const dataSources = [
  ["官网页面", "产品、价格、案例、更新日志", "已采集"],
  ["招聘动态", "岗位、团队扩张、重点能力", "采集中"],
  ["内容矩阵", "公众号、视频号、SEO 页面", "已采集"],
  ["投放素材", "关键词、落地页、转化钩子", "排队中"]
] as const;

const competitors = [
  {
    name: "小鹅通",
    category: "知识付费 / 企业培训",
    score: "91",
    signal: "近期强调 AI 助教、直播转化和企微私域联动",
    risk: "强",
    tags: ["价格页更新", "招聘增长", "内容密集"]
  },
  {
    name: "有赞教育",
    category: "教育 SaaS / 私域运营",
    score: "84",
    signal: "案例页新增连锁培训机构，主打多门店运营能力",
    risk: "中",
    tags: ["案例新增", "渠道扩张", "低价套餐"]
  },
  {
    name: "企微管家",
    category: "CRM / 客户运营",
    score: "76",
    signal: "内容重点从 SCRM 转向 AI 线索跟进和客户分层",
    risk: "中",
    tags: ["定位调整", "关键词变化", "销售招聘"]
  }
] as const;

const conclusions = [
  ["定位变化", "头部竞品正在从工具售卖转向“AI + 私域增长方案”，单纯功能对比已经不够。"],
  ["价格策略", "低门槛套餐用于获客，高阶功能绑定企微、直播和数据分析能力，利于后续升级。"],
  ["获客重点", "内容和案例集中押注教育培训、知识付费、企业内训三个高频场景。"],
  ["反击建议", "优先打造“AI课程增长 + 企微转化”的组合案例，用结果页和落地任务承接。"]
] as const;

const taskFlow = [
  ["1", "输入竞品", "域名、品牌名、关键词或截图"],
  ["2", "脚本采集", "官网、招聘、内容、投放与价格页"],
  ["3", "AI 清洗", "去重、归类、识别异常变化"],
  ["4", "生成结论", "输出威胁等级和应对动作"]
] as const;

function CompetitorDataPage() {
  const [latestScan, setLatestScan] = useState<CompetitorScan | null>(null);
  const [loadError, setLoadError] = useState("");
  const [isScanning, setIsScanning] = useState(false);

  useEffect(() => {
    let active = true;
    competitorApi
      .listScans(20)
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
    return () => {
      active = false;
    };
  }, []);

  const visibleCompetitors = latestScan?.competitors.length ? latestScan.competitors : competitors;
  const visibleConclusions = latestScan?.conclusions.length
    ? latestScan.conclusions.map((item) => [item.title, item.detail] as const)
    : conclusions;

  async function startScan() {
    if (isScanning) return;
    setIsScanning(true);
    try {
      const scan = await competitorApi.createScan({
        targets: ["小鹅通", "有赞教育", "企微管家"],
        focus: "价格、案例、招聘和 AI 功能"
      });
      setLatestScan(scan);
    } catch {
      // Preserve current competitor snapshot; centralized error UI can be added later.
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
          <button className="module-primary-action" disabled={isScanning} onClick={() => void startScan()} type="button">
            {isScanning ? "采集中..." : "启动采集任务"}
          </button>
        </div>
        {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}

        <section className="module-overview-card competitor-data-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">智能客服行业竞品扫描</span>
            <h2>把分散的公开信号合成一张可行动的竞品地图</h2>
            <p>系统会从官网、招聘、内容、价格、案例和投放素材中提取变化，识别竞品正在抢什么客户、推什么能力、用什么话术。</p>
            <div className="module-stat-strip">
              {dataStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact competitor-data-input">
            <label htmlFor="competitor-target">输入竞品或关键词</label>
            <textarea
              id="competitor-target"
              aria-label="输入竞品或关键词"
              placeholder="例如：小鹅通、有赞教育、企微管家；重点关注价格、案例、招聘和 AI 功能..."
            />
            <button type="button">生成采集计划</button>
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
            {visibleCompetitors.map((item) => (
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
                  <Link to="/competitor-monitoring">加入监测</Link>
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
              {visibleConclusions.map(([title, detail]) => (
                <article key={title}>
                  <strong>{title}</strong>
                  <p>{detail}</p>
                </article>
              ))}
            </div>
          </div>

          <aside className="competitor-action-card" aria-label="建议动作">
            <h2>建议动作</h2>
            <strong>先补案例页，再打价格差异</strong>
            <p>竞品正在强化 AI 私域增长叙事。建议用 2 个真实案例证明“课程增长 + 企微转化”闭环，同时生成一份对比型销售材料。</p>
            <Link to="/tasks">生成反击任务</Link>
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default CompetitorDataPage;
