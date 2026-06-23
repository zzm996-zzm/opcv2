import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

const monitoringStats = [
  ["监测中竞品", "12"],
  ["今日新增动态", "18"],
  ["高风险预警", "4"],
  ["已生成任务", "7"]
] as const;

const trackedCompetitors = [
  {
    name: "小鹅通",
    category: "知识付费 / 企业培训",
    status: "高频变化",
    threat: "强",
    lastSeen: "12 分钟前",
    channels: ["价格页", "招聘", "公众号"],
    signal: "新增 AI 助教套餐介绍，并同步发布 3 个直播转化案例。"
  },
  {
    name: "有赞教育",
    category: "私域运营 / 教育 SaaS",
    status: "稳定监测",
    threat: "中",
    lastSeen: "46 分钟前",
    channels: ["案例页", "SEO", "投放"],
    signal: "官网案例页新增连锁校区样板，关键词从开课转向门店增长。"
  },
  {
    name: "企微管家",
    category: "CRM / 客户运营",
    status: "定位漂移",
    threat: "中",
    lastSeen: "1 小时前",
    channels: ["招聘", "内容矩阵", "产品页"],
    signal: "销售岗位 JD 强调 AI 线索跟进，内容标题开始绑定客户分层。"
  }
] as const;

const timeline = [
  ["09:42", "小鹅通", "价格页新增 AI 助教权益", "套餐页把直播答疑、课后作业批改和私域转化写入核心卖点。", "强"],
  ["10:18", "有赞教育", "案例页新增连锁培训机构", "突出多门店排课、统一运营和企微客户沉淀，适合做销售话术对照。", "中"],
  ["11:07", "企微管家", "招聘岗位出现 AI 客户运营", "新增 2 个增长运营岗位，要求会用 AI 做线索分层和内容触达。", "中"],
  ["12:26", "增长黑盒", "投放素材集中测试低价课", "素材主打 9.9 元训练营入口，后端承接企业内训方案。", "低"]
] as const;

const channelHealth = [
  ["官网 / 价格页", "每 6 小时", "正常"],
  ["招聘动态", "每 12 小时", "正常"],
  ["内容矩阵", "每 3 小时", "密集"],
  ["投放素材", "每日", "排队"]
] as const;

const alertRules = [
  ["价格变化", "捕捉套餐、权益、免费试用和低价入口调整"],
  ["招聘扩张", "识别销售、增长、AI 产品岗位的异常增长"],
  ["内容爆发", "监测公众号、视频号、SEO 页面标题变化"],
  ["产品转向", "从产品页和案例页判断定位、场景和客群变化"]
] as const;

const nextActions = [
  "生成小鹅通 AI 助教套餐对比稿",
  "把有赞教育连锁案例拆成销售问答",
  "为企微管家定位变化创建跟进任务"
] as const;

function CompetitorMonitoringPage() {
  return (
    <V4PageShell className="competitor-monitoring-shell">
      <section className="module-page competitor-monitoring-page" aria-label="竞品动态监测">
        <div className="page-title-row">
          <div>
            <h1>竞品动态监测</h1>
            <p>持续盯住竞品的价格、招聘、内容、投放和产品页变化，把异常信号自动沉淀成反击动作</p>
          </div>
          <button className="module-primary-action" type="button">新增监测对象</button>
        </div>

        <section className="module-overview-card competitor-monitoring-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">实时竞争雷达</span>
            <h2>从“偶尔看看竞品”升级成持续预警系统</h2>
            <p>监测规则会自动巡检竞品公开页面和内容渠道，识别高频变化、定位漂移与获客动作，并把可执行建议推送到任务中心。</p>
            <div className="module-stat-strip">
              {monitoringStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <div className="monitoring-radar-card" aria-label="竞品动态雷达">
            <div className="monitoring-radar">
              <span className="radar-sweep" aria-hidden="true" />
              <i className="dot hot" aria-hidden="true" />
              <i className="dot warm" aria-hidden="true" />
              <i className="dot cool" aria-hidden="true" />
              <strong>4</strong>
              <small>高风险预警</small>
            </div>
            <p>最近 24 小时内，价格页和内容矩阵出现连续变化，建议优先生成销售对比材料。</p>
          </div>
        </section>

        <section className="monitoring-grid">
          <div className="monitoring-main-card">
            <div className="module-section-head">
              <div>
                <h2>监测中竞品</h2>
                <p>按威胁等级和最近变化排序，点击后可进入全盘数据破解</p>
              </div>
              <div className="module-chip-row compact">
                {["全部", "强预警", "价格变化", "招聘扩张"].map((view, index) => (
                  <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
                ))}
              </div>
            </div>

            <div className="monitoring-competitor-list">
              {trackedCompetitors.map((item) => (
                <article key={item.name}>
                  <span className={`monitoring-pulse ${item.threat === "强" ? "hot" : ""}`} aria-hidden="true" />
                  <div>
                    <h3>{item.name}</h3>
                    <small>{item.category} · {item.lastSeen}</small>
                  </div>
                  <em className={item.threat === "强" ? "hot" : ""}>威胁 {item.threat}</em>
                  <span className="monitoring-state">{item.status}</span>
                  <p>{item.signal}</p>
                  <div className="tool-tags">
                    {item.channels.map((channel) => <span key={channel}>{channel}</span>)}
                  </div>
                </article>
              ))}
            </div>
          </div>

          <aside className="monitoring-side-card" aria-label="监测频率">
            <h2>监测频率</h2>
            {channelHealth.map(([source, frequency, status]) => (
              <article key={source}>
                <span>
                  <strong>{source}</strong>
                  <small>{frequency}</small>
                </span>
                <em>{status}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="monitoring-lower-grid">
          <div className="monitoring-timeline-card">
            <div className="module-section-head">
              <div>
                <h2>动态时间线</h2>
                <p>把散落的变化按时间串起来，方便判断竞品动作是否连续</p>
              </div>
            </div>
            <div className="monitoring-timeline">
              {timeline.map(([time, company, title, detail, level]) => (
                <article key={`${time}-${title}`}>
                  <time>{time}</time>
                  <div>
                    <strong>{company}</strong>
                    <h3>{title}</h3>
                    <p>{detail}</p>
                  </div>
                  <em className={level === "强" ? "hot" : ""}>{level}</em>
                </article>
              ))}
            </div>
          </div>

          <aside className="monitoring-action-card" aria-label="预警动作">
            <h2>今日预警</h2>
            <strong>先处理小鹅通价格页变化</strong>
            <p>该变化与 AI 助教、直播转化和企微私域同时关联，可能影响销售对比话术。</p>
            <div>
              {nextActions.map((action) => <span key={action}>{action}</span>)}
            </div>
            <Link to="/tasks">生成反击任务</Link>
          </aside>
        </section>

        <section className="monitoring-rule-section">
          <div className="module-section-head">
            <div>
              <h2>预警规则</h2>
              <p>第一版先用固定规则承接，后续可接入真实爬虫、队列和模型评分</p>
            </div>
          </div>
          <div className="monitoring-rule-grid">
            {alertRules.map(([title, detail]) => (
              <article key={title}>
                <strong>{title}</strong>
                <p>{detail}</p>
              </article>
            ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

export default CompetitorMonitoringPage;
