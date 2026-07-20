import { ArrowLeft, ArrowRight, Check, CircleHelp, Edit3, Lightbulb, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxSession } from "../../lib/sandboxApi";

type SandboxQuestionsViewProps = {
  onComplete: () => Promise<SandboxSession>;
  onFinished: (session: SandboxSession) => void;
  onSave: (questionKey: string, answer: string, skipped: boolean) => Promise<SandboxSession>;
  session: SandboxSession;
};

function SandboxQuestionsView({ onComplete, onFinished, onSave, session }: SandboxQuestionsViewProps) {
  const questions = useMemo(() => session.intake?.questions ?? [], [session.intake?.questions]);
  const initialIndex = useMemo(() => {
    const firstOpen = questions.findIndex((question) => !question.answer && !question.skipped);
    return firstOpen >= 0 ? firstOpen : Math.max(questions.length - 1, 0);
  }, [questions]);
  const [index, setIndex] = useState(initialIndex);
  const [answer, setAnswer] = useState(questions[initialIndex]?.answer ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const current = questions[index];
  const answeredCount = questions.filter((question) => question.answer || question.skipped).length;
  const progress = questions.length ? Math.round((answeredCount / questions.length) * 100) : 0;

  useEffect(() => {
    setAnswer(current?.answer ?? "");
  }, [current?.answer, current?.key]);

  async function persist(skipped: boolean) {
    if (!current || saving) return;
    const value = answer.trim();
    if (current.required && !skipped && !value) {
      setError("请回答当前问题，或选择暂时不清楚继续。");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const updated = await onSave(current.key, skipped ? "" : value, skipped);
      const last = index >= questions.length - 1;
      if (!last) {
        setIndex((valueIndex) => valueIndex + 1);
        return;
      }
      const completed = await onComplete();
      onFinished(completed ?? updated);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "答案保存失败，请稍后重试。");
    } finally {
      setSaving(false);
    }
  }

  if (!current) {
    return (
      <SandboxFrame copilotMode="questions" copilotProgress={100} copilotProject={session.intake?.initial_idea}>
        <SandboxStepper active={1} />
        <section className="sb-empty-panel">
          <Sparkles size={28} />
          <h1>关键信息已经整理完成</h1>
          <p>当前会话没有待回答问题，可以直接查看信息摘要。</p>
          <button onClick={async () => onFinished(await onComplete())} type="button">查看信息摘要<ArrowRight size={17} /></button>
        </section>
      </SandboxFrame>
    );
  }

  return (
    <SandboxFrame copilotMode="questions" copilotProgress={progress} copilotProject={session.intake?.initial_idea}>
      <SandboxStepper active={1} />
      <section className="sb-work-panel sb-question-panel">
        <header className="sb-panel-header">
          <div className="sb-heading-with-icon">
            <span><Sparkles size={18} /></span>
            <div><h1>推演发起配置</h1><small>AI 动态提问中</small></div>
          </div>
          <button type="button"><Edit3 size={15} />编辑初始设想</button>
        </header>
        <p className="sb-panel-intro">已识别你提供的初步想法，以下问题将帮助建立更准确的沙盘模型。每次回答都会自动保存。</p>
        <article className="sb-initial-idea-card">
          <strong>你的初步想法</strong>
          <p>{session.intake?.initial_idea || session.goal}</p>
        </article>
        {session.intake?.recognized_fields?.length ? (
          <section className="sb-recognized-grid" aria-label="AI 已识别的信息">
            {session.intake.recognized_fields.map((field) => (
              <article key={field.key}>
                <span><Check size={13} strokeWidth={3} /></span>
                <small>{field.label}</small>
                <strong>{field.value}</strong>
              </article>
            ))}
          </section>
        ) : null}
        <section className="sb-question-focus">
          <div className="sb-question-counter">
            <strong>第 {index + 1} 题 / 共 {questions.length} 题</strong>
            <span><CircleHelp size={14} />{current.required ? "必填" : "可跳过"}</span>
          </div>
          <h2>{current.title}</h2>
          <p>{current.hint}</p>
          <textarea
            aria-label={current.title}
            maxLength={current.max_length || 1500}
            onChange={(event) => setAnswer(event.target.value)}
            placeholder={current.placeholder}
            value={answer}
          />
          <div className="sb-answer-meta">
            <span><Lightbulb size={15} />不确定也没关系，可以先留空继续，稍后返回修改。</span>
            <b>{answer.length}/{current.max_length || 1500}</b>
          </div>
          <footer>
            <button disabled={index === 0 || saving} onClick={() => setIndex((value) => Math.max(0, value - 1))} type="button"><ArrowLeft size={17} />上一题</button>
            <button className="is-secondary" disabled={saving} onClick={() => void persist(true)} type="button">暂时不清楚，留空继续</button>
            <button className="is-primary" disabled={saving || (current.required && !answer.trim())} onClick={() => void persist(false)} type="button">
              {saving ? "保存中..." : index === questions.length - 1 ? "完成并查看摘要" : "下一题"}<ArrowRight size={18} />
            </button>
          </footer>
          {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
        </section>
      </section>
    </SandboxFrame>
  );
}

export default SandboxQuestionsView;
