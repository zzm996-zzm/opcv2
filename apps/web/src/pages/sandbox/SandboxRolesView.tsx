import { ArrowLeft, ArrowRight, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxRoleCard from "../../components/sandbox/SandboxRoleCard";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxRole, SandboxSession } from "../../lib/sandboxApi";

type SandboxRolesViewProps = {
  onSave: (roles: string[]) => Promise<SandboxSession>;
  roles: SandboxRole[];
  session: SandboxSession;
};

function SandboxRolesView({ onSave, roles, session }: SandboxRolesViewProps) {
  const [selected, setSelected] = useState<string[]>(session.roles ?? []);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => setSelected(session.roles ?? []), [session.id, session.roles]);

  function toggle(role: SandboxRole) {
    setSelected((current) => current.includes(role.label) ? current.filter((label) => label !== role.label) : [...current, role.label]);
  }

  async function save() {
    if (!selected.length || saving) return;
    setSaving(true);
    setError("");
    try {
      await onSave(selected);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "角色保存失败，请稍后重试。");
    } finally {
      setSaving(false);
    }
  }

  return (
    <SandboxFrame copilotMode="roles" copilotProject={session.intake?.initial_idea || session.goal}>
      <SandboxStepper active={2} />
      <section className="sb-work-panel sb-roles-panel">
        <header className="sb-panel-header sb-roles-heading">
          <div><h1>你希望从谁的视角进行推演？</h1><p>选择一个或多个角色，AI 将基于该角色的立场对话并给出建议。</p></div>
          <span className="sb-selected-count">已选 {selected.length} 个</span>
        </header>
        <div className="sb-role-select-grid">
          {roles.map((role) => <SandboxRoleCard key={role.key} onToggle={toggle} role={role} selected={selected.includes(role.label)} />)}
        </div>
        <aside className="sb-role-help">
          <Sparkles size={23} />
          <div><strong>为什么选择多个角色？</strong><p>多角色视角能帮助你更全面地发现机会、识别风险，获得更立体的推演结果。</p></div>
          <span className="sb-help-pedestal" aria-hidden="true" />
        </aside>
        <footer className="sb-flow-footer">
          <Link className="is-quiet" to={`/sandbox/questions?session=${session.id}`}><ArrowLeft size={17} />上一步</Link>
          <button className="is-primary" disabled={!selected.length || saving} onClick={() => void save()} type="button">
            {saving ? "保存中..." : "确认角色，进入下一步"}<ArrowRight size={18} />
          </button>
        </footer>
        {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
      </section>
    </SandboxFrame>
  );
}

export default SandboxRolesView;

