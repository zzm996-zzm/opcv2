import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningProgress } from "../lib/learningApi";

const historyStats = [
  ["已学习课程", "18 门", "较上周 +3", "book"],
  ["学习时长", "42.6 小时", "较上周 +6.2", "clock"],
  ["连续学习天数", "12 天", "最长连续 18 天", "fire"],
  ["完课数", "6 门", "完成率 33%", "badge"]
] as const;

const records = [
  ["AI基础入门", "入门 · 共18课时", "第6章 提示词工程实战", "6.1 提示词优化技巧", "68%", "今天 10:24", "violet"],
  ["提示词工程实战", "实战 · 共24课时", "第7章 复杂场景提示词设计", "7.2 多轮对话策略", "42%", "昨天 15:48", "cyan"],
  ["智能客服应用案例", "行业 · 共20课时", "第4章 客服话术优化", "4.3 情绪识别与应答策略", "55%", "昨天 09:12", "purple"],
  ["AI行业分析方法", "工具 · 共22课时", "第5章 竞品分析模板实战", "5.1 竞品数据采集与清洗", "30%", "前天 20:31", "blue"]
] as const;

const calendarDays = [
  ["12", "done"],
  ["13", "done active"],
  ["14", "done"],
  ["15", "missed"],
  ["16", "today"],
  ["17", "idle"],
  ["18", "idle"]
] as const;

const trendPoints = [24, 14, 18, 13, 14, 24, 28, 33, 34, 27, 34, 43, 41, 46, 40, 48, 38, 35, 34, 31, 18, 28, 36, 30, 39, 34, 42] as const;

function formatUpdatedAt(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function toRecord(progress: LearningProgress, index: number) {
  const tones = ["violet", "cyan", "purple", "blue"] as const;
  return [
    progress.course_title,
    progress.course_slug,
    progress.last_lesson,
    progress.recommended_action || "继续完成下一节课程",
    `${progress.percent}%`,
    formatUpdatedAt(progress.updated_at),
    tones[index % tones.length]
  ] as const;
}

function LearningHistoryPage() {
  const [apiProgress, setApiProgress] = useState<LearningProgress[]>([]);

  useEffect(() => {
    let active = true;
    learningApi
      .listProgress()
      .then((payload) => {
        if (active) setApiProgress(payload.progress);
      })
      .catch(() => {
        if (active) setApiProgress([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleRecords = apiProgress.length > 0 ? apiProgress.map(toRecord) : records;

  return (
    <V4PageShell>
      <section className="learning-page history-page" aria-label="学习历史进度">
        <div className="history-main">
          <header className="history-hero">
            <div>
              <div className="diagnosis-breadcrumb">
                <Link to="/learning">AI教学</Link>
                <span>/</span>
                <strong>学习历史进度</strong>
              </div>
              <h1>学习历史进度</h1>
              <p>跟踪你的学习轨迹，回顾成长每一步</p>
            </div>
            <div className="history-hero-art" aria-hidden="true">
              <i />
              <span />
            </div>
          </header>

          <section className="history-stats-grid" aria-label="学习统计">
            {historyStats.map(([label, value, note, tone]) => (
              <article className={`history-stat ${tone}`} key={label}>
                <i aria-hidden="true" />
                <div>
                  <span>{label}</span>
                  <strong>{value}</strong>
                  <small>{note}</small>
                </div>
              </article>
            ))}
          </section>

          <div className="history-filter-row">
            <div className="history-tabs" aria-label="学习记录状态">
              <button className="active" type="button">进行中（12）</button>
              <button type="button">已完成（6）</button>
            </div>
            <button className="history-filter" type="button">全部课程类型 <span aria-hidden="true">⌄</span></button>
          </div>

          <section className="history-record-card" aria-label="最近学习记录">
            <h2>最近学习记录</h2>
            <div className="history-table">
              <header>
                <span>课程信息</span>
                <span>当前章节</span>
                <span>学习进度</span>
                <span>最近学习时间</span>
                <span>操作</span>
              </header>
              {visibleRecords.map(([title, meta, chapter, lesson, progress, time, tone]) => (
                <article key={title}>
                  <div className="history-course-cell">
                    <i className={tone} aria-hidden="true" />
                    <div>
                      <strong>{title}</strong>
                      <small>{meta}</small>
                    </div>
                  </div>
                  <div className="history-chapter-cell">
                    <strong>{chapter}</strong>
                    <small>{lesson}</small>
                  </div>
                  <div className="history-progress-cell">
                    <strong>{progress}</strong>
                    <span style={{ "--progress": progress } as React.CSSProperties} />
                  </div>
                  <time>{time}</time>
                  <Link to="/learning/courses/detail">继续学习 <span aria-hidden="true">›</span></Link>
                </article>
              ))}
            </div>
            <Link className="history-view-all" to="/learning/courses">查看全部进行中课程 <span aria-hidden="true">⌄</span></Link>
          </section>

          <section className="history-lower-grid">
            <article className="history-calendar-card">
              <header>
                <div>
                  <h2>学习打卡日历</h2>
                  <p>坚持学习，养成习惯</p>
                </div>
                <div className="calendar-tools" aria-hidden="true">
                  <span>‹</span>
                  <span>›</span>
                  <button type="button">回到今天</button>
                </div>
              </header>
              <strong>2025 年 5 月</strong>
              <div className="calendar-week">
                {["一", "二", "三", "四", "五", "六", "日"].map((day) => (
                  <span key={day}>{day}</span>
                ))}
              </div>
              <div className="calendar-days">
                {calendarDays.map(([day, state]) => (
                  <span className={state} key={day}>{day}</span>
                ))}
              </div>
            </article>

            <article className="history-trend-card">
              <h2>连续学习趋势</h2>
              <p>近30天连续学习天数</p>
              <div className="trend-chart" aria-label="近30天连续学习趋势">
                {trendPoints.map((point, index) => (
                  <i key={`${point}-${index}`} style={{ "--point": `${point}%` } as React.CSSProperties} />
                ))}
                <b>12 天</b>
              </div>
              <footer>
                <span>04-17</span>
                <span>04-24</span>
                <span>05-01</span>
                <span>05-08</span>
                <span>05-15</span>
              </footer>
            </article>
          </section>
        </div>

        <aside className="learning-copilot history-copilot" aria-label="智活 Copilot 学习历史助手">
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

          <div className="learning-chat history-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>早上好，张婧！<br />你已经连续学习 12 天啦，继续保持，学习效果更佳哦！</p>
            </article>
          </div>

          <section className="history-advice-card" aria-label="今日学习建议">
            <h2>今日学习建议</h2>
            <strong>建议继续学习</strong>
            <Link to="/learning/courses/detail">
              <i aria-hidden="true" />
              <span><b>AI基础入门 · 第6章</b><small>预计用时 25 分钟</small></span>
            </Link>
            <Link className="history-primary-link" to="/learning/courses/detail">继续学习 <span aria-hidden="true">›</span></Link>
          </section>

          <section className="history-next-card">
            <p>根据你的学习进度，推荐你接下来学习：</p>
            <Link to="/learning/courses/detail">
              <i aria-hidden="true" />
              <span><b>提示词工程实战</b><small>第7章 复杂场景提示词设计</small></span>
            </Link>
            <Link className="history-primary-link" to="/learning/courses/detail">去学习 <span aria-hidden="true">›</span></Link>
          </section>

          <nav className="learning-copilot-actions" aria-label="学习历史助手快捷入口">
            <Link to="/learning/report">查看学习周报 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/report">查看学习时长分析 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">学习计划管理 <span aria-hidden="true">›</span></Link>
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

export default LearningHistoryPage;
