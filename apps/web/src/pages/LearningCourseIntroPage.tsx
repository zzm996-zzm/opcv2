import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";

function LearningCourseIntroPage() {
  const navigate = useNavigate();
  const { courseSlug = "ai-market-analysis" } = useParams();
  const [course, setCourse] = useState<LearningCourse | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [starting, setStarting] = useState(false);
  const [startError, setStartError] = useState("");

  useEffect(() => {
    let active = true;
    learningApi.getCourse(courseSlug)
      .then((payload) => { if (active) setCourse(payload); })
      .catch(() => { if (active) setLoadError("课程不存在或暂未发布。"); })
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
  if (!course) return <CourseState title={loadError || "未找到课程"} />;

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
              <p>{course.level} · {course.hours}课时 · {course.learners} 人学习 · {course.price_label}</p>
              <button disabled={starting} onClick={startCourse} type="button">{starting ? "正在开始..." : "开始学习"}</button>
              {startError ? <p role="alert">{startError}</p> : null}
            </div>
          </section>
          <section className="course-info-grid" aria-label="课程信息">
            <article className="course-description-card"><h2>课程简介</h2><p>{course.description}</p></article>
            <article className="course-chapter-card"><h2>课程大纲</h2>{(course.outline ?? []).length ? <ol>{(course.outline ?? []).map((item) => <li key={item}>{item}</li>)}</ol> : <p>当前课程暂未发布大纲。</p>}</article>
          </section>
        </div>
        <aside className="learning-copilot course-intro-copilot" aria-label="智活 Copilot 课程介绍助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>课程介绍说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>课程信息来自已发布目录；当前没有生成 PDF 报告或自动课程推荐。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="课程介绍助手快捷入口"><Link to="/learning/courses">返回课程目录 ›</Link></nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function CourseState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page learning-card" role="status"><h1>{title}</h1><Link to="/learning/courses">返回课程目录</Link></section></V4PageShell>;
}

export default LearningCourseIntroPage;
