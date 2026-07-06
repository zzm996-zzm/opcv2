import { FormEvent, ReactNode, useState } from "react";

import { copilotApi, type CopilotMessage } from "../lib/copilotApi";

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
  content: string;
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
  threadTitlePrefix = ""
}: MiniCopilotFormProps) {
  const [draft, setDraft] = useState("");
  const [threadID, setThreadID] = useState<number | null>(null);
  const [liveMessages, setLiveMessages] = useState<LiveMessage[]>([]);
  const [isSending, setIsSending] = useState(false);
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
      let nextThreadID = threadID;
      if (!nextThreadID) {
        const title = `${threadTitlePrefix}${content}`.slice(0, 28) || "新会话";
        const thread = await copilotApi.createThread({ title, mode: "chat", model });
        nextThreadID = thread.id;
        setThreadID(thread.id);
      }
      const result = await copilotApi.sendMessage(nextThreadID, { content, model });
      setLiveMessages((current) => replaceOptimisticMessage(current, optimisticID, result.user_message, result.assistant_message));
    } catch (requestError) {
      setDraft(content);
      setLiveMessages((current) => current.filter((message) => message.id !== optimisticID));
      setError(requestError instanceof Error ? requestError.message : "发送失败，请稍后重试");
    } finally {
      setIsSending(false);
    }
  }

  return (
    <>
      <form className={className} onSubmit={submit}>
        <button aria-label={attachLabel} onClick={onAttach} type="button">{attachIcon}</button>
        <input
          aria-label={inputAriaLabel}
          onChange={(event) => setDraft(event.target.value)}
          placeholder={placeholder}
          value={draft}
        />
        <button aria-label="发送" disabled={!draft.trim() || isSending} type="submit">{isSending ? "…" : sendIcon}</button>
      </form>
      {(liveMessages.length > 0 || error) && (
        <div className="mini-copilot-live-thread" aria-live="polite">
          {liveMessages.map((message) => (
            <article className={message.role === "user" ? "user" : ""} key={message.id}>
              {message.role === "assistant" && <span className="ai-avatar">A</span>}
              <p>{message.content}</p>
            </article>
          ))}
          {error && <p className="mini-copilot-error" role="alert">{error}</p>}
        </div>
      )}
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
    { id: `user-${userMessage.id}`, role: "user" as const, content: userMessage.content },
    { id: `assistant-${assistantMessage.id}`, role: "assistant" as const, content: assistantMessage.content }
  ];
}

export default MiniCopilot;
