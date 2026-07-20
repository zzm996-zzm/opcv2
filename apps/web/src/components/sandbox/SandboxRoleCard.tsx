import { Check } from "lucide-react";

import type { SandboxRole } from "../../lib/sandboxApi";

const roleArtwork: Record<string, string> = {
  user: "/sandbox/role-user.jpg",
  investor: "/sandbox/role-investor.jpg",
  channel: "/sandbox/role-channel.jpg",
  competitor: "/sandbox/role-competitor.jpg",
  operator: "/sandbox/role-operator.jpg"
};

export function sandboxRoleKey(role: Pick<SandboxRole, "key" | "label">) {
  if (roleArtwork[role.key]) return role.key;
  if (role.label.includes("投资")) return "investor";
  if (role.label.includes("代理") || role.label.includes("渠道")) return "channel";
  if (role.label.includes("竞争")) return "competitor";
  if (role.label.includes("运营")) return "operator";
  return "user";
}

export function sandboxRoleArtwork(role: Pick<SandboxRole, "key" | "label">) {
  return roleArtwork[sandboxRoleKey(role)] ?? roleArtwork.user;
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
      <strong>{role.label}</strong>
      <small>{role.description}</small>
      {!compact && <em>{role.badge}</em>}
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

