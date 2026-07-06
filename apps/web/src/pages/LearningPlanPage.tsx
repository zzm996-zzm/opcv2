import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningPlan, type LearningPlanStage } from "../lib/learningApi";

const learningStages = [
  {
    number: "1",
    title: "AI基础认知",
    subtitle: "建立AI思维与基础能力",
    status: "已完成 35%",
    state: "done",
    courses: ["AI基础入门", "AI能力地图与应用场景"],
    duration: "4.5 小时",
    goal: "理解AI基本概念与能力边界，建立用AI解决问题的思维框架。",
    milestone: "完成AI基础测验"
  },
  {
    number: "2",
    title: "提示词与工具实操",
    subtitle: "掌握提问技巧与实战工具",
    status: "进行中 20%",
    state: "current",
    courses: ["提示词工程实战", "AI工具箱实践指南"],
    duration: "6.5 小时",
    goal: "掌握高质量提示词编写方法，熟练使用常用AI工具完成任务。",
    milestone: "完成提示词实战任务"
  },
  {
    number: "3",
    title: "行业分析方法",
    subtitle: "学会用AI做市场与竞品分析",
    status: "未开始 0%",
    state: "pending violet",
    courses: ["AI行业分析方法", "竞品全盘数据破解实战"],
    duration: "7.0 小时",
    goal: "掌握市场与竞品分析方法，输出可落地的分析报告。",
    milestone: "提交行业分析报告"
  },
  {
    number: "4",
    title: "智能客服应用案例",
    subtitle: "落地客服场景与优化运营",
    status: "未开始 0%",
    state: "pending blue",
    courses: ["智能客服应用案例", "客服数据分析与优化"],
    duration: "6.0 小时",
    goal: "学会梳理与优化智能客服方案，提升客户满意度与转化率。",
    milestone: "完成客服方案优化项目"
  }
] as const;

const planMetrics = [
  ["预计总学习时长", "24 小时", "约 3-4 周完成", "clock"],
  ["每周学习节奏建议", "每周 6-8 小时", "建议每周学习 2-3 次", "calendar"],
  ["学习目标产出", "掌握AI工具与提示词实战能力", "独立完成行业与竞品分析报告｜落地智能客服优化方案", "target"]
] as const;

function stageState(stage: LearningPlanStage, index: number) {
  if (stage.status.includes("完成")) return "done";
  if (stage.status.includes("进行")) return "current";
  return index % 2 === 0 ? "pending violet" : "pending blue";
}

function LearningPlanPage() {
  const [plan, setPlan] = useState<LearningPlan | null>(null);

  useEffect(() => {
    let active = true;
    learningApi
      .getLatestPlan()
      .then((payload) => {
        if (active) setPlan(payload);
      })
      .catch(() => {
        if (active) setPlan(null);
      });
    return () => {
      active = false;
    };
  }, []);

  const pathTitle = plan?.title ?? "系统学习路径";
  const pathDescription = plan?.description ?? "基于你的项目方向与能力诊断结果，为你量身定制学习路径，助你高效掌握“智能客服与市场分析”相关能力。";
  const visibleRecommendations = plan?.recommendations.length ? plan.recommendations : [];
  const visibleStages = plan?.stages.length
    ? plan.stages.map((stage, index) => ({
      number: String(stage.number),
      title: stage.title,
      subtitle: stage.goal,
      status: stage.status,
      state: stageState(stage, index),
      courses: stage.courses,
      duration: stage.duration,
      goal: stage.goal,
      milestone: stage.milestone
    }))
    : learningStages;
  const visiblePlanMetrics = plan
    ? [
      ["预计总学习时长", `${plan.estimated_hours} 小时`, "约 3-4 周完成", "clock"],
      ["每周学习节奏建议", plan.weekly_suggestion, "建议每周学习 2-3 次", "calendar"],
      ["学习目标产出", "掌握诊断推荐能力", "完成阶段练习与能力报告", "target"]
    ] as const
    : planMetrics;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page learning-plan-page" aria-label="系统学习路径">
        <div className="diagnosis-main learning-plan-main">
          <section className="diagnosis-hero learning-plan-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <strong>系统学习路径</strong>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>{pathTitle}</h1>
              <span>张婧的专属学习路径</span>
              <p>{pathDescription}</p>
            </div>
            <div className="learning-plan-hero-art" aria-hidden="true">
              <span />
              <span />
              <span />
            </div>
          </section>

          <section className="diagnosis-card learning-path-card" aria-label="四阶段学习路径">
            {visibleRecommendations.length > 0 ? (
              <header>
                <h2>诊断推荐行动</h2>
                <ul>
                  {visibleRecommendations.map((recommendation) => (
                    <li key={recommendation}>{recommendation}</li>
                  ))}
                </ul>
              </header>
            ) : null}
            <div className="learning-path-line" aria-hidden="true" />
            <div className="learning-path-grid">
              {visibleStages.map((stage) => (
                <article className={`learning-stage ${stage.state}`} key={stage.title}>
                  <span className="learning-stage-number">{stage.number}</span>
                  <header>
                    <div>
                      <h2>{stage.title}</h2>
                      <p>{stage.subtitle}</p>
                    </div>
                    <strong>{stage.status}</strong>
                  </header>
                  <div className="learning-stage-body">
                    <section>
                      <h3>推荐课程</h3>
                      <ul>
                        {stage.courses.map((course) => <li key={course}>{course}</li>)}
                      </ul>
                      <small>共2门课程</small>
                    </section>
                    <section>
                      <h3>预计时长</h3>
                      <b>{stage.duration}</b>
                    </section>
                    <section>
                      <h3>阶段目标</h3>
                      <p>{stage.goal}</p>
                    </section>
                  </div>
                  <footer>
                    <b>里程碑</b>
                    <span>{stage.milestone}</span>
                    <i aria-label={`${stage.title} 阶段状态`} />
                  </footer>
                </article>
              ))}
            </div>
          </section>

          <section className="diagnosis-card learning-plan-summary" aria-label="学习路径摘要">
            {visiblePlanMetrics.map(([title, value, desc, icon]) => (
              <article className={icon} key={title}>
                <i aria-hidden="true" />
                <div>
                  <h2>{title}</h2>
                  <strong>{value}</strong>
                  <p>{desc}</p>
                </div>
              </article>
            ))}
          </section>

          <footer className="learning-plan-footer">
            <p>路径将根据你的学习进度与测评结果动态调整，保持学习效果最优。</p>
            <span>上次更新：{plan ? new Date(plan.generated_at).toLocaleString("zh-CN", { hour12: false }) : "2024-05-20 10:30"}</span>
            <button type="button">刷新路径</button>
          </footer>
        </div>

        <aside className="learning-copilot diagnosis-copilot learning-plan-copilot" aria-label="智活 Copilot 学习路径助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全球 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌃</span>
            </div>
          </header>

          <div className="learning-chat learning-plan-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我已根据你的能力诊断结果，为你生成了专属学习路径。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p><b>为什么生成这条路径？</b><br />✓ 基于你当前的能力诊断：AI基础较弱，提示词与行业分析能力待提升。<br />✓ 结合你的项目方向：聚焦智能客服与市场分析场景的能力构建。<br />✓ 按由浅入深、学以致用的顺序设计，帮助你快速落地并产生业务价值。</p>
            </article>
          </div>

          <div className="learning-plan-copilot-panel">
            <strong>接下来你可以：</strong>
            <Link to="/learning/recommendation">调整路径</Link>
            <Link to="/learning/courses/intro">开始第一阶段</Link>
          </div>

          <nav className="learning-copilot-actions" aria-label="学习路径助手快捷入口">
            <Link to="/learning/recommended-courses">推荐适合我的课程 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">为我制定学习计划 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/history">查看学习进度 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningPlanPage;
