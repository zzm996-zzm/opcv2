import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";
import { referenceCourses } from "../lib/learningReference";

const learningActions = [
  ["能力诊断", "提交目标和项目上下文，生成可追溯的模型评估。", "开始诊断", "/learning/diagnosis"],
  ["系统学习路径", "查看诊断快照生成的阶段计划并记录完成状态。", "查看路径", "/learning/plan"],
  ["学习历史进度", "查看已经持久化的课程学习进度。", "查看进度", "/learning/history"]
] as const;

function formatLearners(value: number) {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}w人学习`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function LearningPage() {
  const [courses, setCourses] = useState<LearningCourse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.listCourses({ limit: 4 })
      .then((payload) => { if (active) setCourses(payload.courses.length ? payload.courses : referenceCourses.slice(0, 4)); })
      .catch(() => { if (active) setCourses(referenceCourses.slice(0, 4)); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  return (
    <V4PageShell>
      <section className="learning-page" aria-label="AI教学首页">
        <div className="learning-main">
          <section className="learning-hero" aria-label="课程学习">
            <div className="learning-hero-copy"><h1>课程学习</h1><p>收录丰富 AI 课程，按主题系统学习</p></div>
            <div className="learning-hero-art" aria-hidden="true" />
          </section>

          <section className="learning-card recommended-courses" aria-label="课程目录预览">
            <div className="learning-section-head"><div><h2>推荐课程</h2><p>基于你的项目、任务、工具使用与能力诊断，为你智能推荐</p></div><Link to="/learning/courses">查看全部课程 ›</Link></div>
            {loading ? <p role="status">正在加载课程...</p> : null}
            {!loading && courses.length === 0 ? <p role="status">当前暂无已发布课程。</p> : null}
            <div className="course-row">
              {courses.map((course, index) => (
                <Link className="course-card" key={course.slug} to={`/learning/courses/${course.slug}`}>
                  <div className={`course-cover ${["blue", "cyan", "violet", "deep"][index % 4]}`} style={{ backgroundImage: `url(/learning/course-0${index + 1}.jpg)` }}><span>{course.category}</span><i aria-hidden="true" /></div>
                  <div className="course-copy"><h3>{course.title}</h3><p>{course.description}</p><footer><small>{course.category} · {course.hours}课时</small><small>{formatLearners(course.learners)}</small><strong>{course.price_label}</strong></footer></div>
                </Link>
              ))}
            </div>
          </section>

          <section className="learning-card learning-functions" aria-label="学习功能">
            <h2>学习功能</h2><div className="learning-function-grid">
              {learningActions.map(([title, description, action, href]) => <article className="learning-function-card" key={title}><div><h3>{title}</h3><p>{description}</p><Link to={href}>{action} ›</Link></div></article>)}
            </div>
          </section>
        </div>

        <aside className="learning-copilot" aria-label="智活 Copilot 课程助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>AI教学使用说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>课程目录与学习进度来自已保存记录；完成诊断后，页面会明确显示模型假设和数据依据。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="课程助手快捷入口"><Link to="/learning/diagnosis">发起能力诊断 ›</Link><Link to="/learning/plan">查看学习计划 ›</Link><Link to="/learning/history">查看学习进度 ›</Link></nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningPage;
