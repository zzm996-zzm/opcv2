import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { projectsApi, type UserProject } from "../lib/projectsApi";
import "../user-projects.css";

function statusLabel(status: UserProject["status"]) {
  return status === "active" ? "进行中" : status === "archived" ? "已归档" : "草稿";
}

export default function UserProjectsPage() {
  const [projects, setProjects] = useState<UserProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

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
                <h2>{project.name}</h2>
                <p>{project.description || "暂未填写项目描述"}</p>
              </article>
            ))}
          </section>
        )}
      </main>
    </V4PageShell>
  );
}
