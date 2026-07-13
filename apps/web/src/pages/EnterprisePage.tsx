import { useEffect, useRef, useState } from "react";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import {
  enterpriseApi,
  type EnterpriseContactConfig,
  type EnterprisePublicCase,
  type EnterprisePublicOverview
} from "../lib/enterpriseApi";

function EnterprisePage() {
  const enterpriseNeedRef = useRef<HTMLTextAreaElement | null>(null);
  const [overview, setOverview] = useState<EnterprisePublicOverview | null>(null);
  const [cases, setCases] = useState<EnterprisePublicCase[]>([]);
  const [contactConfig, setContactConfig] = useState<EnterpriseContactConfig | null>(null);
  const [loadError, setLoadError] = useState("");
  const [caseLoadError, setCaseLoadError] = useState("");
  const [contactLoadError, setContactLoadError] = useState("");
  const [company, setCompany] = useState("");
  const [name, setName] = useState("");
  const [contact, setContact] = useState("");
  const [need, setNeed] = useState("");
  const [budget, setBudget] = useState("");
  const [timeline, setTimeline] = useState("");
  const [submitStatus, setSubmitStatus] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    enterpriseApi
      .publicOverview()
      .then((payload) => {
        if (!active) return;
        setOverview(payload);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setOverview(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取企业服务介绍"));
      });
    enterpriseApi
      .publicCases(6)
      .then((payload) => {
        if (!active) return;
        setCases(payload.cases ?? []);
        setCaseLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setCases([]);
        setCaseLoadError(apiErrorMessage(error, "暂时无法读取企业案例"));
      });
    enterpriseApi
      .contactConfig()
      .then((payload) => {
        if (!active) return;
        setContactConfig(payload);
        setContactLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setContactConfig(null);
        setContactLoadError(apiErrorMessage(error, "暂时无法读取顾问联系方式"));
      });
    return () => {
      active = false;
    };
  }, []);

  const proofPoints = overview?.proof_points ?? [];
  const stats = overview?.stats ?? [];
  const serviceSteps = overview?.service_steps ?? [];
  const hasOverviewData = Boolean(overview?.headline || overview?.description || proofPoints.length || stats.length || serviceSteps.length);

  const focusEnterpriseNeed = () => {
    if (typeof enterpriseNeedRef.current?.scrollIntoView === "function") {
      enterpriseNeedRef.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    enterpriseNeedRef.current?.focus();
  };

  const submitInquiry = async () => {
    const normalizedName = name.trim();
    const normalizedContact = contact.trim();
    const normalizedNeed = need.trim();
    if (!normalizedName || !normalizedContact || !normalizedNeed) {
      setSubmitStatus("");
      setSubmitError("请填写联系人、联系方式和企业需求");
      return;
    }
    setSubmitting(true);
    setSubmitStatus("");
    setSubmitError("");
    try {
      const inquiry = await enterpriseApi.createInquiry({
        company: company.trim(),
        name: normalizedName,
        phone: looksLikePhone(normalizedContact) ? normalizedContact : undefined,
        email: looksLikeEmail(normalizedContact) ? normalizedContact : undefined,
        wechat: !looksLikePhone(normalizedContact) && !looksLikeEmail(normalizedContact) ? normalizedContact : undefined,
        need: normalizedNeed,
        budget: budget.trim(),
        timeline: timeline.trim(),
        source_page: "/enterprise"
      });
      setSubmitStatus(inquiry.crm_customer_id ? "咨询已提交，顾问将跟进联系" : "咨询已提交，等待顾问配置后跟进");
      setCompany("");
      setName("");
      setContact("");
      setNeed("");
      setBudget("");
      setTimeline("");
    } catch (error) {
      setSubmitError(apiErrorMessage(error, "暂时无法提交企业咨询"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <V4PageShell className="enterprise-shell">
      <section className="module-page enterprise-page" aria-label="企业定制化陪跑">
        <div className="page-title-row">
          <div>
            <h1>企业定制化陪跑</h1>
            <p>{overview?.subheadline || "面向企业团队提供诊断、方案、系统搭建、训练和复盘的一体化增长陪跑"}</p>
          </div>
          <button className="module-primary-action" onClick={focusEnterpriseNeed} type="button">预约企业咨询</button>
        </div>

        <section className="module-overview-card enterprise-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">企业服务 · 公开介绍</span>
            <h2>{overview?.headline || "企业服务内容待发布"}</h2>
            <p>{overview?.description || "后台发布企业服务介绍后，这里会展示真实的服务范围、交付方式和来源说明。"}</p>
            {overview?.source_name && <p className="form-success" role="note">来源：{overview.source_name}{overview.source_updated_at ? ` · ${formatDate(overview.source_updated_at)}` : ""}</p>}
            {loadError && <p className="form-error" role="alert">{loadError}</p>}
            {!loadError && !hasOverviewData && <p className="form-success" role="status">暂无已发布企业服务介绍。</p>}
            <div className="module-stat-strip">
              {stats.length === 0 ? (
                <article>
                  <small>公开指标</small>
                  <strong>待发布</strong>
                </article>
              ) : stats.map((item) => (
                <article key={item.key}>
                  <small>{item.label}</small>
                  <strong>{item.value}</strong>
                  {item.note && <small>{item.note}</small>}
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact enterprise-diagnosis-card">
            <label htmlFor="enterprise-company">企业名称</label>
            <input
              id="enterprise-company"
              aria-label="企业名称"
              placeholder="例如：启明星教育"
              value={company}
              onChange={(event) => setCompany(event.target.value)}
            />
            <label htmlFor="enterprise-contact-name">联系人</label>
            <input
              id="enterprise-contact-name"
              aria-label="联系人"
              placeholder="你的姓名"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
            <label htmlFor="enterprise-contact">联系方式</label>
            <input
              id="enterprise-contact"
              aria-label="联系方式"
              placeholder="手机号、邮箱或微信"
              value={contact}
              onChange={(event) => setContact(event.target.value)}
            />
            <label htmlFor="enterprise-need">描述企业需求</label>
            <textarea
              id="enterprise-need"
              ref={enterpriseNeedRef}
              aria-label="描述企业需求"
              placeholder="例如：30人销售团队，希望用 AI 提升线索开发、客户跟进和经营复盘效率..."
              value={need}
              onChange={(event) => setNeed(event.target.value)}
            />
            <label htmlFor="enterprise-budget">预算范围</label>
            <input
              id="enterprise-budget"
              aria-label="预算范围"
              placeholder="可选"
              value={budget}
              onChange={(event) => setBudget(event.target.value)}
            />
            <label htmlFor="enterprise-timeline">启动时间</label>
            <input
              id="enterprise-timeline"
              aria-label="启动时间"
              placeholder="可选"
              value={timeline}
              onChange={(event) => setTimeline(event.target.value)}
            />
            {submitStatus && <p className="form-success" role="status">{submitStatus}</p>}
            {submitError && <p className="form-error" role="alert">{submitError}</p>}
            <button type="button" onClick={submitInquiry} disabled={submitting}>
              {submitting ? "提交中..." : "提交企业咨询"}
            </button>
          </form>
        </section>

        <section className="enterprise-workbench">
          <div className="enterprise-plan-card">
            <div className="module-section-head">
              <div>
                <h2>服务路径</h2>
                <p>后台发布后展示真实交付步骤；未发布时保持空态。</p>
              </div>
            </div>

            <div className="enterprise-plan-grid">
              {serviceSteps.length === 0 && <p>暂无已发布服务路径</p>}
              {serviceSteps.map((step) => (
                <article key={step.title}>
                  <header>
                    <div>
                      <h3>{step.title}</h3>
                    </div>
                  </header>
                  <p>{step.detail || "暂无说明"}</p>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-delivery-card" aria-label="顾问联系方式">
            <h2>顾问联系方式</h2>
            {contactLoadError && <p className="form-error" role="alert">{contactLoadError}</p>}
            {!contactLoadError && !hasContactConfig(contactConfig) && <p>暂无已发布顾问联系方式</p>}
            {contactConfig?.qr_image_url && <img alt={contactConfig.consultant_name || "企业顾问二维码"} src={contactConfig.qr_image_url} />}
            {contactConfig?.consultant_name && <article><span><strong>{contactConfig.consultant_name}</strong><small>{contactConfig.title || "企业服务顾问"}</small></span></article>}
            {contactConfig?.description && <article><span><strong>说明</strong><small>{contactConfig.description}</small></span></article>}
            {contactConfig?.phone && <article><span><strong>电话</strong><small>{contactConfig.phone}</small></span></article>}
            {contactConfig?.email && <article><span><strong>邮箱</strong><small>{contactConfig.email}</small></span></article>}
            {contactConfig?.wechat && <article><span><strong>微信</strong><small>{contactConfig.wechat}</small></span></article>}
            {contactConfig?.contact_url && <a href={contactConfig.contact_url}>打开预约链接</a>}
          </aside>
        </section>

        <section className="enterprise-lower-grid">
          <div className="enterprise-milestone-card">
            <div className="module-section-head">
              <div>
                <h2>服务依据</h2>
                <p>只展示后台发布的公开说明，不使用静态营销结论。</p>
              </div>
            </div>
            <div className="enterprise-milestones">
              {proofPoints.length === 0 && <p>暂无已发布服务依据</p>}
              {proofPoints.map((item) => (
                <article key={item.title}>
                  <b>依据</b>
                  <strong>{item.title}</strong>
                  <small>{item.detail || "暂无说明"}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-case-card" aria-label="企业案例">
            <h2>企业案例</h2>
            {caseLoadError && <p className="form-error" role="alert">{caseLoadError}</p>}
            {!caseLoadError && cases.length === 0 && <p>暂无已发布企业案例</p>}
            {cases.map((item) => (
              <article key={item.id}>
                <strong>{item.title || item.company}</strong>
                <small>{item.company}{item.industry ? ` · ${item.industry}` : ""}</small>
                {item.summary && <small>{item.summary}</small>}
                {item.result && <small>{item.result}</small>}
                {item.services.length > 0 && (
                  <div className="tool-tags">
                    {item.services.map((service) => <span key={service}>{service}</span>)}
                  </div>
                )}
                {item.source_name && <small>来源：{item.source_name}</small>}
              </article>
            ))}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

function looksLikePhone(value: string) {
  return /^[+\d][\d\s-]{5,}$/.test(value);
}

function looksLikeEmail(value: string) {
  return value.includes("@");
}

function hasContactConfig(config: EnterpriseContactConfig | null) {
  return Boolean(config?.consultant_name || config?.phone || config?.email || config?.wechat || config?.qr_image_url || config?.contact_url);
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

export default EnterprisePage;
