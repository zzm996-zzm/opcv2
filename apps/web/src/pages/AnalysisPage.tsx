import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";

import { analysisApi, type DirectionResult } from "../lib/analysisApi";

function AnalysisPage() {
  const [intent, setIntent] = useState("");
  const [result, setResult] = useState<DirectionResult | null>(null);
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!intent.trim() || status === "submitting") return;
    setStatus("submitting");
    setError("");
    try {
      const next = await analysisApi.startDirection({ intent });
      setResult(next);
    } catch {
      setError("暂时无法生成分析，请稍后重试");
    } finally {
      setStatus("idle");
    }
  }

  return (
    <main className="analysis-page">
      <header className="placeholder-header">
        <Link className="brand" to="/" aria-label="返回智活AI OPC 首页">
          <span className="brand-mark" aria-hidden="true" />
          <span>智活AI · OPC</span>
        </Link>
        <Link to="/membership">会员中心</Link>
      </header>

      <section className="analysis-hero">
        <p className="section-eyebrow">免费分析</p>
        <h1>告诉我们你的资源或目标，AI 将为你生成可落地的方向与策略建议。</h1>
        <form className="analysis-form" onSubmit={submit}>
          <div className="analysis-tabs" aria-label="分析类型">
            <span>✦ 我有什么，适合做什么</span>
            <span>▣ 拆解一个对标公司</span>
          </div>
          <label htmlFor="analysis-intent">描述你的资源和目标</label>
          <textarea
            id="analysis-intent"
            aria-label="描述你的资源和目标"
            onChange={(event) => setIntent(event.target.value)}
            placeholder="例如：我有10年教培经验，5万本金，目前在成都，每周20小时，想做线上+线下项目"
            rows={5}
            value={intent}
          />
          <button disabled={!intent.trim() || status === "submitting"} type="submit">
            {status === "submitting" ? "分析中" : "生成分析"}
          </button>
        </form>
        {error && <p className="form-error" role="alert">{error}</p>}
      </section>

      {result?.status === "needs_input" && (
        <section className="question-panel">
          <p className="section-eyebrow">中间提问</p>
          <h2>先回答这几个关键问题，结果会更可执行</h2>
          <div className="question-list">
            {result.questions?.map((question) => (
              <article key={question.key}>
                <h3>{question.text}</h3>
                <div>
                  {question.options.map((option) => (
                    <span key={option}>{option}</span>
                  ))}
                </div>
              </article>
            ))}
          </div>
        </section>
      )}

      {result?.status === "completed" && (
        <section className="direction-results">
          <p className="section-eyebrow">你的创业分析结果</p>
          <h2>AI 已根据你的输入生成 3 份方向报告。</h2>
          <div className="direction-grid">
            {result.cards?.map((card) => (
              <article key={card.name} className="direction-card">
                <div className="score">{card.score}</div>
                <h3>{card.name}</h3>
                <p>{card.market_evidence}</p>
                <ul>
                  {card.reasons.map((reason) => (
                    <li key={reason}>{reason}</li>
                  ))}
                </ul>
                <strong>启动难度：{card.difficulty.level}</strong>
                <div className="action-list">
                  {card.actions.map((action, index) => (
                    <span key={action}>{index + 1}. {action}</span>
                  ))}
                </div>
                <p className="upsell">{card.upsell}</p>
              </article>
            ))}
          </div>
        </section>
      )}
    </main>
  );
}

export default AnalysisPage;
