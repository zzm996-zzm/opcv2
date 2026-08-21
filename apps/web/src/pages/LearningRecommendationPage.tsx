import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import LearningFlowSteps from "../components/LearningFlowSteps";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningRecommendations } from "../lib/learningApi";
import { referenceRecommendations } from "../lib/learningReference";

function LearningRecommendationPage() {
  const [recommendations, setRecommendations] = useState<LearningRecommendations | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    learningApi.getLatestRecommendations()
      .then((payload) => { if (active) setRecommendations(payload); })
      .catch(() => { if (active) setRecommendations(referenceRecommendations); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  if (loading) return <RecommendationState title="正在加载学习建议..." />;
  if (!recommendations) return <RecommendationState title="暂无学习建议" />;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page recommendation-page" aria-label="学习推荐方案">
        <div className="diagnosis-main recommendation-main">
          <section className="diagnosis-hero recommendation-hero">
            <div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>推荐方案</strong></div>
            <div className="diagnosis-hero-copy"><h1>推荐方案</h1><p>基于诊断结果，为你推荐最值得优先学习的课程与提升动作</p></div>
          </section>

          <LearningFlowSteps active={5} />

          <section className="diagnosis-card recommendation-focus-card" aria-label="优先补强方向">
            <h2>优先补强方向</h2>
            <div className="recommendation-focus-grid">
              {recommendations.focus.map((item, index) => (
                <article key={item.name}><b>{index + 1}</b><h3>{item.name}</h3><span>{item.priority}</span><p>{item.summary}</p></article>
              ))}
            </div>
          </section>

          <section className="diagnosis-card recommendation-course-card" aria-label="模型学习建议">
            <h2>模型学习建议</h2><ul>{recommendations.recommendations.map((item) => <li key={item}>{item}</li>)}</ul>
            <p>课程目录尚未与诊断结果做自动匹配，请在全部课程中按能力名称核对。</p>
            <Link to="/learning/courses">浏览真实课程目录</Link>
          </section>

          <section className="diagnosis-card recommendation-method-card" aria-label="建议学习方式">
            <h2>建议学习方式</h2>
            <div>{recommendations.methods.map((method) => <article key={method.title}><h3>{method.title}</h3><strong>{method.value}</strong><p>{method.detail}</p></article>)}</div>
          </section>

          <section className="diagnosis-card" aria-label="评估依据与假设">
            <h2>关键假设</h2><ul>{(recommendations.assumptions ?? []).map((item) => <li key={item}>{item}</li>)}</ul>
          </section>

          <section className="recommendation-footer-actions" aria-label="下一步可选动作">
            <h2>下一步可选动作</h2><Link to="/learning/plan">查看系统学习路径</Link><Link to="/learning/diagnosis">重新诊断</Link>
          </section>
        </div>
        <aside className="learning-copilot diagnosis-copilot recommendation-copilot" aria-label="智活 Copilot 推荐方案助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>学习建议说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>这些建议来自诊断时保存的快照，不会随页面刷新自行改写。</p></article></div>
          <MiniCopilotForm activeFilters={{ module: "learning", view: "recommendations" }} className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function RecommendationState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page diagnosis-card" role="status"><h1>{title}</h1><Link to="/learning/diagnosis">发起能力诊断</Link></section></V4PageShell>;
}

export default LearningRecommendationPage;
