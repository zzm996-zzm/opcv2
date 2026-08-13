import { ArrowRight, History, Search } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";

import type { SandboxRun } from "../../lib/sandboxApi";
import SandboxFrame from "../../components/sandbox/SandboxFrame";

type SandboxHomeViewProps = {
  onCreate: (initialIdea: string) => Promise<SandboxRun>;
};

const pillars = [
  { title: "八角色独立推演", detail: "每个角色拥有隔离的分析会话", artwork: "/sandbox/pillar-roles.jpg" },
  { title: "风险预警", detail: "强制加入悲观者检验关键假设", artwork: "/sandbox/pillar-risks.jpg" },
  { title: "消费概率估算", detail: "明确标注模型推演与判断依据", artwork: "/sandbox/pillar-decisions.jpg" },
  { title: "可执行建议", detail: "报告可转任务与增长测算输入", artwork: "/sandbox/pillar-history.jpg" }
];

function SandboxHomeView({ onCreate }: SandboxHomeViewProps) {
  const [idea, setIdea] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    const value = idea.trim();
    if (!value) {
      setError("请先描述要推演的项目或情况。");
      return;
    }
    if (submitting) return;
    setSubmitting(true);
    setError("");
    try {
      await onCreate(value);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "创建推演失败，请稍后重试。");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <SandboxFrame copilotMode="home" title={false}>
      <div className="sb-home-main">
        <section className="sb-home-hero">
          <div className="sb-home-copy">
            <h1>商业沙盘</h1>
            <h2>多角色模拟未来，推演不同视角，判断项目机会与风险</h2>
            <p>客户、投资、竞争、渠道、供应、行业、反方和合伙视角彼此隔离分析，再由独立报告会话汇总共识与分歧。</p>
          </div>
          <img alt="多角色商业沙盘" className="sb-home-art" src="/sandbox/home-hero.png" />
          <form className="sb-home-start" onSubmit={submit}>
            <Search aria-hidden="true" className="sb-home-search" size={22} />
            <textarea
              aria-label="描述你要推演的项目或情况"
              id="sandbox-initial-idea"
              onChange={(event) => {
                setIdea(event.target.value);
                if (error) setError("");
              }}
              placeholder="请描述你要推演的项目 / 情况，并补充行业、目标用户、当前阶段、关键约束，以及你最想判断的问题"
              value={idea}
            />
            <div className="sb-home-actions">
              <button disabled={submitting} type="submit">
                {submitting ? "分析中..." : "开始推演"}<ArrowRight size={19} />
              </button>
              <Link to="/sandbox/history"><History size={17} />沙盘历史</Link>
            </div>
            {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
          </form>
        </section>
        <section aria-label="商业沙盘能力" className="sb-pillar-strip">
          {pillars.map(({ artwork, detail, title }) => (
            <article key={title}>
              <span><img alt="" src={artwork} /></span>
              <div><strong>{title}</strong><small>{detail}</small></div>
            </article>
          ))}
        </section>
      </div>
    </SandboxFrame>
  );
}

export default SandboxHomeView;
