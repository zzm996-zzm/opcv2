import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { useAuthSession } from "../lib/authSession";
import { tasksApi, type TaskPriority } from "../lib/tasksApi";

type ManualTaskForm = {
  title: string;
  description: string;
  project: string;
  assignee: string;
  dueAt: string;
  priority: TaskPriority;
  tags: string;
  tools: string;
  learning: string;
};

function emptyManualTask(assignee = ""): ManualTaskForm {
  return {
    title: "",
    description: "",
    project: "",
    assignee,
    dueAt: "",
    priority: "medium",
    tags: "",
    tools: "",
    learning: ""
  };
}

function parseList(value: string) {
  return Array.from(new Set(value
    .split(/[,，]/)
    .map((tool) => tool.trim())
    .filter(Boolean)));
}

function TaskCreatePage() {
  const navigate = useNavigate();
  const { user } = useAuthSession();
  const defaultAssignee = user?.nickname?.trim() ?? "";
  const [form, setForm] = useState<ManualTaskForm>(() => emptyManualTask(defaultAssignee));
  const [projects, setProjects] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    tasksApi.listProjects()
      .then((payload) => {
        if (active) setProjects(payload.projects);
      })
      .catch(() => {
        if (active) setProjects([]);
      });
    return () => {
      active = false;
    };
  }, []);

  function updateField<Key extends keyof ManualTaskForm>(key: Key, value: ManualTaskForm[Key]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function createTask(continueAdding: boolean) {
    if (saving) return;
    const title = form.title.trim();
    const project = form.project.trim();
    const assignee = form.assignee.trim();
    if (!title || !project || !assignee) {
      setError("请填写任务标题、所属项目和负责人");
      return;
    }
    setSaving(true);
    setError("");
    setMessage("");
    try {
      const task = await tasksApi.createTask({
        title,
        description: form.description.trim(),
        project,
        assignee,
        priority: form.priority,
        dueAt: form.dueAt ? new Date(form.dueAt).toISOString() : undefined,
        tags: parseList(form.tags),
        tools: parseList(form.tools),
        learning: form.learning.trim()
      });
      if (continueAdding) {
        setForm(emptyManualTask(defaultAssignee));
        setMessage(`已创建任务：${task.title}`);
      } else {
        navigate("/tasks");
      }
    } catch (requestError) {
      setError(apiErrorMessage(requestError, "暂时无法创建任务"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <V4PageShell>
      <section className="module-page task-create-page" aria-label="新建任务">
        <div className="task-create-heading">
          <Link aria-label="返回任务中心" to="/tasks">‹</Link>
          <div>
            <h1>新建任务</h1>
            <p>明确任务目标、执行时间和关联项目</p>
          </div>
        </div>
        {message ? <p className="form-success" role="status">{message}</p> : null}
        <form className="task-create-form" onSubmit={(event) => {
          event.preventDefault();
          void createTask(false);
        }}>
          <section>
            <header><h2>基本信息</h2></header>
            <div className="task-create-grid">
              <label className="wide">
                <span>任务标题 <b>*</b></span>
                <input aria-label="任务标题" maxLength={100} onChange={(event) => updateField("title", event.target.value)} placeholder="输入清晰、可执行的任务目标" value={form.title} />
                <small>{form.title.length}/100</small>
              </label>
              <label className="wide">
                <span>任务描述</span>
                <textarea aria-label="任务描述" maxLength={1000} onChange={(event) => updateField("description", event.target.value)} placeholder="说明任务背景、目标、范围和完成标准" value={form.description} />
                <small>{form.description.length}/1000</small>
              </label>
              <label>
                <span>所属项目 <b>*</b></span>
                <input aria-label="所属项目" list="task-project-options" onChange={(event) => updateField("project", event.target.value)} placeholder="输入或选择项目" value={form.project} />
                <datalist id="task-project-options">
                  {projects.map((project) => <option key={project} value={project}>{project}</option>)}
                </datalist>
              </label>
              <label>
                <span>负责人 <b>*</b></span>
                <input aria-label="负责人" maxLength={100} onChange={(event) => updateField("assignee", event.target.value)} placeholder="输入负责人或外部协作人" value={form.assignee} />
              </label>
              <label>
                <span>截止时间</span>
                <input aria-label="截止时间" onChange={(event) => updateField("dueAt", event.target.value)} type="datetime-local" value={form.dueAt} />
              </label>
              <label>
                <span>优先级 <b>*</b></span>
                <select aria-label="优先级" onChange={(event) => updateField("priority", event.target.value as TaskPriority)} value={form.priority}>
                  <option value="high">高</option>
                  <option value="medium">中</option>
                  <option value="low">低</option>
                </select>
              </label>
              <label>
                <span>初始状态</span>
                <input disabled value="待开始" />
              </label>
              <label>
                <span>标签</span>
                <input aria-label="标签" onChange={(event) => updateField("tags", event.target.value)} placeholder="多个标签使用逗号分隔" value={form.tags} />
              </label>
            </div>
          </section>
          <section>
            <header><h2>执行资源</h2></header>
            <div className="task-create-grid">
              <label className="wide">
                <span>建议工具</span>
                <input aria-label="建议工具" onChange={(event) => updateField("tools", event.target.value)} placeholder="多个工具使用逗号分隔" value={form.tools} />
              </label>
              <label className="wide">
                <span>补课内容</span>
                <textarea aria-label="补课内容" onChange={(event) => updateField("learning", event.target.value)} placeholder="记录执行前需要补充的知识" value={form.learning} />
              </label>
            </div>
          </section>
          {error ? <p className="form-error" role="alert">{error}</p> : null}
          <footer>
            <Link to="/tasks">取消</Link>
            <button disabled={saving} onClick={() => void createTask(true)} type="button">保存并继续添加</button>
            <button className="primary" disabled={saving} type="submit">{saving ? "保存中..." : "保存任务"}</button>
          </footer>
        </form>
      </section>
    </V4PageShell>
  );
}

export default TaskCreatePage;
