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
  return <><header className="payment-title"><Link to="/membership/upgrade" aria-label="返回">←</Link><div><h1>下单 / 支付页</h1><p>请确认订单信息并完成支付，开通会员后即可享受全部权益</p></div></header><div className="checkout-layout"><main className="checkout-main-card">
    <section className="checkout-plan"><div><h2>已选择的套餐</h2><article><strong>♛ 会员版 <b>推荐</b></strong><p>适合企业团队及专业用户，享受更高配额与专属能力</p><span>高配额</span><span>多模型切换</span><span>深度思考</span><span>专属支持</span></article></div><div><h2>选择付费周期</h2><div className="cycle-choice"><button className="active" type="button"><b>年付 <em>更优惠</em></b><strong>¥980 <small>/ 年</small></strong><span>折合 ¥81.67 / 月</span><i>立省 ¥196</i></button><button type="button"><b>月付</b><strong>¥98 <small>/ 月</small></strong></button></div></div></section>
    <section className="checkout-benefits"><h2>权益摘要</h2><div>{benefits.map(([name, value]) => <span key={name}><i>◇</i><strong>{name}</strong><b>{value}</b></span>)}</div></section>
    <section className="coupon-row"><h2>优惠券 / 兑换码</h2><div><input aria-label="优惠券或兑换码" placeholder="请输入优惠券码" /><button type="button">应用</button><span>暂无可用优惠券 ⓘ<Link to="/membership">查看可用优惠券 ›</Link></span></div></section>
    <section className="payment-methods"><h2>支付方式</h2><div><button className="active" type="button"><i className="wechat-pay" />微信支付<small>推荐使用微信扫一扫支付</small><b>✓</b></button><button type="button"><i className="alipay" />支付宝<small>推荐使用支付宝扫一扫支付</small></button></div></section>
    <footer className="payment-trust"><span>♢ <strong>安全可靠</strong><small>支付过程安全加密</small></span><span>♙ <strong>官方服务</strong><small>智活AI官方直销</small></span><span>▤ <strong>发票保障</strong><small>支持企业增值税发票</small></span></footer>
  </main><aside className="checkout-side"><section className="order-summary"><h2>订单摘要</h2><p><span>商品名称</span><b>会员版（年付）</b></p><p><span>付费周期</span><b>1 年</b></p><p><span>原价</span><b>¥1,176</b></p><p><span>优惠金额</span><b className="discount">-¥196</b></p><hr /><strong>应付金额 <b>¥980.00</b></strong><small>◉ 支付成功后自动开通会员、支持开票</small></section><section className="qr-payment"><h2>微信支付</h2><p>请使用微信扫一扫完成支付</p><div className="payment-qr" aria-label="微信支付二维码" /><button type="button">⟳ 刷新二维码</button><hr /><small>◷ 订单有效期 <b>14:52</b></small><Link to="/membership/success">立即支付 ¥980.00</Link><p>支付即表示您已阅读并同意《会员服务协议》</p></section></aside></div></>;
}

function QuotaModalPage() {
  return <><V4PageShell className="payment-shell"><section className="profile-page quota-background" aria-hidden="true"><div className="page-title-row"><div><h1>个人中心</h1><p>管理您的账户信息、使用额度与个性化设置</p></div></div></section></V4PageShell><div className="ui-modal-scrim quota-exhausted-scrim" role="dialog" aria-label="本月额度已用尽"><section className="quota-exhausted-modal"><Link className="modal-close" to="/profile">×</Link><div className="quota-exhausted-hero"><div className="quota-lock-art" aria-hidden="true"><span>!</span><i>∞</i></div><div><h1>本月额度已用尽<br />当前功能已锁定</h1><p>您已使用完本月「商业沙盘推演」的可用次数<br />升级会员后即可继续使用，享受更多权益</p><strong>♧ 本月已使用： 10 / 10 次 <span>重置时间 2025-06-01</span></strong></div></div><section><h2>升级会员，解锁更多权益</h2><div>{[["▤", "更高使用额度", "享受更高使用次数"], ["✦", "全量功能解锁", "畅享全部高效功能"], ["⬡", "专属模型切换", "支持更多模型选择"], ["⌁", "优先响应支持", "专属客服快速响应"]].map(([icon, title, desc]) => <article key={title}><i>{icon}</i><strong>{title}</strong><small>{desc}</small></article>)}</div></section><footer><span>♢ <strong>安全可靠的企业级数据保护</strong><small>您的数据安全与隐私受到严格保护</small></span><Link to="/profile">稍后再说</Link><Link className="primary" to="/membership/upgrade">升级套餐</Link></footer></section></div></>;
}

function SuccessContent() {
  return <div className="success-layout"><main className="success-main"><header><span>✓</span><h1>支付成功，会员已激活！</h1><p>感谢您的信任与支持，智活AI助您数据驱动，增长确定。</p></header><section className="activation-facts">{[["♛", "会员等级", "会员版 推荐"], ["▣", "生效时间", "2025-06-01 10:15"], ["▣", "有效期至", "2026-05-31"], ["⟳", "下次重置日", "2025-07-01"]].map(([icon, label, value]) => <article key={label}><i>{icon}</i><span>{label}</span><strong>{value}</strong></article>)}</section><section className="renewed-benefits"><h2>额度已刷新，可立即使用</h2><div className="success-table"><div><span>功能权益</span><span>本次刷新额度</span><span>有效期内可用</span></div>{benefits.slice(0, 5).map(([name, value]) => <div key={name}><span>◇ {name}</span><span>+{value}</span><span>{value}</span></div>)}</div><p>✓ 您的会员权益已生效，当前所有功能均可使用</p></section><footer><Link to="/profile">返回个人中心</Link><Link className="primary" to="/">立即使用</Link><p>♢ 支付遇到问题？ <Link to="/help">查看支付帮助</Link> 或 <Link to="/help">联系在线客服 ›</Link></p></footer></main><aside className="success-side"><section><h2>本次开通信息</h2>{[["开通套餐", "会员版（年付）"], ["支付金额", "¥980"], ["支付方式", "微信支付"], ["订单编号", "20250601101545987612"], ["支付时间", "2025-06-01 10:15:45"]].map(([label, value]) => <p key={label}><span>{label}</span><b>{value}</b></p>)}</section><article><i>♛</i><div><strong>会员专属特权</strong><p>更多高阶能力，助力企业精准决策，持续增长</p><Link to="/membership">查看全部权益 ›</Link></div></article></aside></div>;
}

export default MembershipPaymentPage;
