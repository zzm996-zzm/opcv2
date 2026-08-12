import { FormEvent, useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { createPortal } from "react-dom";
import { ChevronUp } from "lucide-react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";

import { useRegisteredCopilotPanel } from "../components/CopilotPanelVisibility";
import FloatingCopilotOrb from "../components/FloatingCopilotOrb";
import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { membershipApi, type MembershipPlanOption } from "../lib/membershipApi";
import { trackProjectEvent } from "../lib/projectAnalytics";
import { projectsApi, type EvidenceCaseDetail, type EvidenceCaseItem, type ProjectCase, type ProjectCatalogFavorite, type ProjectContentBlock, type ProjectExport, type ProjectFavorite, type ProjectMatch, type ProjectMatchFile, type ProjectMatchSession, type ProjectMatchWorkflow, type ProjectOpportunity } from "../lib/projectsApi";
import { tasksApi } from "../lib/tasksApi";

type ProjectMarketVariant =
  | "home"
  | "match"
  | "explore"
  | "cases"
  | "caseDetail"
  | "questions"
  | "results"
  | "history"
  | "paywall"
  | "detail"
  | "detailUnlock"
  | "diagnosis"
  | "compare"
  | "export";

type ProjectsPageProps = {
  variant?: ProjectMarketVariant;
};

type DisplayProject = {
  rank: string;
  opportunitySlug?: string;
  title: string;
  score: string;
  tags: readonly string[];
  budget: string;
  reasons: readonly string[];
  risk: string;
};

type MatchHistoryRow = {
  id: number;
  date: string;
  time: string;
  title: string;
  count: string;
  detail: string;
  tags: string[];
  href: string;
};

type IntentFact = {
  key: string;
  label: string;
  value: string;
  icon: string;
};

type OpportunitySection = NonNullable<ProjectOpportunity["sections"]>[number];

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

const featuredOpportunityArt = [
  "/project-market/case-01.jpg",
  "/project-market/case-02.jpg",
  "/project-market/featured-fitness.png",
  "/project-market/case-03.jpg"
] as const;

const featuredOpportunityCopy = [
  { title: "精品咖啡连锁品牌", summary: "打造社区精品咖啡连锁品牌", tags: ["轻资产", "可复制", "低竞争"] },
  { title: "功效护肤品电商", summary: "专注科学功效护肤的DTC品牌", tags: ["轻资产", "SaaS", "可复制"] },
  { title: "智能健身房", summary: "AI+硬件驱动的智能健身新模式", tags: ["轻资产", "可复制", "低竞争"] },
  { title: "AI短视频创作工具", summary: "一键生成爆款短视频内容", tags: ["SaaS", "可复制", "低竞争"] }
] as const;

const detailTabs = [
  ["path", "成功路径", "成功路径"],
  ["data", "当前数据", "当前数据"],
  ["swot", "优劣势", "优劣势"],
  ["learning", "可学经验", "可学经验"],
  ["avoid", "要避免的行为", "要避免的行为"]
] as const;

function saveProjectExport(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

function projectFileStatus(file: ProjectMatchFile) {
  switch (file.parse_status) {
    case "uploading":
      return "上传中";
    case "scanning":
      return "安全扫描中";
    case "parsing":
      return "解析中";
    case "ready":
      return "解析完成";
    case "failed":
      return file.error_code === "ocr_unavailable" ? "图片 OCR 暂不可用" : "解析失败";
    default:
      return "不可用";
  }
}

function idempotencyKey(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function workflowQuestions(session: ProjectMatchWorkflow) {
  return (session.questions ?? []).map((question) => ({
    key: question.id,
    field: question.field,
    text: question.question,
    options: question.options ?? []
  }));
}

function workflowToSession(session: ProjectMatchWorkflow): ProjectMatchSession {
  return {
    id: session.match_id,
    user_id: 0,
    intent: session.need ?? "",
    status: session.status === "clarifying" ? "needs_input" : "completed",
    questions: workflowQuestions(session).map((question) => ({ key: question.key, text: question.text, options: question.options })),
    result: session.generation?.result,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };
}

function projectGenerationStep(step?: string) {
  const labels: Record<string, string> = {
    queued: "等待开始",
    analyzing: "分析需求",
    retrieving_kb: "检索项目库",
    researching_web: "补充外部证据",
    merging: "整理证据",
    generating: "生成推荐",
    done: "匹配完成",
    partial: "部分完成",
    error: "生成失败",
    canceled: "已取消"
  };
  return labels[step ?? ""] ?? "正在恢复匹配进度";
}


function toDisplayProject(project: ProjectMatch): DisplayProject {
  return {
    rank: String(project.rank),
    opportunitySlug: project.opportunity_slug,
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
  const createdAt = new Date(session.created_at);
  return {
    id: session.id,
    date: Number.isNaN(createdAt.getTime()) ? "--" : createdAt.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" }).replaceAll("/", "-"),
    time: Number.isNaN(createdAt.getTime()) ? "--:--" : createdAt.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false }),
    title: firstProject?.title ?? session.intent,
    count: `${projects.length || 0} 个项目`,
    detail: [session.intent, ...(session.answers ?? []).map((answer) => answer.value)].filter(Boolean).join(" · ") || session.status,
    tags: firstProject?.tags?.slice(0, 3) ?? [],
    href: session.status === "needs_input" ? `/projects/matches/${session.id}/questions` : `/projects/matches/${session.id}/results`
  };
}

function intentFacts(intent: string): IntentFact[] {
  const normalized = intent.trim();
  const budget = normalized.match(/(?:预算|资金|本金)[^，。；;]{0,12}/)?.[0]
    ?? normalized.match(/\d+(?:\.\d+)?\s*(?:万|元)/)?.[0]
    ?? "待补充";
  const time = normalized.match(/(?:每天|每周|每日|每月)?\s*\d+(?:-\d+)?\s*(?:小时|天|周|个月)/)?.[0]
    ?? (/全职/.test(normalized) ? "全职投入" : /兼职/.test(normalized) ? "兼职投入" : "待补充");
  const scale = /一人公司|一个人|个人/.test(normalized)
    ? "一人公司"
    : /团队|合伙/.test(normalized)
      ? "小团队"
      : "待补充";
  const direction = /内容|短视频|写作|课程/.test(normalized)
    ? "内容创作"
    : /销售|获客|客户/.test(normalized)
      ? "销售获客"
      : /数据|报表|自动化/.test(normalized)
        ? "数据提效"
        : "待补充";
  const model = /线上/.test(normalized)
    ? "线上优先"
    : /本地|线下/.test(normalized)
      ? "本地服务"
      : /轻资产/.test(normalized)
        ? "轻资产"
        : "待补充";

  return [
    { key: "budget", label: "预算信息", value: budget, icon: "wallet" },
    { key: "scale", label: "团队规模", value: scale, icon: "person" },
    { key: "direction", label: "能力方向", value: direction, icon: "pen" },
    { key: "model", label: "项目偏好", value: model, icon: "cube" },
    { key: "time", label: "投入时间", value: time, icon: "clock" }
  ];
}

function ProjectsPage({ variant = "home" }: ProjectsPageProps) {
  const needsPublicConfig = variant === "results" || variant === "paywall" || variant === "detailUnlock" || variant === "export";
  const [featurePaywallEnabled, setFeaturePaywallEnabled] = useState<boolean | null>(needsPublicConfig ? null : false);

  useEffect(() => {
    if (!needsPublicConfig) {
      setFeaturePaywallEnabled(false);
      return;
    }
    let active = true;
    setFeaturePaywallEnabled(null);
    projectsApi.getPublicConfig()
      .then((config) => { if (active) setFeaturePaywallEnabled(config.feature_paywall_enabled === true); })
      .catch(() => { if (active) setFeaturePaywallEnabled(false); });
    return () => { active = false; };
  }, [needsPublicConfig]);

  useEffect(() => {
    if (variant === "home") trackProjectEvent("project_home_view", { entry_from: document.referrer || "direct" }, "home");
  }, [variant]);

  return (
    <V4PageShell className="project-market-shell" showCopilotMini={false}>
      <section className="project-market-page" aria-label="项目超市">
        <div className={`project-market-layout pm-layout-${variant}`}>
          <main className="project-market-main">
            {variant === "home" && <MarketHome />}
            {variant === "match" && <MatchRequest />}
            {variant === "explore" && <OpportunityExplore />}
            {variant === "cases" && <CaseLibrary />}
            {variant === "caseDetail" && <CaseDetail />}
            {variant === "questions" && <MatchQuestions />}
            {variant === "results" && <MatchResults paywallEnabled={featurePaywallEnabled === true} />}
            {variant === "history" && <MatchHistory />}
            {variant === "paywall" && (
              <>
                <MatchResults paywallEnabled={featurePaywallEnabled === true} />
                <PaywallOverlay paywallEnabled={featurePaywallEnabled} />
              </>
            )}
            {variant === "detail" && <ProjectDetail />}
            {variant === "detailUnlock" && (
              <>
                <ProjectDetail />
                <PaywallOverlay paywallEnabled={featurePaywallEnabled} source="detail" />
              </>
            )}
            {variant === "diagnosis" && (
              <>
                <ProjectDetail />
                <DiagnosisOverlay />
              </>
            )}
            {variant === "compare" && <ProjectCompare />}
            {variant === "export" && (
              <>
                <MatchResults paywallEnabled={featurePaywallEnabled === true} />
                <ExportOverlay paywallEnabled={featurePaywallEnabled === true} />
              </>
            )}
          </main>
          <ProjectCopilot variant={variant === "paywall" && featurePaywallEnabled !== true ? "results" : variant === "detailUnlock" ? "detail" : variant} />
        </div>
      </section>
    </V4PageShell>
  );
}

function MarketHome() {
  const [featured, setFeatured] = useState<ProjectOpportunity[]>([]);
  const [featuredLoading, setFeaturedLoading] = useState(true);
  const [query, setQuery] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    let active = true;
    projectsApi.getHome().then((payload) => {
      if (active) setFeatured(payload.featured.slice(0, 4));
    }).catch(() => {
      if (active) setFeatured([]);
    }).finally(() => {
      if (active) setFeaturedLoading(false);
    });
    return () => { active = false; };
  }, []);

  return (
    <>
      <section className="ref-project-hero">
        <div className="ref-project-hero-copy">
          <h1 className="pm-asset-copy-sr">项目超市</h1>
          <h2 className="pm-asset-copy-sr">发现下一个可落地机会</h2>
          <p className="pm-asset-copy-sr">从真实案例、赛道数据、失败教训和增长路径中筛出适合你的项目。</p>
          <form className="ref-project-search" onSubmit={(event) => { event.preventDefault(); const normalized = query.trim(); trackProjectEvent("project_banner_submit", { query: normalized, parsed_filters: {} }, "home_banner"); navigate(`/projects/explore${normalized ? `?q=${encodeURIComponent(normalized)}` : ""}`); }}>
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索项目名称、行业、关键词" onChange={(event) => setQuery(event.target.value)} placeholder="搜索项目名称、行业、关键词" value={query} />
            <button type="submit" aria-label="搜索">⌕</button>
          </form>
        </div>
        <img className="ref-project-hero-art" alt="" src="/project-market/home-hero.jpg" />
      </section>

      <section className="ref-project-badges" aria-label="项目机会标签">
        {opportunityBadges.map(([, title, detail], index) => (
          <article key={title}>
            <span className={`pm-home-badge-icon badge-${index + 1}`} aria-hidden="true" />
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
          <Link to="/projects/explore">查看全部</Link>
        </div>
        {featuredLoading ? <div aria-label="正在加载精选机会" className="ref-project-opportunity-grid pm-skeleton-grid" role="status">
          {featuredOpportunityCopy.map((display) => <article className="pm-skeleton-card" key={display.title}><span /><b /><i /><i /><em /></article>)}
        </div> : featured.length === 0 ? <div className="module-empty-state" role="status">暂无精选项目，去机会探索查看全部已发布项目</div> : <div className="ref-project-opportunity-grid">
          {featured.map((item, index) => (
            <article className={`pm-project-${item.slug}`} key={item.id}>
              <div
                className="pm-thumb"
                style={{ "--featured-art": `url(${featuredOpportunityArt[index] ?? featuredOpportunityArt[0]})` } as CSSProperties}
              />
              <h3>{item.title}</h3>
              <p>{item.summary}</p>
              <div>
                {item.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}
              </div>
              <Link onClick={() => trackProjectEvent("project_card_click", { project_id: item.id, position: index + 1, list_type: "featured" }, "home_featured")} to={`/projects/${item.slug}`}>查看机会</Link>
            </article>
          ))}
        </div>}
      </section>

    </>
  );
}

function MatchRequest() {
  const exampleIntent = "我想找适合一个人做的线上项目，预算3万以内，每天投入1-2小时，希望尽快看到第一笔收入。";
  const [intent, setIntent] = useState("");
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [files, setFiles] = useState<ProjectMatchFile[]>([]);
  const [uploadingNames, setUploadingNames] = useState<string[]>([]);
  const [error, setError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const trackedParseResults = useRef(new Set<number>());
  const navigate = useNavigate();
  const readyFiles = files.filter((file) => file.parse_status === "ready");
  const canSubmit = (Boolean(intent.trim()) || readyFiles.length > 0) && uploadingNames.length === 0;

  useEffect(() => {
    const pending = files.filter((file) => file.parse_status === "uploading" || file.parse_status === "scanning" || file.parse_status === "parsing");
    if (pending.length === 0) return;
    const timer = window.setInterval(() => {
      void Promise.all(pending.map((file) => projectsApi.getProjectMatchFile(file.id)))
        .then((updated) => {
          updated.forEach((file) => trackFileParseResult(file));
          setFiles((current) => current.map((file) => updated.find((item) => item.id === file.id) ?? file));
        })
        .catch(() => undefined);
    }, 1200);
    return () => window.clearInterval(timer);
  }, [files]);

  function trackFileParseResult(file: ProjectMatchFile) {
    if ((file.parse_status !== "ready" && file.parse_status !== "failed") || trackedParseResults.current.has(file.id)) return;
    trackedParseResults.current.add(file.id);
    trackProjectEvent("match_file_parse_result", {
      file_type: file.detected_mime || file.mime_type,
      size_bytes: file.size_bytes,
      status: file.parse_status,
      error_code: file.error_code ?? ""
    }, "match_request");
  }

  async function uploadFiles(selected: FileList | null) {
    const candidates = Array.from(selected ?? []);
    if (fileInputRef.current) fileInputRef.current.value = "";
    if (candidates.length === 0) return;
    if (files.length + candidates.length > 10) {
      setError("最多只能上传 10 个文件");
      return;
    }
    for (const candidate of candidates) {
      setUploadingNames((current) => [...current, candidate.name]);
      try {
        const uploaded = await projectsApi.uploadProjectMatchFile(candidate);
        setFiles((current) => [...current, uploaded]);
        trackProjectEvent("match_file_upload", { file_type: uploaded.detected_mime || candidate.type, size_bytes: uploaded.size_bytes }, "match_request");
        trackFileParseResult(uploaded);
        setError("");
      } catch (uploadError) {
        setError(`上传“${candidate.name}”失败：${apiErrorMessage(uploadError, "请检查文件后重试")}`);
      } finally {
        setUploadingNames((current) => current.filter((name) => name !== candidate.name));
      }
    }
  }

  async function removeFile(file: ProjectMatchFile) {
    const previous = files;
    setFiles((current) => current.filter((item) => item.id !== file.id));
    try {
      await projectsApi.deleteProjectMatchFile(file.id);
      setError("");
    } catch (removeError) {
      setFiles(previous);
      setError(`删除文件失败：${apiErrorMessage(removeError, "请稍后重试")}`);
    }
  }

  async function retryFile(file: ProjectMatchFile) {
    setFiles((current) => current.map((item) => item.id === file.id ? { ...item, parse_status: "parsing", error_code: undefined } : item));
    try {
      const updated = await projectsApi.retryProjectMatchFile(file.id);
      setFiles((current) => current.map((item) => item.id === file.id ? updated : item));
      trackFileParseResult(updated);
      setError("");
    } catch (retryError) {
      setFiles((current) => current.map((item) => item.id === file.id ? file : item));
      setError(`重试解析失败：${apiErrorMessage(retryError, "请稍后重试")}`);
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!canSubmit || status === "submitting") return;
    setStatus("submitting");
    setError("");
    try {
      const next = await projectsApi.createProjectMatch({ need: intent, file_ids: readyFiles.map((file) => file.id) }, idempotencyKey("project-match"));
      trackProjectEvent("match_start", { match_id: next.match_id }, "match_request");
      navigate(next.status === "clarifying"
        ? `/projects/matches/${next.match_id}/questions`
        : `/projects/matches/${next.match_id}/results`);
    } catch (error) {
      setError(apiErrorMessage(error, "暂时无法生成项目匹配，请稍后重试"));
    } finally {
      setStatus("idle");
    }
  }

  const recognizedFactors = [
    { title: "需求描述", detail: intent.trim() ? `已输入 ${intent.trim().length} 字` : readyFiles.length > 0 ? `已上传 ${readyFiles.length} 份资料` : "待填写", icon: "description" },
    { title: "预算信息", detail: /预算|资金|本金|\d+\s*万/.test(intent) ? "已包含" : "待补充", icon: "budget" },
    { title: "投入时间", detail: /小时|全职|兼职|每周|每天/.test(intent) ? "已包含" : "待补充", icon: "time" },
    { title: "项目偏好", detail: /线上|本地|服务|产品|一人公司|轻资产/.test(intent) ? "已包含" : "待补充", icon: "preference" }
  ] as const;

  return (
    <>
      <ProjectHero
        variant="match"
        breadcrumb={["项目超市", "AI匹配"]}
        title="AI匹配"
        subtitle="让智活 Copilot 根据你的目标、资源与偏好，帮你筛出最适合的项目机会"
        action="匹配历史"
        href="/projects/history"
      />
      <form onSubmit={submit}>
        <section className="pm-panel pm-input-panel">
          <h2>告诉我你的目标、资源与偏好</h2>
          <textarea
            aria-label="项目匹配需求"
            onChange={(event) => setIntent(event.target.value)}
            placeholder="例如：我想找适合一个人做的线上项目，预算3万以内，有1-2小时/天时间，希望尽快见到收入..."
            value={intent}
          />
          {(files.length > 0 || uploadingNames.length > 0) ? (
            <div className="pm-upload-list" aria-label="已上传资料">
              {uploadingNames.map((name) => <article key={`uploading-${name}`}><span><strong>{name}</strong><small>上传中</small></span></article>)}
              {files.map((file) => (
                <article key={file.id}>
                  <span><strong>{file.name}</strong><small>{projectFileStatus(file)}</small></span>
                  {file.parse_status === "failed" ? <button onClick={() => void retryFile(file)} type="button">重试</button> : null}
                  <button aria-label={`删除文件 ${file.name}`} onClick={() => void removeFile(file)} type="button">×</button>
                </article>
              ))}
            </div>
          ) : null}
          <div className="pm-input-tools">
            <button className="pm-input-tool example" onClick={() => setIntent(exampleIntent)} type="button">参考案例</button>
            <label className={`pm-input-tool upload${files.length + uploadingNames.length >= 10 ? " disabled" : ""}`}>
              <span>上传资料</span>
              <input
                accept=".pdf,.docx,.txt,.md,.csv,.xlsx,.pptx,.png,.jpg,.jpeg"
                aria-label="选择项目匹配资料"
                disabled={files.length + uploadingNames.length >= 10}
                multiple
                onChange={(event) => void uploadFiles(event.target.files)}
                ref={fileInputRef}
                type="file"
              />
            </label>
            <button className="pm-input-tool voice" disabled title="语音输入暂未开放" type="button">语音输入</button>
            <button aria-label="提交匹配需求" disabled={!canSubmit || status === "submitting"} type="submit">→</button>
          </div>
        </section>
        <section className="pm-panel pm-recognized">
          <div className="pm-section-head">
            <h2>已识别的信息 <small>（实时）</small></h2>
            <button onClick={() => setIntent("")} type="button">清空重填</button>
          </div>
          <div className="pm-factor-grid">
            {recognizedFactors.map((factor) => (
              <article key={factor.title}>
                <span className={factor.icon} aria-hidden="true" />
                <strong>{factor.title}</strong>
                <small>{factor.detail}</small>
              </article>
            ))}
          </div>
          <button className="pm-primary-button" disabled={!canSubmit || status === "submitting"} type="submit">
            <span aria-hidden="true">✦</span>{status === "submitting" ? "匹配中..." : "提交给 AI 分析"}
          </button>
          <small className="pm-analysis-estimate">约 20-30 秒生成结果</small>
          {error && <p className="form-error" role="alert">{error}</p>}
        </section>
      </form>
      <Considerations />
    </>
  );
}

function OpportunityExplore() {
  const [searchParams, setSearchParams] = useSearchParams();
  const submittedQuery = searchParams.get("q") ?? "";
  const [items, setItems] = useState<ProjectOpportunity[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState("");
  const [query, setQuery] = useState(submittedQuery);
  const [loading, setLoading] = useState(true);
  const [reloadToken, setReloadToken] = useState(0);
  const [favoriteSlugs, setFavoriteSlugs] = useState<string[]>([]);
  const [compareSlugs, setCompareSlugs] = useState<string[]>([]);
  const [collectionPending, setCollectionPending] = useState<string[]>([]);
  const [collectionError, setCollectionError] = useState("");
  const requestID = useRef(0);
  const track = searchParams.get("track") ?? "";
  const budget = searchParams.get("budget") ?? "";
  const difficulty = searchParams.get("difficulty") ?? "";
  const resource = searchParams.get("resource") ?? "";
  const sortMode = searchParams.get("sort") === "latest" ? "latest" : "heat";
  const activeDirection = searchParams.get("direction") ?? "高潜力机会";
  const requestedPage = Number(searchParams.get("page") ?? "1");
  const page = Number.isSafeInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1;
  const pageSize = 8;
  const directions = ["高潜力机会", "低竞争蓝海", "小成本启动", "近期爆发", "一人公司", "可复制案例"];
  const directionKeyword: Record<string, string> = {
    "小成本启动": "低成本启动",
    "一人公司": "一人公司",
    "可复制案例": "可复制"
  };
  const effectiveKeyword = submittedQuery.trim() || directionKeyword[activeDirection] || "";
  const effectiveDifficulty = difficulty || (activeDirection === "低竞争蓝海" ? "较低" : "");
  const effectiveSort = activeDirection === "近期爆发" ? "latest" : sortMode;

  const updateParams = useCallback((updates: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value) next.set(key, value);
      else next.delete(key);
    });
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    setQuery(submittedQuery);
  }, [submittedQuery]);

  useEffect(() => {
    const currentRequest = ++requestID.current;
    setLoading(true);
    projectsApi.listProjects({
      keyword: effectiveKeyword || undefined,
      track: track || undefined,
      budget: budget || undefined,
      difficulty: effectiveDifficulty || undefined,
      resource: resource || undefined,
      sort: effectiveSort,
      page,
      pageSize
    }).then((payload) => {
      if (currentRequest !== requestID.current) return;
      setItems(payload.items ?? []);
      setTotal(Number.isFinite(payload.total) ? payload.total : 0);
      setError("");
      trackProjectEvent("project_explore_view", { group: activeDirection, query: effectiveKeyword }, "explore_catalog");
      if (effectiveKeyword) trackProjectEvent("project_search", { keyword: effectiveKeyword, result_count: payload.total ?? 0 }, "explore_catalog");
    }).catch((loadError) => {
      if (currentRequest !== requestID.current) return;
      setItems([]);
      setTotal(0);
      setError(apiErrorMessage(loadError, "暂时无法读取项目机会"));
    }).finally(() => {
      if (currentRequest === requestID.current) setLoading(false);
    });
  }, [activeDirection, budget, effectiveDifficulty, effectiveKeyword, effectiveSort, page, reloadToken, resource, track]);

  useEffect(() => {
    let active = true;
    Promise.allSettled([projectsApi.listProjectFavorites(), projectsApi.listProjectCompareItems()]).then(([favoritesResult, compareResult]) => {
      if (!active) return;
      if (favoritesResult.status === "fulfilled") {
        setFavoriteSlugs((favoritesResult.value.favorites ?? []).filter((item) => item.project_id > 0).map((item) => item.slug));
      }
      if (compareResult.status === "fulfilled") {
        setCompareSlugs((compareResult.value.items ?? []).filter((item) => item.project_id > 0).map((item) => item.slug));
      }
    });
    return () => { active = false; };
  }, []);

  function submitQuery(value: string) {
    const normalized = value.trim();
    setQuery(normalized);
    if (normalized === submittedQuery) {
      setReloadToken((current) => current + 1);
      return;
    }
    updateParams({ q: normalized, page: "" });
  }

  function updateTrackedFilters(updates: Record<string, string>) {
    const filters = Object.fromEntries(Object.entries({ track, budget, difficulty, resource, sort: sortMode, direction: activeDirection, ...updates })
      .filter(([key, value]) => key !== "page" && Boolean(value)));
    trackProjectEvent("project_filter_apply", { filters }, "explore_filters");
    updateParams(updates);
  }

  function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    submitQuery(query);
  }

  async function toggleCatalogCollection(item: ProjectOpportunity, kind: "favorite" | "compare") {
    const pendingKey = `${kind}:${item.slug}`;
    if (collectionPending.includes(pendingKey)) return;
    const values = kind === "favorite" ? favoriteSlugs : compareSlugs;
    const setValues = kind === "favorite" ? setFavoriteSlugs : setCompareSlugs;
    const removing = values.includes(item.slug);
    const previous = values;
    setValues(removing ? values.filter((slug) => slug !== item.slug) : [...values, item.slug]);
    setCollectionPending((current) => [...current, pendingKey]);
    setCollectionError("");
    try {
      if (kind === "favorite") {
        if (removing) await projectsApi.unfavoriteProject(item.slug);
        else await projectsApi.favoriteProject(item.slug);
      } else if (removing) {
        await projectsApi.removeProjectCompareItem(item.slug);
      } else {
        await projectsApi.addProjectCompareItem(item.slug);
        trackProjectEvent("project_compare_add", { project_ids: [item.id] }, "explore_catalog");
      }
    } catch (collectionActionError) {
      setValues(previous);
      const operation = kind === "favorite" ? (removing ? "取消收藏" : "收藏项目") : (removing ? "移出对比" : "加入对比");
      setCollectionError(`${operation}失败：${apiErrorMessage(collectionActionError, "请稍后重试")}`);
    } finally {
      setCollectionPending((current) => current.filter((key) => key !== pendingKey));
    }
  }

  const filterOptions = useMemo(() => ({
    tracks: [...new Set([track, ...items.map((item) => item.track || item.industry || "")].filter(Boolean))],
    budgets: [...new Set([budget, ...items.map((item) => item.budget_band)].filter(Boolean))],
    difficulties: [...new Set([difficulty, ...items.map((item) => item.difficulty)].filter(Boolean))],
    resources: [...new Set([resource, ...items.flatMap((item) => item.resource_requirements)].filter(Boolean))]
  }), [budget, difficulty, items, resource, track]);
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  return (
    <>
      <section className="pm-catalog-hero explore">
        <div className="pm-catalog-hero-copy">
          <h1>机会探索</h1>
          <p>浏览真实案例、赛道数据和增长路径，发现最适合你的项目机会</p>
          <form aria-busy={loading} className="pm-catalog-search" onSubmit={handleSearch}>
            <span aria-hidden="true">⌕</span>
            <input aria-label="搜索机会赛道" onChange={(event) => setQuery(event.target.value)} placeholder="搜索项目名称、行业、关键词或痛点" value={query} />
            <button aria-label="搜索机会" disabled={loading} type="submit"><span aria-hidden="true">{loading ? "…" : "⌕"}</span></button>
          </form>
          <small>AI 智能推荐词</small>
          <div className="pm-hero-chip-row" aria-label="AI 智能推荐词">
            {["一人公司", "低成本启动", "可复制项目", "副业变现", "AI应用", "出海机会"].map((item) => <button key={item} onClick={() => submitQuery(item)} type="button">{item}</button>)}
          </div>
        </div>
        <div className="pm-catalog-hero-art" aria-hidden="true" />
        <aside className="pm-hero-advice" aria-label="AI 智能筛选建议">
          <strong><span aria-hidden="true">✦</span> AI 智能筛选建议</strong>
          <small>根据当前目录推荐</small>
          <ul>
            <li>偏好：一人公司、低预算启动</li>
            <li>优势：内容创作、AI工具应用</li>
            <li>关注：可复制、长期增长</li>
          </ul>
          <Link to="/projects/match">查看完整画像分析 →</Link>
        </aside>
      </section>

      <section className="pm-catalog-filter" aria-label="项目机会筛选">
        <label><span>行业</span><select aria-label="按行业筛选" onChange={(event) => updateTrackedFilters({ track: event.target.value, page: "" })} value={track}><option value="">全部行业</option>{filterOptions.tracks.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
        <label><span>预算区间</span><select aria-label="按预算筛选" onChange={(event) => updateTrackedFilters({ budget: event.target.value, page: "" })} value={budget}><option value="">全部预算</option>{filterOptions.budgets.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
        <label><span>难度</span><select aria-label="按难度筛选" onChange={(event) => updateTrackedFilters({ difficulty: event.target.value, page: "" })} value={difficulty}><option value="">全部难度</option>{filterOptions.difficulties.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
        <label><span>资源要求</span><select aria-label="按资源要求筛选" onChange={(event) => updateTrackedFilters({ resource: event.target.value, page: "" })} value={resource}><option value="">全部资源</option>{filterOptions.resources.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
        <div className="pm-sort-control" aria-label="项目排序">
          <span>排序</span>
          <div>
            {(["heat", "latest"] as const).map((mode) => <button className={sortMode === mode ? "active" : ""} key={mode} onClick={() => updateTrackedFilters({ sort: mode === "heat" ? "" : mode, page: "" })} type="button">{{ heat: "热度", latest: "最新" }[mode]}</button>)}
          </div>
        </div>
      </section>

      <section className="pm-direction-strip">
        <header><h2>热门探索方向</h2><button onClick={() => updateParams({ direction: "", page: "" })} type="button">查看全部方向 →</button></header>
        <div>
          {directions.map((item, index) => (
            <button aria-label={item} className={activeDirection === item ? "active" : ""} key={item} onClick={() => updateParams({ direction: item === "高潜力机会" ? "" : item, page: "" })} type="button">
              <i aria-hidden="true">{["↗", "≈", "¥", "ϟ", "◉", "▣"][index]}</i>
              <span><strong>{item}</strong><small>{["优质赛道机会", "避开红海竞争", "低成本低风险", "趋势上升赛道", "轻量高效模式", "验证可复制性"][index]}</small></span>
            </button>
          ))}
        </div>
      </section>

      <section aria-busy={loading} className="pm-explore-grid" aria-label="项目机会列表">
        {collectionError ? <p className="form-error pm-collection-error" role="alert">{collectionError}</p> : null}
        {loading ? <div className="module-empty-state" role="status">正在搜索项目机会…</div> : null}
        {!loading && error ? <div className="module-empty-state"><p className="form-error" role="alert">{error}</p><button onClick={() => setReloadToken((current) => current + 1)} type="button">重新加载</button></div> : null}
        {!loading && !error && items.length === 0 ? <div className="module-empty-state" role="status">{effectiveKeyword ? `未找到“${effectiveKeyword}”相关的已发布项目机会` : "暂无符合条件的已发布项目机会"}</div> : null}
        {!loading && !error && items.map((item) => (
          <article className={`pm-explore-card pm-project-${item.slug}`} key={item.id}>
            <div className="pm-thumb" />
            <h2>{item.title}</h2>
            <p>{item.summary}</p>
            <footer>
              <strong>{item.budget_band || "预算待补充"}</strong>
              <small>{item.difficulty || "难度待补充"}</small>
              <div className="pm-mini-tags">{item.tags.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}</div>
              <div className="pm-card-collection-actions">
                <button aria-label={favoriteSlugs.includes(item.slug) ? `取消收藏 ${item.title}` : `收藏 ${item.title}`} disabled={collectionPending.includes(`favorite:${item.slug}`)} onClick={() => void toggleCatalogCollection(item, "favorite")} type="button">{favoriteSlugs.includes(item.slug) ? "★" : "☆"}</button>
                <button aria-label={compareSlugs.includes(item.slug) ? `移出对比 ${item.title}` : `加入对比 ${item.title}`} disabled={collectionPending.includes(`compare:${item.slug}`)} onClick={() => void toggleCatalogCollection(item, "compare")} type="button">{compareSlugs.includes(item.slug) ? "已对比" : "对比"}</button>
              </div>
              <Link aria-label="查看机会" onClick={() => trackProjectEvent("project_card_click", { project_id: item.id, position: (page - 1) * pageSize + items.indexOf(item) + 1, list_type: "explore" }, "explore_catalog")} to={`/projects/${item.slug}`}>查看拆解 →</Link>
            </footer>
          </article>
        ))}
      </section>

      {!loading && !error && items.length > 0 ? <nav className="pm-pagination" aria-label="项目机会分页">
        <button aria-label="上一页" disabled={page === 1} onClick={() => updateParams({ page: String(Math.max(1, page - 1)) })} type="button">‹</button>
        {Array.from({ length: Math.min(pageCount, 5) }, (_, index) => index + 1).map((item) => <button aria-current={page === item ? "page" : undefined} className={page === item ? "active" : ""} key={item} onClick={() => updateParams({ page: item === 1 ? "" : String(item) })} type="button">{item}</button>)}
        {pageCount > 5 ? <span>… {pageCount}</span> : null}
        <button aria-label="下一页" disabled={page >= pageCount} onClick={() => updateParams({ page: String(Math.min(pageCount, page + 1)) })} type="button">›</button>
        <small>共 {total} 条</small>
      </nav> : null}
      {compareSlugs.length > 0 ? createPortal(
        <aside className="pm-persisted-compare-bar" aria-label="项目对比栏"><span>已加入对比 {compareSlugs.length}/5</span><Link to="/projects/compare">打开对比</Link></aside>,
        document.body,
      ) : null}
    </>
  );
}

function CaseLibrary() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [cases, setCases] = useState<EvidenceCaseItem[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [reloadToken, setReloadToken] = useState(0);
  const [sortMode, setSortMode] = useState<"latest" | "sources">("latest");
  const caseType = searchParams.get("type") ?? "";
  const caseIndustry = searchParams.get("industry") ?? "";
  const caseScale = searchParams.get("scale") ?? "";
  const requestedPage = Number(searchParams.get("page") ?? "1");
  const page = Number.isSafeInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1;
  const pageSize = 12;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  const updateCaseParams = useCallback((updates: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value) next.set(key, value);
      else next.delete(key);
    });
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    projectsApi.listEvidenceCases({
      caseType: caseType || undefined,
      industry: caseIndustry || undefined,
      scale: caseScale || undefined,
      page,
      pageSize
    }).then((payload) => {
      if (active) {
        setCases(payload.items ?? []);
        setTotal(Number.isFinite(payload.total) ? payload.total : 0);
        setError("");
      }
    }).catch((loadError) => {
      if (active) {
        setCases([]);
        setTotal(0);
        setError(apiErrorMessage(loadError, "暂时无法读取项目案例"));
      }
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [caseIndustry, caseScale, caseType, page, reloadToken]);

  const visibleCases = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase("zh-CN");
    const searched = normalized
      ? cases.filter((item) => `${item.title}${item.result_summary}${item.industry ?? ""}${item.scale ?? ""}`.toLocaleLowerCase("zh-CN").includes(normalized))
      : cases;
    if (sortMode === "sources") return [...searched].sort((left, right) => right.source_count - left.source_count);
    return [...searched].sort((left, right) => Date.parse(right.published_at ?? "") - Date.parse(left.published_at ?? ""));
  }, [cases, query, sortMode]);

  const caseIndustries = [...new Set([caseIndustry, ...cases.map((item) => item.industry ?? "")].filter(Boolean))];
  const caseScales = [...new Set([caseScale, ...cases.map((item) => item.scale ?? "")].filter(Boolean))];

  return (
    <>
      <section className="pm-catalog-hero cases">
        <div className="pm-catalog-hero-copy">
          <p>项目超市&nbsp;&nbsp;/&nbsp;&nbsp;真实案例库</p>
          <h1>真实案例库</h1>
          <strong>浏览成功与失败案例，学习可复制的方法，也避开常见陷阱</strong>
          <div className="pm-case-hero-controls">
            <div className="pm-catalog-search compact">
              <span aria-hidden="true">⌕</span>
              <input aria-label="搜索真实案例" onChange={(event) => setQuery(event.target.value)} placeholder="搜索案例名称、行业、关键词" value={query} />
              <button aria-label="搜索案例" type="button">⌕</button>
            </div>
            <div className="pm-case-tabs" aria-label="案例分类">
              {[
                ["全部", ""],
                ["成功案例", "success"],
                ["失败案例", "failure"],
              ].map(([label, value]) => <button className={caseType === value ? "active" : ""} key={label} onClick={() => updateCaseParams({ type: value, page: "" })} type="button">{label}</button>)}
            </div>
          </div>
        </div>
        <div className="pm-catalog-hero-art" aria-hidden="true" />
        <aside className="pm-hero-assurance" aria-label="案例来源保障">
          <strong><span aria-hidden="true">◆</span> 全部案例均标注来源，可追溯</strong>
          <ul><li>来源清晰可靠</li><li>数据真实可查</li><li>方法可复现</li></ul>
          <small>已收录案例 <b>{total.toLocaleString("zh-CN")}</b> 个</small>
        </aside>
      </section>

      <section className="pm-catalog-filter cases" aria-label="真实案例筛选">
        <label><span>行业</span><select aria-label="案例行业" onChange={(event) => updateCaseParams({ industry: event.target.value, page: "" })} value={caseIndustry}><option value="">全部行业</option>{caseIndustries.map((item) => <option key={item}>{item}</option>)}</select></label>
        <label><span>规模</span><select aria-label="案例规模" onChange={(event) => updateCaseParams({ scale: event.target.value, page: "" })} value={caseScale}><option value="">全部规模</option>{caseScales.map((item) => <option key={item}>{item}</option>)}</select></label>
        <div className="pm-sort-control" aria-label="案例排序">
          <span>排序</span>
          <div>
            {(["latest", "sources"] as const).map((mode) => <button className={sortMode === mode ? "active" : ""} key={mode} onClick={() => setSortMode(mode)} type="button">{{ latest: "最新发布", sources: "证据最多" }[mode]}</button>)}
          </div>
        </div>
      </section>

      <section className="pm-case-featured">
        <header><h2>精选案例</h2><span>全部案例均带来源记录</span></header>
        <div className="pm-case-grid">
          {loading ? <div className="module-empty-state" role="status">正在读取证据案例…</div> : null}
          {error ? <div className="module-empty-state"><p className="form-error" role="alert">{error}</p><button onClick={() => setReloadToken((current) => current + 1)} type="button">重新加载</button></div> : null}
          {!loading && !error && visibleCases.length === 0 ? <div className="module-empty-state" role="status">暂无符合条件的已发布案例</div> : null}
          {visibleCases.slice(0, 4).map((item) => (
            <article className={`pm-case-card pm-case-${item.id}`} key={item.id}>
              <div className="pm-thumb" />
              <span className="pm-case-save" aria-hidden="true">☆</span>
              <div>
                <h2>{item.title}</h2>
                <strong>{item.result_summary}</strong>
                <div className="pm-mini-tags"><span>{item.type === "success" ? "成功案例" : "失败复盘"}</span><span>{item.source_count} 条证据</span></div>
                <p>{[item.industry, item.scale].filter(Boolean).join(" · ") || "已核验证据案例"}</p>
              </div>
              <footer>
                <Link to={`/project-cases/${item.id}`}>查看案例</Link>
                {item.project_id ? <Link to={`/projects/${item.project_id}`}>查看项目</Link> : <Link to="/projects/explore">浏览项目</Link>}
                <a href={item.primary_source_url} onClick={() => trackProjectEvent("project_case_source_click", { case_id: item.id, case_type: item.type, source_id: "primary", field: "primary_source_url" }, "case_library")} rel="noreferrer" target="_blank">查看首要来源</a>
              </footer>
            </article>
          ))}
        </div>
      </section>

      <section className="pm-case-learning-grid">
        <article>
          <header><h2 aria-label="案例共性">案例共性</h2><span>来自当前筛选结果</span></header>
          {visibleCases.length === 0 ? <div className="module-empty-state">暂无已发布经验</div> : visibleCases.slice(0, 4).map((item, index) => <div key={item.id}><b aria-hidden="true">{index + 1}</b><span><strong>{item.result_summary}</strong><small>{item.source_count} 条可追溯证据</small></span></div>)}
        </article>
        <article className="failure">
          <header><h2>失败教训</h2><span>查看全部 →</span></header>
          {visibleCases.filter((item) => item.type === "fail").length === 0 ? <div className="module-empty-state">暂无失败教训</div> : visibleCases.filter((item) => item.type === "fail").slice(0, 3).map((item) => <div key={item.id}><b aria-hidden="true">!</b><span><strong>{item.result_summary}</strong><small><Link to={`/project-cases/${item.id}`}>查看证据明细</Link></small></span></div>)}
        </article>
      </section>
      {!loading && !error && cases.length > 0 ? <nav className="pm-pagination" aria-label="案例分页">
        <button aria-label="上一页" disabled={page === 1} onClick={() => updateCaseParams({ page: String(Math.max(1, page - 1)) })} type="button">‹</button>
        <span>第 {page} / {pageCount} 页</span>
        <button aria-label="下一页" disabled={page >= pageCount} onClick={() => updateCaseParams({ page: String(Math.min(pageCount, page + 1)) })} type="button">›</button>
        <small>共 {total} 个案例</small>
      </nav> : null}
    </>
  );
}

function CaseDetail() {
  const { caseRef } = useParams();
  const [caseDetail, setCaseDetail] = useState<EvidenceCaseDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!caseRef) {
      setError("案例地址无效");
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    projectsApi.getEvidenceCase(caseRef).then((payload) => {
      if (active) {
        setCaseDetail(payload);
        setError("");
        trackProjectEvent("project_case_view", { case_id: payload.id, case_type: payload.type }, "case_detail");
      }
    }).catch((loadError) => {
      if (active) setError(apiErrorMessage(loadError, "暂时无法读取案例详情"));
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [caseRef]);

  return (
    <>
      <section className="pm-catalog-hero cases">
        <div className="pm-catalog-hero-copy">
          <p>项目超市&nbsp;&nbsp;/&nbsp;&nbsp;真实案例库&nbsp;&nbsp;/&nbsp;&nbsp;案例详情</p>
          <h1>{caseDetail?.title ?? "案例详情"}</h1>
          <strong>{caseDetail?.result_summary ?? "查看经过核验的事实、分析与来源"}</strong>
          <Link to="/projects/cases">返回案例库</Link>
        </div>
        <div className="pm-catalog-hero-art" aria-hidden="true" />
      </section>
      {loading ? <div className="module-empty-state" role="status">正在读取案例详情…</div> : null}
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      {!loading && !error && caseDetail ? <section className="pm-detail-content pm-structured-content">
        <article className="pm-structured-block type-cards">
          <h2>核验事实</h2>
          <div className="pm-structured-grid" style={{ "--pm-columns": 2 } as CSSProperties}>
            {caseDetail.facts.map((fact) => <section key={fact.field}><div><strong>{fact.field}</strong><p>{fact.value}</p><small>来源编号：{fact.source_refs.join("、")}</small></div></section>)}
          </div>
        </article>
        <article className="pm-structured-block type-cards">
          <h2>分析判断</h2>
          <div className="pm-structured-grid" style={{ "--pm-columns": 2 } as CSSProperties}>
            {caseDetail.analyses.map((analysis, index) => <section key={`${analysis.point}-${index}`}><div><strong>{analysis.point}</strong>{analysis.detail ? <p>{analysis.detail}</p> : null}<small>AI 辅助分析 · 来源编号：{analysis.source_refs.join("、") || "未单独引用"}</small></div></section>)}
          </div>
        </article>
        <article className="pm-structured-block type-sources">
          <h2>来源与证据</h2>
          <div className="pm-structured-grid" style={{ "--pm-columns": 3 } as CSSProperties}>
            {caseDetail.sources.map((source) => <a href={source.url} key={source.id} onClick={() => trackProjectEvent("project_case_source_click", { case_id: caseDetail.id, case_type: caseDetail.type, source_id: source.id, field: source.claim_fields.join(",") || "source" }, "case_detail")} rel="noreferrer" target="_blank"><strong>{source.title || source.publisher || "公开来源"}</strong><small>{source.kind}{source.is_primary ? " · 首要来源" : ""}</small><span>查看出处 →</span></a>)}
          </div>
        </article>
      </section> : null}
    </>
  );
}

function MatchQuestions() {
  const { matchId } = useParams();
  const navigate = useNavigate();
  const [session, setSession] = useState<ProjectMatchSession | null>(null);
  const [workflow, setWorkflow] = useState<ProjectMatchWorkflow | null>(null);
  const [selectedAnswers, setSelectedAnswers] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const questions = session?.questions ?? [];
  const completedCount = questions.filter((question) => Boolean(selectedAnswers[question.key])).length;
  const allAnswered = questions.length > 0 && completedCount === questions.length;
  const baseFacts = session ? intentFacts(session.intent) : [];

  useEffect(() => {
    const id = Number(matchId);
    if (!Number.isFinite(id) || id <= 0) {
      setError("匹配记录不存在");
      setLoading(false);
      return;
    }
    let active = true;
    projectsApi.getProjectMatch(id).then((payload) => {
      if (!active) return;
      if (payload.status === "ready" || payload.generation?.status === "completed" || payload.generation?.status === "partial") {
        navigate(`/projects/matches/${id}/results`, { replace:true });
        return;
      }
      setWorkflow(payload);
      setSession(workflowToSession(payload));
      setError("");
    }).catch((loadError) => {
      if (active) setError(apiErrorMessage(loadError, "暂时无法读取补充问题"));
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [matchId, navigate]);

  async function submitAnswers() {
    if (!session || !allAnswered || submitting) return;
    setSubmitting(true);
    setError("");
    try {
      if (!workflow) return;
      const next = await projectsApi.answerProjectMatch(session.id, {
        revision: workflow.revision,
        answers: workflowQuestions(workflow).map((question) => ({
          question_id: question.key,
          field: question.field,
          value: selectedAnswers[question.key]
        }))
      }, idempotencyKey("project-match-answer"));
      trackProjectEvent("match_answer", { match_id: session.id, rounds: next.revision, completeness: next.completeness }, "match_questions");
      navigate(next.status === "ready"
        ? `/projects/matches/${session.id}/results`
        : `/projects/matches/${session.id}/questions`);
    } catch (submitError) {
      setError(apiErrorMessage(submitError, "暂时无法生成匹配结果"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <ProjectHero
        variant="questions"
        breadcrumb={["项目超市", "AI匹配", "AI补充提问"]}
        title="AI补充提问"
        subtitle="为了更精准地匹配适合你的项目，Copilot 需要再确认几个关键信息"
        action="匹配历史"
        href="/projects/history"
      />
      <section className="pm-step-card">
        {[
          ["提交需求", "已完成"],
          ["AI补充提问", "进行中"],
          ["匹配结果", "待生成"]
        ].map(([title, state], index) => (
          <article className={index === 0 ? "complete" : index === 1 ? "active" : ""} key={title}>
            <b>{index === 0 ? "✓" : index + 1}</b>
            <span><strong>{title}</strong><small>{state}</small></span>
          </article>
        ))}
      </section>
      {session ? (
        <section className="pm-panel pm-session-summary" aria-label="已识别的基础需求">
          <h2>已识别你的基础需求</h2>
          <div className="pm-session-facts">
            {baseFacts.map((fact) => (
              <article key={fact.key}>
                <i className={fact.icon} aria-hidden="true" />
                <span><strong>{fact.value}</strong><small>{fact.label}</small></span>
              </article>
            ))}
          </div>
        </section>
      ) : null}
      {loading ? <div className="module-empty-state" role="status">正在读取补充问题...</div> : null}
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      {!loading && !error && questions.length === 0 ? <div className="module-empty-state" role="status">当前记录没有待回答的问题。</div> : null}
      {questions.length > 0 ? (
      <section className="pm-panel pm-question-panel">
        <header className="pm-question-heading">
          <h2>Copilot 还想确认以下问题</h2>
          <span aria-live="polite">已完成 {completedCount}/{questions.length}</span>
        </header>
        {questions.map((question, index) => (
          <article key={question.key}>
            <b>{String(index + 1).padStart(2, "0")}</b>
            <span>
              <strong>{question.text}</strong>
              <small>补充后可以提升项目匹配准确度</small>
            </span>
            <div>
              {question.options.map((answer) => (
                <button
                  className={selectedAnswers[question.key] === answer ? "active" : ""}
                  aria-pressed={selectedAnswers[question.key] === answer}
                  key={answer}
                  onClick={() => setSelectedAnswers((current) => ({ ...current, [question.key]: answer }))}
                  type="button"
                >
                  {answer}
                </button>
              ))}
            </div>
          </article>
        ))}
        <div className="pm-question-actions">
          <button className="pm-primary-button" disabled={!allAnswered || submitting} onClick={() => void submitAnswers()} title={allAnswered ? undefined : "请先完成全部问题"} type="button"><span aria-hidden="true">✦</span>{submitting ? "生成中..." : "生成匹配结果"}</button>
          <Link className="pm-question-back" to="/projects/match"><span aria-hidden="true">←</span><span>返回修改基础需求</span></Link>
          <Link className="pm-question-defer" to="/projects">稍后继续</Link>
        </div>
        <small className="pm-question-privacy">你的选择将帮助 AI 更精准推荐，信息仅用于匹配分析</small>
      </section>
      ) : null}
    </>
  );
}

function MatchResults({ projects, sessionId, paywallEnabled = false }: { projects?: readonly DisplayProject[]; sessionId?: number; paywallEnabled?: boolean }) {
  const { matchId } = useParams();
  const navigate = useNavigate();
  const [loadedSession, setLoadedSession] = useState<ProjectMatchSession | null>(null);
  const [workflow, setWorkflow] = useState<ProjectMatchWorkflow | null>(null);
  const [loading, setLoading] = useState(projects === undefined);
  const [loadError, setLoadError] = useState("");
  const [favorited, setFavorited] = useState(false);
  const [favoritePending, setFavoritePending] = useState(false);
  const trackedResultKey = useRef("");
  const persistedProjects = projects ?? workflow?.generation?.result?.projects?.map(toDisplayProject) ?? loadedSession?.result?.projects?.map(toDisplayProject) ?? [];
  const routeSessionId = Number(matchId);
  const requestedSessionId = Number.isFinite(routeSessionId) && routeSessionId > 0 ? routeSessionId : undefined;
  const persistedSessionId = sessionId ?? loadedSession?.id ?? requestedSessionId;
  const [taskMessage, setTaskMessage] = useState("");
  const [taskError, setTaskError] = useState("");
  const generation = workflow?.generation;
  const generationActive = generation?.status === "queued" || generation?.status === "running";
  const matchFacts = loadedSession?.intent
    ? intentFacts([loadedSession.intent, ...(loadedSession.answers ?? []).map((answer) => answer.value)].join(" "))
    : [];

  useEffect(() => {
    if (!persistedSessionId || !generation || (generation.status !== "completed" && generation.status !== "partial")) return;
    const key = `${persistedSessionId}:${generation.attempt}:${generation.status}`;
    if (trackedResultKey.current === key) return;
    trackedResultKey.current = key;
    const evidence = generation.result?.evidence ?? [];
    const webSources = evidence.filter((item) => item.source_type === "web").length;
    const researchProperties = {
      match_id: persistedSessionId,
      kb_sufficiency: generation.result?.evidence_status ?? "unknown",
      web_trigger_reason: webSources > 0 ? "insufficient_kb_evidence" : "not_triggered",
      source_count: evidence.length
    };
    trackProjectEvent("match_research", researchProperties, "match_results");
    trackProjectEvent("match_result", {
      ...researchProperties,
      rounds: workflow?.revision ?? generation.attempt,
      completeness: workflow?.completeness ?? 1,
      result_count: generation.result?.projects?.length ?? persistedProjects.length
    }, "match_results");
  }, [generation, persistedProjects.length, persistedSessionId, workflow?.completeness, workflow?.revision]);

  useEffect(() => {
    if (projects !== undefined) return;
    let active = true;
    setLoading(true);
    const request = requestedSessionId
      ? projectsApi.getProjectMatch(requestedSessionId).then(async (current) => {
        if (current.status === "clarifying") {
          navigate(`/projects/matches/${requestedSessionId}/questions`, { replace:true });
          return null;
        }
        if (current.status === "ready" || current.generation?.status === "failed" || current.generation?.status === "canceled") {
          const nextGeneration = current.status === "ready"
            ? await projectsApi.generateProjectMatch(requestedSessionId)
            : current.generation;
          if (!nextGeneration) return workflowToSession(current);
          const next = { ...current, status: nextGeneration.status, generation: nextGeneration } as ProjectMatchWorkflow;
          setWorkflow(next);
          return workflowToSession(next);
        }
        setWorkflow(current);
        return workflowToSession(current);
      })
      : projectsApi.listMatches().then((payload) => (payload.matches ?? []).find((item) => item.status === "completed" && (item.result?.projects?.length ?? 0) > 0) ?? null);
    request.then((latest) => {
      if (!active) return;
      setLoadedSession(latest);
      setLoadError("");
    }).catch((error) => {
      if (active) setLoadError(apiErrorMessage(error, "暂时无法读取匹配结果"));
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [navigate, projects, requestedSessionId]);

  useEffect(() => {
    if (!requestedSessionId || !generationActive) return;
    const controller = new AbortController();
    let active = true;
    let lastEventId = 0;
    const terminal = new Set(["completed", "partial", "failed", "canceled"]);
    const refresh = async () => {
      const current = await projectsApi.getProjectMatch(requestedSessionId);
      if (!active) return current;
      setWorkflow(current);
      setLoadedSession(workflowToSession(current));
      return current;
    };
    const poll = async () => {
      while (active) {
        await new Promise((resolve) => window.setTimeout(resolve, 2000));
        if (!active) return;
        const current = await refresh();
        if (current.generation && terminal.has(current.generation.status)) return;
      }
    };
    const connect = async (remainingReconnects: number): Promise<void> => {
      try {
        await projectsApi.streamProjectMatch(requestedSessionId, lastEventId, (event) => {
          lastEventId = Math.max(lastEventId, event.id);
          void refresh();
        }, controller.signal);
        const current = await refresh();
        if (!active || (current.generation && terminal.has(current.generation.status))) return;
      } catch {
        if (!active || controller.signal.aborted) return;
      }
      if (remainingReconnects > 0) {
        await new Promise((resolve) => window.setTimeout(resolve, 500));
        if (active) await connect(remainingReconnects - 1);
        return;
      }
      if (active) await poll();
    };
    void connect(2);
    return () => {
      active = false;
      controller.abort();
    };
  }, [generationActive, requestedSessionId]);

  async function retryGeneration() {
    if (!requestedSessionId) return;
    setLoadError("");
    try {
      const next = await projectsApi.generateProjectMatch(requestedSessionId);
      setWorkflow((current) => current ? { ...current, status: next.status, generation: next } : current);
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法重试匹配"));
    }
  }

  async function cancelGeneration() {
    if (!requestedSessionId) return;
    try {
      const next = await projectsApi.cancelProjectMatch(requestedSessionId);
      setWorkflow((current) => current ? { ...current, status: next.status, generation: next } : current);
      setLoadError("");
    } catch (error) {
      setLoadError(apiErrorMessage(error, "暂时无法取消匹配"));
    }
  }

  useEffect(() => {
    if (!persistedSessionId) return;
    let active = true;
    projectsApi.listFavorites().then((payload) => {
      if (active) setFavorited((payload.favorites ?? []).some((item) => item.session_id === persistedSessionId));
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
          <Link className="rematch" to="/projects/match">重新匹配</Link>
          <button disabled={!persistedSessionId || favoritePending} onClick={() => void toggleFavorite()} type="button">
            {favoritePending ? "处理中..." : favorited ? "取消收藏" : "收藏结果"}
          </button>
          <Link to={persistedSessionId ? `/projects/matches/${persistedSessionId}/export` : "/projects/export"}>导出报告</Link>
          <Link to={`/projects/compare${persistedProjects.some((item) => item.opportunitySlug) ? `?items=${persistedProjects.map((item) => item.opportunitySlug).filter(Boolean).join(",")}` : ""}`}>加入对比</Link>
        </div>
      </div>
      {loading ? <div className="module-empty-state" role="status">正在读取最近一次匹配结果...</div> : null}
      {loadError ? <p className="form-error" role="alert">{loadError}</p> : null}
      {!loading && generationActive ? (
        <section className="pm-generation-progress" aria-label="匹配生成进度">
          <div><strong>{projectGenerationStep(generation?.current_step)}</strong><span>{generation?.progress_percent ?? 0}%</span></div>
          <progress max="100" value={generation?.progress_percent ?? 0} />
          <button onClick={() => void cancelGeneration()} type="button">取消生成</button>
        </section>
      ) : null}
      {!loading && (generation?.status === "failed" || generation?.status === "canceled") ? (
        <section className="pm-generation-progress failed" role="status">
          <strong>{generation.status === "canceled" ? "本次生成已取消" : "生成未完成"}</strong>
          <small>{generation.error_code ? `错误码：${generation.error_code}` : "可以从当前持久状态重新开始"}</small>
          <button onClick={() => void retryGeneration()} type="button">重试生成</button>
        </section>
      ) : null}
      {!loading && generation?.status === "partial" ? <p className="form-warning" role="status">证据不完整，当前结果已标记为部分完成，请优先核对来源。</p> : null}
      {!loading && !loadError && !generationActive && generation?.status !== "failed" && generation?.status !== "canceled" && persistedProjects.length === 0 ? (
        <div className="module-empty-state" role="status">暂无已完成的匹配结果，请先提交项目匹配需求。</div>
      ) : null}
      {taskMessage ? <p className="form-success" role="status">{taskMessage}</p> : null}
      {taskError ? <p className="form-error" role="alert">{taskError}</p> : null}
      {matchFacts.length > 0 ? (
        <section className="pm-result-condition-strip" aria-label="本次匹配条件">
          {matchFacts.map((fact) => <article key={fact.key}><i className={fact.icon} aria-hidden="true" /><span><small>{fact.label}</small><strong>{fact.value}</strong></span></article>)}
        </section>
      ) : null}
      {persistedProjects.length > 0 ? <section className="pm-results-layout">
        <div className="pm-result-list">
          {persistedProjects.map((project) => (
            <article key={project.title} className={`pm-result-card${project.opportunitySlug ? ` pm-project-${project.opportunitySlug}` : ""}`}>
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
                <Link to={project.opportunitySlug ? `/projects/${project.opportunitySlug}` : `/projects/matches/${persistedSessionId}`}>查看拆解</Link>
                <Link aria-label={`加入对比 ${project.title}`} to={project.opportunitySlug ? `/projects/compare?items=${project.opportunitySlug}` : "/projects/compare"}>加入对比</Link>
                <button disabled={!persistedSessionId || favoritePending} onClick={() => void toggleFavorite()} type="button">{favorited ? "取消收藏" : "收藏结果"}</button>
              </div>
              <p>风险提示：{project.risk}</p>
            </article>
          ))}
        </div>
        <aside className="pm-reason-card">
          <h2>推荐依据说明</h2>
          <p>以下内容直接来自当前匹配记录，最终仍需结合项目来源逐项验证。</p>
          <div className="pm-reason-evidence">
            {(generation?.result?.evidence?.slice(0, 4) ?? []).map((evidence, index) => (
              <article key={`${evidence.source_type}-${evidence.url ?? evidence.source_id ?? index}`}><b>{index + 1}</b><span>{evidence.title} · {Math.round(evidence.quality * 100)}%</span></article>
            ))}
          </div>
          <small>来源：匹配记录 #{persistedSessionId} · {generation?.result?.evidence?.length ?? 0} 条持久证据</small>
          <Link to={persistedProjects[0]?.opportunitySlug ? `/projects/${persistedProjects[0].opportunitySlug}` : `/projects/matches/${persistedSessionId}`}>查看项目完整拆解</Link>
          {paywallEnabled ? <Link className="pm-unlock-report" to={persistedSessionId ? `/projects/matches/${persistedSessionId}/results/paywall` : "/projects/results/paywall"}>解锁完整报告</Link> : null}
          <button disabled={!persistedSessionId} onClick={() => void generateTasks()} type="button">生成落地任务</button>
        </aside>
      </section> : null}
    </>
  );
}

function MatchHistory() {
  const [sessions, setSessions] = useState<ProjectMatchSession[]>([]);
  const [favorites, setFavorites] = useState<ProjectFavorite[]>([]);
  const [projectFavorites, setProjectFavorites] = useState<ProjectCatalogFavorite[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [timeFilter, setTimeFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("");
  const [historySort, setHistorySort] = useState<"latest" | "oldest">("latest");
  const [showAllHistory, setShowAllHistory] = useState(false);

  useEffect(() => {
    let active = true;
    projectsApi.listMatches()
      .then((matchesPayload) => {
        if (active) {
          setSessions(matchesPayload.matches ?? []);
          setError("");
        }
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取匹配历史"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    projectsApi.listFavorites()
      .then((favoritesPayload) => {
        if (active) setFavorites(favoritesPayload.favorites ?? []);
      })
      .catch(() => {
        if (active) setFavorites([]);
      });
    projectsApi.listProjectFavorites()
      .then((favoritesPayload) => {
        if (active) setProjectFavorites(favoritesPayload.favorites ?? []);
      })
      .catch(() => {
        if (active) setProjectFavorites([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const allRows = sessions.map(toHistoryRow);
  const historyTags = [...new Set(allRows.flatMap((row) => row.tags))];
  const rows = allRows
    .filter((row) => !tagFilter || row.tags.includes(tagFilter))
    .filter((row) => {
      if (timeFilter === "all") return true;
      const createdAt = new Date(sessions.find((session) => session.id === row.id)?.created_at ?? "").getTime();
      const days = Number(timeFilter);
      return Number.isFinite(createdAt) && Date.now() - createdAt <= days * 24 * 60 * 60 * 1000;
    })
    .sort((left, right) => historySort === "latest" ? right.id - left.id : left.id - right.id);

  async function removeFavorite(sessionId: number) {
    try {
      await projectsApi.unfavoriteMatch(sessionId);
      setFavorites((current) => current.filter((item) => item.session_id !== sessionId));
      setError("");
    } catch (removeError) {
      setError(apiErrorMessage(removeError, "暂时无法取消收藏"));
    }
  }

  async function removeProjectFavorite(favorite: ProjectCatalogFavorite) {
    const previous = projectFavorites;
    setProjectFavorites((current) => current.filter((item) => item.project_id !== favorite.project_id));
    try {
      await projectsApi.unfavoriteProject(favorite.slug);
      setError("");
    } catch (removeError) {
      setProjectFavorites(previous);
      setError(`取消收藏失败：${apiErrorMessage(removeError, "请稍后重试")}`);
    }
  }

  async function addFavoriteToCompare(favorite: ProjectCatalogFavorite) {
    try {
      await projectsApi.addProjectCompareItem(favorite.slug);
      setError("");
    } catch (compareError) {
      setError(`加入对比失败：${apiErrorMessage(compareError, "请稍后重试")}`);
    }
  }

  return (
    <>
      <section className="pm-history-hero">
        <div className="pm-history-hero-copy">
          <p>项目超市&nbsp;&nbsp;/&nbsp;&nbsp;AI匹配&nbsp;&nbsp;/&nbsp;&nbsp;匹配记录</p>
          <h1>匹配历史与收藏</h1>
          <strong>回看过往匹配结果，管理你保存的项目与导出记录</strong>
          <nav aria-label="匹配记录栏目">
            <a className="active" href="#pm-history-list">匹配历史</a>
            <a href="#pm-saved-projects">收藏结果</a>
            <Link to="/projects/export">导出报告</Link>
          </nav>
        </div>
        <div className="pm-history-hero-art" aria-hidden="true" />
      </section>

      <section className="pm-history-filter" aria-label="匹配历史筛选">
        <label><span>时间</span><select aria-label="按时间筛选匹配历史" onChange={(event) => setTimeFilter(event.target.value)} value={timeFilter}><option value="all">全部时间</option><option value="7">近 7 天</option><option value="30">近 30 天</option><option value="90">近 90 天</option></select></label>
        <label><span>标签</span><select aria-label="按标签筛选匹配历史" onChange={(event) => setTagFilter(event.target.value)} value={tagFilter}><option value="">全部标签</option>{historyTags.map((tag) => <option key={tag} value={tag}>{tag}</option>)}</select></label>
        <label><span>排序</span><select aria-label="匹配历史排序" onChange={(event) => setHistorySort(event.target.value as "latest" | "oldest")} value={historySort}><option value="latest">匹配时间（最新）</option><option value="oldest">匹配时间（最早）</option></select></label>
        <button onClick={() => { setTimeFilter("all"); setTagFilter(""); setHistorySort("latest"); }} type="button">↻ 刷新</button>
      </section>

      <section className="pm-history-table" id="pm-history-list">
        <h2 className="pm-visually-hidden">历史匹配</h2>
        <header aria-hidden="true"><span>时间</span><span>我的条件</span><span>匹配数量</span><span>最佳推荐</span><span>操作</span></header>
        {error && <p className="form-error" role="alert">{error}</p>}
        {loading ? <div className="module-empty-state" role="status">正在读取匹配历史...</div> : null}
        {!loading && !error && rows.length === 0 ? <div className="module-empty-state" role="status">暂无匹配历史</div> : null}
        {rows.slice(0, showAllHistory ? rows.length : 3).map((row) => (
          <article className="pm-history-row" key={row.href}>
            <div className="pm-history-date"><i aria-hidden="true">▦</i><span><strong>{row.date}</strong><small>{row.time}</small></span></div>
            <div><strong>我的条件</strong><small>{row.detail}</small><div className="pm-mini-tags">{row.tags.map((tag) => <span key={tag}>{tag}</span>)}</div></div>
            <div className="pm-history-count"><strong>{row.count.split(" ")[0]}</strong><span>个项目</span></div>
            <div><strong>{row.title}</strong><div className="pm-mini-tags">{row.tags.map((tag) => <span key={tag}>{tag}</span>)}</div></div>
            <div className="pm-history-actions">
              <Link aria-label="查看结果" to={row.href}>再次查看</Link>
              <Link to="/projects/match">重新匹配</Link>
              <Link to={`/projects/matches/${row.id}/export`}>导出</Link>
            </div>
          </article>
        ))}
        {rows.length > 3 ? <button className="pm-history-more" onClick={() => setShowAllHistory((current) => !current)} type="button">{showAllHistory ? "收起匹配历史" : "查看更多匹配历史"} <span aria-hidden="true">⌄</span></button> : null}
      </section>

      <section className="pm-saved-section" id="pm-saved-projects">
        <header><h2>收藏结果</h2><span>你收藏的匹配记录，方便随时查看与决策</span><a href="#pm-saved-projects">查看全部收藏 →</a></header>
        <div className="pm-opportunity-grid saved">
          {!loading && !error && favorites.length === 0 ? <div className="module-empty-state" role="status">暂无收藏结果</div> : null}
          {favorites.slice(0, 5).map((favorite) => favorite.session ? (
            <article className={`pm-project-${favorite.session.result?.projects?.[0]?.opportunity_slug ?? "default"}`} key={favorite.id}>
              <div className="pm-thumb" />
              <h3>{favorite.session.result?.projects?.[0]?.title ?? favorite.session.intent}</h3>
              <div className="pm-mini-tags">{favorite.session.result?.projects?.[0]?.tags?.slice(0, 3).map((tag) => <span key={tag}>{tag}</span>)}</div>
              <small>收藏于 {favorite.created_at ? new Date(favorite.created_at).toLocaleDateString("zh-CN") : "匹配记录"}</small>
              <Link to={`/projects/matches/${favorite.session_id}/results`}>查看结果</Link>
              <button aria-label="取消收藏" onClick={() => void removeFavorite(favorite.session_id)} type="button">☆</button>
            </article>
          ) : null)}
        </div>
      </section>

      <section className="pm-saved-section" aria-label="收藏项目">
        <header><h2>收藏项目</h2><span>跨设备保存的项目机会，可直接查看或加入对比</span><Link to="/projects/explore">继续浏览 →</Link></header>
        <div className="pm-opportunity-grid saved">
          {!loading && projectFavorites.length === 0 ? <div className="module-empty-state" role="status">暂无收藏项目</div> : null}
          {projectFavorites.map((favorite) => (
            <article className={`pm-project-${favorite.slug}`} key={favorite.project_id}>
              <div className="pm-thumb" />
              <h3>{favorite.title}</h3>
              <small>收藏于 {favorite.created_at ? new Date(favorite.created_at).toLocaleDateString("zh-CN") : "项目目录"}</small>
              <Link to={`/projects/${favorite.slug}`}>查看项目</Link>
              <button className="pm-saved-compare" onClick={() => void addFavoriteToCompare(favorite)} type="button">加入对比</button>
              <button aria-label={`取消收藏 ${favorite.title}`} className="pm-saved-remove" onClick={() => void removeProjectFavorite(favorite)} type="button">☆</button>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

function toProjectCase(item: EvidenceCaseItem): ProjectCase {
  return {
    id: item.id,
    slug: String(item.id),
    opportunity_id: item.project_id,
    title: item.title,
    summary: item.result_summary,
    industry: item.industry,
    case_type: item.type === "success" ? "success" : "failure",
    outcome: item.result_summary,
    key_actions: [],
    lessons: [],
    pitfalls: [],
    source_title: "查看首要来源",
    source_url: item.primary_source_url,
    captured_at: item.verified_at ?? item.published_at ?? ""
  };
}

function ProjectDetail() {
  const { matchId, opportunitySlug, projectRef } = useParams();
  const activeProjectRef = projectRef ?? opportunitySlug;
  const [searchParams] = useSearchParams();
  const [session, setSession] = useState<ProjectMatchSession | null>(null);
  const [opportunity, setOpportunity] = useState<ProjectOpportunity | null>(null);
  const [evidence, setEvidence] = useState<ProjectCase[]>([]);
  const [loading, setLoading] = useState(Boolean(matchId || activeProjectRef));
  const [error, setError] = useState("");
  const [projectFavorited, setProjectFavorited] = useState(false);
  const [projectCompared, setProjectCompared] = useState(false);
  const [collectionsLoading, setCollectionsLoading] = useState(Boolean(activeProjectRef));
  const [collectionPending, setCollectionPending] = useState<"favorite" | "compare" | "">("");
  const [collectionError, setCollectionError] = useState("");
  const trackedDetail = useRef("");
  const requestedSection = searchParams.get("section") ?? "path";
  const activeSectionKey = detailTabs.some(([key]) => key === requestedSection) ? requestedSection : "path";

  useEffect(() => {
    if (activeProjectRef) {
      let active = true;
      setLoading(true);
      setError("");
      projectsApi.getProject(activeProjectRef).then((payload) => {
        if (active) {
          setOpportunity(payload);
          setError("");
        }
        projectsApi.listEvidenceCases({ pageSize: 100 }).then((casesPayload) => {
          if (active) setEvidence(casesPayload.items.filter((item) => item.project_id === payload.id).map(toProjectCase));
        }).catch(() => {
          if (active) setEvidence([]);
        });
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
  }, [activeProjectRef, matchId]);

  useEffect(() => {
    if (!opportunity) return;
    const detailKey = `${opportunity.id}:${activeSectionKey}`;
    if (trackedDetail.current === detailKey) return;
    trackedDetail.current = detailKey;
    trackProjectEvent("project_detail_view", { project_id: opportunity.id, tab: activeSectionKey }, "project_detail");
  }, [activeSectionKey, opportunity]);

  useEffect(() => {
    if (!activeProjectRef || !opportunity) {
      setCollectionsLoading(false);
      return;
    }
    let active = true;
    setCollectionsLoading(true);
    Promise.allSettled([projectsApi.listProjectFavorites(), projectsApi.listProjectCompareItems()])
      .then(([favoriteResult, compareResult]) => {
        if (!active) return;
        setProjectFavorited(favoriteResult.status === "fulfilled" && favoriteResult.value.favorites.some((item) => item.project_id === opportunity.id || item.slug === opportunity.slug));
        setProjectCompared(compareResult.status === "fulfilled" && compareResult.value.items.some((item) => item.project_id === opportunity.id || item.slug === opportunity.slug));
      })
      .finally(() => { if (active) setCollectionsLoading(false); });
    return () => { active = false; };
  }, [activeProjectRef, opportunity]);

  async function toggleProjectFavorite() {
    if (!opportunity || collectionPending) return;
    const previous = projectFavorited;
    setProjectFavorited(!previous);
    setCollectionPending("favorite");
    setCollectionError("");
    try {
      if (previous) await projectsApi.unfavoriteProject(opportunity.slug);
      else await projectsApi.favoriteProject(opportunity.slug);
    } catch (favoriteError) {
      setProjectFavorited(previous);
      const detail = apiErrorMessage(favoriteError, "请稍后重试");
      setCollectionError(`${previous ? "取消收藏" : "收藏项目"}失败：${detail}`);
    } finally {
      setCollectionPending("");
    }
  }

  async function toggleProjectCompare() {
    if (!opportunity || collectionPending) return;
    const previous = projectCompared;
    setProjectCompared(!previous);
    setCollectionPending("compare");
    setCollectionError("");
    try {
      if (previous) await projectsApi.removeProjectCompareItem(opportunity.slug);
      else {
        await projectsApi.addProjectCompareItem(opportunity.slug);
        trackProjectEvent("project_compare_add", { project_ids: [opportunity.id] }, "project_detail");
      }
    } catch (compareError) {
      setProjectCompared(previous);
      const detail = apiErrorMessage(compareError, "请稍后重试");
      setCollectionError(`${previous ? "移出对比" : "加入对比"}失败：${detail}`);
    } finally {
      setCollectionPending("");
    }
  }

  const project = session?.result?.projects?.[0];
  const title = opportunity?.title ?? project?.title ?? "项目详情";
  const subtitle = opportunity?.summary ?? (project
    ? `来自匹配需求：${session?.intent ?? "项目匹配"}`
    : "请选择一条已发布项目机会或已保存的匹配记录");
  const profileItems = opportunity?.sections?.flatMap((section) => section.blocks ?? []).find((block) => block.type === "profile")?.items ?? [];
  const profileMetrics: [string, string, string][] = profileItems.slice(0, 4).map((item, index) => [
    item.title || "项目指标",
    item.value || "待补充",
    ["wallet", "level", "resource", "people"][index] ?? "resource"
  ]);
  const detailMetrics = opportunity
    ? profileMetrics.length > 0 ? profileMetrics : [
        ["启动预算", opportunity.budget_band || "待补充", "wallet"],
        ["项目难度", opportunity.difficulty || "待补充", "level"],
        ["资源要求", `${opportunity.resource_requirements.length} 项`, "resource"],
        ["适合方向", opportunity.tags.slice(0, 2).join(" / ") || opportunity.industry || "待补充", "people"]
      ] as [string, string, string][]
    : project
      ? [
          ["匹配度", `${project.score}分`, "score"],
          ["启动预算", project.budget, "wallet"],
          ["项目标签", `${project.tags.length} 项`, "resource"],
          ["风险提示", project.risk, "people"]
        ]
      : [];
  const activeSectionTitle = detailTabs.find(([key]) => key === activeSectionKey)?.[2] ?? "成功路径";
  const activeSection = opportunity?.sections?.find((section) => section.key === activeSectionKey || section.title === activeSectionTitle)
    ?? (activeSectionKey === "path" ? opportunity?.sections?.[0] : undefined);
  const detailBasePath = activeProjectRef ? `/projects/${activeProjectRef}` : "/projects/detail";

  function trackDetailTab(nextKey: string) {
    if (!opportunity || nextKey === activeSectionKey) return;
    trackProjectEvent("project_tab_switch", { project_id: opportunity.id, from: activeSectionKey, to: nextKey }, "project_detail");
  }

  return (
    <>
      <section className="pm-detail-hero">
        <div className="pm-detail-copy">
          <p>⌂&nbsp;&nbsp;项目超市&nbsp;&nbsp;/&nbsp;&nbsp;机会探索&nbsp;&nbsp;/&nbsp;&nbsp;项目详情</p>
          <h1>{title} <span aria-hidden="true">☆</span></h1>
          <small>{subtitle}</small>
          {opportunity ? <div className="pm-mini-tags">{[opportunity.industry, ...opportunity.tags].filter(Boolean).slice(0, 5).map((item) => <span key={item}>{item}</span>)}</div> : null}
          {activeProjectRef ? (
            <div className="pm-detail-actions">
              <button aria-label={projectFavorited ? "取消收藏项目" : "收藏项目"} disabled={!opportunity || collectionsLoading || Boolean(collectionPending)} onClick={() => void toggleProjectFavorite()} type="button">{collectionPending === "favorite" ? "处理中..." : projectFavorited ? "取消收藏" : "收藏项目"}</button>
              <button aria-label={projectCompared ? "移出项目对比" : "加入项目对比"} disabled={!opportunity || collectionsLoading || Boolean(collectionPending)} onClick={() => void toggleProjectCompare()} type="button">{collectionPending === "compare" ? "处理中..." : projectCompared ? "移出对比" : "加入对比"}</button>
              {projectCompared ? <Link to="/projects/compare">查看对比</Link> : null}
              <Link to={`${detailBasePath}/diagnosis`}>诊断我能否做</Link>
              <Link className="primary" to="/projects/match">AI 匹配类似项目</Link>
              {opportunity ? (
                <Link
                  className="pm-unlock-detail"
                  onClick={() => trackProjectEvent("project_unlock_click", { project_id: opportunity.id, tab: activeSectionKey }, "project_detail")}
                  to={`${detailBasePath}/unlock?section=${activeSectionKey}`}
                >
                  解锁完整拆解
                </Link>
              ) : null}
            </div>
          ) : null}
        </div>
        <div className={`pm-detail-art pm-detail-art-${activeProjectRef ?? "default"}`} aria-hidden="true" />
        {detailMetrics.length > 0 ? (
          <div className="pm-detail-metric-strip">
            {detailMetrics.map(([label, value, icon]) => <article key={label}><i className={icon} aria-hidden="true" /><span><small>{label}</small><strong>{value}</strong></span>{project ? <span className="pm-visually-hidden">{label === "启动预算" ? `预算 ${value}` : `${label} ${value}`}</span> : null}</article>)}
          </div>
        ) : null}
      </section>
      {loading ? <div className="module-empty-state" role="status">正在读取项目详情...</div> : null}
      {error && <p className="form-error" role="alert">{error}</p>}
      {collectionError ? <p className="form-error" role="alert">{collectionError}</p> : null}
      {!loading && !error && opportunity ? (
        <>
          <nav className="pm-detail-tabs" aria-label="项目详情栏目">
            {detailTabs.map(([key, label]) => (
              <Link className={activeSectionKey === key ? "active" : ""} key={key} onClick={() => trackDetailTab(key)} to={`${detailBasePath}?section=${key}`}>{label}</Link>
            ))}
          </nav>
          {activeSection ? <ProjectDetailSection activeKey={activeSectionKey} evidence={evidence} opportunity={opportunity} section={activeSection} /> : <div className="module-empty-state" role="status">暂无项目详情章节</div>}
        </>
      ) : null}
      {!loading && !error && session ? <MatchRecordDetail session={session} /> : null}
      {!loading && !error && !opportunity && !session ? (
        <div className="module-empty-state" role="status">当前地址没有关联项目记录。<Link to="/projects/explore">浏览已发布项目机会</Link></div>
      ) : null}
    </>
  );
}

function ProjectEvidenceCards({ evidence }: { evidence: ProjectCase[] }) {
  if (evidence.length === 0) return <div className="module-empty-state" role="status">暂无可追溯案例来源</div>;
  return (
    <div className="pm-case-mini-grid">
      {evidence.slice(0, 4).map((item) => (
        <article className={`pm-case-${item.slug}`} key={item.id}>
          <div className="pm-thumb" aria-hidden="true" />
          <span>{item.case_type === "success" ? "成功案例" : "失败教训"}</span>
          <strong>{item.title}</strong>
          <small>{item.summary}</small>
          <b>{item.outcome}</b>
          {item.source_url ? <a href={item.source_url} onClick={() => trackProjectEvent("project_explore_source_click", { ...(item.opportunity_id ? { opportunity_id: item.opportunity_id } : {}), group: item.case_type, query: item.title, source_id: item.id }, "project_detail_evidence")} rel="noreferrer" target="_blank">{item.source_title || "查看来源"}</a> : <small>来源链接待补充</small>}
        </article>
      ))}
    </div>
  );
}

function contentBlockIcon(block: ProjectContentBlock, index: number) {
  const itemTitle = block.items?.[index]?.title ?? "";
  const iconNames: Record<string, string> = {
    "核心价值": "content-value", "主要变现": "growth-stage", "交付周期": "delivery-period",
    "企业品牌方": "brand-power", "个体 IP": "audience", "本地商家": "solo-business", "MCN / 代运营": "customer-access",
    "常见报价区间": "pricing", "平均交付周期": "delivery-period", "平均修改次数": "revision-count", "平均毛利率": "gross-margin",
    "小红书": "content-value", "抖音": "content-efficiency", "视频号": "delivery-efficiency", "私域转介绍": "customer-access",
    "咨询人数": "traffic-funnel", "脚本下单数": "service-package", "复购客户数": "customer-trust",
    "订单数": "service-package", "询盘转化率": "traffic-funnel", "复购率": "customer-trust", "客户满意度": "customer-feedback", "平均客单价": "pricing",
    "交付快": "fast-delivery", "低启动成本": "low-cost", "适合轻资产创业": "solo-business", "可标准化服务": "standard-service", "内容生产效率高": "content-efficiency", "易做案例积累": "easy-replication",
    "审美与表达门槛": "trust-gap", "客户信任建立慢": "customer-trust", "前期获客不稳定": "market-volatility", "同质化竞争": "homogeneous-competition", "修改返工风险": "customer-churn", "规模化管理要求": "process-awareness",
    "适合的人": "content-skill", "不适合的人": "difficulty", "机会区": "growth-stage", "警惕区": "failure-warning", "优化区": "process-awareness", "风险区": "business-risk", "总体判断": "content-value", "关键门槛能力": "content-skill",
    "先做小而明确的场景": "positioning", "用样稿降低成交门槛": "case-growth", "用复盘沉淀模板": "template", "通过反馈优化沟通": "customer-feedback", "建立内容素材库": "asset-library",
    "个人 IP 拍摄": "audience", "产品测评单": "success-case", "知识付费类": "content-value", "企业知识 IP": "brand-power", "短视频代运营": "delivery-efficiency", "脚本咨询": "service-package", "内容策划服务": "content-skill",
    "高频踩坑": "high-frequency-risk", "业务风险": "business-risk", "影响严重": "severity", "可控可防": "preventable",
    "一开始什么都接": "demand-loss", "只拼低价": "low-cost", "没有标准交付流程": "process-awareness", "忽视客户反馈": "customer-feedback", "过度承诺效果": "severity", "案例和定位不清晰": "trust-gap",
    "返工过多": "rework", "需求失控": "demand-loss", "获客渠道单一": "traffic-funnel", "沟通能力": "customer-feedback", "项目管理": "process-awareness", "内容审核": "content-skill", "基础数据复盘": "action-checklist"
  };
  const fallbackByType: Record<string, string[]> = {
    overview: ["content-value", "growth-stage", "delivery-period", "audience"],
    metrics: ["brand-power", "pricing", "delivery-period", "gross-margin"],
    channels: ["content-value", "content-efficiency", "delivery-efficiency", "customer-access"],
    funnel: ["traffic-funnel", "service-package", "customer-trust"],
    matrix: ["growth-stage", "failure-warning", "process-awareness", "business-risk"],
    cards: ["positioning", "service-package", "case-growth", "delivery-efficiency"]
  };
  const fallback = fallbackByType[block.type] ?? ["content-value"];
  return `/project-market/detail-icons/${iconNames[itemTitle] ?? fallback[index % fallback.length]}.png`;
}

function ProjectContentBlockView({ block, evidence, opportunityID }: { block: ProjectContentBlock; evidence: ProjectCase[]; opportunityID: number }) {
  const items = block.items ?? [];
  const series = block.series ?? [];
  const columns = Math.max(1, Math.min(block.columns ?? (items.length > 0 ? items.length : 1), 6));

  if (block.type === "chart") {
    const chartPoints = series.map((item, index) => ({
      x: series.length <= 1 ? 500 : 20 + (index / (series.length - 1)) * 960,
      y: 92 - Math.max(6, Math.min(Number(item.progress) || 0, 100)) * 0.8
    }));
    const lastSeriesItem = series[series.length - 1];
    return (
      <article className="pm-structured-block type-chart">
        {block.title ? <h2>{block.title}</h2> : null}
        {block.subtitle ? <p>{block.subtitle}</p> : null}
        <div className="pm-structured-chart" aria-label={block.title ?? "数据趋势"}>
          <svg aria-hidden="true" preserveAspectRatio="none" viewBox="0 0 1000 100">
            <defs><linearGradient id="pm-chart-line" x1="0" x2="1"><stop offset="0" stopColor="#7560f5" /><stop offset="1" stopColor="#2d75e9" /></linearGradient></defs>
            <polyline points={chartPoints.map((point) => `${point.x},${point.y}`).join(" ")} />
            {chartPoints.map((point, index) => <circle cx={point.x} cy={point.y} key={`${series[index]?.title}-${index}`} r="4" />)}
          </svg>
          <div className="pm-structured-chart-labels" style={{ "--pm-chart-points": Math.max(series.length, 1) } as CSSProperties}>
            {series.map((item, index) => <small key={`${item.title}-${index}`}>{item.title}</small>)}
          </div>
          {lastSeriesItem ? <output>{lastSeriesItem.meta || lastSeriesItem.value}</output> : null}
        </div>
      </article>
    );
  }

  if (block.type === "sources" && evidence.length > 0) {
    return (
      <article className="pm-structured-block type-sources">
        {block.title ? <h2 aria-label="来源与证据">{block.title}</h2> : null}
        <div className="pm-structured-grid" style={{ "--pm-columns": columns } as CSSProperties}>
          {evidence.filter((item) => item.source_url).slice(0, columns).map((item) => <a href={item.source_url} key={item.id} onClick={() => trackProjectEvent("project_explore_source_click", { opportunity_id: opportunityID, group: item.case_type, query: item.title, source_id: item.id }, "project_detail_evidence")} rel="noreferrer" target="_blank"><strong>{item.source_title}</strong><small>{item.title}</small><span>查看出处 →</span></a>)}
        </div>
      </article>
    );
  }

  if (/真实案例/.test(block.title ?? "") && evidence.length > 0) {
    return (
      <article className="pm-structured-block type-evidence">
        {block.title ? <h2>{block.title}</h2> : null}
        {block.subtitle ? <p>{block.subtitle}</p> : null}
        <ProjectEvidenceCards evidence={evidence.slice(0, columns)} />
      </article>
    );
  }

  return (
    <article className={`pm-structured-block type-${block.type || "cards"}`}>
      {block.title ? <h2>{block.title}</h2> : null}
      {block.subtitle ? <p>{block.subtitle}</p> : null}
      <div className="pm-structured-grid" style={{ "--pm-columns": columns } as CSSProperties}>
        {items.map((item, index) => {
          const tone = ["primary", "positive", "warning", "danger", "neutral"].includes(item.tone ?? "") ? item.tone : "neutral";
          const progress = Math.max(0, Math.min(Number(item.progress) || 0, 100));
          const source = /真实案例/.test(block.title ?? "") ? evidence[index] : undefined;
          return (
            <section className={`tone-${tone}`} key={`${item.title}-${item.value}-${index}`}>
              {block.type === "checklist" ? <input aria-label={`已具备 ${item.title ?? `第 ${index + 1} 项`}`} type="checkbox" /> : block.type === "steps" || block.type === "actions" ? <i aria-hidden="true">{block.type === "steps" ? index + 1 : item.meta ?? index + 1}</i> : <i aria-hidden="true" className="pm-structured-icon" style={{ backgroundImage: `url("${contentBlockIcon(block, index)}")` }} />}
              <div>
                {item.title ? <strong>{item.title}</strong> : null}
                {item.value ? <b>{item.value}</b> : null}
                {item.detail ? <p>{item.detail}</p> : null}
                {item.meta && block.type !== "actions" ? <small>{item.meta}</small> : null}
                {progress > 0 ? <span className="pm-structured-progress"><em style={{ width: `${progress}%` }} /></span> : null}
                {(item.tags ?? []).length > 0 ? <span className="pm-structured-tags">{item.tags?.map((tag) => <small key={tag}>{tag}</small>)}</span> : null}
                {source?.source_url ? <a href={source.source_url} onClick={() => trackProjectEvent("project_explore_source_click", { opportunity_id: opportunityID, group: source.case_type, query: source.title, source_id: source.id }, "project_detail_evidence")} rel="noreferrer" target="_blank">{source.source_title || "查看来源"} →</a> : null}
              </div>
            </section>
          );
        })}
      </div>
    </article>
  );
}

function ProjectDetailSection({ activeKey, evidence, opportunity, section }: {
  activeKey: string;
  evidence: ProjectCase[];
  opportunity: ProjectOpportunity;
  section: OpportunitySection;
}) {
  const items = section.items ?? [];
  const structuredBlocks = (section.blocks ?? []).filter((block) => block.type !== "profile");

  if (structuredBlocks.length > 0) {
    return <section className={`pm-detail-content pm-structured-content ${activeKey}`}>{structuredBlocks.map((block, index) => <ProjectContentBlockView block={block} evidence={evidence} key={`${block.type}-${block.title}-${index}`} opportunityID={opportunity.id} />)}</section>;
  }

  return (
    <section className={`pm-detail-content pm-detail-legacy ${activeKey}`}>
      <article className="pm-detail-block">
        <h2>{section.title}</h2>
        <p>{section.body}</p>
        {items.length > 0 ? <ul>{items.map((item) => <li key={item}>{item}</li>)}</ul> : <div className="module-empty-state">该章节暂无结构化内容</div>}
        <div className="pm-mini-tags">{[opportunity.industry, ...opportunity.tags].filter(Boolean).slice(0, 5).map((item) => <span key={item}>{item}</span>)}</div>
      </article>
      {evidence.some((item) => item.source_url) ? <article className="pm-detail-block"><h2>来源与证据</h2><ProjectEvidenceCards evidence={evidence} /></article> : null}
    </section>
  );
}

function DiagnosisOverlay() {
  const { opportunitySlug } = useParams();
  const navigate = useNavigate();
  const closeHref = `/projects/opportunities/${opportunitySlug ?? "ai-short-video-studio"}`;
  const steps = [
    ["1", "读取项目要求", "从项目目录读取技能、预算与资源门槛"],
    ["2", "补充诊断目标", "由你填写可投入时间、能力方向与当前卡点"],
    ["3", "完成能力问答", "使用本次真实回答形成评估依据"],
    ["4", "输出模型建议", "列明适配度、关键假设与后续学习方向"]
  ] as const;

  return createPortal(
    <div className="pm-modal-scrim pm-diagnosis-scrim">
      <section className="pm-diagnosis-modal" role="dialog" aria-modal="true" aria-labelledby="project-diagnosis-title">
        <Link className="pm-modal-close" aria-label="关闭项目诊断" to={closeHref}>×</Link>
        <header>
          <div><h2 id="project-diagnosis-title">诊断我能否做这个项目？</h2><p>系统将结合当前项目要求与您的数据，进行综合评估。</p></div>
          <div className="pm-diagnosis-hero-art" aria-hidden="true" />
        </header>
        <div className="pm-diagnosis-steps">
          {steps.map(([number, title, detail]) => (
            <article key={number}><b>{number}</b><strong>{title}</strong><small>{detail}</small></article>
          ))}
        </div>
        <div className="pm-diagnosis-meta">
          <div className="pm-diagnosis-sources"><strong>本次评估的可用依据</strong><span><i aria-hidden="true">▦</i><b>项目超市</b><small>项目要求</small></span><span><i aria-hidden="true">✓</i><b>本次填写</b><small>目标与时间</small></span><span><i aria-hidden="true">⌘</i><b>能力问答</b><small>真实回答</small></span><span><i aria-hidden="true">A</i><b>模型输出</b><small>假设与建议</small></span></div>
          <div><strong>完成方式</strong><b>按实际填写进度</b><small>不会读取未接入的模块数据</small></div>
          <div><strong>你将获得</strong><span>□ 模型适配度</span><span>□ 关键能力差距</span><span>□ 评估依据</span><span>□ 学习建议</span></div>
        </div>
        <footer><Link to={closeHref}>稍后再说</Link><button aria-label="开始诊断" onClick={() => { trackProjectEvent("project_diagnose_submit", {}, "project_diagnosis"); navigate(`/learning/diagnosis?project=${encodeURIComponent(opportunitySlug ?? "")}`); }} type="button">✦ 开始诊断</button></footer>
        <small className="pm-modal-privacy">所有数据仅用于能力评估与学习建议，严格保护你的隐私安全</small>
      </section>
    </div>,
    document.body
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
  const [searchParams, setSearchParams] = useSearchParams();
  const setSearchParamsRef = useRef(setSearchParams);
  const requestedSlugs = useRef((searchParams.get("items") ?? "").split(",").filter(Boolean).slice(0, 5));
  const [options, setOptions] = useState<ProjectOpportunity[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [comparison, setComparison] = useState<ProjectOpportunity[]>([]);
  const [comparisonId, setComparisonId] = useState<number | null>(null);
  const [comparisonExport, setComparisonExport] = useState<ProjectExport | null>(null);
  const [creatingComparison, setCreatingComparison] = useState(false);
  const [downloadingComparison, setDownloadingComparison] = useState(false);
  const [collectionsLoading, setCollectionsLoading] = useState(true);
  const [collectionPending, setCollectionPending] = useState<string[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    setSearchParamsRef.current = setSearchParams;
  }, [setSearchParams]);
  useEffect(() => {
    let active = true;
    projectsApi.listProjects({ pageSize: 100 }).then((payload) => { if (active) setOptions(payload.items); })
      .catch((loadError) => { if (active) setError(apiErrorMessage(loadError, "暂时无法读取项目目录")); });

    async function loadPersistedSelection() {
      let next: string[] = [];
      try {
        const payload = await projectsApi.listProjectCompareItems();
        next = payload.items.map((item) => item.slug).filter(Boolean).slice(0, 5);
      } catch (loadError) {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取已保存的对比项目"));
      }
      for (const slug of requestedSlugs.current) {
        if (next.includes(slug)) continue;
        try {
          const item = await projectsApi.addProjectCompareItem(slug);
          next.push(item.slug);
        } catch (addError) {
          if (active) setError(apiErrorMessage(addError, "暂时无法加入对比"));
          break;
        }
      }
      if (!active) return;
      setSelected(next);
      setSearchParamsRef.current(next.length > 0 ? { items: next.join(",") } : {}, { replace: true });
      setCollectionsLoading(false);
    }
    void loadPersistedSelection();
    return () => { active = false; };
  }, []);
  async function toggle(slug: string) {
    if (collectionPending.includes(slug)) return;
    const previous = selected;
    const removing = selected.includes(slug);
    const next = removing ? selected.filter((item) => item !== slug) : [...selected, slug];
    setComparison([]);
    setComparisonId(null);
    setComparisonExport(null);
    setSelected(next);
    setSearchParams(next.length > 0 ? { items: next.join(",") } : {}, { replace: true });
    setCollectionPending((current) => [...current, slug]);
    setError("");
    try {
      if (removing) await projectsApi.removeProjectCompareItem(slug);
      else await projectsApi.addProjectCompareItem(slug);
    } catch (toggleError) {
      setSelected(previous);
      setSearchParams(previous.length > 0 ? { items: previous.join(",") } : {}, { replace: true });
      setError(apiErrorMessage(toggleError, removing ? "暂时无法移出对比" : "暂时无法加入对比"));
    } finally {
      setCollectionPending((current) => current.filter((item) => item !== slug));
    }
  }
  async function clearSelection() {
    if (selected.length === 0 || collectionPending.length > 0) return;
    const previous = selected;
    setSelected([]);
    setSearchParams({}, { replace: true });
    setComparison([]);
    setComparisonId(null);
    setComparisonExport(null);
    setCollectionPending(previous);
    setError("");
    try {
      await Promise.all(previous.map((slug) => projectsApi.removeProjectCompareItem(slug)));
    } catch (clearError) {
      try {
        const payload = await projectsApi.listProjectCompareItems();
        const restored = payload.items.map((item) => item.slug).filter(Boolean).slice(0, 5);
        setSelected(restored);
        setSearchParams(restored.length > 0 ? { items: restored.join(",") } : {}, { replace: true });
      } catch {
        setSelected(previous);
        setSearchParams({ items: previous.join(",") }, { replace: true });
      }
      setError(apiErrorMessage(clearError, "暂时无法清空对比项目"));
    } finally {
      setCollectionPending([]);
    }
  }
  async function createComparison() {
    if (selected.length < 2 || creatingComparison) return;
    setCreatingComparison(true);
    try {
      const result = await projectsApi.createComparison(selected);
      setComparison(result.items);
      setComparisonId(result.id);
      setError("");
      trackProjectEvent("project_compare_view", { project_ids: result.items.map((item) => item.id) }, "project_compare");
    } catch (createError) {
      setError(apiErrorMessage(createError, "暂时无法创建项目对比"));
    } finally {
      setCreatingComparison(false);
    }
  }
  async function exportComparison() {
    if (!comparisonId) return;
    try {
      trackProjectEvent("match_export_click", { match_id: comparisonId, plan: "comparison_pdf", blocked: false }, "project_compare");
      const result = await projectsApi.createExport("comparison", comparisonId);
      setComparisonExport(result);
      setError("");
    } catch (exportError) {
      setError(apiErrorMessage(exportError, "暂时无法导出项目对比"));
    }
  }
  useEffect(() => {
    if (!comparisonExport || comparisonExport.status === "ready" || comparisonExport.status === "failed") return;
    let active = true;
    const timer = window.setInterval(() => {
      projectsApi.getExport(comparisonExport.id).then((item) => {
        if (!active) return;
        setComparisonExport(item);
        if (item.status === "failed") setError("PDF 生成失败，请重新创建导出");
      }).catch((loadError) => {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取导出进度"));
      });
    }, 800);
    return () => { active = false; window.clearInterval(timer); };
  }, [comparisonExport]);
  async function downloadComparison() {
    if (!comparisonExport || comparisonExport.status !== "ready" || downloadingComparison) return;
    setDownloadingComparison(true);
    try {
      const blob = await projectsApi.downloadExport(comparisonExport.id);
      saveProjectExport(blob, `project-comparison-${comparisonId ?? comparisonExport.id}.pdf`);
      setError("");
    } catch (downloadError) {
      setError(apiErrorMessage(downloadError, "暂时无法下载项目对比"));
    } finally {
      setDownloadingComparison(false);
    }
  }
  const compareRows = [
    ["启动预算", (item: ProjectOpportunity) => item.budget_band || "待补充"],
    ["验证周期", () => "需通过真实订单验证"],
    ["难度", (item: ProjectOpportunity) => item.difficulty || "待补充"],
    ["资源要求", (item: ProjectOpportunity) => item.resource_requirements.join("、") || "待补充"],
    ["可复制性", (item: ProjectOpportunity) => item.tags.includes("可复制") || item.tags.includes("可复用交付") ? "目录标记为可复制" : "需通过案例验证"],
    ["增长潜力", (item: ProjectOpportunity) => item.summary],
    ["风险提示", (item: ProjectOpportunity) => item.sections?.find((section) => section.title === "要避免的行为")?.body || "需查看案例与来源"],
    ["适合人群", (item: ProjectOpportunity) => [item.industry, ...item.tags].filter(Boolean).slice(0, 3).join("、")],
    ["能力工具", (item: ProjectOpportunity) => item.resource_requirements.slice(0, 3).join("、") || "待补充"]
  ] as const;
  const primaryComparison = comparison[0];
  return (
    <>
      <div className="pm-result-head">
        <div>
          <p>项目超市 / 项目对比</p>
          <h1>项目对比</h1>
          <small>并排比较多个项目的预算、门槛、风险与增长潜力，帮助你更快做决策</small>
        </div>
        <div>
          <button disabled={selected.length === 0 || collectionPending.length > 0} onClick={() => void clearSelection()} type="button">移除项目</button>
          <Link to={comparison[0] ? `/projects/${comparison[0].slug}` : "/projects/explore"}>查看拆解</Link>
          <Link to="/projects/history#pm-saved-projects">查看收藏记录</Link>
          <button disabled={creatingComparison || collectionPending.length > 0 || (comparisonId ? false : selected.length < 2)} onClick={() => void (comparisonId ? exportComparison() : createComparison())} type="button">{creatingComparison ? "对比中..." : comparisonId ? "导出对比" : "开始对比"}</button>
        </div>
      </div>
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      <p className="pm-compare-saved-state" role="status">{collectionsLoading ? "正在读取已保存的对比项目..." : `已保存 ${selected.length}/5 个项目`}</p>
      {comparisonExport && comparisonExport.status !== "ready" && comparisonExport.status !== "failed" ? <p role="status">{comparisonExport.status === "queued" ? "对比报告已进入生成队列..." : "正在排版生成对比报告..."}</p> : null}
      {comparisonExport ? <button className="pm-compare-download" disabled={comparisonExport.status !== "ready" || downloadingComparison} onClick={() => void downloadComparison()} type="button">{downloadingComparison ? "下载中..." : comparisonExport.status === "ready" ? "下载 PDF 对比报告" : "等待 PDF 生成"}</button> : null}
      {comparison.length === 0 ? (
        <section className="pm-panel pm-compare-picker">
          <h2>选择 2-5 个项目</h2>
          <div>{options.map((item) => <label key={item.id}><input aria-label={item.title} checked={selected.includes(item.slug)} disabled={collectionsLoading || collectionPending.includes(item.slug)} onChange={() => void toggle(item.slug)} type="checkbox" /><span><strong>{item.title}</strong><small>{item.industry} · {item.budget_band}</small></span></label>)}</div>
        </section>
      ) : (
        <>
          <section className="pm-compare-portrait">
            <strong>当前参考画像：</strong>
            <span>已选 {comparison.length} 个项目</span>
            <span>{[...new Set(comparison.map((item) => item.industry))].join(" / ")}</span>
            <span>{[...new Set(comparison.flatMap((item) => item.tags))].slice(0, 3).join(" / ")}</span>
            <Link to="/projects/match">修改画像</Link>
          </section>
          <section className="pm-compare-table" style={{ "--pm-compare-count": comparison.length } as CSSProperties}>
            <header>
              <strong>项目名称</strong>
              {comparison.map((project) => <article className={`pm-project-${project.slug}`} key={project.id}><button aria-label={`移除 ${project.title}`} onClick={() => void toggle(project.slug)} type="button">×</button><h2>{project.title}</h2><div className="pm-result-image" /></article>)}
            </header>
            {compareRows.map(([label, getValue]) => (
              <div className="pm-compare-row" key={label}>
                <strong>{label}</strong>
                {comparison.map((project) => <span key={project.id}>{getValue(project)}</span>)}
              </div>
            ))}
          </section>
          <section className="pm-compare-summary">
            <div className="pm-compare-trophy" aria-hidden="true" />
            <div><small>优先核对 · 按已选项目顺序</small><strong className="pm-compare-recommend-title">{primaryComparison?.title}</strong><p>先核对预算、资源与风险项，再进入完整拆解验证项目是否适合当前条件。</p></div>
            <ul>{primaryComparison ? [primaryComparison.budget_band, primaryComparison.difficulty, ...primaryComparison.resource_requirements.slice(0, 2)].filter(Boolean).map((item) => <li key={item}>{item}</li>) : null}</ul>
            <Link to={primaryComparison ? `/projects/${primaryComparison.slug}` : "/projects/explore"}>查看完整项目拆解 →</Link>
          </section>
        </>
      )}
    </>
  );
}

function ProjectHero({ variant, breadcrumb, title, subtitle, action, href }: {
  variant: "match" | "questions";
  breadcrumb: string[];
  title: string;
  subtitle: string;
  action: string;
  href: string;
}) {
  return (
    <section className={`pm-flow-hero ${variant}`}>
      <div className="pm-flow-hero-copy">
        <p>{breadcrumb.map((item, index) => <span key={item}>{index > 0 ? " / " : ""}{item}</span>)}</p>
        <h1>{title}</h1>
        <small>{subtitle}</small>
      </div>
      <div className="pm-flow-hero-art" aria-hidden="true" />
      <Link className="pm-flow-hero-action" to={href}><span aria-hidden="true">◷</span>{action}</Link>
    </section>
  );
}

function Considerations() {
  const considerations = [
    ["ability", "能力适配", "基于你的技能与经验评估项目匹配度"],
    ["budget", "预算门槛", "核对启动与运营成本是否落在预算范围"],
    ["resource", "资源要求", "评估人力、工具、渠道等资源可获得性"],
    ["growth", "增长路径", "分析市场需求、变现方式与长期增长潜力"]
  ] as const;
  return (
    <section className="pm-panel pm-considerations">
      <h2>匹配逻辑会考虑</h2>
      <div className="pm-factor-grid">
        {considerations.map(([icon, title, detail]) => (
          <article key={title}>
            <i className={icon} aria-hidden="true" />
            <span><strong>{title}</strong><small>{detail}</small></span>
          </article>
        ))}
      </div>
      <small className="pm-data-safety">智活AI 严格保护你的隐私，匹配过程不会被用于任何其他用途</small>
    </section>
  );
}

function ProjectCopilot({ variant }: { variant: ProjectMarketVariant }) {
  const copilotPanel = useRegisteredCopilotPanel();
  const { matchId, opportunitySlug, projectRef } = useParams();
  const activeProjectRef = projectRef ?? opportunitySlug;
  const [searchParams] = useSearchParams();
  const [opportunities, setOpportunities] = useState<ProjectOpportunity[]>([]);
  const [cases, setCases] = useState<EvidenceCaseItem[]>([]);
  const [sessions, setSessions] = useState<ProjectMatchSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ProjectMatchSession | null>(null);
  const [contextOpportunity, setContextOpportunity] = useState<ProjectOpportunity | null>(null);

  useEffect(() => {
    let active = true;
    if (["home", "explore", "compare"].includes(variant)) {
      projectsApi.listProjects({ pageSize: 20 }).then((payload) => { if (active) setOpportunities(payload.items ?? []); }).catch(() => { if (active) setOpportunities([]); });
    }
    if (variant === "cases") {
      projectsApi.listEvidenceCases({ pageSize: 20 }).then((payload) => { if (active) setCases(payload.items ?? []); }).catch(() => { if (active) setCases([]); });
    }
    if (["match", "questions", "history", "results", "paywall", "export"].includes(variant)) {
      const routeSessionId = Number(matchId);
      if (Number.isFinite(routeSessionId) && routeSessionId > 0) {
        projectsApi.getMatch(routeSessionId).then((payload) => { if (active) setCurrentSession(payload); }).catch(() => { if (active) setCurrentSession(null); });
      } else {
        projectsApi.listMatches().then((payload) => { if (active) setSessions(payload.matches ?? []); }).catch(() => { if (active) setSessions([]); });
      }
    }
    if (activeProjectRef && ["detail", "diagnosis"].includes(variant)) {
      projectsApi.getProject(activeProjectRef).then((payload) => { if (active) setContextOpportunity(payload); }).catch(() => { if (active) setContextOpportunity(null); });
    }
    return () => { active = false; };
  }, [activeProjectRef, matchId, variant]);

  const selectedSlugs = (searchParams.get("items") ?? "").split(",").filter(Boolean);
  const comparedOpportunities = opportunities.filter((item) => selectedSlugs.includes(item.slug));
  const latestSession = currentSession ?? sessions.find((item) => item.status === "completed" && (item.result?.projects?.length ?? 0) > 0) ?? sessions[0];
  const latestProject = latestSession?.result?.projects?.[0];
  const activeDetailKey = detailTabs.some(([key]) => key === searchParams.get("section")) ? searchParams.get("section") ?? "path" : "path";
  const activeDetailTitle = detailTabs.find(([key]) => key === activeDetailKey)?.[2];
  const activeDetailSection = contextOpportunity?.sections?.find((section) => section.key === activeDetailKey || section.title === activeDetailTitle)
    ?? (activeDetailKey === "path" ? contextOpportunity?.sections?.[0] : undefined);
  const detailItems = (activeDetailSection?.items?.length
    ? activeDetailSection.items
    : activeDetailSection?.blocks?.flatMap((block) => block.items?.map((item) => item.title || item.value || "").filter(Boolean) ?? []) ?? []).slice(0, 4);
  const caseLessons = cases.map((item) => item.result_summary).slice(0, 3);

  if (!copilotPanel.isPanelOpen) {
    return copilotPanel.hasSharedController ? null : <FloatingCopilotOrb onActivate={copilotPanel.openPanel} />;
  }

  let conversation: React.ReactNode;
  if (variant === "match" || variant === "questions") {
    const steps = variant === "match" ? ["提交需求，AI 开始分析", "确认关键信息", "生成匹配结果与建议"] : ["补齐预算、时间与项目偏好", "提交当前问题的真实选择", "按回答生成匹配结果"];
    conversation = <>{latestSession ? <div className="pm-copilot-bubble user"><b>我</b><p>{latestSession.intent}</p></div> : null}<div className="pm-copilot-bubble assistant"><b>AI</b><p>{variant === "match" ? "填写需求后，我会识别目标、预算、时间与资源，再决定是否需要补充问题。" : "当前补充问题来自这条匹配记录，完成选择后即可继续生成结果。"}</p></div><div className="pm-copilot-summary"><strong>{variant === "match" ? "接下来会发生" : "本次补充流程"}</strong>{steps.map((item, index) => <span key={item}><b>{index + 1}</b> {item}</span>)}</div></>;
  } else if (variant === "home") {
    conversation = <>
      <div className="pm-copilot-bubble user"><b>我</b><p>请<br />帮我找一些适合一人公司、轻资产、<br />可快速起盘的项目</p></div>
      <div className="pm-copilot-bubble assistant"><b>AI</b><p>好的！我会基于你的偏好分析适合的<br />机会，并推荐可快速起盘的项目。<br /><br />以下是为你精选的推荐：</p></div>
      <div className="pm-copilot-projects">
        {opportunities.length > 0 ? opportunities.slice(0, 3).map((item, index) => (
          <Link key={item.id} to={`/projects/${item.slug}`}>
            <i className="pm-project-thumb" style={{ backgroundImage: `url(${featuredOpportunityArt[index] ?? featuredOpportunityArt[0]})` }} />
            <span><strong>{item.title}</strong><small>{item.tags.slice(0, 2).join(" · ") || item.track || "已发布项目"}</small></span>
          </Link>
        )) : <span>当前目录暂无可推荐项目</span>}
      </div>
      <Link className="pm-copilot-cta" to="/projects/match">查看 AI 匹配页 →</Link>
    </>;
  } else if (variant === "explore") {
    const copilotOpportunities = opportunities.slice(0, 4);
    conversation = <>
      <div className="pm-copilot-bubble user"><b>我</b><p>我想找适合一个人公司、预算有限的项目机会，有什么推荐？</p></div>
      <div className="pm-copilot-bubble assistant"><b>AI</b><p><strong>智活 Copilot</strong>为你筛选出以下几个方向，综合考虑启动成本、可复制性和变现潜力：</p></div>
      <div className="pm-copilot-explore-list">
        {copilotOpportunities.length > 0 ? copilotOpportunities.map((item) => <Link key={item.id} to={`/projects/${item.slug}`}><strong>{item.title}</strong><small>{item.tags.slice(0, 2).join(" · ")} · {item.budget_band}</small></Link>) : <span>当前目录暂无可推荐项目</span>}
      </div>
      <p className="pm-copilot-question">想查看更多项目拆解和匹配度分析吗？</p>
      <Link className="pm-copilot-cta" to="/projects/match">去 AI 匹配 →</Link>
    </>;
  } else if (variant === "cases") {
    conversation = <><div className="pm-copilot-bubble assistant"><b>AI</b><p>以下案例来自证据案例库，并保留各自的首要来源。</p></div><div className="pm-copilot-projects cases">{cases.slice(0, 3).map((item) => <Link key={item.id} to={`/project-cases/${item.id}`}><i className={`pm-case-thumb pm-case-${item.id}`} /><span><strong>{item.title}</strong><small>{item.type === "success" ? "成功案例" : "失败复盘"} · {item.source_count} 条证据</small></span></Link>)}</div><div className="pm-copilot-summary"><strong>案例库中的核验结论</strong>{caseLessons.length > 0 ? caseLessons.map((item) => <span key={item}>✓ {item}</span>) : <span>当前接口没有已发布经验</span>}</div></>;
  } else if (variant === "history") {
    conversation = <><div className="pm-copilot-bubble assistant"><b>AI</b><p>当前接口返回 {sessions.length} 条匹配记录。</p></div><div className="pm-copilot-summary">{latestProject ? <><span>✓ 最近完成记录的首项：{latestProject.title}</span><span>✓ 可重新打开并核对原始匹配条件</span><span>✓ 可导出已保存的服务端记录</span></> : <span>完成首次匹配后，这里会显示历史记录摘要。</span>}</div></>;
  } else if (variant === "results") {
    conversation = <><div className="pm-copilot-bubble assistant"><b>AI</b><p>{latestProject ? `根据当前匹配记录，优先验证 ${latestProject.title}。` : "匹配结果正在从会话接口读取。"}</p></div>{latestProject ? <div className="pm-copilot-result"><small>TOP 1 推荐</small><strong>{latestProject.title}</strong><b>匹配度 {latestProject.score} 分</b><p>这是最适合当前条件的项目：</p>{latestProject.reasons.map((item) => <span key={item}>✓ {item}</span>)}<Link to={latestProject.opportunity_slug ? `/projects/${latestProject.opportunity_slug}` : "/projects/history"}>查看项目完整拆解 →</Link></div> : null}</>;
  } else if (variant === "detail" || variant === "diagnosis") {
    conversation = <><div className="pm-copilot-bubble assistant"><b>AI</b><p>{contextOpportunity ? `已读取 ${contextOpportunity.title} 的“${activeDetailSection?.title ?? "项目详情"}”接口内容。` : "正在读取当前项目上下文。"}</p></div><div className="pm-copilot-summary"><strong>{activeDetailSection?.title ?? "当前章节"}</strong>{detailItems.length > 0 ? detailItems.map((item) => <span key={item}>• {item}</span>) : <span>当前章节暂无结构化数据</span>}</div>{contextOpportunity ? <Link className="pm-copilot-cta" to={`/tasks?project=${encodeURIComponent(contextOpportunity.slug)}`}>生成落地计划表</Link> : null}</>;
  } else if (variant === "compare") {
    const firstSelected = comparedOpportunities[0];
    conversation = <>{firstSelected ? <div className="pm-copilot-result"><small>已选项目摘要</small><strong>{firstSelected.title}</strong><p>以下内容来自项目目录：</p><span>✓ 启动预算：{firstSelected.budget_band}</span><span>✓ 项目难度：{firstSelected.difficulty}</span><span>✓ 资源：{firstSelected.resource_requirements.slice(0, 2).join("、") || "待补充"}</span><Link to={`/projects/${firstSelected.slug}`}>查看项目拆解详情 →</Link></div> : <div className="pm-copilot-bubble assistant"><b>AI</b><p>选择至少两个项目后，这里会展示接口中的项目条件摘要。</p></div>}</>;
  } else {
    conversation = <div className="pm-copilot-bubble assistant"><b>AI</b><p>{variant === "paywall" ? "解锁前请先核对权益、价格和可导出内容。" : "导出文件将按当前匹配会话的服务端记录生成。"}</p></div>;
  }

  return (
    <aside className={`pm-copilot pm-copilot-${variant}`} aria-label="智活 Copilot">
      <header>
        <span aria-hidden="true">✦</span>
        <div>
          <strong aria-label="智活 Copilot">智活 Copilot</strong>
          <small>你的全球 AI 助手，随时为你提供帮助</small>
        </div>
        <Link className="pm-copilot-settings" aria-label="项目超市 Copilot 设置" title="Copilot 设置" to="/assistant/settings">⚙</Link>
        <button
          type="button"
          aria-label="收起项目超市 Copilot"
          title="收起 Copilot"
          onClick={copilotPanel.closePanel}
        >
          <ChevronUp aria-hidden="true" size={18} />
        </button>
      </header>
      <div className="project-copilot-body" id="project-copilot-body">
        <div className="pm-copilot-conversation">{conversation}</div>
        <nav>
          <Link to="/projects/explore">分析市场机会</Link>
          <Link to="/tools/recommend">推荐工具</Link>
          <Link to="/tasks">制定落地计划</Link>
        </nav>
        <MiniCopilotForm className="pm-copilot-input" attachIcon="＋" sendIcon="↗" />
      </div>
    </aside>
  );
}

function PaywallOverlay({ paywallEnabled, source = "results" }: { paywallEnabled: boolean | null; source?: "results" | "detail" }) {
  const { matchId, opportunitySlug, projectRef } = useParams();
  const [searchParams] = useSearchParams();
  const activeProjectRef = projectRef ?? opportunitySlug;
  const section = searchParams.get("section");
  const closeHref = source === "detail" && activeProjectRef
    ? `/projects/${activeProjectRef}${section ? `?section=${encodeURIComponent(section)}` : ""}`
    : matchId ? `/projects/matches/${matchId}/results` : "/projects/results";
  const [plan, setPlan] = useState<MembershipPlanOption | null>(null);
  const [planStatus, setPlanStatus] = useState<"loading" | "ready" | "error">("loading");
  useEffect(() => {
    if (paywallEnabled !== true) return;
    let active = true;
    membershipApi.listPlans().then((payload) => {
      if (!active) return;
      const nextPlan = payload.plans.find((item) => item.recommended) ?? payload.plans.find((item) => item.code !== "free") ?? null;
      setPlan(nextPlan);
      setPlanStatus(nextPlan ? "ready" : "error");
    }).catch(() => {
      if (active) {
        setPlan(null);
        setPlanStatus("error");
      }
    });
    return () => { active = false; };
  }, [paywallEnabled]);
  if (paywallEnabled === null) return null;
  const shellOnly = paywallEnabled === false;
  const shellBenefits = ["完整成功路径", "全部当前数据 + 出处", "案例完整复盘", "报告导出"];
  return createPortal(
    <div className="pm-modal-scrim">
      <section className="pm-paywall-modal" role="dialog" aria-label="解锁完整拆解" aria-modal="true">
        <Link className="pm-modal-close" aria-label="关闭解锁弹层" to={closeHref}>×</Link>
        <header>
          <div><h2>解锁完整拆解</h2><small>unlock the full project path, real cases, and advanced analysis</small><p>查看完整项目路径、真实案例、数据依据和进阶分析。</p></div>
          <div className="pm-paywall-hero-art" aria-hidden="true" />
        </header>
        <div className="pm-paywall-content">
          <section className="pm-plan-comparison" aria-label="会员权益对比">
            <div className="pm-plan-comparison-head"><strong>功能权益</strong><strong>当前账号</strong><strong>{shellOnly ? "即将开放" : plan?.name ?? "会员版"}</strong></div>
            {(shellOnly ? shellBenefits : plan?.features ?? []).slice(0, 6).map((feature) => (
              <div key={feature}><span>{feature}</span><small>{shellOnly ? "当前已开放" : "以账号权限为准"}</small><b>{shellOnly ? "完整可见" : "方案包含"}</b></div>
            ))}
            {!shellOnly && planStatus === "loading" ? <div className="module-empty-state" role="status">正在读取会员方案...</div> : null}
            {!shellOnly && planStatus === "error" ? <div className="module-empty-state" role="alert">会员方案暂不可用，请稍后重试</div> : null}
          </section>
          <aside>
            <header><strong>{shellOnly ? "完整拆解权益" : plan?.name ?? "智活AI会员"}</strong><small>{shellOnly ? "功能预告" : "解锁全部高级内容"}</small></header>
            <p>{shellOnly ? "本期项目内容已全部开放，无需付费即可查看。" : "海量高质量项目 · 深度拆解 · 持续更新"}</p>
            {!shellOnly && plan ? <div className="pm-paywall-price"><b>{`¥${(plan.price_cents / 100).toLocaleString("zh-CN")}`}</b><span>/{plan.billing_cycle === "year" ? "年" : "月"}</span></div> : null}
            {!shellOnly ? <small className="pm-cancel-note">随时可取消 · 未消费额度按方案规则处理</small> : null}
            {shellOnly ? <button className="pm-primary-button" disabled type="button">即将开放</button> : plan ? <Link className="pm-primary-button" to="/membership/upgrade">立即解锁 <span aria-hidden="true">→</span></Link> : <button className="pm-primary-button" disabled type="button">方案暂不可用</button>}
            {!shellOnly ? <Link to="/membership">先看摘要</Link> : null}
          </aside>
        </div>
      </section>
    </div>,
    document.body
  );
}

function ExportOverlay({ paywallEnabled }: { paywallEnabled: boolean }) {
  const { matchId } = useParams();
  const closeHref = matchId ? `/projects/matches/${matchId}/results` : "/projects/results";
  const [sourceID, setSourceID] = useState<number | null>(null);
  const [exportID, setExportID] = useState<number | null>(null);
  const [exportState, setExportState] = useState<ProjectExport | null>(null);
  const [error, setError] = useState("");
  const [exporting, setExporting] = useState(false);
  const [downloading, setDownloading] = useState(false);
  useEffect(() => {
    const routeID = Number(matchId);
    if (Number.isFinite(routeID) && routeID > 0) {
      let active = true;
      setSourceID(null);
      projectsApi.getProjectMatch(routeID).then((session) => {
        if (!active) return;
        if (session.generation?.status !== "completed" && session.generation?.status !== "partial") {
          setError("该匹配记录尚未完成，暂时不能导出");
          return;
        }
        setSourceID(session.match_id);
        setError("");
      }).catch((loadError) => {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取匹配记录"));
      });
      return () => { active = false; };
    }
    if (matchId) {
      setSourceID(null);
      setError("匹配记录不存在");
      return;
    }
    let active = true;
    projectsApi.listMatches().then((payload) => {
      if (!active) return;
      const latest = (payload.matches ?? []).find((item) => item.status === "completed");
      setSourceID(latest?.id ?? null);
      if (!latest) setError("暂无可导出的已完成匹配记录");
    }).catch((loadError) => { if (active) setError(apiErrorMessage(loadError, "暂时无法读取匹配记录")); });
    return () => { active = false; };
  }, [matchId]);
  useEffect(() => {
    if (!exportID || exportState?.status === "ready" || exportState?.status === "failed") return;
    let active = true;
    const timer = window.setInterval(() => {
      projectsApi.getExport(exportID).then((item) => {
        if (!active) return;
        setExportState(item);
        if (item.status === "failed") setError("PDF 生成失败，请重新创建导出");
      }).catch((loadError) => {
        if (active) setError(apiErrorMessage(loadError, "暂时无法读取导出进度"));
      });
    }, 800);
    return () => { active = false; window.clearInterval(timer); };
  }, [exportID, exportState?.status]);
  async function createExport() {
    if (!sourceID || exporting) return;
    setExporting(true); setError("");
    try {
      trackProjectEvent("match_export_click", { match_id: sourceID, plan: paywallEnabled ? "membership_pdf" : "standard_pdf", blocked: false }, "match_export");
      const item = await projectsApi.createExport("match", sourceID);
      setExportID(item.id);
      setExportState(item);
    }
    catch (createError) { setError(apiErrorMessage(createError, "暂时无法导出报告")); }
    finally { setExporting(false); }
  }
  async function downloadExport() {
    if (!exportID || exportState?.status !== "ready" || downloading) return;
    setDownloading(true); setError("");
    try {
      const blob = await projectsApi.downloadExport(exportID);
      saveProjectExport(blob, `project-match-${sourceID ?? exportID}.pdf`);
    } catch (downloadError) {
      setError(apiErrorMessage(downloadError, "暂时无法下载报告"));
    } finally {
      setDownloading(false);
    }
  }
  return createPortal(
    <div className="pm-modal-scrim">
      <section className="pm-export-modal" role="dialog" aria-label="导出匹配报告" aria-modal="true">
        <Link className="pm-modal-close" aria-label="关闭导出弹层" to={closeHref}>×</Link>
        <header><span aria-hidden="true">⇧</span><div><h2>导出匹配报告</h2><p>选择实际支持的格式与报告内容，生成当前匹配会话的服务端快照。</p></div></header>
        <div className="pm-export-layout">
          <section className="pm-export-formats">
            <h3>1. 选择导出格式</h3>
            <button className="active" type="button"><i aria-hidden="true">PDF</i><b>PDF 报告</b><span>服务端排版</span><small>适合归档、打印与分享</small></button>
            <button disabled type="button"><i aria-hidden="true">X</i><b>Excel 表格</b><span>即将开放</span><small>适合详细数据与二次分析</small></button>
            <button disabled type="button"><i aria-hidden="true">↗</i><b>在线分享链接</b><span>即将开放</span><small>生成加密链接，便于在线分享</small></button>
          </section>
          <section className="pm-export-content-options">
            <h3>2. 服务端报告内容 <small>（固定快照）</small></h3>
            {["匹配需求", "项目结果排名", "项目标签与预算", "模型推荐理由", "风险提示", "会话时间"].map((item, index) => (
              <label key={item}><input checked disabled readOnly type="checkbox" /><i aria-hidden="true">{["◎", "▥", "▦", "✓", "!", "◷"][index]}</i><span><strong>{item}</strong><small>来自当前匹配会话的服务端记录</small></span></label>
            ))}
            {paywallEnabled ? <Link to="/membership">会员专享 · 高级导出选项 →</Link> : null}
          </section>
          <article className="pm-export-preview">
            <h3>3. 报告预览</h3>
            <div><strong>智活AI 匹配报告</strong><small>项目匹配分析报告</small><div className="pm-export-preview-art" aria-hidden="true" /></div>
            <p>当前会话 #{sourceID ?? "--"}</p>
            <div className="pm-export-preview-metrics"><span>需求摘要</span><span>匹配排名</span><span>风险提示</span><span>后续行动</span></div>
            <small>页面 preview · 实际导出内容将按服务端记录生成</small>
          </article>
        </div>
        {error ? <p className="form-error" role="alert">{error}</p> : null}
        {exportState && exportState.status !== "ready" && exportState.status !== "failed" ? <p role="status">{exportState.status === "queued" ? "PDF 已进入生成队列..." : "正在排版生成 PDF..."}</p> : null}
        {exportID ? <button className="pm-export-download" disabled={exportState?.status !== "ready" || downloading} onClick={() => void downloadExport()} type="button">{downloading ? "下载中..." : exportState?.status === "ready" ? "下载 PDF 报告" : "等待 PDF 生成"}</button> : null}
        <footer>
          <small>{paywallEnabled ? "付费用户可导出无水印或详细报告，解锁更多专业内容与数据详情。" : "当前项目内容已全部开放，导出文件按服务端记录生成。"}</small>
          <Link to={closeHref}>取消</Link>
          <button disabled={!sourceID || exporting} onClick={() => void createExport()} type="button">{exporting ? "导出中..." : "确认导出"}</button>
        </footer>
      </section>
    </div>,
    document.body
  );
}

export default ProjectsPage;
