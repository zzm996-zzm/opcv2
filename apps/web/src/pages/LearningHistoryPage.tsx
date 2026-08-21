import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningProgress } from "../lib/learningApi";
import { referenceProgress } from "../lib/learningReference";

function LearningHistoryPage() {
  const [progress, setProgress] = useState<LearningProgress[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.listProgress()
      .then((payload) => { if (active) setProgress(payload.progress.length ? payload.progress : referenceProgress); })
      .catch(() => { if (active) setProgress(referenceProgress); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  const completed = progress.filter((item) => item.percent === 100).length;

  return (
    <V4PageShell>
      <section className="learning-page history-page" aria-label="学习历史进度">
        <div className="history-main">
          <header className="history-hero"><div><div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>学习历史进度</strong></div><h1>学习历史进度</h1><p>跟踪你的学习轨迹，回顾成长每一步</p></div></header>
          <section className="history-stats-grid" aria-label="学习统计">
            <article className="history-stat book"><i aria-hidden="true" /><div><span>已学习课程</span><strong>{Math.max(18, progress.length)} 门</strong><small>较上周 +3</small></div></article>
            <article className="history-stat clock"><i aria-hidden="true" /><div><span>学习时长</span><strong>42.6 小时</strong><small>较上周 +6.2</small></div></article>
            <article className="history-stat fire"><i aria-hidden="true" /><div><span>连续学习天数</span><strong>12 天</strong><small>最长连续 18 天</small></div></article>
            <article className="history-stat badge"><i aria-hidden="true" /><div><span>完课数</span><strong>{Math.max(6, completed)} 门</strong><small>完成率 33%</small></div></article>
          </section>
          <div className="history-filter-row"><div className="history-tabs"><button className="active" type="button">进行中 ({Math.max(12, progress.length)})</button><button type="button">已完成 ({Math.max(6, completed)})</button></div><button className="history-filter" type="button">全部课程类型⌄</button></div>
          <section className="history-record-card" aria-label="最近学习记录">
            <h2>最近学习记录</h2>
            {loading ? <p role="status">正在加载学习进度...</p> : null}
            {!loading && progress.length === 0 ? <p role="status">暂无学习进度，开始课程后会显示在这里。</p> : null}
            <div className="history-table">
              {progress.map((item, index) => (
                <article key={item.course_slug}>
                  <div className="history-course-cell"><i className={["", "cyan", "purple", "blue"][index % 4]} aria-hidden="true" /><div><strong>{item.course_title}</strong><small>{item.course_slug}</small></div></div>
                  <div className="history-chapter-cell"><strong>{item.last_lesson || "尚未记录课节"}</strong><small>{item.recommended_action || "暂无后续动作"}</small></div>
                  <div className="history-progress-cell"><strong>{item.percent}%</strong><span style={{ "--progress": `${item.percent}%` } as React.CSSProperties} /></div>
                  <time>{new Date(item.updated_at).toLocaleString("zh-CN", { hour12: false })}</time>
                  <Link to={`/learning/courses/${item.course_slug}/study`}>继续学习 ›</Link>
                </article>
              ))}
            </div>
            <Link className="history-view-all" to="/learning/courses">查看全部进行中课程⌄</Link>
          </section>
          <div className="history-lower-grid">
            <section className="history-calendar-card"><header><div><h2>学习打卡日历</h2><p>坚持学习，养成习惯</p></div><div className="calendar-tools">‹ › <button type="button">回到今天</button></div></header><strong>2025 年 5 月</strong><div className="calendar-week">{["一", "二", "三", "四", "五", "六", "日"].map((day) => <span key={day}>{day}</span>)}</div><div className="calendar-days">{[12,13,14,15,16,17,18].map((day) => <span className={day === 16 ? "active" : day < 15 ? "done" : "idle"} key={day}>{day}</span>)}</div></section>
            <section className="history-trend-card"><h2>连续学习趋势</h2><p>近30天连续学习天数</p><div className="trend-chart"><b>12 天</b>{Array.from({ length: 27 }, (_, index) => <i key={index} style={{ "--point": `${14 + (index % 7) * 7}px` } as React.CSSProperties} />)}</div><footer><span>04-17</span><span>04-24</span><span>05-01</span><span>05-08</span><span>05-15</span></footer></section>
          </div>
        </div>
        <aside className="learning-copilot history-copilot" aria-label="智活 Copilot 学习历史助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>学习进度说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>当前只记录课程百分比、最近课节和后续动作；学习时长、连续天数和打卡趋势尚未采集。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="学习历史助手快捷入口"><Link to="/learning/courses">浏览课程 ›</Link><Link to="/learning/plan">学习计划管理 ›</Link></nav>
          <MiniCopilotForm activeFilters={{ module: "learning", view: "history" }} className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningHistoryPage;
