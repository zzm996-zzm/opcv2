import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import LearningFlowSteps from "../components/LearningFlowSteps";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningDiagnosis } from "../lib/learningApi";
import { referenceDiagnosis } from "../lib/learningReference";

function LearningAssessmentPage() {
  const location = useLocation();
  const routedDiagnosis = (location.state as { diagnosis?: LearningDiagnosis } | null)?.diagnosis ?? null;
  const [diagnosis, setDiagnosis] = useState<LearningDiagnosis | null>(routedDiagnosis);
  const [loading, setLoading] = useState(!routedDiagnosis);

  useEffect(() => {
    if (routedDiagnosis) {
      setDiagnosis(routedDiagnosis);
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    learningApi.getLatestAssessment()
      .then((payload) => {
        if (active) setDiagnosis(payload);
      })
      .catch(() => { if (active) setDiagnosis(referenceDiagnosis); })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [routedDiagnosis]);

  if (loading) return <AssessmentState title="正在加载能力评估..." />;
  if (!diagnosis) return <AssessmentState title="暂无能力评估" />;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page assessment-page" aria-label="能力评估">
        <div className="diagnosis-main">
          <section className="diagnosis-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link><span>/</span><Link to="/learning/diagnosis">能力诊断</Link>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你的项目、任务与工具使用情况，精准发现能力差距</p>
            </div>
            <div className="diagnosis-target-art" aria-hidden="true" />
          </section>

          <LearningFlowSteps active={2} />

          <section className="diagnosis-card assessment-progress-card" aria-label="评估概览">
            <header>
              <div><h2>{diagnosis.project}</h2><p>目标：{diagnosis.goal}</p></div>
              <div className="assessment-percent"><span>模型评估</span><strong>{diagnosis.overall_score}/100</strong></div>
            </header>
          </section>

          <div className="assessment-workbench">
            <section className="diagnosis-card ability-stage-card" aria-label="模型评估维度">
              <header><h2>模型评估维度</h2><p>分数仅用于当前学习优先级排序</p></header>
              <div className="ability-metric-list">
                {diagnosis.dimensions.map((dimension, index) => (
                  <article key={dimension.name}>
                    <span>{index + 1}</span><strong>{dimension.name}</strong>
                    <i aria-hidden="true"><b style={{ width: `${dimension.score}%` }} /></i>
                    <em>{dimension.score}</em><small>{dimension.summary}</small>
                  </article>
                ))}
              </div>
            </section>

            <section className="diagnosis-card ability-radar-card" aria-label="评估依据与假设">
              <header><div><h2>评估依据与假设</h2><p>只列出本次快照实际记录的信息</p></div></header>
              <div className="ability-radar-visual" aria-label="能力评估雷达图"><span /><span /><span /></div>
              <h3>输入依据</h3>
              <ul>
                {(diagnosis.evidence_sources ?? []).map((source) => <li key={`${source.type}-${source.label}`}>{source.label}</li>)}
              </ul>
              <h3>关键假设</h3>
              <ul>{(diagnosis.assumptions ?? []).map((assumption) => <li key={assumption}>{assumption}</li>)}</ul>
            </section>
          </div>

          <section className="diagnosis-card assessment-action-bar" aria-label="评估后续操作">
            <p>评估快照已保存，可继续查看差距、报告和学习计划。</p>
            <div>
              <Link to="/learning/gap-analysis">查看差距分析</Link>
              <Link to="/learning/report">查看诊断报告</Link>
              <Link to="/learning/plan">查看学习计划</Link>
            </div>
          </section>
        </div>

        <aside className="learning-copilot diagnosis-copilot assessment-copilot" aria-label="智活 Copilot 诊断助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>能力评估说明</p></div></header>
          <div className="learning-chat assessment-chat">
            <article><span className="ai-avatar">A</span><p>这是一份基于提交信息生成的模型评估，不是考试成绩或能力认证。</p></article>
            <article><span className="ai-avatar">A</span><p>建议用课程练习、项目交付物和复盘记录持续验证，再重新诊断。</p></article>
          </div>
          <nav className="learning-copilot-actions" aria-label="诊断助手快捷入口">
            <Link to="/learning/diagnosis">重新诊断 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommendation">查看学习建议 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm activeFilters={{ module: "learning", view: "assessment" }} className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function AssessmentState({ title }: { title: string }) {
  return (
    <V4PageShell>
      <section className="learning-page diagnosis-card" role="status">
        <h1>{title}</h1><Link to="/learning/diagnosis">发起能力诊断</Link>
      </section>
    </V4PageShell>
  );
}

export default LearningAssessmentPage;
