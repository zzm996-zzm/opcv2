import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import LearningFlowSteps from "../components/LearningFlowSteps";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningGaps } from "../lib/learningApi";
import { referenceGaps } from "../lib/learningReference";

function LearningGapAnalysisPage() {
  const [gaps, setGaps] = useState<LearningGaps | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.getLatestGaps()
      .then((payload) => {
        if (active) setGaps(payload);
      })
      .catch(() => { if (active) setGaps(referenceGaps); })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, []);

  if (loading) return <GapState title="正在加载差距分析..." />;
  if (!gaps) return <GapState title="暂无差距分析" />;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page gap-analysis-page" aria-label="能力差距分析">
        <div className="diagnosis-main gap-main">
          <section className="diagnosis-hero gap-hero">
            <div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>差距分析</strong></div>
            <div className="diagnosis-hero-copy"><h1>能力诊断</h1><p>{gaps.disclaimer || "该历史分析未记录免责声明，请重新诊断后使用。"}</p></div>
          </section>

          <LearningFlowSteps active={3} />

          <section className="diagnosis-card gap-comparison-card" aria-label="目标要求与当前水平对比">
            <header><h2>目标要求 vs 当前模型评估</h2></header>
            <div className="gap-comparison-body">
              <article className="goal-direction-card">
                <small>目标项目</small><h3>{gaps.project}</h3><p><strong>学习目标</strong> {gaps.goal}</p>
                <p><strong>综合模型分</strong> {gaps.overall_score}/100</p>
              </article>
              <section className="gap-cause-list" aria-label="能力差距列表">
                {gaps.gaps.map((gap, index) => (
                  <article key={gap.name}>
                    <span className="gap-number">{index + 1}</span>
                    <div>
                      <h3>{gap.name}<small>{gap.priority}</small></h3>
                      <p><b>当前 {gap.current}</b><b>目标 {gap.target}</b><b>差距 {gap.gap}</b></p>
                      <p>{gap.summary}</p><p>建议：{gap.recommended}</p>
                    </div>
                  </article>
                ))}
              </section>
            </div>
          </section>

          <div className="gap-details-grid">
            <section className="diagnosis-card gap-evidence-card" aria-label="判断依据">
              <h2>判断依据</h2>
              {(gaps.evidence_sources ?? []).length === 0 ? <p>该历史诊断未记录结构化输入依据。</p> : (
                <ul>{(gaps.evidence_sources ?? []).map((source) => <li key={`${source.type}-${source.label}`}>{source.label}</li>)}</ul>
              )}
            </section>
            <section className="diagnosis-card gap-priority-card" aria-label="关键假设">
              <h2>关键假设</h2><ul>{(gaps.assumptions ?? []).map((item) => <li key={item}>{item}</li>)}</ul>
            </section>
          </div>

          <section className="gap-footer-actions" aria-label="差距分析操作">
            <Link className="gap-secondary" to="/learning/diagnosis">重新诊断</Link>
            <Link className="gap-primary" to="/learning/report">查看诊断报告 <span aria-hidden="true">→</span></Link>
          </section>
        </div>

        <aside className="learning-copilot diagnosis-copilot gap-copilot" aria-label="智活 Copilot 差距分析助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>差距分析说明</p></div></header>
          <div className="learning-chat gap-chat">
            <article><span className="ai-avatar">A</span><p>差距排序来自本次持久化模型评估快照，仅用于安排学习优先级。</p></article>
          </div>
          <nav className="learning-copilot-actions" aria-label="差距分析助手快捷入口"><Link to="/learning/recommendation">查看学习建议 <span aria-hidden="true">›</span></Link></nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function GapState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page diagnosis-card" role="status"><h1>{title}</h1><Link to="/learning/diagnosis">发起能力诊断</Link></section></V4PageShell>;
}

export default LearningGapAnalysisPage;
