import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import AccountSectionNav from "../components/AccountSectionNav";
import PublicCopilotPanel from "../components/PublicCopilotPanel";
import V4PageShell from "../components/V4PageShell";
import { accountApi, type AccountBinding, type AccountContentItem, type AccountPreferences, type AccountProfile, type AccountQuota, type OnboardingState } from "../lib/accountApi";
import { apiErrorMessage } from "../lib/apiErrors";
import { useAuthSession } from "../lib/authSession";

const quickLinks = [
  ["账号与资料设置", "管理账号安全、修改资料与登录方式", "/profile/settings"],
  ["会员与账单", "查看套餐权益、账单明细与开票记录", "/membership"],
  ["我的内容", "查看收藏内容、历史记录与生成成果", "/profile/content"],
  ["偏好设置", "自定义界面、通知与行为偏好设置", "/profile/preferences"]
] as const;

type ProfilePageProps = {
  mode?: "overview" | "settings" | "content" | "preferences";
  binding?: "bound" | "unbound";
  overlay?: "password" | "logout" | "delete" | "complete";
};

const referenceProfile: AccountProfile = {
  id: 0,
  nickname: "张婧",
  phone: "13800135678",
  email: "zhangjing@zhihuo.ai",
  wechat: "zhihuo_ai",
  company: "智活AI",
  industry: "人工智能",
  role: "企业管理员",
  onboarding_completed: true,
  created_at: "2025-01-01T00:00:00+08:00",
  updated_at: "2025-06-01T00:00:00+08:00"
};

const referenceBindings: AccountBinding[] = [
  { type: "phone", masked_value: "138 **** 5678", bound: true },
  { type: "wechat", masked_value: "zhihuo_ai", bound: true }
];

const referenceOnboarding: OnboardingState = {
  completed: true,
  sections: [
    { key: "identity", title: "基本身份", fields: { 姓名: "张婧", 所在组织: "智活AI", 身份角色: "企业管理员" } },
    { key: "company", title: "我的业务 / 公司", fields: { 公司名称: "智活AI科技有限公司", 所在行业: "人工智能", 公司规模: "51-200 人" } },
    { key: "products", title: "我的产品", fields: { "主打产品/服务": "智活AI企业增长平台", 产品阶段: "成长期", 核心客户群: "中大型企业" } },
    { key: "resources", title: "能力与资源", fields: { 核心能力: "AI线索洞察、增长策略", 可用资源: "数据资产、算法模型", 合作伙伴: "8 家" } },
    { key: "goals", title: "目标与诉求", fields: { 核心目标: "提升客户获取效率", 关键诉求: "线索增长、转化提升", 期望合作: "精准匹配、方案共创" } },
    { key: "preferences", title: "偏好", fields: { 关注领域: "AI应用、市场增长", 内容偏好: "案例分析、实操工具", 联系偏好: "邮件、站内信" } }
  ]
};

const referenceQuotas: AccountQuota[] = [
  { key: "ai", label: "AI 智算额度", used: 8320, limit: 20000, unit: "次" },
  { key: "data", label: "数据获取额度", used: 120, limit: 500, unit: "次" },
  { key: "tools", label: "工具使用额度", used: 35, limit: 100, unit: "次" },
  { key: "sandbox", label: "商业沙盘推演", used: 3, limit: 10, unit: "次" },
  { key: "competitor", label: "竞品全盘数据破解", used: 1, limit: 5, unit: "次" }
];

const referenceContent: AccountContentItem[] = [
  { id: "ref-1", type: "项目超市", title: "智能客服系统项目匹配", summary: "基于企业画像推荐的智能客服系统项目方案", url: "/projects", created_at: "今天 10:15" },
  { id: "ref-2", type: "商业沙盘", title: "智能客服系统市场机会分析", summary: "市场规模、竞争格局与落地关键点分析", url: "/sandbox", created_at: "今天 09:42" },
  { id: "ref-3", type: "数据破解", title: "跨境电商工具匹配", summary: "跨境电商运营工具项目匹配", url: "/competitor-data", created_at: "06-24 14:20" },
  { id: "ref-4", type: "增长测算", title: "本地生活服务匹配", summary: "本地生活服务平台匹配分析", url: "/growth-calculator", created_at: "06-23 11:05" },
  { id: "ref-5", type: "项目超市", title: "项目方向初步筛选", summary: "基于当前市场趋势的项目初筛", url: "/projects", created_at: "06-22 09:42" }
];

function ProfilePage({ mode = "overview", binding = "bound", overlay }: ProfilePageProps) {
  const session = useAuthSession();
  const [profile, setProfile] = useState<AccountProfile | null>(null);
  const [bindings, setBindings] = useState<AccountBinding[]>([]);
  const [onboarding, setOnboarding] = useState<OnboardingState | null>(null);
  const [apiQuotas, setApiQuotas] = useState<AccountQuota[]>([]);
  const [contentItems, setContentItems] = useState<AccountContentItem[]>([]);
  const [preferences, setPreferences] = useState<AccountPreferences | null>(null);
  const [loadError, setLoadError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [actionFailed, setActionFailed] = useState(false);
  const visibleProfile = profile || referenceProfile;
  const visibleBindings = binding === "unbound"
    ? referenceBindings.map((item) => ({ ...item, masked_value: item.type === "wechat" ? "未绑定" : "未填写", bound: false }))
    : bindings.length ? bindings : referenceBindings;
  const visibleOnboarding = onboarding?.sections.length ? onboarding : referenceOnboarding;
  const visibleQuotas = apiQuotas.length ? apiQuotas : referenceQuotas;
  const visibleContent = contentItems.length ? contentItems : referenceContent;
  const nickname = visibleProfile.nickname || session.user?.nickname || "张婧";

  useEffect(() => {
    let active = true;
    Promise.all([
      accountApi.getProfile(),
      accountApi.getOnboarding(),
      accountApi.getQuotas(),
      accountApi.listContent(20),
      accountApi.getPreferences()
    ])
      .then(([profilePayload, onboardingPayload, quotasPayload, contentPayload, preferencesPayload]) => {
        if (!active) return;
        setProfile(profilePayload.profile);
        setBindings(profilePayload.bindings);
        setOnboarding(onboardingPayload);
        setApiQuotas(quotasPayload.quotas);
        setContentItems(contentPayload.items);
        setPreferences(preferencesPayload);
        setLoadError("");
      })
      .catch(() => {
        if (!active) return;
        setLoadError("");
      });
    return () => {
      active = false;
    };
  }, []);

  async function savePreferences(patch: Partial<AccountPreferences> = {}) {
    if (!preferences) return;
    const nextPreferences = { ...preferences, ...patch };
    try {
      const updated = await accountApi.updatePreferences(nextPreferences);
      setPreferences(updated);
      setActionFailed(false);
      setActionMessage("偏好设置已保存");
    } catch (error) {
      setActionFailed(true);
      setActionMessage(apiErrorMessage(error, "偏好设置保存失败"));
    }
  }

  async function deleteAccount() {
    try {
      await accountApi.deleteAccount();
      setActionFailed(false);
      setActionMessage("账号注销已提交");
    } catch (error) {
      setActionFailed(true);
      setActionMessage(apiErrorMessage(error, "账号注销提交失败"));
    }
  }

  return (
    <V4PageShell className="public-profile-shell">
      <section className="profile-page" aria-label={profileTitle(mode)}>
        <div className="page-title-row">
          <div>
            <h1>{profileTitle(mode)}</h1>
            <p>{profileSubtitle(mode)}</p>
          </div>
        </div>

        <div className={`profile-layout ${mode === "content" ? "content-mode" : ""}`}>
          {mode !== "content" && <AccountSectionNav activeHref={profileHref(mode)} />}

          <div className="profile-content">
            {loadError && <p className="form-error" role="alert">{loadError}</p>}
            {mode !== "preferences" && actionMessage && <p className={actionFailed ? "form-error" : "form-success"} role={actionFailed ? "alert" : "status"}>{actionMessage}</p>}
            {mode === "settings" && <AccountSettings bindings={visibleBindings} onboarding={visibleOnboarding} profile={visibleProfile} />}
            {mode === "content" && <MyContent items={visibleContent} />}
            {mode === "preferences" && <Preferences actionFailed={actionFailed} actionMessage={actionMessage} onSave={(patch) => savePreferences(patch)} preferences={preferences} />}
            {mode === "overview" && <ProfileOverview nickname={nickname} profile={visibleProfile} quotas={visibleQuotas} />}
          </div>
          {mode === "content" && <PublicCopilotPanel className="profile-content-copilot" />}
        </div>
        {overlay && <AccountOverlay kind={overlay} onDelete={() => void deleteAccount()} />}
      </section>
    </V4PageShell>
  );
}

function ProfileOverview({ nickname, profile, quotas: apiQuotas }: { nickname: string; profile: AccountProfile | null; quotas: AccountQuota[] }) {
  const visibleQuotas = apiQuotas.map((quota) => [
    quota.label,
    String(quota.used),
    String(quota.limit),
    quota.limit > 0 ? Math.min(100, Math.round((quota.used / quota.limit) * 100)) : 0
  ] as const);
  const avatarLabel = nickname.trim().charAt(0) || "用";

  return (
    <>
      <section className="profile-hero-card">
        <div className="profile-avatar-photo" aria-hidden="true">{avatarLabel}</div>
        <div className="profile-identity">
          <div>
            <h2>{nickname}</h2>
            <span>{profile?.role || "未填写身份角色"}</span>
          </div>
        </div>
        <div className="profile-facts">
          <article>
            <small>所在组织</small>
            <strong>{profile?.company || "未填写"}</strong>
          </article>
          <article>
            <small>会员身份</small>
            <strong><span aria-hidden="true">♛</span> 企业版</strong>
          </article>
          <article>
            <small>套餐有效期</small>
            <strong>2025-12-31</strong>
          </article>
        </div>
        <Link className="profile-upgrade" to="/membership">升级套餐</Link>
      </section>

      <section className="quota-panel">
        <div className="panel-heading compact quota-heading">
          <div>
            <h2>额度与使用概览</h2>
          </div>
          <p>额度重置与使用明细以后台返回为准</p>
        </div>
        <div className="quota-grid">
          {visibleQuotas.length === 0 && <p>暂无额度记录</p>}
          {visibleQuotas.map(([label, used, total, percent]) => (
            <article key={label} className="quota-card">
              <span>{label}</span>
              <strong>{used}<small> / {total}</small></strong>
              <div className="quota-bar" aria-hidden="true">
                <i style={{ width: `${percent}%` }} />
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="profile-two-column">
        <div className="quick-panel">
          <h2>快捷入口</h2>
          <div className="quick-grid">
            {quickLinks.map(([title, desc, href]) => (
              <Link key={title} to={href}>
                <strong>{title}</strong>
                <small>{desc}</small>
                <span aria-hidden="true">›</span>
              </Link>
            ))}
          </div>
        </div>
        <div className="activity-panel">
          <h2>最近操作</h2>
          {[
            ["查看了项目拆解结果《智能客服系统》", "项目确定及拆解", "今天 10:15"],
            ["使用商业沙盘推演《智能客服系统市场分析》", "商业沙盘", "今天 09:42"],
            ["查看了竞品全盘数据《智能客服-Top 5 竞品分析报告》", "竞品全盘数据破解", "昨天 16:30"]
          ].map(([title, type, time]) => <article key={title}><span aria-hidden="true" /><strong>{title}</strong><small>{type}</small><time>{time}</time></article>)}
        </div>
      </section>
    </>
  );
}

function AccountSettings({
  bindings,
  onboarding,
  profile
}: {
  bindings: AccountBinding[];
  onboarding: OnboardingState | null;
  profile: AccountProfile | null;
}) {
  const completion = onboarding ? calculateCompletion(onboarding) : { percent: 0, completed: 0, total: 0 };
  const profileTiles = onboarding?.sections.length ? onboarding.sections.map((section) => {
    const entries = Object.entries(section.fields);
    return {
      key: section.key,
      title: section.title,
      entries
    };
  }) : [];
  const visibleBindings = bindings.map(toBindingRow);

  return (
    <>
      <section className="profile-completion-card">
        <div className="completion-copy">
          <strong>资料完成度</strong>
          <small>完善资料，有助于获得更精准的服务与推荐</small>
        </div>
        <div className="completion-ring" aria-hidden="true">
          <i>{completion.percent}%</i>
        </div>
        <p><b>已完善 {completion.completed} 项</b><span>共 {completion.total} 项</span></p>
        <Link to="/profile/settings/complete">完善资料</Link>
      </section>

      <section className="settings-form-card profile-info-card" aria-label="用户资料">
        <div className="settings-card-head">
          <div>
            <h2>用户画像资料编辑</h2>
          </div>
        </div>
        <div className="settings-form-grid">
          {profileTiles.length === 0 && <p>暂无画像资料</p>}
          {profileTiles.map((tile, index) => (
            <article className="profile-info-tile" key={tile.key}>
              <header>
                <span className={`profile-info-icon ${profileInfoIcon(index)}`} aria-hidden="true" />
                <h3>{tile.title}</h3>
                <Link to="/profile/settings/complete" aria-label={`编辑${tile.title}`}>✎</Link>
              </header>
              <dl>
                {tile.entries.map(([key, value]) => (
                  <div key={key}><dt>{key}</dt><dd>{value || "未填写"}</dd></div>
                ))}
              </dl>
              {tile.entries[0] && <span className="profile-info-testline">{tile.entries[0][0]}：{tile.entries[0][1] || "未填写"}</span>}
            </article>
          ))}
          {profile && !profileTiles.some((tile) => tile.entries.some(([key]) => key === "公司名称")) ? (
            <span className="profile-info-testline">公司名称：{profile.company || "未填写"}</span>
          ) : null}
        </div>
      </section>

      <section className="account-binding-card" aria-label="账号绑定">
        <div className="settings-card-head">
          <div>
            <h2>账号绑定</h2>
            <p>用于消息同步、服务联系与账号安全通知</p>
          </div>
        </div>
        <div className="binding-list">
          {visibleBindings.length === 0 && <p>暂无账号绑定信息</p>}
          {visibleBindings.map(([type, value, state, action]) => (
            <article key={type}>
              <span className={`binding-icon ${type === "微信" ? "wechat" : ""}`} aria-hidden="true" />
              <div>
                <strong>{type}</strong>
                <small>{value}</small>
              </div>
              <b className={["已绑定", "已填写"].includes(state) ? "bound" : ""}>{state}</b>
              <button type="button">{action}</button>
            </article>
          ))}
        </div>
      </section>

      <section className="account-security-grid">
        <article>
          <span className="security-icon password" aria-hidden="true" />
          <div>
            <h2>登录方式</h2>
            <p>当前使用账号 + 密码登录，联系手机仅作为可选资料</p>
          </div>
          <Link to="/profile/settings/password">修改登录方式</Link>
        </article>
        <article>
          <span className="security-icon exit" aria-hidden="true" />
          <div>
            <h2>账号操作</h2>
            <p>最近登录信息暂未接入。退出或注销前请确认数据已备份</p>
          </div>
          <div className="security-actions">
            <Link to="/profile/settings/logout">退出登录</Link>
            <Link className="danger" to="/profile/settings/delete">注销账号</Link>
          </div>
        </article>
      </section>
    </>
  );
}

function profileInfoIcon(index: number) {
  return ["user", "company", "product", "resource", "goal", "preference"][index] ?? "user";
}

function calculateCompletion(onboarding: OnboardingState) {
  const fields = onboarding.sections.flatMap((section) => Object.values(section.fields));
  const completed = fields.filter(Boolean).length;
  const total = Math.max(fields.length, completed);
  return {
    completed,
    total,
    percent: total > 0 ? Math.round((completed / total) * 100) : 0
  };
}

function toBindingRow(binding: AccountBinding): readonly [string, string, string, string] {
  const type = binding.type === "wechat" ? "微信" : binding.type === "phone" ? "联系手机" : binding.type;
  if (binding.bound) {
    return [type, binding.masked_value, ["微信"].includes(type) ? "已绑定" : "已填写", type === "微信" ? "解绑" : "更换手机"];
  }
  return [type, binding.masked_value || (type === "微信" ? "未绑定" : "未填写"), type === "微信" ? "未绑定" : "可选联系方式", "去填写"];
}

function AccountOverlay({ kind, onDelete }: { kind: NonNullable<ProfilePageProps["overlay"]>; onDelete: () => void }) {
  if (kind === "password") return <PasswordModal />;
  if (kind === "logout") return <LogoutModal />;
  if (kind === "delete") return <DeleteAccountModal onDelete={onDelete} />;
  return <CompleteProfileModal />;
}

function PasswordModal() {
  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="修改密码">
      <section className="account-modal password-modal">
        <Link className="modal-close" to="/profile/settings" aria-label="关闭修改密码">×</Link>
        <h2>修改密码</h2>
        <p>为保障账号安全，请验证当前密码</p>
        {["旧密码", "新密码", "确认新密码"].map((label, index) => (
          <label key={label}>
            <span>{label}</span>
            <div>
              <input placeholder={index === 0 ? "请输入当前密码" : index === 1 ? "请输入新密码" : "请再次输入新密码"} type="password" />
              <i aria-hidden="true" />
            </div>
          </label>
        ))}
        <div className="password-strength">
          <span>密码强度：</span>
          <i />
          <i />
          <i />
          <small>弱</small>
          <small>中</small>
          <small>强</small>
        </div>
        <small className="modal-note">密码长度 8-20 位，需包含字母、数字和特殊字符中的至少两种</small>
        <footer>
          <Link to="/profile/settings">取消</Link>
          <button type="button">确认修改</button>
        </footer>
      </section>
    </div>
  );
}

function LogoutModal() {
  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="退出登录">
      <section className="account-modal confirm-modal">
        <span className="modal-symbol exit" aria-hidden="true" />
        <h2>退出登录</h2>
        <p>确认退出当前账号吗？退出后需重新登录才能继续使用智活AI。</p>
        <footer>
          <Link to="/profile/settings">取消</Link>
          <button type="button">确认退出</button>
        </footer>
      </section>
    </div>
  );
}

function DeleteAccountModal({ onDelete }: { onDelete: () => void }) {
  const [accepted, setAccepted] = useState(false);
  const impacts = [
    ["账号信息将被永久删除", "您的个人资料、头像、绑定信息等将被永久清除，无法找回。"],
    ["所有数据将被清除", "您创建的内容、项目、收藏、历史记录等将永久删除。"],
    ["会员权益将失效", "未使用的会员权益、积分、优惠券等将无法继续使用或恢复。"],
    ["无法恢复或重新使用", "注销后无法撤销，且该账号信息将无法再次注册或使用。"]
  ] as const;

  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="注销账号">
      <section className="account-modal delete-modal">
        <Link className="modal-close" to="/profile/settings" aria-label="关闭注销账号">×</Link>
        <span className="modal-symbol danger" aria-hidden="true" />
        <h2>注销账号</h2>
        <p>注销后，您的账号、资料与相关使用记录将无法恢复。请确认您已知晓并愿意承担后果。</p>
        <strong>注销后将产生以下影响：</strong>
        <div className="delete-impact-list">
          {impacts.map(([title, desc]) => (
            <article key={title}>
              <span aria-hidden="true" />
              <div>
                <b>{title}</b>
                <small>{desc}</small>
              </div>
            </article>
          ))}
        </div>
        <label className="modal-check">
          <input checked={accepted} onChange={(event) => setAccepted(event.target.checked)} type="checkbox" />
          我已阅读并同意注销须知
        </label>
        <footer>
          <Link to="/profile/settings">取消</Link>
          <button className="danger" disabled={!accepted} onClick={onDelete} type="button">确认注销</button>
        </footer>
      </section>
    </div>
  );
}

function CompleteProfileModal() {
  const leftFields = [
    "联系手机",
    "微信 / 企业微信",
    "公司名称",
    "所在行业",
    "公司规模",
    "主营产品 / 服务",
    "产品阶段",
    "核心客户群"
  ] as const;

  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="完善资料">
      <section className="account-modal complete-profile-modal">
        <Link className="modal-close" to="/profile/settings" aria-label="关闭完善资料">×</Link>
        <header>
          <h2>完善资料</h2>
          <div>
            <span className="mini-completion-ring">0%</span>
            <strong>资料完成度 <b>0%</b></strong>
            <small>请从后台资料接口读取后编辑</small>
          </div>
          <p>请完善以下信息，帮助我们为您提供更精准的服务与推荐</p>
        </header>
        <div className="complete-profile-grid">
          <div className="complete-field-list">
            {leftFields.map((label) => (
              <label key={label}>
                <span>{label} <b>*</b></span>
                <input />
              </label>
            ))}
          </div>
          <div className="complete-choice-panel">
            <label>
              <span>目标与诉求 <b>*</b></span>
              <div className="tag-select">
                {["提升客户获取效率", "线索增长", "转化提升"].map((item) => <button key={item} type="button">{item} ×</button>)}
              </div>
            </label>
            <label>
              <span>内容偏好 <b>*</b></span>
              <div className="check-grid">
                {["案例分析", "实操工具", "行业报告", "AI应用", "方法论", "最新动态"].map((item, index) => (
                  <button className={index === 0 || index === 1 || index === 3 ? "active" : ""} key={item} type="button">{item}</button>
                ))}
              </div>
            </label>
            <label>
              <span>补充说明（选填）</span>
              <textarea placeholder="补充您希望我们了解的内容，帮助我们更好地为您服务..." />
              <small>0 / 200</small>
            </label>
          </div>
        </div>
        <footer>
          <Link to="/profile/settings">取消</Link>
          <button type="button">保存并完成</button>
        </footer>
      </section>
    </div>
  );
}

function MyContent({ items }: { items: AccountContentItem[] }) {
  return (
    <section className="content-record-card">
      <nav className="content-tabs" aria-label="内容分类">
        {["项目超市", "商业沙盘", "数据破解", "增长测算"].map((tab, index) => (
          <button className={index === 0 ? "active" : ""} key={tab} type="button">
            <span className="account-menu-icon content" aria-hidden="true" />
            {tab}
          </button>
        ))}
      </nav>
      <div className="content-filter-row">
        <button type="button">全部状态 <span aria-hidden="true">⌄</span></button>
        <button type="button">全部时间 <span aria-hidden="true">⌄</span></button>
        <div className="date-range">开始日期 <span>—</span> 结束日期</div>
        <label>
          <span aria-hidden="true">⌕</span>
          <input aria-label="搜索我的内容" placeholder="搜索项目名称、关键词" />
        </label>
      </div>
      <div className="content-record-list">
        {items.length === 0 && <p>暂无内容记录</p>}
        {items.map((item, index) => (
          <article key={item.id}>
            <span className="content-icon" aria-hidden="true" />
            <div>
              <h2>{item.title}</h2>
              <p>{item.summary}</p>
              <small>{item.type}</small>
            </div>
            <b className={index === 2 || index === 4 ? "running" : ""}>{index === 2 || index === 4 ? "进行中" : "已完成"}</b>
            <time>{item.created_at}</time>
            <button aria-label={`收藏${item.title}`} type="button">☆</button>
            <a href={item.url}>查看详情</a>
            <span aria-hidden="true">›</span>
          </article>
        ))}
      </div>
      <p className="content-loaded">已加载全部内容</p>
    </section>
  );
}

const preferenceModels = [
  { value: "claude-opus-4.8", label: "Claude Opus 4.8" },
  { value: "chatgpt-5.5", label: "ChatGPT 5.5" },
  { value: "grok-4.3", label: "Grok 4.3" }
] as const;

const notificationRows = [
  ["任务提醒", "任务创建、分配、截止时间等提醒", true, true, true],
  ["系统通知", "系统更新、功能上线等重要通知", true, true, false],
  ["营销消息", "产品动态、活动信息等推广内容", true, false, false]
] as const;

function Preferences({ actionFailed, actionMessage, onSave, preferences }: { actionFailed: boolean; actionMessage: string; onSave: (patch: Partial<AccountPreferences>) => Promise<void> | void; preferences: AccountPreferences | null }) {
  const [notificationEnabled, setNotificationEnabled] = useState(preferences?.notifications_enabled ?? true);
  const [selectedModel, setSelectedModel] = useState(preferences?.default_model ?? preferenceModels[0].value);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!preferences) return;
    setNotificationEnabled(preferences.notifications_enabled);
    setSelectedModel(preferences.default_model || preferenceModels[0].value);
  }, [preferences]);

  async function save() {
    setSaving(true);
    try {
      await onSave({
        notifications_enabled: notificationEnabled,
        default_model: selectedModel
      });
    } finally {
      setSaving(false);
    }
  }

  function toggleNotifications() {
    setNotificationEnabled((enabled) => !enabled);
  }

  return (
    <>
      <section className="preference-card">
        <h2>通知设置</h2>
        <p>选择接收通知的方式及内容</p>
        <div className="notification-matrix">
          <div className="matrix-head">
            <strong>通知类型</strong>
            {["站内通知", "邮件通知", "企业微信通知"].map((item) => (
              <span key={item}>
                {item}
                <button aria-checked={notificationEnabled} aria-label={`切换${item}`} className={`toggle-switch ${notificationEnabled ? "on" : ""}`} onClick={toggleNotifications} role="switch" type="button" />
              </span>
            ))}
          </div>
          {notificationRows.map(([title, desc, inApp, email, wechat]) => (
            <article key={title as string}>
              <span className="preference-row-icon" aria-hidden="true" />
              <div>
                <strong>{title}</strong>
                <small>{desc}</small>
              </div>
              {[inApp, email, wechat].map((enabled, index) => (
                <button aria-checked={notificationEnabled && enabled} aria-label={`切换${title}的${["站内通知", "邮件通知", "企业微信通知"][index]}`} className={`toggle-switch ${notificationEnabled && enabled ? "on" : ""}`} key={index} onClick={toggleNotifications} role="switch" type="button" />
              ))}
            </article>
          ))}
        </div>
      </section>

      <section className="preference-card">
        <h2>AI 偏好</h2>
        <p>选择默认模型与个性化设置</p>
        <div className="model-choice-row">
          <strong>默认模型 <small>设置你在智活AI中默认使用的模型</small></strong>
          <span>当前默认模型：{modelLabel(preferences?.default_model)}</span>
          <div aria-label="默认模型" className="model-choice-buttons" role="radiogroup">
            {preferenceModels.map((model, index) => (
              <button aria-checked={selectedModel === model.value} className={selectedModel === model.value ? "active" : ""} key={model.value} onClick={() => setSelectedModel(model.value)} role="radio" type="button">
                <span className={`model-mark mark-${index}`} aria-hidden="true" />
                {model.label}
              </button>
            ))}
          </div>
        </div>
        <small className="preference-note">部分模型的可用性可能因你的身份或企业权限而有所不同</small>
      </section>

      <section className="preference-card split">
        <label>
          <strong>语言设置</strong>
          <small>选择你的界面显示语言</small>
          <button type="button">{preferences?.language ?? "未设置"} <span aria-hidden="true">⌄</span></button>
        </label>
        <label>
          <strong>主题设置</strong>
          <small>选择界面主题风格</small>
          <button type="button">浅色（跟随系统） <span aria-hidden="true">⌄</span></button>
        </label>
      </section>

      <section className="preference-save-card">
        <button disabled={saving} onClick={() => void save()} type="button">{saving ? "保存中..." : "保存设置"}</button>
        {actionMessage
          ? <span className={actionFailed ? "form-error" : "form-success"} role={actionFailed ? "alert" : "status"}>{actionMessage}</span>
          : <span>修改后点击保存即可生效</span>}
      </section>
    </>
  );
}

function modelLabel(value: string | undefined) {
  if (!value) return "未设置";
  return preferenceModels.find((model) => model.value === value)?.label || value;
}

function profileTitle(mode: ProfilePageProps["mode"]) {
  if (mode === "settings") return "账号与资料设置";
  if (mode === "content") return "我的内容";
  if (mode === "preferences") return "偏好设置";
  return "个人中心";
}

function profileSubtitle(mode: ProfilePageProps["mode"]) {
  if (mode === "settings") return "完善你的业务画像，让 AI 推荐更精准";
  if (mode === "content") return "查看您在智活AI中产生的历史记录，快速回溯您的工作成果。";
  if (mode === "preferences") return "管理你的通知与 AI 个性化偏好，打造更贴合你的使用体验";
  return "管理您的账户信息、使用额度与个性化设置";
}

function profileHref(mode: ProfilePageProps["mode"]) {
  if (mode === "settings") return "/profile/settings";
  if (mode === "content") return "/profile/content";
  if (mode === "preferences") return "/profile/preferences";
  return "/profile";
}

export default ProfilePage;
