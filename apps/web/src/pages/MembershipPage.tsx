import { Link } from "react-router-dom";

import { CdkTopNav } from "./AnalysisPage";

const plans = [
  {
    name: "免费版",
    desc: "体验基础功能",
    price: "¥0",
    note: "",
    action: "当前使用",
    tone: "free",
    recommended: false,
    quota: ["每月获客积分 50 积分", "可核实线索 10 条/月", "对标拆解 5 次/月", "AI 生成内容 10 次/月"],
    modules: ["模块一：免费分析（限次）", "模块二：内容严选客", "社群：免费层"]
  },
  {
    name: "基础会员",
    desc: "适合B2B或本地实体",
    price: "¥69",
    note: "¥828/年　省35%",
    action: "立即订阅",
    tone: "basic",
    recommended: false,
    quota: ["每月获客积分 800 积分", "可核实线索 80 条/月", "对标拆解 30 次/月", "AI 生成内容 200 次/月"],
    modules: ["模块一：免费分析（不限次）", "模块二：实战获客（单一身份）", "模块三：内容整合部内容", "社群：进阶层"]
  },
  {
    name: "标准会员",
    desc: "适合高客单B2C",
    price: "¥199",
    note: "¥2,388/年　省35%",
    action: "立即订阅",
    tone: "standard",
    recommended: true,
    quota: ["每月获客积分 2,400 积分", "可核实线索 250 条/月", "对标拆解 100 次/月", "AI 生成内容 600 次/月"],
    modules: ["模块一：免费分析（不限次）", "模块二：实战获客（高客单身份）", "模块三：内容整合部内容", "社群：进阶层", "导出结果去水印"]
  },
  {
    name: "高级/企业版",
    desc: "适合跨境及团队使用",
    price: "¥499",
    note: "¥5,988/年　省35%",
    action: "立即订阅",
    tone: "enterprise",
    recommended: false,
    quota: ["每月获客积分 6,000 积分", "可核实线索 800 条/月", "对标拆解 300 次/月", "AI 生成内容 1,500 次/月"],
    modules: ["模块一：免费分析（不限次）", "模块二：实战获客（团队身份）", "模块三：内容整合部内容", "社群：企业定制层", "多成员协作（最多 10 人）", "数据导出 / API 接入"]
  }
] as const;

const orders = [
  ["ZS-20250531-0012", "企业版（续行）", "2025-05-31 10:25", "已支付", "¥9,999.00", "已开票"],
  ["ZS-20240428-0018", "企业版（新付）", "2024-04-28 09:18", "已支付", "¥9,999.00", "已开票"]
] as const;

const benefits = ["线索数据实时更新", "去水印导出结果", "优先处理与客服支持", "社群活动内容与活动", "积分可加购，未用完不累计"] as const;

type MembershipPageProps = {
  showUpgrade?: boolean;
};

function MembershipPage({ showUpgrade = false }: MembershipPageProps) {
  if (showUpgrade) {
    return <RedeemPage />;
  }

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
          {plans.map((plan) => (
            <article className={`cdk-plan-card ${plan.tone} ${plan.recommended ? "recommended" : ""}`} key={plan.name}>
              {plan.recommended && <i>推荐</i>}
              <header>
                <span aria-hidden="true" />
                <div>
                  <h2>{plan.name}</h2>
                  <p>{plan.desc}</p>
                </div>
              </header>
              <strong>{plan.price}<small>/月</small></strong>
              {plan.note && <p className="cdk-plan-note">{plan.note}</p>}
              <button type="button">{plan.action}</button>
              <div className="cdk-plan-quota">
                {plan.quota.map((item) => <span key={item}>{item}</span>)}
              </div>
              <section>
                <h3>可用板块</h3>
                {plan.modules.map((item) => <small key={item}>✓ {item}</small>)}
              </section>
            </article>
          ))}
        </div>

        <aside className="cdk-membership-side">
          <article>
            <h2>统一货币：获客积分</h2>
            <p>积分可用于解锁各项AI功能，按使用扣除，当月有效</p>
            <div>
              <span>对标拆解（每次）<b>10 积分</b></span>
              <span>可核实线索（每条）<b>10~30 积分</b></span>
              <span>AI 生成内容（每次）<b>5 积分</b></span>
            </div>
          </article>
          <article>
            <h2>包含权益</h2>
            {benefits.map((item) => <p key={item}>✓ {item}</p>)}
          </article>
        </aside>
      </section>

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
          {orders.map((order) => (
            <div className="order-row" key={order[0]}>
              {order.map((item, index) => (
                <span className={index === 3 ? "paid" : ""} key={`${order[0]}-${item}`}>{item}</span>
              ))}
              <a href="/membership">查看详情</a>
            </div>
          ))}
        </div>
      </section>

      <h2 className="sr-only">会员与账单</h2>
      <h2 className="sr-only">企业版 会员生效</h2>
    </main>
  );
}

function RedeemPage() {
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
        <article className="cdk-redeem-card">
          <div className="cdk-redeem-tabs">
            <button className="active" type="button">兑换会员/权益</button>
            <button type="button">兑换积分</button>
          </div>
          <label>
            兑换码
            <div>
              <input aria-label="兑换码" placeholder="请输入16-32位兑换码（区分大小写）" />
              <button type="button">粘贴兑换码</button>
            </div>
          </label>
          <button type="button">立即兑换</button>
          <section>
            <h2>兑换码说明</h2>
            <ul>
              <li>兑换码可通过官方活动、合作伙伴、购买赠送等方式获得</li>
              <li>兑换成功后，权益或积分将自动发放到您的账户</li>
              <li>兑换码一旦使用，不可退换或转让</li>
              <li>如遇问题，请联系客服</li>
            </ul>
          </section>
        </article>

        <aside className="cdk-redeem-side">
          <section>
            <h2>当前账户信息</h2>
            <div><span>当前会员</span><strong>标准会员</strong><Link to="/membership">查看详情 ›</Link></div>
            <div><span>当前积分</span><strong>2,400 积分</strong><Link to="/membership">去购买 ›</Link></div>
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
