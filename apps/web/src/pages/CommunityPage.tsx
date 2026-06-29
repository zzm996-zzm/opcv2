import { Link } from "react-router-dom";

import { CdkTopNav } from "./AnalysisPage";

const freeFeatures = [
  ["每周干货分享", "精选实战内容"],
  ["同行经验交流", "互助答疑解惑"],
  ["资源信息互通", "项目 · 渠道 · 工具"]
] as const;

const vipFeatures = [
  ["大咖主题直播", "每周深度分享"],
  ["专家问诊答疑", "1V1定向解答"],
  ["精选资源对接", "人脉 · 资本 · 渠道"],
  ["线下私享活动", "高质量闭门会"]
] as const;

const communityPosts = [
  ["小林创业中", "AI如何搭建私域的3个关键动作", "2小时前", "精华", "23", "12"],
  ["在线上", "AI+内容如何打造低成本获客闭环？", "5小时前", "", "18", "9"],
  ["运营老张", "7天提升转化率的落地SOP分享", "昨天", "", "31", "15"]
] as const;

const events = [
  ["直播分享", "企业私域增长的底层逻辑与实操打法", "智活AI增长顾问 · 老K", "预约"],
  ["案例拆解", "从冷启动到月入百万：真实案例拆解", "私域操盘手 · Abby", "预约"],
  ["线下沙龙", "深圳创业者线下闭门交流会（限定20人）", "智活AI · 社群运营", "报名"]
] as const;

const valueStats = [
  ["活跃成员", "1,200+", "本周新增 67"],
  ["本周互动", "328", "话题回复数"],
  ["干货分享", "56", "本周新增"],
  ["资源对接", "89", "本周新增"]
] as const;

const growthSteps = [
  ["1", "免费分析", "输入你的项目与目标，AI为你生成分析报告与增长建议"],
  ["2", "社群沉淀", "加入免费社群，学习方法、组织伙伴、互助成长"],
  ["3", "VIP获客", "解锁更多工具与资源，精准获客，加速增长"],
  ["4", "私域一对一", "专属顾问深度陪伴，定制方案，长期增长"]
] as const;

function CommunityPage() {
  return (
    <main className="cdk-analysis-page cdk-community-page">
      <CdkTopNav active="社群" />
      <section className="community-page cdk-community-shell" aria-label="AI社群">
        <main className="community-main">
          <header className="community-hero">
            <span>连接 · 学习 · 成长</span>
            <div className="community-hero-row">
              <div>
                <h1>加入智活社群，与优秀创业者一起增长</h1>
                <p>从免费分析到VIP获客，我们陪伴你每一步成长，助力生意持续增长。</p>
              </div>
              <aside>
                <div className="community-avatar-stack" aria-hidden="true">
                  <i />
                  <i />
                  <i />
                  <i />
                </div>
                <strong>已聚集 1,200+ 创业者一起成长</strong>
              </aside>
            </div>
          </header>

          <section className="community-entry-grid" aria-label="社群入口">
            <article className="community-entry-card free">
              <span>免费社群</span>
              <h2>创业成长互助社区</h2>
              <p>免费分析用户专属 · 共同学习 · 资源互助</p>
              <small>适合初创者、正在探索方向的你，获得方法、案例与伙伴支持。</small>
              <div className="community-feature-row">
                {freeFeatures.map(([title, desc]) => (
                  <section key={title}>
                    <i aria-hidden="true" />
                    <strong>{title}</strong>
                    <small>{desc}</small>
                  </section>
                ))}
              </div>
              <Link to="/community/members">免费加入社群</Link>
              <footer>已加入 892 人 · 本周新增 67 人</footer>
            </article>

            <article className="community-entry-card vip">
              <span>VIP社群</span>
              <h2>企业家陪伴成长圈</h2>
              <p>VIP/企业用户专属 · 深度链接 · 高价值陪伴</p>
              <small>面向成长型创业者与企业主，链接优质人脉与资源，解决关键增长难题。</small>
              <div className="community-feature-row vip-grid">
                {vipFeatures.map(([title, desc]) => (
                  <section key={title}>
                    <i aria-hidden="true" />
                    <strong>{title}</strong>
                    <small>{desc}</small>
                  </section>
                ))}
              </div>
              <Link to="/community/enterprise">升级VIP加入</Link>
              <footer>已加入 326 家企业 · 续费率 78%</footer>
            </article>
          </section>

          <section className="community-info-grid">
            <article className="community-panel community-posts">
              <header>
                <h2>社群动态</h2>
                <button type="button">最新评论</button>
                <Link to="/community/members">查看全部 ›</Link>
              </header>
              {communityPosts.map(([author, title, time, tag, likes, comments]) => (
                <section key={title}>
                  <i aria-hidden="true" />
                  <div>
                    <strong>{title}</strong>
                    <small>{author} · {time}</small>
                  </div>
                  {tag && <span>{tag}</span>}
                  <footer aria-label="互动数据">♡ {likes} / ◎ {comments}</footer>
                </section>
              ))}
              <Link className="community-panel-link" to="/community/members">查看全部讨论 ›</Link>
            </article>

            <article className="community-panel community-events">
              <header>
                <h2>本周活动预告</h2>
                <Link to="/community/members">全部活动 ›</Link>
              </header>
              {events.map(([type, title, host, action]) => (
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

            <article className="community-panel community-value">
              <header>
                <h2>社群价值数据</h2>
                <small>社区成长中</small>
              </header>
              <div className="community-value-grid">
                {valueStats.map(([label, value, note]) => (
                  <section key={label}>
                    <span>{label}</span>
                    <strong>{value}</strong>
                    <small>{note}</small>
                  </section>
                ))}
              </div>
              <blockquote>
                在社群里认识了很多同频的创业者，获得了宝贵的建议和资源，少走了很多弯路。
                <cite>Lisa · 教育行业创始人</cite>
              </blockquote>
            </article>
          </section>

          <p className="community-hidden-report">智能客服系统机会分析报告</p>

          <section className="community-growth-path" aria-label="你的成长路径">
            <h2>你的成长路径</h2>
            <p>从陌生到信任，从增长到成功</p>
            <div>
              {growthSteps.map(([step, title, desc]) => (
                <article key={step}>
                  <i>{step}</i>
                  <strong>{title}</strong>
                  <small>{desc}</small>
                </article>
              ))}
            </div>
          </section>
        </main>
      </section>
    </main>
  );
}

export default CommunityPage;
