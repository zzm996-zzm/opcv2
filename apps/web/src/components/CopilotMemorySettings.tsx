import { useEffect, useState, type FormEvent } from "react";
import { ChevronRight, Trash2 } from "lucide-react";
import { Link } from "react-router-dom";

import { copilotApi, type CopilotMemory } from "../lib/copilotApi";

type CopilotMemorySettingsProps = {
  limit?: number;
};

function CopilotMemorySettings({ limit = 4 }: CopilotMemorySettingsProps) {
  const [memories, setMemories] = useState<CopilotMemory[]>([]);
  const [memoryKey, setMemoryKey] = useState("");
  const [memoryValue, setMemoryValue] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [deletingID, setDeletingID] = useState<number | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    copilotApi.listMemories(limit)
      .then((payload) => {
        if (!active) return;
        setMemories(payload.memories);
        setError("");
      })
      .catch(() => {
        if (active) setError("暂时无法读取记忆");
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [limit]);

  async function saveMemory(event: FormEvent) {
    event.preventDefault();
    const key = memoryKey.trim();
    const value = memoryValue.trim();
    if (!key || !value || isSaving) return;
    setIsSaving(true);
    setError("");
    try {
      const memory = await copilotApi.saveMemory({ key, value, confidence: 1, source: "manual" });
      setMemories((current) => [memory, ...current.filter((item) => item.id !== memory.id && item.key !== memory.key)].slice(0, limit));
      setMemoryKey("");
      setMemoryValue("");
    } catch {
      setError("暂时无法保存记忆");
    } finally {
      setIsSaving(false);
    }
  }

  async function deleteMemory(memory: CopilotMemory) {
    if (deletingID !== null) return;
    setDeletingID(memory.id);
    setError("");
    try {
      await copilotApi.deleteMemory(memory.id);
      setMemories((current) => current.filter((item) => item.id !== memory.id));
    } catch {
      setError("暂时无法删除记忆");
    } finally {
      setDeletingID(null);
    }
  }

  return (
    <section className="copilot-memory-settings" aria-label="长期记忆">
      <header>
        <span><strong>长期记忆</strong><small>用于后续对话中的偏好与背景</small></span>
        <Link to="/copilot/memories">完整管理 <ChevronRight aria-hidden="true" /></Link>
      </header>
      <form onSubmit={(event) => void saveMemory(event)}>
        <input aria-label="记忆名称" maxLength={80} onChange={(event) => setMemoryKey(event.target.value)} placeholder="记忆名称" value={memoryKey} />
        <input aria-label="记忆内容" maxLength={1000} onChange={(event) => setMemoryValue(event.target.value)} placeholder="例如：优先给出行动清单" value={memoryValue} />
        <button aria-label="保存记忆" disabled={!memoryKey.trim() || !memoryValue.trim() || isSaving} type="submit">{isSaving ? "保存中" : "保存"}</button>
      </form>
      {error ? <small className="copilot-memory-error" role="alert">{error}</small> : null}
      {isLoading ? <small className="copilot-memory-empty" role="status">正在读取...</small> : null}
      {!isLoading && memories.length === 0 && !error ? <small className="copilot-memory-empty">暂无长期记忆</small> : null}
      {memories.length > 0 ? <div className="copilot-memory-list">
        {memories.map((memory) => <article key={memory.id}>
          <span><strong>{memory.key}</strong><small>{memory.value}</small></span>
          <button
            aria-label={`删除记忆 ${memory.key}`}
            disabled={deletingID !== null}
            onClick={() => void deleteMemory(memory)}
            title={`删除记忆 ${memory.key}`}
            type="button"
          >
            <Trash2 aria-hidden="true" />
          </button>
        </article>)}
      </div> : null}
    </section>
  );
}

export default CopilotMemorySettings;
