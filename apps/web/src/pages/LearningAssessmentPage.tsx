import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningDiagnosis, type LearningDimension } from "../lib/learningApi";

const assessmentSteps = [
  ["✓", "收集信息", "已完成", "complete"],
  ["2", "能力评估", "进行中", "active"],
  ["3", "差距分析", "等待中", ""],
  ["4", "生成报告", "等待中", ""],
  ["5", "推荐方案", "等待中", ""]
] as const;

const progressSteps = [
  ["collect", "收集信息", "已完成", "complete"],
  ["assess", "能力评估", "进行中", "active"],
  ["gap", "差距分析", "等待中", ""],
  ["report", "生成报告", "等待中", ""],
  ["plan", "推荐方案", "等待中", ""]
] as const;

const abilityMetrics = [
  ["1", "市场分析能力", 100, "已完成"],
  ["2", "数据分析能力", 75, "分析中"],
  ["3", "用户洞察能力", 60, "分析中"],
  ["4", "产品策划能力", 20, "评估中"],
  ["5", "运营执行能力", 0, "等待中"],
  ["6", "内容创作能力", 0, "等待中"],
  ["7", "增长思维能力", 0, "等待中"],
  ["8", "资源整合能力", 0, "等待中"]
] as const;

const radarAxis = ["专业知识", "实战经验", "方法工具", "思维认知", "执行落地", "学习能力"];

function toAbilityMetric(dimension: LearningDimension, index: number) {
  return [
    String(index + 1),
    dimension.name,
    dimension.score,
    dimension.summary
  ] as const;
}

function LearningAssessmentPage() {
  const location = useLocation();
  const routedDiagnosis = (location.state as { diagnosis?: LearningDiagnosis } | null)?.diagnosis ?? null;
  const [diagnosis, setDiagnosis] = useState<LearningDiagnosis | null>(routedDiagnosis);

  useEffect(() => {
    let active = true;
    if (routedDiagnosis) {
      setDiagnosis(routedDiagnosis);
      return () => {
        active = false;
      };
    }
    learningApi
      .getLatestDiagnosis()
      .then((payload) => {
        if (active) setDiagnosis(payload);
      })
      .catch(() => {
        if (active) setDiagnosis(null);
      });
    return () => {
      active = false;
    };
  }, [routedDiagnosis]);

  const visiblePercent = diagnosis?.overall_score ?? 65;
  const visibleMetrics = diagnosis?.dimensions.length ? diagnosis.dimensions.map(toAbilityMetric) : abilityMetrics;
  const currentStage = diagnosis ? "当前分析阶段：诊断报告已生成" : "当前分析阶段：能力评估";

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page assessment-page" aria-label="能力评估">
        <div className="diagnosis-main">
          <section className="diagnosis-hero">
            <div className="diagnosis-breadcrumb">
              <Link to="/learning">AI教学</Link>
              <span>/</span>
              <Link to="/learning/diagnosis">能力诊断</Link>
            </div>
            <div className="diagnosis-hero-copy">
              <h1>能力诊断</h1>
              <p>基于你的项目、任务与工具使用情况，精准发现能力差距</p>
            </div>
            <div className="diagnosis-target-art" aria-hidden="true" />
          </section>

          <section className="diagnosis-card diagnosis-steps assessment-steps" aria-label="诊断流程">
            {assessmentSteps.map(([number, title, desc, state]) => (
              <article className={state} key={title}>
                <span>{number}</span>
                <div>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </div>
              </article>
            ))}
          </section>

          <section className="diagnosis-card assessment-progress-card" aria-label="诊断进度">
            <header>
              <div>
                <h2>诊断进度</h2>
                <p>AI 正在从多个维度分析你的能力水平</p>
              </div>
              <div className="assessment-percent">
                <span>进行中</span>
                <strong>{visiblePercent}%</strong>
                <i aria-hidden="true"><b /></i>
              </div>
            </header>

            <div className="assessment-progress-track">
              {progressSteps.map(([icon, title, state, tone]) => (
                <article className={tone} key={title}>
                  <i className={`assessment-flow-icon ${icon}`} aria-hidden="true" />
                  <strong>{title}</strong>
                  <small>{state}</small>
                </article>
              ))}
            </div>
          </section>

          <div className="assessment-workbench">
            <section className="diagnosis-card ability-stage-card" aria-label="当前分析阶段">
              <header>
                <h2>{currentStage}</h2>
                <p>AI 正在逐项评估你的各项能力</p>
              </header>
              <div className="ability-metric-list">
                {visibleMetrics.map(([number, label, value, status]) => (
                  <article className={value === 0 ? "pending" : value === 100 ? "done" : ""} key={label}>
                    <span>{number}</span>
                    <strong>{label}</strong>
                    <i aria-hidden="true">
                      <b style={{ width: `${value}%` }} />
                    </i>
                    <em>{value}%</em>
                    <small>{status}</small>
                  </article>
                ))}
              </div>
            </section>

            <section className="diagnosis-card ability-radar-card" aria-label="能力评估模型">
              <header>
                <div>
                  <h2>能力评估模型</h2>
                  <p>我们从 6 个维度、24 个具体能力点进行综合评估</p>
                </div>
                <span aria-label="评估模型说明">i</span>
              </header>

              <div className="radar-wrap" aria-label="能力评估雷达图">
                <svg viewBox="0 0 360 300" role="img">
                  <title>能力评估雷达图</title>
                  <polygon className="radar-grid fill" points="180,42 293,107 293,193 180,258 67,193 67,107" />
                  <polygon className="radar-grid" points="180,78 262,125 262,175 180,222 98,175 98,125" />
                  <polygon className="radar-grid" points="180,114 231,143 231,157 180,186 129,157 129,143" />
                  <line x1="180" y1="42" x2="180" y2="258" />
                  <line x1="293" y1="107" x2="67" y2="193" />
                  <line x1="293" y1="193" x2="67" y2="107" />
                  <polygon className="radar-excellent" points="180,76 259,118 250,188 180,234 95,190 106,112" />
                  <polygon className="radar-target" points="180,105 236,132 224,177 180,204 116,178 126,126" />
                  <polygon className="radar-current" points="180,130 214,146 209,168 180,182 137,171 148,139" />
                  {radarAxis.map((axis, index) => {
                    const coords = [
                      [180, 24],
                      [320, 98],
                      [318, 214],
                      [180, 286],
                      [40, 214],
                      [38, 98]
                    ][index];
                    return (
                      <text key={axis} x={coords[0]} y={coords[1]} textAnchor="middle">
                        {axis}
                      </text>
                    );
                  })}
                </svg>
                <div className="radar-legend" aria-label="雷达图图例">
                  <span className="current">你的水平</span>
                  <span className="target">目标需求</span>
                  <span className="excellent">优秀水平</span>
                </div>
              </div>
            </section>
          </div>

          <section className="diagnosis-card assessment-action-bar" aria-label="诊断辅助操作">
            <p>诊断过程中你可以：继续完善资料、查看学习资源、或与AI助手交流</p>
            <div>
              <Link to="/learning/diagnosis"><span aria-hidden="true">✎</span> 完善资料</Link>
              <Link to="/learning/courses"><span aria-hidden="true">▣</span> 浏览课程</Link>
              <Link to="/copilot"><span aria-hidden="true">⌕</span> 咨询AI助手</Link>
            </div>
          </section>
        </div>

        <aside className="learning-copilot diagnosis-copilot assessment-copilot" aria-label="智活 Copilot 诊断助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全球 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌃</span>
            </div>
          </header>

          <div className="learning-chat assessment-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>嗨，张婧！<br />我正在为你进行能力诊断分析。我会基于你的项目目标和当前情况，全面评估你的能力水平。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>当前正在进行第 2 阶段：能力评估，预计还需要 2-3 分钟完成分析。</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <div>
                <p>我已经完成了 5 个维度的分析，正在评估产品策划能力，请稍候……</p>
                <span className="typing-dots" aria-label="AI 正在输入"><i /><i /><i /></span>
              </div>
            </article>
          </div>

          <nav className="learning-copilot-actions" aria-label="诊断助手快捷入口">
            <Link to="/learning/diagnosis">诊断相关问题 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommendation">如何提升这些能力? <span aria-hidden="true">›</span></Link>
            <Link to="/learning/recommended-courses">查看学习建议 <span aria-hidden="true">›</span></Link>
          </nav>
          <MiniCopilotForm className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default LearningAssessmentPage;
