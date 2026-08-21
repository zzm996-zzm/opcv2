import { ArrowLeft } from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../V4PageShell";
import SandboxCopilot, { type SandboxCopilotMode } from "./SandboxCopilot";

type SandboxFrameProps = {
  children: ReactNode;
  className?: string;
  copilot?: ReactNode;
  copilotFacts?: Array<{ label: string; value: string }>;
  copilotMode: SandboxCopilotMode;
  copilotProgress?: number;
  copilotProject?: string;
  copilotRunID?: number;
  title?: boolean;
};

function SandboxFrame({
  children,
  className = "",
  copilot,
  copilotFacts,
  copilotMode,
  copilotProgress,
  copilotProject,
  copilotRunID,
  title = true
}: SandboxFrameProps) {
  const isHome = copilotMode === "home" && title === false;

  return (
    <V4PageShell className={`sb-shell${isHome ? " sb-home-shell" : ""}`} showCopilotMini={false}>
      <section className={`sb-app${isHome ? " sb-home-frame" : ""} ${className}`} aria-label="商业沙盘">
        <div className="sb-page-grid">
          <main className="sb-page-content">
            {title && (
              <header className="sb-page-title">
                <Link aria-label="返回商业沙盘首页" to="/sandbox"><ArrowLeft size={21} /></Link>
                <strong>商业沙盘</strong>
                <span>多角色模拟未来，判断项目机会与风险</span>
              </header>
            )}
            {children}
          </main>
          {copilot ?? (
            <SandboxCopilot facts={copilotFacts} mode={copilotMode} progress={copilotProgress} project={copilotProject} runID={copilotRunID} />
          )}
        </div>
      </section>
    </V4PageShell>
  );
}

export default SandboxFrame;
