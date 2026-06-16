import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { useAuthSession } from "../lib/authSession";

const quotaRows = [
  ["AI智算额度", "8,320", "20,000", 41],
  ["数据获取额度", "120", "500", 24],
  ["工具使用额度", "35", "100", 35],
  ["商业沙盘推演", "3", "10", 30],
  ["竞品全盘数据破解", "1", "5", 20]
] as const;

const orders = [
  ["ZS-20250531-0012", "企业版（续行）", "2025-05-31 10:25", "已支付", "¥9,999.00", "已开票"],
  ["ZS-20240428-0018", "企业版（新付）", "2024-04-28 09:18", "已支付", "¥9,999.00", "已开票"],
  ["ZS-20231201-0015", "企业版（新付）", "2023-12-01 14:32", "已支付", "¥9,999.00", "已开票"],
  ["ZS-20230915-0003", "企业版（新付）", "2023-09-15 11:07", "已支付", "¥2,999.00", "已开票"],
  ["ZS-20230810-0002", "企业版（新付）", "2025-08-10 16:44", "已支付", "¥2,999.00", "已开票"]
] as const;

type MembershipPageProps = {
  showUpgrade?: boolean;
};

function MembershipPage({ showUpgrade = false }: MembershipPageProps) {
  const session = useAuthSession();
  const nickname = session.user?.nickname || "张婧";

  return (
    <V4PageShell>
      <section className="profile-page billing-page" aria-label="会员与账单">
        <div className="page-title-row">
          <div>
            <h1>会员与账单</h1>
            <p>管理订阅套餐、额度使用与账单记录</p>
          </div>
        </div>

        <div className="profile-layout billing-layout">
          <aside className="profile-side-tabs" aria-label="个人中心导航">
            {[
              ["个人中心", "/profile"],
              ["账号与资料设置", "/profile/settings"],
              ["会员与账单", "/membership"],
              ["我的内容", "/profile/content"],
              ["偏好设置", "/profile/preferences"]
            ].map(([label, href]) => (
              <Link key={label} className={label === "会员与账单" ? "active" : ""} to={href}>
                {label}
                <span aria-hidden="true">›</span>
              </Link>
            ))}
          </aside>

          <div className="profile-content">
            <section className="billing-plan-card">
              <div className="plan-main">
                <span className="plan-crown" aria-hidden="true">A</span>
                <div>
                  <small>当前套餐</small>
                  <h2>企业版 <b>会员生效</b></h2>
                  <p>套餐有效期 <strong>2025-12-31</strong> <em>剩余 213 天</em></p>
                </div>
              </div>
              <div className="plan-actions">
                <Link to="/membership/upgrade">升级套餐</Link>
                <button type="button">立即续费</button>
              </div>
              <div className="plan-summary">
                <strong>权益摘要</strong>
                {["团队成员上限", "项目数上限", "数据策略查询", "专属客户成功", "更多高级功能"].map((item, index) => (
                  <span key={item}>{item}<b>{index === 1 ? "不限" : index === 3 ? "1v1服务" : index === 4 ? "全部开放" : ""}</b></span>
                ))}
              </div>
            </section>

            <section className="billing-quota-card">
              <div className="module-section-head">
                <div>
                  <h2>各功能额度</h2>
                  <p>本月周期：2025-06-01 至 2025-06-30</p>
                </div>
                <span>所有额度均按自然月重置，本月重置日：2025-06-01</span>
              </div>
              <div className="quota-grid billing-quota-grid">
                {quotaRows.map(([label, used, total, percent]) => (
                  <article key={label} className="quota-card">
                    <span>{label}</span>
                    <strong>{used}<small> / {total}</small></strong>
                    <div className="quota-bar" aria-hidden="true"><i style={{ width: `${percent}%` }} /></div>
                    <small>剩余 {percent}%</small>
                    <small>重置日：2025-06-01</small>
                  </article>
                ))}
              </div>
            </section>

            <section className="order-card">
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
              <button className="view-all-order" type="button">查看全部订单</button>
            </section>
          </div>

          <aside className="billing-copilot-card" aria-label="智活 Copilot">
            <strong>智活 <b>Copilot</b></strong>
            <p>你的全球 AI 助手，随时为你提供帮助</p>
            <article>嗨，{nickname}！今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</article>
            <article className="blue">帮我分析一下智能客服系统的市场机会和落地关键点。</article>
            <article>好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点。</article>
            <Link to="/analysis">分析项目机会 ›</Link>
            <Link to="/tools">推荐工具 ›</Link>
            <Link to="/tasks">制定落地计划 ›</Link>
          </aside>
        </div>
        {showUpgrade && <MembershipUpgradeModal />}
      </section>
    </V4PageShell>
  );
}

function MembershipUpgradeModal() {
  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="升级套餐">
      <section className="upgrade-modal">
        <div className="upgrade-head">
          <div>
            <h2>升级套餐</h2>
            <p>对比普通版与会员版权益，选择更适合你的方案</p>
          </div>
          <div className="billing-toggle" aria-label="计费周期">
            <button type="button">月付</button>
            <button className="active" type="button">年付 <span>更优惠</span></button>
          </div>
          <Link className="modal-close" to="/membership" aria-label="关闭升级套餐">×</Link>
        </div>

        <div className="upgrade-plan-grid">
          <article className="upgrade-plan-card">
            <div className="plan-label">
              <span className="plan-mark plain" aria-hidden="true" />
              <strong>普通版</strong>
              <small>当前方案</small>
            </div>
            <p><b>¥0</b> / 免费</p>
            {["商业沙盘推演：1 次/月", "竞品全盘数据破解：5 次/月", "竞品动态监测：共用查询额度", "其他功能：本版暂不限次", "默认单一模型"].map((item) => (
              <span key={item}>{item}</span>
            ))}
          </article>

          <article className="upgrade-plan-card featured">
            <i>推荐</i>
            <div className="plan-label">
              <span className="plan-mark crown" aria-hidden="true" />
              <strong>会员版</strong>
              <small>推荐</small>
            </div>
            <p><b>¥980</b> / 年 <em>折合 ¥81.67 / 月</em></p>
            {[
              "商业沙盘推演：20 次/月",
              "竞品全盘数据破解：200 次/月",
              "竞品动态监测：共用查询额度",
              "其他功能：本版暂不限次",
              "支持 GPT / Claude / Grok 模型切换",
              "支持深度思考"
            ].map((item) => (
              <span key={item}>{item}</span>
            ))}
          </article>
        </div>

        <div className="benefit-table">
          <h3>权益对比一览</h3>
          {[
            ["商业沙盘推演", "1 次 / 月", "20 次 / 月"],
            ["竞品全盘数据破解查询", "5 次 / 月", "200 次 / 月"],
            ["竞品动态监测", "共用查询额度", "共用查询额度"],
            ["模型切换", "默认单一模型", "支持 GPT / Claude / Grok 模型切换"],
            ["其他功能", "本版暂不限次", "本版暂不限次"]
          ].map(([name, normal, vip]) => (
            <div key={name}>
              <span>{name}</span>
              <span>{normal}</span>
              <strong>{vip}</strong>
            </div>
          ))}
        </div>

        <footer className="upgrade-footer">
          <span className="qr-box compact" aria-hidden="true" />
          <div>
            <strong>联系企业微信升级权限</strong>
            <small>添加顾问，获取开通帮助</small>
          </div>
          <Link to="/membership">稍后再说</Link>
          <button type="button">立即开通</button>
        </footer>
      </section>
    </div>
  );
}

export default MembershipPage;
