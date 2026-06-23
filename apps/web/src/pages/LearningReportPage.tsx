import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const reportSteps = [
  ["1", "收集信息", "获取项目与数据", "complete"],
  ["2", "能力评估", "多维度能力测评", "complete"],
  ["3", "差距分析", "定位关键弱项", "complete"],
  ["4", "生成报告", "输出诊断结果", "active"],
  ["5", "推荐方案", "推荐学习路径与课程", ""]
] as const;

const abilityScores = [
  ["AI基础理解", "70分", "中等", "掌握度 70%", "cube", "medium"],
  ["提示词工程实战", "55分", "较弱", "掌握度 55%", "chat", "weak"],
  ["行业分析方法", "60分", "中等", "掌握度 60%", "bars", "medium"],
  ["智能客服案例拆解", "45分", "较弱", "掌握度 45%", "headset", "weak"],
  ["数据洞察能力", "58分", "中等", "掌握度 58%", "target", "medium"]
] as const;

const priorityGaps = [
  ["智能客服案例拆解", "差距最大", "提升服务场景理解与方案拆解能力", "45分"],
  ["提示词工程实战", "重点提升", "掌握高质量提示词设计与优化技巧", "55分"],
  ["数据洞察能力", "持续补强", "提升数据解读与洞察发现能力", "58分"]
] as const;

const evidenceItems = [
  ["项目超市", "分析了 2 个相关项目", "grid"],
  ["任务中心", "分析了 5 个已完成任务", "check"],
  ["工具箱", "分析了 8 个常用工具", "case"]
] as const;

function LearningReportPage() {
  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page learning-report-page" aria-label="能力诊断报告">
        <div className="diagnosis-main learning-report-main">
          <section className="diagnosis-hero learning-report-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <strong>能力诊断</strong>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你的项目、任务与工具使用情况，精准发现能力差距</p>
            </div>
            <div className="learning-report-hero-art" aria-hidden="true" />
          </section>

          <section className="diagnosis-card diagnosis-steps learning-report-steps" aria-label="诊断流程">
            {reportSteps.map(([number, title, desc, state]) => (
              <article className={state} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card report-overview-card" aria-label="诊断概览">
            <header>
              <h2>诊断概览</h2>
              <p>整体能力水平与目标方向对比</p>
              <div className="report-legend" aria-hidden="true">
                <span className="good">优势</span>
                <span className="mid">中等</span>
                <span className="weak">较弱</span>
              </div>
            </header>
            <div className="report-overview-body">
              <article className="report-target-card">
                <h3>你的目标方向</h3>
                <strong>智能客服与市场分析能力提升</strong>
                <button type="button">修改目标</button>
                <div>
                  <span>整体匹配度</span>
                  <b>62分 <small>/ 100</small></b>
                  <i aria-hidden="true" />
                  <p>当前能力较目标方向仍有提升空间</p>
                </div>
              </article>
              <section className="report-score-panel" aria-label="能力差距分布">
                <h3>能力差距分布</h3>
                <div className="report-score-grid">
                  {abilityScores.map(([title, score, level, mastery, icon, state]) => (
                    <article className={state} key={title}>
                      <i className={`report-score-icon ${icon}`} aria-hidden="true" />
                      <h4>{title}</h4>
                      <strong>{score}</strong>
                      <span>{level}</span>
                      <small>{mastery}</small>
                      <em aria-hidden="true" />
                    </article>
                  ))}
                </div>
              </section>
            </div>
          </section>

          <div className="report-lower-grid">
            <section className="diagnosis-card report-priority-card" aria-label="优先补齐能力">
              <h2>优先补齐能力 <span>推荐学习顺序</span></h2>
              <div className="report-priority-body">
                <div className="report-priority-list">
                  {priorityGaps.map(([title, badge, desc, score], index) => (
                    <article key={title}>
                      <b>{index + 1}</b>
                      <div>
                        <h3>{title}<span>{badge}</span></h3>
                        <p>{desc}</p>
                      </div>
                      <strong>{score}<small>掌握度</small></strong>
                    </article>
                  ))}
                </div>
                <div className="report-growth-card">
                  <div className="report-stairs" aria-hidden="true">
                    <span />
                    <span />
                    <span />
                  </div>
                  <p>建议先从差距最大的能力开始补齐，循序渐进提升整体能力水平。</p>
                  <Link to="/learning/recommendation">去补这些课程 <span aria-hidden="true">→</span></Link>
                  <small>预计投入 10-16 小时，可提升整体匹配度至 85分+</small>
                </div>
              </div>
            </section>

            <section className="diagnosis-card report-evidence-card" aria-label="诊断依据">
              <h2>诊断依据 <span>已分析</span></h2>
              <div>
                {evidenceItems.map(([title, desc, icon]) => (
                  <article key={title}>
                    <i className={`report-evidence-icon ${icon}`} aria-hidden="true" />
                    <div>
                      <h3>{title}</h3>
                      <p>{desc}</p>
                    </div>
                  </article>
                ))}
              </div>
              <Link to="/learning/diagnosis">重新诊断</Link>
            </section>
          </div>

          <p className="report-footnote">
            <span aria-hidden="true">i</span> 诊断结果基于历史数据，持续学习将动态优化评估结果。
          </p>
        </div>

        <aside className="learning-copilot diagnosis-copilot learning-report-copilot" aria-label="智活 Copilot 诊断报告助手">
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

          <div className="learning-chat learning-report-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>张婧，我已基于你的项目、任务与工具使用情况，完成了能力诊断。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p><b>整体诊断结果：</b><br />整体匹配度 62 分，主要差距集中在“智能客服案例拆解”和“提示词工程实战”两个方面。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>建议你优先从“智能客服案例拆解”开始补齐，掌握不同场景下的服务方案拆解方法，快速提升业务落地能力。</p>
            </article>
            <article className="report-pdf-card">
              <span className="ai-avatar">A</span>
              <div>
                <p><b>下一步建议：</b><br />✓ 优先学习推荐的 3 门课程<br />✓ 结合任务中心实践巩固<br />✓ 完成后再次诊断，持续提升</p>
                <a href="/learning/report" aria-label="能力诊断报告.pdf">
                  <i aria-hidden="true">PDF</i>
                  <span><b>能力诊断报告.pdf</b><small>PDF · 1.2 MB</small></span>
                </a>
              </div>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="诊断报告助手快捷入口">
            <Link to="/learning/recommendation">推荐适合我的课程 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">为我制定学习计划 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/history">查看学习进度 <span aria-hidden="true">›</span></Link>
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

export default LearningReportPage;
