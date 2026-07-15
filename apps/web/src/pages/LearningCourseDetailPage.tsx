import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningCourse, type LearningCourseMaterial } from "../lib/learningApi";
import { referenceCourse, referenceMaterials, referenceProgress } from "../lib/learningReference";

function LearningCourseDetailPage() {
  const { courseSlug = "ai-market-analysis" } = useParams();
  const [course, setCourse] = useState<LearningCourse | null>(null);
  const [materials, setMaterials] = useState<LearningCourseMaterial[]>([]);
  const [percent, setPercent] = useState(0);
  const [lastLesson, setLastLesson] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState("");
  const [progressError, setProgressError] = useState("");

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all([
      learningApi.getCourse(courseSlug),
      learningApi.listCourseMaterials(courseSlug),
      learningApi.getProgress(courseSlug).catch(() => null)
    ]).then(([coursePayload, materialPayload, progress]) => {
      if (!active) return;
      setCourse(coursePayload);
      setMaterials(materialPayload.materials ?? []);
      setPercent(progress?.percent ?? 0);
      setLastLesson(progress?.last_lesson || coursePayload.outline?.[0] || "");
    }).catch(() => {
      if (!active) return;
      const fallbackCourse = referenceCourse(courseSlug);
      const fallbackProgress = referenceProgress.find((item) => item.course_slug === fallbackCourse.slug);
      setCourse(fallbackCourse);
      setMaterials(referenceMaterials.map((item) => ({ ...item, course_slug: fallbackCourse.slug })));
      setPercent(fallbackProgress?.percent ?? 32);
      setLastLesson(fallbackProgress?.last_lesson || fallbackCourse.outline[1] || fallbackCourse.outline[0]);
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [courseSlug]);

  async function saveProgress() {
    if (!course || saving) return;
    setSaving(true);
    setProgressError("");
    setSaveMessage("");
    try {
      const progress = await learningApi.updateProgress(course.slug, { percent, last_lesson: lastLesson, recommended_action: percent === 100 ? "课程已完成" : "继续选择下一课节学习" });
      setPercent(progress.percent);
      setLastLesson(progress.last_lesson);
      setSaveMessage("学习进度已保存");
    } catch {
      setProgressError("学习进度保存失败，请稍后重试。");
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <CourseStudyState title="正在加载课程学习页..." />;
  if (!course) return <CourseStudyState title="未找到课程" />;

  return (
    <V4PageShell>
      <section className="learning-page course-detail-page" aria-label="课程学习">
        <div className="course-detail-main">
          <header className="course-study-header"><div className="diagnosis-breadcrumb"><Link to="/learning/courses">全部课程</Link><span>/</span><Link to={`/learning/courses/${course.slug}`}>{course.title}</Link></div><h1>{course.title}</h1></header>
          <div className="course-study-grid">
            <section className="course-player-card">
              <div className="course-player-visual"><small>当前学习：第2章 行业环境分析</small><h2>2.2 行业生命周期与发展阶段判断</h2><p>理解行业所处生命周期阶段，判断其发展潜力与竞争格局变化趋势。</p></div>
              <h2>{lastLesson || "尚未选择课节"}</h2><p>{course.description}</p>
              <label>学习进度：{percent}%<input aria-label="学习进度百分比" max="100" min="0" onChange={(event) => setPercent(Number(event.target.value))} type="range" value={percent} /></label>
              <label>最近课节<select aria-label="最近学习课节" onChange={(event) => setLastLesson(event.target.value)} value={lastLesson}><option value="">尚未选择</option>{(course.outline ?? []).map((item) => <option key={item}>{item}</option>)}</select></label>
              <button disabled={saving} onClick={saveProgress} type="button">{saving ? "保存中..." : "保存学习进度"}</button>
              {saveMessage ? <p role="status">{saveMessage}</p> : null}{progressError ? <p role="alert">{progressError}</p> : null}
            </section>
            <aside className="course-outline-card" aria-label="课程大纲"><h2>课程大纲</h2>{(course.outline ?? []).length ? <ol>{(course.outline ?? []).map((item) => <li key={item}>{item}</li>)}</ol> : <p>当前课程暂未发布大纲。</p>}</aside>
          </div>
          <section className="course-learning-notes"><article><h2>学习资料</h2>{materials.length === 0 ? <p>当前课程暂无已发布学习资料。</p> : <ul>{materials.map((material) => <li key={material.id}>{material.content_url ? <a href={material.content_url}>{material.title}</a> : material.title}<small>{material.material_type}</small></li>)}</ul>}</article></section>
        </div>
        <aside className="learning-copilot course-study-copilot" aria-label="智活 Copilot 课程学习助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>学习进度说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>请按实际学习情况选择课节和百分比；系统不会自动推断观看时长或掌握程度。</p></article></div>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function CourseStudyState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page learning-card" role="status"><h1>{title}</h1><Link to="/learning/courses">返回课程目录</Link></section></V4PageShell>;
}

export default LearningCourseDetailPage;
