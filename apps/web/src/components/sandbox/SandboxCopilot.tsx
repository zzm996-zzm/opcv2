import { ChevronRight, ChevronUp, Paperclip, Send, Settings, Sparkles } from "lucide-react";
import { useLayoutEffect, type ReactNode } from "react";
import { Link } from "react-router-dom";

import { useRegisteredCopilotPanel } from "../CopilotPanelVisibility";
import FloatingCopilotOrb from "../FloatingCopilotOrb";
import { MiniCopilotForm } from "../MiniCopilot";

export type SandboxCopilotMode = "home" | "questions" | "summary" | "roles" | "start" | "run" | "report" | "history";

type SandboxCopilotProps = {
  children?: ReactNode;
  facts?: Array<{ label: string; value: string }>;
  mode: SandboxCopilotMode;
  progress?: number;
  project?: string;
  runID?: number;
};

const modeCopy: Record<SandboxCopilotMode, { user: string; assistant: string; actions: Array<[string, string]> }> = {
  home: {
    user: "我有一个智能家居新品的想法，帮我从多角度分析市场并评估机会与潜在风险。",
    assistant: "好的，既然从用户、竞争对手、运营策略、增长路径和风险研判等多个角色进行模拟推演，为你输出机会与风险判断结果。",
    actions: [["开始多角色推演", "/sandbox/setup"], ["查看推演思路", "/sandbox/questions"], ["生成推演大纲", "/sandbox/setup"], ["历史推演记录", "/sandbox/history"]]
  },
  questions: {
    user: "请帮我补齐这次推演需要的信息。",
    assistant: "我会一次询问一个关键问题。回答会保存到当前沙盘，完成后可随时返回修改。",
    actions: [["返回项目设想", "/sandbox/setup"], ["查看历史推演", "/sandbox/history"]]
  },
  summary: {
    user: "这些信息已经够开始推演了吗？",
    assistant: "关键信息已经整理完成。确认摘要后，选择本轮要参与推演的角色。",
    actions: [["查看沙盘示例", "/sandbox/history"], ["帮助中心", "/help"]]
  },
  roles: {
    user: "我应该选择哪些角色？",
    assistant: "建议至少选择用户、投资人和一个执行视角。角色越互补，机会与风险判断越立体。",
    actions: [["角色选择建议", "/sandbox/roles"], ["返回修改信息", "/sandbox/setup"]]
  },
  start: {
    user: "基于当前信息和所选角色，开始商业沙盘推演。",
    assistant: "我会按所选深度生成角色结论、机会风险、验证动作与阶段路线。启动前仍可调整输出设置。",
    actions: [["查看完整信息", "/sandbox/setup"], ["历史推演记录", "/sandbox/history"]]
  },
  run: {
    user: "请给我本轮推演的下一步建议。",
    assistant: "推演会先汇总各角色判断，再收敛为机会、风险和可执行验证动作。完成后可继续向任一角色追问。",
    actions: [["聚焦风险分析", "/sandbox/run"], ["查看推演报告", "/sandbox/report"], ["保存当前记录", "/sandbox/history"]]
  },
  report: {
    user: "请总结本次推演的关键结论。",
    assistant: "优先验证高影响、高不确定性的假设，并把报告中的动作转成小范围真实实验。",
    actions: [["查看完整建议动作", "/sandbox/report"], ["返回推演记录", "/sandbox/history"]]
  },
  history: {
    user: "请总结本次推演的下一步建议。",
    assistant: "可从评分、风险和参与角色快速筛选记录，再进入报告复盘假设与验证动作。",
    actions: [["开始新推演", "/sandbox/setup"], ["历史推演记录", "/sandbox/history"]]
  }
};

function SandboxCopilot({ children, facts = [], mode, progress, project, runID }: SandboxCopilotProps) {
  const copilotPanel = useRegisteredCopilotPanel();
  const openPanel = copilotPanel.openPanel;
  const copy = modeCopy[mode];
  const isQuestionAnalysis = mode === "questions";

  useLayoutEffect(() => {
    // A sandbox route owns its Copilot context. Re-open it when entering a
    // new step so a collapsed panel from another page cannot leave a blank
    // grid column beside the workflow.
    openPanel();
  }, [mode, openPanel]);

  if (!copilotPanel.isPanelOpen) {
    return copilotPanel.hasSharedController ? null : <FloatingCopilotOrb onActivate={copilotPanel.openPanel} />;
  }

  return (
    <aside className={`sb-copilot${mode === "home" ? " sb-copilot-home" : ""}`} aria-label="智活 Copilot">
      <header>
        <span className="sb-copilot-mark" aria-hidden="true"><Sparkles size={20} /></span>
        <div>
          <strong>智活 Copilot</strong>
          <small>你的全能 AI 助手，随时为你提供帮助</small>
        </div>
        <div className="sb-copilot-controls">
          <Link aria-label="Copilot 设置" to="/profile/preferences"><Settings size={17} /></Link>
          <button aria-label="收起 Copilot" onClick={copilotPanel.closePanel} type="button">
            <ChevronUp size={18} />
          </button>
        </div>
      </header>
      <>
          <div className={`sb-copilot-thread${isQuestionAnalysis ? " is-analysis" : ""}`}>
            {isQuestionAnalysis ? <>
              <article className="is-user"><span>我</span><div><strong>我正在分析你的初步想法</strong><p>我已经梳理了你的想法，现在正在从市场、用户、产品、商业模式等多个维度进行分析，为你生成最关键的补充问题。</p><p>我会一次只问一个问题，请认真回答，这将帮助 AI 生成更精准的推演结果。</p></div></article>
              <section className="sb-copilot-facts"><h3>我已识别到的信息</h3><ul>{facts.map((fact) => <li key={fact.label}>{fact.label}：{fact.value}</li>)}</ul></section>
              <section className="sb-copilot-why"><h3>为什么先问这个问题？</h3><p>明确目标用户细分后，才能更准确地判断市场规模、需求强度与竞争格局，为后续推演奠定基础。</p></section>
            </> : <>
              <article className="is-user"><span>我</span><p>{project || copy.user}</p></article>
              <article><span><Sparkles size={14} /></span><p>{copy.assistant}</p></article>
            </>}
          </div>
          {children}
          {typeof progress === "number" && (
            <section className="sb-copilot-progress" aria-label={`当前进度 ${progress}%`}>
              <div><strong>当前进度</strong><span>{isQuestionAnalysis ? "1 / 5" : `${progress}%`}</span></div>
              <i><span style={{ width: `${Math.max(0, Math.min(progress, 100))}%` }} /></i>
              {isQuestionAnalysis ? <small>填写本题后进入下一题</small> : null}
            </section>
          )}
          {!isQuestionAnalysis ? <nav aria-label="Copilot 快捷操作">
            {copy.actions.map(([label, href]) => (
              <Link key={label} to={href}><span>{label}</span><ChevronRight size={16} /></Link>
            ))}
          </nav> : null}
          {!isQuestionAnalysis ? <MiniCopilotForm
            activeFilters={{
              module: "sandbox",
              view: mode,
              ...(runID ? { run_id: String(runID) } : {})
            }}
            attachIcon={<Paperclip size={16} />}
            className="sb-copilot-input"
            inputAriaLabel="向沙盘 Copilot 提问"
            placeholder="询问任何问题..."
            sendIcon={<Send size={16} />}
            threadTitlePrefix="商业沙盘："
          /> : null}
      </>
    </aside>
  );
}

export default SandboxCopilot;
