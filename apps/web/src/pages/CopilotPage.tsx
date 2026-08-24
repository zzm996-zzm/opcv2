import { ChangeEvent, DragEvent, FormEvent, forwardRef, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import MarkdownMessage from "../components/MarkdownMessage";
import { ApiRequestError } from "../lib/apiRequest";
import { copilotApi, type CompareAnswer, type CopilotAIRun, type CopilotFile, type CopilotMemory, type CopilotMessage, type CopilotModelOption, type CopilotThread, type ModelSmokeResult, type SendMessageResult } from "../lib/copilotApi";
import { membershipApi, type MembershipUsageItem } from "../lib/membershipApi";
import { quotaKeys, quotaSummary } from "../lib/quotaUsage";

export type CopilotVariant = "home" | "new" | "models" | "files" | "memories" | "compare" | "rename" | "delete";
type ComposerPopover = "models" | "files" | "memories" | null;
const MAX_COMPARE_MODELS = 3;

type SpeechRecognitionLike = {
  lang: string;
  continuous: boolean;
  interimResults: boolean;
  onresult: ((event: SpeechRecognitionEventLike) => void) | null;
  onerror: (() => void) | null;
  onend: (() => void) | null;
  start: () => void;
  stop: () => void;
};

type SpeechRecognitionEventLike = {
  resultIndex: number;
  results: ArrayLike<{
    isFinal: boolean;
    0: { transcript: string };
  }>;
};

type SpeechRecognitionConstructor = new () => SpeechRecognitionLike;

const quickActions = [
  ["trend", "分析项目机会", "请帮我分析当前项目的市场机会、目标客户、竞争格局和落地风险。"],
  ["briefcase", "推荐工具", "请根据当前目标，推荐适合的 AI 工具、使用场景、成本和落地优先级。"],
  ["doc", "制定落地计划", "请帮我制定一份可执行的落地计划，包含阶段目标、关键任务、负责人和验收标准。"],
  ["page", "总结当前页面", "请总结当前页面的核心信息，并提炼下一步最应该推进的行动。"]
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
  { name: "GPT-4o", value: "gpt-main", icon: "swirl", selected: true },
  { name: "Claude Opus 4.8", value: "claude-opus", icon: "ai" },
  { name: "Grok 4.3", value: "grok", icon: "black" }
] as const;

function CopilotPage({ variant = "home" }: { variant?: CopilotVariant }) {
  const isNew = variant === "new";
  const isCompare = variant === "compare";
  const showModelPicker = variant === "models";
  const showReferencePicker = variant === "files";
  const showMemoryPanel = variant === "memories";
  const showRename = variant === "rename";
  const showDelete = variant === "delete";
  const routePopover = useMemo<ComposerPopover>(() => {
    if (showModelPicker) return "models";
    if (showReferencePicker) return "files";
    if (showMemoryPanel) return "memories";
    return null;
  }, [showMemoryPanel, showModelPicker, showReferencePicker]);
  const [threads, setThreads] = useState<CopilotThread[]>([]);
  const [availableModels, setAvailableModels] = useState<CopilotModel[]>(fallbackModels);
  const [activeThreadID, setActiveThreadID] = useState<number | null>(null);
  const [messages, setMessages] = useState<CopilotMessage[]>([]);
  const [compareQuestion, setCompareQuestion] = useState<CopilotMessage | null>(null);
  const [compareAnswers, setCompareAnswers] = useState<CompareAnswer[]>([]);
  const [compareSummary, setCompareSummary] = useState<CopilotMessage | null>(null);
  const [compareModelValues, setCompareModelValues] = useState<string[]>(fallbackModels.map((model) => model.value));
  const [memories, setMemories] = useState<CopilotMemory[]>([]);
  const [files, setFiles] = useState<CopilotFile[]>([]);
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const [selectedReferenceIDs, setSelectedReferenceIDs] = useState<number[]>([]);
  const [aiRuns, setAIRuns] = useState<CopilotAIRun[]>([]);
  const [smokeResult, setSmokeResult] = useState<ModelSmokeResult | null>(null);
  const [draft, setDraft] = useState("");
  const [activePopover, setActivePopover] = useState<ComposerPopover>(routePopover);
  const [typewriterContent, setTypewriterContent] = useState<Record<number, string>>({});
  const [streamingContent, setStreamingContent] = useState("");
  const [selectedModel, setSelectedModel] = useState(fallbackModels[0].value);
  const [isSending, setIsSending] = useState(false);
  const [isSummarizing, setIsSummarizing] = useState(false);
  const [isSavingMemory, setIsSavingMemory] = useState(false);
  const [isSavingFile, setIsSavingFile] = useState(false);
  const [isTestingModel, setIsTestingModel] = useState(false);
	const [toolBusyID, setToolBusyID] = useState<number | null>(null);
  const [historyCollapsed, setHistoryCollapsed] = useState(false);
  const [error, setError] = useState("");
  const chatAreaRef = useRef<HTMLDivElement | null>(null);
  const typewriterTimersRef = useRef<Map<number, number>>(new Map());
  const composerWrapRef = useRef<HTMLDivElement | null>(null);
  const sendAbortRef = useRef<AbortController | null>(null);

  const activeThread = useMemo(
    () => threads.find((thread) => thread.id === activeThreadID) ?? null,
    [activeThreadID, threads]
  );
  const activeQuota = quotaSummary(
    usage,
    isCompare ? quotaKeys.copilotCompareCalls : quotaKeys.copilotMessages,
    isCompare ? "Copilot 多模型对比" : "Copilot 对话"
  );

  useEffect(() => {
    let active = true;
    Promise.allSettled([copilotApi.listThreads(), copilotApi.listModels()])
      .then(([threadsResult, modelsResult]) => {
        if (!active) return;
        const nextThreads = threadsResult.status === "fulfilled" ? threadsResult.value.threads : [];
        setThreads(nextThreads);
        setActiveThreadID(isNew ? null : nextThreads[0]?.id ?? null);
        if (threadsResult.status === "rejected") setError("暂时无法加载会话记录，你仍可新建会话");
        if (modelsResult.status === "fulfilled" && modelsResult.value.models.length > 0) {
          const nextModels = toDisplayModels(modelsResult.value.models);
          setAvailableModels(nextModels);
          setSelectedModel(nextModels.find((model) => model.selected)?.value ?? nextModels[0].value);
          setCompareModelValues(nextModels.filter((model) => model.selected).map((model) => model.value).slice(0, MAX_COMPARE_MODELS));
        }
      })
    return () => {
      active = false;
    };
  }, [isNew]);

  useEffect(() => {
    let active = true;
    void membershipApi.usage()
      .then((payload) => {
        if (active) setUsage(payload.usage ?? []);
      })
      .catch(() => {
        if (active) setUsage([]);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    setActivePopover(routePopover);
  }, [routePopover]);

  useEffect(() => {
    if (!activeThreadID) {
      setMessages([]);
      setCompareQuestion(null);
      setCompareAnswers([]);
      setCompareSummary(null);
      return;
    }
    let active = true;
    copilotApi
      .listMessages(activeThreadID)
      .then((payload) => {
        if (!active) return;
        setMessages((current) => {
          const optimisticMessages = current.filter((message) => message.id < 0 && message.thread_id === activeThreadID);
          return optimisticMessages.length > 0 ? [...payload.messages, ...optimisticMessages] : payload.messages;
        });
        if (isCompare) {
          const rebuilt = rebuildCompareState(payload.messages);
          setCompareQuestion(rebuilt.question);
          setCompareAnswers(rebuilt.answers);
          setCompareSummary(rebuilt.summary);
        }
      })
      .catch(() => {
        if (active) setError("暂时无法加载会话消息");
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
    if (activePopover !== "files") return;
    let active = true;
    copilotApi
      .listFiles()
      .then((payload) => {
        if (!active) return;
        setFiles(payload.files);
        setSelectedReferenceIDs((current) => current.filter((fileID) => payload.files.some((file) => file.id === fileID)));
      })
      .catch(() => {
        if (active) setError("暂时无法加载引用文件");
      });
    return () => {
      active = false;
    };
  }, [activePopover]);

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

  useEffect(() => {
    const timers = typewriterTimersRef.current;
    return () => {
      timers.forEach((timer) => window.clearInterval(timer));
      timers.clear();
    };
  }, []);

  useEffect(() => {
    if (!activePopover) return;
    function handlePointerDown(event: PointerEvent) {
      const composerWrap = composerWrapRef.current;
      if (composerWrap?.contains(event.target as Node)) return;
      setActivePopover(null);
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setActivePopover(null);
    }
    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [activePopover]);

  function startTypewriter(message: CopilotMessage) {
    const characters = Array.from(message.content);
    if (characters.length === 0) return;
    const messageID = message.id;
    const existingTimer = typewriterTimersRef.current.get(messageID);
    if (existingTimer) window.clearInterval(existingTimer);

    let index = 1;
    setTypewriterContent((current) => ({ ...current, [messageID]: characters.slice(0, index).join("") }));
    const timer = window.setInterval(() => {
      index += 1;
      if (index >= characters.length) {
        window.clearInterval(timer);
        typewriterTimersRef.current.delete(messageID);
        setTypewriterContent((current) => {
          const next = { ...current };
          delete next[messageID];
          return next;
        });
        return;
      }
      setTypewriterContent((current) => ({ ...current, [messageID]: characters.slice(0, index).join("") }));
    }, 24);
    typewriterTimersRef.current.set(messageID, timer);
  }

  async function startThread(initialContent?: string) {
    const title = (initialContent || "新会话").slice(0, 28);
    const thread = await copilotApi.createThread({ title, mode: isCompare ? "compare" : "chat", model: selectedModel });
    setThreads((current) => [thread, ...current]);
    setActiveThreadID(thread.id);
    setMessages([]);
    return thread;
  }

  async function replaceMissingThread(threadID: number, initialContent: string) {
    setThreads((current) => current.filter((thread) => thread.id !== threadID));
    setActiveThreadID((current) => current === threadID ? null : current);
    return startThread(initialContent);
  }

  async function handleSend(event: FormEvent) {
    event.preventDefault();
    const content = draft.trim();
    if (!content || isSending) return;
    const abortController = new AbortController();
    sendAbortRef.current = abortController;
    setDraft("");
    setStreamingContent("");
    setError("");
    setIsSending(true);
    let streamAccepted = false;
    let streamedAssistant: CopilotMessage | null = null;
    let streamedContent = "";
    let sentThreadID: number | null = activeThreadID;
    let sentUserMessage: CopilotMessage | null = null;
    try {
      let thread = activeThread ?? await startThread(content);
      sentThreadID = thread.id;
      if (isCompare) {
        const models = compareModelValues.length > 0 ? compareModelValues : [selectedModel];
        const showOptimisticQuestion = (targetThread: CopilotThread) => {
          setCompareQuestion(optimisticMessage({
            content,
            model: compareModelValues.join(","),
            role: "user",
            threadID: targetThread.id,
            userID: targetThread.user_id,
            metadata: { kind: "compare_question" }
          }));
          setCompareAnswers([]);
          setCompareSummary(null);
        };
        const sendComparison = (targetThread: CopilotThread) => copilotApi.compareMessages(
          targetThread.id,
          { content, models, request_id: newRequestID("compare") },
          abortController.signal
        );
        showOptimisticQuestion(thread);
        let result: Awaited<ReturnType<typeof copilotApi.compareMessages>>;
        try {
          result = await sendComparison(thread);
        } catch (requestError) {
          if (!isThreadNotFoundError(requestError)) throw requestError;
          thread = await replaceMissingThread(thread.id, content);
          showOptimisticQuestion(thread);
          result = await sendComparison(thread);
        }
        if (abortController.signal.aborted) return;
        setCompareQuestion(result.user_message);
        setCompareAnswers(result.answers);
        setCompareSummary(null);
        setThreads((current) => current.map((item) => item.id === thread.id ? { ...item, updated_at: result.user_message.created_at } : item));
        await refreshUsage();
        return;
      }
      const messageInput = {
        content,
        model: selectedModel,
        reference_ids: selectedReferenceIDs,
        request_id: newRequestID("message")
      };
      let optimisticUserMessage = optimisticMessage({
        content,
        model: selectedModel,
        role: "user",
        threadID: thread.id,
        userID: thread.user_id
      });
      setMessages((current) => [...current, optimisticUserMessage]);
      sentUserMessage = optimisticUserMessage;
      const sendChatMessage = async (targetThread: CopilotThread, optimisticMessageID: number) => {
        let usedFallback = false;
        let result: SendMessageResult;
        try {
          result = await copilotApi.streamMessage(targetThread.id, messageInput, {
            onUserMessage: (message) => {
              streamAccepted = true;
              setMessages((current) => [...current.filter((item) => item.id !== optimisticMessageID && item.id !== message.id), message]);
            },
            onDelta: (delta) => {
              streamedContent += delta;
              setStreamingContent((current) => current + delta);
            },
            onAssistantMessage: (message) => {
              streamAccepted = true;
              streamedAssistant = message;
            }
          }, abortController.signal);
        } catch (streamError) {
          const canFallback = streamError instanceof ApiRequestError && streamError.code !== "thread_not_found" && (
            streamError.code === "streaming_not_supported" || streamError.status === 404 || streamError.status === 501
          );
          if (!canFallback) throw streamError;
          usedFallback = true;
          result = await copilotApi.sendMessage(targetThread.id, messageInput, abortController.signal);
        }
        return { result, usedFallback };
      };
      let sendResult: { result: SendMessageResult; usedFallback: boolean };
      try {
        sendResult = await sendChatMessage(thread, optimisticUserMessage.id);
      } catch (requestError) {
        if (!isThreadNotFoundError(requestError)) throw requestError;
        setStreamingContent("");
        setMessages((current) => current.filter((message) => message.id !== optimisticUserMessage.id));
        thread = await replaceMissingThread(thread.id, content);
        sentThreadID = thread.id;
        optimisticUserMessage = optimisticMessage({
          content,
          model: selectedModel,
          role: "user",
          threadID: thread.id,
          userID: thread.user_id
        });
        sentUserMessage = optimisticUserMessage;
        setMessages((current) => [...current, optimisticUserMessage]);
        streamAccepted = false;
        streamedAssistant = null;
        streamedContent = "";
        sendResult = await sendChatMessage(thread, optimisticUserMessage.id);
      }
      if (abortController.signal.aborted) return;
      const { result, usedFallback } = sendResult;
      setMessages((current) => [
        ...current.filter((message) => message.id !== optimisticUserMessage.id && message.id !== result.user_message.id),
        result.user_message,
        result.assistant_message
      ]);
      setStreamingContent("");
      if (usedFallback) startTypewriter(result.assistant_message);
      setThreads((current) => current.map((item) => item.id === thread.id ? { ...item, updated_at: result.assistant_message.created_at } : item));
      await refreshUsage();
    } catch (requestError) {
      if (abortController.signal.aborted) {
        setMessages((current) => {
          const next = current.filter((message) => message.id >= 0);
          const targetThreadID = sentThreadID ?? sentUserMessage?.thread_id;
          const hasServerUser = next.some((message) => message.role === "user" && message.thread_id === targetThreadID && message.content === content);
          if (!hasServerUser && sentUserMessage && (streamAccepted || streamedContent.trim())) next.push(sentUserMessage);
          if (streamedAssistant && !next.some((message) => message.id === streamedAssistant?.id)) next.push(streamedAssistant);
          else if (streamedContent.trim() && targetThreadID) next.push({
            id: -Date.now(),
            user_id: sentUserMessage?.user_id ?? 0,
            thread_id: targetThreadID,
            role: "assistant",
            content: streamedContent.trim(),
            status: "failed",
            model: selectedModel,
            created_at: new Date().toISOString()
          });
          return next;
        });
        setStreamingContent("");
        if (isCompare) setCompareQuestion((current) => current && current.id < 0 ? null : current);
        setError("已暂停本次对话");
        return;
      }
      if (!streamAccepted && !streamedAssistant && !streamedContent.trim()) {
        setDraft(content);
      } else {
        setDraft("");
      }
      setStreamingContent("");
      setMessages((current) => {
        const next = current.filter((message) => message.id >= 0);
        const targetThreadID = sentThreadID ?? sentUserMessage?.thread_id;
        const hasServerUser = next.some((message) => message.role === "user" && message.thread_id === targetThreadID && message.content === content);
        if (!hasServerUser && sentUserMessage && (streamAccepted || streamedContent.trim())) next.push(sentUserMessage);
        if (streamedAssistant && !next.some((message) => message.id === streamedAssistant?.id)) {
          next.push(streamedAssistant);
        } else if (streamedContent.trim() && targetThreadID && !next.some((message) => message.role === "assistant" && message.thread_id === targetThreadID && message.content === streamedContent.trim())) {
          next.push({
            id: -Date.now(),
            user_id: sentUserMessage?.user_id ?? 0,
            thread_id: targetThreadID,
            role: "assistant",
            content: streamedContent.trim(),
            status: "failed",
            model: selectedModel,
            created_at: new Date().toISOString()
          });
        }
        return next;
      });
      if (isCompare) {
        setCompareQuestion((current) => current && current.id < 0 ? null : current);
      }
      setError(requestError instanceof Error ? requestError.message : "发送失败，请稍后重试");
    } finally {
      if (sendAbortRef.current === abortController) sendAbortRef.current = null;
      setIsSending(false);
    }
  }

  function handlePauseConversation() {
    sendAbortRef.current?.abort();
    typewriterTimersRef.current.forEach((timer) => window.clearInterval(timer));
    typewriterTimersRef.current.clear();
    setTypewriterContent({});
    setIsSending(false);
    setError("已暂停本次对话");
  }

	async function handleToolDecision(message: CopilotMessage, decision: "confirm" | "cancel") {
	  if (!activeThreadID || toolBusyID !== null) return;
	  setToolBusyID(message.id);
	  setError("");
	  try {
	    const result = await copilotApi.confirmTool(activeThreadID, message.id, decision);
	    setMessages((current) => current.map((item) => item.id === result.message.id ? result.message : item));
	  } catch (requestError) {
	    setError(requestError instanceof Error ? requestError.message : "操作失败，请稍后重试");
	  } finally {
	    setToolBusyID(null);
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

  async function refreshUsage() {
    const payload = await membershipApi.usage().catch(() => null);
    if (payload) setUsage(payload.usage ?? []);
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

  async function handleUpdateMemory(id: number, input: { key?: string; value?: string; status?: CopilotMemory["status"] }) {
    setError("");
    try {
      const memory = await copilotApi.updateMemory(id, input);
      setMemories((current) => current.map((item) => item.id === id ? memory : item));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "更新记忆失败，请稍后重试");
    }
  }

  async function handleSaveFile(input: { name: string; content: string; mime_type?: string }) {
    setIsSavingFile(true);
    setError("");
    try {
      const file = await copilotApi.saveFile({ name: input.name, mime_type: input.mime_type || "text/plain", content: input.content });
      setFiles((current) => [file, ...current.filter((item) => item.id !== file.id)]);
      setSelectedReferenceIDs((current) => current.includes(file.id) ? current : [...current, file.id]);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "保存文件失败");
    } finally {
      setIsSavingFile(false);
    }
  }

  async function handleUploadFile(upload: File) {
    setIsSavingFile(true);
    setError("");
    try {
      const file = await copilotApi.uploadFile(upload);
      setFiles((current) => [file, ...current.filter((item) => item.id !== file.id)]);
      setSelectedReferenceIDs((current) => current.includes(file.id) ? current : [...current, file.id]);
      await refreshUsage();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "上传文件失败");
    } finally {
      setIsSavingFile(false);
    }
  }

  async function handleDeleteFile(id: number) {
    try {
      await copilotApi.deleteFile(id);
      setFiles((current) => current.filter((file) => file.id !== id));
      setSelectedReferenceIDs((current) => current.filter((fileID) => fileID !== id));
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "删除文件失败");
    }
  }

  function handleToggleReference(id: number) {
    setSelectedReferenceIDs((current) => (
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id]
    ));
  }

  function handleToggleCompareModel(value: string) {
    setCompareModelValues((current) => {
      if (current.includes(value)) {
        if (current.length === 1) return current;
        return current.filter((item) => item !== value);
      }
      return [...current, value].slice(0, MAX_COMPARE_MODELS);
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
        answers: compareAnswers,
        request_id: newRequestID("summary")
      });
      setCompareSummary(result.summary_message);
      setThreads((current) => current.map((item) => item.id === activeThreadID ? { ...item, updated_at: result.summary_message.created_at } : item));
      await refreshUsage();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "总结失败，请稍后重试");
    } finally {
      setIsSummarizing(false);
    }
  }

  return (
    <V4PageShell className="copilot-shell" showCopilotMini={false}>
      <section
        className={"copilot-workbench " + (isCompare ? "compare-mode " : "") + (historyCollapsed ? "history-collapsed " : "") + (showRename || showDelete ? "modal-open" : "")}
        aria-label="智活 Copilot 工作台"
      >
        <div className="copilot-main-panel">
          {isCompare ? (
            <CompareHeader
              models={availableModels}
              selectedModels={compareModelValues}
              onToggleModel={handleToggleCompareModel}
            />
          ) : <CopilotHeader onPrompt={setDraft} />}

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
            ) : (
	              <ChatThread
                messages={messages}
                isSending={isSending}
                streamingContent={streamingContent}
                typewriterContent={typewriterContent}
	                onPrompt={(prompt) => setDraft(prompt)}
	                onToolDecision={handleToolDecision}
	                toolBusyID={toolBusyID}
	              />
            )}
          </div>

          <div className={activeQuota.blocked ? "copilot-quota-inline depleted" : "copilot-quota-inline"}>
            <span>{activeQuota.label} {activeQuota.value}</span>
            <small>{activeQuota.unit}</small>
            {activeQuota.blocked && <Link to="/membership">升级套餐</Link>}
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
            onPause={handlePauseConversation}
            activePopover={activePopover}
            onPopoverChange={setActivePopover}
            ref={composerWrapRef}
            memories={memories}
            files={files}
            selectedReferenceIDs={selectedReferenceIDs}
            isSavingMemory={isSavingMemory}
            isSavingFile={isSavingFile}
            onSaveMemory={handleSaveMemory}
            onUpdateMemory={handleUpdateMemory}
            onDeleteMemory={handleDeleteMemory}
            onSaveFile={handleSaveFile}
            onUploadFile={handleUploadFile}
            onDeleteFile={handleDeleteFile}
            onToggleReference={handleToggleReference}
            onInsertReferences={() => setActivePopover(null)}
            compare={isCompare}
          />
        </div>

        <ConversationSidebar
          activeThreadID={activeThreadID}
          collapsed={historyCollapsed}
          threads={threads}
          showThreadMenu={showModelPicker}
          onToggleCollapsed={() => setHistoryCollapsed((current) => !current)}
          onSelectThread={setActiveThreadID}
          onNewThread={handleNewThread}
        />
        {showRename && <RenameDialog title={activeThread?.title ?? "智能客服系统项目机会分析"} onConfirm={handleRenameThread} />}
        {showDelete && <DeleteDialog onConfirm={handleArchiveThread} />}
      </section>
    </V4PageShell>
  );
}

function CopilotHeader({ onPrompt }: { onPrompt: (prompt: string) => void }) {
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
        {quickActions.map(([icon, label, prompt]) => (
          <button key={label} onClick={() => onPrompt(prompt)} type="button">
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
  const modelByValue = new Map(models.map((model) => [model.value, model]));
  const selectedModelOptions = selectedModels
    .map((value) => modelByValue.get(value))
    .filter((model): model is CopilotModel => Boolean(model));
  const displayedModels = [
    ...selectedModelOptions,
    ...models.filter((model) => !selectedModels.includes(model.value)).slice(0, Math.max(MAX_COMPARE_MODELS - selectedModelOptions.length, 0))
  ];

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
        <button className={selectedCount === 2 ? "active" : ""} onClick={() => selectedCount > 2 && onToggleModel(selectedModels[selectedModels.length - 1])} type="button">2 模型</button>
        <button
          className={selectedCount === 3 ? "active" : ""}
          onClick={() => {
            const nextModel = models.find((model) => !selectedModels.includes(model.value));
            if (selectedCount < 3 && nextModel) onToggleModel(nextModel.value);
          }}
          type="button"
        >
          3 模型
        </button>
        {displayedModels.map((model) => {
          const selected = selectedModels.includes(model.value);
          const disabled = !selected && selectedModels.length >= MAX_COMPARE_MODELS;
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
        <button className="compare-add-model" type="button">＋ 选择模型</button>
      </div>
    </header>
  );
}

function ChatThread({
  messages,
  isSending,
  streamingContent,
  typewriterContent,
  onPrompt,
  onToolDecision,
  toolBusyID
}: {
  messages: CopilotMessage[];
  isSending: boolean;
  streamingContent: string;
  typewriterContent?: Record<number, string>;
  onPrompt: (prompt: string) => void;
  onToolDecision: (message: CopilotMessage, decision: "confirm" | "cancel") => void;
  toolBusyID: number | null;
}) {
  if (messages.length > 0) {
    return (
      <div className="copilot-chat-thread" aria-label="会话内容">
        {messages.map((message) => (
          <article key={message.id} className={`copilot-message ${message.role === "user" ? "user" : "assistant"} ${message.metadata?.kind === "reference_report" ? "report-continuation" : ""}`}>
            {message.role !== "user" && message.metadata?.kind !== "reference_report" && <span className="v4-logo" aria-hidden="true" />}
            {message.role === "user" ? (
              <UserMessageBubble content={message.content} time={formatTime(message.created_at)} />
            ) : message.metadata?.kind === "reference_report" ? (
              <ReferenceAnalysisReport time={formatTime(message.created_at)} />
            ) : message.metadata?.kind === "reference_thinking" ? (
              <div className="copilot-thinking">
                正在思考中
                <span /><span /><span /><span />
              </div>
            ) : (
              <div className={"copilot-bubble compact " + (typewriterContent && message.id in typewriterContent ? "typing" : "")}>
                <MarkdownMessage content={typewriterContent?.[message.id] ?? message.content} />
                {message.metadata?.tool_preview?.status === "pending" && (
                  <ToolPreviewCard
                    preview={message.metadata.tool_preview}
                    busy={toolBusyID === message.id}
                    onDecision={(decision) => onToolDecision(message, decision)}
                  />
                )}
                {message.metadata?.tool_result && <ToolResultCard result={message.metadata.tool_result} />}
                <time>{formatTime(message.created_at)}</time>
              </div>
            )}
            {message.role === "user" && <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>}
          </article>
        ))}
        {isSending && (streamingContent ? <StreamingMessage content={streamingContent} /> : <ThinkingMessage />)}
      </div>
    );
  }

  return isSending ? (
    <div className="copilot-chat-thread" aria-label="会话内容">
      {streamingContent ? <StreamingMessage content={streamingContent} /> : <ThinkingMessage />}
    </div>
  ) : <EmptyConversation onPrompt={onPrompt} />;
}

function ToolPreviewCard({
  preview,
  busy,
  onDecision
}: {
  preview: NonNullable<NonNullable<CopilotMessage["metadata"]>["tool_preview"]>;
  busy: boolean;
  onDecision: (decision: "confirm" | "cancel") => void;
}) {
  const isTask = preview.call.tool === "create_task";
  const title = isTask ? preview.call.arguments.title : preview.call.arguments.intent;
  return (
    <div className="copilot-tool-preview" role="group" aria-label="待确认的 Copilot 操作">
      <strong>{isTask ? "准备创建任务" : "准备发起项目匹配"}</strong>
      <p>{title || "未命名操作"}</p>
      {isTask && <small>确认后进入正式任务统计，并按任务权限处理通知。</small>}
      {!isTask && <small>确认后创建正式项目匹配会话。</small>}
      <div>
        <button disabled={busy} onClick={() => onDecision("cancel")} type="button">取消</button>
        <button disabled={busy} onClick={() => onDecision("confirm")} type="button">{busy ? "执行中..." : "确认执行"}</button>
      </div>
    </div>
  );
}

function ToolResultCard({ result }: { result: NonNullable<NonNullable<CopilotMessage["metadata"]>["tool_result"]> }) {
  const action = result.tool === "create_task" ? "查看任务" : "查看匹配";
  return (
    <div className="copilot-tool-result">
      <span className="copilot-ui-icon check" aria-hidden="true" />
      <div>
        <strong>{result.title || result.message}</strong>
        <small>{result.status === "completed" ? "执行完成" : "等待补充信息"}</small>
      </div>
      <Link to={result.url}>{action}</Link>
    </div>
  );
}

function StreamingMessage({ content }: { content: string }) {
  return (
    <article className="copilot-message assistant streaming">
      <span className="v4-logo" aria-hidden="true" />
      <div className="copilot-bubble compact typing"><MarkdownMessage content={content} /></div>
    </article>
  );
}

function UserMessageBubble({ content, time }: { content: string; time: string }) {
  return (
    <div className="user-message-bubble">
      <p>{content}</p>
      <time>{time}</time>
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

function ReferenceAnalysisReport({ time }: { time: string }) {
  return (
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
      <time>{time}</time>
    </div>
  );
}

function EmptyConversation({ onPrompt }: { onPrompt: (prompt: string) => void }) {
  return (
    <div className="copilot-empty-state" aria-label="会话内容">
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
  const hasCompareContent = Boolean(question || hasBackendAnswers || summary || isSending);
  const answerCountClass = `answer-count-${Math.min(Math.max(answers.length, 1), MAX_COMPARE_MODELS)}`;

  if (!hasCompareContent) {
    return (
      <div className="copilot-comparison empty" aria-label="对比分析内容">
        <div className="comparison-empty-card">
          <span className="copilot-ui-icon trend" aria-hidden="true" />
          <h2>开始 AI 对比分析</h2>
          <p>选择模型后发送问题，回答会在这里按模型并排展示。</p>
        </div>
      </div>
    );
  }

  return (
    <div className="copilot-comparison">
      {question && (
        <article className="copilot-message user compare-user-question">
          <UserMessageBubble content={question.content} time={formatTime(question.created_at)} />
          <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
        </article>
      )}
      {isSending && <ThinkingMessage />}
      <div className={`comparison-grid ${answerCountClass}`}>
        {answers.map((answer) => (
          <article key={answer.model} className="comparison-card">
            <header>
              <span className={"model-glyph " + modelIcon(answer.model, 0)} aria-hidden="true" />
              <h2>{modelLabel(answer.model, models)}</h2>
              <b>{answer.error_code ? "回答失败" : "回答完成"}</b>
              <time>{formatTime(answer.assistant_message.created_at)}</time>
            </header>
            <section>
              <MarkdownMessage content={answer.assistant_message.content} />
            </section>
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
            <MarkdownMessage content={summary.content} />
          </section>
        </article>
      )}
      {hasBackendAnswers && (
        <button className="compare-summary-button" disabled={isSummarizing} onClick={onSummarize} type="button">
          <span className="copilot-ui-icon doc" aria-hidden="true" />
          {isSummarizing ? "正在总结" : "总结本次对比分析"}
        </button>
      )}
    </div>
  );
}

const Composer = forwardRef<HTMLDivElement, {
  draft: string;
  model: string;
  models: CopilotModel[];
  memories: CopilotMemory[];
  files: CopilotFile[];
  selectedReferenceIDs: number[];
  isSending: boolean;
  isSavingMemory: boolean;
  isSavingFile: boolean;
  error: string;
  onDraftChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onSaveMemory: (input: { key: string; value: string }) => void;
  onUpdateMemory: (id: number, input: { key?: string; value?: string; status?: CopilotMemory["status"] }) => Promise<void>;
  onDeleteMemory: (id: number) => void;
  onSaveFile: (input: { name: string; content: string; mime_type?: string }) => Promise<void>;
  onUploadFile: (file: File) => Promise<void>;
  onDeleteFile: (id: number) => Promise<void>;
  onToggleReference: (id: number) => void;
  onInsertReferences: () => void;
  onSubmit: (event: FormEvent) => void;
  onPause: () => void;
  activePopover?: ComposerPopover;
  onPopoverChange: (popover: ComposerPopover) => void;
  compare?: boolean;
}>(function Composer({
  draft,
  model,
  models,
  memories,
  files,
  selectedReferenceIDs,
  isSending,
  isSavingMemory,
  isSavingFile,
  error,
  onDraftChange,
  onModelChange,
  onSaveMemory,
  onUpdateMemory,
  onDeleteMemory,
  onSaveFile,
  onUploadFile,
  onDeleteFile,
  onToggleReference,
  onInsertReferences,
  onSubmit,
  onPause,
  activePopover,
  onPopoverChange,
  compare
}, ref) {
  const selectedModelLabel = models.find((item) => item.value === model)?.name ?? model;
  const [deepThinking, setDeepThinking] = useState(false);
  const [isListening, setIsListening] = useState(false);
  const [voiceStatus, setVoiceStatus] = useState("");
  const recognitionRef = useRef<SpeechRecognitionLike | null>(null);
  const showModelPicker = activePopover === "models";
  const showReferencePicker = activePopover === "files";
  const showMemoryPanel = activePopover === "memories";
  const selectedReferenceFiles = useMemo(
    () => files.filter((file) => selectedReferenceIDs.includes(file.id)),
    [files, selectedReferenceIDs]
  );

  useEffect(() => {
    return () => {
      recognitionRef.current?.stop();
    };
  }, []);

  function handleVoiceDraft() {
    if (isListening) {
      recognitionRef.current?.stop();
      setIsListening(false);
      setVoiceStatus("语音输入已暂停");
      return;
    }
    const SpeechRecognition = (
      window as Window & { SpeechRecognition?: SpeechRecognitionConstructor; webkitSpeechRecognition?: SpeechRecognitionConstructor }
    ).SpeechRecognition ?? (
      window as Window & { SpeechRecognition?: SpeechRecognitionConstructor; webkitSpeechRecognition?: SpeechRecognitionConstructor }
    ).webkitSpeechRecognition;
    if (!SpeechRecognition) {
      setVoiceStatus("当前浏览器不支持语音输入");
      return;
    }
    const recognition = new SpeechRecognition();
    recognition.lang = "zh-CN";
    recognition.continuous = true;
    recognition.interimResults = true;
    recognition.onresult = (event) => {
      let transcript = "";
      for (let index = event.resultIndex; index < event.results.length; index += 1) {
        transcript += event.results[index][0].transcript;
      }
      const nextDraft = `${draft.trim() ? `${draft.trim()} ` : ""}${transcript.trim()}`.trim();
      onDraftChange(nextDraft);
      setVoiceStatus(transcript.trim() ? "正在识别语音" : "请开始说话");
    };
    recognition.onerror = () => {
      setIsListening(false);
      setVoiceStatus("语音识别失败，请重试");
    };
    recognition.onend = () => {
      setIsListening(false);
    };
    recognitionRef.current = recognition;
    setIsListening(true);
    setVoiceStatus("请开始说话");
    recognition.start();
  }

  return (
    <div className="copilot-composer-wrap" ref={ref}>
      {showModelPicker && <ModelPicker models={models} selectedModel={model} onSelect={onModelChange} />}
      {showReferencePicker && (
        <ReferencePicker
          files={files}
          isSavingFile={isSavingFile}
          selectedIDs={selectedReferenceIDs}
          onInsert={onInsertReferences}
          onSaveFile={onSaveFile}
          onUploadFile={onUploadFile}
          onDeleteFile={onDeleteFile}
          onToggle={onToggleReference}
        />
      )}
      {showMemoryPanel && (
        <MemoryPanel
          isSaving={isSavingMemory}
          memories={memories}
          onDelete={onDeleteMemory}
          onSave={onSaveMemory}
          onUpdate={onUpdateMemory}
        />
      )}
      {selectedReferenceFiles.length > 0 && (
        <div className="selected-reference-bar" role="status" aria-label="已插入引用">
          <span>已引用</span>
          {selectedReferenceFiles.map((file) => (
            <button
              aria-label={`移除引用 ${file.name}`}
              key={file.id}
              onClick={() => onToggleReference(file.id)}
              type="button"
            >
              <span className="reference-file sheet" aria-hidden="true" />
              <strong>{file.name}</strong>
              <em aria-hidden="true">×</em>
            </button>
          ))}
        </div>
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
          <Link onClick={() => onPopoverChange("files")} to="/copilot/files">
            <span className="copilot-ui-icon clip" aria-hidden="true" />
            上传文件
          </Link>
          <Link className={showReferencePicker ? "active" : ""} onClick={() => onPopoverChange("files")} to="/copilot/files">
            <span className="copilot-ui-icon link" aria-hidden="true" />
            引用
          </Link>
          <Link className={showMemoryPanel ? "active" : ""} onClick={() => onPopoverChange("memories")} to="/copilot/memories">
            <span className="copilot-ui-icon memory" aria-hidden="true" />
            记忆
          </Link>
          <Link
            aria-disabled={compare ? "true" : undefined}
            className={showModelPicker ? "active" : ""}
            onClick={(event) => {
              if (compare) event.preventDefault();
              else onPopoverChange("models");
            }}
            to="/copilot/models"
          >
            <span className="model-glyph swirl" aria-hidden="true" />
            {selectedModelLabel}
            <span aria-hidden="true">{showModelPicker ? "⌃" : "⌄"}</span>
          </Link>
          <Link className={"compare-chip " + (compare ? "active" : "")} to="/copilot/compare">
            <span className="copilot-ui-icon doc" aria-hidden="true" />
            AI 对比分析
            <b>New</b>
          </Link>
          <button aria-pressed={deepThinking} className={deepThinking ? "active" : ""} onClick={() => setDeepThinking((current) => !current)} type="button">
            <span className="copilot-ui-icon think" aria-hidden="true" />
            深度思考
            <b className="vip">VIP</b>
          </button>
          <span className="composer-spacer" />
          <button aria-label={isListening ? "暂停语音输入" : "语音输入"} className={"icon-only voice-button " + (isListening ? "listening" : "")} onClick={handleVoiceDraft} type="button">
            <span className="mic-icon" aria-hidden="true" />
          </button>
          {isSending ? (
            <button aria-label="暂停对话" className="send-button pause-button" onClick={onPause} type="button">
              <span aria-hidden="true">Ⅱ</span>
            </button>
          ) : (
            <button aria-label="发送" className="send-button" disabled={!draft.trim()} type="submit">
              <span aria-hidden="true">↗</span>
            </button>
          )}
        </div>
      </form>
      {error && <p className="ai-disclaimer">{error}</p>}
      {voiceStatus && <p className="ai-disclaimer voice-status">{voiceStatus}</p>}
      <p className="ai-disclaimer">ⓘ 内容由 AI 生成，请注意甄别准确性</p>
    </div>
  );
});

function MemoryPanel({
  memories,
  isSaving,
  onSave,
  onDelete,
  onUpdate
}: {
  memories: CopilotMemory[];
  isSaving: boolean;
  onSave: (input: { key: string; value: string }) => void;
  onDelete: (id: number) => void;
  onUpdate: (id: number, input: { key?: string; value?: string; status?: CopilotMemory["status"] }) => Promise<void>;
}) {
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [editingID, setEditingID] = useState<number | null>(null);
  const [editingKey, setEditingKey] = useState("");
  const [editingValue, setEditingValue] = useState("");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const nextKey = key.trim();
    const nextValue = value.trim();
    if (!nextKey || !nextValue || isSaving) return;
    onSave({ key: nextKey, value: nextValue });
    setKey("");
    setValue("");
  }

  function startEditing(memory: CopilotMemory) {
    setEditingID(memory.id);
    setEditingKey(memory.key);
    setEditingValue(memory.value);
  }

  async function handleEditSubmit(event: FormEvent) {
    event.preventDefault();
    if (editingID === null || !editingKey.trim() || !editingValue.trim()) return;
    await onUpdate(editingID, { key: editingKey.trim(), value: editingValue.trim() });
    setEditingID(null);
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
        ) : memories.map((memory) => {
          const status = memory.status ?? "active";
          return (
          <article className={status === "inactive" ? "inactive" : status === "pending" ? "pending" : undefined} key={memory.id}>
            {editingID === memory.id ? (
              <form className="memory-edit-form" onSubmit={(event) => void handleEditSubmit(event)}>
                <input aria-label={`编辑记忆名称 ${memory.key}`} onChange={(event) => setEditingKey(event.target.value)} value={editingKey} />
                <input aria-label={`编辑记忆内容 ${memory.key}`} onChange={(event) => setEditingValue(event.target.value)} value={editingValue} />
                <button disabled={!editingKey.trim() || !editingValue.trim()} type="submit">保存</button>
                <button onClick={() => setEditingID(null)} type="button">取消</button>
              </form>
            ) : (
              <>
                <span>
                  <strong>{memory.key}</strong>
                  <small>{memory.value}</small>
                  <em>{status === "pending" ? "待确认" : status === "inactive" ? "已停用" : memory.source === "manual" ? "手动维护" : "已启用"}</em>
                </span>
                <div>
                  {status === "pending" && <button aria-label={`确认记忆 ${memory.key}`} onClick={() => void onUpdate(memory.id, { status: "active" })} type="button">确认</button>}
                  {status === "pending" && <button aria-label={`忽略记忆 ${memory.key}`} onClick={() => void onUpdate(memory.id, { status: "inactive" })} type="button">忽略</button>}
                  {status === "inactive" && <button aria-label={`启用记忆 ${memory.key}`} onClick={() => void onUpdate(memory.id, { status: "active" })} type="button">启用</button>}
                  <button aria-label={`编辑记忆 ${memory.key}`} onClick={() => startEditing(memory)} type="button">编辑</button>
                  <button aria-label={`删除记忆 ${memory.key}`} onClick={() => onDelete(memory.id)} type="button">删除</button>
                </div>
              </>
            )}
          </article>
          );
        })}
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

function ReferencePicker({
  files,
  isSavingFile,
  selectedIDs,
  onSaveFile,
  onUploadFile,
  onDeleteFile,
  onToggle,
  onInsert
}: {
  files: CopilotFile[];
  isSavingFile: boolean;
  selectedIDs: number[];
  onSaveFile: (input: { name: string; content: string; mime_type?: string }) => Promise<void>;
  onUploadFile: (file: File) => Promise<void>;
  onDeleteFile: (id: number) => Promise<void>;
  onToggle: (id: number) => void;
  onInsert: () => void;
}) {
  const [fileName, setFileName] = useState("");
  const [fileContent, setFileContent] = useState("");
  const [isDragging, setIsDragging] = useState(false);
  const [uploadStatus, setUploadStatus] = useState("");
  const hasBackendFiles = files.length > 0;

  function handleSave(event: FormEvent) {
    event.preventDefault();
    const name = fileName.trim();
    const content = fileContent.trim();
    if (!name || !content || isSavingFile) return;
    onSaveFile({ name, content });
    setFileName("");
    setFileContent("");
  }

  async function saveDroppedFiles(fileList: FileList | File[]) {
    const nextFiles = Array.from(fileList);
    if (nextFiles.length === 0 || isSavingFile) return;
    setUploadStatus(`正在上传 ${nextFiles.length} 个文件`);
    for (const file of nextFiles) {
      await onUploadFile(file);
    }
    setUploadStatus(`已上传 ${nextFiles.length} 个文件`);
  }

  function handleFileInput(event: ChangeEvent<HTMLInputElement>) {
    const nextFiles = event.target.files;
    if (!nextFiles) return;
    void saveDroppedFiles(nextFiles).finally(() => {
      event.target.value = "";
    });
  }

  function handleDrop(event: DragEvent<HTMLLabelElement>) {
    event.preventDefault();
    setIsDragging(false);
    void saveDroppedFiles(event.dataTransfer.files);
  }

  return (
    <div className="copilot-popover reference-picker" role="dialog" aria-label="引用">
      <h2>引用</h2>
      <details className="reference-upload-tools">
        <summary>上传或粘贴文件</summary>
        <label
          className={"reference-dropzone " + (isDragging ? "dragging" : "")}
          onDragEnter={(event) => {
            event.preventDefault();
            setIsDragging(true);
          }}
          onDragOver={(event) => event.preventDefault()}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
        >
          <input
            accept=".txt,.md,.markdown,.csv,.tsv,.json,.yaml,.yml,.xml,.html,.htm,.docx"
            aria-label="选择上传文件"
            multiple
            onChange={handleFileInput}
            type="file"
          />
          <span className="copilot-ui-icon clip" aria-hidden="true" />
          <strong>拖拽文件到这里，或点击选择</strong>
          <small>支持文本、Markdown、CSV、JSON、YAML、XML、HTML、DOCX，单个不超过 10MB</small>
        </label>
        {uploadStatus && <p className="reference-upload-status">{uploadStatus}</p>}
        <form className="reference-upload-form" onSubmit={handleSave}>
          <label>
            <span>文件名称</span>
            <input aria-label="文件名称" onChange={(event) => setFileName(event.target.value)} placeholder="例如：客户访谈纪要.txt" value={fileName} />
          </label>
          <label>
            <span>文件内容</span>
            <textarea aria-label="文件内容" onChange={(event) => setFileContent(event.target.value)} placeholder="粘贴需要 Copilot 引用的文本内容" value={fileContent} />
          </label>
          <button disabled={isSavingFile || !fileName.trim() || !fileContent.trim()} type="submit">{isSavingFile ? "保存中" : "保存文件"}</button>
        </form>
      </details>
      <label className="reference-search">
        <span aria-hidden="true">⌕</span>
        <input placeholder="搜索文件、对话或我的内容" />
      </label>
      <div className="reference-list">
        {hasBackendFiles ? (
          <div>
            {groupReferenceFiles(files).map(([group, groupedFiles]) => (
              <section key={group}>
                <h3>{group}</h3>
                {groupedFiles.map((file) => {
              const selected = selectedIDs.includes(file.id);
              const ready = file.status === "ready";
              const icon = file.mime_type === "application/pdf" ? "pdf" : file.mime_type === "conversation" ? "chat" : "sheet";
              return (
                <div key={file.id} className={`reference-list-item ${selected ? "checked" : ""}`}>
                  <label>
                  <input
                    aria-label={`引用 ${file.name}`}
                    checked={selected}
                    disabled={!ready}
                    onChange={() => onToggle(file.id)}
                    type="checkbox"
                  />
                  <span className={`reference-file ${icon}`} aria-hidden="true" />
                  <span>
                    <strong>{file.name}</strong>
                    <small>{referenceFileMeta(file)}</small>
                  </span>
                  <em>{file.source === "current" ? "当前" : file.source === "recent" ? "昨天" : file.source === "history" ? "2024-06-20 09:15" : "昨天"}</em>
                  </label>
                  <button aria-label={`删除文件 ${file.name}`} onClick={() => void onDeleteFile(file.id)} type="button">删除</button>
                </div>
              );
                })}
              </section>
            ))}
          </div>
        ) : (
          <p className="reference-empty">暂无可引用文件，上传后会自动选中。</p>
        )}
      </div>
      <footer>
        <span>已选择 {selectedIDs.length} 项</span>
        <button onClick={onInsert} type="button">插入引用</button>
      </footer>
    </div>
  );
}

function ConversationSidebar({
  activeThreadID,
  collapsed,
  threads,
  showThreadMenu,
  onToggleCollapsed,
  onSelectThread,
  onNewThread
}: {
  activeThreadID: number | null;
  collapsed: boolean;
  threads: CopilotThread[];
  showThreadMenu?: boolean;
  onToggleCollapsed: () => void;
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
  })) : [];

  return (
    <aside className={"copilot-history " + (collapsed ? "collapsed" : "")} aria-label="会话记录">
      <header>
        <button aria-label={collapsed ? "展开会话记录" : "收起会话记录"} onClick={onToggleCollapsed} type="button">
          {collapsed ? "«" : "»"}
        </button>
        {!collapsed && <button onClick={onNewThread} type="button">＋ 新建会话</button>}
      </header>
      {!collapsed && (
        <>
          <label className="history-search">
            <span aria-hidden="true">⌕</span>
            <input placeholder="搜索会话" />
          </label>
          <div className="history-list">
            {displayThreads.map((item, index) => {
              const group = referenceThreadGroup(index);
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
        </>
      )}
      {showThreadMenu && !collapsed && (
        <div className="thread-action-menu" role="menu" aria-label="会话操作">
          <Link role="menuitem" to="/copilot/rename">重命名</Link>
          <Link role="menuitem" to="/copilot/delete">删除</Link>
        </div>
      )}
    </aside>
  );
}

function groupReferenceFiles(files: CopilotFile[]): Array<[string, CopilotFile[]]> {
  const labels: Record<string, string> = {
    current: "当前页面",
    recent: "最近上传文件",
    history: "历史对话",
    content: "我的内容",
    upload: "最近上传文件",
    pasted: "最近上传文件"
  };
  const groups = new Map<string, CopilotFile[]>();
  files.forEach((file) => {
    const label = labels[file.source] ?? "最近上传文件";
    groups.set(label, [...(groups.get(label) ?? []), file]);
  });
  return Array.from(groups.entries());
}

function referenceFileMeta(file: CopilotFile) {
  if (file.mime_type === "application/pdf") return `PDF · ${formatFileSize(file.size_bytes)}`;
  if (file.mime_type.includes("sheet")) return `XLSX · ${formatFileSize(file.size_bytes)}`;
  if (file.mime_type === "conversation") return "对话";
  if (file.mime_type === "document") return "文档";
  return `${file.mime_type} · ${formatFileSize(file.size_bytes)} · ${file.status === "ready" ? "解析完成" : "解析失败"}`;
}

function referenceThreadGroup(index: number) {
  if (index < 3) return "今天";
  return "更早";
}

function formatTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false });
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function newRequestID(prefix: string) {
  const randomID = typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  return `${prefix}-${randomID}`;
}

function isThreadNotFoundError(error: unknown) {
  return error instanceof ApiRequestError && error.code === "thread_not_found";
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
