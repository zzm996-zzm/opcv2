import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { contentApi, type CommunityConfig } from "../lib/contentApi";
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

const growthSteps = [
  ["1", "免费分析", "输入你的项目与目标，AI为你生成分析报告与增长建议"],
  ["2", "社群沉淀", "加入免费社群，学习方法、组织伙伴、互助成长"],
  ["3", "VIP获客", "解锁更多工具与资源，精准获客，加速增长"],
  ["4", "私域一对一", "专属顾问深度陪伴，定制方案，长期增长"]
] as const;

function CommunityPage() {
  const [config, setConfig] = useState<CommunityConfig | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    contentApi
      .getCommunityConfig()
      .then((payload) => {
        if (active) setConfig(payload);
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取社群配置"));
      });
    return () => {
      active = false;
    };
  }, []);

  return (
    <main className="cdk-analysis-page cdk-community-page">
      <CdkTopNav active="社群" />
      <section className="community-page cdk-community-shell" aria-label="AI社群">
        <main className="community-main">
          <header className="community-hero">
            <span>连接 · 学习 · 成长</span>
            <div className="community-hero-row">
              <div>
                <h1>{config?.headline ?? "加入智活社群，与优秀创业者一起增长"}</h1>
                <p>{config?.description ?? "从免费分析到VIP获客，我们陪伴你每一步成长，助力生意持续增长。"}</p>
                {error && <small className="form-error" role="alert">{error}</small>}
              </div>
              <aside>
                <div className="community-avatar-stack" aria-hidden="true">
                  <i />
                  <i />
                  <i />
                  <i />
                </div>
                <strong>社群人数待接入</strong>
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
              <footer>成员数据待接入</footer>
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
              <footer>企业数据待接入</footer>
            </article>
          </section>

          <section className="community-info-grid">
            <article className="community-panel community-posts">
              <header>
                <h2>社群动态</h2>
                <Link to="/community/members">查看全部 ›</Link>
              </header>
              <p>暂无社群动态</p>
              <Link className="community-panel-link" to="/community/members">查看全部讨论 ›</Link>
            </article>

            <article className="community-panel community-events">
              <header>
                <h2>本周活动预告</h2>
                <Link to="/community/members">全部活动 ›</Link>
              </header>
              <p>暂无活动数据</p>
            </article>

            <article className="community-panel community-value">
              <header>
                <h2>社群价值数据</h2>
                <small>等待社区统计接口</small>
              </header>
              <p>暂无社群价值数据</p>
            </article>
          </section>

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
