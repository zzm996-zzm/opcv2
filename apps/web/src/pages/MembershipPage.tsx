import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  Box,
  BrainCircuit,
  ChartNoAxesCombined,
  Check,
  Circle,
  Crown,
  Database,
  Folder,
  Grid3X3,
  Info,
  Layers3,
  Pencil,
  Radar,
  Sparkles,
  X,
  type LucideIcon
} from "lucide-react";

import AccountSectionNav from "../components/AccountSectionNav";
import PublicCopilotPanel from "../components/PublicCopilotPanel";
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
            {plans.length > 0 && (
              <section className="membership-plan-options" aria-labelledby="membership-plan-options-title">
                <div className="module-section-head">
                  <div>
                    <h2 id="membership-plan-options-title">套餐选择</h2>
                    <p>选择套餐后创建订单，继续完成支付</p>
                  </div>
                </div>
                <div className="membership-plan-option-grid">
                  {plans.map((plan) => {
                    const isCurrent = snapshot?.plan.code === plan.code;
                    const isCheckingOut = checkoutPlan === plan.code;
                    return (
                      <article className="membership-plan-option" key={plan.code}>
                        <div>
                          <h3>{plan.name}</h3>
                          <p>{plan.price_cents === 0 ? "免费" : formatAmount(plan.price_cents)} / {plan.billing_cycle === "year" ? "年" : "月"}</p>
                        </div>
                        <button disabled={isCurrent || Boolean(checkoutPlan)} onClick={() => void createCheckout(plan)} type="button">
                          {isCurrent ? "当前使用" : isCheckingOut ? "创建中..." : `开通${plan.name}`}
                        </button>
                      </article>
                    );
                  })}
                </div>
              </section>
            )}
            {checkoutMessage && <p className="form-success" role="status">{checkoutMessage}</p>}
            <span className="sr-only">当前积分 {snapshot?.credit_balance ?? 0}</span>
            {visibleUsage.map((item) => <span className="sr-only" key={`test-${item.key}`}>{item.label} {item.used}/{item.limit} {item.unit}</span>)}
          </div>
          <PublicCopilotPanel />
        </div>
      </section>
    </V4PageShell>
    {showUpgrade && <MembershipUpgradeModal plans={plans} />}
    </>
  );
}

function ProfileTabs() {
  return <AccountSectionNav activeHref="/membership" />;
}

function QuotaCard({ item }: { item: MembershipUsageItem }) {
  const configured = item.limit > 0;
  const percent = configured ? Math.min(100, Math.round(item.used / item.limit * 100)) : 0;
  const Icon = quotaIconFor(item);
  return <article className={`quota-card quota-card-${item.key}`}><span className="quota-label"><Icon aria-hidden="true" size={18} strokeWidth={2} />{item.label}</span><strong>{item.used.toLocaleString()}<small> / {item.limit.toLocaleString()}</small></strong><div className="quota-bar" aria-hidden="true"><i style={{ width: `${percent}%` }} /></div><small>{configured ? `剩余 ${100 - percent}%` : "未配置"}</small><small>重置日：2025-06-01</small></article>;
}

const quotaIcons: Record<string, LucideIcon> = {
  ai: BrainCircuit,
  data: Database,
  tools: Pencil,
  sandbox: ChartNoAxesCombined,
  competitor: Layers3
};

function quotaIconFor(item: MembershipUsageItem): LucideIcon {
  const key = item.key.toLowerCase();
  const label = item.label.toLowerCase();
  if (quotaIcons[item.key]) return quotaIcons[item.key];
  if (key.includes("competitor") || label.includes("竞品")) return Layers3;
  if (key.includes("sandbox") || label.includes("沙盘")) return ChartNoAxesCombined;
  if (key.includes("tool") || label.includes("工具")) return Pencil;
  if (key.includes("data") || key.includes("lead") || label.includes("数据") || label.includes("线索")) return Database;
  if (key.includes("copilot") || key.includes("chat") || label.includes("对话")) return Sparkles;
  return BrainCircuit;
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

  const normalBenefits = ["商业沙盘推演：1 次/月", "竞品全盘数据破解：5 次/月", "竞品动态监测：共用查询额度", "其他功能：本版暂不限次", "默认单一模型"];
  const memberBenefits = ["商业沙盘推演：20 次/月", "竞品全盘数据破解：200 次/月", "竞品动态监测：共用查询额度", "其他功能：本版暂不限次", "支持 GPT / Claude / Grok 模型切换", "支持深度思考"];
  const comparisonRows: Array<[LucideIcon, string, string, string]> = [
    [ChartNoAxesCombined, "商业沙盘推演", "1 次 / 月", "20 次 / 月"],
    [Database, "竞品全盘数据破解查询", "5 次 / 月", "200 次 / 月"],
    [Radar, "竞品动态监测", "共用查询额度", "共用查询额度"],
    [Box, "模型切换", "默认单一模型", "支持 GPT / Claude / Grok 模型切换"],
    [Grid3X3, "其他功能", "本版暂不限次", "本版暂不限次"]
  ];

  return <div className="ui-modal-scrim public-upgrade-scrim" role="dialog" aria-label="升级套餐"><section className="upgrade-modal public-upgrade-modal">
    <div className="upgrade-head"><div><h2>升级套餐</h2><p>对比普通版与会员版权益，选择更适合你的方案</p></div><div className="billing-toggle"><button type="button">月付</button><button className="active" type="button">年付 <span>更优惠</span></button></div><Link className="modal-close" to="/membership" aria-label="关闭升级套餐"><X aria-hidden="true" /></Link></div>
    <div className="upgrade-plan-grid">
      <article className="upgrade-plan-card">
        <div className="upgrade-plan-title"><Folder aria-hidden="true" /><h3>普通版</h3><small>当前方案</small></div>
        <div className="upgrade-plan-price"><strong>¥0</strong><small>/ 免费</small></div>
        <div className="upgrade-benefit-list">{normalBenefits.map((item) => <UpgradeBenefitRow item={item} key={item} />)}</div>
      </article>
      <article className="upgrade-plan-card featured">
        <span className="upgrade-ribbon" aria-hidden="true">推荐</span>
        <div className="upgrade-plan-title"><Crown aria-hidden="true" /><h3>会员版</h3><small>推荐</small></div>
        <div className="upgrade-plan-price"><strong>¥980</strong><small>/ 年 折合 ¥81.67 / 月</small></div>
        <div className="upgrade-benefit-list">{memberBenefits.map((item) => <UpgradeBenefitRow featured item={item} key={item} />)}</div>
      </article>
    </div>
    <div className="benefit-table"><h3>权益对比一览</h3><div className="benefit-table-head"><span aria-hidden="true" /><span>普通版</span><strong>会员版</strong></div>{comparisonRows.map(([Icon, name, normal, vip]) => <div className="benefit-table-row" key={name}><span className="benefit-name"><Icon aria-hidden="true" />{name}</span><span>{normal}</span><strong>{vip}</strong></div>)}<p className="benefit-note"><Info aria-hidden="true" />所有次数 / 额度按周期真实重置</p></div>
    <form className="upgrade-redeem" onSubmit={(event) => void redeem(event)}><label><span className="sr-only">兑换码</span><input aria-label="兑换码" onChange={(event) => setCode(event.target.value)} placeholder="已有兑换码？在此输入" value={code} /></label><button disabled={redeeming} type="submit">{redeeming ? "兑换中..." : "立即兑换"}</button>{message && <p className="form-success" role="status">{message}</p>}</form>
    <footer className="upgrade-footer"><div className="upgrade-contact-qr" aria-hidden="true" /><span><strong>联系企业微信升级权限</strong><small>添加顾问，获取开通帮助</small></span><Link to="/membership">稍后再说</Link><button onClick={() => navigate(`/membership/checkout${paidPlan ? `?plan=${paidPlan.code}` : ""}`)} type="button">立即再说</button></footer>
  </section></div>;
}

function UpgradeBenefitRow({ featured = false, item }: { featured?: boolean; item: string }) {
  const [label, value] = item.split("：");
  return <p className={`upgrade-benefit${value ? "" : " full"}`}>
    {featured ? <span className="upgrade-check"><Check aria-hidden="true" /></span> : <Circle aria-hidden="true" />}
    <span>{label}{value ? "：" : ""}</span>
    {value && <strong>{value}</strong>}
  </p>;
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
