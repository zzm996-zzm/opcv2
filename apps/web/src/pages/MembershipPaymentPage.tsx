import { useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

type MembershipPaymentPageProps = {
  mode: "checkout" | "quota" | "success";
};

const benefits = [
  ["商业沙盘推演", "20 次 / 月"],
  ["竞品全盘数据破解", "200 次 / 月"],
  ["竞品动态监测", "共用查询额度"],
  ["其他功能", "本版暂不限次"],
  ["支持 GPT / Claude / Grok 模型切换", "本版提供"],
  ["支持深度思考", "本版提供"]
] as const;

function MembershipPaymentPage({ mode }: MembershipPaymentPageProps) {
  if (mode === "quota") return <QuotaModalPage />;
  return <V4PageShell className="payment-shell" showCopilotMini={mode === "success"}><section className={`payment-page ${mode === "success" ? "payment-success-page" : ""}`}>{mode === "success" ? <SuccessContent /> : <CheckoutContent />}</section></V4PageShell>;
}

function CheckoutContent() {
  const [cycle, setCycle] = useState<"year" | "month">("year");
  const [paymentMethod, setPaymentMethod] = useState<"wechat" | "alipay">("wechat");
  const [couponCode, setCouponCode] = useState("");
  const [couponMessage, setCouponMessage] = useState("");
  const [qrFailed, setQrFailed] = useState(false);
  const [qrVersion, setQrVersion] = useState(0);
  const pricing = cycle === "year"
    ? { cycleLabel: "年付", period: "1 年", original: "¥1,176", discount: "-¥196", amount: "¥980.00" }
    : { cycleLabel: "月付", period: "1 个月", original: "¥98", discount: "¥0", amount: "¥98.00" };
  const paymentLabel = paymentMethod === "wechat" ? "微信支付" : "支付宝";

  function refreshQr() {
    setQrFailed(false);
    setQrVersion((version) => version + 1);
  }

  return <>
    <header className="payment-title"><Link to="/membership/upgrade" aria-label="返回">←</Link><div><h1>下单 / 支付页</h1><p>请确认订单信息并完成支付，开通会员后即可享受全部权益</p></div></header>
    <div className="checkout-layout">
      <main className="checkout-main-card">
        <section className="checkout-plan">
          <div><h2>已选择的套餐</h2><article><strong>♛ 会员版 <b>推荐</b></strong><p>适合企业团队及专业用户，享受更高配额与专属能力</p><span>高配额</span><span>多模型切换</span><span>深度思考</span><span>专属支持</span></article></div>
          <div><h2>选择付费周期</h2><div aria-label="付费周期" className="cycle-choice" role="radiogroup">
            <button aria-checked={cycle === "year"} className={cycle === "year" ? "active" : ""} onClick={() => setCycle("year")} role="radio" type="button"><b>年付 <em>更优惠</em></b><strong>¥980 <small>/ 年</small></strong><span>折合 ¥81.67 / 月</span><i>立省 ¥196</i>{cycle === "year" && <span className="choice-check" aria-hidden="true">✓</span>}</button>
            <button aria-checked={cycle === "month"} className={cycle === "month" ? "active" : ""} onClick={() => setCycle("month")} role="radio" type="button"><b>月付</b><strong>¥98 <small>/ 月</small></strong>{cycle === "month" && <span className="choice-check" aria-hidden="true">✓</span>}</button>
          </div></div>
        </section>
        <section className="checkout-benefits"><h2>权益摘要</h2><div>{benefits.map(([name, value]) => <span key={name}><i>◇</i><strong>{name}</strong><b>{value}</b></span>)}</div></section>
        <section className="coupon-row"><h2>优惠券 / 兑换码</h2><div><input aria-label="优惠券或兑换码" onChange={(event) => { setCouponCode(event.target.value); setCouponMessage(""); }} placeholder="请输入优惠券码" value={couponCode} /><button disabled={!couponCode.trim()} onClick={() => setCouponMessage("优惠码已提交，请以结算结果为准")} type="button">应用</button><span>{couponMessage || "暂无可用优惠券 ⓘ"}<Link to="/membership">查看可用优惠券 ›</Link></span></div></section>
        <section className="payment-methods"><h2>支付方式</h2><div aria-label="支付方式" role="radiogroup">
          <button aria-checked={paymentMethod === "wechat"} className={paymentMethod === "wechat" ? "active" : ""} onClick={() => setPaymentMethod("wechat")} role="radio" type="button"><i className="wechat-pay" />微信支付<small>推荐使用微信扫一扫支付</small>{paymentMethod === "wechat" && <b aria-hidden="true">✓</b>}</button>
          <button aria-checked={paymentMethod === "alipay"} className={paymentMethod === "alipay" ? "active" : ""} onClick={() => setPaymentMethod("alipay")} role="radio" type="button"><i className="alipay" />支付宝<small>推荐使用支付宝扫一扫支付</small>{paymentMethod === "alipay" && <b aria-hidden="true">✓</b>}</button>
        </div></section>
        <footer className="payment-trust"><span>♢ <strong>安全可靠</strong><small>支付过程安全加密</small></span><span>♙ <strong>官方服务</strong><small>智活AI官方直销</small></span><span>▤ <strong>发票保障</strong><small>支持企业增值税发票</small></span></footer>
      </main>
      <aside className="checkout-side">
        <section className="order-summary"><h2>订单摘要</h2><p><span>商品名称</span><b>会员版（{pricing.cycleLabel}）</b></p><p><span>付费周期</span><b>{pricing.period}</b></p><p><span>原价</span><b>{pricing.original}</b></p><p><span>优惠金额</span><b className="discount">{pricing.discount}</b></p><hr /><strong>应付金额 <b>{pricing.amount}</b></strong><small>◉ 支付成功后自动开通会员、支持开票</small></section>
        <section className="qr-payment"><h2>{paymentLabel}</h2><p>请使用{paymentMethod === "wechat" ? "微信" : "支付宝"}扫一扫完成支付</p><div className={`payment-qr ${qrFailed ? "failed" : ""}`} aria-label={`${paymentLabel}二维码`}>{qrFailed ? <span role="status">二维码加载失败，请重试或切换支付方式</span> : <img alt="" onError={() => setQrFailed(true)} src={`/public-components/payment-qr.jpg?refresh=${qrVersion}`} />}</div><button onClick={refreshQr} type="button">⟳ {qrFailed ? "重试二维码" : "刷新二维码"}</button><hr /><small>◷ 订单有效期 <b>14:52</b></small><Link className="checkout-pay-action" to="/membership/success"><span>应付金额 <strong>{pricing.amount}</strong></span><b>立即支付 <em>{pricing.amount}</em></b></Link><p>支付即表示您已阅读并同意<Link to="/terms">《会员服务协议》</Link></p></section>
      </aside>
    </div>
  </>;
}

function QuotaModalPage() {
  return <><V4PageShell className="payment-shell"><section className="profile-page quota-background" aria-hidden="true"><div className="page-title-row"><div><h1>个人中心</h1><p>管理您的账户信息、使用额度与个性化设置</p></div></div></section></V4PageShell><div className="ui-modal-scrim quota-exhausted-scrim" role="dialog" aria-label="本月额度已用尽"><section className="quota-exhausted-modal"><Link className="modal-close" to="/profile">×</Link><div className="quota-exhausted-hero"><div className="quota-lock-art" aria-hidden="true"><span>!</span><i>∞</i></div><div><h1>本月额度已用尽<br />当前功能已锁定</h1><p>您已使用完本月「商业沙盘推演」的可用次数<br />升级会员后即可继续使用，享受更多权益</p><strong>♧ 本月已使用： 10 / 10 次 <span>重置时间 2025-06-01</span></strong></div></div><section><h2>升级会员，解锁更多权益</h2><div>{[["▤", "更高使用额度", "享受更高使用次数"], ["✦", "全量功能解锁", "畅享全部高效功能"], ["⬡", "专属模型切换", "支持更多模型选择"], ["⌁", "优先响应支持", "专属客服快速响应"]].map(([icon, title, desc]) => <article key={title}><i>{icon}</i><strong>{title}</strong><small>{desc}</small></article>)}</div></section><footer><span>♢ <strong>安全可靠的企业级数据保护</strong><small>您的数据安全与隐私受到严格保护</small></span><Link to="/profile">稍后再说</Link><Link className="primary" to="/membership/upgrade">升级套餐</Link></footer></section></div></>;
}

function SuccessContent() {
  const [copied, setCopied] = useState(false);
  const orderNumber = "20250601101545987612";

  async function copyOrderNumber() {
    setCopied(true);
    try {
      await navigator.clipboard?.writeText(orderNumber);
    } catch {
      // Clipboard access can be unavailable in embedded browsers; the inline state still confirms the action.
    }
  }

  return <div className="success-layout"><main className="success-main"><header><span>✓</span><h1>支付成功，会员已激活！</h1><p>感谢您的信任与支持，智活AI助您数据驱动，增长确定。</p></header><section className="activation-facts">{[["♛", "会员等级", "会员版 推荐"], ["▣", "生效时间", "2025-06-01 10:15"], ["▣", "有效期至", "2026-05-31"], ["⟳", "下次重置日", "2025-07-01"]].map(([icon, label, value]) => <article key={label}><i>{icon}</i><span>{label}</span><strong>{value}</strong></article>)}</section><section className="renewed-benefits"><h2>额度已刷新，可立即使用</h2><div className="success-table"><div><span>功能权益</span><span>本次刷新额度</span><span>有效期内可用</span></div>{benefits.slice(0, 5).map(([name, value]) => <div key={name}><span>◇ {name}</span><span>+{value}</span><span>{value}</span></div>)}</div><p>✓ 您的会员权益已生效，当前所有功能均可使用</p></section><footer><Link to="/profile">返回个人中心</Link><Link className="primary" to="/">立即使用</Link><p>♢ 支付遇到问题？ <Link to="/help">查看支付帮助</Link> 或 <Link to="/help">联系在线客服 ›</Link></p></footer></main><aside className="success-side"><section><h2>本次开通信息</h2>{[["开通套餐", "会员版（年付）"], ["支付金额", "¥980"], ["支付方式", "微信支付"], ["订单编号", orderNumber], ["支付时间", "2025-06-01 10:15:45"]].map(([label, value]) => <p key={label}><span>{label}</span><b className={label === "订单编号" ? "order-number" : ""}><span>{value}</span>{label === "订单编号" && <button aria-label="复制订单编号" onClick={() => void copyOrderNumber()} type="button">{copied ? "已复制" : "复制"}</button>}</b></p>)}</section><article><i>♛</i><div><strong>会员专属特权</strong><p>更多高阶能力，助力企业精准决策，持续增长</p><Link to="/membership">查看全部权益 ›</Link></div></article></aside></div>;
}

export default MembershipPaymentPage;
