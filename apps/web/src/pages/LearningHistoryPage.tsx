import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningProgress } from "../lib/learningApi";

function LearningHistoryPage() {
  const [progress, setProgress] = useState<LearningProgress[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let active = true;
    learningApi.listProgress()
      .then((payload) => { if (active) setProgress(payload.progress); })
      .catch(() => { if (active) setLoadError("学习进度加载失败，请稍后重试。"); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  const completed = progress.filter((item) => item.percent === 100).length;

  return (
    <V4PageShell>
      <section className="learning-page history-page" aria-label="学习历史进度">
        <div className="history-main">
          <header className="history-hero"><div><div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>学习历史进度</strong></div><h1>学习历史进度</h1><p>仅展示已持久化的课程进度</p></div></header>
          <section className="history-stats-grid" aria-label="学习统计">
            <article className="history-stat book"><div><span>已有进度课程</span><strong>{progress.length} 门</strong></div></article>
            <article className="history-stat badge"><div><span>已完成课程</span><strong>{completed} 门</strong></div></article>
          </section>
          <section className="history-record-card" aria-label="最近学习记录">
            <h2>最近学习记录</h2>
            {loading ? <p role="status">正在加载学习进度...</p> : null}
            {loadError ? <p role="alert">{loadError}</p> : null}
            {!loading && !loadError && progress.length === 0 ? <p role="status">暂无学习进度，开始课程后会显示在这里。</p> : null}
            <div className="history-table">
              {progress.map((item) => (
                <article key={item.course_slug}>
                  <div className="history-course-cell"><div><strong>{item.course_title}</strong><small>{item.course_slug}</small></div></div>
                  <div className="history-chapter-cell"><strong>{item.last_lesson || "尚未记录课节"}</strong><small>{item.recommended_action || "暂无后续动作"}</small></div>
                  <div className="history-progress-cell"><strong>{item.percent}%</strong><span style={{ "--progress": `${item.percent}%` } as React.CSSProperties} /></div>
                  <time>{new Date(item.updated_at).toLocaleString("zh-CN", { hour12: false })}</time>
                  <Link to={`/learning/courses/${item.course_slug}/study`}>继续学习 ›</Link>
                </article>
              ))}
            </div>
          </section>
        </div>
        <aside className="learning-copilot history-copilot" aria-label="智活 Copilot 学习历史助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>学习进度说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>当前只记录课程百分比、最近课节和后续动作；学习时长、连续天数和打卡趋势尚未采集。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="学习历史助手快捷入口"><Link to="/learning/courses">浏览课程 ›</Link><Link to="/learning/plan">学习计划管理 ›</Link></nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningHistoryPage;
