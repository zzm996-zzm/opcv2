import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { tasksApi, type Task, type TaskPriority, type TaskStats, type TaskStatus } from "../lib/tasksApi";

type TaskRow = {
  id?: number;
  title: string;
  assignee: string;
  project: string;
  status: string;
  statusCode?: TaskStatus;
  priority: string;
  due: string;
  tools: string[];
  learning: string;
};

type TaskView = "list" | "board" | "calendar";

const taskPageSize = 20;

type TaskEditForm = {
  title: string;
  description: string;
  assignee: string;
  project: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueAt: string;
  tools: string;
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
  { label: "已完成", status: "completed" },
  { label: "提醒中", status: "reminder" }
];

const taskViews: Array<{ label: string; value: TaskView }> = [
  { label: "列表", value: "list" },
  { label: "看板", value: "board" },
  { label: "日历", value: "calendar" }
];

const taskViewDescriptions: Record<TaskView, string> = {
  list: "按优先级和状态快速处理任务",
  board: "按执行状态查看跨项目任务",
  calendar: "按截止日期安排任务节奏"
};

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

function formatCalendarDate(dueAt: string) {
  return new Date(dueAt).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric"
  });
}

function toDateTimeLocal(dueAt?: string) {
  if (!dueAt) return "";
  const date = new Date(dueAt);
  const localTime = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return localTime.toISOString().slice(0, 16);
}

function toTaskEditForm(task: Task): TaskEditForm {
  return {
    title: task.title,
    description: task.description ?? "",
    assignee: task.assignee ?? "",
    project: task.project,
    status: task.status,
    priority: task.priority,
    dueAt: toDateTimeLocal(task.due_at),
    tools: task.tools.join("，"),
    learning: task.learning
  };
}

function parseTools(value: string) {
  return value
    .split(/[,，]/)
    .map((tool) => tool.trim())
    .filter(Boolean);
}

function toTaskRow(task: Task): TaskRow {
  return {
    id: task.id,
    title: task.title,
    assignee: task.assignee ?? "",
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
  const [selectedProject, setSelectedProject] = useState("");
  const [selectedPriority, setSelectedPriority] = useState<TaskPriority | "">("");
  const [searchInput, setSearchInput] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [taskPage, setTaskPage] = useState(1);
  const [taskTotal, setTaskTotal] = useState(0);
  const [projectOptions, setProjectOptions] = useState<string[]>([]);
  const [projectsLoaded, setProjectsLoaded] = useState(false);
  const [projectsLoading, setProjectsLoading] = useState(false);
  const [listLoading, setListLoading] = useState(false);
  const [activeView, setActiveView] = useState<TaskView>("list");
  const [listError, setListError] = useState("");
  const [savingTaskID, setSavingTaskID] = useState<number | null>(null);
  const [taskGoal, setTaskGoal] = useState("");
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [createMessage, setCreateMessage] = useState("");
  const [createError, setCreateError] = useState("");
  const [detailTaskID, setDetailTaskID] = useState<number | null>(null);
  const [detailTask, setDetailTask] = useState<Task | null>(null);
  const [detailForm, setDetailForm] = useState<TaskEditForm | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailSaving, setDetailSaving] = useState(false);
  const [detailDeleting, setDetailDeleting] = useState(false);
  const [detailError, setDetailError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    let active = true;
    setListLoading(true);
    tasksApi
      .listTasks({
        status: selectedStatus,
        project: selectedProject || undefined,
        priority: selectedPriority || undefined,
        q: searchQuery || undefined,
        limit: taskPageSize,
        offset: taskPage > 1 ? (taskPage - 1) * taskPageSize : undefined
      })
      .then((payload) => {
        if (!active) return;
        setApiTasks(payload.tasks);
        setTaskTotal(payload.total ?? payload.tasks.length);
        setListError("");
      })
      .catch((error) => {
        if (!active) return;
        setApiTasks([]);
        setTaskTotal(0);
        setListError(apiErrorMessage(error, "暂时无法读取任务列表"));
      })
      .finally(() => {
        if (active) setListLoading(false);
      });
    return () => {
      active = false;
    };
  }, [selectedStatus, selectedProject, selectedPriority, searchQuery, taskPage]);

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
  const scheduledTasks = apiTasks
    .filter((task) => task.due_at)
    .sort((left, right) => new Date(left.due_at as string).getTime() - new Date(right.due_at as string).getTime());
  const calendarGroups = Array.from(scheduledTasks.reduce<Map<string, Task[]>>((groups, task) => {
    const label = formatCalendarDate(task.due_at as string);
    groups.set(label, [...(groups.get(label) ?? []), task]);
    return groups;
  }, new Map()));
  const unscheduledTasks = apiTasks.filter((task) => !task.due_at);
  const totalPages = Math.max(1, Math.ceil(taskTotal / taskPageSize));
  const hasActiveFilters = Boolean(selectedStatus || selectedProject || selectedPriority || searchQuery);

  useEffect(() => {
    if (taskPage > totalPages) setTaskPage(totalPages);
  }, [taskPage, totalPages]);

  async function refreshTaskStats() {
    try {
      setApiStats(await tasksApi.stats());
    } catch {
      // Keep the last known aggregate stats when refresh is temporarily unavailable.
    }
  }

  async function loadProjectOptions() {
    if (projectsLoaded || projectsLoading) return;
    setProjectsLoading(true);
    try {
      const payload = await tasksApi.listProjects();
      setProjectOptions(payload.projects);
      setProjectsLoaded(true);
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法读取项目选项"));
    } finally {
      setProjectsLoading(false);
    }
  }

  function selectStatus(status?: TaskStatus) {
    setTaskPage(1);
    setSelectedStatus(status);
  }

  function applyTaskSearch() {
    setTaskPage(1);
    setSearchQuery(searchInput.trim());
  }

  function clearTaskFilters() {
    setSearchInput("");
    setSearchQuery("");
    setSelectedStatus(undefined);
    setSelectedProject("");
    setSelectedPriority("");
    setTaskPage(1);
  }

  function taskMatchesCurrentFilters(task: Task) {
    const normalizedQuery = searchQuery.toLowerCase();
    return (!selectedStatus || task.status === selectedStatus) &&
      (!selectedProject || task.project === selectedProject) &&
      (!selectedPriority || task.priority === selectedPriority) &&
      (!normalizedQuery || [task.title, task.description ?? "", task.assignee ?? "", task.project, task.learning].some((value) => value.toLowerCase().includes(normalizedQuery)));
  }

  async function updateTaskStatus(taskID: number, currentStatus: TaskStatus) {
    const action = taskStatusAction(currentStatus);
    setSavingTaskID(taskID);
    try {
      const updated = await tasksApi.updateTask(taskID, { status: action.nextStatus });
      const remainsVisible = taskMatchesCurrentFilters(updated);
      setApiTasks((current) => !remainsVisible
        ? current.filter((task) => task.id !== taskID)
        : current.map((task) => task.id === taskID ? updated : task));
      if (!remainsVisible) setTaskTotal((current) => Math.max(0, current - 1));
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
      if (taskMatchesCurrentFilters(task)) {
        setTaskTotal((current) => current + 1);
        if (taskPage === 1) {
          setApiTasks((current) => [task, ...current].slice(0, taskPageSize));
        } else {
          setTaskPage(1);
        }
      }
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

  async function openTaskDetail(taskID: number) {
    setDetailTaskID(taskID);
    setDetailTask(null);
    setDetailForm(null);
    setDetailLoading(true);
    setDetailError("");
    setConfirmDelete(false);
    try {
      const task = await tasksApi.getTask(taskID);
      setDetailTask(task);
      setDetailForm(toTaskEditForm(task));
    } catch (error) {
      setDetailError(apiErrorMessage(error, "暂时无法读取任务详情"));
    } finally {
      setDetailLoading(false);
    }
  }

  function closeTaskDetail() {
    if (detailSaving || detailDeleting) return;
    setDetailTaskID(null);
    setDetailTask(null);
    setDetailForm(null);
    setDetailError("");
    setConfirmDelete(false);
  }

  function updateDetailField<Key extends keyof TaskEditForm>(key: Key, value: TaskEditForm[Key]) {
    setDetailForm((current) => current ? { ...current, [key]: value } : current);
  }

  async function saveTaskDetail() {
    if (!detailTask || !detailForm || detailSaving) return;
    const title = detailForm.title.trim();
    const project = detailForm.project.trim();
    if (!title || !project) {
      setDetailError("任务标题和所属项目不能为空");
      return;
    }
    setDetailSaving(true);
    setDetailError("");
    try {
      const updated = await tasksApi.updateTask(detailTask.id, {
        title,
        description: detailForm.description !== (detailTask.description ?? "") ? detailForm.description : undefined,
        assignee: detailForm.assignee !== (detailTask.assignee ?? "") ? detailForm.assignee : undefined,
        project,
        status: detailForm.status,
        priority: detailForm.priority,
        dueAt: detailForm.dueAt ? new Date(detailForm.dueAt).toISOString() : undefined,
        clearDueAt: !detailForm.dueAt && Boolean(detailTask.due_at) ? true : undefined,
        tools: parseTools(detailForm.tools),
        learning: detailForm.learning.trim()
      });
      const remainsVisible = taskMatchesCurrentFilters(updated);
      setApiTasks((current) => !remainsVisible
        ? current.filter((task) => task.id !== updated.id)
        : current.map((task) => task.id === updated.id ? updated : task));
      if (!remainsVisible) setTaskTotal((current) => Math.max(0, current - 1));
      await refreshTaskStats();
      setCreateMessage(`已更新任务：${updated.title}`);
      setListError("");
      setDetailTaskID(null);
      setDetailTask(null);
      setDetailForm(null);
    } catch (error) {
      setDetailError(apiErrorMessage(error, "暂时无法保存任务"));
    } finally {
      setDetailSaving(false);
    }
  }

  async function deleteTaskDetail() {
    if (!detailTask || detailDeleting) return;
    setDetailDeleting(true);
    setDetailError("");
    try {
      await tasksApi.deleteTask(detailTask.id);
      setApiTasks((current) => current.filter((task) => task.id !== detailTask.id));
      setTaskTotal((current) => Math.max(0, current - 1));
      await refreshTaskStats();
      setCreateMessage(`已删除任务：${detailTask.title}`);
      setListError("");
      setDetailTaskID(null);
      setDetailTask(null);
      setDetailForm(null);
      setConfirmDelete(false);
    } catch (error) {
      setDetailError(apiErrorMessage(error, "暂时无法删除任务"));
    } finally {
      setDetailDeleting(false);
    }
  }

  function renderTaskCard(task: Task) {
    const action = taskStatusAction(task.status);
    return (
      <article className="task-view-card" key={task.id}>
        <div className="task-view-card-head">
          <div>
            <h3>{task.title}</h3>
            <small>{task.project} · 负责人 {task.assignee || "未指定"}</small>
          </div>
          <span className={`task-priority ${task.priority === "high" ? "high" : task.priority === "medium" ? "mid" : ""}`}>
            {priorityLabels[task.priority]}
          </span>
        </div>
        <div className="task-view-card-meta">
          <span className="task-state">{statusLabels[task.status]}</span>
          <span>{task.due_at ? `截止 ${formatDueAt(task.due_at)}` : "待安排"}</span>
        </div>
        <div className="task-view-card-links">
          <Link to="/tools">工具：{task.tools.join(" / ")}</Link>
          <Link to="/learning">补课：{task.learning}</Link>
        </div>
        <div className="task-view-card-actions">
          <button
            disabled={savingTaskID === task.id}
            onClick={() => void updateTaskStatus(task.id, task.status)}
            type="button"
          >
            {savingTaskID === task.id ? "更新中..." : action.label}
          </button>
          <button onClick={() => void openTaskDetail(task.id)} type="button">查看任务详情</button>
        </div>
      </article>
    );
  }

  return (
    <V4PageShell>
      <section className="module-page tasks-page" aria-label="任务中心">
        <div className="page-title-row">
          <div>
            <h1>任务中心</h1>
            <p>把当前情况和目标拆成可执行任务，并联动工具箱与 AI 教学</p>
          </div>
          <Link className="module-primary-action" to="/tasks/new">新建任务</Link>
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

        <section className={`task-workbench ${activeView === "list" ? "" : "full-width"}`}>
          <div className="task-list-panel">
            <div className="module-section-head">
              <div>
                <h2>任务总览</h2>
                <p>{taskViewDescriptions[activeView]}</p>
              </div>
              <div className="module-chip-row compact">
                {taskViews.map((view) => (
                  <button
                    aria-pressed={activeView === view.value}
                    className={activeView === view.value ? "active" : ""}
                    key={view.value}
                    onClick={() => setActiveView(view.value)}
                    type="button"
                  >
                    {view.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="module-chip-row compact" aria-label="任务状态筛选">
              {statusFilters.map((filter) => (
                <button
                  aria-label={`筛选${filter.label}`}
                  className={selectedStatus === filter.value ? "active" : ""}
                  key={filter.label}
                  onClick={() => selectStatus(filter.value)}
                  type="button"
                >
                  {filter.label}
                </button>
              ))}
            </div>
            <form className="task-filter-bar" aria-label="任务搜索与筛选" onSubmit={(event) => {
              event.preventDefault();
              applyTaskSearch();
            }}>
              <div className="task-search-field">
                <label className="sr-only" htmlFor="task-search">搜索任务</label>
                <input
                  id="task-search"
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder="搜索标题、描述、负责人、项目或补课内容"
                  value={searchInput}
                />
                <button type="submit">搜索</button>
              </div>
              <label className="sr-only" htmlFor="task-project-filter">按项目筛选</label>
              <select
                id="task-project-filter"
                onChange={(event) => {
                  setTaskPage(1);
                  setSelectedProject(event.target.value);
                }}
                onFocus={() => void loadProjectOptions()}
                value={selectedProject}
              >
                <option value="">{projectsLoading ? "读取项目中..." : "全部项目"}</option>
                {projectOptions.map((project) => <option key={project} value={project}>{project}</option>)}
              </select>
              <label className="sr-only" htmlFor="task-priority-filter">按优先级筛选</label>
              <select
                id="task-priority-filter"
                onChange={(event) => {
                  setTaskPage(1);
                  setSelectedPriority(event.target.value as TaskPriority | "");
                }}
                value={selectedPriority}
              >
                <option value="">全部优先级</option>
                <option value="high">高优先级</option>
                <option value="medium">中优先级</option>
                <option value="low">低优先级</option>
              </select>
              {hasActiveFilters ? <button className="task-filter-clear" onClick={clearTaskFilters} type="button">清除筛选</button> : null}
            </form>
            {activeView === "list" ? (
              <div aria-label="任务列表" className="task-table">
                {visibleTasks.length === 0 ? (
                  <div className="module-empty-state" role="status">暂无任务数据</div>
                ) : visibleTasks.map((task) => (
                  <article key={task.id ?? task.title}>
                    <span className={`task-status-dot ${task.status === "已完成" ? "done" : task.status === "进行中" ? "doing" : ""}`} aria-hidden="true" />
                    <div>
                      <h3>{task.title}</h3>
                      <small>{task.project} · 负责人 {task.assignee || "未指定"} · 截止 {task.due}</small>
                    </div>
                    <span className={`task-priority ${task.priority === "高" ? "high" : task.priority === "中" ? "mid" : ""}`}>{task.priority}</span>
                    <span className="task-state">{task.status}</span>
                    {task.id && task.statusCode ? (
                      <div className="task-row-actions">
                        <button
                          disabled={savingTaskID === task.id}
                          onClick={() => void updateTaskStatus(task.id as number, task.statusCode as TaskStatus)}
                          type="button"
                        >
                          {savingTaskID === task.id ? "更新中..." : taskStatusAction(task.statusCode).label}
                        </button>
                        <button onClick={() => void openTaskDetail(task.id as number)} type="button">查看任务详情</button>
                      </div>
                    ) : null}
                    <Link to="/tools">建议工具：{task.tools.join(" / ")}</Link>
                    <Link to="/learning">补课：{task.learning}</Link>
                  </article>
                ))}
              </div>
            ) : null}
            {activeView === "board" ? (
              <div aria-label="任务看板" className="task-view-board" role="region">
                {boardColumnLabels.map((column) => {
                  const columnTasks = apiTasks.filter((task) => task.status === column.status);
                  return (
                    <section aria-label={`${column.label}任务`} className="task-view-column" key={column.status}>
                      <header>
                        <h3>{column.label}</h3>
                        <span>{columnTasks.length}</span>
                      </header>
                      <div>
                        {columnTasks.length === 0
                          ? <div className="module-empty-state">暂无任务</div>
                          : columnTasks.map(renderTaskCard)}
                      </div>
                    </section>
                  );
                })}
              </div>
            ) : null}
            {activeView === "calendar" ? (
              <div aria-label="任务日历" className="task-calendar-view" role="region">
                {apiTasks.length === 0 ? <div className="module-empty-state" role="status">暂无任务数据</div> : null}
                {calendarGroups.map(([label, tasks]) => (
                  <section aria-label={`${label}任务`} className="task-calendar-group" key={label}>
                    <header>
                      <h3>{label}</h3>
                      <span>{tasks.length} 项</span>
                    </header>
                    <div>{tasks.map(renderTaskCard)}</div>
                  </section>
                ))}
                {unscheduledTasks.length > 0 ? (
                  <section aria-label="待安排任务" className="task-calendar-group unscheduled">
                    <header>
                      <h3>待安排</h3>
                      <span>{unscheduledTasks.length} 项</span>
                    </header>
                    <div>{unscheduledTasks.map(renderTaskCard)}</div>
                  </section>
                ) : null}
              </div>
            ) : null}
            {taskTotal > 0 ? (
              <footer className="task-pagination" aria-label="任务分页">
                <button
                  aria-label="上一页"
                  disabled={listLoading || taskPage <= 1}
                  onClick={() => setTaskPage((current) => Math.max(1, current - 1))}
                  type="button"
                >
                  ‹
                </button>
                <span>第 {taskPage} / {totalPages} 页 · 共 {taskTotal} 条</span>
                <button
                  aria-label="下一页"
                  disabled={listLoading || taskPage >= totalPages}
                  onClick={() => setTaskPage((current) => Math.min(totalPages, current + 1))}
                  type="button"
                >
                  ›
                </button>
              </footer>
            ) : null}
          </div>

          {activeView === "list" ? (
            <aside className="task-board-panel" aria-label="任务看板预览">
              <h2>跨项目看板</h2>
              <p>切换到看板视图可按状态处理跨项目任务</p>
              <div>
                {boardColumns.map(([title, items]) => (
                  <section key={title}>
                    <strong>{title}</strong>
                    {items.length === 0 ? <span>暂无任务</span> : items.map((item) => <span key={item}>{item}</span>)}
                  </section>
                ))}
              </div>
            </aside>
          ) : null}
        </section>
        {detailTaskID !== null ? (
          <div className="task-modal-backdrop">
            <section aria-label="任务详情" aria-modal="true" className="task-detail-dialog" role="dialog">
              <button aria-label="关闭任务详情" className="task-detail-close" onClick={closeTaskDetail} type="button">×</button>
              <header>
                <span>任务 #{detailTaskID}</span>
                <h2>任务详情</h2>
                <p>查看并更新任务的执行信息</p>
              </header>
              {detailLoading ? <div className="module-empty-state" role="status">正在读取任务详情...</div> : null}
              {!detailLoading && detailError && !detailForm ? <p className="form-error" role="alert">{detailError}</p> : null}
              {detailForm && detailTask ? (
                <form onSubmit={(event) => {
                  event.preventDefault();
                  void saveTaskDetail();
                }}>
                  <div className="task-detail-form-grid">
                    <label className="wide">
                      <span>任务标题</span>
                      <input onChange={(event) => updateDetailField("title", event.target.value)} value={detailForm.title} />
                    </label>
                    <label className="wide">
                      <span>任务描述</span>
                      <textarea maxLength={1000} onChange={(event) => updateDetailField("description", event.target.value)} value={detailForm.description} />
                    </label>
                    <label>
                      <span>所属项目</span>
                      <input onChange={(event) => updateDetailField("project", event.target.value)} value={detailForm.project} />
                    </label>
                    <label>
                      <span>负责人</span>
                      <input maxLength={100} onChange={(event) => updateDetailField("assignee", event.target.value)} placeholder="输入负责人或外部协作人" value={detailForm.assignee} />
                    </label>
                    <label>
                      <span>截止时间</span>
                      <input onChange={(event) => updateDetailField("dueAt", event.target.value)} type="datetime-local" value={detailForm.dueAt} />
                    </label>
                    <label>
                      <span>任务状态</span>
                      <select onChange={(event) => updateDetailField("status", event.target.value as TaskStatus)} value={detailForm.status}>
                        {statusFilters.filter((item) => item.value).map((item) => (
                          <option key={item.value} value={item.value}>{item.label}</option>
                        ))}
                      </select>
                    </label>
                    <label>
                      <span>优先级</span>
                      <select onChange={(event) => updateDetailField("priority", event.target.value as TaskPriority)} value={detailForm.priority}>
                        <option value="low">低</option>
                        <option value="medium">中</option>
                        <option value="high">高</option>
                      </select>
                    </label>
                    <label className="wide">
                      <span>建议工具</span>
                      <input onChange={(event) => updateDetailField("tools", event.target.value)} value={detailForm.tools} />
                    </label>
                    <label className="wide">
                      <span>补课内容</span>
                      <textarea onChange={(event) => updateDetailField("learning", event.target.value)} value={detailForm.learning} />
                    </label>
                  </div>
                  {detailError ? <p className="form-error" role="alert">{detailError}</p> : null}
                  {confirmDelete ? (
                    <div className="task-delete-confirm" role="alert">
                      <span>删除后无法恢复，请确认当前任务不再需要。</span>
                      <button disabled={detailDeleting} onClick={() => void deleteTaskDetail()} type="button">
                        {detailDeleting ? "删除中..." : "确认删除"}
                      </button>
                    </div>
                  ) : null}
                  <footer>
                    <button className="danger" disabled={detailSaving || detailDeleting} onClick={() => setConfirmDelete((current) => !current)} type="button">
                      {confirmDelete ? "取消删除" : "删除任务"}
                    </button>
                    <div>
                      <button disabled={detailSaving || detailDeleting} onClick={closeTaskDetail} type="button">取消</button>
                      <button className="primary" disabled={detailSaving || detailDeleting} type="submit">
                        {detailSaving ? "保存中..." : "保存修改"}
                      </button>
                    </div>
                  </footer>
                </form>
              ) : null}
            </section>
          </div>
        ) : null}
      </section>
    </V4PageShell>
  );
}

export default TasksPage;
