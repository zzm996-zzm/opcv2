import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { authSession } from "../lib/authSession";
import { contentApi, type CommunityConfig } from "../lib/contentApi";
import { MiniCopilotForm } from "./MiniCopilot";
import ReferenceShell from "./ReferenceShell";
import UnifiedCopilotPanel from "./UnifiedCopilotPanel";
import V4PageShell from "./V4PageShell";

export type CommunityReferenceVariant = "overview" | "members" | "enterprise";

type CommunityReferencePageProps = {
  variant: CommunityReferenceVariant;
};

const overviewMemberFeatures = [
  ["每周干货分享", "精选实战内容"],
  ["同行经验交流", "互助答疑解惑"],
  ["资源信息互通", "项目 · 渠道 · 工具"]
] as const;

const overviewEnterpriseFeatures = [
  ["大咖主题直播", "每周深度分享"],
  ["专家问诊答疑", "1V1定向解答"],
  ["精准资源对接", "人脉 · 资本 · 渠道"],
  ["线下私享活动", "高质量闭门会"]
] as const;

const memberFeatures = [
  ["经验分享", "实战经验与避坑指南"],
  ["同行交流", "链接优质伙伴"],
  ["资源互通", "信息与资源高效对接"],
  ["共同成长", "携手前进，一起突破"]
] as const;

const enterpriseFeatures = [
  ["深度对接", "优质资源精准匹配"],
  ["线下活动", "高质量闭门交流"]
] as const;

const posts = [
  ["从0到1搭建私域的3个关键动作", "小林创业中 · 2小时前", "23", "12"],
  ["AI+内容如何打造低成本获客闭环？", "在路上 · 5小时前", "18", "9"],
  ["7天提升转化量的落地SOP分享", "运营专家 · 昨天", "31", "15"]
] as const;

const events = [
  ["直播分享", "企业私域增长的底层逻辑与实操打法", "智活AI增长顾问 · 老K", "预约"],
  ["案例拆解", "从冷启动到月入百万：真实案例拆解", "私域操盘手 · Abby", "预约"],
  ["线下沙龙", "深圳创业者线下闭门交流会（限定20人）", "智活AI · 社群运营", "报名"]
] as const;

const valueItems = [
  ["连接高价值人脉", "连接 1,500+ 优质创业者与企业决策者"],
  ["实战经验共享", "真实案例、避坑指南、增长方法"],
  ["优质资源对接", "资金、渠道、技术、人才资源"],
  ["行业机会发现", "洞察趋势、发现合作与增长机会"]
] as const;

const growthSteps = [
  ["1", "免费分析", "输入你的项目与目标，AI为你生成分析报告与增长建议"],
  ["2", "社群沉淀", "加入免费社群，学习方法、组织伙伴、互助成长"],
  ["3", "VIP获客", "解锁更多工具与资源，精准获客，加速增长"],
  ["4", "私域一对一", "专属顾问深度陪伴，定制方案，长期增长"]
] as const;

const memberStats = [
  ["活跃成员", "1,204+", "较上周 +18%"],
  ["本周互动", "358", "较上周 +22%"],
  ["干货分享", "62", "本周新增"],
  ["资源对接", "95", "本周新增"]
] as const;

function CommunityOverview({ config }: { config: CommunityConfig | null }) {
  return (
    <ReferenceShell className="ref-community-shell" mainClassName="ref-community-page">
      <section className="ref-community-layout" aria-label="AI社群">
        <main className="ref-community-main">
          <header className="ref-community-heading">
            <span>◉ 连接 · 学习 · 成长</span>
            <h1 className="ref-community-sr-only">{config?.headline ?? "加入智活社群，与优秀创业者一起增长"}</h1>
            <p>{config?.description ?? "从免费分析到VIP获客，我们陪伴你每一步成长，助力生意持续增长。"}</p>
            <aside>
              <div>{[1, 4, 6].map((item) => <img alt="" key={item} src={`/community/avatar-0${item}.jpg`} />)}</div>
              <strong>已聚集 1,200+ 创业者一起成长</strong>
            </aside>
          </header>

          <section className="ref-community-entries" aria-label="社群入口">
            <article className="ref-community-card member">
              <img alt="" src="/community/member-hero.jpg" />
              <span>免费社群</span>
              <h1>创业成长互助社区</h1>
              <p>免费分析用户专属 · 共同学习 · 资源互助</p>
              <small>适合初创者、正在探索方向的你，获得方法、案例与伙伴支持。</small>
              <div className="ref-community-features three">
                {overviewMemberFeatures.map(([title, description], index) => (
                  <section key={title}><i>{["▣", "♙", "▤"][index]}</i><span><strong>{title}</strong><small>{description}</small></span></section>
                ))}
              </div>
              <Link to="/community/members">♧ 免费加入社群</Link>
              <footer>已加入 1,204 人 · 本周新增 86 人</footer>
            </article>

            <article className="ref-community-card enterprise">
              <img alt="" src="/community/enterprise-hero.jpg" />
              <span>VIP社群</span>
              <h1>企业家陪伴成长圈</h1>
              <p>VIP/企业用户专属 · 深度链接 · 高价值陪伴</p>
              <small>面向成长型创业者与企业主，链接优质人脉与资源，解决关键增长难题。</small>
              <div className="ref-community-features">
                {overviewEnterpriseFeatures.map(([title, description], index) => (
                  <section key={title}><i>{["▣", "♙", "▤", "⌁"][index]}</i><span><strong>{title}</strong><small>{description}</small></span></section>
                ))}
              </div>
              <Link to="/community/enterprise">♛ 升级VIP加入</Link>
              <footer>已加入 326 家企业 · 活跃度 78%</footer>
            </article>
          </section>

          <section className="ref-community-info">
            <article className="ref-community-panel">
              <header><h2>社群动态 <small>最新讨论</small></h2><Link to="/community/members">查看全部 ›</Link></header>
              {posts.map(([title, meta, likes, comments], index) => (
                <div className="ref-community-list-row" key={title}><img alt="" src={`/community/avatar-0${index + 1}.jpg`} /><span><strong>{title}</strong><small>{meta}</small></span><em>♧ {likes} <span>◯ {comments}</span></em></div>
              ))}
              <Link to="/community/members">查看全部讨论 ›</Link>
            </article>
            <article className="ref-community-panel">
              <header><h2>本周活动预告</h2><Link to="/community/members">全部活动 ›</Link></header>
              {events.map(([tag, title, host, action], index) => (
                <div className="ref-community-list-row event" key={title}><img alt="" src={`/community/avatar-0${index + 4}.jpg`} /><span><b>{tag}</b><strong>{title}</strong><small>{host}</small></span><button type="button">{action}</button></div>
              ))}
            </article>
            <article className="ref-community-panel ref-community-stats">
              <header><h2>社群价值数据</h2><strong>社区成长中</strong></header>
              <div>{[["活跃成员","1,200+"],["本周互动","328"],["干货分享","56"],["资源对接","89"]].map(([label,value]) => <section key={label}><small>{label}</small><b>{value}</b></section>)}</div>
              <blockquote>在社群里认识了很多同频的创业者，获得了宝贵的建议和资源，少走了很多弯路。<cite>— Lisa · 教育行业创始人</cite></blockquote>
            </article>
          </section>

          <section className="ref-community-growth">
            <header><h2>你的成长路径 <small>（从陌生到信任，从增长到成功）</small></h2></header>
            <div>{growthSteps.map(([step, title, desc]) => <article key={step}><i>{step}</i><span><strong>{title}</strong><small>{desc}</small></span></article>)}</div>
          </section>
        </main>

        <aside className="ref-community-copilot" aria-label="智活 Copilot 社群助手">
          <header><div><strong><span>✦</span> 智活 <b>Copilot</b></strong><p>你的全能 AI 助手，随时为你提供帮助</p></div><div className="ref-community-copilot-tools"><span>⚙</span><span>⌃</span></div></header>
          <div className="ref-community-chat">
            <article><span>A</span><p>嗨，张婧！<br />今天想聚焦哪个方向？我可以帮你分析机会、推荐工具或制定落地计划。</p></article>
            <small>猜你想问</small>
            <div className="question">帮我分析一下智能硬件赛道的市场机会</div>
            <article><span>A</span><p>好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点。</p></article>
          </div>
          <section className="ref-community-report"><i>▦</i><span><strong>智能硬件市场机会分析报告</strong><small>PDF · 2.4MB</small></span></section>
          <nav><Link to="/analysis">▥ 分析项目机会 <span>›</span></Link><Link to="/tools/recommend">▣ 推荐工具 <span>›</span></Link><Link to="/learning/plan">▤ 制定落地计划 <span>›</span></Link></nav>
          <MiniCopilotForm className="ref-community-input" placeholder="询问任何问题..." sendIcon="↗" />
        </aside>
      </section>
    </ReferenceShell>
  );
}

function CommunityCard({ currentVariant, onOpen, overview, type }: {
  currentVariant: CommunityReferenceVariant;
  onOpen: () => void;
  overview: boolean;
  type: "member" | "enterprise";
}) {
  const isMember = type === "member";
  const isCurrentRoute = currentVariant === (isMember ? "members" : "enterprise");
  const features = overview
    ? isMember ? overviewMemberFeatures : overviewEnterpriseFeatures
    : isMember ? memberFeatures : enterpriseFeatures;

  return (
    <article className={`community-ref-entry ${isMember ? "member" : "enterprise"}`}>
      <img alt="" src={isMember ? "/community/member-hero.jpg" : "/community/enterprise-hero.jpg"} />
      <span>{overview ? isMember ? "免费社群" : "VIP社群" : isMember ? "会员社群" : "企业社群"}</span>
      <h2>{overview ? isMember ? "创业成长互助社区" : "企业家陪伴成长圈" : isMember ? "创业成长互助社区" : "企业决策者交流圈"}</h2>
      <p>{overview
        ? isMember ? "免费分析用户专属 · 共同学习 · 资源互助" : "VIP/企业用户专属 · 深度链接 · 高价值陪伴"
        : isMember ? "与1,200+ 创业者一起学习、交流、成长" : "面向企业主、投资人、行业专家，共享资源与机会"}</p>
      {overview ? <small>{isMember ? "适合初创者、正在探索方向的你，获得方法、案例与伙伴支持。" : "面向成长型创业者与企业主，链接优质人脉与资源，解决关键增长难题。"}</small> : null}
      <div className={`community-ref-features ${features.length === 3 ? "three" : ""}`}>
        {features.map(([title, description], index) => (
          <section key={title}>
            <i aria-hidden="true">{["▣", "♙", "▤", "⌁"][index]}</i>
            <span><strong>{title}</strong><small>{description}</small></span>
          </section>
        ))}
      </div>
      <Link onClick={isCurrentRoute ? onOpen : undefined} to={isMember ? "/community/members" : "/community/enterprise"}>
        <span aria-hidden="true">{isMember ? "♧" : "♛"}</span>
        {overview ? isMember ? "免费加入社群" : "升级VIP加入" : isMember ? "加入会员社群" : "加入企业社群"}
      </Link>
      <footer>{isMember ? "已加入 1,204 人 · 本周新增 86 人" : "已加入 326 家企业 · 活跃度 78%"}</footer>
    </article>
  );
}

function CommunityPosts({ compact = false }: { compact?: boolean }) {
  return (
    <div className={`community-ref-post-list ${compact ? "compact" : ""}`}>
      {posts.map(([title, meta, likes, comments], index) => (
        <article key={title}>
          <img alt="" src={`/community/avatar-0${index + 1}.jpg`} />
          <div><strong>{title}</strong><small>{meta}</small></div>
          <span>♧ {likes}</span><span>◯ {comments}</span>
        </article>
      ))}
    </div>
  );
}

function CommunityEvents({ compact = false }: { compact?: boolean }) {
  return (
    <div className={`community-ref-event-list ${compact ? "compact" : ""}`}>
      {events.map(([tag, title, host, action], index) => (
        <article key={title}>
          <img alt="" src={`/community/avatar-0${index + 4}.jpg`} />
          <div><span>{tag}</span><strong>{title}</strong><small>{host}</small></div>
          <button type="button">{action}</button>
        </article>
      ))}
    </div>
  );
}

function CommunityStats({ overview }: { overview: boolean }) {
  const stats = overview
    ? [["活跃成员", "1,200+", "本周新增 67"], ["本周互动", "328", "话题回复"], ["干货分享", "56", "本周新增"], ["资源对接", "89", "本周新增"]]
    : memberStats;

  return (
    <div className="community-ref-stat-grid">
      {stats.map(([label, value, note]) => <section key={label}><span>{label}</span><strong>{value}</strong><small>{note}</small></section>)}
    </div>
  );
}

function CommunityCopilot() {
  return <UnifiedCopilotPanel
    actions={[
      { href: "/analysis", label: "分析项目机会" },
      { href: "/tools/recommend", label: "推荐工具" },
      { href: "/learning/plan", label: "制定落地计划" }
    ]}
    ariaLabel="智活 Copilot 社群助手"
    className="community-ref-copilot"
    inputAriaLabel="询问社群 Copilot"
    report={{ title: "智能硬件市场机会分析报告", meta: "PDF · 2.4 MB" }}
    response="好的，已为你生成分析报告，包含市场规模、竞争格局和落地要点。"
    userPrompt="帮我分析一下智能硬件赛道的市场机会。"
  />;
}

function JoinModal({ config, onClose, onSubmit, status, variant }: {
  config: CommunityConfig | null;
  onClose: () => void;
  onSubmit: () => void;
  status: string;
  variant: "members" | "enterprise";
}) {
  const isMember = variant === "members";
  const qrVariant = config?.qr_variants?.find((item) => item.key === variant);
  const fallbackCard = isMember ? "/community/member-join-card.png" : "/community/enterprise-join-card.png";
  const useFallbackCard = !qrVariant;
  const dialogLabel = isMember ? "加入会员社群" : "加入企业社群";
  const qrLabel = isMember ? "会员社群二维码" : "企业社群二维码";

  return (
    <div className={`community-ref-modal-layer ${isMember ? "member" : "enterprise"}`}>
      <button aria-label="关闭入群弹窗遮罩" className="community-ref-modal-scrim" onClick={onClose} type="button" />
      <section aria-label={dialogLabel} aria-modal="true" className={`community-ref-modal ${useFallbackCard ? "template" : ""}`} role="dialog">
        <button aria-label={isMember ? "关闭加入会员社群弹窗" : "关闭加入企业社群弹窗"} onClick={onClose} type="button">×</button>
        {useFallbackCard ? (
          <>
            <h2 className="ref-community-sr-only">{dialogLabel}</h2>
            <img alt={qrLabel} className="community-ref-join-card" src={fallbackCard} />
          </>
        ) : (
          <>
            {!isMember ? <i className="community-ref-crown" aria-hidden="true">♛</i> : null}
            <h2>{dialogLabel}</h2>
            <p>{config?.headline ?? (isMember ? <>与 <b>1,200+</b> 创业者一起交流成长</> : <>与 <b>300+</b> 企业决策者一起链接资源</>)}</p>
            <div className="community-ref-qr">
              <img alt={qrVariant.label || qrLabel} src={qrVariant.image_url} />
            </div>
            <small><span aria-hidden="true">●●</span>{qrVariant.description || "使用微信扫一扫，添加社群小助手，拉你入群"}</small>
            {qrVariant.join_url ? <a href={qrVariant.join_url}>打开入群链接</a> : null}
            {status ? <strong role="status">{status}</strong> : null}
          </>
        )}
        <button onClick={onSubmit} type="button">我知道了</button>
      </section>
    </div>
  );
}

function CommunityReferencePage({ variant }: CommunityReferencePageProps) {
  const overview = variant === "overview";
  const [config, setConfig] = useState<CommunityConfig | null>(null);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("");
  const [modalOpen, setModalOpen] = useState(!overview);

  useEffect(() => {
    let active = true;
    contentApi.getCommunityConfig()
      .then((payload) => { if (active) setConfig(payload); })
      .catch(() => {
        // Reference content remains available when the optional community config is offline.
      });
    return () => { active = false; };
  }, []);

  async function submitJoinRequest() {
    if (overview) return;
    setError("");
    try {
      const user = authSession.get().user;
      await contentApi.joinCommunity({
        community: variant,
        contact: user?.phone || user?.account || user?.nickname || "unknown",
        note: variant === "members" ? "希望加入创业成长互助社区" : "希望加入企业决策者交流圈"
      });
      setStatus("申请已提交");
    } catch (caught) {
      setError(apiErrorMessage(caught, "申请提交失败，请稍后再试"));
    }
  }

  if (overview) return <CommunityOverview config={config} />;

  return (
    <V4PageShell className="community-ref-shell">
      <section className={`community-ref-page ${variant}`} aria-label={overview ? "AI社群" : variant === "members" ? "会员社群" : "企业社群"}>
        <main className="community-ref-main">
          {overview ? (
            <header className="community-ref-hero">
              <span>◉ 连接 · 学习 · 成长</span>
              <div><h1>{config?.headline ?? "加入智活社群，与优秀创业者一起增长"}</h1><p>{config?.description ?? "从免费分析到VIP获客，我们陪伴你每一步成长，助力生意持续增长。"}</p></div>
              <aside><div>{[1, 4, 6].map((item) => <img alt="" key={item} src={`/community/avatar-0${item}.jpg`} />)}</div><strong>已聚集 1,200+ 创业者一起成长</strong></aside>
            </header>
          ) : (
            <header className="community-ref-title"><h1>AI社群</h1><p>{config?.description ?? "连接高质量创业者与企业用户，分享经验、对接资源、共同成长"}</p></header>
          )}
          {error ? <small className="form-error community-ref-error" role="alert">{error}</small> : null}

          <section className="community-ref-entry-grid" aria-label="社群入口">
            <CommunityCard currentVariant={variant} onOpen={() => setModalOpen(true)} overview={overview} type="member" />
            <CommunityCard currentVariant={variant} onOpen={() => setModalOpen(true)} overview={overview} type="enterprise" />
          </section>

          {overview ? (
            <section className="community-ref-info-grid">
              <article className="community-ref-panel"><header><h2>社群动态 <small>最新讨论</small></h2><Link to="/community/members">查看全部 ›</Link></header><CommunityPosts /><Link to="/community/members">查看全部讨论 ›</Link></article>
              <article className="community-ref-panel"><header><h2>本周活动预告</h2><Link to="/community/members">全部活动 ›</Link></header><CommunityEvents /></article>
              <article className="community-ref-panel community-ref-overview-stats"><header><h2>社群价值数据</h2><small>社区成长中</small></header><CommunityStats overview /><blockquote>在社群里认识了很多同频的创业者，获得了宝贵的建议和资源，少走了很多弯路。<cite>— Lisa · 教育行业创始人</cite></blockquote></article>
            </section>
          ) : (
            <section className="community-ref-member-grid">
              <article className="community-ref-panel community-ref-values"><h2>社群价值</h2>{valueItems.map(([title, desc], index) => <section key={title}><i>{["♧", "◴", "⌘", "▣"][index]}</i><div><strong>{title}</strong><small>{desc}</small></div></section>)}</article>
              <article className="community-ref-panel community-ref-member-feed"><header><h2>社群动态</h2><Link to={`/community/${variant}`}>查看全部 ›</Link></header><div><CommunityPosts compact /><CommunityEvents compact /></div><Link to={`/community/${variant}`}>查看全部动态 ›</Link></article>
              <article className="community-ref-panel community-ref-member-stats"><h2>社群价值数据</h2><CommunityStats overview={false} /></article>
            </section>
          )}

          {overview ? <section className="community-ref-growth"><header><h2>你的成长路径 <small>（从陌生到信任，从增长到成功）</small></h2></header><div>{growthSteps.map(([step, title, desc]) => <article key={step}><i>{step}</i><div><strong>{title}</strong><small>{desc}</small></div></article>)}</div></section> : null}
        </main>
        <CommunityCopilot />
        {modalOpen && !overview ? <JoinModal config={config} onClose={() => setModalOpen(false)} onSubmit={submitJoinRequest} status={status} variant={variant} /> : null}
      </section>
    </V4PageShell>
  );
}

export default CommunityReferencePage;
