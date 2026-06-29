import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { analysisApi, type AnalysisSession } from "../lib/analysisApi";
import { CdkTopNav } from "./AnalysisPage";

const historyFallback = [
  ["本地教培转化顾问", "围绕教培经验、低预算启动和本地机构获客给出的方向报告", "91分", "已完成"],
  ["AI 内容获客陪跑", "拆解内容生产、私域承接和小单咨询入口的低成本组合", "86分", "已完成"],
  ["社区门店增长方案", "适合线下资源、人脉转介绍和门店服务能力的轻资产方向", "待补充", "需要补充信息"]
] as const;

const historyFilters = ["全部报告", "已完成", "待补充", "近7天"] as const;

function AnalysisHistoryPage() {
  const [sessions, setSessions] = useState<AnalysisSession[]>([]);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");

  useEffect(() => {
    let active = true;
    analysisApi
      .listSessions(20)
      .then((payload) => {
        if (!active) return;
        setSessions(payload.sessions);
        setStatus("ready");
      })
      .catch(() => {
        if (active) setStatus("error");
      });

    return () => {
      active = false;
    };
  }, []);

  const completedCount = sessions.filter((session) => session.status === "completed").length;
  const inputCount = sessions.filter((session) => session.status === "needs_input").length;
  const hasSessions = status === "ready" && sessions.length > 0;

  return (
    <main className="cdk-analysis-page cdk-history-page">
      <CdkTopNav active="免费分析" />

      <section className="cdk-history-hero" aria-label="分析历史">
        <div>
          <span>REPORT HISTORY</span>
          <h1>分析历史</h1>
          <p>查看过往方向分析、待补充问题和已生成的方向卡，继续推进最适合你的创业路径。</p>
        </div>
        <Link className="analysis-history-link" to="/analysis">新建分析</Link>
      </section>

      <section className="cdk-history-layout">
        <div className="cdk-history-main">
          <section className="analysis-panel analysis-history-summary">
          <article>
            <small>历史报告</small>
            <strong>{sessions.length}</strong>
          </article>
          <article>
            <small>已完成</small>
            <strong>{completedCount}</strong>
          </article>
          <article>
            <small>待补充</small>
            <strong>{inputCount}</strong>
          </article>
        </section>

          <section className="cdk-history-toolbar" aria-label="历史报告筛选">
            <div>
              {historyFilters.map((filter, index) => (
                <button className={index === 0 ? "active" : ""} key={filter} type="button">{filter}</button>
              ))}
            </div>
            <label>
              <span aria-hidden="true">⌕</span>
              <input aria-label="搜索历史报告" placeholder="搜索报告名称、目标或关键词" />
            </label>
          </section>

        {status === "loading" && (
          <section className="analysis-panel analysis-state-card">
            <h2>正在加载历史报告</h2>
            <p>稍等片刻，系统正在读取你最近的分析记录。</p>
          </section>
        )}

        {status === "error" && (
          <section className="analysis-panel analysis-state-card">
            <h2>暂时无法读取历史报告</h2>
            <p>请稍后重试，或先新建一次分析。</p>
          </section>
        )}

        {status === "ready" && sessions.length === 0 && (
          <section className="analysis-panel analysis-state-card">
            <h2>还没有分析记录</h2>
            <p>完成一次免费分析后，方向卡和追问记录会出现在这里。</p>
            <Link to="/analysis">开始第一次分析</Link>
          </section>
        )}

        {status === "ready" && sessions.length > 0 && (
          <section className="analysis-history-list" aria-label="历史报告列表">
            {sessions.map((session) => (
              <article className="analysis-history-row" key={session.id}>
                <em aria-hidden="true" />
                <div>
                  <span>{session.status === "completed" ? "已完成" : "需要补充信息"}</span>
                  <h2>{sessionTitle(session)}</h2>
                  <p>{session.intent}</p>
                </div>
                <div className="analysis-history-meta">
                  <strong>{sessionScore(session)}</strong>
                  <small>{formatDate(session.created_at)}</small>
                </div>
                <div className="analysis-history-actions">
                  <Link to={`/analysis/sessions/${session.id}`}>查看报告</Link>
                </div>
              </article>
            ))}
          </section>
        )}

          {!hasSessions && status !== "loading" && (
            <section className="cdk-history-samples" aria-label="历史报告样例">
              {historyFallback.map(([title, desc, score, state], index) => (
                <article key={title}>
                  <i className={`sample-${index + 1}`} aria-hidden="true" />
                  <span>{state}</span>
                  <h2>{title}</h2>
                  <p>{desc}</p>
                  <strong>{score}</strong>
                </article>
              ))}
            </section>
          )}
        </div>

        <aside className="cdk-history-aside" aria-label="历史报告操作建议">
          <section>
            <h2>下一步建议</h2>
            <p>优先打开最高分报告，确认启动路径后进入 VIP 获客，生成首批可触达客户。</p>
            <Link to="/leads">进入 VIP 获客 ›</Link>
          </section>
          <section>
            <h2>报告状态</h2>
            <div><b />已完成报告可查看完整方向、对标案例和行动清单。</div>
            <div><b />待补充报告需要回答 AI 追问后继续生成。</div>
            <div><b />所有报告仅对当前账号可见，支持后续同步到 CRM。</div>
          </section>
        </aside>
      </section>
    </main>
  );
}

function sessionTitle(session: AnalysisSession) {
  return session.result?.cards?.[0]?.name || session.questions?.[0]?.text || "方向分析记录";
}

function sessionScore(session: AnalysisSession) {
  const score = session.result?.cards?.[0]?.score;
  return typeof score === "number" ? `${score}分` : "--";
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}

export default AnalysisHistoryPage;
