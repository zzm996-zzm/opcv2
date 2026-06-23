import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const memberFeatures = [
  ["经验分享", "实战经验与避坑指南"],
  ["同行交流", "结识同频伙伴"],
  ["资源互通", "信息与资源高效对接"],
  ["共同成长", "携手进步，一起突破"]
] as const;

const enterpriseFeatures = [
  ["资源对接", "优质资源精准匹配"],
  ["线下活动", "高质量闭门交流"]
] as const;

const valueItems = [
  ["高质量人脉圈", "连接 1,500+ 优质创业者与企业决策者"],
  ["实战经验共享", "真实案例、避坑指南、增长方法"],
  ["优质资源对接", "资金、渠道、技术、人才资源"],
  ["行业机会发现", "洞察趋势、发现合作与增长机会"]
] as const;

const memberPosts = [
  ["产品", "实战经验：如何用AI提升私域运营效率", "王磊 · 3小时前", "23", "18"],
  ["小林创业中", "从0到1搭建私域的3个关键动作", "2小时前", "18", "19"],
  ["在线上", "AI+内容如何打造低成本获客闭环？", "5小时前", "31", "15"]
] as const;

const weekEvents = [
  ["直播分享", "企业私域增长的底层逻辑与实操打法", "智活AI增长顾问 · 老K", "预约"],
  ["案例拆解", "从冷启动到月入百万：真实案例拆解", "私域操盘手 · Abby", "预约"],
  ["线下沙龙", "深圳创业者线下闭门交流会（限定20人）", "智活AI · 社群运营", "报名"]
] as const;

const memberStats = [
  ["活跃成员", "1,204+", "本周新增 86"],
  ["本周互动", "358", "话题回复数"],
  ["干货分享", "62", "本周新增"],
  ["资源对接", "95", "本周新增"]
] as const;

function CommunityMembersPage() {
  return (
    <V4PageShell>
      <section className="community-page community-members-page" aria-label="会员社群">
        <main className="community-main community-members-main">
          <header className="community-members-head">
            <h1>AI社群</h1>
            <p>连接高质量创业者与企业用户，分享经验、对接资源、共同成长</p>
          </header>

          <section className="community-entry-grid community-members-entry" aria-label="社群入口">
            <article className="community-entry-card community-member-card free">
              <span>会员社群</span>
              <h2>创业成长互助社区</h2>
              <p>与1,200+ 创业者一起学习、交流、成长</p>
              <div className="community-feature-row community-member-feature-grid">
                {memberFeatures.map(([title, desc]) => (
                  <section key={title}>
                    <i aria-hidden="true" />
                    <strong>{title}</strong>
                    <small>{desc}</small>
                  </section>
                ))}
              </div>
              <Link to="/community/members">加入会员社群</Link>
              <footer>已加入 1,204 人 · 本周新增 86 人</footer>
            </article>

            <article className="community-entry-card community-member-card vip">
              <span>企业社群</span>
              <h2>企业决策者交流圈</h2>
              <p>面向企业主、投资人、行业专家，共享资源与机会</p>
              <div className="community-feature-row community-enterprise-feature-grid">
                {enterpriseFeatures.map(([title, desc]) => (
                  <section key={title}>
                    <i aria-hidden="true" />
                    <strong>{title}</strong>
                    <small>{desc}</small>
                  </section>
                ))}
              </div>
              <Link to="/community/enterprise">加入企业社群</Link>
              <footer>已加入 326 家企业 · 活跃率 78%</footer>
            </article>
          </section>

          <section className="community-members-lower">
            <article className="community-panel community-member-value">
              <h2>社群价值</h2>
              {valueItems.map(([title, desc]) => (
                <section key={title}>
                  <i aria-hidden="true" />
                  <div>
                    <strong>{title}</strong>
                    <small>{desc}</small>
                  </div>
                </section>
              ))}
            </article>

            <article className="community-panel community-posts community-member-posts">
              <header>
                <h2>社群动态</h2>
              </header>
              {memberPosts.map(([author, title, meta, likes, comments]) => (
                <section key={title}>
                  <i aria-hidden="true" />
                  <div>
                    <strong>{title}</strong>
                    <small>{author} · {meta}</small>
                  </div>
                  <footer aria-label="互动数据">♡ {likes} / ◎ {comments}</footer>
                </section>
              ))}
              <Link className="community-panel-link" to="/community/members">查看全部动态 ›</Link>
            </article>

            <article className="community-panel community-events community-member-events">
              <header>
                <h2>本周活动预告</h2>
                <Link to="/community/members">查看全部 ›</Link>
              </header>
              {weekEvents.map(([type, title, host, action]) => (
                <section key={title}>
                  <i aria-hidden="true" />
                  <div>
                    <span>{type}</span>
                    <strong>{title}</strong>
                    <small>{host}</small>
                  </div>
                  <button type="button">{action}</button>
                </section>
              ))}
            </article>

            <article className="community-panel community-value community-member-stats">
              <h2>社群价值数据</h2>
              <div className="community-value-grid">
                {memberStats.map(([label, value, note]) => (
                  <section key={label}>
                    <span>{label}</span>
                    <strong>{value}</strong>
                    <small>{note}</small>
                  </section>
                ))}
              </div>
            </article>
          </section>

          <section className="community-join-modal" role="dialog" aria-label="加入会员社群" aria-modal="true">
            <button aria-label="关闭加入会员社群弹窗" type="button">×</button>
            <h2>加入会员社群</h2>
            <p>与 1,200+ 创业者一起交流成长</p>
            <div className="community-qr" aria-label="会员社群二维码">
              <i aria-hidden="true" />
            </div>
            <small>使用微信扫一扫，添加社群小助手，拉你入群</small>
            <button type="button">我知道了</button>
          </section>
        </main>

        <aside className="learning-copilot community-copilot community-members-copilot" aria-label="智活 Copilot 社群助手">
          <header>
            <div>
              <strong><span aria-hidden="true">✦</span> 智活 <b>Copilot</b></strong>
              <p>你的全球 AI 助手，随时为你提供帮助</p>
            </div>
            <div className="learning-copilot-tools" aria-hidden="true">
              <span>⚙</span>
              <span>⌃</span>
            </div>
          </header>

          <div className="learning-chat community-chat community-member-chat">
            <article>
              <span className="ai-avatar">A</span>
              <p>您好，张婧 👋<br />我是您的智能助手，可以帮您：<br />· 找工具、做分析、出方案、提建议</p>
            </article>
            <article>
              <span className="ai-avatar">A</span>
              <p>帮我分析一下智能硬件赛道的市场机会</p>
            </article>
          </div>

          <section className="community-report-card">
            <Link to="/analysis" aria-label="智能硬件市场分析报告.pdf">
              <i aria-hidden="true">PDF</i>
              <span><b>智能硬件市场分析报告.pdf</b><small>PDF · 2.4MB</small></span>
            </Link>
          </section>

          <nav className="learning-copilot-actions" aria-label="社群助手快捷入口">
            <Link to="/analysis">分析项目机会 <span aria-hidden="true">›</span></Link>
            <Link to="/tools/recommend">推荐工具 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">制定落地计划 <span aria-hidden="true">›</span></Link>
          </nav>

          <form className="learning-copilot-input">
            <button aria-label="添加附件" type="button">+</button>
            <input aria-label="向 Copilot 提问" placeholder="向我提问，或输入 @ 调用技能" />
            <button aria-label="发送" type="button">›</button>
          </form>
        </aside>
      </section>
    </V4PageShell>
  );
}

export default CommunityMembersPage;
