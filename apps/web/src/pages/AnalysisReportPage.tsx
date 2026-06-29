import { type CSSProperties, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { analysisApi, type AnalysisActionItem, type AnalysisSession, type DirectionCard } from "../lib/analysisApi";
import { CdkTopNav } from "./AnalysisPage";

const reportSections = [
  ["为什么适合你", ["你具备内容与营销理解，擅长信息整合与表达。", "熟悉中小企业获客痛点，能快速切入落地场景。", "有内容输出经验与资源，具备产品化与商业化潜力。"], "robot"],
  ["市场机会", ["中小企业内容生产外包需求强，年增速 25%+。", "AI 降本提效明显，付费转化路径清晰。", "从内容生成切入，易与获客、转化、数据分析形成闭环。"], "chart"],
  ["0 → 1 起步建议", ["调研 20 家目标客户的内容需求与付费意愿。", "用 Notion + AI 工具做出 MVP Demo 验证。", "锁定 1-2 个细分场景进行试点，迭代产品与交付 SOP。"], "rocket"],
  ["推荐对标公司 / 产品", [], "benchmarks"],
  ["风险提醒", ["同质化竞争激烈，需做出差异化定位与模板壁垒。", "客户对效果敏感，需建立可衡量的交付标准与案例验证。"], "risk"]
] as const;

const planItems = [
  ["Day 1", "明确最小目标客户与核心痛点", "输出用户画像与场景清单"],
  ["Day 2", "验证需求与付费意愿", "访谈 5-10 位目标用户，收集反馈"],
  ["Day 3", "提供最小可行方案（MVP）", "用现有资源做出演示版本"],
  ["Day 4", "小范围试点与迭代", "服务 3-5 位用户，收集改进点"],
  ["Day 5", "打磨产品/服务与定价", "确定核心服务包与定价策略"],
  ["Day 6", "打造信任与案例", "输出案例与用户见证，建立信任"],
  ["Day 7", "决策与放大", "复盘数据，决定是否规模投入"]
] as const;

const relatedReports = [
  ["行业垂直知识付费社区", 84],
  ["本地生活数字化服务", 78]
] as const;

function AnalysisReportPage() {
  const { sessionId } = useParams();
  const [session, setSession] = useState<AnalysisSession | null>(null);
  const [actionItems, setActionItems] = useState<AnalysisActionItem[]>([]);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");

  useEffect(() => {
    const id = Number(sessionId);
    if (!Number.isFinite(id) || id <= 0) {
      setStatus("error");
      return;
    }

    let active = true;
    analysisApi
      .getSession(id)
      .then(async (payload) => {
        const actionPayload = payload.status === "completed" ? await analysisApi.listActionItems(id) : { items: [] };
        if (!active) return;
        setSession(payload);
        setActionItems(actionPayload.items);
        setStatus("ready");
      })
      .catch(() => {
        if (active) setStatus("error");
      });

    return () => {
      active = false;
    };
  }, [sessionId]);

  async function toggleActionItem(item: AnalysisActionItem) {
    const completed = !item.completed;
    setActionItems((items) => items.map((next) => (next.id === item.id ? { ...next, completed } : next)));
    try {
      const updated = await analysisApi.updateActionItem(item.session_id, item.id, completed);
      setActionItems((items) => items.map((next) => (next.id === updated.id ? updated : next)));
    } catch {
      setActionItems((items) => items.map((next) => (next.id === item.id ? item : next)));
    }
  }

  const cards = session?.result?.cards ?? [];
  const primary = cards[0];

  return (
    <main className="cdk-analysis-page cdk-report-page">
      <CdkTopNav />

      {status === "loading" && (
        <section className="cdk-report-state">
          <h1>正在加载分析报告</h1>
          <p>系统正在读取本次分析结果。</p>
        </section>
      )}

      {status === "error" && (
        <section className="cdk-report-state">
          <h1>暂时无法读取分析报告</h1>
          <p>请返回历史列表重新打开，或新建一次分析。</p>
          <Link to="/analysis">新建分析</Link>
        </section>
      )}

      {status === "ready" && session?.status === "needs_input" && (
        <section className="cdk-report-state cdk-report-needs-input">
          <Link to="/analysis/history">← 返回报告列表</Link>
          <h1>需要补充信息</h1>
          <p>这次分析还没有生成完整报告，先回答以下问题会更准确。</p>
          <div>
            {session.questions?.map((question, index) => (
              <section key={question.key}>
                <b>{index + 1}</b>
                <strong>{question.text}</strong>
                <small>{question.options.join(" / ")}</small>
              </section>
            ))}
          </div>
        </section>
      )}

      {status === "ready" && primary && (
        <>
          <section className="cdk-report-hero">
            <div>
              <Link to="/analysis/history">← 返回报告列表</Link>
              <h1>{primary.name}</h1>
              <span>{matchLabel(0)}</span>
              <p>基于你的资源与目标，AI 判断该方向与你的经验、能力与市场机会高度匹配。</p>
            </div>
            <div className="cdk-report-actions">
              <button type="button">下载报告（PDF）›</button>
              <button type="button">分享报告</button>
            </div>
          </section>

          <section className="cdk-report-metrics">
            <div className="score">
              <i style={{ "--score": `${primary.score}%` } as CSSProperties} />
              <strong>{primary.score}<small>/100</small></strong>
              <span>匹配度</span>
            </div>
            <div><i className="building" /><strong>{primary.difficulty.level}等</strong><span>启动难度<br />需验证产品与渠道</span></div>
            <div><i className="people" /><strong>{primary.benchmarks.length}</strong><span>推荐对标<br />可学习与对标的公司</span></div>
          </section>

          <section className="cdk-report-layout">
            <div className="cdk-report-main">
              {reportSections.map(([title, bullets, icon], index) => (
                <article className="cdk-report-section" key={title}>
                  <header>
                    <b>{index + 1}</b>
                    <h2>{title}</h2>
                  </header>
                  {icon === "benchmarks" ? (
                    <div className="cdk-benchmark-grid">
                      {benchmarkCards(primary).map((benchmark) => (
                        <section key={benchmark.name}>
                          <i>{benchmark.logo}</i>
                          <strong>{benchmark.name}<span>国际</span></strong>
                          <p>{benchmark.description}</p>
                          <small>官网：{benchmark.site}</small>
                        </section>
                      ))}
                    </div>
                  ) : (
                    <ul>
                      {(index === 0 ? primary.reasons : index === 1 ? [primary.market_evidence, ...primary.difficulty.notes] : bullets)
                        .slice(0, 3)
                        .map((item) => <li key={item}>{item}</li>)}
                    </ul>
                  )}
                  {index === 1 && (
                    <div className="cdk-market-score">
                      <span>市场机会评分</span>
                      <i /><i /><i /><i /><i /><i />
                      <strong>高</strong>
                    </div>
                  )}
                  <em className={`cdk-report-art ${icon}`} aria-hidden="true" />
                </article>
              ))}

              <div className="cdk-report-ctas">
                <section className="vip">
                  <i>♛</i>
                  <strong>升级 VIP，获取精准线索与成交工具</strong>
                  <p>获取目标客户线索、话术模板、成交 SOP 与自动化跟进工具包。</p>
                  <Link to="/membership/upgrade">升级 VIP 获取 ›</Link>
                </section>
                <section>
                  <i>□</i>
                  <strong>加入社群，领取完整创业分析模板</strong>
                  <p>含：市场分析模板、用户访谈清单、MVP 画布、商业计划书框架等 12 份实用资料。</p>
                  <Link to="/community">加入社群领取 ›</Link>
                </section>
              </div>
            </div>

            <aside className="cdk-report-sidebar">
              <section className="cdk-action-timeline">
                <h2>本报告下一步行动</h2>
                {(actionItems.length > 0 ? actionItems : fallbackActionItems()).map((item) => (
                  <label className={item.completed ? "completed" : ""} key={`${item.day_index}-${item.title}`}>
                    <input
                      aria-label={item.title}
                      checked={item.completed}
                      disabled={item.id === 0}
                      onChange={() => {
                        if (item.id !== 0) void toggleActionItem(item);
                      }}
                      type="checkbox"
                    />
                    <span>Day {item.day_index}</span>
                    <strong>{item.title}</strong>
                    <small>{item.detail}</small>
                  </label>
                ))}
              </section>

              <section className="cdk-ai-follow">
                <h2>AI 追问补充</h2>
                <p>完善以下信息，AI 将生成更精准的建议。</p>
                <div>
                  <button type="button">你的目标客户是谁？</button>
                  <button type="button">你更擅长内容还是销售？</button>
                  <button type="button">你希望快速验证还是重交付？</button>
                </div>
                <Link to="/analysis">更多追问问题</Link>
              </section>

              <section className="cdk-related-reports">
                <h2>相关报告</h2>
                {relatedReports.map(([name, score], index) => (
                  <Link key={name} to="/analysis/history">
                    <b>{index + 2}</b>
                    <strong>{name}</strong>
                    <span>匹配度 {score}</span>
                  </Link>
                ))}
              </section>
            </aside>
          </section>
        </>
      )}
    </main>
  );
}

function fallbackActionItems(): AnalysisActionItem[] {
  return planItems.map(([day, title, detail], index) => ({
    id: 0,
    user_id: 0,
    session_id: 0,
    day_index: index + 1,
    title,
    detail,
    completed: false,
    created_at: "",
    updated_at: day
  }));
}

function matchLabel(index: number) {
  if (index === 0) return "高度匹配";
  if (index === 1) return "较高匹配";
  return "可行方向";
}

function benchmarkCards(card: DirectionCard) {
  const names = card.benchmarks.length > 0 ? card.benchmarks : ["Jasper", "Copy.ai"];
  return names.slice(0, 2).map((name, index) => ({
    name,
    logo: index === 0 ? "J" : "c",
    description: index === 0 ? "AI 文案生成与品牌内容平台，专注长文与营销场景。" : "面向中小企业的 AI 文案工具，轻量易用。",
    site: index === 0 ? "jasper.ai" : "copy.ai"
  }));
}

export default AnalysisReportPage;
