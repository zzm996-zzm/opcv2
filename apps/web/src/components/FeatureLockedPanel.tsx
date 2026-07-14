import { Link } from "react-router-dom";

import type { FeatureAccess } from "../lib/membershipApi";

type FeatureLockedPanelProps = {
  feature: FeatureAccess;
  title?: string;
  description?: string;
  eyebrow?: string;
  accent?: string;
  highlights?: Array<{
    label: string;
    value: string;
    detail: string;
  }>;
  workflow?: Array<{
    step: string;
    title: string;
    detail: string;
  }>;
  preview?: Array<{
    title: string;
    detail: string;
    tag: string;
  }>;
};

function FeatureLockedPanel({ accent = "产品预告", eyebrow, feature, highlights = [], preview = [], title, description, workflow = [] }: FeatureLockedPanelProps) {
  return (
    <section className="feature-locked-panel" aria-label={`${feature.label}暂未开放`}>
      <div className="feature-locked-aurora" aria-hidden="true" />
      <div className="feature-locked-hero">
        <div className="feature-locked-copy">
          <span className="feature-locked-kicker">{eyebrow || `${accent} · ${feature.required_plan === "pro" ? "专业版候补开放" : "升级开通"}`}</span>
          <h2>{title || `${feature.label}当前版本暂未开放`}</h2>
          <p>{description || feature.message || "该模块当前只开放入口展示，真实工作流暂未启用。"}</p>
          <p className="feature-locked-policy" role="status">当前不会读取业务数据，也不会创建任务、客户或分析请求。</p>
          <div className="feature-locked-actions">
            {feature.upgrade_url && <Link className="feature-locked-primary" to={feature.upgrade_url}>{feature.cta || "查看升级方案"}</Link>}
            {feature.contact_url && <Link className="feature-locked-secondary" to={feature.contact_url}>预约企业顾问</Link>}
          </div>
        </div>
        <div className="feature-locked-device" aria-hidden="true">
          <div className="feature-device-top">
            <span />
            <span />
            <span />
          </div>
          <div className="feature-device-orbit">
            <strong>{feature.label}</strong>
            <small>LOCKED PREVIEW</small>
          </div>
          <div className="feature-device-lines">
            <i />
            <i />
            <i />
          </div>
        </div>
      </div>

      {highlights.length > 0 ? (
        <div className="feature-locked-highlight-grid">
          {highlights.map((item) => (
            <article key={item.label}>
              <small>{item.label}</small>
              <strong>{item.value}</strong>
              <span>{item.detail}</span>
            </article>
          ))}
        </div>
      ) : null}

      <div className="feature-locked-lower">
        {workflow.length > 0 ? (
          <section className="feature-locked-workflow" aria-label="开放后工作流">
            <h3>开放后工作流</h3>
            <div>
              {workflow.map((item) => (
                <article key={item.step}>
                  <span>{item.step}</span>
                  <div>
                    <strong>{item.title}</strong>
                    <p>{item.detail}</p>
                  </div>
                </article>
              ))}
            </div>
          </section>
        ) : null}

        {preview.length > 0 ? (
          <section className="feature-locked-preview" aria-label="能力预览">
            <h3>能力预览</h3>
            {preview.map((item) => (
              <article key={item.title}>
                <span>{item.tag}</span>
                <div>
                  <strong>{item.title}</strong>
                  <p>{item.detail}</p>
                </div>
              </article>
            ))}
          </section>
        ) : null}
      </div>
    </section>
  );
}

export default FeatureLockedPanel;
