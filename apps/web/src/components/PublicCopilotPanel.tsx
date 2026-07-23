import UnifiedCopilotPanel from "./UnifiedCopilotPanel";

type PublicCopilotPanelProps = {
  className?: string;
};

function PublicCopilotPanel({ className = "" }: PublicCopilotPanelProps) {
  return (
    <UnifiedCopilotPanel
      className={`billing-copilot-card public-copilot-panel ${className}`}
      inputAriaLabel="询问 Copilot"
      report={{ title: "智能客服系统机会分析报告", meta: "PDF · 1.2 MB" }}
      response="好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点，点击下方查看详情。"
      userPrompt="帮我分析一下智能客服系统的市场机会和落地关键点。"
    />
  );
}

export default PublicCopilotPanel;
