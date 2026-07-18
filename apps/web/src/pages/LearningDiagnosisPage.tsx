import { useEffect, useState, type MouseEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import LearningFlowSteps from "../components/LearningFlowSteps";
import V4PageShell from "../components/V4PageShell";
import { learningApi } from "../lib/learningApi";
import { projectsApi } from "../lib/projectsApi";

const focusTags = ["提升专业能力", "优化工作效率", "拓展业务视野", "职业发展提升"];
const timeTags = ["1-2 小时", "3-5 小时", "5-8 小时", "8 小时以上"];
const defaultProject = "";

const analysisItems = [
  ["AI基础理解", "评估你对AI概念、原理和应用场景的理解程度", "ai"],
  ["提示词工程实战", "评估提示词设计与优化能力及实战应用水平", "chat"],
  ["行业分析方法", "评估市场分析、竞品分析方法掌握程度", "bars"],
  ["智能客服案例拆解", "评估案例理解、方案拆解与优化能力", "headset"],
  ["数据洞察能力", "评估数据收集、分析与洞察发现能力", "pie"],
  ["执行落地能力", "评估方案执行、项目推进与落地能力", "target"]
] as const;

function LearningDiagnosisPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [goal, setGoal] = useState(focusTags[0]);
  const [project, setProject] = useState(defaultProject);
  const [focusAbility, setFocusAbility] = useState("");
  const [weeklyTime, setWeeklyTime] = useState(timeTags[1]);
  const [bottleneck, setBottleneck] = useState("");
  const [isStarting, setIsStarting] = useState(false);
  const [startError, setStartError] = useState("");
  const [projectUpdatedAt, setProjectUpdatedAt] = useState("");
  const [projectSourceState, setProjectSourceState] = useState<"idle" | "loading" | "loaded" | "error">("idle");
  const [projectRefreshKey, setProjectRefreshKey] = useState(0);

  useEffect(() => {
    const opportunitySlug = searchParams.get("project")?.trim();
    if (!opportunitySlug) {
      setProjectSourceState("idle");
      setProjectUpdatedAt("");
      return;
    }

    let active = true;
    setProjectSourceState("loading");
    projectsApi.getOpportunity(opportunitySlug)
      .then((opportunity) => {
        if (active) {
          setProject(opportunity.title);
          setProjectUpdatedAt(opportunity.updated_at ?? "");
          setProjectSourceState("loaded");
        }
      })
      .catch(() => {
        if (active) setProjectSourceState("error");
      });

    return () => {
      active = false;
    };
  }, [projectRefreshKey, searchParams]);

  const projectSourceTime = projectUpdatedAt
    ? new Date(projectUpdatedAt).toLocaleString("zh-CN", { hour12: false })
    : "接口未提供更新时间";

  async function startDiagnosis(event: MouseEvent<HTMLAnchorElement>) {
    event.preventDefault();
    if (isStarting || !goal.trim() || !project.trim()) return;
    setIsStarting(true);
    setStartError("");
    try {
      const diagnosis = await learningApi.submitAssessment({
        goal,
        project,
        focus_abilities: focusAbility ? [focusAbility] : [],
        weekly_time: weeklyTime,
        bottleneck,
        answers: [
          { key: "focus_ability", question: "希望提升的能力", answer: focusAbility || "未指定" },
          { key: "weekly_time", question: "每周可投入时间", answer: weeklyTime },
          ...(bottleneck.trim() ? [{ key: "bottleneck", question: "当前最大卡点", answer: bottleneck.trim() }] : [])
        ]
      });
      navigate("/learning/assessment", { state: { diagnosis } });
    } catch {
      setStartError("诊断启动失败，请稍后重试。");
    } finally {
      setIsStarting(false);
    }
  }

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page learning-diagnosis-entry" aria-label="能力诊断">
        <div className="diagnosis-main">
          <section className="diagnosis-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <strong>能力诊断</strong>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你提交的信息与可用用户画像生成模型评估</p>
            </div>
            <div className="diagnosis-target-art" aria-hidden="true">
              <span />
              <span />
              <span />
            </div>
          </section>

          <LearningFlowSteps active={1} />

          <section className="diagnosis-card data-source-card" aria-label="已接入分析的数据源">
            <div className="diagnosis-section-head">
              <h2>本次评估使用的数据</h2><button disabled={!searchParams.get("project") || projectSourceState === "loading"} onClick={() => setProjectRefreshKey((value) => value + 1)} type="button">↻ 重新读取项目</button>
            </div>
            <div className="data-source-grid">
              <article>
                <i className="source-icon cube" aria-hidden="true" />
                <div>
                  <h3>项目超市</h3><p>{projectSourceState === "loaded" ? "已从项目接口读取" : projectSourceState === "loading" ? "正在读取项目接口" : projectSourceState === "error" ? "项目接口读取失败" : "尚未关联项目"}</p><strong>{project || "请填写目标项目"}</strong><small>{projectSourceState === "loaded" ? `更新时间：${projectSourceTime}` : "仅在关联项目后接入"}</small>
                </div>
              </article>
              <article>
                <i className="source-icon profile" aria-hidden="true" />
                <div>
                  <h3>本次填写</h3><p>使用页面中的真实选择</p><strong>{goal}</strong><small>每周可投入：{weeklyTime}</small>
                </div>
              </article>
              <article><i className="source-icon cube" aria-hidden="true" /><div><h3>能力问答</h3><p>提交后生成诊断依据</p><strong>尚未开始</strong><small>不会预填虚构答题结果</small></div></article>
              <article><i className="source-icon profile" aria-hidden="true" /><div><h3>外部画像</h3><p>本次评估未接入</p><strong>不参与计算</strong><small>结果只使用明确列出的输入</small></div></article>
            </div>
          </section>

          <div className="diagnosis-workbench">
            <section className="diagnosis-card diagnosis-goals" aria-label="补充你的诊断目标">
              <h2>补充你的诊断目标</h2>
              <label className="diagnosis-field">
                <span>目标项目或应用场景</span>
                <input aria-label="目标项目或应用场景" onChange={(event) => setProject(event.target.value)} placeholder="例如：智能客服系统试点" value={project} />
              </label>
              <div className="goal-row">
                <span>目标方向</span>
                <div className="goal-chips">
                  {focusTags.map((tag) => (
                    <button
                      className={goal === tag ? "active" : ""}
                      key={tag}
                      onClick={() => setGoal(tag)}
                      type="button"
                    >
                      {tag}
                    </button>
                  ))}
                </div>
              </div>
              <label className="diagnosis-field">
                <span>希望提升的能力</span>
                <select aria-label="希望提升的能力" onChange={(event) => setFocusAbility(event.target.value)} value={focusAbility}>
                  <option value="" disabled>请选择核心想要提升的能力（可多选）</option>
                  {analysisItems.map(([title]) => <option key={title}>{title}</option>)}
                </select>
              </label>
              <div className="goal-row">
                <span>每周可投入时间</span>
                <div className="goal-chips">
                  {timeTags.map((tag) => (
                    <button
                      className={weeklyTime === tag ? "active" : ""}
                      key={tag}
                      onClick={() => setWeeklyTime(tag)}
                      type="button"
                    >
                      {tag}
                    </button>
                  ))}
                </div>
              </div>
              <label className="diagnosis-field textarea">
                <span>当前最大卡点</span>
                <textarea
                  aria-label="当前最大卡点"
                  maxLength={100}
                  onChange={(event) => setBottleneck(event.target.value)}
                  placeholder="请描述你当前遇到的主要困难或挑战（选填）"
                  value={bottleneck}
                />
                <small>{bottleneck.length}/100</small>
              </label>
            </section>

            <section className="diagnosis-card analysis-scope" aria-label="本次诊断将分析什么">
              <div>
                <h2>本次诊断将分析什么</h2>
                <p>模型会参考以下候选能力，并根据实际输入生成评估维度</p>
              </div>
              <div className="analysis-grid">
                {analysisItems.map(([title, desc, icon]) => (
                  <article key={title}>
                    <i className={`analysis-icon ${icon}`} aria-hidden="true" />
                    <div>
                      <h3>{title}</h3>
                      <p>{desc}</p>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          </div>

          <div className="diagnosis-footer-actions">
            <Link
              aria-disabled={isStarting || !project.trim()}
              className="diagnosis-primary"
              onClick={startDiagnosis}
              to="/learning/assessment"
            >
              {isStarting ? "诊断生成中..." : "提交并生成模型评估"} <span aria-hidden="true">→</span>
            </Link>
            <Link className="diagnosis-secondary" to="/learning">稍后继续补充</Link>
            {startError ? <p role="alert">{startError}</p> : null}
            <p><span aria-hidden="true">♢</span> 结果为模型评估，不代表标准化考试成绩或能力认证。</p>
          </div>
        </div>

        <aside className="learning-copilot diagnosis-copilot" aria-label="智活 Copilot 诊断助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全局 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌃</span>
            </div>
          </header>

          <div className="learning-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>填写真实目标与项目场景后，系统会生成一次可追溯的模型评估快照。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>系统不会声称读取未接入的数据。结果页会列出本次实际使用的输入依据、关键假设和免责声明。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>如果缺少标准化测验或作品证据，分数只能用于安排学习优先级，不能视为客观能力证明。</p>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="诊断助手快捷入口">
            <Link to="/learning/diagnosis">查看诊断逻辑 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/diagnosis">补充诊断信息 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommended-courses">查看学习建议 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningDiagnosisPage;
