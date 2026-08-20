import { useState, type ReactNode } from "react";
import { ChevronRight, ChevronUp, FileText, Settings } from "lucide-react";
import { Link } from "react-router-dom";

import { useAuthSession } from "../lib/authSession";
import { useRegisteredCopilotPanel } from "./CopilotPanelVisibility";
import CopilotMemorySettings from "./CopilotMemorySettings";
import FloatingCopilotOrb from "./FloatingCopilotOrb";
import { MiniCopilotForm } from "./MiniCopilot";

export type UnifiedCopilotAction = {
  href: string;
  label: string;
};

type UnifiedCopilotPanelProps = {
  actionContent?: ReactNode;
  actions?: readonly UnifiedCopilotAction[];
  ariaLabel?: string;
  actionsAriaLabel?: string;
  className?: string;
  content?: ReactNode;
  inputAriaLabel?: string;
  bodyId?: string;
  report?: {
    meta: string;
    title: string;
  };
  response?: string;
  showComposer?: boolean;
  subtitle?: string;
  userPrompt?: string;
  taskID?: number;
  currentView?: string;
  activeFilters?: Record<string, string>;
};

const defaultActions: readonly UnifiedCopilotAction[] = [
  { href: "/analysis", label: "分析项目机会" },
  { href: "/tools/recommend", label: "推荐工具" },
  { href: "/tasks", label: "制定落地计划" }
];

function UnifiedCopilotPanel({
  actionContent,
  actions = defaultActions,
  actionsAriaLabel = "Copilot 快捷入口",
  ariaLabel = "智活 Copilot",
  bodyId,
  className = "",
  content,
  inputAriaLabel = "询问智活 Copilot",
  report,
  response = "好的，我会结合当前页面信息，为你整理重点并给出下一步建议。",
  showComposer = true,
  subtitle = "你的全球 AI 助手，随时为你提供帮助",
	userPrompt,
	taskID,
	currentView,
	activeFilters
}: UnifiedCopilotPanelProps) {
  const session = useAuthSession();
  const panel = useRegisteredCopilotPanel();
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [showQuickActions, setShowQuickActions] = useState(true);
  const nickname = session.user?.nickname || "张婧";

  if (!panel.isPanelOpen) {
    return panel.hasSharedController ? null : <FloatingCopilotOrb onActivate={panel.openPanel} />;
  }

  return (
    <aside className={`unified-copilot-panel ${className}`.trim()} aria-label={ariaLabel}>
      <header className="unified-copilot-head">
        <div>
          <strong><span className="copilot-title-logo v4-logo" aria-hidden="true" />智活 <b>Copilot</b></strong>
          <p>{subtitle}</p>
        </div>
        <div className="unified-copilot-controls">
          <button
            aria-expanded={settingsOpen}
            aria-label={settingsOpen ? "关闭 Copilot 设置" : "打开 Copilot 设置"}
            onClick={() => setSettingsOpen((open) => !open)}
            title="Copilot 设置"
            type="button"
          >
            <Settings aria-hidden="true" />
          </button>
          <button
            aria-label="收起 Copilot"
            onClick={() => {
              setSettingsOpen(false);
              panel.closePanel();
            }}
            title="收起 Copilot"
            type="button"
          >
            <ChevronUp aria-hidden="true" />
          </button>
        </div>
      </header>

      {settingsOpen && (
        <section className="unified-copilot-settings" role="dialog" aria-label="Copilot 设置">
          <label className="unified-copilot-quick-setting"><input aria-label="显示快捷建议" checked={showQuickActions} onChange={(event) => setShowQuickActions(event.target.checked)} type="checkbox" /><span><strong>显示快捷建议</strong><small>已附带当前页面上下文</small></span></label>
          <Link className="unified-copilot-preferences-link" to="/profile/preferences">更多偏好设置 <ChevronRight aria-hidden="true" /></Link>
          <CopilotMemorySettings />
        </section>
      )}

      <div className="unified-copilot-body" id={bodyId}>
          {content ?? <div className="unified-copilot-thread">
            <article className="assistant">
              <span className="ai-avatar" aria-hidden="true">A</span>
              <p><strong>嗨，{nickname}！</strong>今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</p>
            </article>
            {userPrompt && <article className="user"><p>{userPrompt}</p></article>}
            <article className="assistant">
              <span className="ai-avatar" aria-hidden="true">A</span>
              <p>{response}</p>
            </article>
            {report && (
              <Link className="unified-copilot-report" to="/copilot">
                <i aria-hidden="true"><FileText /></i>
                <span><strong>{report.title}</strong><small>{report.meta}</small></span>
              </Link>
            )}
          </div>}

          {showQuickActions && (actionContent ?? (
            <nav className="unified-copilot-actions" aria-label={actionsAriaLabel}>
              {actions.map((action) => (
                <Link key={`${action.href}-${action.label}`} to={action.href}>{action.label}<ChevronRight aria-hidden="true" /></Link>
              ))}
            </nav>
          ))}

          {showComposer && <MiniCopilotForm
            className="unified-copilot-input"
            inputAriaLabel={inputAriaLabel}
	            sendIcon="➤"
	            threadTitlePrefix="页面助手："
	            taskID={taskID}
	            currentView={currentView}
	            activeFilters={activeFilters}
	          />}
      </div>
    </aside>
  );
}

export default UnifiedCopilotPanel;
