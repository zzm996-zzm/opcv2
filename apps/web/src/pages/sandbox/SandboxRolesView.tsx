import { ArrowLeft, ArrowRight, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxRoleCard from "../../components/sandbox/SandboxRoleCard";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxRole, SandboxRun } from "../../lib/sandboxApi";

type Props = { onSave: (roles: string[]) => Promise<SandboxRun>; roles: SandboxRole[]; run: SandboxRun };
export default function SandboxRolesView({ onSave, roles, run }: Props) {
  const [selected, setSelected] = useState(run.roles ?? []); const [saving, setSaving] = useState(false); const [error, setError] = useState("");
  useEffect(() => setSelected(run.roles ?? []), [run.id, run.roles]);
  function toggle(role: SandboxRole) { if (role.is_required) return; setSelected((current) => current.includes(role.role_code) ? current.filter((item) => item !== role.role_code) : [...current, role.role_code]); }
  async function save() { if (selected.length < 3 || selected.length > 8 || saving || !selected.includes("skeptic")) { setError("至少选择 3 个、最多 8 个角色，且必须包含悲观者。"); return; } setSaving(true); setError(""); try { await onSave(selected); } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "角色保存失败，请稍后重试。"); } finally { setSaving(false); } }
  return <SandboxFrame copilotMode="roles" copilotProject={run.product.name}><SandboxStepper active={3} /><section className="sb-work-panel sb-roles-panel"><header className="sb-panel-header sb-roles-heading"><div><h1>选择推演角色</h1><p>每个角色都是独立会话；报告阶段才会比较共识和分歧。</p></div><span className="sb-selected-count">已选 {selected.length} / 8</span></header><div className="sb-role-select-grid">{roles.map((role) => <SandboxRoleCard key={role.role_code} onToggle={toggle} role={role} selected={selected.includes(role.role_code)} />)}</div><aside className="sb-role-help"><Sparkles size={23} /><div><strong>悲观者不可取消</strong><p>强制保留反方意见，避免报告只呈现乐观结论。预计耗时约 {Math.max(1, Math.ceil(selected.length * 0.7))} 分钟。</p></div></aside><footer className="sb-flow-footer"><Link className="is-quiet" to={`/sandbox/new?run=${run.id}&step=2`}><ArrowLeft size={17} />上一步</Link><button className="is-primary" disabled={saving} onClick={() => void save()} type="button">{saving ? "保存中..." : "确认角色，下一步"}<ArrowRight size={18} /></button></footer>{error ? <p className="sb-inline-error" role="alert">{error}</p> : null}</section></SandboxFrame>;
}
