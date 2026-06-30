import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { tasksApi, type Task, type TaskPriority, type TaskStatus } from "../lib/tasksApi";

type TaskRow = {
  id?: number;
  title: string;
  project: string;
  status: string;
  statusCode?: TaskStatus;
  priority: string;
  due: string;
  tools: string[];
  learning: string;
};

const taskStats = [
  ["今日待办", "5"],
  ["进行中", "8"],
  ["已完成", "27"],
  ["提醒中", "4"]
] as const;

const taskRows: TaskRow[] = [
  {
    title: "完成智能客服系统项目商业画布",
    project: "智能客服系统",
    status: "进行中",
    priority: "高",
    due: "今天 18:00",
    tools: ["Notion AI", "ChatGPT"],
    learning: "B 端需求访谈"
  },
  {
    title: "整理首批 20 个潜在客户名单",
    project: "AI线索开发",
    status: "待开始",
    priority: "中",
    due: "明天 10:00",
    tools: ["表格助手", "CRM"],
    learning: "线索评分"
  },
  {
    title: "为短视频代运营项目生成报价模板",
    project: "AI 短视频代运营",
    status: "进行中",
    priority: "中",
    due: "06-16 15:00",
    tools: ["Canva", "剪映专业版"],
    learning: "服务产品化"
  },
  {
    title: "复盘竞品招聘动态并输出应对建议",
    project: "竞品动态监测",
    status: "已完成",
    priority: "低",
    due: "06-12 17:30",
    tools: ["竞品监测"],
    learning: "竞争分析"
  }
];

const boardColumns = [
  ["待开始", ["整理客户名单", "预约顾问沟通"]],
  ["进行中", ["项目商业画布", "报价模板"]],
  ["已完成", ["竞品动态复盘", "工具清单整理"]]
] as const;

const statusLabels: Record<TaskStatus, string> = {
  todo: "待开始",
  in_progress: "进行中",
  completed: "已完成",
  reminder: "提醒中"
};

const priorityLabels: Record<TaskPriority, string> = {
  low: "低",
  medium: "中",
  high: "高"
};

function formatDueAt(dueAt?: string) {
  if (!dueAt) return "待安排";
  return new Date(dueAt).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function toTaskRow(task: Task): TaskRow {
  return {
    id: task.id,
    title: task.title,
    project: task.project,
    status: statusLabels[task.status],
    statusCode: task.status,
    priority: priorityLabels[task.priority],
    due: formatDueAt(task.due_at),
    tools: task.tools,
    learning: task.learning
  };
}

function buildTaskStats(tasks: Task[]) {
  if (tasks.length === 0) return taskStats;
  const countByStatus = tasks.reduce<Record<TaskStatus, number>>(
    (acc, task) => {
      acc[task.status] += 1;
      return acc;
    },
    { todo: 0, in_progress: 0, completed: 0, reminder: 0 }
  );
  return [
    ["今日待办", String(countByStatus.todo + countByStatus.in_progress + countByStatus.reminder)],
    ["进行中", String(countByStatus.in_progress)],
    ["已完成", String(countByStatus.completed)],
    ["提醒中", String(countByStatus.reminder)]
  ] as const;
}

function TasksPage() {
  const [apiTasks, setApiTasks] = useState<Task[]>([]);
  const [listError, setListError] = useState("");
  const [savingTaskID, setSavingTaskID] = useState<number | null>(null);

  useEffect(() => {
    let active = true;
    tasksApi
      .listTasks(20)
      .then((payload) => {
        if (!active) return;
        setApiTasks(payload.tasks);
        setListError("");
      })
      .catch((error) => {
        if (!active) return;
        setApiTasks([]);
        setListError(apiErrorMessage(error, "暂时无法读取任务列表"));
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleTasks = apiTasks.length > 0 ? apiTasks.map(toTaskRow) : taskRows;
  const visibleStats = buildTaskStats(apiTasks);

  async function completeTask(taskID: number) {
    setSavingTaskID(taskID);
    try {
      const updated = await tasksApi.updateTask(taskID, { status: "completed" });
      setApiTasks((current) => current.map((task) => task.id === taskID ? updated : task));
    } catch {
      // Keep the current row unchanged; a global toast system can surface this later.
    } finally {
      setSavingTaskID(null);
    }
  }

  return (
    <V4PageShell>
      <section className="module-page tasks-page" aria-label="任务中心">
        <div className="page-title-row">
          <div>
            <h1>任务中心</h1>
            <p>把当前情况和目标拆成可执行任务，并联动工具箱与 AI 教学</p>
          </div>
          <button className="module-primary-action" type="button">新建任务</button>
        </div>
        {listError ? <p className="form-error" role="alert">{listError}</p> : null}

        <section className="module-overview-card tasks-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">任务总览</span>
            <h2>把计划变成今天能推进的任务</h2>
            <p>任务中心承接项目、工具和竞品动态提醒，形成持续推进的执行闭环。</p>
            <div className="module-stat-strip">
              {visibleStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>
          <form className="module-ai-box compact">
            <label htmlFor="task-goal">AI 生成任务</label>
            <textarea id="task-goal" aria-label="描述任务目标" placeholder="输入当前情况 + 目标..." />
            <button type="button">生成任务表</button>
          </form>
        </section>

        <section className="task-workbench">
          <div className="task-list-panel">
            <div className="module-section-head">
              <div>
                <h2>任务总览</h2>
                <p>列表 / 看板 / 日历多视图，当前显示列表视图</p>
              </div>
              <div className="module-chip-row compact">
                {["列表", "看板", "日历"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>
            <div className="task-table">
              {visibleTasks.map((task) => (
                <article key={task.title}>
                  <span className={`task-status-dot ${task.status === "已完成" ? "done" : task.status === "进行中" ? "doing" : ""}`} aria-hidden="true" />
                  <div>
                    <h3>{task.title}</h3>
                    <small>{task.project} · 截止 {task.due}</small>
                  </div>
                  <span className={`task-priority ${task.priority === "高" ? "high" : task.priority === "中" ? "mid" : ""}`}>{task.priority}</span>
                  <span className="task-state">{task.status}</span>
                  <Link to="/tools">建议工具：{task.tools.join(" / ")}</Link>
                  <Link to="/learning">补课：{task.learning}</Link>
                  {task.id ? (
                    <button
                      disabled={task.statusCode === "completed" || savingTaskID === task.id}
                      onClick={() => void completeTask(task.id as number)}
                      type="button"
                    >
                      {task.statusCode === "completed" ? "已完成" : savingTaskID === task.id ? "更新中..." : "标记完成"}
                    </button>
                  ) : null}
                </article>
              ))}
            </div>
          </div>

          <aside className="task-board-panel" aria-label="任务看板预览">
            <h2>跨项目看板</h2>
            <p>付费后可跨项目拖拽、分组和日历同步</p>
            <div>
              {boardColumns.map(([title, items]) => (
                <section key={title}>
                  <strong>{title}</strong>
                  {items.map((item) => <span key={item}>{item}</span>)}
                </section>
              ))}
            </div>
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default TasksPage;
