import { useEffect, useState } from "react";
import { Archive, Check, Pencil, Save, Trash2, X } from "lucide-react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { projectsApi, type UserProject } from "../lib/projectsApi";
import "../user-projects.css";

function statusLabel(status: UserProject["status"]) {
  return status === "active" ? "已发布" : status === "archived" ? "已归档" : "草稿";
}

export default function UserProjectsPage() {
  const [projects, setProjects] = useState<UserProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editingID, setEditingID] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [busyID, setBusyID] = useState<number | null>(null);

  useEffect(() => {
    let active = true;
    projectsApi.listUserProjects()
      .then((payload) => {
        if (active) setProjects(payload.projects ?? []);
      })
      .catch(() => {
        if (active) setError("项目列表暂时无法加载");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  function startEditing(project: UserProject) {
    setEditingID(project.id);
    setEditName(project.name);
    setEditDescription(project.description ?? "");
    setError("");
  }

  async function saveEditing(project: UserProject) {
    setBusyID(project.id);
    setError("");
    try {
      const updated = await projectsApi.updateUserProject(project.id, { name: editName, description: editDescription });
      setProjects((current) => current.map((item) => item.id === updated.id ? updated : item));
      setEditingID(null);
    } catch {
      setError("项目保存失败，请检查名称后重试");
    } finally {
      setBusyID(null);
    }
  }

  async function changeStatus(project: UserProject) {
    setBusyID(project.id);
    setError("");
    try {
      const updated = project.status === "active"
        ? await projectsApi.archiveUserProject(project.id)
        : await projectsApi.publishUserProject(project.id);
      setProjects((current) => current.map((item) => item.id === updated.id ? updated : item));
    } catch {
      setError("项目状态更新失败，请稍后重试");
    } finally {
      setBusyID(null);
    }
  }

  async function removeProject(project: UserProject) {
    if (!window.confirm(`确定删除“${project.name}”吗？删除后无法恢复。`)) return;
    setBusyID(project.id);
    setError("");
    try {
      await projectsApi.deleteUserProject(project.id);
      setProjects((current) => current.filter((item) => item.id !== project.id));
    } catch {
      setError("项目删除失败，请稍后重试");
    } finally {
      setBusyID(null);
    }
  }

  return (
    <V4PageShell className="user-projects-shell">
      <main className="user-projects-page">
        <header className="user-projects-head">
          <div>
            <p className="user-projects-eyebrow">我的工作区</p>
            <h1>我的项目</h1>
            <p>在 Copilot 中确认创建的项目草稿都会保存在这里。</p>
          </div>
          <Link className="user-projects-back" to="/copilot">返回 Copilot</Link>
        </header>
        {loading && <p className="user-projects-state">正在加载项目...</p>}
        {!loading && error && <p className="user-projects-state is-error">{error}</p>}
        {!loading && !error && projects.length === 0 && (
          <section className="user-projects-empty">
            <strong>还没有项目</strong>
            <p>在对话框中说“创建一个项目：项目名称”，确认后就会保存到这里。</p>
            <Link to="/copilot">打开 Copilot</Link>
          </section>
        )}
        {!loading && !error && projects.length > 0 && (
          <section className="user-projects-grid" aria-label="我的项目列表">
            {projects.map((project) => (
              <article className="user-project-card" key={project.id}>
                <div className="user-project-card-head">
                  <span className="user-project-status">{statusLabel(project.status)}</span>
                  <time>{new Date(project.created_at).toLocaleDateString("zh-CN")}</time>
                </div>
                {editingID === project.id ? (
                  <div className="user-project-edit-form">
                    <label>项目名称<input value={editName} onChange={(event) => setEditName(event.target.value)} /></label>
                    <label>项目描述<textarea value={editDescription} onChange={(event) => setEditDescription(event.target.value)} rows={4} /></label>
                    <div className="user-project-actions">
                      <button disabled={busyID === project.id} onClick={() => void saveEditing(project)} type="button"><Save aria-hidden="true" size={15} />保存</button>
                      <button disabled={busyID === project.id} onClick={() => setEditingID(null)} type="button"><X aria-hidden="true" size={15} />取消</button>
                    </div>
                  </div>
                ) : (
                  <>
                    <h2>{project.name}</h2>
                    <p>{project.description || "暂未填写项目描述"}</p>
                    <div className="user-project-actions">
                      <button disabled={busyID === project.id} onClick={() => startEditing(project)} type="button"><Pencil aria-hidden="true" size={15} />编辑</button>
                      <button disabled={busyID === project.id} onClick={() => void changeStatus(project)} type="button">
                        {project.status === "active" ? <Archive aria-hidden="true" size={15} /> : <Check aria-hidden="true" size={15} />}
                        {project.status === "active" ? "归档" : "发布"}
                      </button>
                      <button className="danger" disabled={busyID === project.id} onClick={() => void removeProject(project)} type="button"><Trash2 aria-hidden="true" size={15} />删除</button>
                    </div>
                  </>
                )}
              </article>
            ))}
          </section>
        )}
      </main>
    </V4PageShell>
  );
}
