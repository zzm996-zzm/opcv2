import { ArrowLeft, Check, Edit3, Rocket, Users } from "lucide-react";
import { Link } from "react-router-dom";
import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxRoleCard from "../../components/sandbox/SandboxRoleCard";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import type { SandboxRole, SandboxRun } from "../../lib/sandboxApi";

type Props = { onStart: () => Promise<void>; roles: SandboxRole[]; run: SandboxRun };
export default function SandboxStartView({ onStart, roles, run }: Props) {
  return <SandboxFrame copilotMode="start" copilotProject={run.product.name} copilotRunID={run.id}><SandboxStepper active={4} /><section className="sb-work-panel sb-start-panel"><div className="sb-start-columns"><section className="sb-start-brief"><header><h1>确认项目信息</h1><Link to={`/sandbox/new?run=${run.id}&step=2`}><Edit3 size={15} />编辑</Link></header><dl><div><dt>产品 / 生意</dt><dd>{run.product.name}</dd></div><div><dt>卖点</dt><dd>{run.product.selling_point || "未补充"}</dd></div><div><dt>目标客户</dt><dd>{run.context.target_customer || "未补充"}</dd></div><div><dt>渠道</dt><dd>{run.context.channel || "未补充"}</dd></div><div><dt>假设</dt><dd>{run.assumptions?.length ? run.assumptions.join("；") : "暂无"}</dd></div></dl></section><section className="sb-start-roles"><header><h1><Users size={18} />参与角色 <small>已选 {run.roles.length} 个</small></h1></header><div className="sb-start-role-grid">{run.roles.map((code) => { const role = roles.find((item) => item.role_code === code); return role ? <SandboxRoleCard compact key={code} role={role} selected /> : null; })}</div></section></div><footer className="sb-flow-footer sb-start-footer"><Link className="is-quiet" to={`/sandbox/new?run=${run.id}&step=3`}><ArrowLeft size={17} />上一步</Link><div className="sb-start-assurance"><Check size={15} /><span>结果会标注“模型推演”，不替代真实市场验证。</span></div><button className="is-primary" disabled={run.roles.length < 3 || !run.roles.includes("skeptic")} onClick={() => void onStart()} type="button">开始推演<Rocket size={18} /></button></footer></section></SandboxFrame>;
}
