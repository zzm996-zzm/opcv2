import { useState, type MouseEvent } from "react";
import { Link, useNavigate } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi } from "../lib/learningApi";

const diagnosisSteps = [
  ["1", "收集信息", "获取相关数据与目标"],
  ["2", "能力评估", "多维度能力评估打分"],
  ["3", "差距分析", "定位差距与原因"],
  ["4", "生成报告", "输出诊断结果"],
  ["5", "推荐方案", "推荐学习路径与课程"]
] as const;

const focusTags = ["提升专业能力", "优化工作效率", "拓展业务视野", "职业发展提升"];
const timeTags = ["1-2 小时", "3-5 小时", "5-8 小时", "8 小时以上"];

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
  const [goal, setGoal] = useState(focusTags[0]);
  const [project, setProject] = useState("");
  const [focusAbility, setFocusAbility] = useState("");
  const [weeklyTime, setWeeklyTime] = useState(timeTags[1]);
  const [bottleneck, setBottleneck] = useState("");
  const [isStarting, setIsStarting] = useState(false);
  const [startError, setStartError] = useState("");

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

          <section className="diagnosis-card diagnosis-steps" aria-label="诊断流程">
            {diagnosisSteps.map(([number, title, desc], index) => (
              <article className={index === 0 ? "active" : ""} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card data-source-card" aria-label="诊断依据说明">
            <div className="diagnosis-section-head">
              <h2>诊断依据说明</h2>
            </div>
            <div className="data-source-grid">
              <article>
                <i className="source-icon profile" aria-hidden="true" />
                <div>
                  <h3>本次评估提交</h3>
                  <p>目标、项目、重点能力、投入时间与当前卡点</p>
                  <strong>由你确认后提交</strong>
                </div>
              </article>
              <article>
                <i className="source-icon cube" aria-hidden="true" />
                <div>
                  <h3>已保存用户画像</h3>
                  <p>仅在账户已有可用画像时作为辅助上下文</p>
                  <strong>实际使用来源会写入诊断结果</strong>
                </div>
              </article>
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
