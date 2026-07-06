import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";

const outlineSections = [
  {
    chapter: "第1章",
    count: "3/3",
    title: "行业分析概述",
    lessons: [],
    tone: "done"
  },
  {
    chapter: "第2章",
    count: "2/4",
    title: "行业环境分析",
    lessons: [
      ["2.1 宏观环境分析（PEST分析）", "08:12", "已完成", "done"],
      ["2.2 行业生命周期与发展阶段判断", "12:45", "学习中", "active"],
      ["2.3 行业规模与增长趋势分析", "未开始", "", "idle"],
      ["2.4 产业链结构与价值分布", "未开始", "", "idle"]
    ],
    tone: "active"
  },
  {
    chapter: "第3章",
    count: "0/4",
    title: "市场竞争格局分析",
    lessons: [],
    tone: "locked"
  },
  {
    chapter: "第4章",
    count: "0/3",
    title: "目标市场与客户分析",
    lessons: [],
    tone: "locked"
  },
  {
    chapter: "第5章",
    count: "0/2",
    title: "行业机会与风险识别",
    lessons: [],
    tone: "locked"
  },
  {
    chapter: "第6章",
    count: "0/2",
    title: "行业分析报告输出",
    lessons: [],
    tone: "locked"
  }
] as const;

const keyPoints = [
  "行业生命周期的四个阶段特征与判断标准",
  "关键指标：市场增长率、集中度、技术变革、利润水平",
  "不同阶段的竞争格局与企业战略重点"
] as const;

const materials = [
  ["PPT", "行业生命周期分析框架.pptx", "PPTX · 2.4 MB"],
  ["PDF", "行业发展阶段判断案例.pdf", "PDF · 1.8 MB"],
  ["XLS", "行业关键指标参考模板.xlsx", "XLSX · 3.1 MB"]
] as const;

function LearningCourseDetailPage() {
  return (
    <V4PageShell>
      <section className="learning-page course-detail-page" aria-label="课程学习">
        <div className="course-detail-main">
          <div className="diagnosis-breadcrumb">
            <Link to="/learning">AI教学</Link>
            <span>/</span>
            <Link to="/learning/courses/intro">AI行业分析方法</Link>
            <span>/</span>
            <strong>课程学习</strong>
          </div>

          <section className="course-study-shell">
            <article className="course-player-card">
              <div className="player-copy">
                <span>当前学习： 第2章 行业环境分析</span>
                <h1>2.2 行业生命周期与发展阶段判断</h1>
                <p>理解行业所处生命周期阶段，判断其发展潜力与竞争格局变化趋势。</p>
              </div>

              <div className="course-video-frame" aria-label="课程视频播放器">
                <i className="video-wave" aria-hidden="true" />
                <i className="video-stage" aria-hidden="true" />
                <i className="video-bars" aria-hidden="true" />
                <i className="video-line" aria-hidden="true" />
                <div className="video-controls" aria-label="视频播放控制">
                  <button type="button" aria-label="播放">▶</button>
                  <button type="button" aria-label="音量">▮</button>
                  <span>06:38 / 12:45</span>
                  <b>1.25x</b>
                  <b>标清</b>
                  <button type="button" aria-label="全屏">⌗</button>
                </div>
              </div>

              <footer className="course-progress-row">
                <strong>课程总进度</strong>
                <progress max="100" value="32">32%</progress>
                <b>32%</b>
                <span>已学 6 / 18 节</span>
                <button type="button">继续下一节</button>
              </footer>
            </article>

            <aside className="course-outline-card" aria-label="课程大纲">
              <h2>课程大纲</h2>
              <div className="outline-list">
                {outlineSections.map((section) => (
                  <section className={section.tone} key={section.title}>
                    <header>
                      <span>{section.chapter}</span>
                      <strong>{section.title}</strong>
                      <b>{section.count}</b>
                      <i aria-hidden="true">{section.tone === "active" ? "⌃" : "›"}</i>
                    </header>
                    {section.lessons.length > 0 && (
                      <div className="outline-lessons">
                        {section.lessons.map(([title, time, state, tone]) => (
                          <article className={tone} key={title}>
                            <i aria-hidden="true" />
                            <div>
                              <strong>{title}</strong>
                              <small>{time} {state && <span>| {state}</span>}</small>
                            </div>
                          </article>
                        ))}
                      </div>
                    )}
                  </section>
                ))}
              </div>
            </aside>
          </section>

          <section className="course-study-lower">
            <article className="course-study-card lesson-summary-card">
              <h2>本节内容摘要</h2>
              <p>本节介绍行业生命周期的四个阶段：导入期、成长期、成熟期和衰退期，帮助你从市场规模、增长率、竞争强度、客户需求等维度判断行业所处阶段，识别阶段特征与关键指标，为后续策略制定提供依据。</p>
              <div className="lesson-keypoint-box">
                <div>
                  <strong>核心知识点</strong>
                  <ul>
                    {keyPoints.map((item) => (
                      <li key={item}>{item}</li>
                    ))}
                  </ul>
                </div>
                <i aria-hidden="true" />
              </div>
            </article>

            <article className="course-study-card material-card">
              <h2>学习资料</h2>
              <div className="material-list">
                {materials.map(([type, title, meta]) => (
                  <a href="/learning/courses/detail" key={title}>
                    <i className={type.toLowerCase()} aria-hidden="true">{type}</i>
                    <span><b>{title}</b><small>{meta}</small></span>
                    <strong aria-hidden="true">↓</strong>
                  </a>
                ))}
              </div>
              <Link to="/learning/courses/detail">查看全部资料（3）</Link>
            </article>
          </section>
        </div>

        <aside className="learning-copilot course-study-copilot" aria-label="智活 Copilot 课程学习助手">
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

          <div className="learning-chat course-study-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我已定位到你正在学习的内容：第2章 行业环境分析 · 2.2 行业生命周期与发展阶段判断</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>需要我为你总结本节要点吗？</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>好的，以下是本节核心要点总结：<br />· 行业生命周期分为导入期、成长期、成熟期和衰退期；<br />· 可通过市场增长率、集中度、技术变革等指标判断阶段；<br />· 不同阶段的竞争格局与企业战略重点存在显著差异。</p>
            </article>
          </div>

          <section className="next-lesson-card">
            <p>建议你接下来学习：</p>
            <strong>2.3 行业规模与增长趋势分析</strong>
            <span>通过数据分析行业规模、增长驱动因素，进一步验证行业潜力。</span>
            <Link to="/learning/courses/detail">去学习下一节 <b aria-hidden="true">›</b></Link>
          </section>

          <nav className="learning-copilot-actions" aria-label="课程学习助手快捷入口">
            <Link to="/learning/courses/detail">总结本节关键知识点 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/courses/detail">生成学习笔记 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommended-courses">推荐相关案例学习 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningCourseDetailPage;
