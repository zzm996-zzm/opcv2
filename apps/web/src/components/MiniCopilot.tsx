import { FormEvent, ReactNode, useState } from "react";

import { ApiRequestError } from "../lib/apiRequest";
import { copilotApi, type CopilotMessage } from "../lib/copilotApi";
import MarkdownMessage from "./MarkdownMessage";

type InitialMessage = {
  content: ReactNode;
  role?: "assistant" | "user";
};

type MiniCopilotFormProps = {
  className: string;
  inputAriaLabel?: string;
  placeholder?: string;
  attachLabel?: string;
  attachIcon?: ReactNode;
  onAttach?: () => void;
  sendIcon?: ReactNode;
  model?: string;
  threadTitlePrefix?: string;
  taskID?: number;
  currentView?: string;
  activeFilters?: Record<string, string>;
};

type MiniCopilotProps = Omit<MiniCopilotFormProps, "className"> & {
  actions?: ReactNode;
  ariaLabel: string;
  className: string;
  header?: ReactNode;
  inputClassName: string;
  initialMessages?: InitialMessage[];
  messagesClassName: string;
  title: string;
  subtitle?: string;
};

type LiveMessage = {
  id: string;
  messageID?: number;
  content: string;
  metadata?: CopilotMessage["metadata"];
  role: "assistant" | "user";
};

export function MiniCopilotForm({
  className,
  inputAriaLabel = "向 Copilot 提问",
  placeholder = "询问任何问题...",
  attachLabel = "添加附件",
  attachIcon = "+",
  onAttach,
  sendIcon = "⌁",
  model,
	threadTitlePrefix = "",
	taskID,
	currentView,
	activeFilters
}: MiniCopilotFormProps) {
  const [draft, setDraft] = useState("");
  const [threadID, setThreadID] = useState<number | null>(null);
  const [liveMessages, setLiveMessages] = useState<LiveMessage[]>([]);
  const [isSending, setIsSending] = useState(false);
	const [toolBusyID, setToolBusyID] = useState<number | null>(null);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    const content = draft.trim();
    if (!content || isSending) return;

    const optimisticID = `user-${Date.now()}`;
    setDraft("");
    setError("");
    setIsSending(true);
    setLiveMessages((current) => [...current, { id: optimisticID, role: "user", content }]);

    try {
      const createThread = async () => {
        const title = `${threadTitlePrefix}${content}`.slice(0, 28) || "新会话";
        const thread = await copilotApi.createThread({ title, mode: "chat", model });
        setThreadID(thread.id);
        return thread.id;
      };
      let nextThreadID = threadID ?? await createThread();
      const messageInput = {
	        content,
	        model,
	        task_id: taskID,
	        current_view: currentView || window.location.pathname,
	        active_filters: activeFilters
	      };
      let result;
      try {
        result = await copilotApi.sendMessage(nextThreadID, messageInput);
      } catch (requestError) {
        if (!(requestError instanceof ApiRequestError) || requestError.code !== "thread_not_found") throw requestError;
        nextThreadID = await createThread();
        result = await copilotApi.sendMessage(nextThreadID, messageInput);
      }
      setLiveMessages((current) => replaceOptimisticMessage(current, optimisticID, result.user_message, result.assistant_message));
    } catch (requestError) {
      setDraft(content);
      setLiveMessages((current) => current.filter((message) => message.id !== optimisticID));
      setError(requestError instanceof Error ? requestError.message : "发送失败，请稍后重试");
    } finally {
      setIsSending(false);
    }
  }

	async function decideTool(message: LiveMessage, decision: "confirm" | "cancel") {
	  if (!threadID || !message.messageID || toolBusyID !== null) return;
	  setToolBusyID(message.messageID);
	  setError("");
	  try {
	    const result = await copilotApi.confirmTool(threadID, message.messageID, decision);
	    setLiveMessages((current) => current.map((item) => item.messageID === result.message.id ? liveMessageFrom(result.message) : item));
	  } catch (requestError) {
	    setError(requestError instanceof Error ? requestError.message : "操作失败，请稍后重试");
	  } finally {
	    setToolBusyID(null);
	  }
	}

  return (
    <>
      {(liveMessages.length > 0 || error) && (
        <div className="mini-copilot-live-thread" aria-live="polite">
          {liveMessages.map((message) => (
            <article className={message.role === "user" ? "user" : ""} key={message.id}>
              {message.role === "assistant" && <span className="ai-avatar">A</span>}
              <div>
                {message.role === "assistant" ? <MarkdownMessage content={message.content} /> : <p>{message.content}</p>}
                {message.metadata?.tool_preview?.status === "pending" && (
                  <div className="mini-copilot-tool-actions" aria-label="待确认操作">
                    <button disabled={toolBusyID !== null} onClick={() => void decideTool(message, "cancel")} type="button">取消</button>
                    <button disabled={toolBusyID !== null} onClick={() => void decideTool(message, "confirm")} type="button">
                      {toolBusyID === message.messageID ? "执行中..." : "确认执行"}
                    </button>
                  </div>
                )}
                {message.metadata?.tool_preview?.status === "cancelled" && <small>操作已取消</small>}
              </div>
            </article>
          ))}
          {error && <p className="mini-copilot-error" role="alert">{error}</p>}
        </div>
      )}
      <form className={className} onSubmit={submit}>
        <button aria-label={attachLabel} disabled={!onAttach} onClick={onAttach} title={onAttach ? attachLabel : `${attachLabel}（暂未开放）`} type="button">{attachIcon}</button>
        <input
          aria-label={inputAriaLabel}
          onChange={(event) => setDraft(event.target.value)}
          placeholder={placeholder}
          value={draft}
        />
        <button aria-label="发送" disabled={!draft.trim() || isSending} type="submit">{isSending ? "…" : sendIcon}</button>
      </form>
    </>
  );
}

function MiniCopilot({
  actions,
  ariaLabel,
  className,
  header,
  inputClassName,
  initialMessages = [],
  messagesClassName,
  title,
  subtitle = "你的全球 AI 助手，随时为你提供帮助",
  ...formProps
}: MiniCopilotProps) {
  return (
    <aside className={className} aria-label={ariaLabel}>
      {header ?? (
        <header>
          <div>
            <strong><span aria-hidden="true">✦</span> {title}</strong>
            <p>{subtitle}</p>
          </div>
        </header>
      )}
      {initialMessages.length > 0 && (
        <div className={messagesClassName}>
          {initialMessages.map((message, index) => (
            <article className={message.role === "user" ? "user" : ""} key={index}>
              {message.role !== "user" && <span className="ai-avatar">A</span>}
              <p>{message.content}</p>
            </article>
          ))}
        </div>
      )}
      {actions}
      <MiniCopilotForm className={inputClassName} {...formProps} />
    </aside>
  );
}

function replaceOptimisticMessage(
  current: LiveMessage[],
  optimisticID: string,
  userMessage: CopilotMessage,
  assistantMessage: CopilotMessage
) {
  const withoutOptimistic = current.filter((message) => message.id !== optimisticID);
	  return [
	    ...withoutOptimistic,
	    liveMessageFrom(userMessage),
	    liveMessageFrom(assistantMessage)
	  ];
}

function liveMessageFrom(message: CopilotMessage): LiveMessage {
	return {
	  id: `${message.role}-${message.id}`,
	  messageID: message.id,
	  role: message.role === "user" ? "user" : "assistant",
	  content: message.content,
	  metadata: message.metadata
	};
}

export default MiniCopilot;
