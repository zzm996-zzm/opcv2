import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { MiniCopilotForm } from "../components/MiniCopilot";
import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { authSession } from "../lib/authSession";
import { contentApi, type CommunityConfig } from "../lib/contentApi";

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
  ["高质量人脉圈", "连接优质创业者与企业决策者"],
  ["实战经验共享", "真实案例、避坑指南、增长方法"],
  ["优质资源对接", "资金、渠道、技术、人才资源"],
  ["行业机会发现", "洞察趋势、发现合作与增长机会"]
] as const;

function CommunityMembersPage() {
  const [config, setConfig] = useState<CommunityConfig | null>(null);
  const [status, setStatus] = useState("");
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

  async function submitJoinRequest() {
    setError("");
    try {
      const user = authSession.get().user;
      await contentApi.joinCommunity({
        community: "members",
        contact: user?.phone || user?.account || user?.nickname || "unknown",
        note: "希望加入创业成长互助社区"
      });
      setStatus("申请已提交");
    } catch (error) {
      setError(apiErrorMessage(error, "申请提交失败，请稍后再试"));
    }
  }

  return (
    <V4PageShell>
      <section className="community-page community-members-page" aria-label="会员社群">
        <main className="community-main community-members-main">
          <header className="community-members-head">
            <h1>AI社群</h1>
            <p>{config?.description ?? "连接高质量创业者与企业用户，分享经验、对接资源、共同成长"}</p>
            {error && <small className="form-error" role="alert">{error}</small>}
          </header>

          <section className="community-entry-grid community-members-entry" aria-label="社群入口">
            <article className="community-entry-card community-member-card free">
              <span>会员社群</span>
              <h2>创业成长互助社区</h2>
              <p>与创业者一起学习、交流、成长</p>
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
              <footer>成员数据待接入</footer>
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
              <footer>企业数据待接入</footer>
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
              <p>暂无社群动态</p>
              <Link className="community-panel-link" to="/community/members">查看全部动态 ›</Link>
            </article>

            <article className="community-panel community-events community-member-events">
              <header>
                <h2>本周活动预告</h2>
                <Link to="/community/members">查看全部 ›</Link>
              </header>
              <p>暂无活动数据</p>
            </article>

            <article className="community-panel community-value community-member-stats">
              <h2>社群价值数据</h2>
              <p>暂无社群价值数据</p>
            </article>
          </section>

          <section className="community-join-modal" role="dialog" aria-label="加入会员社群" aria-modal="true">
            <button aria-label="关闭加入会员社群弹窗" type="button">×</button>
            <h2>加入会员社群</h2>
            <p>{config?.headline ?? "提交申请后由社群助手联系入群"}</p>
            <div className="community-qr" aria-label="会员社群二维码">
              <i aria-hidden="true" />
            </div>
            <small>使用微信扫一扫，添加社群小助手，拉你入群</small>
            {status && <strong>{status}</strong>}
            <button onClick={submitJoinRequest} type="button">提交会员社群申请</button>
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
            <p>暂无社群助手对话</p>
          </div>

          <section className="community-report-card">
            <p>暂无社群报告</p>
          </section>

          <nav className="learning-copilot-actions" aria-label="社群助手快捷入口">
            <Link to="/analysis">分析项目机会 <span aria-hidden="true">›</span></Link>
            <Link to="/tools/recommend">推荐工具 <span aria-hidden="true">›</span></Link>
            <Link to="/learning/plan">制定落地计划 <span aria-hidden="true">›</span></Link>
          </nav>

          <MiniCopilotForm className="learning-copilot-input" placeholder="向我提问，或输入 @ 调用技能" sendIcon="›" />
        </aside>
      </section>
    </V4PageShell>
  );
}

export default CommunityMembersPage;
