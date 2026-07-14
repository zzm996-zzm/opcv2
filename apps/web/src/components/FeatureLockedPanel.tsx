import { Link } from "react-router-dom";

import type { FeatureAccess } from "../lib/membershipApi";

type FeatureLockedPanelProps = {
  feature: FeatureAccess;
  title?: string;
  description?: string;
};

function FeatureLockedPanel({ feature, title, description }: FeatureLockedPanelProps) {
  return (
    <section className="module-overview-card feature-locked-panel" aria-label={`${feature.label}暂未开放`}>
      <div className="module-overview-copy">
        <span className="module-kicker">功能未开放 · {feature.required_plan === "pro" ? "专业版" : "升级开通"}</span>
        <h2>{title || `${feature.label}当前版本暂未开放`}</h2>
        <p>{description || feature.message || "该模块当前只开放入口展示，真实工作流暂未启用。"}</p>
        <p className="form-success" role="status">当前不会读取业务数据，也不会创建任务、客户或分析请求。</p>
        <div className="module-chip-row compact">
          {feature.upgrade_url && <Link to={feature.upgrade_url}>{feature.cta || "查看升级方案"}</Link>}
          {feature.contact_url && <Link to={feature.contact_url}>预约企业顾问</Link>}
        </div>
      </div>
    </section>
  );
}

export default FeatureLockedPanel;
