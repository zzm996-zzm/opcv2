import { useEffect, useRef, useState } from "react";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { enterpriseApi, type EnterpriseOverview } from "../lib/enterpriseApi";

function EnterprisePage() {
  const enterpriseNeedRef = useRef<HTMLTextAreaElement | null>(null);
  const [overview, setOverview] = useState<EnterpriseOverview | null>(null);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let active = true;
    enterpriseApi
      .overview()
      .then((payload) => {
        if (!active) return;
        setOverview(payload);
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setOverview(null);
        setLoadError(apiErrorMessage(error, "暂时无法读取企业陪跑数据"));
      });
    return () => {
      active = false;
    };
  }, []);

  const stats = overview?.stats ?? [];
  const plans = overview?.plans ?? [];
  const deliveryBoard = overview?.delivery_board ?? [];
  const milestones = overview?.milestones ?? [];
  const cases = overview?.cases ?? [];
  const hasOverviewData = stats.length > 0 || plans.length > 0 || deliveryBoard.length > 0 || milestones.length > 0 || cases.length > 0;

  const focusEnterpriseNeed = () => {
    if (typeof enterpriseNeedRef.current?.scrollIntoView === "function") {
      enterpriseNeedRef.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    enterpriseNeedRef.current?.focus();
  };

  return (
    <V4PageShell className="enterprise-shell">
      <section className="module-page enterprise-page" aria-label="企业定制化陪跑">
        <div className="page-title-row">
          <div>
            <h1>企业定制化陪跑</h1>
            <p>面向企业团队提供诊断、方案、系统搭建、训练和复盘的一体化增长陪跑</p>
          </div>
          <button className="module-primary-action" onClick={focusEnterpriseNeed} type="button">预约企业诊断</button>
        </div>

        <section className="module-overview-card enterprise-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">企业服务 · 定制交付</span>
            <h2>从业务问题到团队上线，陪企业把 AI 增长流程真正跑起来</h2>
            <p>通过企业诊断、工具配置、实战陪跑和交付验收，把项目超市、GEO 获客、AI 线索开发和 CRM 组合成企业自己的增长系统。</p>
            {loadError && <p className="form-error" role="alert">{loadError}</p>}
            {!loadError && !hasOverviewData && (
              <p className="form-success" role="status">暂无企业陪跑概览数据，提交一次诊断需求后将逐步沉淀方案、交付和案例数据。</p>
            )}
            <div className="module-stat-strip">
              {stats.length === 0 ? (
                <>
                  <article>
                    <small>服务企业数</small>
                    <strong>0</strong>
                  </article>
                  <article>
                    <small>平均周期</small>
                    <strong>未接入</strong>
                  </article>
                  <article>
                    <small>交付任务数</small>
                    <strong>0</strong>
                  </article>
                  <article>
                    <small>续约率数据</small>
                    <strong>未接入</strong>
                  </article>
                </>
              ) : stats.map((item) => (
                <article key={item.key}>
                  <small>{item.label}</small>
                  <strong>{item.value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact enterprise-diagnosis-card">
            <label htmlFor="enterprise-need">描述企业需求</label>
            <textarea
              id="enterprise-need"
              ref={enterpriseNeedRef}
              aria-label="描述企业需求"
              placeholder="例如：30人销售团队，希望用 AI 提升线索开发、客户跟进和经营复盘效率..."
            />
            <button type="button">生成诊断提纲</button>
          </form>
        </section>

        <section className="enterprise-workbench">
          <div className="enterprise-plan-card">
            <div className="module-section-head">
              <div>
                <h2>陪跑方案</h2>
                <p>按企业团队规模、增长目标和系统复杂度选择交付模式</p>
              </div>
              <div className="module-chip-row compact">
                {["标准", "获客", "系统", "定制"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>

            <div className="enterprise-plan-grid">
              {plans.length === 0 && <p>暂无陪跑方案</p>}
              {plans.map((plan) => (
                <article key={plan.id}>
                  <header>
                    <div>
                      <h3>{plan.title}</h3>
                      <small>{plan.audience || "未标注适用对象"}</small>
                    </div>
                    <strong>{plan.price_label || "未报价"}</strong>
                  </header>
                  <p>{plan.result || "暂无交付结果"}</p>
                  <div className="tool-tags">
                    {plan.focus.map((item) => <span key={item}>{item}</span>)}
                  </div>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-delivery-card" aria-label="交付看板">
            <h2>交付看板</h2>
            {deliveryBoard.length === 0 && <p>暂无交付看板数据</p>}
            {deliveryBoard.map((item) => (
              <article key={item.stage}>
                <span>
                  <strong>{item.stage}</strong>
                  <small>{item.detail || "暂无说明"}</small>
                </span>
                <em>{item.count}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="enterprise-lower-grid">
          <div className="enterprise-milestone-card">
            <div className="module-section-head">
              <div>
                <h2>陪跑里程碑</h2>
                <p>把企业服务拆成可验收、可复盘、可续约的交付节奏</p>
              </div>
            </div>
            <div className="enterprise-milestones">
              {milestones.length === 0 && <p>暂无陪跑里程碑</p>}
              {milestones.map((item) => (
                <article key={`${item.time_label}-${item.title}`}>
                  <b>{item.time_label}</b>
                  <strong>{item.title}</strong>
                  <small>{item.detail || "暂无说明"}</small>
                </article>
              ))}
            </div>
          </div>

          <aside className="enterprise-case-card" aria-label="企业案例">
            <h2>企业案例</h2>
            {cases.length === 0 && <p>暂无企业案例</p>}
            {cases.map((item) => (
              <article key={item.id}>
                <strong>{item.company}</strong>
                <small>{item.result || "暂无案例结果"}</small>
              </article>
            ))}
          </aside>
        </section>
      </section>
    </V4PageShell>
  );
}

export default EnterprisePage;
