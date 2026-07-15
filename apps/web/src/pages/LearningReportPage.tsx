import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import LearningFlowSteps from "../components/LearningFlowSteps";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningReport } from "../lib/learningApi";
import { referenceReport } from "../lib/learningReference";

function LearningReportPage() {
  const [report, setReport] = useState<LearningReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.getLatestReport()
      .then((payload) => { if (active) setReport(payload); })
      .catch(() => { if (active) setReport(referenceReport); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  if (loading) return <ReportState title="正在加载诊断报告..." />;
  if (!report) return <ReportState title="暂无诊断报告" />;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page learning-report-page" aria-label="能力诊断报告">
        <div className="diagnosis-main learning-report-main">
          <section className="diagnosis-hero learning-report-hero">
            <div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>能力诊断报告</strong></div>
            <div className="diagnosis-hero-copy"><h1>能力诊断</h1><p>基于你的项目、任务与工具使用情况，精准发现能力差距</p></div>
          </section>

          <LearningFlowSteps active={4} />

          <section className="diagnosis-card report-overview-card" aria-label="诊断概览">
            <header><h2>诊断概览</h2><p>{report.project} · {report.goal}</p></header>
            <div className="report-overview-body">
              <article className="report-target-card"><h3>综合模型评估</h3><strong>{report.overall_score}/100</strong><p>该分数不是标准化考试成绩。</p></article>
              <section className="report-score-panel" aria-label="能力差距分布">
                <h3>能力差距分布</h3>
                <div className="report-score-grid">
                  {report.dimensions.map((dimension) => (
                    <article key={dimension.name}><h4>{dimension.name}</h4><strong>{dimension.score}</strong><span>差距 {dimension.gap}</span><small>{dimension.summary}</small></article>
                  ))}
                </div>
              </section>
            </div>
          </section>

          <div className="report-lower-grid">
            <section className="diagnosis-card report-priority-card" aria-label="优先补齐能力">
              <h2>优先补齐能力</h2>
              <div className="report-priority-list">
                {report.priority_gaps.map((gap, index) => (
                  <article key={gap.name}><b>{index + 1}</b><div><h3>{gap.name}</h3><p>{gap.recommended}</p></div><strong>{gap.current}<small>模型分</small></strong></article>
                ))}
              </div>
              <Link to="/learning/recommendation">查看学习建议 <span aria-hidden="true">→</span></Link>
            </section>

            <section className="diagnosis-card report-evidence-card" aria-label="诊断依据">
              <h2>诊断依据</h2>
              {(report.evidence_sources ?? []).length === 0 ? <p>该历史报告未记录结构化输入依据。</p> : <ul>{(report.evidence_sources ?? []).map((source) => <li key={`${source.type}-${source.label}`}>{source.label}</li>)}</ul>}
              <h3>关键假设</h3><ul>{(report.assumptions ?? []).map((item) => <li key={item}>{item}</li>)}</ul>
            </section>
          </div>

          <section className="diagnosis-card"><h2>模型建议</h2><ul>{report.recommendations.map((item) => <li key={item}>{item}</li>)}</ul></section>
        </div>

        <aside className="learning-copilot diagnosis-copilot learning-report-copilot" aria-label="智活 Copilot 诊断报告助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>报告边界说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>当前没有实现 PDF 导出，因此这里只提供持久化网页报告。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="诊断报告助手快捷入口"><Link to="/learning/plan">查看学习计划 <span aria-hidden="true">›</span></Link><Link to="/learning/diagnosis">重新诊断 <span aria-hidden="true">›</span></Link></nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function ReportState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page diagnosis-card" role="status"><h1>{title}</h1><Link to="/learning/diagnosis">发起能力诊断</Link></section></V4PageShell>;
}

export default LearningReportPage;
