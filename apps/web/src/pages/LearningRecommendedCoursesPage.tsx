import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";

const recommendTabs = ["全部", "入门", "实战", "行业", "工具"] as const;
const sortTabs = ["匹配度", "热度", "最新"] as const;

const recommendedCourses = [
  ["基于能力诊断", "智能客服核心能力全景解析", "系统掌握智能客服的技术架构、业务流程与实施能力，建立完整认知框架。", "2.2小时", "18.6k人学习", "会员免费", "headset"],
  ["适合当前项目", "智能客服实战：从需求到落地", "结合实际项目，完成智能客服方案设计、模型选型与落地部署全流程。", "4.6小时", "12.4k人学习", "会员免费", "chat"],
  ["关联任务推荐", "市场分析方法与数据洞察", "掌握市场分析框架与数据洞察方法，为市场策略与产品优化提供数据支撑。", "3.1小时", "22.1k人学习", "会员免费", "bars"],
  ["基于能力诊断", "用户画像与需求挖掘实战", "学会构建精准用户画像，挖掘真实需求，提升客服响应与转化效果。", "2.7小时", "16.9k人学习", "会员免费", "target"],
  ["适合当前项目", "竞品分析与差异化策略", "系统开展竞品分析，发现差异化机会，制定可落地的竞争策略。", "2.9小时", "14.3k人学习", "会员免费", "search"],
  ["关联任务推荐", "数据驱动的市场细分与定位", "基于数据进行市场细分与用户画像定位，明确目标人群与价值主张。", "2.8小时", "11.2k人学习", "会员免费", "cube"],
  ["工具推荐", "ChatGPT 提示词工程实战", "掌握高级提示词技巧，提升AI在客服与市场分析场景中的应用效果。", "1.9小时", "25.7k人学习", "会员免费", "laptop"],
  ["行业专项", "电商行业智能客服最佳实践", "结合电商行业案例，学习智能客服的场景设计与运营优化方法。", "3.4小时", "9.6k人学习", "会员免费", "columns"]
] as const;

function formatLearners(value: number) {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}w人学习`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function toRecommendedCourse(course: LearningCourse, index: number) {
  const icons = ["headset", "chat", "bars", "target", "search", "cube", "laptop", "columns"] as const;
  return [
    `${course.category}推荐`,
    course.title,
    course.description,
    `${course.hours}课时`,
    formatLearners(course.learners),
    course.price_label,
    icons[index % icons.length]
  ] as const;
}

function LearningRecommendedCoursesPage() {
  const [apiCourses, setApiCourses] = useState<LearningCourse[]>([]);

  useEffect(() => {
    let active = true;
    learningApi
      .listCourses({ limit: 8 })
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

  const visibleCourses = apiCourses.length > 0 ? apiCourses.map(toRecommendedCourse) : recommendedCourses;

  return (
    <V4PageShell>
      <section className="learning-page recommended-page" aria-label="推荐课程">
        <div className="recommended-main">
          <section className="recommended-hero">
            <div>
              <div className="diagnosis-breadcrumb">
                <Link to="/learning">AI教学</Link>
                <span>/</span>
                <strong>推荐课程</strong>
              </div>
              <h1>推荐课程</h1>
              <p>基于你当前的项目、能力诊断结果和任务上下文，为你智能推荐最适合的课程，助力能力提升与目标达成。</p>
            </div>
            <div className="recommended-goal">
              <span>当前学习目标</span>
              <strong>提升智能客服与市场分析能力</strong>
              <small>修改目标 ›</small>
              <p>关联项目：智能客服体验升级项目 / 关联任务：市场分析与竞品洞察</p>
            </div>
            <div className="recommended-hero-art" aria-hidden="true" />
          </section>

          <div className="recommended-toolbar" aria-label="推荐课程筛选">
            <div className="course-tabs">
              {recommendTabs.map((tab, index) => (
                <button className={index === 0 ? "active" : ""} key={tab} type="button">{tab}</button>
              ))}
            </div>
            <div className="recommended-sort">
              <span>排序：</span>
              {sortTabs.map((tab, index) => (
                <button className={index === 0 ? "active" : ""} key={tab} type="button">{tab}</button>
              ))}
            </div>
          </div>

          <section className="recommended-grid" aria-label="推荐课程列表">
            {visibleCourses.map(([badge, title, desc, hours, learners, price, icon], index) => (
              <Link className="recommended-course-card" key={title} to={index === 0 ? "/learning/courses/intro" : "/learning/courses/detail"}>
                <div className={`recommended-cover ${icon}`}>
                  <span>{badge}</span>
                  <i aria-hidden="true" />
                  <b aria-hidden="true" />
                </div>
                <div className="recommended-card-copy">
                  <h2>{title}</h2>
                  <p>{desc}</p>
                  <footer>
                    <small>{hours}</small>
                    <small>{learners}</small>
                    <strong>{price}</strong>
                  </footer>
                </div>
              </Link>
            ))}
          </section>

          <p className="recommended-endline"><span /> 已经到底了 <span /></p>
        </div>

        <aside className="learning-copilot recommended-copilot" aria-label="智活 Copilot 推荐课程助手">
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

          <div className="learning-chat recommended-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />这些课程是我根据你的项目、能力诊断和任务上下文为你推荐的。</p>
            </article>
          </div>

          <section className="recommended-logic-panel" aria-label="推荐逻辑说明">
            <h2>推荐逻辑说明</h2>
            <ul>
              <li>基于能力诊断结果，优先补齐短板</li>
              <li>结合当前项目目标，匹配关键技能</li>
              <li>根据关联任务需求，推荐相关知识</li>
              <li>参考学习偏好与历史行为，个性化排序</li>
            </ul>
          </section>

          <section className="recommended-actions-panel" aria-label="推荐课程行动">
            <p>你可以通过以下方式更高效学习 👇</p>
            <Link to="/learning/plan">为我制定学习计划</Link>
            <Link to="/learning/history">查看学习进度</Link>
            <a href="/learning/report" aria-label="能力诊断报告.pdf">
              <i aria-hidden="true">PDF</i>
              <span><b>能力诊断报告.pdf</b><small>PDF · 1.2 MB</small></span>
            </a>
          </section>

          <nav className="learning-copilot-actions" aria-label="推荐课程助手快捷入口">
            <Link to="/learning/recommendation">如何理解推荐逻辑? <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">这些课程的学习顺序是什么? <span aria-hidden="true">›</span></Link>
            <Link to="/learning/courses/detail">如何将课程应用到我的项目中? <span aria-hidden="true">›</span></Link>
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

export default LearningRecommendedCoursesPage;
