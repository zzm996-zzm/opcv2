import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningRecommendations } from "../lib/learningApi";

const recommendationSteps = [
  ["1", "收集信息", "获取项目信息与目标", "complete"],
  ["2", "能力评估", "多维度能力评估打分", "complete"],
  ["3", "差距分析", "定位差距与原因", "complete"],
  ["4", "生成报告", "输出诊断结果", "complete"],
  ["5", "推荐方案", "推荐学习路径与动作", "active"]
] as const;

const focusDirections = [
  ["智能客服案例拆解", "优先级 高", "服务流程理解与方案拆解能力短板明显", "headset"],
  ["提示词工程实战", "优先级 高", "提示词设计与优化能力需要补强", "chat"],
  ["数据洞察能力", "优先级 中", "数据分析与洞察输出能力可进一步提升", "bars"]
] as const;

const recommendedCourses = [
  {
    badge: "优先推荐",
    title: "智能客服应用案例",
    desc: "打造高效客服应用服务，提升满意度",
    tags: ["案例拆解", "场景实战"],
    learners: "6.3k人学习",
    hours: "20课时",
    tone: "violet"
  },
  {
    badge: "实战推荐",
    title: "提示词工程实战",
    desc: "掌握高质量提示词设计与优化技巧",
    tags: ["提示工程", "实战训练"],
    learners: "8.7k人学习",
    hours: "24课时",
    tone: "cyan"
  },
  {
    badge: "行业推荐",
    title: "AI行业分析方法",
    desc: "用 AI 洞察市场趋势与竞争格局",
    tags: ["行业分析", "方法论"],
    learners: "9.1k人学习",
    hours: "22课时",
    tone: "blue"
  },
  {
    badge: "实验推荐",
    title: "数据洞察与竞品研究实战",
    desc: "从数据中发现机会，输出洞察报告",
    tags: ["数据洞察", "竞品研究"],
    learners: "7.4k人学习",
    hours: "24课时",
    tone: "orange"
  }
] as const;

const studyMethods = [
  ["建议每周学习节奏", "每周 6-8 小时", "建议每周学习 2-3 次", "clock"],
  ["预计完成周期", "3-4 周", "约 24-32 小时学习量", "calendar"],
  ["建议学习顺序", "案例拆解 → 提示词 → 行业分析 → 数据洞察", "先补关键短板，再巩固通用能力", "path"],
  ["学习目标产出", "3 个能力交付物", "完成 1 份实战练习与能力报告", "target"]
] as const;

function LearningRecommendationPage() {
  const [recommendations, setRecommendations] = useState<LearningRecommendations | null>(null);

  useEffect(() => {
    let active = true;
    learningApi
      .getLatestRecommendations()
      .then((payload) => {
        if (active) setRecommendations(payload);
      })
      .catch(() => {
        if (active) setRecommendations(null);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleFocusDirections = recommendations?.focus.length
    ? recommendations.focus.map((item, index) => [
      item.name,
      item.priority === "high" ? "优先级 高" : item.priority === "medium" ? "优先级 中" : "优先级 普通",
      item.summary,
      index === 0 ? "headset" : index === 1 ? "chat" : "bars"
    ] as const)
    : focusDirections;
  const visibleStudyMethods = recommendations?.methods.length
    ? recommendations.methods.map((item, index) => [
      item.title,
      item.value,
      item.detail,
      index === 0 ? "clock" : index === 1 ? "calendar" : index === 2 ? "path" : "target"
    ] as const)
    : studyMethods;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page recommendation-page" aria-label="推荐方案">
        <div className="diagnosis-main">
          <section className="diagnosis-hero recommendation-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <Link to="/learning/diagnosis">能力诊断</Link>
              <span>/</span>
              <strong>推荐方案</strong>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>推荐方案</h1>
              <p>基于诊断结果，为你推荐最值得优先学习的课程与提升动作</p>
            </div>
            <div className="recommendation-hero-art" aria-hidden="true" />
          </section>

          <section className="diagnosis-card diagnosis-steps recommendation-steps" aria-label="诊断流程">
            {recommendationSteps.map(([number, title, desc, state]) => (
              <article className={state} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card recommendation-focus-card" aria-label="优先补强方向">
            <h2><span>1</span> 优先补强方向</h2>
            <div className="recommendation-focus-grid">
              {visibleFocusDirections.map(([title, priority, desc, icon]) => (
                <article key={title}>
                  <i className={`recommendation-icon ${icon}`} aria-hidden="true" />
                  <div>
                    <h3>{title}<small>{priority}</small></h3>
                    <p>{desc}</p>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section className="diagnosis-card recommendation-course-card" aria-label="推荐课程">
            <h2><span>2</span> 推荐课程</h2>
            <div className="recommendation-course-grid">
              {recommendedCourses.map((course) => (
                <article className="recommendation-course" key={course.title}>
                  <div className={`recommendation-course-cover ${course.tone}`}>
                    <strong>{course.badge}</strong>
                    <i aria-hidden="true" />
                  </div>
                  <div className="recommendation-course-copy">
                    <h3>{course.title}</h3>
                    <p>{course.desc}</p>
                    <div>
                      {course.tags.map((tag) => <span key={tag}>{tag}</span>)}
                    </div>
                    <footer>
                      <small>{course.learners}</small>
                      <small>{course.hours}</small>
                    </footer>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section className="diagnosis-card recommendation-method-card" aria-label="建议学习方式">
            <h2><span>3</span> 建议学习方式</h2>
            <div className="recommendation-method-grid">
              {visibleStudyMethods.map(([title, value, desc, icon]) => (
                <article key={title}>
                  <i className={`recommendation-method-icon ${icon}`} aria-hidden="true" />
                  <div>
                    <h3>{title}</h3>
                    <strong>{value}</strong>
                    <p>{desc}</p>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section className="diagnosis-card recommendation-next-card" aria-label="下一步可选动作">
            <div>
              <h2><span>4</span> 下一步可选动作</h2>
            <p>如果你需要系统化学习，可一键生成阶段化学习路径。</p>
            {recommendations?.recommendations.length ? (
              <ul>
                {recommendations.recommendations.map((item) => <li key={item}>{item}</li>)}
              </ul>
            ) : null}
            </div>
            <Link className="recommendation-primary" to="/learning/plan">生成系统学习路径 <span aria-hidden="true">→</span></Link>
            <Link className="recommendation-secondary" to="/learning/courses">查看全部课程</Link>
            <Link className="recommendation-secondary" to="/learning/report">返回诊断报告</Link>
          </section>

          <p className="recommendation-footnote">
            <span aria-hidden="true">i</span> 推荐结果已同步到你的学习计划，可在学习历史进度中持续追踪。
          </p>
        </div>

        <aside className="learning-copilot diagnosis-copilot recommendation-copilot" aria-label="智活 Copilot 推荐助手">
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

          <div className="learning-chat recommendation-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>我已根据你的能力诊断结果，为你整理了优先学习方向和推荐课程。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>建议优先补齐“智能客服案例拆解”和“提示词工程实战”，这样能最快提升你的业务理解与落地能力。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>我已经按先易后难和业务关联度，为你搭配了 4 门课程。你也可以继续让我生成系统学习路径。</p>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="推荐助手快捷入口">
            <Link to="/learning/plan">生成系统学习路径 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommended-courses">调整推荐课程 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/history">查看学习历史 <span aria-hidden="true">›</span></Link>
          </nav>

          <form className="learning-copilot-input">
            <button aria-label="添加附件" type="button">+</button>
            <input aria-label="向 Copilot 提问" placeholder="询问任何问题..." />
            <button aria-label="发送" type="button">⌁</button>
          </form>
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningRecommendationPage;
