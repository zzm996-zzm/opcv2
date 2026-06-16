import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const diagnosisSteps = [
  ["1", "收集信息", "获取相关数据与目标"],
  ["2", "能力评估", "多维度能力评估打分"],
  ["3", "差距分析", "定位差距与原因"],
  ["4", "生成报告", "输出诊断结果"],
  ["5", "推荐方案", "推荐学习路径与课程"]
] as const;

const dataSources = [
  ["项目超市", "已识别当前项目", "智能客服与市场分析", "cube"],
  ["任务中心", "已读取最近任务", "5 条", "check"],
  ["工具箱", "已分析常用工具", "8 个", "case"],
  ["用户画像", "已同步目标与", "时间投入偏好", "profile"]
] as const;

const focusTags = ["提升专业能力", "优化工作效率", "拓展业务视野", "职业发展提升"];
const timeTags = ["1-2 小时", "3-5 小时", "5-8 小时", "8 小时以上"];

const analysisItems = [
  ["AI基础理解", "评估你对AI概念、原理和应用场景的理解程度", "ai"],
  ["提示词工程实战", "评估提示词设计与优化能力及实战应用水平", "chat"],
  ["行业分析方法", "评估市场分析、竞品分析方法掌握程度", "bars"],
  ["智能客服案例拆解", "评估案例理解、方案拆解与优化能力", "headset"],
  ["数据洞察能力", "评估数据收集、分析与洞察发现能力", "pie"],
  ["执行落地能力", "评估方案执行、项目推进与落地能力", "target"]
] as const;

function LearningDiagnosisPage() {
  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page" aria-label="能力诊断">
        <div className="diagnosis-main">
          <section className="diagnosis-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <strong>能力诊断</strong>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你的项目、任务与工具使用情况，精准发现能力差距</p>
            </div>
            <div className="diagnosis-target-art" aria-hidden="true" />
          </section>

          <section className="diagnosis-card diagnosis-steps" aria-label="诊断流程">
            {diagnosisSteps.map(([number, title, desc], index) => (
              <article className={index === 0 ? "active" : ""} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card data-source-card" aria-label="已接入分析的数据源">
            <div className="diagnosis-section-head">
              <h2>已接入分析的数据源</h2>
              <button type="button">↻ 更新数据源</button>
            </div>
            <div className="data-source-grid">
              {dataSources.map(([title, desc, value, icon]) => (
                <article key={title}>
                  <i className={`source-icon ${icon}`} aria-hidden="true" />
                  <div>
                    <h3>{title}</h3>
                    <p>{desc}</p>
                    <strong>{value}</strong>
                  </div>
                  <span aria-label={`${title} 已接入`}>✓</span>
                  <time>更新时间：2024-05-20 10:30</time>
                </article>
              ))}
            </div>
          </section>

          <div className="diagnosis-workbench">
            <section className="diagnosis-card diagnosis-goals" aria-label="补充你的诊断目标">
              <h2>补充你的诊断目标</h2>
              <div className="goal-row">
                <span>目标方向</span>
                <div className="goal-chips">
                  {focusTags.map((tag, index) => (
                    <button className={index === 0 ? "active" : ""} key={tag} type="button">{tag}</button>
                  ))}
                </div>
              </div>
              <label className="diagnosis-field">
                <span>希望提升的能力</span>
                <select defaultValue="">
                  <option value="" disabled>请选择核心想要提升的能力（可多选）</option>
                  <option>智能客服方案设计</option>
                </select>
              </label>
              <div className="goal-row">
                <span>每周可投入时间</span>
                <div className="goal-chips">
                  {timeTags.map((tag, index) => (
                    <button className={index === 1 ? "active" : ""} key={tag} type="button">{tag}</button>
                  ))}
                </div>
              </div>
              <label className="diagnosis-field textarea">
                <span>当前最大卡点</span>
                <textarea maxLength={100} placeholder="请描述你当前遇到的主要困难或挑战（选填）" />
                <small>0/100</small>
              </label>
            </section>

            <section className="diagnosis-card analysis-scope" aria-label="本次诊断将分析什么">
              <div>
                <h2>本次诊断将分析什么</h2>
                <p>基于 6 大核心维度全面评估你的能力水平</p>
              </div>
              <div className="analysis-grid">
                {analysisItems.map(([title, desc, icon]) => (
                  <article key={title}>
                    <i className={`analysis-icon ${icon}`} aria-hidden="true" />
                    <div>
                      <h3>{title}</h3>
                      <p>{desc}</p>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          </div>

          <div className="diagnosis-footer-actions">
            <Link className="diagnosis-primary" to="/learning/assessment">开始能力诊断 <span aria-hidden="true">→</span></Link>
            <Link className="diagnosis-secondary" to="/learning">稍后继续补充</Link>
            <p><span aria-hidden="true">♢</span> 诊断过程约需 8-12 分钟<br />我们会严格保护你的数据安全</p>
          </div>
        </div>

        <aside className="learning-copilot diagnosis-copilot" aria-label="智活 Copilot 诊断助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全局 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌃</span>
            </div>
          </header>

          <div className="learning-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我会结合你的项目、任务、工具使用情况完成本次能力诊断，帮你发现提升空间。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>我已经为你接入了以下数据：<br />✓ 项目超市：智能客服与市场分析<br />✓ 任务中心：最近任务 5 条<br />✓ 工具箱：常用工具 8 个<br />✓ 用户画像：目标与投入偏好</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>还需要你补充 2-3 个关键信息，以便更精准评估你的能力水平。建议先完善下方表单，再开始诊断。</p>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="诊断助手快捷入口">
            <Link to="/learning/diagnosis">查看诊断逻辑 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/diagnosis">补充诊断信息 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommended-courses">查看学习建议 <span aria-hidden="true">›</span></Link>
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

export default LearningDiagnosisPage;
