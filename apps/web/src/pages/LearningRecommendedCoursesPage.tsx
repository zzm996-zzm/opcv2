import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";
import { referenceCourses } from "../lib/learningReference";

function formatLearners(value: number) {
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function LearningRecommendedCoursesPage() {
  const [courses, setCourses] = useState<LearningCourse[]>([]);
  const [loading, setLoading] = useState(true);
  const [copilotOpen, setCopilotOpen] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.listCourses({ limit: 20 })
      .then((payload) => { if (active) setCourses(payload.courses.length ? payload.courses : referenceCourses.slice(0, 8)); })
      .catch(() => { if (active) setCourses(referenceCourses.slice(0, 8)); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  return (
    <V4PageShell showCopilotMini={false}>
      <section className={`learning-page recommended-page ${copilotOpen ? "" : "copilot-collapsed"}`} aria-label="课程选择">
        <div className="recommended-main">
          <header className="recommended-header"><div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>推荐课程</strong></div><h1>推荐课程</h1><p>基于你当前的项目、能力诊断结果和任务上下文，为你智能推荐最适合的课程。</p></header>
          {loading ? <p role="status">正在加载课程...</p> : null}
          {!loading && courses.length === 0 ? <p role="status">当前暂无已发布课程。</p> : null}
          <section className="recommended-course-grid" aria-label="已发布课程">
            {courses.map((course, index) => <Link className="recommended-course-card" key={course.slug} to={`/learning/courses/${course.slug}`}><div className={`recommended-cover thumb-${index % 4 + 1}`} style={{ backgroundImage: `url(/learning/course-${String(index % 15 + 1).padStart(2, "0")}.jpg)` }}><span>{course.category}</span></div><div className="recommended-card-copy"><h2>{course.title}</h2><p>{course.description}</p><footer><small>{course.hours}课时</small><small>{formatLearners(course.learners)}</small><strong>{course.price_label}</strong></footer></div></Link>)}
          </section>
        </div>
        <aside className={`learning-copilot recommended-copilot ${copilotOpen ? "" : "collapsed"}`} aria-label="智活 Copilot 课程助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>课程选择说明</p></div><button aria-expanded={copilotOpen} aria-label={copilotOpen ? "收起推荐课程助手" : "展开推荐课程助手"} onClick={() => setCopilotOpen((open) => !open)} type="button">{copilotOpen ? "⌃" : "⌄"}</button></header>
          {copilotOpen ? <><div className="learning-chat"><article><span className="ai-avatar">A</span><p>当前诊断建议与课程目录还没有可验证的自动匹配规则，请按课程标题、简介和标签自行核对。</p></article></div><nav className="learning-copilot-actions" aria-label="课程助手快捷入口"><Link to="/learning/recommendation">查看诊断学习建议 ›</Link><Link to="/learning/plan">查看学习路径 ›</Link></nav><MiniCopilotForm className="learning-copilot-input" /></> : null}
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningRecommendedCoursesPage;
