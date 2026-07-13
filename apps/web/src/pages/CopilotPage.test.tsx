import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { authSession } from "../lib/authSession";
import CopilotPage, { type CopilotVariant } from "./CopilotPage";

describe("CopilotPage", () => {
  beforeEach(() => {
    Object.defineProperty(HTMLElement.prototype, "scrollTo", {
      configurable: true,
      value: vi.fn()
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    authSession.clear();
    vi.restoreAllMocks();
  });

  function renderPage(variant: CopilotVariant = "home") {
    authSession.set({
      access_token: "access-token",
      access_token_expires_at: "2026-06-23T12:00:00Z",
      is_new_user: false,
      user: {
        id: 7,
        nickname: "张婧",
        phone: "",
        account: "zhangjing",
        status: "active"
      }
    });

    return render(
      <MemoryRouter>
        <CopilotPage variant={variant} />
      </MemoryRouter>
    );
  }

  it("renders the conversation workspace", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "智活 Copilot" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开始一段新的对话" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /市场规模与增长趋势/ })).not.toBeInTheDocument();
    expect(screen.queryByText("智能客服市场分析报告.pdf")).not.toBeInTheDocument();
    expect(screen.getByRole("complementary", { name: "会话记录" })).toBeInTheDocument();
  });

  it("shows backend copilot quota usage", async () => {
    mockCopilotBackend();

    renderPage();

    expect(await screen.findByText("Copilot 对话 28/30")).toBeInTheDocument();
  });

  it("collapses and expands the conversation history rail", () => {
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "收起会话记录" }));

    expect(screen.getByRole("button", { name: "展开会话记录" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "＋ 新建会话" })).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText("搜索会话")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "展开会话记录" }));

    expect(screen.getByRole("button", { name: "收起会话记录" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "＋ 新建会话" })).toBeInTheDocument();
  });

  it("renders a clean new conversation", () => {
    renderPage("new");

    expect(screen.getByRole("heading", { name: "开始一段新的对话" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "分析一个新项目机会" })).toBeInTheDocument();
    expect(screen.queryByText("请帮我分析智能客服系统的市场机会和竞争格局。")).not.toBeInTheDocument();
    expect(screen.queryByText("智能客服市场分析报告.pdf")).not.toBeInTheDocument();
  });

  it("fills the composer when a quick action is selected", () => {
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "分析项目机会" }));

    expect(screen.getByLabelText("输入你的问题")).toHaveValue("请帮我分析当前项目的市场机会、目标客户、竞争格局和落地风险。");
  });

  it("exposes real navigation for composer toolbar actions", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "上传文件" })).toHaveAttribute("href", "/copilot/files");
    expect(screen.getByRole("link", { name: "引用" })).toHaveAttribute("href", "/copilot/files");
    expect(screen.getByRole("link", { name: "记忆" })).toHaveAttribute("href", "/copilot/memories");
    expect(screen.getByRole("link", { name: /DeepSeek/ })).toHaveAttribute("href", "/copilot/models");
    expect(screen.getByRole("link", { name: /AI 对比分析/ })).toHaveAttribute("href", "/copilot/compare");
  });

  it("renders the model picker", () => {
    renderPage("models");

    expect(screen.getByRole("dialog", { name: "模型选择" })).toBeInTheDocument();
    expect(screen.getByText("Claude opus4.8")).toBeInTheDocument();
    expect(screen.getByText("Grok4.3")).toBeInTheDocument();
  });

  it("smoke tests the selected model and shows recent AI run diagnostics", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("models");

    expect(await screen.findByRole("region", { name: "模型联调" })).toBeInTheDocument();
    expect(await screen.findByText("provider_model_not_found")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "测试 DeepSeek" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/models/smoke",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ model: "deepseek", prompt: "ping" })
        })
      );
    });
    expect(await screen.findByText("pong")).toBeInTheDocument();
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/ai-runs?limit=20",
        expect.objectContaining({ method: "GET" })
      );
    });
  });

  it("renders the reference picker", async () => {
    mockCopilotBackend();
    renderPage("files");

    expect(screen.getByRole("dialog", { name: "引用" })).toBeInTheDocument();
    expect(await screen.findByText("智能客服竞品功能对比表.txt")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "插入引用" })).toBeInTheDocument();
  });

  it("sends selected reference files with chat messages", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("files");

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.click(await screen.findByRole("checkbox", { name: "引用 智能客服竞品功能对比表.txt" }));
    fireEvent.click(screen.getByRole("button", { name: "插入引用" }));
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "结合文件分析机会" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99/messages",
        expect.objectContaining({
          method: "POST",
          body: expect.stringMatching(/^\{"content":"结合文件分析机会","model":"deepseek","reference_ids":\[17\],"request_id":"message-.+"\}$/)
        })
      );
    });
  });

  it("shows selected reference files after inserting them", async () => {
    mockCopilotBackend();
    renderPage("files");

    fireEvent.click(await screen.findByRole("checkbox", { name: "引用 智能客服竞品功能对比表.txt" }));
    fireEvent.click(screen.getByRole("button", { name: "插入引用" }));

    expect(screen.queryByRole("dialog", { name: "引用" })).not.toBeInTheDocument();
    const selectedReferences = screen.getByRole("status", { name: "已插入引用" });
    expect(selectedReferences).toHaveTextContent("智能客服竞品功能对比表.txt");

    fireEvent.click(screen.getByRole("button", { name: "移除引用 智能客服竞品功能对比表.txt" }));

    expect(screen.queryByRole("status", { name: "已插入引用" })).not.toBeInTheDocument();
  });

  it("uploads pasted text as a copilot reference file", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("files");

    await screen.findByText("智能客服竞品功能对比表.txt");
    fireEvent.change(screen.getByLabelText("文件名称"), { target: { value: "新增访谈纪要.txt" } });
    fireEvent.change(screen.getByLabelText("文件内容"), { target: { value: "客户最关注响应速度和私域转化。" } });
    fireEvent.click(screen.getByRole("button", { name: "保存文件" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/files",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            name: "新增访谈纪要.txt",
            mime_type: "text/plain",
            content: "客户最关注响应速度和私域转化。"
          })
        })
      );
    });
    await waitFor(() => {
      expect(screen.getAllByText("新增访谈纪要.txt").length).toBeGreaterThan(0);
    });
  });

  it("uploads selected local files as copilot references", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("files");

    await screen.findByText("智能客服竞品功能对比表.txt");
    const file = new File(["客户提到预算和交付周期。"], "客户访谈.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("选择上传文件"), { target: { files: [file] } });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/files",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            name: "客户访谈.txt",
            mime_type: "text/plain",
            content: "客户提到预算和交付周期。"
          })
        })
      );
    });
    await waitFor(() => {
      expect(screen.getAllByText("客户访谈.txt").length).toBeGreaterThan(0);
    });
  });

  it("shows voice input status when speech recognition is unavailable", () => {
    renderPage();

    fireEvent.click(screen.getByRole("button", { name: "语音输入" }));

    expect(screen.getByText("当前浏览器不支持语音输入")).toBeInTheDocument();
  });

  it("renders and manages copilot memories", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("memories");

    expect(await screen.findByRole("dialog", { name: "记忆" })).toBeInTheDocument();
    expect(await screen.findByText("偏好风格")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("记忆键"), { target: { value: "行业" } });
    fireEvent.change(screen.getByLabelText("记忆内容"), { target: { value: "教培" } });
    fireEvent.click(screen.getByRole("button", { name: "保存记忆" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/memories",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ key: "行业", value: "教培", confidence: 1, source: "manual" })
        })
      );
    });
    expect(await screen.findByText("行业")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "删除记忆 行业" }));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/memories/8",
        expect.objectContaining({ method: "DELETE" })
      );
    });
  });

  it("closes composer popovers when clicking outside them", () => {
    renderPage("memories");

    expect(screen.getByRole("dialog", { name: "记忆" })).toBeInTheDocument();

    fireEvent.pointerDown(screen.getByLabelText("会话内容"));

    expect(screen.queryByRole("dialog", { name: "记忆" })).not.toBeInTheDocument();
  });

  it("closes composer popovers with Escape", () => {
    renderPage("memories");

    expect(screen.getByRole("dialog", { name: "记忆" })).toBeInTheDocument();

    fireEvent.keyDown(document, { key: "Escape" });

    expect(screen.queryByRole("dialog", { name: "记忆" })).not.toBeInTheDocument();
  });

  it("renders a clean comparison empty state before backend answers", () => {
    renderPage("compare");

    expect(screen.getByRole("heading", { name: "AI 对比分析" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "开始 AI 对比分析" })).toBeInTheDocument();
    expect(screen.queryAllByText("回答完成")).toHaveLength(0);
    expect(screen.queryByRole("button", { name: "总结本次对比分析" })).not.toBeInTheDocument();
  });

  it("submits a comparison request and renders backend model answers", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("compare");

    await screen.findByText("智能客服系统项目机会分析");
    const input = screen.getByLabelText("输入你的问题");
    fireEvent.change(input, { target: { value: "帮我对比三个模型的建议" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99/compare",
        expect.objectContaining({
          method: "POST",
          body: expect.stringMatching(/^\{"content":"帮我对比三个模型的建议","models":\["deepseek"\],"request_id":"compare-.+"\}$/)
        })
      );
    });
    expect(await screen.findByText("DeepSeek 建议先做教培客服场景验证。")).toBeInTheDocument();
    expect(screen.getAllByText("DeepSeek").length).toBeGreaterThan(0);
  });

  it("shows an optimistic user message while sending chat and scrolls to the latest message", async () => {
    const fetchMock = mockCopilotBackend({ delayMessage: true });
    renderPage();

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "马上显示的问题" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    expect(await screen.findByText("马上显示的问题")).toBeInTheDocument();
    expect(HTMLElement.prototype.scrollTo).toHaveBeenCalled();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99/messages",
        expect.objectContaining({ method: "POST" })
      );
    });
    expect(await screen.findByText("后端返回的聊天回复。")).toBeInTheDocument();
  });

  it("pauses an in-flight chat request", async () => {
    mockCopilotBackend({ delayMessage: true });
    renderPage();

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "需要暂停的问题" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));
    fireEvent.click(await screen.findByRole("button", { name: "暂停对话" }));

    expect(await screen.findByText("已暂停本次对话")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText("后端返回的聊天回复。")).not.toBeInTheDocument();
    });
  });

  it("renders new assistant chat replies with a typewriter effect", async () => {
    mockCopilotBackend();
    renderPage();

    await screen.findByText("智能客服系统项目机会分析");
    vi.useFakeTimers();
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "打字机测试" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(screen.getByText("后")).toBeInTheDocument();
    expect(screen.queryByText("后端返回的聊天回复。")).not.toBeInTheDocument();

    await act(async () => {
      vi.advanceTimersByTime(600);
    });

    expect(screen.getByText("后端返回的聊天回复。")).toBeInTheDocument();
  });

  it("uses selected compare models when submitting comparison", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("compare");

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.click(screen.getByRole("button", { name: "选择模型 Claude" }));
    fireEvent.click(screen.getByRole("button", { name: "取消选择 DeepSeek" }));

    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "对比模型能力" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99/compare",
        expect.objectContaining({
          method: "POST",
          body: expect.stringMatching(/^\{"content":"对比模型能力","models":\["claude"\],"request_id":"compare-.+"\}$/)
        })
      );
    });
  });

  it("limits comparison selection to three models", () => {
    renderPage("compare");

    fireEvent.click(screen.getByRole("button", { name: "选择模型 GPT-4o" }));
    fireEvent.click(screen.getByRole("button", { name: "选择模型 Claude opus4.8" }));

    expect(screen.getByRole("button", { name: "3 模型" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "选择模型 Grok4.3" })).not.toBeInTheDocument();
  });

  it("shows an optimistic comparison question while waiting for model answers", async () => {
    mockCopilotBackend({ delayCompare: true });
    renderPage("compare");

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "立即展示的对比问题" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));

    expect(await screen.findByText("立即展示的对比问题")).toBeInTheDocument();
  });

  it("summarizes comparison answers", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("compare");

    await screen.findByText("智能客服系统项目机会分析");
    fireEvent.change(screen.getByLabelText("输入你的问题"), { target: { value: "帮我对比三个模型的建议" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));
    expect(await screen.findByText("DeepSeek 建议先做教培客服场景验证。")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "总结本次对比分析" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99/compare/summary",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining("\"answers\"")
        })
      );
    });
    expect(await screen.findByText("综合建议：先验证教培客服场景，再整理可复制的获客路径。")).toBeInTheDocument();
  });

  it("rebuilds comparison layout from historical messages", async () => {
    mockCopilotBackend({ historicalCompare: true });
    renderPage("compare");

    expect(await screen.findByText("历史问题：对比三个模型怎么落地")).toBeInTheDocument();
    expect(await screen.findByText("历史 DeepSeek：先验证高频客服。")).toBeInTheDocument();
    expect(await screen.findByText("历史 GPT：补齐执行路径。")).toBeInTheDocument();
    expect(await screen.findByText("历史总结：先客服验证，再固化执行模板。")).toBeInTheDocument();
  });

  it("renders the rename dialog", () => {
    renderPage("rename");

    expect(screen.getByRole("dialog", { name: "重命名会话" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("智能客服系统项目机会分析")).toBeInTheDocument();
  });

  it("renders the delete confirmation", () => {
    renderPage("delete");

    expect(screen.getByRole("dialog", { name: "删除会话" })).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: "我已确认删除此会话" })).toBeInTheDocument();
  });

  it("renames the active backend thread", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("rename");

    const input = await screen.findByDisplayValue("智能客服系统项目机会分析");
    fireEvent.change(input, { target: { value: "改名后的会话" } });
    fireEvent.click(screen.getByRole("button", { name: "确认重命名" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99",
        expect.objectContaining({
          method: "PATCH",
          body: JSON.stringify({ title: "改名后的会话" })
        })
      );
    });
    expect(await screen.findByText("改名后的会话")).toBeInTheDocument();
  });

  it("archives the active backend thread after confirmation", async () => {
    const fetchMock = mockCopilotBackend();
    renderPage("delete");

    fireEvent.click(await screen.findByRole("checkbox", { name: "我已确认删除此会话" }));
    fireEvent.click(screen.getByRole("button", { name: "确认删除" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/copilot/threads/99",
        expect.objectContaining({ method: "DELETE" })
      );
    });
    await waitFor(() => expect(screen.queryByText("智能客服系统项目机会分析")).not.toBeInTheDocument());
  });
});

function mockCopilotBackend(options: { historicalCompare?: boolean; delayCompare?: boolean; delayMessage?: boolean } = {}) {
  return vi.spyOn(globalThis, "fetch").mockImplementation((input, init) => {
    const url = String(input);
    if (url === "/api/v1/copilot/threads?limit=20") {
      return Promise.resolve(new Response(JSON.stringify({
        threads: [{
          id: 99,
          user_id: 7,
          title: "智能客服系统项目机会分析",
          mode: "chat",
          model: "deepseek",
          created_at: "2026-07-01T09:00:00Z",
          updated_at: "2026-07-01T09:30:00Z"
        }]
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/models") {
      return Promise.resolve(new Response(JSON.stringify({
        models: [
          { name: "DeepSeek", value: "deepseek", provider: "openai-compatible", is_default: true },
          { name: "GPT-4o", value: "gpt-main", provider: "openai-responses", is_default: false },
          { name: "Claude", value: "claude", provider: "openai-compatible", is_default: false }
        ]
      }), { status: 200 }));
    }
    if (url === "/api/v1/membership/usage") {
      return Promise.resolve(new Response(JSON.stringify({
        usage: [
          { key: "copilot_messages", label: "Copilot 对话", used: 2, limit: 30, unit: "次/月" },
          { key: "copilot_compare_calls", label: "Copilot 多模型对比", used: 0, limit: 30, unit: "模型次/月" }
        ]
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/ai-runs?limit=20") {
      return Promise.resolve(new Response(JSON.stringify({
        runs: [{
          id: 51,
          user_id: 7,
          feature: "copilot.model_smoke",
          prompt_version: "copilot_model_smoke_v1",
          provider: "openai-compatible",
          model: "deepseek",
          status: "failed",
          error_code: "provider_model_not_found",
          error_message: "AI provider model was not found",
          input_tokens: 0,
          output_tokens: 0,
          latency_ms: 180,
          created_at: "2026-07-01T10:25:00Z",
          updated_at: "2026-07-01T10:25:01Z"
        }]
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/files?limit=50") {
      return Promise.resolve(new Response(JSON.stringify({
        files: [{
          id: 17,
          user_id: 7,
          name: "智能客服竞品功能对比表.txt",
          mime_type: "text/plain",
          size_bytes: 64,
          content: "小鹅通：私域工具强；有赞教育：交易能力强。",
          created_at: "2026-07-02T09:00:00Z",
          updated_at: "2026-07-02T09:00:00Z"
        }]
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/files" && init?.method === "POST") {
      const payload = JSON.parse(String(init.body)) as { name: string; mime_type?: string; content: string };
      return Promise.resolve(new Response(JSON.stringify({
        id: 18,
        user_id: 7,
        name: payload.name,
        mime_type: payload.mime_type || "text/plain",
        size_bytes: payload.content.length,
        content: payload.content,
        created_at: "2026-07-02T09:15:00Z",
        updated_at: "2026-07-02T09:15:00Z"
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/models/smoke" && init?.method === "POST") {
      return Promise.resolve(new Response(JSON.stringify({
        ok: true,
        model: "deepseek",
        reply: "pong",
        input_tokens: 1,
        output_tokens: 1
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/threads/99/messages?limit=50") {
      if (options.historicalCompare) {
        return Promise.resolve(new Response(JSON.stringify({ messages: [
          {
            id: 20,
            user_id: 7,
            thread_id: 99,
            role: "user",
            content: "历史问题：对比三个模型怎么落地",
            status: "completed",
            model: "deepseek,gpt-main",
            metadata: { kind: "compare_question" },
            created_at: "2026-07-01T08:50:00Z"
          },
          {
            id: 21,
            user_id: 7,
            thread_id: 99,
            role: "assistant",
            content: "历史 DeepSeek：先验证高频客服。",
            status: "completed",
            model: "deepseek",
            metadata: { kind: "compare_answer" },
            created_at: "2026-07-01T08:51:00Z"
          },
          {
            id: 22,
            user_id: 7,
            thread_id: 99,
            role: "assistant",
            content: "历史 GPT：补齐执行路径。",
            status: "completed",
            model: "gpt-main",
            metadata: { kind: "compare_answer" },
            created_at: "2026-07-01T08:52:00Z"
          },
          {
            id: 23,
            user_id: 7,
            thread_id: 99,
            role: "assistant",
            content: "历史总结：先客服验证，再固化执行模板。",
            status: "completed",
            model: "deepseek",
            metadata: { kind: "compare_summary" },
            created_at: "2026-07-01T08:53:00Z"
          }
        ] }), { status: 200 }));
      }
      return Promise.resolve(new Response(JSON.stringify({ messages: [] }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/threads/99/compare" && init?.method === "POST") {
      const response = new Response(JSON.stringify({
        user_message: {
          id: 10,
          user_id: 7,
          thread_id: 99,
          role: "user",
          content: "帮我对比三个模型的建议",
          status: "completed",
          model: "deepseek",
          created_at: "2026-07-01T09:50:00Z"
        },
        answers: [{
          model: "deepseek",
          assistant_message: {
            id: 11,
            user_id: 7,
            thread_id: 99,
            role: "assistant",
            content: "DeepSeek 建议先做教培客服场景验证。",
            status: "completed",
            model: "deepseek",
            created_at: "2026-07-01T09:51:00Z"
          }
        }]
      }), { status: 200 });
      return options.delayCompare ? new Promise((resolve) => setTimeout(() => resolve(response), 20)) : Promise.resolve(response);
    }
    if (url === "/api/v1/copilot/threads/99/messages" && init?.method === "POST") {
      const response = new Response(JSON.stringify({
        user_message: {
          id: 30,
          user_id: 7,
          thread_id: 99,
          role: "user",
          content: "马上显示的问题",
          status: "completed",
          model: "deepseek",
          created_at: "2026-07-01T09:55:00Z"
        },
        assistant_message: {
          id: 31,
          user_id: 7,
          thread_id: 99,
          role: "assistant",
          content: "后端返回的聊天回复。",
          status: "completed",
          model: "deepseek",
          created_at: "2026-07-01T09:56:00Z"
        }
      }), { status: 200 });
      return options.delayMessage ? delayedResponse(response, init?.signal) : Promise.resolve(response);
    }
    if (url === "/api/v1/copilot/threads/99/compare/summary" && init?.method === "POST") {
      return Promise.resolve(new Response(JSON.stringify({
        summary_message: {
          id: 12,
          user_id: 7,
          thread_id: 99,
          role: "assistant",
          content: "综合建议：先验证教培客服场景，再整理可复制的获客路径。",
          status: "completed",
          model: "deepseek",
          created_at: "2026-07-01T09:52:00Z"
        }
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/memories?limit=50") {
      return Promise.resolve(new Response(JSON.stringify({
        memories: [{
          id: 7,
          user_id: 7,
          key: "偏好风格",
          value: "直接给执行清单",
          confidence: 0.9,
          source: "manual",
          created_at: "2026-07-01T09:00:00Z",
          updated_at: "2026-07-01T09:00:00Z"
        }]
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/memories" && init?.method === "POST") {
      return Promise.resolve(new Response(JSON.stringify({
        id: 8,
        user_id: 7,
        key: "行业",
        value: "教培",
        confidence: 1,
        source: "manual",
        created_at: "2026-07-01T09:45:00Z",
        updated_at: "2026-07-01T09:45:00Z"
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/memories/8" && init?.method === "DELETE") {
      return Promise.resolve(new Response(null, { status: 204 }));
    }
    if (url === "/api/v1/copilot/threads/99" && init?.method === "PATCH") {
      return Promise.resolve(new Response(JSON.stringify({
        id: 99,
        user_id: 7,
        title: "改名后的会话",
        mode: "chat",
        model: "deepseek",
        created_at: "2026-07-01T09:00:00Z",
        updated_at: "2026-07-01T09:40:00Z"
      }), { status: 200 }));
    }
    if (url === "/api/v1/copilot/threads/99" && init?.method === "DELETE") {
      return Promise.resolve(new Response(null, { status: 204 }));
    }
    return Promise.reject(new Error(`unexpected request: ${url}`));
  });
}

function delayedResponse(response: Response, signal?: AbortSignal | null) {
  return new Promise<Response>((resolve, reject) => {
    if (signal?.aborted) {
      reject(new Error("aborted"));
      return;
    }
    const handleAbort = () => {
      clearTimeout(timeout);
      signal?.removeEventListener("abort", handleAbort);
      reject(new Error("aborted"));
    };
    const timeout = setTimeout(() => {
      signal?.removeEventListener("abort", handleAbort);
      resolve(response);
    }, 500);
    signal?.addEventListener("abort", handleAbort, { once: true });
  });
}
