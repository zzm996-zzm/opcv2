import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { tasksApi, type Task, type TaskPriority, type TaskStats, type TaskStatus } from "../lib/tasksApi";

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

const emptyTaskStats = [
  ["今日待办", "0"],
  ["进行中", "0"],
  ["已完成", "0"],
  ["提醒中", "0"]
] as const;

const boardColumnLabels: Array<{ label: string; status: TaskStatus }> = [
  { label: "待开始", status: "todo" },
  { label: "进行中", status: "in_progress" },
  { label: "已完成", status: "completed" }
];

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

const statusFilters: Array<{ label: string; value?: TaskStatus }> = [
  { label: "全部" },
  { label: "待开始", value: "todo" },
  { label: "进行中", value: "in_progress" },
  { label: "已完成", value: "completed" },
  { label: "提醒中", value: "reminder" }
];

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
  if (tasks.length === 0) return emptyTaskStats;
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

function statsFromApi(stats: TaskStats) {
  return [
    ["今日待办", String(stats.todo + stats.in_progress + stats.reminder)],
    ["进行中", String(stats.in_progress)],
    ["已完成", String(stats.completed)],
    ["提醒中", String(stats.reminder)]
  ] as const;
}

function taskStatusAction(status: TaskStatus) {
  switch (status) {
    case "todo":
      return { label: "开始任务", nextStatus: "in_progress" as TaskStatus };
    case "completed":
      return { label: "重新打开", nextStatus: "todo" as TaskStatus };
    default:
      return { label: "标记完成", nextStatus: "completed" as TaskStatus };
  }
}

function TasksPage() {
  const taskGoalRef = useRef<HTMLTextAreaElement | null>(null);
  const [apiTasks, setApiTasks] = useState<Task[]>([]);
  const [apiStats, setApiStats] = useState<TaskStats | null>(null);
  const [selectedStatus, setSelectedStatus] = useState<TaskStatus | undefined>();
  const [listError, setListError] = useState("");
  const [savingTaskID, setSavingTaskID] = useState<number | null>(null);
  const [taskGoal, setTaskGoal] = useState("");
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [createMessage, setCreateMessage] = useState("");
  const [createError, setCreateError] = useState("");

  useEffect(() => {
    let active = true;
    tasksApi
      .listTasks({ status: selectedStatus, limit: 20 })
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
  }, [selectedStatus]);

  useEffect(() => {
    let active = true;
    tasksApi
      .stats()
      .then((stats) => {
        if (active) setApiStats(stats);
      })
      .catch(() => {
        if (active) setApiStats(null);
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleTasks = apiTasks.map(toTaskRow);
  const visibleStats = apiStats ? statsFromApi(apiStats) : buildTaskStats(apiTasks);
  const boardColumns = boardColumnLabels.map((column) => [
    column.label,
    apiTasks.filter((task) => task.status === column.status).map((task) => task.title)
  ] as const);

  async function refreshTaskStats() {
    try {
      setApiStats(await tasksApi.stats());
    } catch {
      // Keep the last known aggregate stats when refresh is temporarily unavailable.
    }
  }

  async function updateTaskStatus(taskID: number, currentStatus: TaskStatus) {
    const action = taskStatusAction(currentStatus);
    setSavingTaskID(taskID);
    try {
      const updated = await tasksApi.updateTask(taskID, { status: action.nextStatus });
      setApiTasks((current) => selectedStatus && updated.status !== selectedStatus
        ? current.filter((task) => task.id !== taskID)
        : current.map((task) => task.id === taskID ? updated : task));
      await refreshTaskStats();
      setListError("");
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法更新任务状态"));
    } finally {
      setSavingTaskID(null);
    }
  }

  async function createTaskFromGoal() {
    if (isCreatingTask) return;
    const title = taskGoal.trim();
    if (!title) {
      setCreateError("请输入任务目标");
      return;
    }
    setIsCreatingTask(true);
    setCreateError("");
    setCreateMessage("");
    try {
      const task = await tasksApi.createTask({
        title,
        project: "任务中心",
        priority: "medium",
        tools: ["任务中心"],
        learning: title
      });
      setApiTasks((current) => selectedStatus && task.status !== selectedStatus ? current : [task, ...current]);
      await refreshTaskStats();
      setTaskGoal("");
      setListError("");
      setCreateMessage(`已生成任务：${task.title}`);
    } catch (error) {
      setCreateError(apiErrorMessage(error, "暂时无法生成任务"));
    } finally {
      setIsCreatingTask(false);
    }
  }

  function focusTaskGoal() {
    if (typeof taskGoalRef.current?.scrollIntoView === "function") {
      taskGoalRef.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    taskGoalRef.current?.focus();
  }

  return (
    <V4PageShell>
      <section className="module-page tasks-page" aria-label="任务中心">
        <div className="page-title-row">
          <div>
            <h1>任务中心</h1>
            <p>把当前情况和目标拆成可执行任务，并联动工具箱与 AI 教学</p>
          </div>
          <button className="module-primary-action" onClick={focusTaskGoal} type="button">新建任务</button>
        </div>
        {listError ? <p className="form-error" role="alert">{listError}</p> : null}
        {createMessage ? <p className="form-success" role="status">{createMessage}</p> : null}

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
          <form className="module-ai-box compact" onSubmit={(event) => {
            event.preventDefault();
            void createTaskFromGoal();
          }}>
            <label htmlFor="task-goal">AI 生成任务</label>
            <textarea
              id="task-goal"
              aria-label="描述任务目标"
              onChange={(event) => setTaskGoal(event.target.value)}
              placeholder="输入当前情况 + 目标..."
              ref={taskGoalRef}
              value={taskGoal}
            />
            {createError ? <small className="form-error" role="alert">{createError}</small> : null}
            <button disabled={isCreatingTask} type="submit">{isCreatingTask ? "生成中..." : "生成任务表"}</button>
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
            <div className="module-chip-row compact" aria-label="任务状态筛选">
              {statusFilters.map((filter) => (
                <button
                  aria-label={`筛选${filter.label}`}
                  className={selectedStatus === filter.value ? "active" : ""}
                  key={filter.label}
                  onClick={() => setSelectedStatus(filter.value)}
                  type="button"
                >
                  {filter.label}
                </button>
              ))}
            </div>
            <div className="task-table">
              {visibleTasks.length === 0 ? (
                <div className="module-empty-state" role="status">暂无任务数据</div>
              ) : visibleTasks.map((task) => (
                <article key={task.id ?? task.title}>
                  <span className={`task-status-dot ${task.status === "已完成" ? "done" : task.status === "进行中" ? "doing" : ""}`} aria-hidden="true" />
                  <div>
                    <h3>{task.title}</h3>
                    <small>{task.project} · 截止 {task.due}</small>
                  </div>
                  <span className={`task-priority ${task.priority === "高" ? "high" : task.priority === "中" ? "mid" : ""}`}>{task.priority}</span>
                  <span className="task-state">{task.status}</span>
                  {task.id && task.statusCode ? (
                    <button
                      disabled={savingTaskID === task.id}
                      onClick={() => void updateTaskStatus(task.id as number, task.statusCode as TaskStatus)}
                      type="button"
                    >
                      {savingTaskID === task.id ? "更新中..." : taskStatusAction(task.statusCode).label}
                    </button>
                  ) : null}
                  <Link to="/tools">建议工具：{task.tools.join(" / ")}</Link>
                  <Link to="/learning">补课：{task.learning}</Link>
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
                  {items.length === 0 ? <span>暂无任务</span> : items.map((item) => <span key={item}>{item}</span>)}
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
