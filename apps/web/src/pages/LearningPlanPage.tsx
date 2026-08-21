import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { learningApi, type LearningPlan } from "../lib/learningApi";
import { referencePlan } from "../lib/learningReference";
import { tasksApi } from "../lib/tasksApi";

function LearningPlanPage() {
  const [plan, setPlan] = useState<LearningPlan | null>(null);
  const [loading, setLoading] = useState(true);
  const [updatingStage, setUpdatingStage] = useState<number | null>(null);
  const [planError, setPlanError] = useState("");
  const [syncingTasks, setSyncingTasks] = useState(false);
  const [taskSyncMessage, setTaskSyncMessage] = useState("");
  const [taskSyncError, setTaskSyncError] = useState("");

  useEffect(() => {
    let active = true;
    learningApi.getLatestPlan()
      .then((payload) => { if (active) setPlan(payload); })
      .catch(() => { if (active) setPlan(referencePlan); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  async function toggleStage(stageNumber: number, completed: boolean) {
    if (!plan || updatingStage !== null) return;
    setUpdatingStage(stageNumber);
    setPlanError("");
    try {
      await learningApi.updatePlanItem(plan.diagnosis_id, stageNumber, completed);
      const refreshed = await learningApi.getPlan(plan.diagnosis_id);
      setPlan(refreshed);
    } catch {
      setPlanError("学习阶段状态保存失败，请稍后重试。");
    } finally {
      setUpdatingStage(null);
    }
  }

  async function syncToTaskCenter() {
    if (!plan || syncingTasks) return;
    setSyncingTasks(true);
    setTaskSyncMessage("");
    setTaskSyncError("");
    const milestones = plan.stages.map((stage) => stage.milestone).filter(Boolean).join("；");
    try {
      const result = await tasksApi.generateTasks(
        `完成${plan.title}${milestones ? `，重点里程碑：${milestones}` : ""}`,
        { sourceType: "learning_diagnosis", sourceId: plan.diagnosis_id, sourceTitle: plan.title, sourceUrl: "/learning/plan" }
      );
      setTaskSyncMessage(`已创建 ${result.tasks.length} 个学习任务`);
    } catch {
      setTaskSyncError("同步任务失败，请稍后重试。");
    } finally {
      setSyncingTasks(false);
    }
  }

  if (loading) return <PlanState title="正在加载学习计划..." />;
  if (!plan) return <PlanState title="暂无学习计划" />;

  const itemByStage = new Map((plan.items ?? []).map((item) => [item.stage_number, item]));

  return (
    <V4PageShell>
      <section className="learning-page diagnosis-page learning-plan-page" aria-label="系统学习路径">
        <div className="diagnosis-main learning-plan-main">
          <section className="diagnosis-hero learning-plan-hero">
            <div className="diagnosis-breadcrumb"><Link to="/learning">AI教学</Link><span>/</span><strong>系统学习路径</strong></div>
            <div className="diagnosis-hero-copy"><h1>{plan.title}</h1><p>{plan.description}</p><small>{plan.disclaimer || "该历史计划未记录免责声明，请重新诊断后使用。"}</small></div>
          </section>

          <section className="diagnosis-card learning-path-card" aria-label="学习路径阶段">
            <header><h2>诊断推荐行动</h2><ul>{plan.recommendations.map((item) => <li key={item}>{item}</li>)}</ul></header>
            <div className="learning-path-grid">
              {plan.stages.map((stage) => {
                const item = itemByStage.get(stage.number);
                const completed = item?.completed ?? stage.status === "completed";
                return (
                  <article className={`learning-stage ${completed ? "done" : "pending"}`} key={stage.number}>
                    <span className="learning-stage-number">{stage.number}</span>
                    <header><div><h2>{stage.title}</h2><p>{stage.goal}</p></div><strong>{completed ? "已完成" : "未完成"}</strong></header>
                    <div className="learning-stage-body">
                      <section><h3>关联课程</h3>{stage.courses.length ? <ul>{stage.courses.map((course) => <li key={course}>{course}</li>)}</ul> : <p>尚未匹配课程目录</p>}</section>
                      {stage.duration ? <section><h3>预计时长</h3><b>{stage.duration}</b></section> : null}
                      <section><h3>里程碑</h3><p>{stage.milestone}</p></section>
                    </div>
                    <footer>
                      <button disabled={updatingStage !== null} onClick={() => void toggleStage(stage.number, !completed)} type="button">
                        {updatingStage === stage.number ? "保存中..." : completed ? "标记为未完成" : "标记为已完成"}
                      </button>
                    </footer>
                  </article>
                );
              })}
            </div>
          </section>

          <section className="diagnosis-card learning-plan-summary" aria-label="学习路径摘要">
            {plan.weekly_suggestion ? <article className="calendar"><div><h2>用户计划投入</h2><strong>{plan.weekly_suggestion}</strong><p>来自诊断提交</p></div></article> : null}
            <article className="target"><div><h2>快照生成时间</h2><strong>{new Date(plan.generated_at).toLocaleString("zh-CN", { hour12: false })}</strong><p>后续完成状态单独持久化</p></div></article>
          </section>

          <footer className="learning-plan-footer">
            <p>学习路径为模型建议；阶段完成状态由你确认并保存。</p>
            <button disabled={syncingTasks} onClick={syncToTaskCenter} type="button">{syncingTasks ? "同步中..." : "同步到任务中心"}</button>
            {planError ? <span role="alert">{planError}</span> : null}
            {taskSyncMessage ? <span role="status">{taskSyncMessage}</span> : null}
            {taskSyncError ? <span role="alert">{taskSyncError}</span> : null}
          </footer>
        </div>

        <aside className="learning-copilot diagnosis-copilot learning-plan-copilot" aria-label="智活 Copilot 学习路径助手">
          <header><div><strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong><p>学习路径说明</p></div></header>
          <div className="learning-chat"><article><span className="ai-avatar">A</span><p>路径来自诊断快照；课程自动匹配尚未实现时会明确显示为空。</p></article></div>
          <nav className="learning-copilot-actions" aria-label="学习路径助手快捷入口"><Link to="/learning/courses">浏览课程目录 <span aria-hidden="true">›</span></Link><Link to="/learning/history">查看课程进度 <span aria-hidden="true">›</span></Link></nav>
          <MiniCopilotForm activeFilters={{ module: "learning", view: "plan" }} className="learning-copilot-input" />
        </aside>
      </section>
    </V4PageShell>
  );
}

function PlanState({ title }: { title: string }) {
  return <V4PageShell><section className="learning-page diagnosis-card" role="status"><h1>{title}</h1><Link to="/learning/diagnosis">发起能力诊断</Link></section></V4PageShell>;
}

export default LearningPlanPage;
