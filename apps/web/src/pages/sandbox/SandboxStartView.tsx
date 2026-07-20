import { ArrowLeft, ArrowRight, Check, ChevronDown, Edit3, FileText, Gauge, ListTree, Rocket, SlidersHorizontal, Users } from "lucide-react";
import { useMemo, useState } from "react";
import { Link } from "react-router-dom";

import SandboxFrame from "../../components/sandbox/SandboxFrame";
import SandboxQuotaDialog from "../../components/sandbox/SandboxQuotaDialog";
import SandboxRoleCard from "../../components/sandbox/SandboxRoleCard";
import SandboxStepper from "../../components/sandbox/SandboxStepper";
import { ApiRequestError } from "../../lib/apiRequest";
import type { MembershipUsageItem } from "../../lib/membershipApi";
import type { SandboxOptions, SandboxRole, SandboxRunSettings, SandboxSession } from "../../lib/sandboxApi";

type SandboxStartViewProps = {
  onStart: (settings: SandboxRunSettings) => Promise<void>;
  options: SandboxOptions;
  session: SandboxSession;
  usage?: MembershipUsageItem;
};

function SandboxStartView({ onStart, options, session, usage }: SandboxStartViewProps) {
  const [settings, setSettings] = useState<SandboxRunSettings>(session.settings ?? options.defaults);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [starting, setStarting] = useState(false);
  const [quotaOpen, setQuotaOpen] = useState(false);
  const [error, setError] = useState("");
  const selectedRoles = useMemo(() => session.roles.map((label) => options.roles.find((role) => role.label === label) ?? fallbackRole(label)), [options.roles, session.roles]);
  const blocked = usage ? usage.limit <= 0 || usage.used >= usage.limit : false;

  function setVariable(key: string, value: string) {
    setSettings((current) => ({ ...current, variables: { ...current.variables, [key]: value } }));
  }

  async function start() {
    if (starting) return;
    if (blocked) {
      setQuotaOpen(true);
      return;
    }
    setStarting(true);
    setError("");
    try {
      await onStart(settings);
    } catch (requestError) {
      if (requestError instanceof ApiRequestError && requestError.status === 402) {
        setQuotaOpen(true);
      } else {
        setError(requestError instanceof Error ? requestError.message : "推演启动失败，请稍后重试。");
      }
    } finally {
      setStarting(false);
    }
  }

  return (
    <SandboxFrame copilotMode="start" copilotProgress={starting ? 60 : undefined} copilotProject={session.intake?.initial_idea || session.goal}>
      <SandboxStepper active={3} />
      <section className="sb-work-panel sb-start-panel">
        <div className="sb-start-columns">
          <section className="sb-start-brief">
            <header><h1><ListTree size={18} />推演设置信息</h1><Link to={`/sandbox/setup?session=${session.id}`}><Edit3 size={15} />编辑</Link></header>
            <dl>
              <div><dt>项目设想</dt><dd>{session.intake?.initial_idea || session.goal}</dd></div>
              {session.intake?.recognized_fields?.map((field) => <div key={field.key}><dt>{field.label}</dt><dd>{field.value}</dd></div>)}
            </dl>
            <Link className="sb-inline-link" to={`/sandbox/setup?session=${session.id}`}>查看完整信息<ArrowRight size={14} /></Link>
          </section>
          <section className="sb-start-roles">
            <header><h1><Users size={18} />参与推演角色 <small>已选 {selectedRoles.length} 个</small></h1></header>
            <div className="sb-start-role-grid">
              {selectedRoles.map((role) => <SandboxRoleCard compact key={role.label} role={role} selected />)}
            </div>
          </section>
        </div>
        <section className="sb-settings-section">
          <h2>推演设置</h2>
          <div className="sb-setting-grid">
            <label>
              <span><Gauge size={18} /><b>推演深度</b></span>
              <div className="sb-select-wrap">
                <select aria-label="推演深度" onChange={(event) => setSettings((current) => ({ ...current, depth: event.target.value as SandboxRunSettings["depth"] }))} value={settings.depth}>
                  {options.depths.map((option) => <option key={option.value} value={option.value}>{option.label}{option.recommended ? "（推荐）" : ""}</option>)}
                </select>
                <ChevronDown size={15} />
              </div>
              <small>{options.depths.find((option) => option.value === settings.depth)?.description}</small>
            </label>
            <label>
              <span><FileText size={18} /><b>输出风格</b></span>
              <div className="sb-select-wrap">
                <select aria-label="输出风格" onChange={(event) => setSettings((current) => ({ ...current, output_style: event.target.value as SandboxRunSettings["output_style"] }))} value={settings.output_style}>
                  {options.output_styles.map((option) => <option key={option.value} value={option.value}>{option.label}{option.recommended ? "（推荐）" : ""}</option>)}
                </select>
                <ChevronDown size={15} />
              </div>
              <small>{options.output_styles.find((option) => option.value === settings.output_style)?.description}</small>
            </label>
            <label className="sb-toggle-setting">
              <span><ListTree size={18} /><b>生成推演大纲</b></span>
              <button aria-pressed={settings.generate_outline} className={settings.generate_outline ? "is-on" : ""} onClick={() => setSettings((current) => ({ ...current, generate_outline: !current.generate_outline }))} type="button"><span /></button>
              <small>先生成大纲，确认后再深入推演</small>
            </label>
            <button aria-expanded={advancedOpen} className="sb-advanced-toggle" onClick={() => setAdvancedOpen((open) => !open)} type="button"><SlidersHorizontal size={17} />高级设置<ArrowRight size={15} /></button>
          </div>
          {advancedOpen && (
            <div className="sb-advanced-settings">
              <label><span>定价策略</span><input onChange={(event) => setVariable("pricing", event.target.value)} placeholder="例如：中档订阅" value={settings.variables.pricing ?? ""} /></label>
              <label><span>获客渠道</span><input onChange={(event) => setVariable("acquisition", event.target.value)} placeholder="例如：内容营销 + 渠道合作" value={settings.variables.acquisition ?? ""} /></label>
              <label><span>目标客群</span><input onChange={(event) => setVariable("audience", event.target.value)} placeholder="例如：10-200 人规模企业" value={settings.variables.audience ?? ""} /></label>
            </div>
          )}
        </section>
        <footer className="sb-flow-footer sb-start-footer">
          <Link className="is-quiet" to={`/sandbox/roles?session=${session.id}`}><ArrowLeft size={17} />上一步</Link>
          <div className="sb-start-assurance"><Check size={15} /><span>推演结果是基于输入条件的 AI 情景分析，关键结论仍需真实验证。</span></div>
          <button className="is-primary" disabled={starting || !session.roles.length} onClick={() => void start()} type="button">
            {starting ? "准备推演中..." : "开始推演"}<Rocket size={18} />
          </button>
        </footer>
        {error ? <p className="sb-inline-error" role="alert">{error}</p> : null}
      </section>
      {quotaOpen ? <SandboxQuotaDialog limit={usage?.limit ?? 1} onClose={() => setQuotaOpen(false)} used={usage?.used ?? usage?.limit ?? 1} /> : null}
    </SandboxFrame>
  );
}

function fallbackRole(label: string): SandboxRole {
  return { key: "user", label, description: "参与本轮多角色推演", badge: "已选择" };
}

export default SandboxStartView;
