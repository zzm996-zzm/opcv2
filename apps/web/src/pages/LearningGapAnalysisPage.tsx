import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningGaps } from "../lib/learningApi";

const gapSteps = [
  ["✓", "收集信息", "获取项目&数据与目标", "complete"],
  ["✓", "能力评估", "多维度能力评估打分", "complete"],
  ["3", "差距分析", "定位差距与原因", "active"],
  ["4", "生成报告", "输出诊断结果", ""],
  ["5", "推荐方案", "推荐学习路径与资源", ""]
] as const;

const targetTags = ["提升专业能力", "行业分析", "智能客服"];

const gapItems = [
  {
    title: "智能客服案例拆解",
    level: "差距较大",
    current: "45 分",
    target: "80 分",
    gap: "35 分",
    reason: "缺少系统化智能客服场景和业务方案拆解经验。",
    evidence: "仅参与过分析类项目，缺少服务方案落地经验。",
    icon: "bot"
  },
  {
    title: "提示词工程实战",
    level: "差距较大",
    current: "55 分",
    target: "78 分",
    gap: "23 分",
    reason: "提示词设计与迭代优化经验不足。",
    evidence: "使用通用提示词较多，缺少场景化优化实践。",
    icon: "chat"
  },
  {
    title: "数据洞察能力",
    level: "差距中等",
    current: "58 分",
    target: "76 分",
    gap: "18 分",
    reason: "数据线索与洞察输出能力仍需加强。",
    evidence: "输出多为基础报表，缺少深入洞察与建议。",
    icon: "bars"
  }
] as const;

const evidenceSources = [
  ["项目超市", "分析了 2 个相关项目", "提取项目目标、类型与你的参与情况进行对比", "cube"],
  ["任务中心", "分析了 5 个已完成任务", "分析任务类型、复杂度与你的完成质量与效率", "check"],
  ["工具箱", "分析了 8 个常用工具", "评估工具使用频率、深度与能力提升的关联", "case"],
  ["用户画像", "读取了目标、时间投入与偏好", "结合学习目标、时间投入与学习偏好综合评估", "profile"]
] as const;

const priorityItems = [
  ["1", "智能客服案例拆解", "差距 35 分"],
  ["2", "提示词工程实战", "差距 23 分"],
  ["3", "数据洞察能力", "差距 18 分"]
] as const;

const radarAxis = ["AI基础认知", "提示词工程实战", "行业分析方法", "智能客服案例拆解", "数据洞察能力", "工具应用熟练度"];

function LearningGapAnalysisPage() {
  const [gaps, setGaps] = useState<LearningGaps | null>(null);

  useEffect(() => {
    let active = true;
    learningApi
      .getLatestGaps()
      .then((payload) => {
        if (active) setGaps(payload);
      })
      .catch(() => {
        if (active) setGaps(null);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleTargetTags = gaps ? ["提升专业能力", gaps.goal, gaps.project] : targetTags;
  const visibleGapItems = gaps?.gaps.length
    ? gaps.gaps.map((item) => ({
      title: item.name,
      level: item.priority === "high" ? "差距较大" : item.priority === "medium" ? "差距中等" : "持续补强",
      current: `${item.current} 分`,
      target: `${item.target} 分`,
      gap: `${item.gap} 分`,
      reason: item.summary,
      evidence: item.evidence,
      icon: item.priority === "high" ? "bot" : item.priority === "medium" ? "chat" : "bars"
    }))
    : gapItems;
  const visibleEvidenceSources = gaps?.evidence.length
    ? gaps.evidence.map((item, index) => [`诊断依据 ${index + 1}`, item, "来自最新能力诊断", index === 0 ? "profile" : index === 1 ? "cube" : "check"] as const)
    : evidenceSources;
  const visiblePriorityItems = gaps?.gaps.length
    ? gaps.gaps.map((item, index) => [String(index + 1), item.name, `差距 ${item.gap} 分`] as const)
    : priorityItems;

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page gap-page" aria-label="差距分析">
        <div className="diagnosis-main">
          <section className="diagnosis-hero gap-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <Link to="/learning/diagnosis">能力诊断</Link>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你的项目、任务与工具使用情况，精准发现能力差距</p>
            </div>
            <div className="diagnosis-target-art gap-target-art" aria-hidden="true" />
          </section>

          <section className="diagnosis-card diagnosis-steps gap-steps" aria-label="诊断流程">
            {gapSteps.map(([number, title, desc, state]) => (
              <article className={state} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card gap-comparison-card" aria-label="目标要求与当前水平对比">
            <header>
              <h2>目标要求 vs 当前水平</h2>
              <div className="gap-legend" aria-label="差距分析图例">
                <span className="current">当前水平</span>
                <span className="target">目标要求</span>
                <span className="excellent">优秀样本</span>
              </div>
            </header>
            <div className="gap-comparison-body">
              <article className="goal-direction-card">
                <small>你的目标方向</small>
                <h3>{gaps?.goal ?? "智能客服与市场分析能力提升"}</h3>
                <div>
                  {visibleTargetTags.map((tag) => <span key={tag}>{tag}</span>)}
                </div>
                <p><strong>目标描述</strong> 掌握{gaps?.project ?? "智能客服"}相关能力，能够独立完成项目落地与优化。</p>
              </article>

              <div className="gap-radar-wrap" aria-label="差距分析雷达图">
                <svg viewBox="0 0 440 300" role="img">
                  <title>差距分析雷达图</title>
                  <polygon className="gap-radar-grid fill" points="220,34 350,98 350,202 220,266 90,202 90,98" />
                  <polygon className="gap-radar-grid" points="220,68 319,117 319,183 220,232 121,183 121,117" />
                  <polygon className="gap-radar-grid" points="220,102 288,136 288,164 220,198 152,164 152,136" />
                  <line x1="220" y1="34" x2="220" y2="266" />
                  <line x1="350" y1="98" x2="90" y2="202" />
                  <line x1="350" y1="202" x2="90" y2="98" />
                  <polygon className="gap-excellent" points="220,72 318,116 323,190 220,234 118,188 124,112" />
                  <polygon className="gap-target" points="220,86 304,126 309,178 220,216 134,180 137,122" />
                  <polygon className="gap-current" points="220,128 263,146 263,166 220,186 176,168 178,142" />
                  {radarAxis.map((axis, index) => {
                    const coords = [
                      [220, 18],
                      [389, 88],
                      [390, 217],
                      [220, 289],
                      [50, 217],
                      [50, 88]
                    ][index];
                    return (
                      <text key={axis} x={coords[0]} y={coords[1]} textAnchor="middle">
                        {axis}
                      </text>
                    );
                  })}
                </svg>
              </div>

              <article className="gap-explain-card">
                <h3>解读说明</h3>
                <p>雷达图对比了你的当前水平与目标要求（及优秀样本）。在各能力维度的表现，蓝色面积越大表示能力越强，绿色虚线为目标要求，紫色线为优秀样本参考。</p>
              </article>
            </div>
          </section>

          <div className="gap-details-grid">
            <section className="diagnosis-card gap-cause-card" aria-label="关键差距与原因">
              <h2>关键差距与原因 <span>按差距从大到小排序</span></h2>
              <div className="gap-cause-list">
                {visibleGapItems.map((item, index) => (
                  <article key={item.title}>
                    <span className="gap-number">{index + 1}</span>
                    <div>
                      <h3>{item.title}<small>{item.level}</small></h3>
                      <p><b>当前 {item.current}</b><b>目标 {item.target}</b><b>差距 {item.gap}</b></p>
                      <p>原因：{item.reason}</p>
                      <p>证据：{item.evidence}</p>
                    </div>
                    <i className={`gap-art ${item.icon}`} aria-hidden="true" />
                  </article>
                ))}
              </div>
            </section>

            <section className="diagnosis-card gap-evidence-card" aria-label="判断依据">
              <h2>判断依据 <span>基于多维数据分析</span></h2>
              <div>
                {visibleEvidenceSources.map(([title, desc, detail, icon]) => (
                  <article key={title}>
                    <i className={`source-icon ${icon}`} aria-hidden="true" />
                    <div>
                      <h3>{title}</h3>
                      <strong>{desc}</strong>
                      <p>{detail}</p>
                    </div>
                  </article>
                ))}
              </div>
            </section>

            <section className="diagnosis-card gap-priority-card" aria-label="优先补齐顺序">
              <h2>优先补齐顺序 <span>推荐学习顺序</span></h2>
              <div className="gap-priority-body">
                <ol>
                  {visiblePriorityItems.map(([number, title, gap]) => (
                    <li key={title}>
                      <span>{number}</span>
                      <strong>{title}</strong>
                      <small>{gap}</small>
                    </li>
                  ))}
                </ol>
                <div className="priority-stairs" aria-hidden="true">
                  <i>1</i>
                  <i>2</i>
                  <i>3</i>
                </div>
              </div>
              <p>建议先从差距最大的能力开始补齐，循序渐进提升整体能力水平。</p>
            </section>
          </div>

          <section className="gap-footer-actions" aria-label="差距分析操作">
            <Link className="gap-secondary" to="/learning/diagnosis">查看详细依据</Link>
            <Link className="gap-primary" to="/learning/report">继续生成诊断报告 <span aria-hidden="true">→</span></Link>
            <p><span aria-hidden="true">◇</span> 诊断过程保留 & 12 分钟<br />我们将严格保护你的数据安全</p>
          </section>
        </div>

        <aside className="learning-copilot diagnosis-copilot gap-copilot" aria-label="智活 Copilot 差距分析助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全球 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌁</span>
            </div>
          </header>

          <div className="learning-chat gap-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>我已完成能力评估，正在为你分析“当前水平”和“目标要求”之间的差距。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>差距最大的 3 个能力点已经识别出来，我会结合项目、任务和工具使用情况说明判断依据。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>例如，“智能客服案例拆解”分数偏低，主要因为你当前项目更偏分析研究，缺少服务方案拆解和优化经验。</p>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="差距分析助手快捷入口">
            <Link to="/learning/diagnosis">查看差距依据 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommendation">为什么会有这个结果? <span aria-hidden="true">›</span></Link>
            <Link to="/learning/report">下一步做什么 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningGapAnalysisPage;
