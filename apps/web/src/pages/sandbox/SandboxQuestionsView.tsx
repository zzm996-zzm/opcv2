import { ArrowLeft, ArrowRight, Check, CircleHelp, Edit3, Lightbulb, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxRun } from "../../lib/sandboxApi";

type Props = { onAnswer: (input: { revision: number; answers?: { key: string; value: string; skipped?: boolean }[]; skip?: boolean }) => Promise<SandboxRun>; onFinished: (run: SandboxRun) => void; run: SandboxRun };

function SandboxQuestionsView({ onAnswer, onFinished, run }: Props) {
  const questions = useMemo(() => run.next_questions?.length ? run.next_questions : run.questions ?? [], [run.next_questions, run.questions]);
  const initialIdea = run.context.extra?.trim() ?? "";
  const isSmartCompletion = Boolean(initialIdea);
  const recognizedFacts = useMemo(() => [
    { label: "产品方向", value: run.product.name || "待补充" },
    { label: "目标人群", value: run.context.target_customer || "待补充" },
    { label: "定价区间", value: run.product.price_cents ? `${formatPrice(run.product.price_cents)} / 次` : run.context.market || "待补充" },
    { label: "销售场景", value: run.context.channel || "待补充" }
  ], [run.context.channel, run.context.market, run.context.target_customer, run.product.name, run.product.price_cents]);
  const [index, setIndex] = useState(0);
  const current = questions[index];
  const [answer, setAnswer] = useState(current?.answer ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => setAnswer(current?.answer ?? ""), [current?.key, current?.answer]);

  async function persist(skipped: boolean) {
    if (!current || saving) return;
    if (!skipped && current.required && !answer.trim()) { setError("请回答当前问题，或选择跳过并记录假设。"); return; }
    setSaving(true); setError("");
    try {
      const updated = await onAnswer({ revision: run.revision, answers: [{ key: current.key, value: skipped ? "" : answer.trim(), skipped }] });
      if (index < questions.length - 1 && updated.status === "clarifying") setIndex((value) => value + 1);
      else onFinished(updated);
    } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "答案保存失败，请稍后重试。"); }
    finally { setSaving(false); }
  }

  async function skipAll() {
    if (saving) return;
    setSaving(true); setError("");
    try { onFinished(await onAnswer({ revision: run.revision, skip: true })); }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : "跳过澄清失败，请稍后重试。"); }
    finally { setSaving(false); }
  }

  if (!current) return <SandboxFrame copilotMode="questions"><SandboxStepper active={2} /><section className="sb-empty-panel"><Sparkles size={28} /><h1>信息已经整理完成</h1><p>当前没有待回答的澄清问题，可以继续选择角色。</p><button onClick={() => onFinished(run)} type="button">继续<ArrowRight size={17} /></button></section></SandboxFrame>;

  return <SandboxFrame copilotMode="questions" copilotProgress={run.completeness * 100} copilotProject={run.product.name}>
    <SandboxStepper active={2} smart={isSmartCompletion} />
    <section className={`sb-work-panel sb-question-panel${isSmartCompletion ? " is-smart-completion" : ""}`}>
      <header className="sb-panel-header">
        <div className="sb-heading-with-icon"><span>{isSmartCompletion ? <Check size={18} strokeWidth={2.8} /> : <Sparkles size={18} />}</span><div><h1>{isSmartCompletion ? "推演发起配置" : "补充关键信息"}</h1><small>{isSmartCompletion ? "AI 动态提问中" : "每轮最多 3 个问题，回答会自动保存"}</small></div></div>
        {isSmartCompletion ? <Link className="sb-restart-link" to="/sandbox"><Edit3 size={15} />重新输入</Link> : <span className="sb-selected-count">完成度 {Math.round(run.completeness * 100)}%</span>}
      </header>
      <p className="sb-panel-intro">{isSmartCompletion ? "已识别你提供的初步想法，AI 正在分析并为你生成最关键的补充问题。请完成后继续下一题。" : `已识别“${run.product.name}”，补充的信息会冻结到本轮推演上下文中。`}</p>
      <article className="sb-initial-idea-card"><strong>{isSmartCompletion ? "你的初步想法" : "当前项目"}</strong><p>{initialIdea || run.product.name}{run.product.selling_point ? `：${run.product.selling_point}` : ""}</p>{isSmartCompletion ? <Link to="/sandbox"><Edit3 size={14} />编辑</Link> : null}</article>
      {isSmartCompletion ? <section className="sb-recognized-facts" aria-label="AI 已识别到的关键信息"><h2>AI 已识别到的关键信息</h2><div className="sb-recognized-grid">{recognizedFacts.map((fact) => <article className={fact.value === "待补充" ? "is-pending" : ""} key={fact.label}><span><Check size={13} strokeWidth={3} /></span><small>{fact.label}</small><strong>{fact.value}</strong></article>)}</div></section> : null}
      <section className="sb-question-focus"><div className="sb-question-counter"><strong>第 {index + 1} 题 / 本轮 {questions.length} 题</strong><span><CircleHelp size={14} />{current.required ? "必填" : "可跳过"}</span></div><h2>{current.question}</h2><textarea aria-label={current.question} onChange={(event) => setAnswer(event.target.value)} placeholder="填写你确定的信息；不确定可以跳过" value={answer} /><div className="sb-answer-meta"><span><Lightbulb size={15} />跳过会作为明确假设写入报告。</span><b>{answer.length}/1500</b></div><footer><button disabled={index === 0 || saving} onClick={() => setIndex((value) => Math.max(0, value - 1))} type="button"><ArrowLeft size={17} />上一题</button><button className="is-secondary" disabled={saving} onClick={() => void persist(true)} type="button">跳过本题</button><button className="is-primary" disabled={saving} onClick={() => void persist(false)} type="button">{saving ? "保存中..." : index === questions.length - 1 ? "完成补充" : "下一题"}<ArrowRight size={18} /></button></footer><button className="sb-quiet-button" disabled={saving} onClick={() => void skipAll()} type="button">跳过全部并进入角色选择</button>{error ? <p className="sb-inline-error" role="alert">{error}</p> : null}</section>
    </section>
  </SandboxFrame>;
}

function formatPrice(priceCents: number) {
  const amount = priceCents / 100;
  return `¥${Number.isInteger(amount) ? amount : amount.toFixed(2)}`;
}

export default SandboxQuestionsView;
