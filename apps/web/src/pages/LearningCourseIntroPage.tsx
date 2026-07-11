import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi } from "../lib/learningApi";

const topicTags = ["行业分析", "市场洞察", "数据分析", "竞争分析", "AI工具应用"] as const;

const courseMeta = [
  ["课程时长", "2小时18分钟"],
  ["学习难度", "中级"],
  ["更新日期", "2024-05-20"],
  ["课程语言", "中文"]
] as const;

const outcomes = [
  "掌握 AI 行业分析的整体框架与关键步骤",
  "学会使用 AI 工具进行行业数据的获取与清洗",
  "掌握市场规模、增长趋势与驱动因素分析方法",
  "学会竞争格局与头部企业分析的实战技巧",
  "输出结构化行业分析报告，支撑业务决策"
] as const;

const chapters = [
  ["第1章", "行业分析概述与框架", "12:35", true],
  ["第2章", "行业数据获取与清洗", "18:42", false],
  ["第3章", "市场规模与增长趋势分析", "22:18", false],
  ["第4章", "行业驱动因素与机会洞察", "19:56", false],
  ["第5章", "竞争格局与头部企业分析", "23:47", false],
  ["第6章", "用户需求与细分市场分析", "16:30", false],
  ["第7章", "行业分析报告撰写与呈现", "24:12", false]
] as const;

const audience = [
  ["产品经理", "需要进行行业研究与市场分析"],
  ["市场分析师", "需要洞察市场趋势与竞争格局"],
  ["运营经理", "需要基于数据制定运营策略"],
  ["创业者与企业管理者", "需要评估行业机会与投资决策"]
] as const;

const relatedCourses = [
  ["提示词工程实战", "掌握高效提示词设计与优化技巧", "8.7k人学习", "cyan"],
  ["智能客服应用案例", "打造高效客服产品，提升满意度", "6.3k人学习", "violet"],
  ["数据可视化与报告呈现", "用数据讲故事，提升汇报影响力", "7.2k人学习", "blue"]
] as const;

function LearningCourseIntroPage() {
  const navigate = useNavigate();
  const [starting, setStarting] = useState(false);
  const [startError, setStartError] = useState("");

  async function startCourse() {
    if (starting) return;
    setStarting(true);
    setStartError("");
    try {
      await learningApi.updateProgress("ai-market-analysis", {
        percent: 1,
        last_lesson: "第1章 行业分析概述与框架",
        recommended_action: "继续学习第1章"
      });
      navigate("/learning/courses/detail");
    } catch {
      setStartError("课程启动失败，请稍后重试。");
    } finally {
      setStarting(false);
    }
  }

  return (
    <V4PageShell>
      <section className="learning-page course-intro-page" aria-label="课程介绍">
        <div className="course-intro-main">
          <section className="course-intro-hero">
            <div className="course-intro-cover" aria-label="AI行业分析方法课程封面">
              <span>热门工具</span>
              <i className="course-stage" aria-hidden="true" />
              <i className="course-bars" aria-hidden="true" />
              <i className="course-lens" aria-hidden="true" />
            </div>

            <div className="course-intro-summary">
              <div className="diagnosis-breadcrumb">
                <Link to="/learning">AI教学</Link>
                <span>/</span>
                <Link to="/learning/courses">全部课程</Link>
                <span>/</span>
                <strong>AI行业分析方法</strong>
              </div>
              <div className="course-intro-kicker">
                <span>AI教学</span>
                <span>行业分析</span>
              </div>
              <h1>AI行业分析方法</h1>
              <p>用 AI 洞察市场趋势与竞争格局，驱动科学决策</p>
              <div className="course-teacher-row">
                <span className="teacher-avatar" aria-hidden="true">李</span>
                <div>
                  <strong>李明远</strong>
                  <small>智活AI 高级数据分析师</small>
                </div>
                <b>9.1k 人学习</b>
                <div className="learner-stack" aria-label="近期学习用户">
                  <span />
                  <span />
                  <span />
                  <span />
                </div>
              </div>
              <div className="course-topic-tags">
                {topicTags.map((tag) => (
                  <span key={tag}>{tag}</span>
                ))}
              </div>
              <div className="course-hero-actions">
                <button disabled={starting} onClick={startCourse} type="button">{starting ? "正在开始..." : "立即学习"}</button>
                <Link to="/learning/plan">加入学习计划</Link>
                <button type="button">收藏</button>
              </div>
              {startError ? <p role="alert">{startError}</p> : null}
            </div>
          </section>

          <section className="course-info-grid" aria-label="课程核心信息">
            <article className="course-info-card course-intro-copy">
              <h2>课程简介</h2>
              <p>本课程将带你掌握 AI 驱动的行业分析全流程方法论，从数据获取、清洗分析到趋势洞察与竞争格局研判，帮助你快速上手行业研究，输出高价值分析报告，为业务决策提供有力支撑。</p>
              <dl>
                {courseMeta.map(([label, value]) => (
                  <div key={label}>
                    <dt>{label}</dt>
                    <dd>{value}</dd>
                  </div>
                ))}
              </dl>
            </article>

            <article className="course-info-card course-outcomes-card">
              <h2>你将学到什么</h2>
              <ul>
                {outcomes.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </article>

            <article className="course-info-card course-goal-card">
              <header>
                <h2>与你的目标强相关</h2>
                <span>当前目标</span>
              </header>
              <div className="goal-match-card">
                <strong>智能客服与市场分析能力提升</strong>
                <label>
                  <span>目标进度</span>
                  <progress max="100" value="35">35%</progress>
                  <b>35%</b>
                </label>
                <p>提升市场洞察与数据分析能力，为智能客服产品优化与市场策略制定提供数据支撑。</p>
                <Link to="/learning/plan">查看目标详情</Link>
              </div>
            </article>
          </section>

          <section className="course-lower-grid">
            <article className="course-info-card course-chapters-card" aria-label="章节目录">
              <header>
                <h2>章节目录</h2>
                <small>共7章</small>
                <button type="button">展开全部</button>
              </header>
              <div className="chapter-list">
                {chapters.map(([chapter, title, duration, unlocked]) => (
                  <div key={chapter}>
                    <span aria-hidden="true">⌄</span>
                    <strong>{chapter}</strong>
                    <p>{title}</p>
                    <time>{duration}</time>
                    <i aria-label={unlocked ? "可播放" : "已锁定"}>{unlocked ? "▶" : "▢"}</i>
                  </div>
                ))}
              </div>
            </article>

            <article className="course-info-card course-audience-card">
              <h2>适合人群</h2>
              {audience.map(([role, desc]) => (
                <section key={role}>
                  <i aria-hidden="true" />
                  <div>
                    <strong>{role}</strong>
                    <p>{desc}</p>
                  </div>
                </section>
              ))}
            </article>

            <article className="course-info-card course-related-card">
              <header>
                <h2>相关课程推荐</h2>
                <Link to="/learning/recommended-courses">查看更多</Link>
              </header>
              {relatedCourses.map(([title, desc, learners, tone]) => (
                <Link key={title} to="/learning/courses/detail">
                  <i className={tone} aria-hidden="true" />
                  <div>
                    <strong>{title}</strong>
                    <p>{desc}</p>
                    <small>{learners}</small>
                  </div>
                </Link>
              ))}
            </article>
          </section>
        </div>

        <aside className="learning-copilot course-intro-copilot" aria-label="智活 Copilot 课程介绍助手">
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

          <div className="learning-chat course-intro-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我可以基于你的项目、任务和能力短板，推荐最适合你的 AI 课程和学习路径。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>我想提升智能客服方向的能力，并了解市场分析方法，有哪些学习建议？</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>为你规划了智能客服与市场分析的学习路径，建议按以下顺序学习：<br />1. AI行业分析方法 / 2. 提示词工程实战<br />3. 智能客服应用案例 / 4. AI行业分析方法</p>
            </article>
          </div>

          <section className="course-report-card" aria-label="能力诊断报告">
            <p>已为你生成《能力诊断报告》，可下载查看详细提升建议。</p>
            <Link to="/learning/report" aria-label="能力诊断报告.pdf">
              <i aria-hidden="true">PDF</i>
              <span><b>能力诊断报告.pdf</b><small>PDF · 1.2 MB</small></span>
            </Link>
          </section>

          <nav className="learning-copilot-actions" aria-label="课程介绍助手快捷入口">
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

export default LearningCourseIntroPage;
