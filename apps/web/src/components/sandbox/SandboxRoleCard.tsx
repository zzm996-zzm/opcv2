import { Check } from "lucide-react";

import type { SandboxRole } from "../../lib/sandboxApi";

const roleArtwork: Record<string, string> = {
  customer: "/sandbox/role-user.jpg",
  investor: "/sandbox/role-investor.jpg",
  channel: "/sandbox/role-channel.jpg",
  competitor: "/sandbox/role-competitor.jpg",
  supply: "/sandbox/role-operator.jpg",
  expert: "/sandbox/role-user.jpg",
  skeptic: "/sandbox/role-competitor.jpg",
  partner: "/sandbox/role-channel.jpg"
};

function sandboxRoleArtwork(role: SandboxRole) {
  return roleArtwork[role.role_code] ?? roleArtwork.customer;
}

type SandboxRoleCardProps = {
  compact?: boolean;
  onToggle?: (role: SandboxRole) => void;
  role: SandboxRole;
  selected?: boolean;
};

function SandboxRoleCard({ compact = false, onToggle, role, selected = false }: SandboxRoleCardProps) {
  const content = (
    <>
      <span className="sb-role-check" aria-hidden="true">{selected ? <Check size={15} strokeWidth={3} /> : null}</span>
      <img alt="" className="sb-role-art" src={sandboxRoleArtwork(role)} />
      <strong>{role.display_name}</strong>
      <small>{role.description}</small>
      {!compact && <em>{role.is_required ? "必须选择" : role.default_selected ? "默认推荐" : "可选角色"}</em>}
    </>
  );

  if (!onToggle) {
    return <article className={`sb-role-card ${selected ? "is-selected" : ""} ${compact ? "is-compact" : ""}`}>{content}</article>;
  }

  return (
    <button
      aria-pressed={selected}
      className={`sb-role-card ${selected ? "is-selected" : ""} ${compact ? "is-compact" : ""}`}
      onClick={() => onToggle(role)}
      type="button"
    >
      {content}
    </button>
  );
}

export default SandboxRoleCard;
