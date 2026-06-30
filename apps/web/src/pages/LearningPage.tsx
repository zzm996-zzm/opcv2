import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";

const courseTabs = ["全部", "入门", "实战", "行业", "工具"];

const courses = [
  {
    badge: "入门推荐",
    title: "AI基础入门",
    desc: "快速了解 AI 核心概念与应用场景",
    meta: "入门 · 18课时",
    learners: "12.4k人学习",
    price: "免费",
    tone: "blue"
  },
  {
    badge: "实战推荐",
    title: "提示词工程实战",
    desc: "掌握高质量提示词设计与优化技巧",
    meta: "实战 · 24课时",
    learners: "8.7k人学习",
    price: "会员免费",
    tone: "cyan"
  },
  {
    badge: "行业推荐",
    title: "智能客服应用案例",
    desc: "打造高效客户服务，提升满意度",
    meta: "行业 · 20课时",
    learners: "6.3k人学习",
    price: "会员免费",
    tone: "violet"
  },
  {
    badge: "热门工具",
    title: "AI行业分析方法",
    desc: "用 AI 洞察市场趋势与竞争格局",
    meta: "实战 · 22课时",
    learners: "9.1k人学习",
    price: "会员免费",
    tone: "deep"
  }
];

const learningActions = [
  {
    title: "能力诊断",
    desc: "分析当前能力短板，推荐适合你的课程。",
    button: "开始诊断",
    href: "/learning/diagnosis",
    art: "target"
  },
  {
    title: "系统学习路径",
    desc: "基于目标与薄弱点，生成分阶段学习路径。",
    button: "查看路径",
    href: "/learning/plan",
    tag: "会员专享",
    art: "path"
  },
  {
    title: "学习历史进度",
    desc: "回顾已学课程，查看进度，继续学习。",
    button: "继续学习",
    href: "/learning/history",
    note: "笔记与收藏已整合至学习历史",
    art: "clock"
  }
];

function formatLearners(value: number) {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}w人学习`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function toCourseCard(course: LearningCourse, index: number) {
  const tones = ["blue", "cyan", "violet", "deep"] as const;
  return {
    badge: `${course.category}推荐`,
    title: course.title,
    desc: course.description,
    meta: `${course.category} · ${course.hours}课时`,
    learners: formatLearners(course.learners),
    price: course.price_label,
    tone: tones[index % tones.length]
  };
}

function LearningPage() {
  const [apiCourses, setApiCourses] = useState<LearningCourse[]>([]);

  useEffect(() => {
    let active = true;
    learningApi
      .listCourses({ limit: 4 })
      .then((payload) => {
        if (active) setApiCourses(payload.courses);
      })
      .catch(() => {
        if (active) setApiCourses([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleCourses = apiCourses.length > 0 ? apiCourses.map(toCourseCard) : courses;

  return (
    <V4PageShell>
      <section className="learning-page" aria-label="AI教学首页">
        <div className="learning-main">
          <section className="learning-hero" aria-label="课程学习">
            <div className="learning-hero-copy">
              <h1>课程学习</h1>
              <p>收录丰富 AI 课程，按主题系统学习</p>
            </div>
            <div className="learning-hero-art" aria-hidden="true">
              <span className="learning-cap" />
              <span className="learning-podium" />
              <span className="learning-doc" />
            </div>
          </section>

          <section className="learning-card recommended-courses" aria-label="推荐课程">
            <div className="learning-section-head">
              <div>
                <h2>推荐课程</h2>
                <p>基于你的项目、任务、工具使用与能力诊断，为你智能推荐</p>
              </div>
              <Link to="/learning/recommended-courses">查看全部课程 <span aria-hidden="true">›</span></Link>
            </div>

            <div className="learning-tabs" role="list" aria-label="课程分类">
              {courseTabs.map((tab, index) => (
                <button className={index === 0 ? "active" : ""} key={tab} type="button">{tab}</button>
              ))}
            </div>

            <div className="course-row">
              {visibleCourses.map((course) => (
                <article className="course-card" key={course.title}>
                  <div className={`course-cover ${course.tone}`}>
                    <span>{course.badge}</span>
                    <i aria-hidden="true" />
                  </div>
                  <div className="course-copy">
                    <h3>{course.title}</h3>
                    <p>{course.desc}</p>
                    <footer>
                      <small>{course.meta}</small>
                      <small>{course.learners}</small>
                      <strong>{course.price}</strong>
                    </footer>
                  </div>
                </article>
              ))}
              <button className="course-next" aria-label="查看下一组课程" type="button">›</button>
            </div>
          </section>

          <section className="learning-card learning-functions" aria-label="学习功能">
            <h2>学习功能</h2>
            <div className="learning-function-grid">
              {learningActions.map((item) => (
                <article className="learning-function-card" key={item.title}>
                  <div>
                    <h3>
                      {item.title}
                      {item.tag && <span>{item.tag}</span>}
                    </h3>
                    <p>{item.desc}</p>
                    <Link to={item.href}>{item.button} <span aria-hidden="true">›</span></Link>
                    {item.note && <small>{item.note}</small>}
                  </div>
                  <i className={`learning-function-art ${item.art}`} aria-hidden="true" />
                </article>
              ))}
            </div>
          </section>
        </div>

        <aside className="learning-copilot" aria-label="智活 Copilot 课程助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全球 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌄</span>
            </div>
          </header>

          <div className="learning-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我可以基于你的项目、任务和能力短板，推荐最适合你的 AI 课程和学习路径。</p>
            </article>
            <article className="user">
              <p>我想提升智能客服方向的能力，并了解市场分析方法，有哪些学习建议?</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <div>
                <p>为你规划了智能客服与市场分析的学习路径，建议按以下顺序学习:</p>
                <ol>
                  <li>AI基础入门</li>
                  <li>提示词工程实战</li>
                  <li>智能客服应用案例</li>
                  <li>AI行业分析方法</li>
                </ol>
                <p>已为你生成《能力诊断报告》，可下载查看详细提升建议。</p>
                <span className="learning-file-chip">能力诊断报告.pdf<small>PDF · 1.2 MB</small></span>
              </div>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="课程助手快捷入口">
            <Link to="/learning/recommended-courses">推荐适合我的课程 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">为我制定学习计划 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/history">查看学习进度 <span aria-hidden="true">›</span></Link>
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

export default LearningPage;
