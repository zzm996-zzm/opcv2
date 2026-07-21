import { Link } from "react-router-dom";

type PublicCopilotPanelProps = {
  className?: string;
};

function PublicCopilotPanel({ className = "" }: PublicCopilotPanelProps) {
  return (
    <aside className={`billing-copilot-card public-copilot-panel ${className}`} aria-label="智活 Copilot">
      <header><strong><b>◆</b> 智活 Copilot</strong><span>⚙⌃</span></header>
      <p>你的全球 AI 助手，随时为你提供帮助</p>
      <article><em>A</em><strong>嗨，张婧！</strong><p>今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</p></article>
      <article className="blue">帮我分析一下智能客服系统的市场机会和落地关键点。</article>
      <article><em>A</em><p>好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点，点击下方查看详情。</p></article>
      <div className="copilot-file-chip"><span className="pdf-thumb">PDF</span><span><strong>智能客服系统机会分析报告</strong><small>PDF · 1.2 MB</small></span></div>
      {["分析项目机会", "推荐工具", "制定落地计划"].map((item) => <Link key={item} to="/copilot">{item}<span>›</span></Link>)}
      <label className="billing-copilot-input"><span>⌾</span><input aria-label="询问 Copilot" placeholder="询问任何问题..." /><b>➤</b></label>
    </aside>
  );
}

export default PublicCopilotPanel;
