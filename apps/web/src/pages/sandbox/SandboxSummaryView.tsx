import { ArrowRight, Check, Edit3, Sparkles } from "lucide-react";
import { Link } from "react-router-dom";
import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxRun } from "../../lib/sandboxApi";

export default function SandboxSummaryView({ run }: { run: SandboxRun }) {
  return <SandboxFrame copilotMode="summary" copilotProgress={100} copilotProject={run.product.name}><SandboxStepper active={2} /><section className="sb-work-panel sb-summary-panel"><header className="sb-panel-header"><div className="sb-heading-with-icon"><span><Sparkles size={18} /></span><div><h1>确认项目事实</h1><small>服务器已保存当前版本</small></div></div><Link className="sb-quiet-button" to={`/sandbox/new?run=${run.id}&step=2`}><Edit3 size={15} />继续修改</Link></header><p className="sb-panel-intro">以下内容会作为所有角色共享的事实包，角色之间不会看到彼此输出。</p><div className="sb-summary-overview"><article><strong>{run.product.name}</strong><p>{run.product.selling_point || "尚未补充卖点"}</p><small>目标客户：{run.context.target_customer || "尚未补充"} · 渠道：{run.context.channel || "尚未补充"}</small></article><div className="sb-summary-orbit" aria-hidden="true"><span>V1.2</span></div></div><h2 className="sb-section-title">已记录的回答与假设</h2><section className="sb-summary-question-grid">{(run.questions ?? []).map((question, index) => <article key={question.key}><span>{index + 1}</span><div><strong>{question.question}</strong><small>{question.field}</small><p>{question.skipped ? "用户跳过，进入报告假设" : question.answer || "未填写"}</p></div><em><Check size={12} />{question.skipped ? "假设" : "已回答"}</em></article>)}</section><footer className="sb-summary-footer"><Link to={`/sandbox/new?run=${run.id}&step=3`}>选择推演角色<ArrowRight size={19} /></Link></footer></section></SandboxFrame>;
}
