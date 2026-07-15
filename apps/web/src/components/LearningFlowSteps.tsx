const steps = [
  ["收集信息", "获取项目数据与目标"],
  ["能力评估", "多维度能力评估打分"],
  ["差距分析", "定位差距与原因"],
  ["生成报告", "输出诊断结果"],
  ["推荐方案", "推荐学习路径与课程"]
] as const;

function LearningFlowSteps({ active }: { active: number }) {
  return (
    <section className="diagnosis-card diagnosis-steps learning-flow-steps" aria-label="诊断流程">
      {steps.map(([title, description], index) => {
        const step = index + 1;
        return (
          <article className={step < active ? "complete" : step === active ? "active" : ""} key={title}>
            <span>{step < active ? "✓" : step}</span>
            <div><strong>{title}</strong><small>{step < active ? "已完成" : step === active ? description : "等待中"}</small></div>
          </article>
        );
      })}
    </section>
  );
}

export default LearningFlowSteps;
