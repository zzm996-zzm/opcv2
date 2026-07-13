import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse } from "../lib/learningApi";

const categoryTabs = ["全部", "入门", "实战", "行业", "工具"] as const;

function formatLearners(value: number) {
  if (value >= 10000) return `${(value / 10000).toFixed(1)}w人学习`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}k人学习`;
  return `${value}人学习`;
}

function LearningCoursesPage() {
  const [courses, setCourses] = useState<LearningCourse[]>([]);
  const [category, setCategory] = useState("全部");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [copilotOpen, setCopilotOpen] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.listCourses({ limit: 100 })
      .then((payload) => { if (active) setCourses(payload.courses); })
      .catch(() => { if (active) setLoadError("课程目录加载失败，请稍后重试。"); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  const visibleCourses = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase();
    return courses.filter((course) => {
      if (category !== "全部" && course.category !== category) return false;
      if (!normalizedQuery) return true;
      return [course.title, course.description, ...course.tags].some((value) => value.toLocaleLowerCase().includes(normalizedQuery));
    });
  }, [category, courses, query]);

  return (
    <V4PageShell showCopilotMini={false}>
      <section className={`learning-page courses-page ${copilotOpen ? "" : "copilot-collapsed"}`} aria-label="全部课程">
        <div className="courses-main">
          <header className="courses-header"><div><div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>全部课程</strong></div><h1>全部课程</h1><p>当前已发布课程目录</p></div>
            <label className="courses-search"><input aria-label="搜索课程" onChange={(event) => setQuery(event.target.value)} placeholder="搜索课程名称或关键词" value={query} /><span aria-hidden="true" /></label>
          </header>
          <div className="courses-toolbar" aria-label="课程筛选"><div className="course-tabs">{categoryTabs.map((tab) => <button className={category === tab ? "active" : ""} key={tab} onClick={() => setCategory(tab)} type="button">{tab}</button>)}</div></div>
          {loading ? <p role="status">正在加载课程目录...</p> : null}
          {loadError ? <p role="alert">{loadError}</p> : null}
          {!loading && !loadError && visibleCourses.length === 0 ? <p role="status">没有符合条件的课程。</p> : null}
          <section className="course-catalog-grid" aria-label="课程列表">
            {visibleCourses.map((course, index) => (
              <Link className="course-catalog-card" key={course.slug} to={`/learning/courses/${course.slug}`}>
                <div className={`course-cover ${["ai", "chat", "headset", "bars"][index % 4]}`}><span>{course.category}</span><i aria-hidden="true" /></div>
                <div className="course-card-copy"><h2>{course.title}</h2><p>{course.description}</p><footer><small>{course.hours}课时</small><small>{formatLearners(course.learners)}</small><strong>{course.price_label}</strong></footer></div>
              </Link>
            ))}
          </section>
        </div>
        <aside className={`learning-copilot courses-copilot ${copilotOpen ? "" : "collapsed"}`} aria-label="智活 Copilot 课程助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>课程目录说明</p></div><button aria-expanded={copilotOpen} aria-label={copilotOpen ? "收起课程助手" : "展开课程助手"} onClick={() => setCopilotOpen((open) => !open)} type="button">{copilotOpen ? "⌃" : "⌄"}</button></header>
          {copilotOpen ? <><div className="learning-chat"><article><span className="ai-avatar">A</span><p>这里展示全部已发布课程。诊断与课程自动匹配尚未实现，因此不会伪装成个性化推荐。</p></article></div><nav className="learning-copilot-actions" aria-label="课程助手快捷入口"><Link to="/learning/diagnosis">发起能力诊断 ›</Link><Link to="/learning/plan">查看学习计划 ›</Link></nav><MiniCopilotForm className="learning-copilot-input" /></> : null}
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningCoursesPage;
