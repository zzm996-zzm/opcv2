import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";

import { analysisApi, type DirectionCard, type DirectionResult } from "../lib/analysisApi";

type AnalysisMode = "direction" | "competitor";

const promptChips = [
  "我擅长内容创作，如何变现？",
  "如何验证一个小项目是否值得做？",
  "上班族下班后可以做什么副业？",
  "有哪些低门槛、高复购的方向？"
] as const;

const historyCards = [
  ["教培副业切入方向分析", "结合教培经验与小额预算，推荐高变现方向", "我有什么，适合做什么", "2024-05-18 14:32"],
  ["少儿英语项目可行性", "市场需求与竞争格局深度分析", "拆解一个对标公司", "2024-05-16 10:21"],
  ["对标某教育机构拆解", "产品模式、获客渠道与商业模式拆解", "拆解一个对标公司", "2024-05-14 16:45"],
  ["知识付费项目评估", "从定位、内容到变现的全链路评估报告", "我有什么，适合做什么", "2024-05-12 09:30"]
] as const;

const competitorSections = [
  "商业模式一句话总结",
  "获客渠道拆解",
  "爆款内容分析",
  "定价与变现方式",
  "可模仿点",
  "差异化切入建议"
] as const;

function AnalysisPage() {
  const [mode, setMode] = useState<AnalysisMode>("direction");
  const [intent, setIntent] = useState("");
  const [competitorTarget, setCompetitorTarget] = useState("");
  const [result, setResult] = useState<DirectionResult | null>(null);
  const [competitorGenerated, setCompetitorGenerated] = useState(false);
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");

    if (mode === "competitor") {
      if (!competitorTarget.trim()) return;
      setCompetitorGenerated(true);
      return;
    }

    if (!intent.trim() || status === "submitting") return;
    setStatus("submitting");
    try {
      const next = await analysisApi.startDirection({ intent });
      setResult(next);
    } catch {
      setError("暂时无法生成分析，请稍后重试");
    } finally {
      setStatus("idle");
    }
  }

  if (result?.status === "needs_input") {
    return (
      <main className="cdk-analysis-page">
        <CdkTopNav />
        <FollowUpPanel result={result} />
      </main>
    );
  }

  if (result?.status === "completed") {
    return (
      <main className="cdk-analysis-page">
        <CdkTopNav />
        <DirectionResults cards={result.cards ?? []} />
      </main>
    );
  }

  return (
    <main className="cdk-analysis-page">
      <CdkTopNav />

      <section className="cdk-analysis-hero">
        <div>
          <h1>免费分析</h1>
          <p>告诉我们的资源或目标，AI 将为你生成可落地的方向与策略建议。</p>
        </div>
        <div className="cdk-hero-illustration" aria-hidden="true">
          <span />
          <b />
          <i />
        </div>
      </section>

      <section className="cdk-analysis-card" aria-label="免费分析">
        <form onSubmit={submit}>
          <div className="cdk-analysis-tabs" role="tablist" aria-label="分析类型">
            <button
              aria-selected={mode === "direction"}
              onClick={() => setMode("direction")}
              role="tab"
              type="button"
            >
              <span aria-hidden="true">✦</span>
              我有什么，适合做什么
            </button>
            <button
              aria-selected={mode === "competitor"}
              onClick={() => setMode("competitor")}
              role="tab"
              type="button"
            >
              <span aria-hidden="true">▣</span>
              拆解一个对标公司
            </button>
          </div>

          {mode === "direction" ? (
            <div className="cdk-input-shell">
              <label className="sr-only" htmlFor="analysis-intent">描述你的资源和目标</label>
              <textarea
                id="analysis-intent"
                aria-label="描述你的资源和目标"
                maxLength={500}
                onChange={(event) => setIntent(event.target.value)}
                placeholder="我有 10 年教培经验、3 万预算，想在线上做副业，适合从什么方向切入？"
                value={intent}
              />
              <div className="cdk-input-actions">
                <span aria-hidden="true">⌁</span>
                <small>{intent.length} / 500</small>
                <button disabled={!intent.trim() || status === "submitting"} type="submit">
                  {status === "submitting" ? "生成中" : "生成分析"}
                </button>
              </div>
            </div>
          ) : (
            <div className="cdk-input-shell">
              <label className="sr-only" htmlFor="competitor-target">输入对标公司或主页链接</label>
              <textarea
                id="competitor-target"
                aria-label="输入对标公司或主页链接"
                maxLength={500}
                onChange={(event) => setCompetitorTarget(event.target.value)}
                placeholder="输入公司名、公众号、抖音主页或小红书链接，例如：小鹅通、知识付费工具..."
                value={competitorTarget}
              />
              <div className="cdk-input-actions">
                <span aria-hidden="true">⌁</span>
                <small>{competitorTarget.length} / 500</small>
                <button disabled={!competitorTarget.trim()} type="submit">生成拆解</button>
              </div>
            </div>
          )}

          <div className="cdk-prompt-row">
            <span>试试这些问题</span>
            {promptChips.map((chip) => (
              <button
                key={chip}
                onClick={() => {
                  setMode("direction");
                  setIntent(chip);
                }}
                type="button"
              >
                {chip}
              </button>
            ))}
          </div>
        </form>
      </section>

      {error && <p className="form-error cdk-analysis-error" role="alert">{error}</p>}

      {mode === "competitor" && <CompetitorOutput generated={competitorGenerated} />}

      <HistorySection />
      <TrustStrip />
    </main>
  );
}

type CdkTopNavProps = {
  active?: "首页" | "免费分析" | "VIP获客" | "会员计划" | "咨询通" | "工具箱" | "社群";
};

export function CdkTopNav({ active = "免费分析" }: CdkTopNavProps) {
  const links = [
    ["首页", "/"],
    ["免费分析", "/analysis"],
    ["VIP获客", "/leads"],
    ["会员计划", "/membership"],
    ["咨询通", "/insights"],
    ["工具箱", "/tools"],
    ["社群", "/community"]
  ] as const;

  return (
    <header className="cdk-topbar">
      <Link className="cdk-brand" to="/">
        <span className="cdk-brand-mark" aria-hidden="true" />
        智活AI · OPC
      </Link>
      <nav aria-label="CDK 顶部导航">
        {links.map(([label, href]) => (
          <Link className={label === active ? "active" : ""} key={label} to={href}>
            {label}
          </Link>
        ))}
      </nav>
      <div className="cdk-top-actions">
        <Link className="case-link" to="/projects/cases">查看案例</Link>
        <button type="button">开始体验</button>
      </div>
    </header>
  );
}

function FollowUpPanel({ result }: { result: DirectionResult }) {
  const questions = result.questions ?? [];
  const activeQuestion = questions[questions.length - 1];
  const answered = questions.slice(0, Math.max(0, questions.length - 1));

  return (
    <section className="cdk-analysis-progress" aria-label="AI 正在为你分析">
      <div className="cdk-progress-title">
        <div>
          <h2>AI 正在为你分析 ✦</h2>
          <p>AI 正在理解你的需求，追问关键信息，并生成更精准的方向报告。</p>
        </div>
        <span>预计完成时间 18 秒<br />完成后将自动生成报告</span>
      </div>

      <div className="cdk-progress-card">
        <div className="cdk-original-question">
          <strong>你的原始问题</strong>
          <p>我有 10 年教培经验、3 万预算，想在线上做副业，适合从什么方向切入？</p>
          <button type="button">查看原始输入</button>
        </div>

        <div className="cdk-stepper">
          {["读取输入", "理解需求", "AI 追问补充", "生成方向报告"].map((label, index) => (
            <div className={index < 2 ? "done" : index === 2 ? "active" : ""} key={label}>
              <b>{index < 2 ? "✓" : index + 1}</b>
              <strong>{label}</strong>
              <small>{index < 2 ? "已完成" : index === 2 ? "进行中" : "待进行"}</small>
            </div>
          ))}
          <em>72%<small>分析进度</small></em>
        </div>

        <div className="cdk-progress-dots"><span /><span /><span /><span /></div>
        <p className="cdk-progress-note">AI 正在整理问题并继续分析...</p>

        <div className="cdk-question-board">
          <div className="cdk-question-list">
            {(answered.length > 0 ? answered : questions).map((question, index) => (
              <div className={index === questions.length - 1 ? "current" : ""} key={question.key}>
                <b>🤖</b>
                <strong>AI 问题 {index + 1}</strong>
                <p>{question.text}</p>
                {question.options[0] && <span>{question.options[0]}</span>}
                {index < questions.length - 1 && <i>✓</i>}
              </div>
            ))}
          </div>

          <div className="cdk-current-question">
            <strong>AI 问题 {Math.max(1, questions.length)}（当前问题）</strong>
            <div>
              <h3>{activeQuestion?.text ?? "你更想服务哪类人群？"}</h3>
              <p>例如：教培从业者、家长、学生、中小机构等，可多选</p>
            </div>
            <label>
              你的回答
              <textarea placeholder="请输入你的回答（可多选）" maxLength={200} />
              <small>0 / 200</small>
            </label>
            <div className="cdk-answer-options">
              {(activeQuestion?.options ?? ["教培从业者", "家长", "学生", "中小机构", "其他人群"]).map((option) => (
                <button key={option} type="button">{option}</button>
              ))}
            </div>
            <button className="cdk-submit-answer" type="button">提交回答并继续分析</button>
            <p className="cdk-safe-note">你的回答仅用于本次分析，我们将严格保护你的数据安全</p>
          </div>
        </div>
      </div>

      <div className="cdk-progress-benefits">
        <div><b>?</b><strong>追问补充：让分析更准确</strong><p>AI 通过追问关键问题，补全信息盲点</p></div>
        <div><b>⌬</b><strong>动态推理：逐步完善判断</strong><p>结合你的回答，逐步优化分析逻辑</p></div>
        <div><b>▤</b><strong>自动生成：输出最终报告</strong><p>完成问答后，自动生成 3 份方向报告</p></div>
      </div>
    </section>
  );
}

function DirectionResults({ cards }: { cards: DirectionCard[] }) {
  return (
    <section className="cdk-result-preview cdk-result-overview">
      <div className="cdk-section-head cdk-result-heading">
        <div>
          <h2>你的创业分析结果</h2>
          <p>AI 已根据你的输入，生成 {cards.length} 份方向报告。请选择你最感兴趣的方向，查看详细报告。</p>
        </div>
        <span>已生成</span>
      </div>

      <div className="cdk-result-statbar">
        <div><i className="doc" /><strong>{cards.length}</strong><small>生成报告<br />份</small></div>
        <div><i className="target" /><strong>{cards.length}</strong><small>推荐方向<br />个</small></div>
        <div><i className="building" /><strong>{cards.reduce((sum, card) => sum + card.benchmarks.length, 0)}</strong><small>推荐对标公司<br />家</small></div>
        <div><i className="crown" /><strong>已排序</strong><small>建议优先级</small></div>
      </div>

      <div className="cdk-result-grid">
        {cards.map((card, index) => (
          <article className={`tone-${index + 1}`} key={card.name}>
            <header>
              <b>{index + 1}</b>
              <h3>{card.name}</h3>
              <span>{index === 0 ? "高度匹配" : index === 1 ? "较高匹配" : "可行方向"}</span>
            </header>
            <p>{card.market_evidence}</p>
            <div className="cdk-result-reasons">
              <section>
                <strong>适合理由</strong>
                <ul>{card.reasons.slice(0, 2).map((reason) => <li key={reason}>{reason}</li>)}</ul>
              </section>
              <section>
                <strong>市场机会</strong>
                <ul>{[card.market_evidence, ...card.difficulty.notes].slice(0, 2).map((reason) => <li key={reason}>{reason}</li>)}</ul>
              </section>
              <section>
                <strong>启动难度</strong>
                <ul>{card.actions.slice(0, 2).map((action) => <li key={action}>{action}</li>)}</ul>
              </section>
            </div>
            <div className="cdk-match-ring"><strong>{card.score}</strong><small>匹配度</small></div>
            <footer>
              <span>生成时间 2024-05-12</span>
              <span>推荐对标 {card.benchmarks.length} 家</span>
            </footer>
            <Link to="/analysis/history">查看详细报告</Link>
          </article>
        ))}
      </div>

      <p className="cdk-result-tip">此处为报告概览页，可先对比 {cards.length} 份方向报告。下载与分享将在单份报告详情页中进行。</p>
      <div className="cdk-result-order">
        <strong>AI 建议查看顺序<small>结合匹配度与落地难度综合推荐</small></strong>
        <span>1<small>先看高度匹配<br />机会清晰，启动更快</small></span>
        <span>2<small>再看较高匹配<br />长期价值，值得投入</small></span>
        <span>3<small>最后看可行方向<br />可作为备选或第二曲线</small></span>
      </div>
    </section>
  );
}

function CompetitorOutput({ generated }: { generated: boolean }) {
  return (
    <section className="cdk-competitor-output">
      <h2>竞品拆解输出</h2>
      <p>{generated ? "已生成前端拆解骨架，后续可接入搜索和模型数据。" : "输入公司、公众号、抖音或小红书链接后，报告会按固定结构输出。"}</p>
      <div>
        {competitorSections.map((section, index) => (
          <span key={section}>
            <b>{index + 1}</b>
            {section}
          </span>
        ))}
      </div>
    </section>
  );
}

function HistorySection() {
  return (
    <section className="cdk-history-card">
      <div className="cdk-section-head">
        <div>
          <h2>历史分析结果</h2>
          <p>共 12 条分析记录</p>
        </div>
        <Link aria-label="查看历史报告" to="/analysis/history">查看全部</Link>
      </div>
      <div className="cdk-history-grid">
        {historyCards.map(([title, desc, tag, time], index) => (
          <div className="cdk-history-item" key={title}>
            <div className={`cdk-history-thumb thumb-${index + 1}`} aria-hidden="true" />
            <div>
              <h3>{title}</h3>
              <p>{desc}</p>
              <span>{tag}</span>
            </div>
            <footer>
              <small>{time}</small>
              <i>已完成</i>
            </footer>
            <Link to="/analysis/history">继续查看 →</Link>
          </div>
        ))}
      </div>
    </section>
  );
}

function TrustStrip() {
  return (
    <section className="cdk-trust-strip">
      <div>
        <span aria-hidden="true">✓</span>
        <strong>数据安全 · 隐私保护</strong>
        <p>你的所有输入仅用于生成分析，绝不泄露或用于其他用途</p>
      </div>
      <div>
        <span className="avatar-stack" aria-hidden="true" />
        <strong>已有 12,804+ 创业者</strong>
        <p>通过免费分析找到方向</p>
      </div>
    </section>
  );
}

export default AnalysisPage;
