import { FormEvent, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { projectsApi, type ProjectCase, type ProjectMatch, type ProjectMatchResult, type ProjectMatchSession, type ProjectOpportunity } from "../lib/projectsApi";

type ProjectMarketVariant =
  | "home"
  | "match"
  | "explore"
  | "cases"
  | "questions"
  | "results"
  | "history"
  | "paywall"
  | "detail"
  | "compare"
  | "export";

type ProjectsPageProps = {
  variant?: ProjectMarketVariant;
};

type DisplayProject = {
  rank: string;
  title: string;
  score: string;
  tags: readonly string[];
  budget: string;
  reasons: readonly string[];
  risk: string;
};

type MatchHistoryRow = {
  title: string;
  count: string;
  detail: string;
  href: string;
};

const opportunityBadges = [
  ["🚀", "高潜力机会", "优质赛道机会"],
  ["🌊", "低竞争蓝海", "赛道竞争切入"],
  ["🪙", "小成本启动", "低成本低风险"],
  ["⚡", "近期爆发", "趋势上升赛道"],
  ["👤", "一人公司", "轻量高效模式"],
  ["⚠", "失败教训", "避坑少走弯路"]
] as const;

const coreEntries = [
  ["AI匹配", "智活 Copilot 根据你的目标、预算、能力与资源推荐项目。", "/projects/match", "去匹配", "robot"],
  ["机会探索", "浏览赛道机会与趋势方向。", "/projects/explore", "查看机会", "lens"],
  ["真实案例库", "看成功/失败案例与拆解。", "/projects/cases", "看案例", "case"]
] as const;

const opportunities = [
  ["精品咖啡连锁品牌", "打造社区精品咖啡连锁品牌", "轻资产", "可复制", "低竞争"],
  ["功效护肤品电商", "专注科学功效护肤的 DTC 品牌", "轻资产", "SaaS", "可复制"],
  ["智能健身房", "AI+硬件驱动的智能健身新模式", "轻资产", "可复制", "低竞争"],
  ["AI短视频创作工具", "一键生成爆款短视频内容", "SaaS", "可复制", "低竞争"]
] as const;

const caseCards = [
  ["AI视频脚本工作室", "3人团队 · 6周跑通", "首月收入 ¥42,000", "从本地教培行业切入，用行业脚本包提升交付效率。"],
  ["社区咖啡小店", "夫妻档 · 90天复盘", "复购率 38%", "放弃大店模型，转向社区会员和团购预售。"],
  ["Excel自动化顾问", "一人公司 · B端服务", "客单价 ¥8,800", "用诊断模板筛选客户，减少低价值定制需求。"],
  ["知识付费训练营", "内容博主 · 2期迭代", "完课率 72%", "从大课改为 14 天任务制，提升交付和续费。"]
] as const;

const diagnosisMetrics = [
  ["能力匹配", "88", "你具备内容结构化、脚本表达和客户沟通基础。"],
  ["资金压力", "92", "启动预算可控，前期主要投入工具和样板制作。"],
  ["时间适配", "76", "每天 1-2 小时可启动，但需要固定交付节奏。"],
  ["获客基础", "68", "需要先验证垂直行业和首批样板客户。"]
] as const;

const pathSteps = [
  ["第 1 周", "做出 3 个垂直行业样板", "选教培、本地商家、知识博主各做 1 套脚本样板。"],
  ["第 2 周", "验证报价和交付边界", "用低价试单换取反馈，沉淀需求清单和标准报价。"],
  ["第 3-4 周", "复制获客话术", "围绕样板案例做私信、社群和短视频获客。"],
  ["第 5-8 周", "标准化交付包", "形成脚本包、选题库、复盘表和续费方案。"]
] as const;

const dataRows = [
  ["需求趋势", "+23.8%", "短视频脚本、口播稿、直播切片需求持续上升。"],
  ["启动成本", "¥2,180", "基础 AI 工具、剪辑模板和素材库即可起步。"],
  ["平均客单", "¥286", "单条脚本低客单，脚本包和月度服务提升收入。"],
  ["风险指数", "4.8/10", "竞争多，但垂直行业定位可显著降低同质化。"]
] as const;

const swotItems = [
  ["优势", "启动快、成本低、交付周期短，适合用样板快速验证。"],
  ["劣势", "同质化竞争明显，纯脚本服务容易陷入低价。"],
  ["机会", "商家短视频常态化，很多团队缺稳定内容供给。"],
  ["威胁", "平台规则和工具能力变化快，需要持续更新方法。"]
] as const;

const learningItems = [
  ["先小后大", "从一个细分行业开始，跑通报价、交付和复购。"],
  ["样板先行", "没有样板前不做大范围投放，先让客户看得见结果。"],
  ["交付产品化", "把脚本拆成选题、结构、钩子、口播、复盘五个标准件。"],
  ["复购设计", "用月度选题包、账号复盘和热点响应提升续费。"]
] as const;

const avoidItems = [
  ["不要一开始做全行业", "行业太散会导致案例、话术和报价都无法沉淀。"],
  ["不要只卖单条脚本", "单条脚本容易低价竞争，应尽快升级为脚本包或月服务。"],
  ["不要忽视版权和素材来源", "交付中要明确素材、案例和客户隐私边界。"],
  ["不要无节奏接单", "必须设置交付周期、修改次数和验收口径。"]
] as const;

const matchFactors = [
  ["预算有限", "预算范围 ≤ 3万"],
  ["内容方向", "未明确"],
  ["可投入时间", "未明确"],
  ["偏好标签", "未明确"]
] as const;

const resultProjects = [
  {
    rank: "1",
    title: "AI短视频脚本工作室",
    score: "94分",
    tags: ["内容创作", "低成本启动", "高需求"],
    budget: "¥1,000 - ¥3,000",
    reasons: ["你擅长内容创作，脚本写作能力突出", "市场需求旺盛，变现路径清晰", "启动成本低，适合快速验证"],
    risk: "同质化竞争较多，需打造差异化定位"
  },
  {
    rank: "2",
    title: "知识付费课程制作",
    score: "87分",
    tags: ["知识变现", "可复用交付", "高毛利"],
    budget: "¥2,000 - ¥5,000",
    reasons: ["你具备内容结构化与表达优势", "知识付费市场持续增长", "复用性强，边际成本低"],
    risk: "需要持续迭代与内容沉淀"
  },
  {
    rank: "3",
    title: "Excel自动化报表定制",
    score: "82分",
    tags: ["办公效率", "刚需服务", "稳定复购"],
    budget: "¥1,500 - ¥4,000",
    reasons: ["你熟悉 Excel，自动化能力突出", "企业降本增效需求强烈", "交付标准化，复购与转介绍机会多"],
    risk: "需注意交付效率与需求管理"
  },
  {
    rank: "4",
    title: "AI智能简历优化服务",
    score: "78分",
    tags: ["轻服务", "低门槛", "高潜需求"],
    budget: "¥800 - ¥2,000",
    reasons: ["你理解表达与逻辑，易输出高质量内容", "目标客户明确，付费意愿强", "可快速启动并通过口碑扩散"],
    risk: "需关注隐私合规与服务质量"
  }
] as const;

const questionRows = [
  ["01", "你更偏好服务型还是产品型？", "这将影响项目类别与盈利模式的匹配", ["服务型", "产品型", "都可以"]],
  ["02", "你希望多久看到第一笔收入？", "不同的变现周期，对应不同的项目类型", ["1个月内", "1-3个月", "3个月以上"]],
  ["03", "你是否接受出镜或打造个人IP？", "这将影响内容类项目的推荐方向", ["可以", "尽量不出镜", "无所谓"]],
  ["04", "更偏好线上项目还是本地项目？", "项目交付与获客方式会有所不同", ["纯线上", "本地服务", "都可以"]]
] as const;

const historyItems = [
  ["线上轻资产项目", "12 个匹配机会", "预算 1-3 万 · 一人公司"],
  ["内容创作方向", "9 个匹配机会", "内容优势 · 1-2 小时/天"],
  ["本地服务项目", "15 个匹配机会", "低成本试跑 · 可复制"]
] as const;

function toDisplayProject(project: ProjectMatch): DisplayProject {
  return {
    rank: String(project.rank),
    title: project.title,
    score: `${project.score}分`,
    tags: project.tags,
    budget: project.budget,
    reasons: project.reasons,
    risk: project.risk
  };
}

function toHistoryRow(session: ProjectMatchSession): MatchHistoryRow {
  const projects = session.result?.projects ?? [];
  const firstProject = projects[0];
  return {
    title: firstProject?.title ?? session.intent,
    count: `${projects.length || 0} 个匹配机会`,
    detail: session.intent || session.status,
    href: `/projects/matches/${session.id}`
  };
}

function ProjectsPage({ variant = "home" }: ProjectsPageProps) {
  return (
    <V4PageShell className="project-market-shell" showCopilotMini={false}>
      <section className="project-market-page" aria-label="项目超市">
        <div className="project-market-layout">
          <main className="project-market-main">
            {variant === "home" && <MarketHome />}
            {variant === "match" && <MatchRequest />}
            {variant === "explore" && <OpportunityExplore />}
            {variant === "cases" && <CaseLibrary />}
            {variant === "questions" && <MatchQuestions />}
            {variant === "results" && <MatchResults />}
            {variant === "history" && <MatchHistory />}
            {variant === "paywall" && (
              <>
                <MatchResults />
                <PaywallOverlay />
              </>
            )}
            {variant === "detail" && <ProjectDetail />}
            {variant === "compare" && <ProjectCompare />}
            {variant === "export" && (
              <>
                <ProjectCompare />
                <ExportOverlay />
              </>
            )}
          </main>
          <ProjectCopilot variant={variant} />
        </div>
      </section>
    </V4PageShell>
  );
}

function MarketHome() {
  const [featured, setFeatured] = useState<ProjectOpportunity[]>([]);
  const [featuredError, setFeaturedError] = useState("");

  useEffect(() => {
    let active = true;
    projectsApi.listOpportunities().then((payload) => {
      if (active) { setFeatured(payload.opportunities.slice(0, 4)); setFeaturedError(""); }
    }).catch((loadError) => {
      if (active) { setFeatured([]); setFeaturedError(apiErrorMessage(loadError, "暂时无法读取精选机会")); }
    });
    return () => { active = false; };
  }, []);

  return (
    <>
      <section className="pm-hero home">
        <div>
          <h1>项目超市</h1>
          <h2>发现下一个可落地机会</h2>
          <p>从真实案例、赛道数据、失败教训和增长路径中筛出适合你的项目。</p>
          <div className="pm-search">
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索项目名称、行业、关键词" placeholder="搜索项目名称、行业、关键词" />
            <button type="button" aria-label="搜索">⌕</button>
          </div>
        </div>
        <AiCubeArt />
      </section>

      <section className="pm-badge-strip" aria-label="项目机会标签">
        {opportunityBadges.map(([icon, title, detail]) => (
          <article key={title}>
            <span>{icon}</span>
            <strong>{title}</strong>
            <small>{detail}</small>
          </article>
        ))}
      </section>

      <section className="pm-core-section">
        <h2>核心入口</h2>
        <div className="pm-core-grid">
          {coreEntries.map(([title, detail, href, action, kind]) => (
            <article className={`pm-core-card ${kind}`} key={title}>
              <div>
                <h3>{title}</h3>
                <p>{detail}</p>
              </div>
              <Link to={href}>{action}</Link>
            </article>
          ))}
        </div>
      </section>

      <section className="pm-opportunity-section">
        <div className="pm-section-head">
          <h2>精选机会</h2>
          <Link to="/projects/results">查看全部</Link>
        </div>
        <div className="pm-opportunity-grid">
          {featuredError ? <p className="form-error" role="alert">{featuredError}</p> : null}
          {!featuredError && featured.length === 0 ? <div className="module-empty-state" role="status">暂无精选机会</div> : null}
          {featured.map((item) => (
            <article key={item.id}>
              <div className="pm-thumb" />
              <h3>{item.title}</h3>
              <p>{item.summary}</p>
              <div>
                {item.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <Link to={`/projects/opportunities/${item.slug}`}>查看机会</Link>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

function MatchRequest() {
  const [intent, setIntent] = useState("");
  const [result, setResult] = useState<ProjectMatchResult | null>(null);
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!intent.trim() || status === "submitting") return;
    setStatus("submitting");
    setError("");
    try {
      const next = await projectsApi.createMatch({ intent });
      setResult(next);
    } catch (error) {
      setError(apiErrorMessage(error, "暂时无法生成项目匹配，请稍后重试"));
    } finally {
      setStatus("idle");
    }
  }

  const matchedProjects = result?.projects?.map(toDisplayProject) ?? [];

  return (
    <>
      <ProjectHero title="AI匹配" subtitle="让智活 Copilot 根据你的目标、资源与偏好，帮你筛出最适合的项目机会" action="匹配历史" href="/projects/history" />
      <form onSubmit={submit}>
        <section className="pm-panel pm-input-panel">
        <h2>告诉我你的目标、资源与偏好</h2>
        <textarea
          aria-label="项目匹配需求"
          onChange={(event) => setIntent(event.target.value)}
          placeholder="例如：我想找适合一个人做的线上项目，预算3万以内，有1-2小时/天时间，希望尽快见到收入..."
          value={intent}
        />
        <div className="pm-input-tools">
          <span>参考案例</span>
          <span>上传资料</span>
          <span>语音输入</span>
          <Link to="/projects/questions">→</Link>
        </div>
        </section>
        <section className="pm-panel pm-recognized">
          <div className="pm-section-head">
            <h2>已识别的信息（示例）</h2>
            <button onClick={() => setIntent("")} type="button">清空重填</button>
          </div>
          <div className="pm-factor-grid">
            {matchFactors.map(([title, detail]) => (
              <article key={title}>
                <span aria-hidden="true" />
                <strong>{title}</strong>
                <small>{detail}</small>
              </article>
            ))}
          </div>
          <button className="pm-primary-button" disabled={!intent.trim() || status === "submitting"} type="submit">
            {status === "submitting" ? "匹配中..." : "提交给 AI 分析"}
          </button>
          {error && <p className="form-error" role="alert">{error}</p>}
        </section>
      </form>
      {result?.status === "needs_input" && (
        <section className="pm-panel pm-question-panel">
          <h2>Copilot 还想确认以下问题</h2>
          {result.questions?.map((question, index) => (
            <article key={question.key}>
              <b>{String(index + 1).padStart(2, "0")}</b>
              <span>
                <strong>{question.text}</strong>
                <small>补充后可以提升项目匹配准确度</small>
              </span>
              <div>
                {question.options.map((option) => <button key={option} type="button">{option}</button>)}
              </div>
            </article>
          ))}
        </section>
      )}
      {result?.status === "completed" && matchedProjects.length > 0 && (
        <MatchResults projects={matchedProjects} sessionId={result.session_id} />
      )}
      <Considerations />
    </>
  );
}

function OpportunityExplore() {
  const [items, setItems] = useState<ProjectOpportunity[]>([]);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");

  async function loadOpportunities(search = "") {
    try {
      const payload = await projectsApi.listOpportunities({ query: search.trim() || undefined });
      setItems(payload.opportunities);
      setError("");
    } catch (loadError) {
      setItems([]);
      setError(apiErrorMessage(loadError, "暂时无法读取项目机会"));
    }
  }

  useEffect(() => {
    void loadOpportunities();
  }, []);

  return (
    <>
      <ProjectHero title="机会探索" subtitle="按赛道热度、启动门槛、投入周期和个人适配度筛选项目机会" action="AI匹配" href="/projects/match" />
      <section className="pm-panel pm-explore-controls">
        <div className="pm-search wide">
          <span aria-hidden="true">⌕</span>
          <input aria-label="搜索机会赛道" onChange={(event) => setQuery(event.target.value)} placeholder="搜索行业、项目、关键词" value={query} />
          <button type="button" aria-label="搜索机会" onClick={() => void loadOpportunities(query)}>⌕</button>
        </div>
        <div className="pm-explore-tabs" aria-label="机会筛选">
          {["全部机会", "高潜力机会", "低竞争蓝海", "小成本启动", "近期爆发", "一人公司"].map((item, index) => (
            <button className={index === 0 ? "active" : ""} key={item} type="button">{item}</button>
          ))}
        </div>
      </section>
      <section className="pm-explore-layout">
        <div className="pm-explore-grid">
          {error ? <p className="form-error" role="alert">{error}</p> : null}
          {!error && items.length === 0 ? <div className="module-empty-state" role="status">暂无已发布项目机会</div> : null}
          {items.map((item, index) => (
            <article className="pm-explore-card" key={item.id}>
              <div className="pm-thumb" />
              <span>#{index + 1}</span>
              <h2>{item.title}</h2>
              <p>{item.summary}</p>
              <div className="pm-mini-tags">
                {[item.industry, item.difficulty, ...item.tags].filter(Boolean).map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <footer>
                <small>{item.budget_band ? `适合预算 ${item.budget_band}` : "预算待运营补充"}</small>
                <Link to={`/projects/opportunities/${item.slug}`}>查看机会</Link>
              </footer>
            </article>
          ))}
        </div>
        <aside className="pm-insight-panel">
          <h2>机会雷达</h2>
          <p>基于当前已发布项目目录筛选，不展示无来源热度数据。</p>
          <div className="module-empty-state">暂无机会洞察</div>
        </aside>
      </section>
    </>
  );
}

function CaseLibrary() {
  const [cases, setCases] = useState<ProjectCase[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    projectsApi.listCases().then((payload) => { if (active) { setCases(payload.cases); setError(""); } })
      .catch((loadError) => { if (active) { setCases([]); setError(apiErrorMessage(loadError, "暂时无法读取项目案例")); } });
    return () => { active = false; };
  }, []);
  return (
    <>
      <ProjectHero title="真实案例库" subtitle="用真实创业样板、失败复盘和可复制经验反推你的启动路径" action="机会探索" href="/projects/explore" />
      <section className="pm-panel pm-case-filter">
        <div className="pm-explore-tabs" aria-label="案例分类">
          {["全部案例", "成功样板", "失败教训", "低成本启动", "可复制模型", "近期更新"].map((item, index) => (
            <button className={index === 0 ? "active" : ""} key={item} type="button">{item}</button>
          ))}
        </div>
      </section>
      <section className="pm-case-layout">
        <div className="pm-case-grid">
          {error ? <p className="form-error" role="alert">{error}</p> : null}
          {!error && cases.length === 0 ? <div className="module-empty-state" role="status">暂无已发布案例</div> : null}
          {cases.map((item) => (
            <article className="pm-case-card" key={item.id}>
              <div className="pm-thumb" />
              <div>
                <small>{item.case_type === "success" ? "成功样板" : "失败复盘"}</small>
                <h2>{item.title}</h2>
                <strong>{item.outcome}</strong>
                <p>{item.summary}</p>
              </div>
              <footer>
                <span>关键动作 {item.key_actions.length} 个</span>
                <span>踩坑提醒 {item.pitfalls.length} 条</span>
                <a href={item.source_url} rel="noreferrer" target="_blank">{item.source_title}</a>
              </footer>
            </article>
          ))}
        </div>
        <aside className="pm-insight-panel">
          <h2>案例共性</h2>
          <p>共性结论需要基于已发布案例证据生成。</p>
          <div className="module-empty-state">暂无案例共性结论</div>
        </aside>
      </section>
    </>
  );
}

function MatchQuestions() {
  return (
    <>
      <ProjectHero title="AI补充提问" subtitle="为了更精准地匹配适合你的项目，Copilot 需要再确认几个关键信息" action="匹配历史" href="/projects/history" />
      <section className="pm-step-card">
        {["提交需求 已完成", "AI补充提问 进行中", "匹配结果 待生成"].map((step, index) => (
          <article className={index === 1 ? "active" : ""} key={step}>
            <b>{index + 1}</b>
            <span>{step}</span>
          </article>
        ))}
      </section>
      <section className="pm-panel pm-question-panel">
        <h2>Copilot 还想确认以下问题</h2>
        {questionRows.map(([number, title, detail, answers]) => (
          <article key={number}>
            <b>{number}</b>
            <span>
              <strong>{title}</strong>
              <small>{detail}</small>
            </span>
            <div>
              {answers.map((answer) => <button key={answer} type="button">{answer}</button>)}
            </div>
          </article>
        ))}
        <div className="pm-question-actions">
          <button className="pm-primary-button" type="button">生成匹配结果</button>
          <Link className="pm-secondary-button" to="/projects/match">返回修改基础需求</Link>
          <button type="button">稍后继续</button>
        </div>
      </section>
    </>
  );
}

function MatchResults({ projects = resultProjects, sessionId }: { projects?: readonly DisplayProject[]; sessionId?: number }) {
  const detailHref = sessionId ? `/projects/matches/${sessionId}` : "/projects/detail";

  return (
    <>
      <div className="pm-result-head">
        <div>
          <p>项目超市 / AI匹配 / 匹配结果</p>
          <h1>为你匹配到的项目机会</h1>
          <small>基于你的能力、预算、时间与目标，智活 Copilot 已完成项目适配分析</small>
        </div>
        <div>
          <Link to="/projects/match">重新匹配</Link>
          <button type="button">收藏结果</button>
          <Link to="/projects/export">导出报告</Link>
          <Link to="/projects/compare">加入对比</Link>
        </div>
      </div>
      <div className="pm-condition-strip">
        {["一人公司", "预算 1-3万", "内容创作优势", "轻资产偏好", "希望 1-3个月启动"].map((item) => <span key={item}>{item}</span>)}
      </div>
      <section className="pm-results-layout">
        <div className="pm-result-list">
          {projects.map((project) => (
            <article key={project.title} className="pm-result-card">
              <b>{project.rank}</b>
              <div className="pm-result-image" />
              <div>
                <h2>{project.title}</h2>
                <div className="pm-mini-tags">
                  {project.tags.map((tag) => <span key={tag}>{tag}</span>)}
                </div>
                <small>启动预算 {project.budget}</small>
              </div>
              <strong>{project.score}</strong>
              <ul>
                {project.reasons.map((reason) => <li key={reason}>{reason}</li>)}
              </ul>
              <div className="pm-result-actions">
                <Link to={detailHref}>查看拆解</Link>
                <button aria-label={`加入对比 ${project.title}`} type="button">加入对比</button>
                <button type="button">收藏</button>
              </div>
              <p>风险提示：{project.risk}</p>
            </article>
          ))}
        </div>
        <aside className="pm-reason-card">
          <h2>为什么推荐这些项目</h2>
          <p>基于多维评估模型，为你筛选最合适的机会。</p>
          {["能力匹配 40%", "预算匹配 25%", "资源门槛 20%", "增长潜力 15%"].map((item) => <article key={item}>{item}</article>)}
          <Link to={detailHref}>查看项目完整拆解</Link>
        </aside>
      </section>
    </>
  );
}

function MatchHistory() {
  const [sessions, setSessions] = useState<ProjectMatchSession[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    projectsApi
      .listMatches()
      .then((payload) => {
        if (active) setSessions(payload.matches);
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取匹配历史"));
      });
    return () => {
      active = false;
    };
  }, []);

  const rows = sessions.length > 0
    ? sessions.map(toHistoryRow)
    : historyItems.map(([title, count, detail]) => ({ title, count, detail, href: "/projects/results" }));

  return (
    <>
      <ProjectHero title="匹配历史与收藏" subtitle="查看过往 AI 匹配记录、收藏项目和最近浏览的项目机会" action="重新匹配" href="/projects/match" />
      <section className="pm-history-layout">
        <div className="pm-panel">
          <h2>历史匹配</h2>
          {error && <p className="form-error" role="alert">{error}</p>}
          {rows.map(({ title, count, detail, href }) => (
            <article className="pm-history-row" key={title}>
              <strong>{title}</strong>
              <span>{count}</span>
              <small>{detail}</small>
              <Link to={href}>查看结果</Link>
            </article>
          ))}
        </div>
        <div className="pm-panel">
          <h2>收藏项目</h2>
          <div className="pm-opportunity-grid saved">
            {opportunities.map(([title, detail]) => (
              <article key={title}>
                <div className="pm-thumb" />
                <h3>{title}</h3>
                <p>{detail}</p>
              </article>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}

function ProjectDetail() {
  const { matchId, opportunitySlug } = useParams();
  const [session, setSession] = useState<ProjectMatchSession | null>(null);
  const [opportunity, setOpportunity] = useState<ProjectOpportunity | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (opportunitySlug) {
      let active = true;
      projectsApi.getOpportunity(opportunitySlug).then((payload) => {
        if (active) { setOpportunity(payload); setError(""); }
      }).catch((loadError) => {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取项目详情"));
      });
      return () => { active = false; };
    }
    if (!matchId) return;
    const id = Number(matchId);
    if (!Number.isFinite(id) || id <= 0) {
      setError("匹配记录不存在");
      return;
    }
    let active = true;
    projectsApi
      .getMatch(id)
      .then((payload) => {
        if (active) {
          setSession(payload);
          setError("");
        }
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取项目详情"));
      });
    return () => {
      active = false;
    };
  }, [matchId, opportunitySlug]);

  const project = session?.result?.projects?.[0];
  const title = opportunity?.title ?? project?.title ?? "AI短视频脚本工作室";
  const subtitle = opportunity?.summary ?? (project
    ? `来自匹配需求：${session?.intent ?? "项目匹配"}`
    : "为知识博主/品牌/商家提供短视频脚本本地化制作服务");
  const conditionTags = opportunity
    ? [opportunity.industry, opportunity.budget_band, opportunity.difficulty, ...opportunity.tags].filter(Boolean)
    : project
    ? [`匹配度 ${project.score}分`, `预算 ${project.budget}`, ...project.tags.slice(0, 2)]
    : ["匹配度 94分", "预算 1-3万", "1-3个月启动", "轻资产"];

  return (
    <>
      <section className="pm-detail-hero">
        <div>
          <p>项目超市 / 匹配结果 / 项目详情</p>
          <h1>{title}</h1>
          <small>{subtitle}</small>
          <div className="pm-condition-strip">
            {conditionTags.map((item) => <span key={item}>{item}</span>)}
          </div>
        </div>
        <AiCubeArt />
      </section>
      {error && <p className="form-error" role="alert">{error}</p>}
      <nav className="pm-detail-tabs" aria-label="项目详情模块">
        {["诊断是否能做", "成功路径", "当前数据", "真实案例库", "优劣势", "可学经验", "要避免行为"].map((item) => <a key={item} href={`#${item}`}>{item}</a>)}
      </nav>
      {opportunity ? (
        <section className="pm-detail-dashboard">
          {opportunity.sections?.length ? opportunity.sections.map((section) => (
            <article className="pm-detail-section" key={section.title}>
              <h2>{section.title}</h2>
              <p>{section.body}</p>
              {section.items.length ? <ul>{section.items.map((item) => <li key={item}>{item}</li>)}</ul> : null}
            </article>
          )) : <div className="module-empty-state" role="status">暂无项目详情章节</div>}
        </section>
      ) : <DetailDashboard />}
    </>
  );
}

function DetailDashboard() {
  return (
    <section className="pm-detail-dashboard">
      <article className="pm-detail-section pm-diagnosis" id="诊断是否能做">
        <div className="pm-section-head">
          <h2>诊断是否能做</h2>
          <span>综合可做度 81分</span>
        </div>
        <p>从能力、时间、资金、获客渠道四个维度判断当前是否适合启动。</p>
        <div className="pm-score-grid">
          {diagnosisMetrics.map(([label, score, detail]) => (
            <article key={label}>
              <strong>{score}</strong>
              <span>{label}</span>
              <small>{detail}</small>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-path" id="成功路径">
        <h2>成功路径</h2>
        <p>第 1 周搭样板，第 2 周验证报价，第 3-4 周复制获客话术。</p>
        <div className="pm-path-list">
          {pathSteps.map(([week, title, detail]) => (
            <article key={week}>
              <b>{week}</b>
              <span>
                <strong>{title}</strong>
                <small>{detail}</small>
              </span>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-data-board" id="当前数据">
        <div>
          <h2>当前数据</h2>
          <p>短视频脚本代写搜索热度上升，低预算工具链成熟，适合轻量试跑。</p>
        </div>
        <div className="pm-trend-chart" aria-label="市场趋势图">
          {[34, 46, 42, 58, 63, 78, 86].map((value, index) => (
            <span key={`${value}-${index}`} style={{ height: `${value}%` }} />
          ))}
        </div>
        <div className="pm-data-grid">
          {dataRows.map(([label, value, detail]) => (
            <article key={label}>
              <strong>{value}</strong>
              <span>{label}</span>
              <small>{detail}</small>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-case-sample" id="真实案例库">
        <h2>真实案例库</h2>
        <p>个人工作室通过垂直行业脚本包切入，首月获得 12 个付费客户。</p>
        <div className="pm-case-mini-grid">
          {caseCards.slice(0, 3).map(([title, meta, result, detail]) => (
            <article key={title}>
              <div className="pm-thumb" />
              <strong>{title}</strong>
              <span>{result}</span>
              <small>{meta} · {detail}</small>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-swot" id="优劣势">
        <h2>优劣势</h2>
        <p>优势是成本低和交付快，劣势是同质化高、需要垂直定位。</p>
        <div className="pm-swot-grid">
          {swotItems.map(([title, detail]) => (
            <article key={title}>
              <strong>{title}</strong>
              <small>{detail}</small>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-learning" id="可学经验">
        <h2>可学经验</h2>
        <p>先服务一个细分行业，再把脚本模板、报价和复购机制标准化。</p>
        <div className="pm-learning-list">
          {learningItems.map(([title, detail], index) => (
            <article key={title}>
              <b>{index + 1}</b>
              <span>
                <strong>{title}</strong>
                <small>{detail}</small>
              </span>
            </article>
          ))}
        </div>
      </article>

      <article className="pm-detail-section pm-avoid" id="要避免行为">
        <h2>要避免行为</h2>
        <p>避免一开始做大而全的内容服务，避免没有样板就直接大范围投放。</p>
        <div className="pm-avoid-list">
          {avoidItems.map(([title, detail]) => (
            <article key={title}>
              <b>!</b>
              <span>
                <strong>{title}</strong>
                <small>{detail}</small>
              </span>
            </article>
          ))}
        </div>
      </article>
    </section>
  );
}

function ProjectCompare() {
  return (
    <>
      <div className="pm-result-head">
        <div>
          <p>项目超市 / 项目对比</p>
          <h1>项目对比</h1>
          <small>从预算、能力、增长潜力和风险维度横向比较候选项目。</small>
        </div>
        <button type="button">导出对比报告</button>
      </div>
      <section className="pm-compare-grid">
        {resultProjects.slice(0, 3).map((project) => (
          <article key={project.title}>
            <div className="pm-result-image" />
            <h2>{project.title}</h2>
            <strong>{project.score}</strong>
            <p>{project.risk}</p>
            <ul>
              {project.reasons.map((reason) => <li key={reason}>{reason}</li>)}
            </ul>
          </article>
        ))}
      </section>
      <section className="pm-panel pm-compare-summary">
        <h2>AI对比建议</h2>
        <p>优先选择 AI短视频脚本工作室作为第一阶段验证项目，同时保留 Excel 自动化作为 B 端稳定现金流备选。</p>
      </section>
    </>
  );
}

function ProjectHero({ title, subtitle, action, href }: { title: string; subtitle: string; action: string; href: string }) {
  return (
    <section className="pm-hero compact">
      <div>
        <p>项目超市 / {title}</p>
        <h1>{title}</h1>
        <small>{subtitle}</small>
      </div>
      <AiCubeArt />
      <Link to={href}>{action}</Link>
    </section>
  );
}

function Considerations() {
  return (
    <section className="pm-panel">
      <h2>匹配逻辑会考虑</h2>
      <div className="pm-factor-grid">
        {["能力适配", "预算门槛", "资源要求", "增长路径"].map((item) => (
          <article key={item}>
            <strong>{item}</strong>
            <small>用于评估项目可做度与启动风险</small>
          </article>
        ))}
      </div>
    </section>
  );
}

function ProjectCopilot({ variant }: { variant: ProjectMarketVariant }) {
  const resultMode = variant === "results" || variant === "paywall" || variant === "detail" || variant === "compare" || variant === "export";
  return (
    <aside className="pm-copilot" aria-label="智活 Copilot">
      <header>
        <span aria-hidden="true">✦</span>
        <div>
          <strong aria-label="项目超市 Copilot">项目超市 Copilot</strong>
          <small>你的全球 AI 助手，随时为你提供帮助</small>
        </div>
        <button type="button">⌃</button>
      </header>
      <div className="pm-chat-bubble mine">帮我找一些适合一人公司、轻资产、可快速起盘的项目</div>
      <div className="pm-chat-bubble">
        {resultMode
          ? "根据你的条件与偏好，我为你匹配了 4 个高契合度项目机会。"
          : "好的！我会基于你的偏好分析适合的机会，并推荐可快速起盘的项目。"}
      </div>
      {resultMode ? (
        <article className="pm-copilot-recommend">
          <small>TOP 1 推荐</small>
          <strong>AI短视频脚本工作室</strong>
          <p>匹配度 94 分，启动成本低，1-3 天即可验证。</p>
          <Link to="/projects/detail">查看项目完整拆解</Link>
        </article>
      ) : (
        <article className="pm-copilot-recommend">
          <small>以下是为你精选的推荐</small>
          <strong>烘焙甜品工作室</strong>
          <p>轻资产、低成本，适合小团队启动。</p>
          <Link to="/projects/match">查看 AI 匹配页</Link>
        </article>
      )}
      <nav>
        <Link to="/projects/results">分析市场机会</Link>
        <Link to="/tools/recommend">推荐工具</Link>
        <Link to="/tasks">制定落地计划</Link>
      </nav>
      <MiniCopilotForm className="pm-copilot-input" attachIcon="＋" sendIcon="↗" />
    </aside>
  );
}

function PaywallOverlay() {
  return (
    <div className="pm-modal-scrim">
      <section className="pm-paywall-modal" role="dialog" aria-label="解锁完整拆解">
        <div>
          <h2>解锁完整拆解</h2>
          <p>查看项目完整路径、当前数据、真实案例、优劣势和避坑清单。</p>
        </div>
        <aside>
          <strong>智活AI会员</strong>
          <b>¥199</b>
          <small>完整解锁项目拆解样板</small>
          <button type="button">立即解锁</button>
          <button type="button">会员权益</button>
        </aside>
      </section>
    </div>
  );
}

function ExportOverlay() {
  return (
    <div className="pm-modal-scrim">
      <section className="pm-export-modal" role="dialog" aria-label="导出匹配报告">
        <h2>导出匹配报告</h2>
        <div>
          {["PDF报告", "Excel对比表", "落地任务清单"].map((item) => <button key={item} type="button">{item}</button>)}
        </div>
        <article>
          <strong>报告内容</strong>
          <small>匹配条件、推荐项目、对比结论、启动路径与风险提示</small>
        </article>
        <footer>
          <button type="button">取消</button>
          <button type="button">确认导出</button>
        </footer>
      </section>
    </div>
  );
}

function AiCubeArt() {
  return (
    <div className="pm-ai-art" aria-hidden="true">
      <span className="pm-orbit one" />
      <span className="pm-orbit two" />
      <strong>AI</strong>
      <i className="cube one" />
      <i className="cube two" />
      <i className="cube three" />
    </div>
  );
}

export default ProjectsPage;
