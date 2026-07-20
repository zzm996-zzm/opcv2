import { ArrowRight, Check, Edit3, Sparkles } from "lucide-react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxSession } from "../../lib/sandboxApi";

type SandboxSummaryViewProps = {
  session: SandboxSession;
};

function SandboxSummaryView({ session }: SandboxSummaryViewProps) {
  const answered = session.intake?.questions ?? [];

  return (
    <SandboxFrame copilotMode="summary" copilotProgress={100} copilotProject={session.intake?.initial_idea}>
      <SandboxStepper active={1} />
      <section className="sb-work-panel sb-summary-panel">
        <header className="sb-panel-header">
          <div className="sb-heading-with-icon">
            <span><Sparkles size={18} /></span>
            <div><h1>推演发起配置</h1><small>AI 智能补充完成</small></div>
          </div>
          <Link className="sb-quiet-button" to={`/sandbox/questions?session=${session.id}`}><Edit3 size={15} />继续修改</Link>
        </header>
        <p className="sb-panel-intro">关键信息已经整理完成。请确认以下内容，确认后选择本轮推演角色。</p>
        <div className="sb-summary-overview">
          <article>
            <strong>你已输入的信息</strong>
            <p>{session.intake?.initial_idea || session.goal}</p>
          </article>
          <div className="sb-summary-orbit" aria-hidden="true"><span>AI</span></div>
        </div>
        <h2 className="sb-section-title">为了更准确地推演，我们已补充以下关键信息</h2>
        <section className="sb-summary-question-grid">
          {answered.map((question, index) => (
            <article key={question.key}>
              <span>{index + 1}</span>
              <div>
                <strong>{question.title}</strong>
                <small>{question.hint}</small>
                <p>{question.skipped ? "暂未补充" : question.answer || "暂未补充"}</p>
              </div>
              <em><Check size={12} />{question.skipped ? "已跳过" : "已回答"}</em>
            </article>
          ))}
        </section>
        <footer className="sb-summary-footer">
          <Link to={`/sandbox/roles?session=${session.id}`}>确认以上信息，选择推演角色<ArrowRight size={19} /></Link>
        </footer>
      </section>
    </SandboxFrame>
  );
}

export default SandboxSummaryView;

