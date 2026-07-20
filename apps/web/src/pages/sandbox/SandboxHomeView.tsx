import { ArrowRight, BarChart3, Clock3, History, Search, ShieldCheck, Users } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";

import type { SandboxSession } from "../../lib/sandboxApi";
import SandboxFrame from "../../components/sandbox/SandboxFrame";

type SandboxHomeViewProps = {
  onCreate: (initialIdea: string) => Promise<SandboxSession>;
};

const pillars = [
  { title: "多角色推演", detail: "五大核心视角全面洞察", icon: Users },
  { title: "机会与风险识别", detail: "推演发现关键影响因素", icon: ShieldCheck },
  { title: "科学决策支持", detail: "数据驱动更优决策", icon: BarChart3 },
  { title: "推演历史沉淀", detail: "复盘迭代持续优化", icon: Clock3 }
];

function SandboxHomeView({ onCreate }: SandboxHomeViewProps) {
  const [idea, setIdea] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    const value = idea.trim();
    if (!value || submitting) return;
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
            <p>从用户、竞争、运营、增长到风险，多维度模拟真实世界的商业逻辑，助力科学决策。</p>
          </div>
          <img alt="多角色商业沙盘" className="sb-home-art" src="/sandbox/home-hero.jpg" />
          <form className="sb-home-start" onSubmit={submit}>
            <Search aria-hidden="true" className="sb-home-search" size={22} />
            <label htmlFor="sandbox-initial-idea">描述你要推演的项目或情况</label>
            <textarea
              id="sandbox-initial-idea"
              onChange={(event) => setIdea(event.target.value)}
              placeholder="请描述项目、行业、目标用户、当前阶段、关键约束，以及你最想判断的问题"
              value={idea}
            />
            <div className="sb-home-actions">
              <button disabled={!idea.trim() || submitting} type="submit">
                {submitting ? "分析中..." : "开始推演"}<ArrowRight size={19} />
              </button>
              <Link to="/sandbox/history"><History size={17} />沙盘历史</Link>
            </div>
            {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
          </form>
        </section>
        <section aria-label="商业沙盘能力" className="sb-pillar-strip">
          {pillars.map(({ detail, icon: Icon, title }, index) => (
            <article key={title}>
              <span><Icon size={20} /></span>
              <div><strong>{title}</strong><small>{detail}</small></div>
              <b>{String(index + 1).padStart(2, "0")}</b>
            </article>
          ))}
        </section>
      </div>
    </SandboxFrame>
  );
}

export default SandboxHomeView;

