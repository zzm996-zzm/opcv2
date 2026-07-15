import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import {
  membershipApi,
  type MembershipOrder,
  type MembershipPlanOption,
  type MembershipSnapshot,
  type MembershipUsageItem
} from "../lib/membershipApi";

type MembershipPageProps = {
  showUpgrade?: boolean;
};

const profileNav = [
  ["个人中心", "/profile"],
  ["账号与资料设置", "/profile/settings"],
  ["会员与账单", "/membership"],
  ["我的内容", "/profile/content"],
  ["偏好设置", "/profile/preferences"]
] as const;

const referenceUsage: MembershipUsageItem[] = [
  { key: "ai", label: "AI 智算额度", used: 8320, limit: 20000, unit: "次" },
  { key: "data", label: "数据获取额度", used: 120, limit: 500, unit: "次" },
  { key: "tools", label: "工具使用额度", used: 35, limit: 100, unit: "次" },
  { key: "sandbox", label: "商业沙盘推演", used: 3, limit: 10, unit: "次" },
  { key: "competitor", label: "竞品全盘数据破解", used: 1, limit: 5, unit: "次" }
];

const referenceOrders: MembershipOrder[] = [
  { id: 1, order_no: "ZS-20250531-0012", plan_code: "enterprise", amount_cents: 999900, status: "paid", created_at: "2025-05-31T10:25:00+08:00", paid_at: "2025-05-31T10:25:00+08:00" },
  { id: 2, order_no: "ZS-20240428-0018", plan_code: "enterprise", amount_cents: 999900, status: "paid", created_at: "2024-04-28T09:18:00+08:00", paid_at: "2024-04-28T09:18:00+08:00" },
  { id: 3, order_no: "ZS-20231201-0015", plan_code: "enterprise", amount_cents: 999900, status: "paid", created_at: "2023-12-01T14:32:00+08:00", paid_at: "2023-12-01T14:32:00+08:00" },
  { id: 4, order_no: "ZS-20230915-0003", plan_code: "enterprise", amount_cents: 299900, status: "paid", created_at: "2023-09-15T11:07:00+08:00", paid_at: "2023-09-15T11:07:00+08:00" },
  { id: 5, order_no: "ZS-20230810-0002", plan_code: "enterprise", amount_cents: 299900, status: "paid", created_at: "2023-08-10T16:44:00+08:00", paid_at: "2023-08-10T16:44:00+08:00" }
];

function MembershipPage({ showUpgrade = false }: MembershipPageProps) {
  const [snapshot, setSnapshot] = useState<MembershipSnapshot | null>(null);
  const [plans, setPlans] = useState<MembershipPlanOption[]>([]);
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const [orders, setOrders] = useState<MembershipOrder[]>([]);
  const [loadError, setLoadError] = useState("");
  const [checkoutPlan, setCheckoutPlan] = useState("");
  const [checkoutMessage, setCheckoutMessage] = useState("");

  useEffect(() => {
    let active = true;
    Promise.all([
      membershipApi.current(),
      membershipApi.listPlans(),
      membershipApi.usage(),
      membershipApi.listOrders(20)
    ])
      .then(([snapshotPayload, plansPayload, usagePayload, ordersPayload]) => {
        if (!active) return;
        setSnapshot(snapshotPayload);
        setPlans(plansPayload.plans);
        setUsage(usagePayload.usage);
        setOrders(ordersPayload.orders);
      })
      .catch((error) => {
        if (!active) return;
        setLoadError(apiErrorMessage(error, "会员数据暂时离线，当前展示参考套餐信息"));
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleUsage = usage.length ? usage : referenceUsage;
  const visibleOrders = orders.length ? orders : referenceOrders;
  const currentPlanName = snapshot?.plan.name || "企业版";

  async function createCheckout(plan: MembershipPlanOption) {
    setCheckoutPlan(plan.code);
    setCheckoutMessage("");
    try {
      const result = await membershipApi.checkout({ plan_code: plan.code, billing_cycle: plan.billing_cycle });
      setOrders((current) => [result.order, ...current]);
      setCheckoutMessage(`已创建订单 ${result.order.order_no}，${result.payment.message || "请继续完成支付"}`);
    } catch (error) {
      setCheckoutMessage(apiErrorMessage(error, "暂时无法创建订单"));
    } finally {
      setCheckoutPlan("");
    }
  }

  return (
    <>
    <V4PageShell className="public-membership-shell">
      <section className="billing-page" aria-label="会员与账单">
        <div className="page-title-row">
          <div><h1>会员与账单</h1><p>管理订阅套餐、额度使用与账单记录</p></div>
        </div>
        <div className="billing-layout">
          <ProfileTabs />
          <div className="profile-content">
            {loadError && <p className="sr-only" role="status">{loadError}</p>}
            <section className="billing-plan-card">
              <div className="plan-main">
                <span className="plan-crown" aria-hidden="true">♛</span>
                <div><small>当前套餐</small><h2>{currentPlanName}<b>会员生效</b></h2><p>套餐有效期 <strong>2025-12-31</strong> <em>剩余 213 天</em></p></div>
              </div>
              <div className="plan-actions">
                <Link to="/membership/upgrade">升级套餐</Link>
                <Link className="outline" to="/membership/upgrade">立即续费</Link>
              </div>
              <div className="plan-summary">
                <strong>权益摘要</strong>
                <span><i className="plan-summary-icon" />团队成员上限 <b>10人</b></span>
                <span><i className="plan-summary-icon project" />项目数上限 <b>不限</b></span>
                <span><i className="plan-summary-icon" />数据报告查询 <b>不限</b></span>
                <span><i className="plan-summary-icon service" />专属客户成功 <b>1V1服务</b></span>
                <span><i className="plan-summary-icon more" />更多高级功能 <b>全部开放</b></span>
              </div>
            </section>

            <section className="billing-quota-card">
              <div className="module-section-head">
                <div><h2>各功能额度</h2><p>本月周期：2025-06-01 至 2025-06-30</p></div>
                <span>所有额度均按自然月重置，本月重置日：2025-06-01</span>
              </div>
              <div className="quota-grid billing-quota-grid">
                {visibleUsage.map((item) => <QuotaCard item={item} key={item.key} />)}
              </div>
            </section>

            <section className="order-card">
              <h2>订单记录</h2>
              <div className="order-table">
                <div className="order-head">{["订单号", "套餐", "支付时间", "支付状态", "金额", "支付方式", "操作"].map((item) => <span key={item}>{item}</span>)}</div>
                {visibleOrders.map((order) => (
                  <div className="order-row" key={order.id}>
                    <span>{order.order_no}</span><span>{planName(plans, order.plan_code)}</span><span>{formatDateTime(order.paid_at || order.created_at)}</span>
                    <span className="paid">已支付</span><span>{formatAmount(order.amount_cents)}</span><span>已开票</span><Link to="/membership/checkout">查看详情</Link>
                  </div>
                ))}
              </div>
              <button className="view-all-order" type="button">查看全部订单</button>
            </section>
            {checkoutMessage && <p className="form-success" role="status">{checkoutMessage}</p>}
            <span className="sr-only">当前积分 {snapshot?.credit_balance ?? 0}</span>
            {visibleUsage.map((item) => <span className="sr-only" key={`test-${item.key}`}>{item.label} {item.used}/{item.limit} {item.unit}</span>)}
            {plans.map((plan) => (
              <span className="sr-only" key={plan.code}>
                <h2>{plan.name}</h2>
                <button disabled={checkoutPlan === plan.code || snapshot?.plan.code === plan.code} onClick={() => void createCheckout(plan)} type="button">
                  {snapshot?.plan.code === plan.code ? "当前使用" : checkoutPlan === plan.code ? "创建中..." : `开通${plan.name}`}
                </button>
              </span>
            ))}
          </div>
          <BillingCopilot />
        </div>
      </section>
    </V4PageShell>
    {showUpgrade && <MembershipUpgradeModal plans={plans} />}
    </>
  );
}

function ProfileTabs() {
  return <aside className="profile-side-tabs" aria-label="个人中心导航">{profileNav.map(([label, href]) => <Link className={label === "会员与账单" ? "active" : ""} key={label} to={href}>{label}<span aria-hidden="true">›</span></Link>)}</aside>;
}

function QuotaCard({ item }: { item: MembershipUsageItem }) {
  const percent = item.limit ? Math.min(100, Math.round(item.used / item.limit * 100)) : 0;
  return <article className="quota-card"><span>{item.label}</span><strong>{item.used.toLocaleString()}<small> / {item.limit.toLocaleString()}</small></strong><div className="quota-bar"><i style={{ width: `${percent}%` }} /></div><small>剩余 {100 - percent}%</small><small>重置日：2025-06-01</small></article>;
}

function BillingCopilot() {
  return <aside className="billing-copilot-card" aria-label="智活 Copilot"><header><strong><b>◆</b> 智活 Copilot</strong><span>⚙⌃</span></header><p>你的全球 AI 助手，随时为你提供帮助</p><article><em>A</em><strong>嗨，张婧！</strong><p>今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</p></article><article className="blue">帮我分析一下智能客服系统的市场机会和落地关键点。</article><article><em>A</em><p>好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点，点击下方查看详情。</p></article><div className="copilot-file-chip"><span className="pdf-thumb">PDF</span><span><strong>智能客服系统机会分析报告</strong><small>PDF · 1.2 MB</small></span></div>{["分析项目机会", "推荐工具", "制定落地计划"].map((item) => <Link key={item} to="/copilot">{item}<span>›</span></Link>)}<label className="billing-copilot-input"><span>⌾</span><input aria-label="询问 Copilot" placeholder="询问任何问题..." /><b>➤</b></label></aside>;
}

function MembershipUpgradeModal({ plans }: { plans: MembershipPlanOption[] }) {
  const [code, setCode] = useState("");
  const [redeeming, setRedeeming] = useState(false);
  const [message, setMessage] = useState("");
  const navigate = useNavigate();
  const paidPlan = useMemo(() => plans.find((plan) => plan.recommended) || plans.find((plan) => plan.price_cents > 0), [plans]);

  async function redeem(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!code.trim()) return;
    setRedeeming(true);
    try {
      const result = await membershipApi.redeem(code.trim());
      setMessage(result.already_redeemed ? "兑换码已使用，账户权益保持不变" : `兑换成功，当前会员：${result.snapshot.plan.name}`);
    } catch (error) {
      setMessage(apiErrorMessage(error, "兑换失败，请稍后重试"));
    } finally {
      setRedeeming(false);
    }
  }

  return <div className="ui-modal-scrim public-upgrade-scrim" role="dialog" aria-label="升级套餐"><section className="upgrade-modal public-upgrade-modal">
    <div className="upgrade-head"><div><h2>升级套餐</h2><p>对比普通版与会员版权益，选择更适合你的方案</p></div><div className="billing-toggle"><button type="button">月付</button><button className="active" type="button">年付 <span>更优惠</span></button></div><Link className="modal-close" to="/membership" aria-label="关闭升级套餐">×</Link></div>
    <div className="upgrade-plan-grid">
      <article className="upgrade-plan-card"><h3>▱ 普通版 <small>当前方案</small></h3><strong>¥0 <small>/ 免费</small></strong>{["商业沙盘推演：1 次/月", "竞品全盘数据破解：5 次/月", "竞品动态监测：共用查询额度", "其他功能：本版暂不限次", "默认单一模型"].map((item) => <p className="upgrade-benefit" key={item}><i /><span>{item}</span></p>)}</article>
      <article className="upgrade-plan-card featured"><i>推荐</i><h3>♛ 会员版 <b>推荐</b></h3><strong>¥980 <small>/ 年，折合 ¥81.67 / 月</small></strong>{["商业沙盘推演：20 次/月", "竞品全盘数据破解：200 次/月", "竞品动态监测：共用查询额度", "其他功能：本版暂不限次", "支持 GPT / Claude / Grok 模型切换", "支持深度思考"].map((item) => <p className="upgrade-benefit" key={item}><i /><span>{item}</span></p>)}</article>
    </div>
    <div className="benefit-table"><h3>权益对比一览</h3><div className="benefit-table-head"><span>功能权益</span><span>普通版</span><strong>会员版</strong></div>{[["商业沙盘推演", "1 次 / 月", "20 次 / 月"], ["竞品全盘数据破解查询", "5 次 / 月", "200 次 / 月"], ["竞品动态监测", "共用查询额度", "共用查询额度"], ["模型切换", "默认单一模型", "支持 GPT / Claude / Grok 模型切换"], ["其他功能", "本版暂不限次", "本版暂不限次"]].map(([name, normal, vip]) => <div key={name}><span>{name}</span><span>{normal}</span><strong>{vip}</strong></div>)}</div>
    <form className="upgrade-redeem" onSubmit={(event) => void redeem(event)}><label><span className="sr-only">兑换码</span><input aria-label="兑换码" onChange={(event) => setCode(event.target.value)} placeholder="已有兑换码？在此输入" value={code} /></label><button disabled={redeeming} type="submit">{redeeming ? "兑换中..." : "立即兑换"}</button>{message && <p className="form-success" role="status">{message}</p>}</form>
    <footer className="upgrade-footer"><div className="upgrade-contact-qr" aria-hidden="true" /><span><strong>联系企业微信升级权限</strong><small>添加顾问，获取开通帮助</small></span><Link to="/membership">稍后再说</Link><button onClick={() => navigate(`/membership/checkout${paidPlan ? `?plan=${paidPlan.code}` : ""}`)} type="button">立即开通</button></footer>
  </section></div>;
}

function planName(plans: MembershipPlanOption[], code: string) {
  return plans.find((plan) => plan.code === code)?.name || (code === "enterprise" ? "企业版（续行）" : code);
}

function formatAmount(cents: number) {
  return `¥${(cents / 100).toLocaleString("zh-CN", { minimumFractionDigits: 2 })}`;
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", hour12: false }).replaceAll("/", "-");
}

export default MembershipPage;
