import { FormEvent, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import ReferenceShell from "../components/ReferenceShell";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { projectsApi, type ProjectCase, type ProjectFavorite, type ProjectMatch, type ProjectMatchResult, type ProjectMatchSession, type ProjectOpportunity } from "../lib/projectsApi";
import { tasksApi } from "../lib/tasksApi";

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


const matchFactors = [
  ["预算有限", "预算范围 ≤ 3万"],
  ["内容方向", "未明确"],
  ["可投入时间", "未明确"],
  ["偏好标签", "未明确"]
] as const;


const questionRows = [
  ["01", "你更偏好服务型还是产品型？", "这将影响项目类别与盈利模式的匹配", ["服务型", "产品型", "都可以"]],
  ["02", "你希望多久看到第一笔收入？", "不同的变现周期，对应不同的项目类型", ["1个月内", "1-3个月", "3个月以上"]],
  ["03", "你是否接受出镜或打造个人IP？", "这将影响内容类项目的推荐方向", ["可以", "尽量不出镜", "无所谓"]],
  ["04", "更偏好线上项目还是本地项目？", "项目交付与获客方式会有所不同", ["纯线上", "本地服务", "都可以"]]
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
  if (variant === "home") {
    return (
      <ReferenceShell className="ref-project-shell" mainClassName="ref-project-page">
        <div className="ref-project-layout">
          <main className="ref-project-main"><MarketHome /></main>
          <ProjectCopilot reference variant={variant} />
        </div>
      </ReferenceShell>
    );
  }

  return (
    <V4PageShell className="project-market-shell" showCopilotMini={false}>
      <section className="project-market-page" aria-label="项目超市">
        <div className="project-market-layout">
          <main className="project-market-main">
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
      <section className="ref-project-hero">
        <div className="ref-project-hero-copy">
          <h1>项目超市</h1>
          <h2>发现下一个可落地机会</h2>
          <p>从真实案例、赛道数据、失败教训和增长路径中筛出适合你的项目。</p>
          <div className="ref-project-search">
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索项目名称、行业、关键词" placeholder="搜索项目名称、行业、关键词" />
            <button type="button" aria-label="搜索">⌕</button>
          </div>
        </div>
        <img className="ref-project-hero-art" alt="" src="/project-market/home-hero.jpg" />
      </section>

      <section className="ref-project-badges" aria-label="项目机会标签">
        {opportunityBadges.map(([icon, title, detail]) => (
          <article key={title}>
            <span>{icon}</span>
            <strong>{title}</strong>
            <small>{detail}</small>
          </article>
        ))}
      </section>

      <section className="ref-project-core">
        <h2>核心入口</h2>
        <div className="ref-project-core-grid">
          {coreEntries.map(([title, detail, href, action, kind]) => (
            <article className={`ref-project-core-card ${kind}`} key={title}>
              <div>
                <h3>{title}</h3>
                <p>{detail}</p>
              </div>
              <Link to={href}>{action}</Link>
            </article>
          ))}
        </div>
      </section>

      <section className="ref-project-opportunities">
        <div className="ref-project-section-head">
          <h2>精选机会</h2>
          <Link to="/projects/results">查看全部</Link>
        </div>
        <div className="ref-project-opportunity-grid">
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
  const [answers, setAnswers] = useState<Record<string, string>>({});

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

  async function submitAnswers() {
    if (!result || result.status !== "needs_input" || status === "submitting") return;
    const payload = (result.questions ?? []).map((question) => ({ key:question.key, value:answers[question.key] ?? "" }));
    if (payload.some((answer) => !answer.value)) { setError("请回答全部补充问题"); return; }
    setStatus("submitting"); setError("");
    try { setResult(await projectsApi.answerMatch(result.session_id, payload)); } catch (submitError) { setError(apiErrorMessage(submitError, "暂时无法生成匹配结果")); } finally { setStatus("idle"); }
  }

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
                {question.options.map((option) => <button className={answers[question.key] === option ? "active" : ""} key={option} onClick={() => setAnswers((current) => ({ ...current, [question.key]:option }))} type="button">{option}</button>)}
              </div>
            </article>
          ))}
          <button className="pm-primary-button" disabled={status === "submitting"} onClick={() => void submitAnswers()} type="button">{status === "submitting" ? "生成中..." : "生成匹配结果"}</button>
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

function MatchResults({ projects, sessionId }: { projects?: readonly DisplayProject[]; sessionId?: number }) {
  const [loadedSession, setLoadedSession] = useState<ProjectMatchSession | null>(null);
  const [loading, setLoading] = useState(projects === undefined);
  const [loadError, setLoadError] = useState("");
  const [favorited, setFavorited] = useState(false);
  const [favoritePending, setFavoritePending] = useState(false);
  const persistedProjects = projects ?? loadedSession?.result?.projects?.map(toDisplayProject) ?? [];
  const persistedSessionId = sessionId ?? loadedSession?.id;
  const detailHref = persistedSessionId ? `/projects/matches/${persistedSessionId}` : "/projects/history";
  const [taskMessage, setTaskMessage] = useState("");
  const [taskError, setTaskError] = useState("");

  useEffect(() => {
    if (projects !== undefined) return;
    let active = true;
    setLoading(true);
    projectsApi.listMatches().then((payload) => {
      if (!active) return;
      const latest = payload.matches.find((item) => item.status === "completed" && (item.result?.projects?.length ?? 0) > 0) ?? null;
      setLoadedSession(latest);
      setLoadError("");
    }).catch((error) => {
      if (active) setLoadError(apiErrorMessage(error, "暂时无法读取匹配结果"));
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [projects]);

  useEffect(() => {
    if (!persistedSessionId) return;
    let active = true;
    projectsApi.listFavorites().then((payload) => {
      if (active) setFavorited(payload.favorites.some((item) => item.session_id === persistedSessionId));
    }).catch(() => {
      if (active) setFavorited(false);
    });
    return () => { active = false; };
  }, [persistedSessionId]);

  async function generateTasks() {
    if (!persistedSessionId || persistedProjects.length === 0) return;
    try {
      const result = await tasksApi.generateTasks(`验证并落地项目：${persistedProjects.map((item) => item.title).join("、")}`, { sourceType:"project_match", sourceId:persistedSessionId, sourceTitle:persistedProjects[0].title, sourceUrl:`/projects/matches/${persistedSessionId}` });
      setTaskMessage(`已创建 ${result.tasks.length} 个项目任务`); setTaskError("");
    } catch { setTaskError("生成任务失败，请稍后重试"); }
  }

  async function toggleFavorite() {
    if (!persistedSessionId || favoritePending) return;
    setFavoritePending(true);
    try {
      if (favorited) await projectsApi.unfavoriteMatch(persistedSessionId);
      else await projectsApi.favoriteMatch(persistedSessionId);
      setFavorited(!favorited);
      setLoadError("");
    } catch (error) {
      setLoadError(apiErrorMessage(error, favorited ? "暂时无法取消收藏" : "暂时无法收藏结果"));
    } finally {
      setFavoritePending(false);
    }
  }

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
          <button disabled={!persistedSessionId || favoritePending} onClick={() => void toggleFavorite()} type="button">
            {favoritePending ? "处理中..." : favorited ? "取消收藏" : "收藏结果"}
          </button>
          <Link to="/projects/export">导出报告</Link>
          <Link to="/projects/compare">加入对比</Link>
          <button disabled={!persistedSessionId} onClick={() => void generateTasks()} type="button">生成落地任务</button>
        </div>
      </div>
      {loading ? <div className="module-empty-state" role="status">正在读取最近一次匹配结果...</div> : null}
      {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
      {!loading && !loadError && persistedProjects.length === 0 ? (
        <div className="module-empty-state" role="status">暂无已完成的匹配结果，请先提交项目匹配需求。</div>
      ) : null}
      {taskMessage ? <p className="form-success" role="status">{taskMessage}</p> : null}
      {taskError ? <p className="form-error" role="alert">{taskError}</p> : null}
      {loadedSession?.intent ? <div className="pm-condition-strip"><span>匹配需求：{loadedSession.intent}</span></div> : null}
      {persistedProjects.length > 0 ? <section className="pm-results-layout">
        <div className="pm-result-list">
          {persistedProjects.map((project) => (
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
                <button disabled={!persistedSessionId || favoritePending} onClick={() => void toggleFavorite()} type="button">{favorited ? "取消收藏" : "收藏"}</button>
              </div>
              <p>风险提示：{project.risk}</p>
            </article>
          ))}
        </div>
        <aside className="pm-reason-card">
          <h2>推荐依据说明</h2>
          <p>以上匹配度、推荐理由和风险提示来自该次 AI 匹配记录，不代表已验证的市场事实。</p>
          <article>来源：匹配记录 #{persistedSessionId}</article>
          <article>口径：模型推演，需结合已发布项目资料与外部证据验证</article>
          <Link to={detailHref}>查看项目完整拆解</Link>
        </aside>
      </section> : null}
    </>
  );
}

function MatchHistory() {
  const [sessions, setSessions] = useState<ProjectMatchSession[]>([]);
  const [favorites, setFavorites] = useState<ProjectFavorite[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    Promise.all([projectsApi.listMatches(), projectsApi.listFavorites()])
      .then(([matchesPayload, favoritesPayload]) => {
        if (active) {
          setSessions(matchesPayload.matches);
          setFavorites(favoritesPayload.favorites);
          setError("");
        }
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取匹配历史"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const rows = sessions.map(toHistoryRow);

  async function removeFavorite(sessionId: number) {
    try {
      await projectsApi.unfavoriteMatch(sessionId);
      setFavorites((current) => current.filter((item) => item.session_id !== sessionId));
      setError("");
    } catch (removeError) {
      setError(apiErrorMessage(removeError, "暂时无法取消收藏"));
    }
  }

  return (
    <>
      <ProjectHero title="匹配历史与收藏" subtitle="查看过往 AI 匹配记录、收藏项目和最近浏览的项目机会" action="重新匹配" href="/projects/match" />
      <section className="pm-history-layout">
        <div className="pm-panel">
          <h2>历史匹配</h2>
          {error && <p className="form-error" role="alert">{error}</p>}
          {loading ? <div className="module-empty-state" role="status">正在读取匹配历史...</div> : null}
          {!loading && !error && rows.length === 0 ? <div className="module-empty-state" role="status">暂无匹配历史</div> : null}
          {rows.map(({ title, count, detail, href }) => (
            <article className="pm-history-row" key={href}>
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
            {!loading && !error && favorites.length === 0 ? <div className="module-empty-state" role="status">暂无收藏项目</div> : null}
            {favorites.map((favorite) => favorite.session ? (
              <article key={favorite.id}>
                <div className="pm-thumb" />
                <h3>{favorite.session.result?.projects?.[0]?.title ?? favorite.session.intent}</h3>
                <p>{favorite.session.intent}</p>
                <Link to={`/projects/matches/${favorite.session_id}`}>查看结果</Link>
                <button onClick={() => void removeFavorite(favorite.session_id)} type="button">取消收藏</button>
              </article>
            ) : null)}
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
  const [evidence, setEvidence] = useState<ProjectCase[]>([]);
  const [loading, setLoading] = useState(Boolean(matchId || opportunitySlug));
  const [error, setError] = useState("");

  useEffect(() => {
    if (opportunitySlug) {
      let active = true;
      setLoading(true);
      Promise.all([projectsApi.getOpportunity(opportunitySlug), projectsApi.listCases()]).then(([payload, casesPayload]) => {
        if (active) {
          setOpportunity(payload);
          setEvidence(casesPayload.cases);
          setError("");
        }
      }).catch((loadError) => {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取项目详情"));
      }).finally(() => {
        if (active) setLoading(false);
      });
      return () => { active = false; };
    }
    if (!matchId) {
      setLoading(false);
      return;
    }
    const id = Number(matchId);
    if (!Number.isFinite(id) || id <= 0) {
      setError("匹配记录不存在");
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
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
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [matchId, opportunitySlug]);

  const project = session?.result?.projects?.[0];
  const title = opportunity?.title ?? project?.title ?? "项目详情";
  const subtitle = opportunity?.summary ?? (project
    ? `来自匹配需求：${session?.intent ?? "项目匹配"}`
    : "请选择一条已发布项目机会或已保存的匹配记录");
  const conditionTags = opportunity
    ? [opportunity.industry, opportunity.budget_band, opportunity.difficulty, ...opportunity.tags].filter(Boolean)
    : project
    ? [`匹配度 ${project.score}分`, `预算 ${project.budget}`, ...project.tags.slice(0, 2)]
    : [];

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
      {loading ? <div className="module-empty-state" role="status">正在读取项目详情...</div> : null}
      {error && <p className="form-error" role="alert">{error}</p>}
      {!loading && !error && opportunity ? (
        <section className="pm-detail-dashboard">
          {opportunity.sections?.length ? opportunity.sections.map((section) => (
            <article className="pm-detail-section" key={section.title}>
              <h2>{section.title}</h2>
              <p>{section.body}</p>
              {section.items.length ? <ul>{section.items.map((item) => <li key={item}>{item}</li>)}</ul> : null}
            </article>
          )) : <div className="module-empty-state" role="status">暂无项目详情章节</div>}
          <article className="pm-detail-section pm-case-sample">
            <h2>来源与证据</h2>
            <p>以下内容来自运营发布的案例证据库；项目章节未引用的结论不自动视为已验证事实。</p>
            {evidence.length === 0 ? <div className="module-empty-state" role="status">暂无可追溯案例来源</div> : null}
            <div className="pm-case-mini-grid">
              {evidence.map((item) => (
                <article key={item.id}>
                  <strong>{item.title}</strong>
                  <span>{item.outcome}</span>
                  <small>{item.summary}</small>
                  {item.source_url ? <a href={item.source_url} rel="noreferrer" target="_blank">{item.source_title || "查看来源"}</a> : <small>来源链接待补充</small>}
                </article>
              ))}
            </div>
          </article>
        </section>
      ) : null}
      {!loading && !error && session ? <MatchRecordDetail session={session} /> : null}
      {!loading && !error && !opportunity && !session ? (
        <div className="module-empty-state" role="status">当前地址没有关联项目记录。<Link to="/projects/explore">浏览已发布项目机会</Link></div>
      ) : null}
    </>
  );
}

function MatchRecordDetail({ session }: { session: ProjectMatchSession }) {
  const projects = session.result?.projects ?? [];
  return (
    <section className="pm-detail-dashboard">
      <article className="pm-detail-section">
        <h2>AI 匹配记录</h2>
        <p>记录 #{session.id} · 创建于 {new Date(session.created_at).toLocaleString("zh-CN")}</p>
        <p>以下分数、理由和风险均为模型基于“{session.intent}”生成的推演，未引用外部证据，不应当作市场实测数据。</p>
      </article>
      {projects.length === 0 ? <div className="module-empty-state" role="status">此匹配记录暂无项目结果</div> : null}
      {projects.map((item) => (
        <article className="pm-detail-section" key={`${item.rank}-${item.title}`}>
          <div className="pm-section-head"><h2>{item.title}</h2><span>模型匹配度 {item.score}分</span></div>
          <p>启动预算：{item.budget}</p>
          <div className="pm-mini-tags">{item.tags.map((tag) => <span key={tag}>{tag}</span>)}</div>
          <h3>模型推荐理由</h3>
          <ul>{item.reasons.map((reason) => <li key={reason}>{reason}</li>)}</ul>
          <p>风险提示：{item.risk}</p>
        </article>
      ))}
    </section>
  );
}

function ProjectCompare() {
  const [options, setOptions] = useState<ProjectOpportunity[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [comparison, setComparison] = useState<ProjectOpportunity[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    projectsApi.listOpportunities().then((payload) => { if (active) setOptions(payload.opportunities); })
      .catch((loadError) => { if (active) setError(apiErrorMessage(loadError, "暂时无法读取项目目录")); });
    return () => { active = false; };
  }, []);
  function toggle(slug: string) { setSelected((current) => current.includes(slug) ? current.filter((item) => item !== slug) : current.length < 4 ? [...current, slug] : current); }
  async function createComparison() { try { const result = await projectsApi.createComparison(selected); setComparison(result.items); setError(""); } catch (createError) { setError(apiErrorMessage(createError, "暂时无法创建项目对比")); } }
  return (
    <>
      <div className="pm-result-head">
        <div>
          <p>项目超市 / 项目对比</p>
          <h1>项目对比</h1>
          <small>从预算、能力、增长潜力和风险维度横向比较候选项目。</small>
        </div>
        <button disabled={selected.length < 2} onClick={() => void createComparison()} type="button">开始对比</button>
      </div>
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      {comparison.length === 0 ? <section className="pm-panel"><h2>选择 2-4 个项目</h2>{options.map((item) => <label key={item.id}><input aria-label={item.title} checked={selected.includes(item.slug)} onChange={() => toggle(item.slug)} type="checkbox" />{item.title}</label>)}</section> : null}
      <section className="pm-compare-grid">
        {comparison.map((project) => (
          <article key={project.id}>
            <div className="pm-result-image" />
            <h2>{project.title}</h2>
            <strong>{project.budget_band || "预算待补充"}</strong>
            <p>{project.summary}</p>
            <ul>
              {[project.industry, project.difficulty, ...project.resource_requirements].filter(Boolean).map((reason) => <li key={reason}>{reason}</li>)}
            </ul>
          </article>
        ))}
      </section>
      {comparison.length > 0 ? <section className="pm-panel pm-compare-summary"><h2>对比说明</h2><p>以上内容来自已发布项目目录快照，不包含未经验证的收益预测。</p></section> : null}
    </>
  );
}

function ProjectHero({ title, subtitle, action, href }: { title: string; subtitle: string; action: string; href: string }) {
  const heroKind = title === "机会探索"
    ? "explore"
    : title === "真实案例库"
      ? "cases"
      : title === "匹配历史与收藏"
        ? "history"
        : "match";

  return (
    <section className={`pm-hero compact ${heroKind}`}>
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

function ProjectCopilot({ reference = false, variant }: { reference?: boolean; variant: ProjectMarketVariant }) {
  const [collapsed, setCollapsed] = useState(false);
  const hasRecordContext = variant === "results" || variant === "paywall" || variant === "detail" || variant === "compare" || variant === "export";
  const copilotClassName = reference ? "ref-project-copilot" : "pm-copilot";
  return (
    <aside className={`${copilotClassName}${collapsed ? " is-collapsed" : ""}`} aria-label="智活 Copilot">
      <header>
        <span aria-hidden="true">✦</span>
        <div>
          <strong aria-label="项目超市 Copilot">项目超市 Copilot</strong>
          <small>你的全球 AI 助手，随时为你提供帮助</small>
        </div>
        <button
          type="button"
          aria-controls="project-copilot-body"
          aria-expanded={!collapsed}
          aria-label={collapsed ? "展开项目超市 Copilot" : "收起项目超市 Copilot"}
          title={collapsed ? "展开 Copilot" : "收起 Copilot"}
          onClick={() => setCollapsed((current) => !current)}
        >
          {collapsed ? "⌄" : "⌃"}
        </button>
      </header>
      <div className="project-copilot-body" id="project-copilot-body" hidden={collapsed}>
        <div className={reference ? "ref-project-chat-note" : "pm-chat-bubble"}>
          {hasRecordContext
            ? "页面中的项目结论只来自当前持久化记录；AI 匹配内容会明确标记为模型推演。"
            : "提交真实需求后，Copilot 会创建可回看的匹配记录；未关联记录时不展示业务推荐。"}
        </div>
        <article className={reference ? "ref-project-match-card" : "pm-copilot-recommend"}>
          <small>数据透明说明</small>
          <strong>先记录，再分析</strong>
          <p>机会详情使用运营发布内容与案例来源；匹配结果使用对应的 AI 会话记录。</p>
          <Link to={hasRecordContext ? "/projects/history" : "/projects/match"}>{hasRecordContext ? "查看匹配记录" : "开始 AI 匹配"}</Link>
        </article>
        <nav>
          <Link to="/projects/results">分析市场机会</Link>
          <Link to="/tools/recommend">推荐工具</Link>
          <Link to="/tasks">制定落地计划</Link>
        </nav>
        <MiniCopilotForm className={reference ? "ref-project-copilot-input" : "pm-copilot-input"} attachIcon="＋" sendIcon="↗" />
      </div>
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
  const [sourceID, setSourceID] = useState<number | null>(null);
  const [downloadURL, setDownloadURL] = useState("");
  const [error, setError] = useState("");
  const [exporting, setExporting] = useState(false);
  useEffect(() => {
    let active = true;
    projectsApi.listMatches().then((payload) => {
      if (!active) return;
      const latest = payload.matches.find((item) => item.status === "completed");
      setSourceID(latest?.id ?? null);
      if (!latest) setError("暂无可导出的已完成匹配记录");
    }).catch((loadError) => { if (active) setError(apiErrorMessage(loadError, "暂时无法读取匹配记录")); });
    return () => { active = false; };
  }, []);
  async function createExport() {
    if (!sourceID || exporting) return;
    setExporting(true); setError("");
    try { const item = await projectsApi.createExport("match", sourceID); setDownloadURL(item.download_url); }
    catch (createError) { setError(apiErrorMessage(createError, "暂时无法导出报告")); }
    finally { setExporting(false); }
  }
  return (
    <div className="pm-modal-scrim">
      <section className="pm-export-modal" role="dialog" aria-label="导出匹配报告">
        <h2>导出匹配报告</h2>
        <div><button className="active" type="button">JSON 数据快照</button></div>
        <article>
          <strong>报告内容</strong>
          <small>匹配需求、补充问题、推荐项目与风险提示的服务端快照</small>
        </article>
        {error ? <p className="form-error" role="alert">{error}</p> : null}
        {downloadURL ? <a href={downloadURL}>下载 JSON 报告</a> : null}
        <footer>
          <button type="button">取消</button>
          <button disabled={!sourceID || exporting} onClick={() => void createExport()} type="button">{exporting ? "导出中..." : "确认导出"}</button>
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
