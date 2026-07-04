import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import {
  membershipApi,
  type MembershipOrder,
  type MembershipPlanOption,
  type MembershipSnapshot,
  type MembershipUsageItem
} from "../lib/membershipApi";
import { CdkTopNav } from "./AnalysisPage";

const benefits = ["线索数据实时更新", "去水印导出结果", "优先处理与客服支持", "社群活动内容与活动", "积分可加购，未用完不累计"] as const;

type MembershipPageProps = {
  showUpgrade?: boolean;
};

const planTone = ["free", "basic", "standard", "enterprise"] as const;

function formatPrice(cents: number) {
  if (cents === 0) return "¥0";
  const amount = cents / 100;
  return Number.isInteger(amount) ? `¥${amount}` : `¥${amount.toFixed(2)}`;
}

function formatAmount(cents: number) {
  return `¥${(cents / 100).toFixed(2)}`;
}

function formatDateTime(value?: string) {
  if (!value) return "未支付";
  return new Date(value).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  });
}

function orderStatusLabel(status: string) {
  if (status === "paid") return "已支付";
  if (status === "pending") return "待支付";
  return status;
}

function planName(plans: MembershipPlanOption[], code: string) {
  return plans.find((plan) => plan.code === code)?.name ?? code;
}

function MembershipPage({ showUpgrade = false }: MembershipPageProps) {
  return showUpgrade ? <RedeemPage /> : <PurchaseMembershipPage />;
}

function PurchaseMembershipPage() {
  const [snapshot, setSnapshot] = useState<MembershipSnapshot | null>(null);
  const [plans, setPlans] = useState<MembershipPlanOption[]>([]);
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const [orders, setOrders] = useState<MembershipOrder[]>([]);
  const [loading, setLoading] = useState(true);
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
        setLoadError("");
      })
      .catch((error) => {
        if (!active) return;
        setLoadError(apiErrorMessage(error, "暂时无法读取会员信息"));
        setSnapshot(null);
        setPlans([]);
        setUsage([]);
        setOrders([]);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function createCheckout(plan: MembershipPlanOption) {
    setCheckoutPlan(plan.code);
    setCheckoutMessage("");
    try {
      const result = await membershipApi.checkout({
        plan_code: plan.code,
        billing_cycle: plan.billing_cycle
      });
      setOrders((current) => [result.order, ...current]);
      setCheckoutMessage(`已创建订单 ${result.order.order_no}，${result.payment.message || "客服会协助完成支付"}`);
    } catch (error) {
      setCheckoutMessage(apiErrorMessage(error, "暂时无法创建订单"));
    } finally {
      setCheckoutPlan("");
    }
  }

  const currentPlanCode = snapshot?.plan.code;
  const currentPlanName = snapshot?.plan.name ?? "未开通";
  const visibleBenefits = plans.find((plan) => plan.code === currentPlanCode)?.features ?? benefits;

  return (
    <main className="cdk-analysis-page cdk-membership-page">
      <CdkTopNav active="会员计划" />
      <section className="cdk-membership-hero" aria-label="会员与账单">
        <h1>选择适合你的会员计划</h1>
        <p>用获客积分解锁更多AI能力，高效获取可核实的精准客户</p>
        <div className="cdk-membership-cycle" aria-label="计费周期">
          <button className="active" type="button">连续包年 <span>省35%</span></button>
          <button type="button">连续包月</button>
          <button type="button">单月购买</button>
        </div>
        <div className="cdk-membership-switch">
          <Link className="active" to="/membership">购买会员</Link>
          <Link to="/membership/upgrade">会员兑换</Link>
        </div>
      </section>

      <section className="cdk-membership-main">
        <div className="cdk-plan-grid" aria-label="会员计划">
          {loading && <p>正在读取会员计划...</p>}
          {loadError && <p className="form-error" role="alert">{loadError}</p>}
          {!loading && !loadError && plans.length === 0 && <p>暂无可购买套餐</p>}
          {plans.map((plan, index) => (
            <article className={`cdk-plan-card ${planTone[index % planTone.length]} ${plan.recommended ? "recommended" : ""}`} key={plan.code}>
              {plan.recommended && <i>推荐</i>}
              <header>
                <span aria-hidden="true" />
                <div>
                  <h2>{plan.name}</h2>
                  <p>{plan.features[0] ?? "适合持续使用 AI 获客能力"}</p>
                </div>
              </header>
              <strong>{formatPrice(plan.price_cents)}<small>/{plan.billing_cycle === "year" ? "年" : "月"}</small></strong>
              {plan.billing_cycle === "year" && <p className="cdk-plan-note">按年计费</p>}
              <button
                disabled={checkoutPlan === plan.code || currentPlanCode === plan.code}
                onClick={() => void createCheckout(plan)}
                type="button"
              >
                {currentPlanCode === plan.code ? "当前使用" : checkoutPlan === plan.code ? "创建中..." : `开通${plan.name}`}
              </button>
              <div className="cdk-plan-quota">
                {plan.quotas.map((item) => <span key={item.key}>{item.label} {item.limit} {item.unit}</span>)}
              </div>
              <section>
                <h3>可用板块</h3>
                {plan.features.map((item) => <small key={item}>✓ {item}</small>)}
              </section>
            </article>
          ))}
        </div>

        <aside className="cdk-membership-side">
          <article>
            <h2>统一货币：获客积分</h2>
            <p>积分可用于解锁各项AI功能，按使用扣除，当月有效</p>
            {snapshot && <strong>当前积分 {snapshot.credit_balance}</strong>}
            <div>
              {usage.length > 0 ? usage.map((item) => (
                <span key={item.key}>{item.label} {item.used}/{item.limit} {item.unit}</span>
              )) : <span>暂无额度使用记录</span>}
            </div>
          </article>
          <article>
            <h2>包含权益</h2>
            {visibleBenefits.map((item) => <p key={item}>✓ {item}</p>)}
          </article>
        </aside>
      </section>
      {checkoutMessage && <p className="form-success" role="status">{checkoutMessage}</p>}

      <section className="cdk-membership-safe">
        <strong>没有风险的订阅体验</strong>
        <p>随时取消订阅，订阅期内权益不受影响。积分按月发放，当月有效，不累计到次月。</p>
        <div><span>支持支付方式</span><b>微信支付</b><b>支付宝</b><b>银联支付</b><b>VISA</b></div>
      </section>

      <section className="order-card cdk-membership-orders">
        <h2>订单记录</h2>
        <div className="order-table">
          <div className="order-head">
            {["订单号", "套餐", "支付时间", "支付状态", "金额", "支付方式", "操作"].map((item) => <span key={item}>{item}</span>)}
          </div>
          {orders.length === 0 && <p>暂无订单记录</p>}
          {orders.map((order) => (
            <div className="order-row" key={order.id}>
              <span>{order.order_no}</span>
              <span>{planName(plans, order.plan_code)}</span>
              <span>{formatDateTime(order.paid_at ?? order.created_at)}</span>
              <span className={order.status === "paid" ? "paid" : ""}>{orderStatusLabel(order.status)}</span>
              <span>{formatAmount(order.amount_cents)}</span>
              <span>{order.status === "paid" ? "手动支付" : "待确认"}</span>
              <a href="/membership">查看详情</a>
            </div>
          ))}
        </div>
      </section>

      <h2 className="sr-only">会员与账单</h2>
      <h2 className="sr-only">{currentPlanName} 会员生效</h2>
    </main>
  );
}

function RedeemPage() {
  const [snapshot, setSnapshot] = useState<MembershipSnapshot | null>(null);
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(true);
  const [redeeming, setRedeeming] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    membershipApi
      .current()
      .then((payload) => {
        if (!active) return;
        setSnapshot(payload);
      })
      .catch(() => {
        if (!active) return;
        setSnapshot(null);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function redeem(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = code.trim();
    if (!trimmed) {
      setMessage("请输入兑换码");
      return;
    }
    setRedeeming(true);
    setMessage("");
    try {
      const result = await membershipApi.redeem(trimmed);
      setSnapshot(result.snapshot);
      setMessage(result.already_redeemed ? "兑换码已使用，账户权益保持不变" : `兑换成功，当前会员：${result.snapshot.plan.name}`);
    } catch (error) {
      setMessage(apiErrorMessage(error, "兑换失败，请稍后重试"));
    } finally {
      setRedeeming(false);
    }
  }

  return (
    <main className="cdk-analysis-page cdk-membership-page cdk-redeem-page">
      <CdkTopNav active="会员计划" />
      <section className="cdk-membership-hero">
        <h1>会员兑换</h1>
        <p>使用兑换码激活会员权益或兑换积分，享受更多 AI 获客能力</p>
        <div className="cdk-membership-switch redeem">
          <Link to="/membership">购买会员</Link>
          <Link className="active" to="/membership/upgrade">会员兑换</Link>
        </div>
      </section>

      <section className="cdk-redeem-layout">
        <form className="cdk-redeem-card" onSubmit={(event) => void redeem(event)}>
          <div className="cdk-redeem-tabs">
            <button className="active" type="button">兑换会员/权益</button>
            <button type="button">兑换积分</button>
          </div>
          <label>
            兑换码
            <div>
              <input
                aria-label="兑换码"
                onChange={(event) => setCode(event.target.value)}
                placeholder="请输入16-32位兑换码（区分大小写）"
                value={code}
              />
              <button type="button">粘贴兑换码</button>
            </div>
          </label>
          <button disabled={redeeming} type="submit">{redeeming ? "兑换中..." : "立即兑换"}</button>
          {message && <p className="form-success" role="status">{message}</p>}
          <section>
            <h2>兑换码说明</h2>
            <ul>
              <li>兑换码可通过官方活动、合作伙伴、购买赠送等方式获得</li>
              <li>兑换成功后，权益或积分将自动发放到您的账户</li>
              <li>兑换码一旦使用，不可退换或转让</li>
              <li>如遇问题，请联系客服</li>
            </ul>
          </section>
        </form>

        <aside className="cdk-redeem-side">
          <section>
            <h2>当前账户信息</h2>
            <div><span>当前会员</span><strong>{loading ? "读取中..." : snapshot?.plan.name ?? "未开通"}</strong><Link to="/membership">查看详情 ›</Link></div>
            <div><span>当前积分</span><strong>{snapshot?.credit_balance ?? 0} 积分</strong><Link to="/membership">去购买 ›</Link></div>
          </section>
          <section>
            <h2>温馨提示</h2>
            {["兑换码区分大小写，请仔细核对", "部分兑换码有使用期限，请在有效期内使用", "积分兑换码可能存在使用限制，请查看活动规则", "企业专属兑换码仅限企业成员使用"].map((tip) => (
              <p key={tip}>✓ {tip}</p>
            ))}
          </section>
        </aside>
      </section>

      <section className="cdk-redeem-more">
        <h2>更多获取兑换码的方式</h2>
        {[
          ["官方活动", "参与平台活动，赢取兑换码"],
          ["合作伙伴", "通过合作伙伴获取专属兑换码"],
          ["购买赠送", "购买会员或积分时获赠兑换码"]
        ].map(([title, desc]) => (
          <article key={title}><i aria-hidden="true" /><strong>{title}</strong><p>{desc}</p><span>›</span></article>
        ))}
      </section>

      <MembershipUpgradeModal />
    </main>
  );
}

function MembershipUpgradeModal() {
  return (
    <div className="ui-modal-scrim cdk-upgrade-scrim" role="dialog" aria-label="升级套餐">
      <section className="upgrade-modal">
        <div className="upgrade-head">
          <div>
            <h2>升级套餐</h2>
            <p>对比普通版与会员版权益，选择更适合你的方案</p>
            <span className="sr-only">会员版</span>
          </div>
          <Link className="modal-close" to="/membership" aria-label="关闭升级套餐">×</Link>
        </div>
        <div className="benefit-table">
          <h3>权益对比一览</h3>
          {[
            ["商业沙盘推演", "1 次 / 月", "20 次 / 月"],
            ["竞品全盘数据破解查询", "5 次 / 月", "200 次 / 月"],
            ["模型切换", "默认单一模型", "支持 GPT / Claude / Grok 模型切换"]
          ].map(([name, normal, vip]) => (
            <div key={name}><span>{name}</span><span>{normal}</span><strong>{vip}</strong></div>
          ))}
        </div>
        <footer className="upgrade-footer">
          <Link to="/membership">稍后再说</Link>
          <button type="button">立即开通</button>
        </footer>
      </section>
    </div>
  );
}

export default MembershipPage;
