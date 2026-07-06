import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";

const categoryTabs = ["全部", "入门", "实战", "行业", "工具"] as const;

const hotCourses = [
  ["01", "AI赋能市场分析实战", "9.8k人在学", "violet"],
  ["02", "提示词工程实战进阶", "8.7k人在学", "cyan"],
  ["03", "AI驱动的增长黑客策略", "7.2k人在学", "blue"]
] as const;

const courses = [
  ["入门", "AI基础入门：从0到1了解AI", "快速掌握 AI 核心概念与应用场景", "18课时", "12.4k人学习", "免费", "ai"],
  ["实战", "提示词工程实战", "掌握高质量提示词设计与优化技巧", "24课时", "8.7k人学习", "会员免费", "chat"],
  ["行业", "智能客服应用案例解析", "打造高效客户服务，提升满意度", "20课时", "6.3k人学习", "会员免费", "headset"],
  ["工具", "AI行业分析方法", "用 AI 洞察市场趋势与竞争格局", "22课时", "9.1k人学习", "会员免费", "bars"],
  ["入门", "大模型原理与能力边界", "理解大模型底层原理与适用边界", "16课时", "5.6k人学习", "免费", "bulb"],
  ["实战", "AI内容创作实战", "高效生成文案、脚本、图片等内容", "22课时", "11.3k人学习", "会员免费", "rocket"],
  ["行业", "AI赋能电商运营", "提升转化率与客户生命周期价值", "20课时", "6.8k人学习", "会员免费", "cart"],
  ["工具", "Python + AI实战入门", "用 Python 调用 AI 接口与自动化", "26课时", "7.4k人学习", "会员免费", "code"],
  ["实战", "数据分析与可视化实战", "从数据洞察到可视化呈现", "21课时", "8.9k人学习", "会员免费", "pie"],
  ["行业", "AI在金融行业的应用", "风控、投研与智能风控案例", "18课时", "4.9k人学习", "会员免费", "bank"],
  ["工具", "AI自动化办公实战", "高效处理文档、表格与邮件", "19课时", "6.1k人学习", "会员免费", "bot"],
  ["行业", "AI医疗行业应用案例", "影像识别、辅助诊断与管理", "17课时", "4.3k人学习", "会员免费", "heart"],
  ["实战", "用户画像与精准营销", "构建用户画像，提升营销效果", "20课时", "6.7k人学习", "会员免费", "target"],
  ["工具", "RAG知识库构建实战", "构建企业专属知识库与问答系统", "24课时", "5.5k人学习", "会员免费", "cloud"],
  ["行业", "AI助力出海增长", "市场洞察、本地化与增长策略", "18课时", "4.8k人学习", "会员免费", "globe"]
] as const;

const recommendedCourses = [
  ["智能客服应用案例解析", "20课时", "6.3k人在学", "行业"],
  ["提示词工程实战进阶", "24课时", "8.7k人在学", "实战"],
  ["AI自动化办公实战", "19课时", "6.1k人在学", "工具"]
] as const;

function formatLearners(value: number) {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}w人学习`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function toCatalogCourse(course: LearningCourse) {
  return [
    course.category,
    course.title,
    course.description,
    `${course.hours}课时`,
    formatLearners(course.learners),
    course.price_label,
    "ai"
  ] as const;
}

function LearningCoursesPage() {
  const [apiCourses, setApiCourses] = useState<LearningCourse[]>([]);

  useEffect(() => {
    let active = true;
    learningApi
      .listCourses({ limit: 20 })
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

  const visibleCourses = apiCourses.length > 0 ? apiCourses.map(toCatalogCourse) : courses;

  return (
    <V4PageShell>
      <section className="learning-page courses-page" aria-label="全部课程">
        <div className="courses-main">
          <header className="courses-header">
            <div>
              <div className="diagnosis-breadcrumb">
                <Link to="/learning">AI教学</Link>
                <span>/</span>
                <strong>全部课程</strong>
              </div>
              <h1>全部课程</h1>
              <p>系统学习 AI 知识与行业应用，提升实战能力</p>
            </div>
            <label className="courses-search">
              <input aria-label="搜索课程" placeholder="搜索课程名称、关键词或讲师" />
              <span aria-hidden="true" />
            </label>
          </header>

          <div className="courses-toolbar" aria-label="课程筛选">
            <div className="course-tabs">
              {categoryTabs.map((tab, index) => (
                <button className={index === 0 ? "active" : ""} key={tab} type="button">{tab}</button>
              ))}
            </div>
            <div className="course-sorters">
              <button type="button">综合排序</button>
              <button type="button">最新上架</button>
              <button type="button">筛选</button>
            </div>
          </div>

          <section className="diagnosis-card hot-courses-card" aria-label="本周热门课程">
            <div>
              <h2>本周热门课程</h2>
              <p>根据学习热度与学员反馈精选</p>
            </div>
            <div className="hot-course-list">
              {hotCourses.map(([rank, title, learners, tone]) => (
                <article className={tone} key={title}>
                  <b>{rank}</b>
                  <i aria-hidden="true" />
                  <div>
                    <strong>{title}</strong>
                    <small>{learners} <span aria-hidden="true">🔥</span></small>
                  </div>
                </article>
              ))}
            </div>
            <Link to="/learning/recommended-courses">查看全部榜单 <span aria-hidden="true">›</span></Link>
          </section>

          <section className="course-catalog-grid" aria-label="课程列表">
            {visibleCourses.map(([category, title, desc, hours, learners, price, icon], index) => (
              <Link className="course-catalog-card" key={title} to={index === 0 ? "/learning/courses/intro" : "/learning/courses/detail"}>
                <div className={`course-cover ${icon}`}>
                  <span>{category}</span>
                  <i aria-hidden="true" />
                </div>
                <div className="course-card-copy">
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

          <nav className="course-pagination" aria-label="课程分页">
            <button type="button" aria-label="上一页">‹</button>
            {[1, 2, 3, 4, 5].map((page) => (
              <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>
            ))}
            <span>...</span>
            <button type="button">20</button>
            <button type="button" aria-label="下一页">›</button>
            <label>
              每页
              <select defaultValue="12" aria-label="每页课程数">
                <option>12</option>
                <option>24</option>
              </select>
            </label>
          </nav>
        </div>

        <aside className="learning-copilot courses-copilot" aria-label="智活 Copilot 课程助手">
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

          <div className="learning-chat courses-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />基于你的项目“智能客服升级”与近期任务，我为你推荐以下课程，助力目标达成。</p>
            </article>
          </div>

          <section className="courses-recommend-panel" aria-label="为你推荐的课程">
            <h2>为你推荐的课程</h2>
            {recommendedCourses.map(([title, hours, learners, tag], index) => (
              <Link key={title} to={index === 0 ? "/learning/courses/intro" : "/learning/courses/detail"}>
                <i className={`course-mini-thumb thumb-${index + 1}`} aria-hidden="true" />
                <div>
                  <strong>{title}</strong>
                  <small>{hours} ｜ {learners}</small>
                </div>
                <span>{tag}</span>
              </Link>
            ))}
          </section>

          <section className="courses-suggestion-panel" aria-label="更多学习建议">
            <h2>更多学习建议</h2>
            <ul>
              <li>强化：提示词工程、数据分析能力</li>
              <li>拓展：RAG知识库构建、自动化工具</li>
              <li>进阶：用户画像与精准营销</li>
            </ul>
          </section>

          <nav className="learning-copilot-actions" aria-label="课程助手快捷入口">
            <Link to="/learning/recommendation">推荐适合我的课程 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">为我制定学习计划 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/history">查看学习进度 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningCoursesPage;
