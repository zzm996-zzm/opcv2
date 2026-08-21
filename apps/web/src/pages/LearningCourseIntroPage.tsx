import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";
import { referenceCourse } from "../lib/learningReference";

function LearningCourseIntroPage() {
  const navigate = useNavigate();
  const { courseSlug = "ai-market-analysis" } = useParams();
  const [course, setCourse] = useState<LearningCourse | null>(null);
  const [loading, setLoading] = useState(true);
  const [starting, setStarting] = useState(false);
  const [startError, setStartError] = useState("");

  useEffect(() => {
    let active = true;
    learningApi.getCourse(courseSlug)
      .then((payload) => { if (active) setCourse(payload); })
      .catch(() => { if (active) setCourse(referenceCourse(courseSlug)); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [courseSlug]);

  async function startCourse() {
    if (!course || starting) return;
    setStarting(true);
    setStartError("");
    try {
      const existing = await learningApi.getProgress(course.slug).catch(() => null);
      if (!existing) {
        await learningApi.updateProgress(course.slug, { percent: 0, last_lesson: course.outline?.[0] || "", recommended_action: "选择课节并记录学习进度" });
      }
      navigate(`/learning/courses/${course.slug}/study`);
    } catch {
      setStartError("课程启动失败，请稍后重试。");
    } finally {
      setStarting(false);
    }
  }

  if (loading) return <CourseState title="正在加载课程..." />;
  if (!course) return <CourseState title="未找到课程" />;

  return (
    <V4PageShell>
      <section className="learning-page course-intro-page" aria-label="课程介绍">
        <div className="course-intro-main">
          <section className="course-intro-hero">
            <div className="course-intro-cover" aria-label={`${course.title}课程封面`} />
            <div className="course-intro-summary">
              <div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><Link to="/learning/courses">全部课程</Link></div>
              <span>{course.category}</span><h1>{course.title}</h1><p>{course.description}</p>
              <div className="course-intro-tags">{(course.tags ?? []).map((tag) => <span key={tag}>{tag}</span>)}</div>
              <p>{course.level} · {course.hours}课时 · {course.learners >= 1000 ? `${(course.learners / 1000).toFixed(1)}k` : course.learners} 人学习 · {course.price_label}</p>
              <button disabled={starting} onClick={startCourse} type="button">{starting ? "正在开始..." : "开始学习"}</button>
              {startError ? <p role="alert">{startError}</p> : null}
            </div>
          </section>
          <section className="course-info-grid" aria-label="课程信息">
            <article className="course-description-card"><h2>课程简介</h2><p>本课程将带你掌握 AI 驱动的行业分析全流程方法论，从数据获取、清洗分析到趋势洞察与竞争格局研判，帮助你快速上手行业研究，输出高价值分析报告，为业务决策提供有力支撑。</p><dl><div><dt>课程时长</dt><dd>2小时18分钟</dd></div><div><dt>学习难度</dt><dd>{course.level}</dd></div><div><dt>更新日期</dt><dd>2024-05-20</dd></div><div><dt>课程语言</dt><dd>中文</dd></div></dl></article>
            <article className="course-learn-card"><h2>你将学到什么</h2><ul>{["掌握 AI 行业分析的整体框架与关键步骤", "学会使用 AI 工具进行行业数据的获取与清洗", "掌握市场规模、增长趋势与驱动因素分析方法", "学会竞争格局与头部企业分析的实战技巧", "输出结构化行业分析报告，支撑业务决策"].map((item) => <li key={item}>✓ {item}</li>)}</ul></article>
            <article className="course-goal-card"><h2>与你的目标强相关</h2><strong>智能客服与市场分析能力提升</strong><p>目标进度 <b>35%</b></p><span><i style={{ width: "35%" }} /></span><Link to="/learning/plan">查看目标详情</Link></article>
          </section>
          <section className="course-intro-lower"><article className="course-chapter-card"><h2>章节目录 <small>共{course.outline.length}节</small></h2>{(course.outline ?? []).length ? <ol>{(course.outline ?? []).map((item, index) => <li key={item}><span>第{index + 1}章</span>{item}<small>{12 + index * 2}:35</small></li>)}</ol> : <p>当前课程暂未发布大纲。</p>}</article><article><h2>适合人群</h2><ul><li>产品经理</li><li>市场分析师</li><li>运营经理</li><li>创业者与企业管理者</li></ul></article><article><h2>相关课程推荐</h2><Link to="/learning/courses/prompt-engineering">提示词工程实战</Link><Link to="/learning/courses/customer-service-cases">智能客服应用案例</Link><Link to="/learning/courses/data-visualization">数据可视化与报告呈现</Link></article></section>
        </div>
        <aside className="learning-copilot course-intro-copilot" aria-label="智活 Copilot 课程介绍助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>课程介绍说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>课程信息来自已发布目录；当前没有生成 PDF 报告或自动课程推荐。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="课程介绍助手快捷入口"><Link to="/learning/courses">返回课程目录 ›</Link></nav>
          <MiniCopilotForm activeFilters={{ module: "learning", view: "course_intro", course_slug: course.slug }} className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function CourseState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page learning-card" role="status"><h1>{title}</h1><Link to="/learning/courses">返回课程目录</Link></section></V4PageShell>;
}

export default LearningCourseIntroPage;
