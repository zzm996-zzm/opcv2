import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { copilotApi, type CompareAnswer, type CopilotAIRun, type CopilotMemory, type CopilotMessage, type CopilotModelOption, type CopilotThread, type ModelSmokeResult } from "../lib/copilotApi";

export type CopilotVariant = "home" | "new" | "models" | "files" | "memories" | "compare" | "rename" | "delete";

const quickActions = [
  ["trend", "分析项目机会"],
  ["briefcase", "推荐工具"],
  ["doc", "制定落地计划"],
  ["page", "总结当前页面"]
] as const;

type ConversationItem = {
  title: string;
  desc: string;
  time: string;
  active?: boolean;
  menu?: boolean;
  group?: string;
};

const conversations: ConversationItem[] = [
  { title: "智能客服系统项目机会分析", desc: "分析市场机会、推荐工具与...", time: "10:32", active: true, menu: true },
  { title: "竞争对手监测方案设计", desc: "如何搭建竞品监测体系?", time: "09:15" },
  { title: "CRM客户管理落地计划", desc: "制定阶段性落地路线图", time: "08:47" },
  { title: "数据资产治理方法论", desc: "企业数据治理的5步进阶步骤", time: "16:22", group: "昨天" },
  { title: "GEO获客策略建议", desc: "针对SaaS产品的获客策略", time: "14:08" },
  { title: "AI教学课程内容设计", desc: "设计面向销售团队的AI课程", time: "11:30" },
  { title: "商业沙盘模拟复盘", desc: "本次沙盘的关键复盘点", time: "06-24", group: "更早" },
  { title: "增长测算模型搭建", desc: "建立业务增长测算模型", time: "06-23" }
] as const;

type CopilotModel = {
  name: string;
  value: string;
  icon: "swirl" | "ai" | "black";
  selected?: boolean;
};

type OptimisticMessageInput = {
  content: string;
  metadata?: CopilotMessage["metadata"];
  model?: string;
  role: CopilotMessage["role"];
  threadID: number;
  userID: number;
};

const fallbackModels: CopilotModel[] = [
  { name: "DeepSeek", value: "deepseek", icon: "swirl", selected: true },
  { name: "GPT-4o", value: "gpt-main", icon: "swirl" },
  { name: "Claude opus4.8", value: "claude-opus", icon: "ai" },
  { name: "Grok4.3", value: "grok", icon: "black" },
  { name: "Development", value: "development-model", icon: "black" }
] as const;

type ReferenceItem = {
  section: string;
  title: string;
  meta: string;
  kind: "pdf" | "sheet" | "chat";
  checked: boolean;
  current?: boolean;
  time?: string;
};

const referenceItems: ReferenceItem[] = [
  { section: "当前页面", title: "智能客服市场分析报告.pdf", meta: "PDF · 1.8 MB", kind: "pdf", checked: true, current: true },
  { section: "最近上传文件", title: "智能客服竞品功能对比表.xlsx", meta: "XLSX · 320 KB", kind: "sheet", checked: false, time: "昨天" },
  { section: "历史对话", title: "竞争对手监测方案设计", meta: "对话 · 2024-06-20 09:15", kind: "chat", checked: false },
  { section: "我的内容", title: "客户成功案例：某政务热线升...", meta: "文档 · 昨天", kind: "sheet", checked: false }
] as const;

const comparisonAnswers = [
  {
    model: "GPT-4o",
    icon: "swirl",
    sections: [
      ["市场规模与增长趋势", "2024年中国智能客服市场规模约为 95.2 亿元，预计到 2027 年将达 181.6 亿元，年复合增长率约 24.0%。"],
      ["竞争格局", "头部集中且持续分化，阿里云、腾讯云、百度智能云、华为云等占据主要市场份额。"],
      ["客户需求", "降本增效、提升客户体验、全渠道整合与个性化服务成为核心诉求。"],
      ["机会点", "大模型驱动的智能化升级、垂直行业解决方案、出海与多语言服务是主要机会。"]
    ]
  },
  {
    model: "Claude opus4.8",
    icon: "ai",
    sections: [
      ["市场规模与增长趋势", "2024 市场规模约 92.3 亿元，受大模型普及推动，预计 2027 年达 175.8 亿元，CAGR 为 23.3%。"],
      ["竞争格局", "市场呈现“一超多强”格局，云厂商+AI厂商+SaaS厂商协同竞争。"],
      ["客户需求", "更注重智能化水平，尤其是 AI 理解与生成能力，和业务闭环效果。"],
      ["机会点", "AI 原生应用、行业 Know-how 沉淀、数据安全与合规能力将形成差异化壁垒。"]
    ]
  },
  {
    model: "Grok4.3",
    icon: "black",
    sections: [
      ["市场规模与增长趋势", "2024 年市场规模约 98.7 亿元，增长强劲，预计 2027 年突破 190 亿元，年复合增长率 24.8%。"],
      ["竞争格局", "竞争激烈，头部厂商加速布局大模型与全渠道，区域性厂商在细分行业突围。"],
      ["客户需求", "对实时响应、复杂问题解决和数据分析洞察的需求显著提升。"],
      ["机会点", "多模态交互、客服+营销一体化、智能体落地是关键机会。"]
    ]
  }
] as const;

function CopilotPage({ variant = "home" }: { variant?: CopilotVariant }) {
  const isNew = variant === "new";
  const isCompare = variant === "compare";
  const showModelPicker = variant === "models";
  const showReferencePicker = variant === "files";
  const showMemoryPanel = variant === "memories";
  const showRename = variant === "rename";
  const showDelete = variant === "delete";
  const [threads, setThreads] = useState<CopilotThread[]>([]);
  const [availableModels, setAvailableModels] = useState<CopilotModel[]>(fallbackModels);
  const [activeThreadID, setActiveThreadID] = useState<number | null>(null);
  const [messages, setMessages] = useState<CopilotMessage[]>([]);
  const [compareQuestion, setCompareQuestion] = useState<CopilotMessage | null>(null);
  const [compareAnswers, setCompareAnswers] = useState<CompareAnswer[]>([]);
  const [compareSummary, setCompareSummary] = useState<CopilotMessage | null>(null);
  const [compareModelValues, setCompareModelValues] = useState<string[]>([fallbackModels[0].value]);
  const [memories, setMemories] = useState<CopilotMemory[]>([]);
  const [aiRuns, setAIRuns] = useState<CopilotAIRun[]>([]);
  const [smokeResult, setSmokeResult] = useState<ModelSmokeResult | null>(null);
  const [draft, setDraft] = useState("");
  const [selectedModel, setSelectedModel] = useState(fallbackModels[0].value);
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isSummarizing, setIsSummarizing] = useState(false);
  const [isSavingMemory, setIsSavingMemory] = useState(false);
  const [isTestingModel, setIsTestingModel] = useState(false);
  const [useHistoryFallback, setUseHistoryFallback] = useState(false);
  const [error, setError] = useState("");
  const chatAreaRef = useRef<HTMLDivElement | null>(null);

  const activeThread = useMemo(
    () => threads.find((thread) => thread.id === activeThreadID) ?? null,
    [activeThreadID, threads]
  );

  useEffect(() => {
    let active = true;
    setIsLoading(true);
    Promise.allSettled([copilotApi.listThreads(), copilotApi.listModels()])
      .then(([threadsResult, modelsResult]) => {
        if (!active) return;
        if (threadsResult.status === "fulfilled") {
          setThreads(threadsResult.value.threads);
          setActiveThreadID(threadsResult.value.threads[0]?.id ?? null);
          setUseHistoryFallback(false);
        } else {
          setUseHistoryFallback(true);
          setError("暂时无法加载历史会话，已显示示例内容");
        }
        if (modelsResult.status === "fulfilled" && modelsResult.value.models.length > 0) {
          const nextModels = toDisplayModels(modelsResult.value.models);
          setAvailableModels(nextModels);
          setSelectedModel(nextModels.find((model) => model.selected)?.value ?? nextModels[0].value);
          setCompareModelValues(nextModels.filter((model) => model.selected).map((model) => model.value).slice(0, 4));
        }
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!activeThreadID) {
      setMessages([]);
      return;
    }
    let active = true;
    copilotApi
      .listMessages(activeThreadID)
      .then((payload) => {
        if (!active) return;
        setMessages(payload.messages);
        if (isCompare) {
          const rebuilt = rebuildCompareState(payload.messages);
          setCompareQuestion(rebuilt.question);
          setCompareAnswers(rebuilt.answers);
          setCompareSummary(rebuilt.summary);
        }
      })
      .catch(() => {
        if (active) setError("暂时无法加载会话消息，已显示示例内容");
      });
    return () => {
      active = false;
    };
  }, [activeThreadID, isCompare]);

  useEffect(() => {
    if (!showMemoryPanel) return;
    let active = true;
    copilotApi
      .listMemories()
      .then((payload) => {
        if (active) setMemories(payload.memories);
      })
      .catch(() => {
        if (active) setError("暂时无法加载记忆");
      });
    return () => {
      active = false;
    };
  }, [showMemoryPanel]);

  useEffect(() => {
    if (!showModelPicker) return;
    let active = true;
    loadAIRuns().catch(() => {
      if (active) setError("暂时无法加载模型运行记录");
    });
    return () => {
      active = false;
    };
  }, [showModelPicker]);

  useEffect(() => {
    const chatArea = chatAreaRef.current;
    if (!chatArea) return;
    if (typeof chatArea.scrollTo === "function") {
      chatArea.scrollTo({ top: chatArea.scrollHeight, behavior: "smooth" });
    } else {
      chatArea.scrollTop = chatArea.scrollHeight;
    }
  }, [messages, compareQuestion, compareAnswers, compareSummary, isSending, isSummarizing]);

  async function startThread(initialContent?: string) {
    const title = (initialContent || "新会话").slice(0, 28);
    const thread = await copilotApi.createThread({ title, mode: isCompare ? "compare" : "chat", model: selectedModel });
    setThreads((current) => [thread, ...current]);
    setActiveThreadID(thread.id);
    setMessages([]);
    return thread;
  }

  async function handleSend(event: FormEvent) {
    event.preventDefault();
    const content = draft.trim();
    if (!content || isSending) return;
    setDraft("");
    setError("");
    setIsSending(true);
    try {
      const thread = activeThreadID ? activeThread : await startThread(content);
      if (!thread) return;
      if (isCompare) {
        const optimisticQuestion = optimisticMessage({
          content,
          model: compareModelValues.join(","),
          role: "user",
          threadID: thread.id,
          userID: thread.user_id,
          metadata: { kind: "compare_question" }
        });
        const models = compareModelValues.length > 0 ? compareModelValues : [selectedModel];
        setCompareQuestion(optimisticQuestion);
        setCompareAnswers([]);
        setCompareSummary(null);
        const result = await copilotApi.compareMessages(thread.id, { content, models });
        setCompareQuestion(result.user_message);
        setCompareAnswers(result.answers);
        setCompareSummary(null);
        setThreads((current) => current.map((item) => item.id === thread.id ? { ...item, updated_at: result.user_message.created_at } : item));
        return;
      }
      const optimisticUserMessage = optimisticMessage({
        content,
        model: selectedModel,
        role: "user",
        threadID: thread.id,
        userID: thread.user_id
      });
      setMessages((current) => [...current, optimisticUserMessage]);
      const result = await copilotApi.sendMessage(thread.id, { content, model: selectedModel });
      setMessages((current) => [
        ...current.filter((message) => message.id !== optimisticUserMessage.id),
        result.user_message,
        result.assistant_message
      ]);
      setThreads((current) => current.map((item) => item.id === thread.id ? { ...item, updated_at: result.assistant_message.created_at } : item));
    } catch (requestError) {
      setDraft(content);
      setMessages((current) => current.filter((message) => message.id >= 0));
      if (isCompare) {
        setCompareQuestion((current) => current && current.id < 0 ? null : current);
      }
      setError(requestError instanceof Error ? requestError.message : "发送失败，请稍后重试");
    } finally {
      setIsSending(false);
    }
  }

  async function handleNewThread() {
    try {
      await startThread();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "新建会话失败");
    }
  }

  async function handleRenameThread(title: string) {
    if (!activeThreadID) return;
    try {
      const updated = await copilotApi.renameThread(activeThreadID, title);
      setThreads((current) => current.map((thread) => thread.id === updated.id ? updated : thread));
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "重命名失败，请稍后重试");
    }
  }

  async function handleArchiveThread() {
    if (!activeThreadID) return;
    try {
      const archivedID = activeThreadID;
      await copilotApi.archiveThread(archivedID);
      setThreads((current) => current.filter((thread) => thread.id !== archivedID));
      setMessages([]);
      setActiveThreadID((current) => {
        if (current !== archivedID) return current;
        const nextThread = threads.find((thread) => thread.id !== archivedID);
        return nextThread?.id ?? null;
      });
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "删除会话失败，请稍后重试");
    }
  }

  async function handleSaveMemory(input: { key: string; value: string }) {
    setIsSavingMemory(true);
    setError("");
    try {
      const memory = await copilotApi.saveMemory({ ...input, confidence: 1, source: "manual" });
      setMemories((current) => [memory, ...current.filter((item) => item.id !== memory.id && item.key !== memory.key)]);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "保存记忆失败，请稍后重试");
    } finally {
      setIsSavingMemory(false);
    }
  }

  async function loadAIRuns() {
    const payload = await copilotApi.listAIRuns();
    setAIRuns(payload.runs);
  }

  async function handleSmokeModel() {
    setIsTestingModel(true);
    setError("");
    try {
      const result = await copilotApi.smokeModel({ model: selectedModel, prompt: "ping" });
      setSmokeResult(result);
      await loadAIRuns();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "模型联调失败，请查看运行记录");
      try {
        await loadAIRuns();
      } catch {
        // Keep the original smoke error visible.
      }
    } finally {
      setIsTestingModel(false);
    }
  }

  async function handleDeleteMemory(id: number) {
    try {
      await copilotApi.deleteMemory(id);
      setMemories((current) => current.filter((memory) => memory.id !== id));
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "删除记忆失败，请稍后重试");
    }
  }

  function handleToggleCompareModel(value: string) {
    setCompareModelValues((current) => {
      if (current.includes(value)) {
        if (current.length === 1) return current;
        return current.filter((item) => item !== value);
      }
      return [...current, value].slice(0, 4);
    });
  }

  async function handleSummarizeComparison() {
    if (!activeThreadID || !compareQuestion || compareAnswers.length === 0 || isSummarizing) return;
    setIsSummarizing(true);
    setError("");
    try {
      const result = await copilotApi.summarizeComparison(activeThreadID, {
        content: compareQuestion.content,
        model: selectedModel,
        answers: compareAnswers
      });
      setCompareSummary(result.summary_message);
      setThreads((current) => current.map((item) => item.id === activeThreadID ? { ...item, updated_at: result.summary_message.created_at } : item));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "总结失败，请稍后重试");
    } finally {
      setIsSummarizing(false);
    }
  }

  return (
    <V4PageShell className="copilot-shell" showCopilotMini={false}>
      <section
        className={"copilot-workbench " + (isCompare ? "compare-mode " : "") + (showRename || showDelete ? "modal-open" : "")}
        aria-label="智活 Copilot 工作台"
      >
        <div className="copilot-main-panel">
          {isCompare ? (
            <CompareHeader
              models={availableModels}
              selectedModels={compareModelValues}
              onToggleModel={handleToggleCompareModel}
            />
          ) : <CopilotHeader />}

          <div className="copilot-chat-area" ref={chatAreaRef}>
            {showModelPicker && (
              <ModelDiagnosticsPanel
                isTesting={isTestingModel}
                modelLabel={availableModels.find((model) => model.value === selectedModel)?.name ?? selectedModel}
                runs={aiRuns}
                smokeResult={smokeResult}
                onSmoke={handleSmokeModel}
              />
            )}
            {isCompare ? (
              <CompareConversation
                answers={compareAnswers}
                isSending={isSending}
                isSummarizing={isSummarizing}
                models={availableModels}
                onSummarize={handleSummarizeComparison}
                question={compareQuestion}
                summary={compareSummary}
              />
            ) : isNew ? (
              <EmptyConversation onPrompt={(prompt) => setDraft(prompt)} />
            ) : (
              <ChatThread messages={messages} isSending={isSending} useFallback={messages.length === 0 && !isLoading} />
            )}
          </div>

          <Composer
            draft={draft}
            model={selectedModel}
            models={availableModels}
            isSending={isSending}
            error={error}
            onDraftChange={setDraft}
            onModelChange={setSelectedModel}
            onSubmit={handleSend}
            showModelPicker={showModelPicker}
            showReferencePicker={showReferencePicker}
            showMemoryPanel={showMemoryPanel}
            memories={memories}
            isSavingMemory={isSavingMemory}
            onSaveMemory={handleSaveMemory}
            onDeleteMemory={handleDeleteMemory}
            compare={isCompare}
          />
        </div>

        <ConversationSidebar
          activeThreadID={activeThreadID}
          threads={threads}
          useFallback={useHistoryFallback}
          showThreadMenu={showModelPicker}
          onSelectThread={setActiveThreadID}
          onNewThread={handleNewThread}
        />
        {showRename && <RenameDialog title={activeThread?.title ?? "智能客服系统项目机会分析"} onConfirm={handleRenameThread} />}
        {showDelete && <DeleteDialog onConfirm={handleArchiveThread} />}
      </section>
    </V4PageShell>
  );
}

function CopilotHeader() {
  return (
    <header className="copilot-workbench-head">
      <div className="copilot-brand-mark" aria-hidden="true">
        <span className="v4-logo" />
      </div>
      <div>
        <h1>智活 Copilot</h1>
        <p>你的全球 AI 助手，随时为你提供专业的分析与建议</p>
      </div>
      <nav className="copilot-quick-actions" aria-label="Copilot 快捷指令">
        {quickActions.map(([icon, label]) => (
          <button key={label} type="button">
            <span className={"copilot-ui-icon " + icon} aria-hidden="true" />
            {label}
          </button>
        ))}
      </nav>
    </header>
  );
}

function CompareHeader({
  models,
  selectedModels,
  onToggleModel
}: {
  models: CopilotModel[];
  selectedModels: string[];
  onToggleModel: (value: string) => void;
}) {
  const selectedCount = selectedModels.length;
  const displayedModels = models.slice(0, 6);

  return (
    <header className="copilot-compare-head">
      <div className="copilot-compare-icon" aria-hidden="true">
        <span className="copilot-ui-icon trend" />
      </div>
      <div>
        <h1>AI 对比分析</h1>
        <p>同时对比多个 AI 模型的回答，获得更全面、客观的洞察</p>
      </div>
      <div className="copilot-compare-settings" aria-label="对比设置">
        <strong>对比设置:</strong>
        <button type="button">{Math.min(selectedCount, 4)} 模型</button>
        {displayedModels.map((model) => {
          const selected = selectedModels.includes(model.value);
          const disabled = !selected && selectedModels.length >= 4;
          return (
          <button
            aria-label={`${selected ? "取消选择" : "选择模型"} ${model.name}`}
            className={selected ? "active" : ""}
            disabled={disabled}
            key={model.value}
            onClick={() => onToggleModel(model.value)}
            type="button"
          >
            <span className={"model-glyph " + model.icon} aria-hidden="true" />
            {model.name}
            {selected && <i aria-hidden="true">×</i>}
          </button>
          );
        })}
      </div>
    </header>
  );
}

function ChatThread({
  messages,
  isSending,
  useFallback
}: {
  messages: CopilotMessage[];
  isSending: boolean;
  useFallback?: boolean;
}) {
  if (!useFallback && messages.length > 0) {
    return (
      <div className="copilot-chat-thread" aria-label="会话内容">
        {messages.map((message) => (
          <article key={message.id} className={"copilot-message " + (message.role === "user" ? "user" : "assistant")}>
            {message.role !== "user" && <span className="v4-logo" aria-hidden="true" />}
            <div className={message.role === "user" ? "" : "copilot-bubble compact"}>
              <p>{message.content}</p>
              <time>{formatTime(message.created_at)}</time>
            </div>
            {message.role === "user" && <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>}
          </article>
        ))}
        {isSending && <ThinkingMessage />}
      </div>
    );
  }

  return (
    <div className="copilot-chat-thread" aria-label="会话内容">
      <article className="copilot-message user">
        <p>请帮我分析智能客服系统的市场机会和竞争格局。</p>
        <time>10:32 ✓✓</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>

      <article className="copilot-message assistant">
        <span className="v4-logo" aria-hidden="true" />
        <div className="copilot-bubble compact">
          <p>好的，我将从市场规模、增长趋势、竞争格局、客户需求与机会点四个维度为你分析智能客服系统的市场机会。</p>
          <time>10:32</time>
        </div>
      </article>

      <article className="copilot-message assistant report">
        <span className="v4-logo" aria-hidden="true" />
        <div className="copilot-bubble report-card">
          <h2>一、市场规模与增长趋势</h2>
          <ul>
            <li>2024年中国智能客服市场规模约为 95.2 亿元，预计 2027 年将达到 181.6 亿元，年复合增长率约 24.0%。</li>
            <li>受益于企业降本增效、用户体验提升与大模型技术普及，市场保持高速增长。</li>
          </ul>
          <h2>二、竞争格局</h2>
          <ul>
            <li>第一梯队：阿里云、腾讯云、百度智能云、华为云等，具备强大技术与生态能力。</li>
            <li>第二梯队：容联云、智齿科技、环信等，聚焦垂直场景与中大型客户。</li>
            <li>新兴玩家：大量AI原生创业公司，依托大模型+场景化能力切入细分赛道。</li>
          </ul>
          <div className="copilot-file-chip">
            <span className="pdf-thumb" aria-hidden="true">PDF</span>
            <span>
              <strong>智能客服市场分析报告.pdf</strong>
              <small>PDF · 1.8 MB</small>
            </span>
          </div>
          <time>10:33</time>
        </div>
      </article>

      <article className="copilot-message user lower">
        <p>请基于上面的分析，推荐适合我们的工具和落地路径。</p>
        <time>10:34 ✓✓</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>

      {isSending && <ThinkingMessage />}
    </div>
  );
}

function ThinkingMessage() {
  return (
    <article className="copilot-message assistant thinking">
      <span className="v4-logo" aria-hidden="true" />
      <div className="copilot-thinking">
        正在思考中
        <span />
        <span />
        <span />
        <span />
      </div>
    </article>
  );
}

function EmptyConversation({ onPrompt }: { onPrompt: (prompt: string) => void }) {
  return (
    <div className="copilot-empty-state">
      <div className="copilot-empty-card">
        <div className="empty-orbit" aria-hidden="true">
          <span />
          <i />
        </div>
        <h1>开始一段新的对话</h1>
        <p>向智活 Copilot 提问，获取专业的分析与建议</p>
        <div className="copilot-prompt-list">
          {["分析一个新项目机会", "帮我制定执行计划", "总结当前页面"].map((prompt) => (
            <button key={prompt} onClick={() => onPrompt(prompt)} type="button">
              {prompt}
              <span aria-hidden="true">↗</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

function CompareConversation({
  answers,
  isSending,
  isSummarizing,
  models,
  onSummarize,
  question,
  summary
}: {
  answers: CompareAnswer[];
  isSending: boolean;
  isSummarizing: boolean;
  models: CopilotModel[];
  onSummarize: () => void;
  question: CopilotMessage | null;
  summary: CopilotMessage | null;
}) {
  const hasBackendAnswers = answers.length > 0;

  return (
    <div className="copilot-comparison">
      <article className="copilot-message user compare-question">
        <p>{question?.content ?? "请分析 2024 年中国智能客服市场的规模、增长趋势、竞争格局、客户需求与机会点。"}</p>
        <time>{question ? formatTime(question.created_at) : "10:35 ✓"}</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>
      {isSending && <ThinkingMessage />}
      <div className="comparison-grid">
        {hasBackendAnswers ? answers.map((answer) => (
          <article key={answer.model} className="comparison-card">
            <header>
              <span className={"model-glyph " + modelIcon(answer.model, 0)} aria-hidden="true" />
              <h2>{modelLabel(answer.model, models)}</h2>
              <b>{answer.error_code ? "回答失败" : "回答完成"}</b>
              <time>{formatTime(answer.assistant_message.created_at)}</time>
            </header>
            <section>
              <p>{answer.assistant_message.content}</p>
            </section>
          </article>
        )) : comparisonAnswers.map((answer) => (
          <article key={answer.model} className="comparison-card">
            <header>
              <span className={"model-glyph " + answer.icon} aria-hidden="true" />
              <h2>{answer.model}</h2>
              <b>回答完成</b>
              <time>10:35</time>
            </header>
            {answer.sections.map(([title, body], index) => (
              <section key={title}>
                <h3>{index + 1}、{title}</h3>
                <p>{body}</p>
              </section>
            ))}
          </article>
        ))}
      </div>
      {summary && (
        <article className="comparison-card comparison-summary-card">
          <header>
            <span className="copilot-ui-icon doc" aria-hidden="true" />
            <h2>综合结论</h2>
            <b>总结完成</b>
            <time>{formatTime(summary.created_at)}</time>
          </header>
          <section>
            <p>{summary.content}</p>
          </section>
        </article>
      )}
      <button className="compare-summary-button" disabled={!hasBackendAnswers || isSummarizing} onClick={onSummarize} type="button">
        <span className="copilot-ui-icon doc" aria-hidden="true" />
        {isSummarizing ? "正在总结" : "总结本次对比分析"}
      </button>
    </div>
  );
}

function Composer({
  draft,
  model,
  models,
  memories,
  isSending,
  isSavingMemory,
  error,
  onDraftChange,
  onModelChange,
  onSaveMemory,
  onDeleteMemory,
  onSubmit,
  showModelPicker,
  showReferencePicker,
  showMemoryPanel,
  compare
}: {
  draft: string;
  model: string;
  models: CopilotModel[];
  memories: CopilotMemory[];
  isSending: boolean;
  isSavingMemory: boolean;
  error: string;
  onDraftChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onSaveMemory: (input: { key: string; value: string }) => void;
  onDeleteMemory: (id: number) => void;
  onSubmit: (event: FormEvent) => void;
  showModelPicker?: boolean;
  showReferencePicker?: boolean;
  showMemoryPanel?: boolean;
  compare?: boolean;
}) {
  const selectedModelLabel = models.find((item) => item.value === model)?.name ?? model;

  return (
    <div className="copilot-composer-wrap">
      {showModelPicker && <ModelPicker models={models} selectedModel={model} onSelect={onModelChange} />}
      {showReferencePicker && <ReferencePicker />}
      {showMemoryPanel && (
        <MemoryPanel
          isSaving={isSavingMemory}
          memories={memories}
          onDelete={onDeleteMemory}
          onSave={onSaveMemory}
        />
      )}
      <form className="copilot-composer" aria-label="Copilot 输入框" onSubmit={onSubmit}>
        <label className="sr-only" htmlFor="copilot-question">输入你的问题</label>
        <input
          id="copilot-question"
          onChange={(event) => onDraftChange(event.target.value)}
          placeholder="输入你的问题，Enter 发送、Shift + Enter 换行"
          value={draft}
        />
        <div className="composer-toolbar">
          <button type="button">
            <span className="copilot-ui-icon clip" aria-hidden="true" />
            上传文件
          </button>
          <button className={showReferencePicker ? "active" : ""} type="button">
            <span className="copilot-ui-icon link" aria-hidden="true" />
            引用
          </button>
          <Link className={showMemoryPanel ? "active" : ""} to="/copilot/memories">
            <span className="copilot-ui-icon memory" aria-hidden="true" />
            记忆
          </Link>
          <button className={showModelPicker ? "active" : ""} type="button" disabled={compare}>
            <span className="model-glyph swirl" aria-hidden="true" />
            {selectedModelLabel}
            <span aria-hidden="true">{showModelPicker ? "⌃" : "⌄"}</span>
          </button>
          <button className="compare-chip active" type="button">
            <span className="copilot-ui-icon doc" aria-hidden="true" />
            AI 对比分析
            <b>New</b>
          </button>
          <button type="button">
            <span className="copilot-ui-icon think" aria-hidden="true" />
            深度思考
            <b className="vip">VIP</b>
          </button>
          <span className="composer-spacer" />
          <button aria-label="语音输入" className="icon-only" type="button">
            <span className="mic-icon" aria-hidden="true" />
          </button>
          <button aria-label="发送" className="send-button" disabled={isSending || !draft.trim()} type="submit">
            <span aria-hidden="true">↗</span>
          </button>
        </div>
      </form>
      {error && <p className="ai-disclaimer">{error}</p>}
      <p className="ai-disclaimer">ⓘ 内容由 AI 生成，请注意甄别准确性</p>
    </div>
  );
}

function MemoryPanel({
  memories,
  isSaving,
  onSave,
  onDelete
}: {
  memories: CopilotMemory[];
  isSaving: boolean;
  onSave: (input: { key: string; value: string }) => void;
  onDelete: (id: number) => void;
}) {
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const nextKey = key.trim();
    const nextValue = value.trim();
    if (!nextKey || !nextValue || isSaving) return;
    onSave({ key: nextKey, value: nextValue });
    setKey("");
    setValue("");
  }

  return (
    <div className="copilot-popover memory-panel" role="dialog" aria-label="记忆">
      <h2>记忆</h2>
      <form onSubmit={handleSubmit}>
        <label>
          <span>记忆键</span>
          <input aria-label="记忆键" onChange={(event) => setKey(event.target.value)} value={key} />
        </label>
        <label>
          <span>记忆内容</span>
          <input aria-label="记忆内容" onChange={(event) => setValue(event.target.value)} value={value} />
        </label>
        <button disabled={isSaving || !key.trim() || !value.trim()} type="submit">保存记忆</button>
      </form>
      <div className="memory-list">
        {memories.length === 0 ? (
          <p>暂无记忆</p>
        ) : memories.map((memory) => (
          <article key={memory.id}>
            <span>
              <strong>{memory.key}</strong>
              <small>{memory.value}</small>
            </span>
            <button aria-label={`删除记忆 ${memory.key}`} onClick={() => onDelete(memory.id)} type="button">删除</button>
          </article>
        ))}
      </div>
    </div>
  );
}

function ModelPicker({
  models,
  selectedModel,
  onSelect
}: {
  models: CopilotModel[];
  selectedModel: string;
  onSelect: (model: string) => void;
}) {
  return (
    <div className="copilot-popover model-picker" role="dialog" aria-label="模型选择">
      {models.map((model) => (
        <button key={model.value} className={model.value === selectedModel ? "selected" : ""} onClick={() => onSelect(model.value)} type="button">
          <span className={"model-glyph " + model.icon} aria-hidden="true" />
          {model.name}
          {model.value === selectedModel && <b aria-hidden="true">✓</b>}
        </button>
      ))}
    </div>
  );
}

function ModelDiagnosticsPanel({
  isTesting,
  modelLabel,
  runs,
  smokeResult,
  onSmoke
}: {
  isTesting: boolean;
  modelLabel: string;
  runs: CopilotAIRun[];
  smokeResult: ModelSmokeResult | null;
  onSmoke: () => void;
}) {
  return (
    <section className="model-diagnostics-panel" aria-label="模型联调">
      <header>
        <span className="copilot-ui-icon think" aria-hidden="true" />
        <div>
          <h2>模型联调</h2>
          <p>测试当前模型路由，并查看最近 Copilot AI run 的安全诊断信息。</p>
        </div>
        <button disabled={isTesting} onClick={onSmoke} type="button">
          {isTesting ? "测试中" : `测试 ${modelLabel}`}
        </button>
      </header>
      {smokeResult && (
        <div className="model-smoke-result" role="status">
          <strong>{smokeResult.model}</strong>
          <span>{smokeResult.reply}</span>
          <small>{smokeResult.input_tokens ?? 0} / {smokeResult.output_tokens ?? 0} tokens</small>
        </div>
      )}
      <div className="ai-run-list">
        {runs.length === 0 ? (
          <p>暂无运行记录</p>
        ) : runs.slice(0, 5).map((run) => (
          <article key={run.id} className={run.status === "failed" ? "failed" : "completed"}>
            <span>
              <strong>{run.model || run.feature}</strong>
              <small>{run.feature} · {formatTime(run.created_at)}</small>
            </span>
            <code>{run.error_code || run.status}</code>
            <em>{run.latency_ms ?? 0}ms</em>
          </article>
        ))}
      </div>
    </section>
  );
}

function ReferencePicker() {
  let previousSection = "";

  return (
    <div className="copilot-popover reference-picker" role="dialog" aria-label="引用">
      <h2>引用</h2>
      <label className="reference-search">
        <span aria-hidden="true">⌕</span>
        <input placeholder="搜索文件、对话或我的内容" />
      </label>
      <div className="reference-list">
        {referenceItems.map((item) => {
          const showSection = item.section !== previousSection;
          previousSection = item.section;

          return (
            <div key={item.title}>
              {showSection && <h3>{item.section}</h3>}
              <label className={item.checked ? "checked" : ""}>
                <input checked={item.checked} readOnly type="checkbox" />
                <span className={"reference-file " + item.kind} aria-hidden="true" />
                <span>
                  <strong>{item.title}</strong>
                  <small>{item.meta}</small>
                </span>
                {item.current && <b>当前</b>}
                {item.time && <em>{item.time}</em>}
              </label>
            </div>
          );
        })}
      </div>
      <footer>
        <span>已选择 1 项</span>
        <button type="button">插入引用</button>
      </footer>
    </div>
  );
}

function ConversationSidebar({
  activeThreadID,
  threads,
  useFallback,
  showThreadMenu,
  onSelectThread,
  onNewThread
}: {
  activeThreadID: number | null;
  threads: CopilotThread[];
  useFallback: boolean;
  showThreadMenu?: boolean;
  onSelectThread: (id: number) => void;
  onNewThread: () => void;
}) {
  let lastGroup = "";
  const displayThreads = threads.length > 0 ? threads.map((thread) => ({
    id: thread.id,
    title: thread.title,
    desc: thread.model || "Copilot 会话",
    time: formatTime(thread.updated_at),
    active: thread.id === activeThreadID
  })) : useFallback ? conversations.map((item, index) => ({
    id: -index - 1,
    title: item.title,
    desc: item.desc,
    time: item.time,
    active: item.active,
    group: item.group
  })) : [];

  return (
    <aside className="copilot-history" aria-label="会话记录">
      <header>
        <button aria-label="收起会话记录" type="button">»</button>
        <button onClick={onNewThread} type="button">＋ 新建会话</button>
      </header>
      <label className="history-search">
        <span aria-hidden="true">⌕</span>
        <input placeholder="搜索会话" />
      </label>
      <div className="history-list">
        {displayThreads.map((item, index) => {
          const group = "group" in item && item.group ? item.group : (index < 3 ? "今天" : "");
          const showGroup = group && group !== lastGroup;
          if (group) lastGroup = group;

          return (
            <div key={item.title}>
              {showGroup && <h2>{group}</h2>}
              <article className={item.active ? "active" : ""}>
                <Link onClick={() => item.id > 0 && onSelectThread(item.id)} to="/copilot">
                  <strong>{item.title}</strong>
                  <span>{item.desc}</span>
                </Link>
                <time>{item.time}</time>
                {item.active && (
                  <button aria-label="会话更多操作" className="history-more" type="button">
                    ⋮
                  </button>
                )}
              </article>
            </div>
          );
        })}
      </div>
      {showThreadMenu && (
        <div className="thread-action-menu" role="menu" aria-label="会话操作">
          <Link role="menuitem" to="/copilot/rename">重命名</Link>
          <Link role="menuitem" to="/copilot/delete">删除</Link>
        </div>
      )}
    </aside>
  );
}

function formatTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false });
}

function toDisplayModels(models: CopilotModelOption[]): CopilotModel[] {
  return models.map((model, index) => ({
    name: model.name || model.value,
    value: model.value,
    icon: modelIcon(model.value, index),
    selected: model.is_default
  }));
}

function optimisticMessage(input: OptimisticMessageInput): CopilotMessage {
  return {
    id: -Date.now(),
    user_id: input.userID,
    thread_id: input.threadID,
    role: input.role,
    content: input.content,
    status: "completed",
    model: input.model,
    metadata: input.metadata,
    created_at: new Date().toISOString()
  };
}

function rebuildCompareState(messages: CopilotMessage[]): {
  question: CopilotMessage | null;
  answers: CompareAnswer[];
  summary: CopilotMessage | null;
} {
  const metadataQuestion = [...messages].reverse().find((message) => message.metadata?.kind === "compare_question") ?? null;
  const metadataAnswers = messages.filter((message) => message.metadata?.kind === "compare_answer");
  const metadataSummary = [...messages].reverse().find((message) => message.metadata?.kind === "compare_summary") ?? null;
  if (metadataQuestion || metadataAnswers.length > 0 || metadataSummary) {
    return {
      question: metadataQuestion,
      answers: metadataAnswers.map((message) => ({
        model: message.model || "unknown",
        assistant_message: message,
        error_code: message.error_code
      })),
      summary: metadataSummary
    };
  }

  const lastUserIndex = messages.map((message) => message.role).lastIndexOf("user");
  if (lastUserIndex < 0) {
    return { question: null, answers: [], summary: null };
  }
  const question = messages[lastUserIndex];
  const assistantMessages = messages.slice(lastUserIndex + 1).filter((message) => message.role === "assistant");
  const summary = assistantMessages.length > 1 ? assistantMessages[assistantMessages.length - 1] : null;
  const answerMessages = summary ? assistantMessages.slice(0, -1) : assistantMessages;
  return {
    question,
    answers: answerMessages.map((message) => ({
      model: message.model || "unknown",
      assistant_message: message,
      error_code: message.error_code
    })),
    summary
  };
}

function modelIcon(value: string, index: number): CopilotModel["icon"] {
  if (value.includes("deepseek") || value.includes("gpt")) return "swirl";
  if (value.includes("claude")) return "ai";
  if (index % 3 === 1) return "ai";
  if (index % 3 === 2) return "black";
  return "swirl";
}

function modelLabel(value: string, models: CopilotModel[]) {
  return models.find((model) => model.value === value)?.name ?? value;
}

function RenameDialog({ title, onConfirm }: { title: string; onConfirm: (title: string) => void }) {
  const [value, setValue] = useState(title);

  return (
    <div className="copilot-modal-backdrop">
      <section className="copilot-dialog rename-dialog" role="dialog" aria-modal="true" aria-label="重命名会话">
        <button aria-label="关闭" className="dialog-close" type="button">×</button>
        <h2>重命名会话</h2>
        <label className="sr-only" htmlFor="conversation-name">会话名称</label>
        <input id="conversation-name" onChange={(event) => setValue(event.target.value)} value={value} />
        <p>修改后仅影响会话标题，不影响对话内容</p>
        <footer>
          <button type="button">取消</button>
          <button className="primary" onClick={() => onConfirm(value)} type="button">确认重命名</button>
        </footer>
      </section>
    </div>
  );
}

function DeleteDialog({ onConfirm }: { onConfirm: () => void }) {
  const [confirmed, setConfirmed] = useState(false);

  return (
    <div className="copilot-modal-backdrop">
      <section className="copilot-dialog delete-dialog" role="dialog" aria-modal="true" aria-label="删除会话">
        <button aria-label="关闭" className="dialog-close" type="button">×</button>
        <div className="delete-warning" aria-hidden="true">!</div>
        <h2>删除会话</h2>
        <p>删除后，该会话的所有内容（包括对话记录及引用的附件）将从历史会话列表中移除，且无法恢复。</p>
        <p>此操作不可撤销，请谨慎确认。</p>
        <label className="confirm-delete">
          <input checked={confirmed} onChange={(event) => setConfirmed(event.target.checked)} type="checkbox" />
          我已确认删除此会话
        </label>
        <footer>
          <button type="button">取消</button>
          <button className="danger" disabled={!confirmed} onClick={onConfirm} type="button">确认删除</button>
        </footer>
      </section>
    </div>
  );
}

export default CopilotPage;
