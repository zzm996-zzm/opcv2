import { Fragment, useEffect, useRef, useState, type CSSProperties, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { ArrowDown, ArrowUp, ChevronDown, ChevronRight, Download, Eye, EyeOff, Paperclip, Plus, RotateCcw, SlidersHorizontal, Trash2, Upload, X } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { useCopilotPanelVisibility } from "../components/CopilotPanelVisibility";
import UnifiedCopilotPanel from "../components/UnifiedCopilotPanel";
import { apiErrorMessage } from "../lib/apiErrors";
import { ApiRequestError } from "../lib/apiRequest";
import { tasksApi, type ReminderRecurrence, type Task, type TaskActivity, type TaskAIDraft, type TaskAttachment, type TaskComment, type TaskGroup, type TaskListColumn, type TaskPriority, type TaskReminder, type TaskSort, type TaskStats, type TaskStatus, type TaskSubtask } from "../lib/tasksApi";

type TaskRow = {
  id?: number;
  title: string;
  assignee: string;
  project: string;
  status: string;
  statusCode?: TaskStatus;
  priority: string;
  due: string;
  tags: string[];
  tools: string[];
  learning: string;
  sourceTitle?: string;
  sourceURL?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
  progress: number;
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
  tags: string;
  tools: string;
  learning: string;
  progress: number;
};

type AIDraftPreviewItem = {
  draftIndex: number;
  selected: boolean;
  title: string;
  description: string;
  assignee: string;
  project: string;
  priority: TaskPriority;
  dueAt: string;
  tags: string;
  tools: string;
  learning: string;
};

function TaskModalPortal({ children }: { children: ReactNode }) {
  useEffect(() => {
    const previousOverflow = document.body.style.overflow;
    const previousPaddingRight = document.body.style.paddingRight;
    const scrollbarWidth = Math.max(0, window.innerWidth - document.documentElement.clientWidth);
    const bodyPaddingRight = Number.parseFloat(window.getComputedStyle(document.body).paddingRight) || 0;

    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) document.body.style.paddingRight = `${bodyPaddingRight + scrollbarWidth}px`;

    return () => {
      document.body.style.overflow = previousOverflow;
      document.body.style.paddingRight = previousPaddingRight;
    };
  }, []);

  return createPortal(children, document.body);
}

const emptyTaskStats = [
  ["今日待办", "0"],
  ["进行中", "0"],
  ["已完成", "0"],
  ["提醒中", "0"]
] as const;

const boardColumnLabels: Array<{ label: string; status: TaskStatus }> = [
  { label: "待开始", status: "todo" },
  { label: "进行中", status: "in_progress" },
  { label: "评审中", status: "review" },
  { label: "阻塞", status: "blocked" },
  { label: "已完成", status: "completed" },
  { label: "已取消", status: "cancelled" },
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
  review: "评审中",
  blocked: "阻塞",
  completed: "已完成",
  cancelled: "已取消",
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
  { label: "评审中", value: "review" },
  { label: "阻塞", value: "blocked" },
  { label: "已完成", value: "completed" },
  { label: "已取消", value: "cancelled" },
  { label: "提醒中", value: "reminder" }
];

const taskSortOptions: Array<{ label: string; value: TaskSort }> = [
  { label: "最新创建", value: "created_at" },
  { label: "最近更新", value: "updated_at" },
  { label: "截止时间", value: "due_at" },
  { label: "优先级", value: "priority" },
  { label: "进度", value: "progress" }
];

const taskGroupOptions: Array<{ label: string; value: TaskGroup | "" }> = [
  { label: "不分组", value: "" },
  { label: "按状态分组", value: "status" },
  { label: "按负责人分组", value: "assignee" },
  { label: "按项目分组", value: "project" },
  { label: "按优先级分组", value: "priority" },
  { label: "按来源分组", value: "source" }
];

const taskColumnDefinitions: Array<{ key: TaskListColumn; label: string; required?: boolean }> = [
  { key: "title", label: "任务标题", required: true },
  { key: "project", label: "所属项目" },
  { key: "assignee", label: "负责人" },
  { key: "due_at", label: "截止时间" },
  { key: "priority", label: "优先级" },
  { key: "status", label: "状态" },
  { key: "tags", label: "标签" },
  { key: "progress", label: "进度" },
  { key: "source", label: "来源模块" },
  { key: "created_at", label: "创建时间" },
  { key: "updated_at", label: "更新时间" }
];

const defaultTaskColumns: TaskListColumn[] = ["title", "project", "assignee", "due_at", "priority", "status", "tags", "progress", "source"];

function validTaskColumns(columns: unknown): columns is TaskListColumn[] {
  if (!Array.isArray(columns) || !columns.includes("title")) return false;
  const allowed = new Set(taskColumnDefinitions.map((column) => column.key));
  return new Set(columns).size === columns.length && columns.every((column) => typeof column === "string" && allowed.has(column as TaskListColumn));
}

function taskStatusFromQuery(value: string | null) {
  return statusFilters.find((filter) => filter.value === value)?.value;
}

function taskPriorityFromQuery(value: string | null): TaskPriority | "" {
  return value === "low" || value === "medium" || value === "high" ? value : "";
}

function taskSortFromQuery(value: string | null): TaskSort {
  return taskSortOptions.some((option) => option.value === value) ? value as TaskSort : "created_at";
}

function taskGroupFromQuery(value: string | null): TaskGroup | "" {
  return taskGroupOptions.some((option) => option.value === value) ? value as TaskGroup | "" : "";
}

function taskViewFromQuery(value: string | null): TaskView {
  return taskViews.some((view) => view.value === value) ? value as TaskView : "list";
}

function taskPageFromQuery(value: string | null) {
  const page = Number(value);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function currentCalendarMonth() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

function calendarMonthFromQuery(value: string | null) {
  return value && /^\d{4}-(0[1-9]|1[0-2])$/.test(value) ? value : currentCalendarMonth();
}

type SubtaskTreeRow = {
  item: TaskSubtask;
  depth: number;
  descendantCount: number;
  hasChildren: boolean;
};

function subtaskDescendantIDs(items: TaskSubtask[], parentID: number) {
  const children = new Map<number, number[]>();
  items.forEach((item) => {
    if (item.parent_subtask_id === undefined) return;
    children.set(item.parent_subtask_id, [...(children.get(item.parent_subtask_id) ?? []), item.id]);
  });
  const descendants: number[] = [];
  const visited = new Set<number>([parentID]);
  const pending = [...(children.get(parentID) ?? [])];
  while (pending.length > 0) {
    const id = pending.shift() as number;
    if (visited.has(id)) continue;
    visited.add(id);
    descendants.push(id);
    pending.push(...(children.get(id) ?? []));
  }
  return descendants;
}

function flattenSubtaskTree(items: TaskSubtask[], collapsedIDs: Set<number>): SubtaskTreeRow[] {
  const itemIDs = new Set(items.map((item) => item.id));
  const children = new Map<number | null, TaskSubtask[]>();
  items.forEach((item) => {
    const parentID = item.parent_subtask_id !== undefined && itemIDs.has(item.parent_subtask_id)
      ? item.parent_subtask_id
      : null;
    children.set(parentID, [...(children.get(parentID) ?? []), item]);
  });
  children.forEach((siblings) => siblings.sort((left, right) => left.created_at.localeCompare(right.created_at) || left.id - right.id));

  const rows: SubtaskTreeRow[] = [];
  const visited = new Set<number>();
  const append = (item: TaskSubtask, depth: number) => {
    if (visited.has(item.id)) return;
    visited.add(item.id);
    const childItems = children.get(item.id) ?? [];
    rows.push({
      item,
      depth,
      descendantCount: subtaskDescendantIDs(items, item.id).length,
      hasChildren: childItems.length > 0
    });
    if (!collapsedIDs.has(item.id)) childItems.forEach((child) => append(child, depth + 1));
  };
  (children.get(null) ?? []).forEach((item) => append(item, 0));
  items.forEach((item) => append(item, 0));
  return rows;
}

function calendarMonthRange(month: string) {
  const [year, monthNumber] = month.split("-").map(Number);
  const nextYear = monthNumber === 12 ? year + 1 : year;
  const nextMonth = monthNumber === 12 ? 1 : monthNumber + 1;
  return {
    from: `${month}-01`,
    to: `${nextYear}-${String(nextMonth).padStart(2, "0")}-01`
  };
}

function shiftCalendarMonth(month: string, offset: number) {
  const [year, monthNumber] = month.split("-").map(Number);
  const shifted = new Date(year, monthNumber - 1 + offset, 1);
  return `${shifted.getFullYear()}-${String(shifted.getMonth() + 1).padStart(2, "0")}`;
}

function calendarMonthLabel(month: string) {
  const [year, monthNumber] = month.split("-").map(Number);
  return `${year}年${monthNumber}月`;
}

function taskGroupLabel(task: Task, group: TaskGroup) {
  switch (group) {
    case "status":
      return statusLabels[task.status];
    case "assignee":
      return task.assignee || "未指定负责人";
    case "project":
      return task.project || "未关联项目";
    case "priority":
      return `${priorityLabels[task.priority]}优先级`;
    case "source":
      return task.source_title || "手工创建";
  }
}

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
    tags: (task.tags ?? []).join("，"),
    tools: task.tools.join("，"),
    learning: task.learning,
    progress: task.progress ?? 0
  };
}

function toAIDraftPreviewItem(task: Task, draftIndex: number): AIDraftPreviewItem {
  return {
    draftIndex,
    selected: true,
    title: task.title,
    description: task.description ?? "",
    assignee: task.assignee ?? "",
    project: task.project,
    priority: task.priority,
    dueAt: toDateTimeLocal(task.due_at),
    tags: (task.tags ?? []).join("，"),
    tools: (task.tools ?? []).join("，"),
    learning: task.learning ?? ""
  };
}

function parseList(value: string) {
  return Array.from(new Set(value
    .split(/[,，]/)
    .map((tool) => tool.trim())
    .filter(Boolean)));
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
    tags: task.tags ?? [],
    tools: task.tools,
    learning: task.learning,
    sourceTitle: task.source_title,
    sourceURL: task.source_url,
    createdAt: formatDueAt(task.created_at),
    updatedAt: formatDueAt(task.updated_at),
    version: task.version,
    progress: task.progress ?? 0
  };
}

function buildTaskStats(tasks: Task[]) {
  if (tasks.length === 0) return emptyTaskStats;
  const countByStatus = tasks.reduce<Record<TaskStatus, number>>(
    (acc, task) => {
      acc[task.status] += 1;
      return acc;
    },
    { todo: 0, in_progress: 0, review: 0, blocked: 0, completed: 0, cancelled: 0, reminder: 0 }
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
    ["提醒中", String(stats.reminder)],
    ["48小时内到期", String(stats.due_soon ?? 0)],
    ["48小时内超时", String(stats.timed_out ?? 0)],
    ["严重逾期", String(stats.overdue ?? 0)]
  ] as const;
}

function taskStatusOptions(status: TaskStatus) {
  const allowed: Record<TaskStatus, TaskStatus[]> = {
    todo: ["todo", "in_progress", "reminder", "completed", "cancelled"],
    in_progress: ["in_progress", "todo", "review", "blocked", "reminder", "completed", "cancelled"],
    review: ["review", "in_progress", "blocked", "completed", "cancelled"],
    blocked: ["blocked", "todo", "in_progress", "completed", "cancelled"],
    completed: ["completed", "todo", "in_progress"],
    cancelled: ["cancelled", "todo"],
    reminder: ["reminder", "todo", "in_progress", "review", "blocked", "completed", "cancelled"]
  };
  return allowed[status].map((value) => ({ value, label: statusLabels[value] }));
}

function taskStatusAction(status: TaskStatus) {
  switch (status) {
    case "todo":
      return { label: "开始任务", nextStatus: "in_progress" as TaskStatus };
    case "completed":
    case "cancelled":
      return { label: "重新打开", nextStatus: "todo" as TaskStatus };
    default:
      return { label: "标记完成", nextStatus: "completed" as TaskStatus };
  }
}

function taskActivitySummary(activity: TaskActivity) {
  if (activity.action === "created") return "创建了任务";
  if (activity.action === "deleted") return "删除了任务";
  if (activity.action === "restored") return "恢复了任务";
  if (activity.action === "comment_added") return "发表了评论";
  if (activity.action === "comment_updated") return "编辑了评论";
  if (activity.action === "comment_deleted") return "删除了评论";
  if (activity.action === "status_changed") {
    const before = typeof activity.before.status === "string" ? statusLabels[activity.before.status as TaskStatus] ?? activity.before.status : "未知状态";
    const after = typeof activity.after.status === "string" ? statusLabels[activity.after.status as TaskStatus] ?? activity.after.status : "未知状态";
    return `将状态从${before}改为${after}`;
  }
  const fieldLabels: Record<string, string> = {
    title: "标题",
    assignee: "负责人",
    project: "项目",
    priority: "优先级",
    tags: "标签",
    due_at: "截止时间",
    progress: "进度"
  };
  const changed = Object.keys(activity.after).filter((key) => key !== "version" && key !== "source_type" && JSON.stringify(activity.before[key]) !== JSON.stringify(activity.after[key]));
  return changed.length > 0 ? `更新了${changed.map((key) => fieldLabels[key] ?? key).join("、")}` : "更新了任务";
}

function formatAttachmentSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.ceil(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function TasksPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const copilotPanel = useCopilotPanelVisibility();
  const taskGoalRef = useRef<HTMLTextAreaElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);
  const taskDeepLinkRef = useRef<number | null>(null);
  const deepLinkedTaskID = searchParams.get("task_id");
  const [apiTasks, setApiTasks] = useState<Task[]>([]);
  const [apiStats, setApiStats] = useState<TaskStats | null>(null);
  const [selectedStatus, setSelectedStatus] = useState<TaskStatus | undefined>(() => taskStatusFromQuery(searchParams.get("status")));
  const [selectedProject, setSelectedProject] = useState(() => searchParams.get("project") ?? "");
  const [selectedPriority, setSelectedPriority] = useState<TaskPriority | "">(() => taskPriorityFromQuery(searchParams.get("priority")));
  const [selectedTag, setSelectedTag] = useState(() => searchParams.get("tag") ?? "");
  const [selectedSort, setSelectedSort] = useState<TaskSort>(() => taskSortFromQuery(searchParams.get("sort")));
  const [selectedGroup, setSelectedGroup] = useState<TaskGroup | "">(() => taskGroupFromQuery(searchParams.get("group")));
  const [searchInput, setSearchInput] = useState(() => searchParams.get("q") ?? "");
  const [searchQuery, setSearchQuery] = useState(() => searchParams.get("q") ?? "");
  const [taskPage, setTaskPage] = useState(() => taskPageFromQuery(searchParams.get("page")));
  const [taskTotal, setTaskTotal] = useState(0);
  const [listRevision, setListRevision] = useState(0);
  const [projectOptions, setProjectOptions] = useState<string[]>([]);
  const [projectsLoaded, setProjectsLoaded] = useState(false);
  const [projectsLoading, setProjectsLoading] = useState(false);
  const [tagOptions, setTagOptions] = useState<string[]>([]);
  const [tagsLoaded, setTagsLoaded] = useState(false);
  const [tagsLoading, setTagsLoading] = useState(false);
  const [listLoading, setListLoading] = useState(false);
  const [activeView, setActiveView] = useState<TaskView>(() => taskViewFromQuery(searchParams.get("view")));
  const [calendarMonth, setCalendarMonth] = useState(() => calendarMonthFromQuery(searchParams.get("month")));
  const [calendarTasks, setCalendarTasks] = useState<Task[] | null>(null);
  const [calendarLoading, setCalendarLoading] = useState(false);
  const [listError, setListError] = useState("");
  const [savingTaskID, setSavingTaskID] = useState<number | null>(null);
  const [selectedTaskIDs, setSelectedTaskIDs] = useState<Set<number>>(() => new Set());
  const [batchStatus, setBatchStatus] = useState<TaskStatus>("completed");
  const [batchEditField, setBatchEditField] = useState<"assignee" | "priority" | "due_at" | "tags">("assignee");
  const [batchAssignee, setBatchAssignee] = useState("");
  const [batchPriority, setBatchPriority] = useState<TaskPriority>("medium");
  const [batchDueAt, setBatchDueAt] = useState("");
  const [batchTags, setBatchTags] = useState("");
  const [batchBusy, setBatchBusy] = useState(false);
  const [confirmBatchDelete, setConfirmBatchDelete] = useState(false);
  const [batchMessage, setBatchMessage] = useState("");
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
  const [subtasks, setSubtasks] = useState<TaskSubtask[]>([]);
  const [subtasksLoading, setSubtasksLoading] = useState(false);
  const [subtaskCreating, setSubtaskCreating] = useState(false);
  const [subtaskBusyID, setSubtaskBusyID] = useState<number | null>(null);
  const [subtaskTitle, setSubtaskTitle] = useState("");
  const [subtaskError, setSubtaskError] = useState("");
  const [subtaskParentID, setSubtaskParentID] = useState<number | null>(null);
  const [collapsedSubtaskIDs, setCollapsedSubtaskIDs] = useState<Set<number>>(new Set());
  const [confirmSubtaskDeleteID, setConfirmSubtaskDeleteID] = useState<number | null>(null);
  const [taskReminder, setTaskReminder] = useState<TaskReminder | null>(null);
  const [reminderLoading, setReminderLoading] = useState(false);
  const [reminderSaving, setReminderSaving] = useState(false);
  const [reminderTime, setReminderTime] = useState("");
  const [reminderRecurrence, setReminderRecurrence] = useState<ReminderRecurrence>("once");
  const [reminderError, setReminderError] = useState("");
  const [reminderMessage, setReminderMessage] = useState("");
  const [reminderNeedsUpgrade, setReminderNeedsUpgrade] = useState(false);
  const [deletedTasks, setDeletedTasks] = useState<Task[]>([]);
  const [deletedTasksMessage, setDeletedTasksMessage] = useState("");
  const [restoringTasks, setRestoringTasks] = useState(false);
  const [aiDraft, setAIDraft] = useState<TaskAIDraft | null>(null);
  const [aiDraftItems, setAIDraftItems] = useState<AIDraftPreviewItem[]>([]);
  const [aiDraftBusy, setAIDraftBusy] = useState(false);
  const [aiDraftError, setAIDraftError] = useState("");
  const [taskActivities, setTaskActivities] = useState<TaskActivity[]>([]);
  const [activityTotal, setActivityTotal] = useState(0);
  const [activitiesLoading, setActivitiesLoading] = useState(false);
  const [activitiesError, setActivitiesError] = useState("");
  const [taskComments, setTaskComments] = useState<TaskComment[]>([]);
  const [commentTotal, setCommentTotal] = useState(0);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [commentSavingID, setCommentSavingID] = useState<number | "new" | null>(null);
  const [commentContent, setCommentContent] = useState("");
  const [commentEditingID, setCommentEditingID] = useState<number | null>(null);
  const [commentEditingContent, setCommentEditingContent] = useState("");
  const [commentsError, setCommentsError] = useState("");
  const [taskAttachments, setTaskAttachments] = useState<TaskAttachment[]>([]);
  const [attachmentsLoading, setAttachmentsLoading] = useState(false);
  const [attachmentBusyID, setAttachmentBusyID] = useState<number | "upload" | null>(null);
  const [attachmentsError, setAttachmentsError] = useState("");
  const [draggingTaskID, setDraggingTaskID] = useState<number | null>(null);
  const [boardUpdatingTaskID, setBoardUpdatingTaskID] = useState<number | null>(null);
  const [taskColumns, setTaskColumns] = useState<TaskListColumn[]>(defaultTaskColumns);
  const [columnConfigOpen, setColumnConfigOpen] = useState(false);
  const [columnsLoading, setColumnsLoading] = useState(true);
  const [columnsSaving, setColumnsSaving] = useState(false);
  const [columnsMessage, setColumnsMessage] = useState("");
  const [columnsError, setColumnsError] = useState("");

  useEffect(() => {
    let active = true;
    setListLoading(true);
    tasksApi
      .listTasks({
        status: selectedStatus,
        project: selectedProject || undefined,
        priority: selectedPriority || undefined,
        tag: selectedTag || undefined,
        q: searchQuery || undefined,
        sort: selectedSort !== "created_at" ? selectedSort : undefined,
        group: selectedGroup || undefined,
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
  }, [selectedStatus, selectedProject, selectedPriority, selectedTag, selectedSort, selectedGroup, searchQuery, taskPage, listRevision]);

  useEffect(() => {
    if (activeView !== "calendar") return;
    let active = true;
    const range = calendarMonthRange(calendarMonth);
    setCalendarLoading(true);
    tasksApi.calendar({
      ...range,
      status: selectedStatus,
      project: selectedProject || undefined,
      priority: selectedPriority || undefined,
      tag: selectedTag || undefined,
      q: searchQuery || undefined
    }).then((payload) => {
      if (!active) return;
      setCalendarTasks([...payload.tasks, ...payload.unscheduled]);
    }).catch(() => {
      if (active) setCalendarTasks(null);
    }).finally(() => {
      if (active) setCalendarLoading(false);
    });
    return () => {
      active = false;
    };
  }, [activeView, calendarMonth, selectedStatus, selectedProject, selectedPriority, selectedTag, searchQuery, listRevision]);

  useEffect(() => {
    const next = new URLSearchParams();
    if (activeView !== "list") next.set("view", activeView);
    if (activeView === "calendar") next.set("month", calendarMonth);
    if (selectedStatus) next.set("status", selectedStatus);
    if (selectedProject) next.set("project", selectedProject);
    if (selectedPriority) next.set("priority", selectedPriority);
    if (selectedTag) next.set("tag", selectedTag);
    if (searchQuery) next.set("q", searchQuery);
    if (selectedSort !== "created_at") next.set("sort", selectedSort);
    if (selectedGroup) next.set("group", selectedGroup);
    if (taskPage > 1) next.set("page", String(taskPage));
    if (deepLinkedTaskID) next.set("task_id", deepLinkedTaskID);
    setSearchParams(next, { replace: true });
  }, [activeView, calendarMonth, deepLinkedTaskID, searchQuery, selectedGroup, selectedPriority, selectedProject, selectedSort, selectedStatus, selectedTag, setSearchParams, taskPage]);

  useEffect(() => {
    const taskID = Number(deepLinkedTaskID);
    if (!Number.isInteger(taskID) || taskID <= 0 || taskDeepLinkRef.current === taskID) return;
    taskDeepLinkRef.current = taskID;
    void openTaskDetail(taskID);
    // The ref prevents repeated loads when filters rewrite the URL around the same task deep link.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deepLinkedTaskID]);

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

  useEffect(() => {
    let active = true;
    tasksApi.getViewPreference().then((preference) => {
      if (!active) return;
      if (validTaskColumns(preference.columns)) setTaskColumns(preference.columns);
      setColumnsError("");
    }).catch(() => {
      if (active) setColumnsError("字段配置暂时无法同步，当前使用默认字段");
    }).finally(() => {
      if (active) setColumnsLoading(false);
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
  const populatedBoardColumns = boardColumns.filter(([, items]) => items.length > 0);
  const showBoardPreview = activeView === "list" && populatedBoardColumns.length > 0;
  const useFullWidthWorkbench = !showBoardPreview && copilotPanel?.isPanelOpen === false;
  const calendarSourceTasks = calendarTasks ?? apiTasks;
  const scheduledTasks = calendarSourceTasks
    .filter((task) => task.due_at)
    .sort((left, right) => new Date(left.due_at as string).getTime() - new Date(right.due_at as string).getTime());
  const calendarGroups = Array.from(scheduledTasks.reduce<Map<string, Task[]>>((groups, task) => {
    const label = formatCalendarDate(task.due_at as string);
    groups.set(label, [...(groups.get(label) ?? []), task]);
    return groups;
  }, new Map()));
  const unscheduledTasks = calendarSourceTasks.filter((task) => !task.due_at);
  const listGroups = selectedGroup
    ? Array.from(apiTasks.reduce<Map<string, TaskRow[]>>((groups, task) => {
      const label = taskGroupLabel(task, selectedGroup);
      groups.set(label, [...(groups.get(label) ?? []), toTaskRow(task)]);
      return groups;
    }, new Map()))
    : [["", visibleTasks] as [string, TaskRow[]]];
  const totalPages = Math.max(1, Math.ceil(taskTotal / taskPageSize));
  const hasActiveFilters = Boolean(selectedStatus || selectedProject || selectedPriority || selectedTag || searchQuery || selectedGroup || selectedSort !== "created_at");
  const currentPageTaskIDs = apiTasks.map((task) => task.id);
  const allCurrentPageSelected = currentPageTaskIDs.length > 0 && currentPageTaskIDs.every((id) => selectedTaskIDs.has(id));
  const selectedTasks = apiTasks.filter((task) => selectedTaskIDs.has(task.id));
  const subtaskTreeRows = flattenSubtaskTree(subtasks, collapsedSubtaskIDs);
  const subtaskParent = subtasks.find((item) => item.id === subtaskParentID);
  const orderedColumnDefinitions = [
    ...taskColumns.map((key) => taskColumnDefinitions.find((column) => column.key === key)).filter(Boolean),
    ...taskColumnDefinitions.filter((column) => !taskColumns.includes(column.key))
  ] as Array<{ key: TaskListColumn; label: string; required?: boolean }>;
  const batchStatusOptions = boardColumnLabels.filter((option) =>
    selectedTasks.length > 0 && selectedTasks.every((task) => taskStatusOptions(task.status).some((item) => item.value === option.status))
  );
  const selectedBatchStatus = batchStatusOptions.some((option) => option.status === batchStatus)
    ? batchStatus
    : (batchStatusOptions[0]?.status ?? "completed");

  useEffect(() => {
    if (taskTotal > 0 && taskPage > totalPages) setTaskPage(totalPages);
  }, [taskPage, taskTotal, totalPages]);

  useEffect(() => {
    setSelectedTaskIDs(new Set());
    setConfirmBatchDelete(false);
  }, [selectedStatus, selectedProject, selectedPriority, selectedTag, selectedSort, selectedGroup, searchQuery, taskPage]);

  useEffect(() => {
    const validIDs = new Set(apiTasks.map((task) => task.id));
    setSelectedTaskIDs((current) => {
      const next = new Set(Array.from(current).filter((id) => validIDs.has(id)));
      return next.size === current.size ? current : next;
    });
  }, [apiTasks]);

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

  async function loadTagOptions() {
    if (tagsLoaded || tagsLoading) return;
    setTagsLoading(true);
    try {
      const payload = await tasksApi.listTags();
      setTagOptions(payload.tags);
      setTagsLoaded(true);
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法读取标签选项"));
    } finally {
      setTagsLoading(false);
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
    setSelectedTag("");
    setSelectedSort("created_at");
    setSelectedGroup("");
    setTaskPage(1);
  }

  function taskMatchesCurrentFilters(task: Task) {
    const normalizedQuery = searchQuery.toLowerCase();
    return (!selectedStatus || task.status === selectedStatus) &&
      (!selectedProject || task.project === selectedProject) &&
      (!selectedPriority || task.priority === selectedPriority) &&
      (!selectedTag || (task.tags ?? []).includes(selectedTag)) &&
      (!normalizedQuery || [task.title, task.description ?? "", task.assignee ?? "", task.project, ...(task.tags ?? []), task.learning].some((value) => value.toLowerCase().includes(normalizedQuery)));
  }

  async function updateTaskStatus(taskID: number, currentStatus: TaskStatus, version: number) {
    const action = taskStatusAction(currentStatus);
    setSavingTaskID(taskID);
    try {
      const updated = await tasksApi.updateTask(taskID, { status: action.nextStatus, version });
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

  async function moveTaskOnBoard(taskID: number, targetStatus: TaskStatus) {
    if (boardUpdatingTaskID !== null) return;
    const original = apiTasks.find((task) => task.id === taskID);
    if (!original || original.status === targetStatus) {
      setDraggingTaskID(null);
      return;
    }
    setDraggingTaskID(null);
    setBoardUpdatingTaskID(taskID);
    setListError("");
    setApiTasks((current) => current.map((task) => task.id === taskID ? { ...task, status: targetStatus } : task));
    try {
      const updated = await tasksApi.updateTask(taskID, { status: targetStatus, version: original.version });
      const remainsVisible = taskMatchesCurrentFilters(updated);
      setApiTasks((current) => !remainsVisible
        ? current.filter((task) => task.id !== taskID)
        : current.map((task) => task.id === taskID ? updated : task));
      if (!remainsVisible) setTaskTotal((current) => Math.max(0, current - 1));
      await refreshTaskStats();
    } catch (error) {
      setApiTasks((current) => current.map((task) => task.id === taskID ? original : task));
      setListError(apiErrorMessage(error, "任务状态更新失败，卡片已回到原位置"));
    } finally {
      setBoardUpdatingTaskID(null);
    }
  }

  function toggleTaskSelection(taskID: number) {
    setConfirmBatchDelete(false);
    setSelectedTaskIDs((current) => {
      const next = new Set(current);
      if (next.has(taskID)) next.delete(taskID);
      else next.add(taskID);
      return next;
    });
  }

  function toggleCurrentPageSelection() {
    setConfirmBatchDelete(false);
    setSelectedTaskIDs((current) => {
      const next = new Set(current);
      if (allCurrentPageSelected) currentPageTaskIDs.forEach((id) => next.delete(id));
      else currentPageTaskIDs.forEach((id) => next.add(id));
      return next;
    });
  }

  async function applyBatchStatus() {
    const ids = currentPageTaskIDs.filter((id) => selectedTaskIDs.has(id));
    if (ids.length === 0 || batchBusy) return;
    setBatchBusy(true);
    setBatchMessage("");
    try {
      await tasksApi.batchUpdateStatus(ids, selectedBatchStatus);
      const selectedIDs = new Set(ids);
      const removedCount = apiTasks.filter((task) => selectedIDs.has(task.id) && !taskMatchesCurrentFilters({ ...task, status: selectedBatchStatus })).length;
      setApiTasks((current) => current
        .map((task) => selectedIDs.has(task.id) ? { ...task, status: selectedBatchStatus } : task)
        .filter(taskMatchesCurrentFilters));
      if (removedCount > 0) setTaskTotal((current) => Math.max(0, current - removedCount));
      setListRevision((current) => current + 1);
      setSelectedTaskIDs(new Set());
      setConfirmBatchDelete(false);
      setListError("");
      setBatchMessage(`已更新 ${ids.length} 条任务状态`);
      await refreshTaskStats();
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法批量更新任务状态"));
    } finally {
      setBatchBusy(false);
    }
  }

  async function applyBatchField(clearDueAt = false) {
    const ids = currentPageTaskIDs.filter((id) => selectedTaskIDs.has(id));
    if (ids.length === 0 || batchBusy) return;
    const input = batchEditField === "assignee"
      ? { ids, assignee: batchAssignee.trim() }
      : batchEditField === "priority"
        ? { ids, priority: batchPriority }
        : batchEditField === "due_at"
          ? { ids, ...(clearDueAt ? { clearDueAt: true } : { dueAt: batchDueAt ? new Date(batchDueAt).toISOString() : undefined }) }
          : { ids, tags: batchTags.split(/[,，]/).map((tag) => tag.trim()).filter(Boolean) };
    if (batchEditField === "due_at" && !clearDueAt && !batchDueAt) {
      setListError("请选择截止时间，或使用清除截止时间");
      return;
    }
    setBatchBusy(true);
    setBatchMessage("");
    setListError("");
    try {
      const result = await tasksApi.batchUpdate(input);
      setListRevision((current) => current + 1);
      setSelectedTaskIDs(new Set());
      setConfirmBatchDelete(false);
      setBatchMessage(result.failed.length > 0
        ? `已更新 ${result.updated} 条，${result.failed.length} 条失败`
        : `已更新 ${result.updated} 条任务`);
      await refreshTaskStats();
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法批量更新任务"));
    } finally {
      setBatchBusy(false);
    }
  }

  async function deleteSelectedTasks() {
    const ids = currentPageTaskIDs.filter((id) => selectedTaskIDs.has(id));
    if (ids.length === 0 || batchBusy) return;
    if (!confirmBatchDelete) {
      setConfirmBatchDelete(true);
      return;
    }
    setBatchBusy(true);
    setBatchMessage("");
    try {
      const deleted = apiTasks.filter((task) => ids.includes(task.id));
      await tasksApi.batchDelete(ids);
      setDeletedTasks(deleted);
      setDeletedTasksMessage(`已删除 ${ids.length} 条任务`);
      const selectedIDs = new Set(ids);
      setApiTasks((current) => current.filter((task) => !selectedIDs.has(task.id)));
      setTaskTotal((current) => Math.max(0, current - ids.length));
      setListRevision((current) => current + 1);
      setSelectedTaskIDs(new Set());
      setConfirmBatchDelete(false);
      setListError("");
      setBatchMessage("");
      await refreshTaskStats();
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法批量删除任务"));
    } finally {
      setBatchBusy(false);
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
    setDeletedTasks([]);
    setDeletedTasksMessage("");
    try {
      const result = await tasksApi.generateTasks(title);
      setAIDraft(result.draft);
      setAIDraftItems(result.tasks.map(toAIDraftPreviewItem));
      setAIDraftError("");
      setListError("");
      setCreateMessage("");
    } catch (error) {
      setCreateError(apiErrorMessage(error, "暂时无法生成任务"));
    } finally {
      setIsCreatingTask(false);
    }
  }

  function updateAIDraftItem<Key extends keyof AIDraftPreviewItem>(draftIndex: number, key: Key, value: AIDraftPreviewItem[Key]) {
    setAIDraftItems((current) => current.map((item) => item.draftIndex === draftIndex ? { ...item, [key]: value } : item));
  }

  async function adoptAIDraftTasks() {
    if (!aiDraft || aiDraftBusy) return;
    const selected = aiDraftItems.filter((item) => item.selected);
    if (selected.length === 0) {
      setAIDraftError("请至少选择一条任务建议");
      return;
    }
    if (selected.some((item) => !item.title.trim() || !item.project.trim())) {
      setAIDraftError("选中任务的标题和所属项目不能为空");
      return;
    }
    setAIDraftBusy(true);
    setAIDraftError("");
    try {
      const result = await tasksApi.adoptTaskAIDraft(aiDraft.id, selected.map((item) => ({
        draftIndex: item.draftIndex,
        title: item.title.trim(),
        description: item.description.trim(),
        assignee: item.assignee.trim(),
        project: item.project.trim(),
        priority: item.priority,
        tags: parseList(item.tags),
        dueAt: item.dueAt ? new Date(item.dueAt).toISOString() : undefined,
        tools: parseList(item.tools),
        learning: item.learning.trim()
      })), `task-ai-draft-${aiDraft.id}`);
      setAIDraft(null);
      setAIDraftItems([]);
      setTaskGoal("");
      setCreateMessage(`已采纳 ${result.tasks.length} 条任务`);
      setListRevision((current) => current + 1);
      await refreshTaskStats();
    } catch (error) {
      setAIDraftError(apiErrorMessage(error, "暂时无法采纳任务草稿"));
    } finally {
      setAIDraftBusy(false);
    }
  }

  function closeAIDraftPreview() {
    if (aiDraftBusy) return;
    setAIDraft(null);
    setAIDraftItems([]);
    setAIDraftError("");
  }

  function regenerateAIDraft() {
    if (aiDraftBusy || isCreatingTask) return;
    if (!window.confirm("重新生成会覆盖当前草稿中的选择和修改，是否继续？")) return;
    closeAIDraftPreview();
    void createTaskFromGoal();
  }

  async function openTaskDetail(taskID: number) {
    setDetailTaskID(taskID);
    setDetailTask(null);
    setDetailForm(null);
    setDetailLoading(true);
    setDetailError("");
    setConfirmDelete(false);
    setSubtasks([]);
    setSubtasksLoading(true);
    setSubtaskTitle("");
    setSubtaskError("");
    setSubtaskParentID(null);
    setCollapsedSubtaskIDs(new Set());
    setConfirmSubtaskDeleteID(null);
    setTaskReminder(null);
    setReminderLoading(true);
    setReminderTime("");
    setReminderRecurrence("once");
    setReminderError("");
    setReminderMessage("");
    setReminderNeedsUpgrade(false);
    setTaskActivities([]);
    setActivityTotal(0);
    setActivitiesLoading(true);
    setActivitiesError("");
    setTaskComments([]);
    setCommentTotal(0);
    setCommentContent("");
    setCommentEditingID(null);
    setCommentEditingContent("");
    setCommentsError("");
    setCommentsLoading(true);
    setTaskAttachments([]);
    setAttachmentsLoading(true);
    setAttachmentBusyID(null);
    setAttachmentsError("");
    void loadTaskComments(taskID);
    const [taskResult, subtasksResult, reminderResult, activitiesResult, attachmentsResult] = await Promise.allSettled([
      tasksApi.getTask(taskID),
      tasksApi.listSubtasks(taskID),
      tasksApi.getReminder(taskID),
      tasksApi.listTaskActivities(taskID),
      tasksApi.listTaskAttachments(taskID)
    ]);
    if (taskResult.status === "fulfilled") {
      setDetailTask(taskResult.value);
      setDetailForm(toTaskEditForm(taskResult.value));
    } else {
      setDetailError(apiErrorMessage(taskResult.reason, "暂时无法读取任务详情"));
    }
    if (subtasksResult.status === "fulfilled") {
      setSubtasks(subtasksResult.value.subtasks);
    } else {
      setSubtaskError(apiErrorMessage(subtasksResult.reason, "暂时无法读取子任务"));
    }
    if (reminderResult.status === "fulfilled") {
      setTaskReminder(reminderResult.value.reminder);
      setReminderTime(toDateTimeLocal(reminderResult.value.reminder?.remind_at));
      setReminderRecurrence(reminderResult.value.reminder?.recurrence ?? "once");
    } else {
      setReminderError(apiErrorMessage(reminderResult.reason, "暂时无法读取提醒设置"));
    }
    if (activitiesResult.status === "fulfilled") {
      setTaskActivities(activitiesResult.value.activities);
      setActivityTotal(activitiesResult.value.total);
    } else {
      setActivitiesError(apiErrorMessage(activitiesResult.reason, "暂时无法读取操作记录"));
    }
    if (attachmentsResult.status === "fulfilled") {
      setTaskAttachments(attachmentsResult.value.attachments);
    } else {
      setAttachmentsError(apiErrorMessage(attachmentsResult.reason, "暂时无法读取附件"));
    }
    setDetailLoading(false);
    setSubtasksLoading(false);
    setReminderLoading(false);
    setActivitiesLoading(false);
    setAttachmentsLoading(false);
  }

  function closeTaskDetail() {
    if (detailSaving || detailDeleting || reminderSaving || attachmentBusyID !== null) return;
    setDetailTaskID(null);
    setDetailTask(null);
    setDetailForm(null);
    setDetailError("");
    setConfirmDelete(false);
    setSubtasks([]);
    setSubtaskTitle("");
    setSubtaskError("");
    setSubtaskParentID(null);
    setCollapsedSubtaskIDs(new Set());
    setConfirmSubtaskDeleteID(null);
    setTaskReminder(null);
    setReminderTime("");
    setReminderRecurrence("once");
    setReminderError("");
    setReminderMessage("");
    setReminderNeedsUpgrade(false);
    setTaskActivities([]);
    setActivityTotal(0);
    setActivitiesError("");
    setActivitiesLoading(false);
    setTaskComments([]);
    setCommentTotal(0);
    setCommentContent("");
    setCommentEditingID(null);
    setCommentEditingContent("");
    setCommentsError("");
    setCommentsLoading(false);
    setTaskAttachments([]);
    setAttachmentsLoading(false);
    setAttachmentBusyID(null);
    setAttachmentsError("");
  }

  function updateDetailField<Key extends keyof TaskEditForm>(key: Key, value: TaskEditForm[Key]) {
    setDetailForm((current) => current ? { ...current, [key]: value } : current);
  }

  async function createDetailSubtask() {
    if (!detailTaskID || subtaskCreating) return;
    const title = subtaskTitle.trim();
    if (!title) {
      setSubtaskError("请输入子任务标题");
      return;
    }
    setSubtaskCreating(true);
    setSubtaskError("");
    try {
      const item = await tasksApi.createSubtask(detailTaskID, {
        title,
        parentSubtaskId: subtaskParentID ?? undefined
      });
      setSubtasks((current) => [...current, item]);
      setSubtaskTitle("");
      setSubtaskParentID(null);
    } catch (error) {
      setSubtaskError(apiErrorMessage(error, "暂时无法添加子任务"));
    } finally {
      setSubtaskCreating(false);
    }
  }

  async function toggleDetailSubtask(item: TaskSubtask) {
    if (!detailTaskID || subtaskBusyID !== null) return;
    setSubtaskBusyID(item.id);
    setSubtaskError("");
    try {
      const updated = await tasksApi.updateSubtask(detailTaskID, item.id, { completed: !item.completed });
      setSubtasks((current) => current.map((subtask) => subtask.id === updated.id ? updated : subtask));
    } catch (error) {
      setSubtaskError(apiErrorMessage(error, "暂时无法更新子任务"));
    } finally {
      setSubtaskBusyID(null);
    }
  }

  async function deleteDetailSubtask(item: TaskSubtask) {
    if (!detailTaskID || subtaskBusyID !== null) return;
    setSubtaskBusyID(item.id);
    setSubtaskError("");
    try {
      await tasksApi.deleteSubtask(detailTaskID, item.id);
      setSubtasks((current) => {
        const removedIDs = new Set([item.id, ...subtaskDescendantIDs(current, item.id)]);
        return current.filter((subtask) => !removedIDs.has(subtask.id));
      });
      setConfirmSubtaskDeleteID(null);
      if (subtaskParentID === item.id) setSubtaskParentID(null);
    } catch (error) {
      setSubtaskError(apiErrorMessage(error, "暂时无法删除子任务"));
    } finally {
      setSubtaskBusyID(null);
    }
  }

  async function saveTaskReminder() {
    if (!detailTaskID || reminderSaving) return;
    const remindAt = new Date(reminderTime);
    if (!reminderTime || Number.isNaN(remindAt.getTime()) || remindAt.getTime() <= Date.now()) {
      setReminderError("请选择未来的提醒时间");
      return;
    }
    setReminderSaving(true);
    setReminderError("");
    setReminderMessage("");
    setReminderNeedsUpgrade(false);
    try {
      const reminder = await tasksApi.upsertReminder(detailTaskID, {
        remindAt: remindAt.toISOString(),
        recurrence: reminderRecurrence
      });
      setTaskReminder(reminder);
      setReminderTime(toDateTimeLocal(reminder.remind_at));
      setReminderRecurrence(reminder.recurrence);
      setReminderMessage("已设置站内提醒");
    } catch (error) {
      setReminderError(apiErrorMessage(error, "暂时无法保存提醒"));
      setReminderNeedsUpgrade(error instanceof ApiRequestError && error.code === "membership_required");
    } finally {
      setReminderSaving(false);
    }
  }

  async function cancelTaskReminder() {
    if (!detailTaskID || !taskReminder || reminderSaving) return;
    setReminderSaving(true);
    setReminderError("");
    setReminderMessage("");
    setReminderNeedsUpgrade(false);
    try {
      await tasksApi.deleteReminder(detailTaskID);
      setTaskReminder(null);
      setReminderTime("");
      setReminderRecurrence("once");
    } catch (error) {
      setReminderError(apiErrorMessage(error, "暂时无法取消提醒"));
    } finally {
      setReminderSaving(false);
    }
  }

  async function loadMoreTaskActivities() {
    if (!detailTaskID || activitiesLoading || taskActivities.length >= activityTotal) return;
    setActivitiesLoading(true);
    setActivitiesError("");
    try {
      const result = await tasksApi.listTaskActivities(detailTaskID, 20, taskActivities.length);
      setTaskActivities((current) => [...current, ...result.activities]);
      setActivityTotal(result.total);
    } catch (error) {
      setActivitiesError(apiErrorMessage(error, "暂时无法读取更多操作记录"));
    } finally {
      setActivitiesLoading(false);
    }
  }

  async function loadTaskComments(taskID: number) {
    try {
      const result = await tasksApi.listTaskComments(taskID);
      setTaskComments(result.comments);
      setCommentTotal(result.total);
      setCommentsError("");
    } catch (error) {
      setCommentsError(apiErrorMessage(error, "暂时无法读取评论"));
    } finally {
      setCommentsLoading(false);
    }
  }

  async function createTaskComment() {
    if (!detailTaskID || commentSavingID !== null) return;
    const content = commentContent.trim();
    if (!content) {
      setCommentsError("请输入评论内容");
      return;
    }
    setCommentSavingID("new");
    setCommentsError("");
    try {
      const comment = await tasksApi.createTaskComment(detailTaskID, content);
      setTaskComments((current) => [comment, ...current]);
      setCommentTotal((current) => current + 1);
      setCommentContent("");
    } catch (error) {
      setCommentsError(apiErrorMessage(error, "暂时无法发表评论"));
    } finally {
      setCommentSavingID(null);
    }
  }

  async function saveTaskComment(commentID: number) {
    if (!detailTaskID || commentSavingID !== null) return;
    const content = commentEditingContent.trim();
    if (!content) {
      setCommentsError("评论内容不能为空");
      return;
    }
    setCommentSavingID(commentID);
    setCommentsError("");
    try {
      const comment = await tasksApi.updateTaskComment(detailTaskID, commentID, content);
      setTaskComments((current) => current.map((item) => item.id === comment.id ? comment : item));
      setCommentEditingID(null);
      setCommentEditingContent("");
    } catch (error) {
      setCommentsError(apiErrorMessage(error, "暂时无法保存评论"));
    } finally {
      setCommentSavingID(null);
    }
  }

  async function deleteTaskComment(commentID: number) {
    if (!detailTaskID || commentSavingID !== null || !window.confirm("确定删除这条评论吗？")) return;
    setCommentSavingID(commentID);
    setCommentsError("");
    try {
      await tasksApi.deleteTaskComment(detailTaskID, commentID);
      setTaskComments((current) => current.filter((item) => item.id !== commentID));
      setCommentTotal((current) => Math.max(0, current - 1));
    } catch (error) {
      setCommentsError(apiErrorMessage(error, "暂时无法删除评论"));
    } finally {
      setCommentSavingID(null);
    }
  }

  async function uploadTaskAttachment(file: File) {
    if (!detailTaskID || attachmentBusyID !== null) return;
    if (file.size <= 0 || file.size > 10 * 1024 * 1024) {
      setAttachmentsError("单个附件不能超过 10MB");
      return;
    }
    setAttachmentBusyID("upload");
    setAttachmentsError("");
    try {
      const attachment = await tasksApi.uploadTaskAttachment(detailTaskID, file);
      setTaskAttachments((current) => [attachment, ...current]);
    } catch (error) {
      setAttachmentsError(apiErrorMessage(error, "暂时无法上传附件"));
    } finally {
      setAttachmentBusyID(null);
      if (attachmentInputRef.current) attachmentInputRef.current.value = "";
    }
  }

  async function downloadTaskAttachment(attachment: TaskAttachment) {
    if (!detailTaskID || attachmentBusyID !== null) return;
    setAttachmentBusyID(attachment.id);
    setAttachmentsError("");
    try {
      await tasksApi.downloadTaskAttachment(detailTaskID, attachment.id, attachment.name);
    } catch (error) {
      setAttachmentsError(apiErrorMessage(error, "暂时无法下载附件"));
    } finally {
      setAttachmentBusyID(null);
    }
  }

  async function deleteTaskAttachment(attachment: TaskAttachment) {
    if (!detailTaskID || attachmentBusyID !== null || !window.confirm(`确定删除附件“${attachment.name}”吗？`)) return;
    setAttachmentBusyID(attachment.id);
    setAttachmentsError("");
    try {
      await tasksApi.deleteTaskAttachment(detailTaskID, attachment.id);
      setTaskAttachments((current) => current.filter((item) => item.id !== attachment.id));
    } catch (error) {
      setAttachmentsError(apiErrorMessage(error, "暂时无法删除附件"));
    } finally {
      setAttachmentBusyID(null);
    }
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
      const tags = parseList(detailForm.tags);
      const updated = await tasksApi.updateTask(detailTask.id, {
        title,
        description: detailForm.description !== (detailTask.description ?? "") ? detailForm.description : undefined,
        assignee: detailForm.assignee !== (detailTask.assignee ?? "") ? detailForm.assignee : undefined,
        project,
        status: detailForm.status,
        priority: detailForm.priority,
        tags: JSON.stringify(tags) !== JSON.stringify(detailTask.tags ?? []) ? tags : undefined,
        dueAt: detailForm.dueAt ? new Date(detailForm.dueAt).toISOString() : undefined,
        clearDueAt: !detailForm.dueAt && Boolean(detailTask.due_at) ? true : undefined,
        tools: parseList(detailForm.tools),
        learning: detailForm.learning.trim(),
        progress: detailForm.progress !== (detailTask.progress ?? 0) ? detailForm.progress : undefined,
        version: detailTask.version
      });
      const remainsVisible = taskMatchesCurrentFilters(updated);
      setApiTasks((current) => !remainsVisible
        ? current.filter((task) => task.id !== updated.id)
        : current.map((task) => task.id === updated.id ? updated : task));
      if (!remainsVisible) setTaskTotal((current) => Math.max(0, current - 1));
      await refreshTaskStats();
      setCreateMessage(`已更新任务：${updated.title}`);
      setDeletedTasks([]);
      setDeletedTasksMessage("");
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
      const deleted = detailTask;
      await tasksApi.deleteTask(detailTask.id);
      setDeletedTasks([deleted]);
      setDeletedTasksMessage(`已删除任务：${deleted.title}`);
      setApiTasks((current) => current.filter((task) => task.id !== detailTask.id));
      setTaskTotal((current) => Math.max(0, current - 1));
      await refreshTaskStats();
      setCreateMessage("");
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

  async function restoreDeletedTasks() {
    if (deletedTasks.length === 0 || restoringTasks) return;
    const tasksToRestore = deletedTasks;
    setRestoringTasks(true);
    setListError("");
    try {
      await Promise.all(tasksToRestore.map((task) => tasksApi.restoreTask(task.id)));
      setDeletedTasks([]);
      setDeletedTasksMessage("");
      setCreateMessage(`已恢复 ${tasksToRestore.length} 条任务`);
      setListRevision((current) => current + 1);
      await refreshTaskStats();
    } catch (error) {
      setListError(apiErrorMessage(error, "暂时无法恢复任务"));
    } finally {
      setRestoringTasks(false);
    }
  }

  function toggleTaskColumn(column: TaskListColumn) {
    if (column === "title") return;
    setTaskColumns((current) => current.includes(column)
      ? current.filter((key) => key !== column)
      : [...current, column]);
    setColumnsMessage("");
  }

  function moveTaskColumn(column: TaskListColumn, direction: -1 | 1) {
    setTaskColumns((current) => {
      const index = current.indexOf(column);
      const nextIndex = index + direction;
      if (index < 0 || nextIndex < 0 || nextIndex >= current.length) return current;
      const next = [...current];
      [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
      return next;
    });
    setColumnsMessage("");
  }

  async function saveTaskColumns() {
    if (columnsSaving) return;
    setColumnsSaving(true);
    setColumnsError("");
    setColumnsMessage("");
    try {
      const preference = await tasksApi.saveViewPreference(taskColumns);
      if (validTaskColumns(preference.columns)) setTaskColumns(preference.columns);
      setColumnsMessage("字段配置已保存");
    } catch (error) {
      setColumnsError(apiErrorMessage(error, "暂时无法保存字段配置"));
    } finally {
      setColumnsSaving(false);
    }
  }

  function renderTaskListColumn(task: TaskRow, column: TaskListColumn) {
    switch (column) {
      case "title":
        return (
          <div className="task-list-title-value">
            <span className={`task-status-dot ${task.status === "已完成" ? "done" : task.status === "进行中" ? "doing" : ""}`} aria-hidden="true" />
            <h3 title={task.title}>{task.title}</h3>
            <div className="task-list-title-links">
              <Link to="/tools">建议工具：{task.tools.join(" / ") || "待补充"}</Link>
              <Link to="/learning">补课：{task.learning || "待补充"}</Link>
            </div>
          </div>
        );
      case "project":
        return <span>{task.project || "未关联"}</span>;
      case "assignee":
        return <span>{task.assignee || "未指定"}</span>;
      case "due_at":
        return <span>{task.due}</span>;
      case "priority":
        return <span className={`task-priority ${task.priority === "高" ? "high" : task.priority === "中" ? "mid" : ""}`}>{task.priority}</span>;
      case "status":
        return <span className="task-state">{task.status}</span>;
      case "tags":
        return task.tags.length ? <div aria-label="任务标签" className="task-label-list">{task.tags.map((tag) => <span key={tag}>{tag}</span>)}</div> : <span>无标签</span>;
      case "progress":
        return <span>{task.progress}%</span>;
      case "source":
        return task.sourceTitle && task.sourceURL ? <Link className="task-source-link" to={task.sourceURL}>来源：{task.sourceTitle}</Link> : <span>{task.sourceTitle || "手工创建"}</span>;
      case "created_at":
        return <span>{task.createdAt}</span>;
      case "updated_at":
        return <span>{task.updatedAt}</span>;
    }
  }

  function renderTaskCard(task: Task, enableDrag = false) {
    const action = taskStatusAction(task.status);
    return (
      <article
        aria-grabbed={enableDrag ? draggingTaskID === task.id : undefined}
        className={`task-view-card${draggingTaskID === task.id ? " dragging" : ""}${boardUpdatingTaskID === task.id ? " updating" : ""}`}
        draggable={enableDrag && boardUpdatingTaskID === null}
        key={task.id}
        onDragEnd={enableDrag ? () => setDraggingTaskID(null) : undefined}
        onDragStart={enableDrag ? (event) => {
          event.dataTransfer.effectAllowed = "move";
          event.dataTransfer.setData("text/plain", String(task.id));
          setDraggingTaskID(task.id);
        } : undefined}
      >
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
          <span>进度 {task.progress ?? 0}%</span>
        </div>
        {task.tags?.length ? (
          <div aria-label="任务标签" className="task-label-list">
            {task.tags.map((tag) => <span key={tag}>{tag}</span>)}
          </div>
        ) : null}
        <div className="task-view-card-links">
          <Link to="/tools">工具：{task.tools.join(" / ")}</Link>
          <Link to="/learning">补课：{task.learning}</Link>
        </div>
        <div className="task-view-card-actions">
          <button
            disabled={savingTaskID === task.id || boardUpdatingTaskID === task.id}
            onClick={() => void updateTaskStatus(task.id, task.status, task.version)}
            type="button"
          >
            {savingTaskID === task.id ? "更新中..." : action.label}
          </button>
          <button onClick={() => void openTaskDetail(task.id)} type="button">查看任务详情</button>
        </div>
      </article>
	    );
	  }

	  const taskCopilotFilters: Record<string, string> = {};
	  if (searchQuery) taskCopilotFilters.query = searchQuery;
	  if (selectedStatus) taskCopilotFilters.status = selectedStatus;
	  if (selectedProject) taskCopilotFilters.project = selectedProject;
	  if (selectedPriority) taskCopilotFilters.priority = selectedPriority;
	  if (selectedTag) taskCopilotFilters.tag = selectedTag;
	  if (selectedSort) taskCopilotFilters.sort = selectedSort;
	  if (selectedGroup) taskCopilotFilters.group = selectedGroup;

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
        {batchMessage ? <p className="form-success" role="status">{batchMessage}</p> : null}
        {deletedTasks.length > 0 ? (
          <div className="form-success task-undo-message" role="status">
            <span>{deletedTasksMessage}</span>
            <button disabled={restoringTasks} onClick={() => void restoreDeletedTasks()} type="button">
              {restoringTasks ? "恢复中..." : "撤销删除"}
            </button>
          </div>
        ) : null}

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

        {aiDraft ? (
          <TaskModalPortal>
            <div className="task-modal-backdrop">
            <section aria-label="AI任务草稿预览" aria-modal="true" className="task-ai-draft-dialog" role="dialog">
              <header>
                <span>AI 建议草稿</span>
                <h2>确认后再进入正式任务</h2>
                <p>{aiDraft.goal} · 共 {aiDraftItems.length} 条建议</p>
              </header>
              <div className="task-ai-draft-toolbar">
                <label>
                  <input
                    checked={aiDraftItems.length > 0 && aiDraftItems.every((item) => item.selected)}
                    disabled={aiDraftBusy}
                    onChange={(event) => setAIDraftItems((current) => current.map((item) => ({ ...item, selected: event.target.checked })))}
                    type="checkbox"
                  />
                  全选建议
                </label>
                <span>已选 {aiDraftItems.filter((item) => item.selected).length} 条</span>
                <button disabled={aiDraftBusy || isCreatingTask} onClick={regenerateAIDraft} type="button">重新生成</button>
              </div>
              <div className="task-ai-draft-list">
                {aiDraftItems.length === 0 ? <div className="module-empty-state" role="status">暂无可采纳建议</div> : null}
                {aiDraftItems.map((item) => (
                  <article className={item.selected ? "task-ai-draft-item selected" : "task-ai-draft-item"} key={item.draftIndex}>
                    <label className="task-ai-draft-select">
                      <input
                        aria-label={`选择建议 ${item.title}`}
                        checked={item.selected}
                        disabled={aiDraftBusy}
                        onChange={(event) => updateAIDraftItem(item.draftIndex, "selected", event.target.checked)}
                        type="checkbox"
                      />
                    </label>
                    <div className="task-ai-draft-fields">
                      <label className="wide">
                        <span>任务标题</span>
                        <input disabled={!item.selected || aiDraftBusy} maxLength={100} onChange={(event) => updateAIDraftItem(item.draftIndex, "title", event.target.value)} value={item.title} />
                      </label>
                      <label>
                        <span>所属项目</span>
                        <input disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "project", event.target.value)} value={item.project} />
                      </label>
                      <label>
                        <span>负责人</span>
                        <input disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "assignee", event.target.value)} value={item.assignee} />
                      </label>
                      <label>
                        <span>优先级</span>
                        <select disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "priority", event.target.value as TaskPriority)} value={item.priority}>
                          <option value="high">高</option>
                          <option value="medium">中</option>
                          <option value="low">低</option>
                        </select>
                      </label>
                      <label>
                        <span>截止时间</span>
                        <input disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "dueAt", event.target.value)} type="datetime-local" value={item.dueAt} />
                      </label>
                      <label className="wide">
                        <span>任务描述</span>
                        <textarea disabled={!item.selected || aiDraftBusy} maxLength={1000} onChange={(event) => updateAIDraftItem(item.draftIndex, "description", event.target.value)} value={item.description} />
                      </label>
                      <label>
                        <span>标签</span>
                        <input disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "tags", event.target.value)} value={item.tags} />
                      </label>
                      <label>
                        <span>建议工具</span>
                        <input disabled={!item.selected || aiDraftBusy} onChange={(event) => updateAIDraftItem(item.draftIndex, "tools", event.target.value)} value={item.tools} />
                      </label>
                    </div>
                    <button aria-label={`删除建议 ${item.title}`} className="task-ai-draft-remove" disabled={aiDraftBusy} onClick={() => setAIDraftItems((current) => current.filter((candidate) => candidate.draftIndex !== item.draftIndex))} type="button">删除建议</button>
                  </article>
                ))}
              </div>
              {aiDraftError ? <p className="form-error" role="alert">{aiDraftError}</p> : null}
              <footer className="task-ai-draft-footer">
                <button disabled={aiDraftBusy} onClick={closeAIDraftPreview} type="button">稍后处理</button>
                <button className="primary" disabled={aiDraftBusy || aiDraftItems.every((item) => !item.selected)} onClick={() => void adoptAIDraftTasks()} type="button">
                  {aiDraftBusy ? "采纳中..." : "采纳选中任务"}
                </button>
              </footer>
            </section>
            </div>
          </TaskModalPortal>
        ) : null}

	        <section className={`task-workbench${useFullWidthWorkbench ? " full-width" : ""}`}>
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
                {activeView === "list" ? (
                  <button
                    aria-expanded={columnConfigOpen}
                    aria-label="配置列表字段"
                    className={columnConfigOpen ? "active" : ""}
                    onClick={() => setColumnConfigOpen((open) => !open)}
                    title="配置列表字段"
                    type="button"
                  ><SlidersHorizontal aria-hidden="true" /><span>字段</span></button>
                ) : null}
              </div>
            </div>
            {activeView === "list" && columnConfigOpen ? (
              <section aria-label="字段配置" className="task-column-config">
                <header>
                  <div>
                    <h3>列表字段</h3>
                    <span>已显示 {taskColumns.length} 项</span>
                  </div>
                  <button aria-label="关闭字段配置" onClick={() => setColumnConfigOpen(false)} title="关闭" type="button"><X aria-hidden="true" /></button>
                </header>
                <div className="task-column-config-list">
                  {orderedColumnDefinitions.map((column) => {
                    const visible = taskColumns.includes(column.key);
                    const index = taskColumns.indexOf(column.key);
                    return (
                      <div className={visible ? "visible" : ""} key={column.key}>
                        <button
                          aria-label={`${visible ? "隐藏" : "显示"}字段 ${column.label}`}
                          disabled={column.required}
                          onClick={() => toggleTaskColumn(column.key)}
                          title={column.required ? "基础字段不可隐藏" : visible ? "隐藏字段" : "显示字段"}
                          type="button"
                        >{visible ? <Eye aria-hidden="true" /> : <EyeOff aria-hidden="true" />}</button>
                        <strong>{column.label}</strong>
                        {column.required ? <small>必选</small> : null}
                        {visible ? (
                          <span className="task-column-order-actions">
                            <button aria-label={`上移字段 ${column.label}`} disabled={index <= 0} onClick={() => moveTaskColumn(column.key, -1)} title="上移" type="button"><ArrowUp aria-hidden="true" /></button>
                            <button aria-label={`下移字段 ${column.label}`} disabled={index < 0 || index >= taskColumns.length - 1} onClick={() => moveTaskColumn(column.key, 1)} title="下移" type="button"><ArrowDown aria-hidden="true" /></button>
                          </span>
                        ) : null}
                      </div>
                    );
                  })}
                </div>
                <footer>
                  <button className="secondary" disabled={columnsSaving} onClick={() => {
                    setTaskColumns(defaultTaskColumns);
                    setColumnsMessage("");
                  }} type="button"><RotateCcw aria-hidden="true" />恢复默认</button>
                  <button disabled={columnsLoading || columnsSaving} onClick={() => void saveTaskColumns()} type="button">{columnsSaving ? "保存中..." : "保存配置"}</button>
                </footer>
                {columnsMessage ? <p role="status">{columnsMessage}</p> : null}
                {columnsError ? <p className="form-error" role="alert">{columnsError}</p> : null}
              </section>
            ) : null}
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
                  placeholder="搜索任务、负责人、项目或标签"
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
                {selectedProject && !projectOptions.includes(selectedProject) ? <option value={selectedProject}>{selectedProject}</option> : null}
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
              <label className="sr-only" htmlFor="task-tag-filter">按标签筛选</label>
	              <select
	                id="task-tag-filter"
                onChange={(event) => {
                  setTaskPage(1);
                  setSelectedTag(event.target.value);
                }}
                onFocus={() => void loadTagOptions()}
                value={selectedTag}
              >
                <option value="">{tagsLoading ? "读取标签中..." : "全部标签"}</option>
                {selectedTag && !tagOptions.includes(selectedTag) ? <option value={selectedTag}>{selectedTag}</option> : null}
                {tagOptions.map((tag) => <option key={tag} value={tag}>{tag}</option>)}
	              </select>
	              <label className="sr-only" htmlFor="task-sort">任务排序</label>
	              <select
	                id="task-sort"
	                onChange={(event) => {
	                  setTaskPage(1);
	                  setSelectedSort(event.target.value as TaskSort);
	                }}
	                value={selectedSort}
	              >
	                {taskSortOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
	              </select>
	              <label className="sr-only" htmlFor="task-group">任务分组</label>
	              <select
	                id="task-group"
	                onChange={(event) => {
	                  setTaskPage(1);
	                  setSelectedGroup(event.target.value as TaskGroup | "");
	                }}
	                value={selectedGroup}
	              >
	                {taskGroupOptions.map((option) => <option key={option.value || "none"} value={option.value}>{option.label}</option>)}
	              </select>
	              {hasActiveFilters ? <button className="task-filter-clear" onClick={clearTaskFilters} type="button">清除筛选</button> : null}
            </form>
            {activeView === "list" ? (
              <>
                {visibleTasks.length > 0 ? (
                  <div className="task-selection-head">
                    <label>
                      <input
                        aria-label="全选当前页任务"
                        checked={allCurrentPageSelected}
                        disabled={batchBusy}
                        onChange={toggleCurrentPageSelection}
                        type="checkbox"
                      />
                      <span>全选当前页</span>
                    </label>
                  </div>
                ) : null}
                {selectedTaskIDs.size > 0 ? (
                  <div aria-label="批量操作" className="task-batch-toolbar" role="toolbar">
                    <strong>已选择 {selectedTaskIDs.size} 项</strong>
                    <label className="sr-only" htmlFor="task-batch-status">批量设置状态</label>
                    <select
                      aria-label="批量设置状态"
                      disabled={batchBusy}
                      id="task-batch-status"
                      onChange={(event) => {
                        setBatchStatus(event.target.value as TaskStatus);
                        setConfirmBatchDelete(false);
                      }}
                      value={selectedBatchStatus}
                    >
                      {batchStatusOptions.map((option) => <option key={option.status} value={option.status}>{option.label}</option>)}
                    </select>
                    <button disabled={batchBusy} onClick={() => void applyBatchStatus()} type="button">应用状态</button>
                    <label className="sr-only" htmlFor="task-batch-field">批量编辑字段</label>
                    <select
                      aria-label="批量编辑字段"
                      disabled={batchBusy}
                      id="task-batch-field"
                      onChange={(event) => setBatchEditField(event.target.value as typeof batchEditField)}
                      value={batchEditField}
                    >
                      <option value="assignee">负责人</option>
                      <option value="priority">优先级</option>
                      <option value="due_at">截止时间</option>
                      <option value="tags">标签</option>
                    </select>
                    {batchEditField === "assignee" ? (
                      <input aria-label="批量负责人" disabled={batchBusy} maxLength={100} onChange={(event) => setBatchAssignee(event.target.value)} placeholder="负责人，可留空" value={batchAssignee} />
                    ) : null}
                    {batchEditField === "priority" ? (
                      <select aria-label="批量优先级" disabled={batchBusy} onChange={(event) => setBatchPriority(event.target.value as TaskPriority)} value={batchPriority}>
                        <option value="high">高</option>
                        <option value="medium">中</option>
                        <option value="low">低</option>
                      </select>
                    ) : null}
                    {batchEditField === "due_at" ? (
                      <input aria-label="批量截止时间" disabled={batchBusy} onChange={(event) => setBatchDueAt(event.target.value)} type="datetime-local" value={batchDueAt} />
                    ) : null}
                    {batchEditField === "tags" ? (
                      <input aria-label="批量标签" disabled={batchBusy} onChange={(event) => setBatchTags(event.target.value)} placeholder="标签用逗号分隔" value={batchTags} />
                    ) : null}
                    <button disabled={batchBusy} onClick={() => void applyBatchField()} type="button">应用字段</button>
                    {batchEditField === "due_at" ? <button className="secondary" disabled={batchBusy} onClick={() => void applyBatchField(true)} type="button">清除截止时间</button> : null}
                    <button className="danger" disabled={batchBusy} onClick={() => void deleteSelectedTasks()} type="button">
                      {confirmBatchDelete ? `确认删除 ${selectedTaskIDs.size} 项` : "批量删除"}
                    </button>
                  </div>
                ) : null}
                <div aria-label="任务列表" className="task-table">
                  {visibleTasks.length === 0 ? (
	                  <div className="module-empty-state task-empty-state" role="status">
	                    <strong>{listLoading ? "正在读取任务..." : "暂无任务数据"}</strong>
	                    {!listLoading && hasActiveFilters ? <button onClick={clearTaskFilters} type="button">清除筛选</button> : null}
	                    {!listLoading && !hasActiveFilters ? <Link to="/tasks/new"><Plus aria-hidden="true" />创建首个任务</Link> : null}
	                  </div>
	                  ) : listGroups.map(([group, tasks]) => (
	                    <Fragment key={group || "all-tasks"}>
	                      {group ? (
	                        <div className="task-list-group-heading">
	                          <strong>{group}</strong>
	                          <span>{tasks.length} 项</span>
	                        </div>
	                      ) : null}
	                      {tasks.map((task) => (
	                    <article key={task.id ?? task.title}>
                      {task.id ? (
                        <input
                          aria-label={`选择任务 ${task.title}`}
                          checked={selectedTaskIDs.has(task.id)}
                          disabled={batchBusy}
                          onChange={() => toggleTaskSelection(task.id as number)}
                          type="checkbox"
                        />
                      ) : null}
                    <div className="task-list-fields">
                      {taskColumns.map((column) => (
                        <div className={`task-list-field task-list-field-${column}`} key={column}>
                          <small>{taskColumnDefinitions.find((definition) => definition.key === column)?.label}</small>
                          {renderTaskListColumn(task, column)}
                        </div>
                      ))}
                    </div>
                    {task.id && task.statusCode ? (
                      <div className="task-row-actions">
                        <button
                          disabled={savingTaskID === task.id}
                          onClick={() => void updateTaskStatus(task.id as number, task.statusCode as TaskStatus, task.version)}
                          type="button"
                        >
                          {savingTaskID === task.id ? "更新中..." : taskStatusAction(task.statusCode).label}
                        </button>
                        <button onClick={() => void openTaskDetail(task.id as number)} type="button">查看任务详情</button>
                      </div>
                    ) : null}
	                    </article>
	                      ))}
	                    </Fragment>
	                  ))}
                </div>
              </>
            ) : null}
            {activeView === "board" ? (
              <div aria-label="任务看板" className="task-view-board" role="region">
                {boardColumnLabels.map((column) => {
                  const columnTasks = apiTasks.filter((task) => task.status === column.status);
                  return (
	                    <section
	                      aria-label={`${column.label}任务`}
	                      className={draggingTaskID !== null ? "task-view-column drop-ready" : "task-view-column"}
	                      key={column.status}
	                      onDragOver={(event) => {
	                        if (draggingTaskID !== null) {
	                          event.preventDefault();
	                          event.dataTransfer.dropEffect = "move";
	                        }
	                      }}
	                      onDrop={(event) => {
	                        event.preventDefault();
	                        const taskID = draggingTaskID ?? Number(event.dataTransfer.getData("text/plain"));
	                        if (Number.isInteger(taskID) && taskID > 0) void moveTaskOnBoard(taskID, column.status);
	                      }}
	                    >
                      <header>
                        <h3>{column.label}</h3>
                        <span>{columnTasks.length}</span>
                      </header>
                      <div>
                        {columnTasks.length === 0
                          ? <div className="module-empty-state">暂无任务</div>
	                          : columnTasks.map((task) => renderTaskCard(task, true))}
                      </div>
                    </section>
                  );
                })}
              </div>
            ) : null}
            {activeView === "calendar" ? (
              <div aria-label="任务日历" className="task-calendar-view" role="region">
                <header className="task-calendar-toolbar">
                  <button aria-label="上个月" disabled={calendarLoading} onClick={() => setCalendarMonth((month) => shiftCalendarMonth(month, -1))} title="上个月" type="button">‹</button>
                  <strong>{calendarMonthLabel(calendarMonth)}</strong>
                  <button aria-label="下个月" disabled={calendarLoading} onClick={() => setCalendarMonth((month) => shiftCalendarMonth(month, 1))} title="下个月" type="button">›</button>
                  <button disabled={calendarLoading || calendarMonth === currentCalendarMonth()} onClick={() => setCalendarMonth(currentCalendarMonth())} type="button">今天</button>
                  {calendarLoading ? <span role="status">正在读取完整月份...</span> : null}
                </header>
                {calendarSourceTasks.length === 0 ? <div className="module-empty-state" role="status">本月暂无任务数据</div> : null}
                {calendarGroups.map(([label, tasks]) => (
                  <section aria-label={`${label}任务`} className="task-calendar-group" key={label}>
                    <header>
                      <h3>{label}</h3>
                      <span>{tasks.length} 项</span>
                    </header>
                    <div>{tasks.map((task) => renderTaskCard(task))}</div>
                  </section>
                ))}
                {unscheduledTasks.length > 0 ? (
                  <section aria-label="待安排任务" className="task-calendar-group unscheduled">
                    <header>
                      <h3>待安排</h3>
                      <span>{unscheduledTasks.length} 项</span>
                    </header>
                    <div>{unscheduledTasks.map((task) => renderTaskCard(task))}</div>
                  </section>
                ) : null}
              </div>
            ) : null}
            {taskTotal > 0 && activeView !== "calendar" ? (
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

	          <div className="task-side-rail">
	            {showBoardPreview ? (
	              <aside className="task-board-panel" aria-label="任务看板预览">
	                <h2>跨项目看板</h2>
	                <p>切换到看板视图可按状态处理跨项目任务</p>
	                <div>
	                  {populatedBoardColumns.map(([title, items]) => (
	                    <section key={title}>
	                      <strong>{title}</strong>
	                      {items.map((item) => <span key={item}>{item}</span>)}
	                    </section>
	                  ))}
	                </div>
	              </aside>
	            ) : null}
	            <UnifiedCopilotPanel
	              activeFilters={taskCopilotFilters}
	              ariaLabel="任务中心 Copilot"
	              className="tasks-copilot-panel"
	              currentView={activeView}
	              inputAriaLabel="向任务中心 Copilot 提问"
	              response={activeView === "board" ? "我会结合当前看板状态识别阻塞、临期和在制任务。" : activeView === "calendar" ? "我会结合当前月份和筛选条件检查排期、临期与逾期风险。" : "我会结合当前列表筛选和任务详情，整理优先级与下一步。"}
	              taskID={detailTask?.id}
	              userPrompt={detailTask ? `分析任务：${detailTask.title}` : "分析当前任务中心"}
	            />
	          </div>
        </section>
        {detailTaskID !== null ? (
          <TaskModalPortal>
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
                      <span>标签</span>
                      <input onChange={(event) => updateDetailField("tags", event.target.value)} placeholder="多个标签使用逗号分隔" value={detailForm.tags} />
                    </label>
                    <label>
                      <span>截止时间</span>
                      <input onChange={(event) => updateDetailField("dueAt", event.target.value)} type="datetime-local" value={detailForm.dueAt} />
                    </label>
                    <label>
                      <span>任务状态</span>
                      <select onChange={(event) => updateDetailField("status", event.target.value as TaskStatus)} value={detailForm.status}>
                        {taskStatusOptions(detailTask.status).map((item) => (
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
                    <label>
                      <span>完成进度 {detailForm.progress}%</span>
                      <input
                        aria-label="任务进度"
                        max={100}
                        min={0}
                        onChange={(event) => updateDetailField("progress", Number(event.target.value))}
                        type="range"
                        value={detailForm.progress}
                      />
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
                  {detailTask.source_title && detailTask.source_url ? (
                    <section aria-label="任务来源" className="task-source-section">
                      <div>
                        <h3>任务来源</h3>
                        <p>{detailTask.source_title}</p>
                      </div>
                      <Link aria-label={`查看来源：${detailTask.source_title}`} to={detailTask.source_url}>查看来源</Link>
                    </section>
                  ) : null}
                  <section aria-label="提醒设置" className="task-reminder-section">
                    <header>
                      <h3>提醒设置</h3>
                      <span>{reminderLoading ? "读取中" : taskReminder?.recurrence !== "once" && taskReminder ? "循环中" : taskReminder?.sent_at ? "已触发" : taskReminder ? "待触发" : "暂无提醒"}</span>
                    </header>
                    <div className="task-reminder-controls">
                      <label>
                        <span>提醒时间</span>
                        <input
                          aria-label="提醒时间"
                          disabled={reminderLoading || reminderSaving}
                          onChange={(event) => {
                            setReminderTime(event.target.value);
                            setReminderMessage("");
                          }}
                          type="datetime-local"
                          value={reminderTime}
                        />
                      </label>
                      <label>
                        <span>提醒频率</span>
                        <select
                          aria-label="提醒频率"
                          disabled={reminderLoading || reminderSaving}
                          onChange={(event) => {
                            setReminderRecurrence(event.target.value as ReminderRecurrence);
                            setReminderMessage("");
                            setReminderNeedsUpgrade(false);
                          }}
                          value={reminderRecurrence}
                        >
                          <option value="once">一次性</option>
                          <option value="daily">每天（会员）</option>
                          <option value="weekly">每周（会员）</option>
                        </select>
                      </label>
                      <button disabled={reminderLoading || reminderSaving || !reminderTime} onClick={() => void saveTaskReminder()} type="button">
                        {reminderSaving ? "处理中..." : "保存提醒"}
                      </button>
                      {taskReminder ? (
                        <button className="cancel" disabled={reminderSaving} onClick={() => void cancelTaskReminder()} type="button">取消提醒</button>
                      ) : null}
                    </div>
                    {reminderMessage ? <p className="task-reminder-message" role="status">{reminderMessage}</p> : null}
                    {reminderError ? <p className="form-error" role="alert">{reminderError}</p> : null}
                    {reminderNeedsUpgrade ? <Link className="task-reminder-upgrade" to="/membership">升级会员</Link> : null}
                  </section>
                  <section aria-label="子任务" className="task-subtask-section">
                    <header>
                      <div>
                        <h3>子任务</h3>
                        <p>将当前任务拆成可逐项完成的执行步骤</p>
                      </div>
                      <span>{subtasks.filter((item) => item.completed).length} / {subtasks.length} 已完成</span>
                    </header>
                    {subtaskParent ? (
                      <div className="task-subtask-parent-target" role="status">
                        <span>正在为“{subtaskParent.title}”添加下级</span>
                        <button aria-label="取消添加下级" onClick={() => setSubtaskParentID(null)} title="取消添加下级" type="button"><X aria-hidden="true" /></button>
                      </div>
                    ) : null}
                    <div className="task-subtask-create">
                      <input
                        aria-label="新建子任务"
                        disabled={subtaskCreating}
                        maxLength={100}
                        onChange={(event) => setSubtaskTitle(event.target.value)}
                        onKeyDown={(event) => {
                          if (event.key !== "Enter") return;
                          event.preventDefault();
                          void createDetailSubtask();
                        }}
                        placeholder={subtaskParent ? "输入下级子任务标题" : "输入子任务标题"}
                        value={subtaskTitle}
                      />
                      <button disabled={subtaskCreating || !subtaskTitle.trim()} onClick={() => void createDetailSubtask()} type="button">
                        {subtaskCreating ? "添加中..." : subtaskParent ? "添加下级" : "添加子任务"}
                      </button>
                    </div>
                    {subtasksLoading ? <div className="task-subtask-empty" role="status">正在读取子任务...</div> : null}
                    {!subtasksLoading && subtasks.length === 0 ? <div className="task-subtask-empty" role="status">暂无子任务</div> : null}
                    {!subtasksLoading && subtasks.length > 0 ? (
                      <div className="task-subtask-list">
                        {subtaskTreeRows.map(({ item, depth, descendantCount, hasChildren }) => (
                          <article
                            className={item.completed ? "completed" : ""}
                            key={item.id}
                            style={{ "--subtask-depth": Math.min(depth, 8) } as CSSProperties}
                          >
                            {hasChildren ? (
                              <button
                                aria-label={`${collapsedSubtaskIDs.has(item.id) ? "展开" : "收起"}子任务 ${item.title}`}
                                className="task-subtask-collapse"
                                onClick={() => setCollapsedSubtaskIDs((current) => {
                                  const next = new Set(current);
                                  if (next.has(item.id)) next.delete(item.id);
                                  else next.add(item.id);
                                  return next;
                                })}
                                title={collapsedSubtaskIDs.has(item.id) ? "展开下级" : "收起下级"}
                                type="button"
                              >
                                {collapsedSubtaskIDs.has(item.id) ? <ChevronRight aria-hidden="true" /> : <ChevronDown aria-hidden="true" />}
                              </button>
                            ) : <span className="task-subtask-collapse-spacer" />}
                            <label>
                              <input
                                aria-label={`完成子任务 ${item.title}`}
                                checked={item.completed}
                                disabled={subtaskBusyID !== null}
                                onChange={() => void toggleDetailSubtask(item)}
                                type="checkbox"
                              />
                            </label>
                            <div className="task-subtask-content">
                              <strong>{item.title}</strong>
                              {item.assignee || item.due_at ? (
                                <small>{item.assignee ? `负责人 ${item.assignee}` : ""}{item.assignee && item.due_at ? " · " : ""}{item.due_at ? `截止 ${formatDueAt(item.due_at)}` : ""}</small>
                              ) : null}
                            </div>
                            <div className="task-subtask-actions">
                              <button
                                aria-label={`为子任务 ${item.title} 添加下级`}
                                disabled={subtaskBusyID !== null}
                                onClick={() => {
                                  setSubtaskParentID(item.id);
                                  setSubtaskTitle("");
                                  setSubtaskError("");
                                }}
                                title="添加下级"
                                type="button"
                              ><Plus aria-hidden="true" /></button>
                              {confirmSubtaskDeleteID === item.id ? (
                                <>
                                  <button
                                    className="danger"
                                    disabled={subtaskBusyID !== null}
                                    onClick={() => void deleteDetailSubtask(item)}
                                    type="button"
                                  >确认删除{descendantCount > 0 ? `（含 ${descendantCount} 个下级）` : ""}</button>
                                  <button aria-label={`取消删除子任务 ${item.title}`} onClick={() => setConfirmSubtaskDeleteID(null)} title="取消删除" type="button"><X aria-hidden="true" /></button>
                                </>
                              ) : (
                                <button
                                  aria-label={`删除子任务 ${item.title}`}
                                  disabled={subtaskBusyID !== null}
                                  onClick={() => setConfirmSubtaskDeleteID(item.id)}
                                  title="删除子任务"
                                  type="button"
                                ><Trash2 aria-hidden="true" /></button>
                              )}
                            </div>
                          </article>
                        ))}
                      </div>
                    ) : null}
                    {subtaskError ? <p className="form-error" role="alert">{subtaskError}</p> : null}
                  </section>
                  <section aria-label="任务附件" className="task-attachment-section">
                    <header>
                      <div>
                        <h3>任务附件</h3>
                        <p>文件受任务权限保护，下载地址会短期失效</p>
                      </div>
                      <span>{taskAttachments.length} / 10</span>
                    </header>
                    <input
                      accept=".txt,.md,.markdown,.csv,.pdf,.docx,.xlsx,.pptx,.png,.jpg,.jpeg"
                      aria-label="选择任务附件"
                      className="task-attachment-input"
                      onChange={(event) => {
                        const file = event.target.files?.[0];
                        if (file) void uploadTaskAttachment(file);
                      }}
                      ref={attachmentInputRef}
                      type="file"
                    />
                    <button
                      className="task-attachment-upload"
                      disabled={attachmentsLoading || attachmentBusyID !== null || taskAttachments.length >= 10}
                      onClick={() => attachmentInputRef.current?.click()}
                      type="button"
                    >
                      <Upload aria-hidden="true" />
                      {attachmentBusyID === "upload" ? "上传中..." : "上传附件"}
                    </button>
                    {attachmentsLoading ? <div className="task-subtask-empty" role="status">正在读取附件...</div> : null}
                    {!attachmentsLoading && taskAttachments.length === 0 && !attachmentsError ? <div className="task-subtask-empty" role="status">暂无附件</div> : null}
                    {taskAttachments.length > 0 ? (
                      <div className="task-attachment-list">
                        {taskAttachments.map((attachment) => (
                          <article key={attachment.id}>
                            <Paperclip aria-hidden="true" />
                            <div>
                              <strong title={attachment.name}>{attachment.name}</strong>
                              <small>{formatAttachmentSize(attachment.size_bytes)} · {formatDueAt(attachment.created_at)}</small>
                            </div>
                            <button
                              aria-label={`下载附件 ${attachment.name}`}
                              disabled={attachmentBusyID !== null}
                              onClick={() => void downloadTaskAttachment(attachment)}
                              title="下载附件"
                              type="button"
                            ><Download aria-hidden="true" /></button>
                            <button
                              aria-label={`删除附件 ${attachment.name}`}
                              className="danger"
                              disabled={attachmentBusyID !== null}
                              onClick={() => void deleteTaskAttachment(attachment)}
                              title="删除附件"
                              type="button"
                            ><Trash2 aria-hidden="true" /></button>
                          </article>
                        ))}
                      </div>
                    ) : null}
                    {attachmentsError ? <p className="form-error" role="alert">{attachmentsError}</p> : null}
                  </section>
                  <section aria-label="任务评论" className="task-comment-section">
                    <header>
                      <div>
                        <h3>协作评论</h3>
                        <p>记录任务讨论和执行补充</p>
                      </div>
                      <span>{commentTotal} 条</span>
                    </header>
                    <div className="task-comment-create">
                      <textarea
                        aria-label="评论内容"
                        disabled={commentSavingID !== null}
                        maxLength={2000}
                        onChange={(event) => setCommentContent(event.target.value)}
                        placeholder="输入评论内容"
                        value={commentContent}
                      />
                      <button disabled={commentSavingID !== null || !commentContent.trim()} onClick={() => void createTaskComment()} type="button">
                        {commentSavingID === "new" ? "发表中..." : "发表评论"}
                      </button>
                    </div>
                    {commentsLoading ? <div className="task-subtask-empty" role="status">正在读取评论...</div> : null}
                    {!commentsLoading && taskComments.length === 0 && !commentsError ? <div className="task-subtask-empty" role="status">暂无评论</div> : null}
                    {taskComments.length > 0 ? (
                      <div className="task-comment-list">
                        {taskComments.map((comment) => (
                          <article key={comment.id}>
                            {commentEditingID === comment.id ? (
                              <textarea aria-label={`编辑评论 ${comment.id}`} disabled={commentSavingID !== null} maxLength={2000} onChange={(event) => setCommentEditingContent(event.target.value)} value={commentEditingContent} />
                            ) : (
                              <p>{comment.content}</p>
                            )}
                            <footer>
                              <small>你 · {formatDueAt(comment.created_at)}{comment.updated_at !== comment.created_at ? " · 已编辑" : ""}</small>
                              {commentEditingID === comment.id ? (
                                <>
                                  <button disabled={commentSavingID !== null} onClick={() => {
                                    setCommentEditingID(null);
                                    setCommentEditingContent("");
                                  }} type="button">取消</button>
                                  <button disabled={commentSavingID !== null || !commentEditingContent.trim()} onClick={() => void saveTaskComment(comment.id)} type="button">保存</button>
                                </>
                              ) : (
                                <>
                                  <button disabled={commentSavingID !== null} onClick={() => {
                                    setCommentEditingID(comment.id);
                                    setCommentEditingContent(comment.content);
                                  }} type="button">编辑</button>
                                  <button className="danger" disabled={commentSavingID !== null} onClick={() => void deleteTaskComment(comment.id)} type="button">删除</button>
                                </>
                              )}
                            </footer>
                          </article>
                        ))}
                      </div>
                    ) : null}
                    {commentsError ? <p className="form-error" role="alert">{commentsError}</p> : null}
                  </section>
                  <section aria-label="操作记录" className="task-activity-section">
                    <header>
                      <div>
                        <h3>操作记录</h3>
                        <p>系统自动记录任务的重要变更</p>
                      </div>
                      <span>{activityTotal} 条</span>
                    </header>
                    {activitiesLoading && taskActivities.length === 0 ? <div className="task-subtask-empty" role="status">正在读取操作记录...</div> : null}
                    {!activitiesLoading && taskActivities.length === 0 && !activitiesError ? <div className="task-subtask-empty" role="status">暂无操作记录</div> : null}
                    {taskActivities.length > 0 ? (
                      <div className="task-activity-list">
                        {taskActivities.map((activity) => (
                          <article key={activity.id}>
                            <span className="task-activity-dot" aria-hidden="true" />
                            <div>
                              <strong>{taskActivitySummary(activity)}</strong>
                              <small>你 · {formatDueAt(activity.created_at)}</small>
                            </div>
                          </article>
                        ))}
                      </div>
                    ) : null}
                    {activitiesError ? <p className="form-error" role="alert">{activitiesError}</p> : null}
                    {taskActivities.length < activityTotal ? (
                      <button className="task-activity-more" disabled={activitiesLoading} onClick={() => void loadMoreTaskActivities()} type="button">
                        {activitiesLoading ? "读取中..." : "加载更多记录"}
                      </button>
                    ) : null}
                  </section>
                  {detailError ? <p className="form-error" role="alert">{detailError}</p> : null}
                  {confirmDelete ? (
                    <div className="task-delete-confirm" role="alert">
                      <span>任务会从当前列表隐藏，可在顶部提示中撤销删除。</span>
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
          </TaskModalPortal>
        ) : null}
      </section>
    </V4PageShell>
  );
}

export default TasksPage;
