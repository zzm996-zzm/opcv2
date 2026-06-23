import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { useAuthSession } from "../lib/authSession";

const profileNav = [
  ["个人中心", "/profile"],
  ["账号与资料设置", "/profile/settings"],
  ["会员与账单", "/membership"],
  ["我的内容", "/profile/content"],
  ["偏好设置", "/profile/preferences"]
] as const;

const quotas = [
  ["AI智算额度", "8,320", "20,000", 42],
  ["数据获取额度", "120", "500", 24],
  ["工具使用额度", "35", "100", 35],
  ["商业沙盘推演", "3", "10", 30],
  ["竞品全盘数据破解", "1", "5", 20]
] as const;

const quickLinks = [
  ["账号与资料设置", "管理账号安全、修改资料与登录方式", "/profile/settings"],
  ["会员与账单", "查看套餐权益、账单明细与开票记录", "/membership"],
  ["我的内容", "查看收藏内容、历史记录与生成成果", "/profile/content"],
  ["偏好设置", "自定义界面、通知与行为偏好设置", "/profile/preferences"]
] as const;

const profileForms = [
  ["姓名", "张晨", "企业管理员"],
  ["公司名称", "智活科技有限公司", "已认证"],
  ["所属行业", "人工智能 / SaaS", "可编辑"],
  ["业务阶段", "增长期", "可编辑"],
  ["核心产品", "AI 智能客服系统", "已完善"],
  ["关注方向", "线索开发、竞品监测、增长测算", "已完善"]
] as const;

const accountBindings = {
  bound: [
    ["联系手机", "138****5678", "已填写", "更换手机"],
    ["微信", "zhihuo_ai", "已绑定", "解绑"]
  ],
  unbound: [
    ["联系手机", "未填写", "可选联系方式", "去填写"],
    ["微信", "未绑定", "暂未同步微信消息", "去绑定"]
  ]
} as const;

type ProfilePageProps = {
  mode?: "overview" | "settings" | "content" | "preferences";
  binding?: "bound" | "unbound";
  overlay?: "password" | "logout" | "delete" | "complete";
};

function ProfilePage({ binding = "bound", mode = "overview", overlay }: ProfilePageProps) {
  const session = useAuthSession();
  const nickname = session.user?.nickname || "张婧";

  return (
    <V4PageShell>
      <section className="profile-page" aria-label={profileTitle(mode)}>
        <div className="page-title-row">
          <div>
            <h1>{profileTitle(mode)}</h1>
            <p>{profileSubtitle(mode)}</p>
          </div>
        </div>

        <div className="profile-layout">
          <aside className="profile-side-tabs" aria-label="个人中心导航">
            {profileNav.map(([label, href]) => (
              <Link key={label} className={activeClass(mode, label)} to={href}>
                {label}
                <span aria-hidden="true">›</span>
              </Link>
            ))}
          </aside>

          <div className="profile-content">
            {mode === "settings" && <AccountSettings binding={binding} />}
            {mode === "content" && <MyContent />}
            {mode === "preferences" && <Preferences />}
            {mode === "overview" && <ProfileOverview nickname={nickname} />}
          </div>
        </div>
        {overlay && <AccountOverlay kind={overlay} />}
      </section>
    </V4PageShell>
  );
}

function ProfileOverview({ nickname }: { nickname: string }) {
  return (
    <>
      <section className="profile-hero-card">
        <div className="profile-avatar-photo" aria-hidden="true">张</div>
        <div className="profile-identity">
          <div>
            <h2>{nickname}</h2>
            <span>企业管理员</span>
          </div>
        </div>
        <div className="profile-facts">
          <article>
            <small>所在组织</small>
            <strong>智活AI</strong>
          </article>
          <article>
            <small>会员身份</small>
            <strong><span aria-hidden="true">♕</span> 企业版</strong>
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
          <p>所有额度均按自然月重置，本月重置日：2025-06-01</p>
        </div>
        <div className="quota-grid">
          {quotas.map(([label, used, total, percent]) => (
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
            ["生成智能客服系统机会分析", "今天 10:18"],
            ["查看竞品动态监测报告", "昨天 18:10"],
            ["更新 AI 智能硬件项目任务", "06-23 14:00"]
          ].map(([title, time]) => (
            <article key={title}>
              <span aria-hidden="true" />
              <strong>{title}</strong>
              <time>{time}</time>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}

function AccountSettings({ binding }: { binding: NonNullable<ProfilePageProps["binding"]> }) {
  return (
    <>
      <section className="profile-completion-card">
        <div className="completion-copy">
          <span>资料完成度</span>
          <strong>78%</strong>
          <small>已完善 11 项 / 共 14 项，补齐后可获得更精准的项目推荐与增长建议</small>
        </div>
        <div className="completion-ring" aria-hidden="true">
          <i>78%</i>
        </div>
        <Link to="/profile/settings/complete">完善资料</Link>
      </section>

      <section className="settings-form-card" aria-label="用户资料">
        <div className="settings-card-head">
          <div>
            <h2>用户资料</h2>
            <p>这些信息会用于生成更贴合你业务场景的 AI 建议</p>
          </div>
          <Link to="/profile/settings/complete">编辑资料</Link>
        </div>
        <div className="settings-form-grid">
          {profileForms.map(([label, value, note]) => (
            <label key={label}>
              <span>{label}</span>
              <strong>{value}</strong>
              <small>{note}</small>
            </label>
          ))}
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
          {accountBindings[binding].map(([type, value, state, action]) => (
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
            <p>最近登录：今天 09:12 · 上海。退出或注销前请确认数据已备份</p>
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

function AccountOverlay({ kind }: { kind: NonNullable<ProfilePageProps["overlay"]> }) {
  if (kind === "password") return <PasswordModal />;
  if (kind === "logout") return <LogoutModal />;
  if (kind === "delete") return <DeleteAccountModal />;
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

function DeleteAccountModal() {
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
          <input type="checkbox" />
          我已阅读并同意注销须知
        </label>
        <footer>
          <Link to="/profile/settings">取消</Link>
          <button className="danger" type="button">确认注销</button>
        </footer>
      </section>
    </div>
  );
}

function CompleteProfileModal() {
  const leftFields = [
    ["联系手机", "138 **** 5678"],
    ["微信 / 企业微信", "zhihuo_ai"],
    ["公司名称", "智活AI科技有限公司"],
    ["所在行业", "人工智能"],
    ["公司规模", "51-200 人"],
    ["主营产品 / 服务", "智活AI企业增长平台"],
    ["产品阶段", "成长期"],
    ["核心客户群", "中大型企业"]
  ] as const;

  return (
    <div className="ui-modal-scrim" role="dialog" aria-label="完善资料">
      <section className="account-modal complete-profile-modal">
        <Link className="modal-close" to="/profile/settings" aria-label="关闭完善资料">×</Link>
        <header>
          <h2>完善资料</h2>
          <div>
            <span className="mini-completion-ring">78%</span>
            <strong>资料完成度 <b>78%</b></strong>
            <small>已完成 <b>11 / 14</b> 项</small>
          </div>
          <p>请完善以下信息，帮助我们为您提供更精准的服务与推荐</p>
        </header>
        <div className="complete-profile-grid">
          <div className="complete-field-list">
            {leftFields.map(([label, value]) => (
              <label key={label}>
                <span>{label} <b>*</b></span>
                <input defaultValue={value} />
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

const contentRows = [
  ["智能客服系统项目匹配", "智能客服SaaS项目的智能匹配分析", "匹配条件：智能客服 | SaaS | 中小企业", "已完成", "今天 10:15"],
  ["AI教育平台方向匹配", "AI教育平台项目方向匹配", "匹配条件：教育科技 | 在线教育 | AI工具", "已完成", "昨天 16:30"],
  ["跨境电商工具匹配", "跨境电商运营工具项目匹配", "匹配条件：跨境电商 | 工具类 | 运营辅助", "进行中", "06-24 14:20"],
  ["本地生活服务匹配", "本地生活服务平台匹配分析", "匹配条件：本地生活 | 服务平台 | O2O", "已完成", "06-23 11:05"],
  ["项目方向初步筛选", "基于当前市场趋势的项目初筛", "匹配条件：综合 | 趋势分析 | 初步筛选", "进行中", "06-22 09:42"]
] as const;

function MyContent() {
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
        {contentRows.map(([title, desc, meta, status, time]) => (
          <article key={title}>
            <span className="content-icon" aria-hidden="true" />
            <div>
              <h2>{title}</h2>
              <p>{desc}</p>
              <small>{meta}</small>
            </div>
            <b className={status === "进行中" ? "running" : ""}>{status}</b>
            <time>{time}</time>
            <button aria-label={`收藏${title}`} type="button">☆</button>
            <a href="/projects">查看详情</a>
            <span aria-hidden="true">›</span>
          </article>
        ))}
      </div>
      <p className="content-loaded">已加载全部内容</p>
    </section>
  );
}

function Preferences() {
  return (
    <>
      <section className="preference-card">
        <h2>通知设置</h2>
        <p>选择接收通知的方式及内容</p>
        <div className="notification-matrix">
          <div className="matrix-head">
            {["站内通知", "邮件通知", "企业微信通知"].map((item) => (
              <span key={item}>{item}<i className="toggle-switch on" aria-hidden="true" /></span>
            ))}
          </div>
          {[
            ["任务提醒", "任务创建、分配、截止时间等提醒", true, true, true],
            ["系统通知", "系统更新、功能上线等重要通知", true, true, false],
            ["营销消息", "产品动态、活动信息等推广内容", true, false, false]
          ].map(([title, desc, inApp, email, wechat]) => (
            <article key={title as string}>
              <span className="preference-row-icon" aria-hidden="true" />
              <div>
                <strong>{title}</strong>
                <small>{desc}</small>
              </div>
              {[inApp, email, wechat].map((enabled, index) => (
                <i className={`toggle-switch ${enabled ? "on" : ""}`} key={index} aria-hidden="true" />
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
          {["Claude opus4.8", "Chatgpt 5.5", "Gork4.3"].map((model, index) => (
            <button className={index === 0 ? "active" : ""} key={model} type="button">
              <span className={`model-mark mark-${index}`} aria-hidden="true" />
              {model}
            </button>
          ))}
        </div>
        <small className="preference-note">部分模型的可用性可能因你的身份或企业权限而有所不同</small>
      </section>

      <section className="preference-card split">
        <label>
          <strong>语言设置</strong>
          <small>选择你的界面显示语言</small>
          <button type="button">简体中文 <span aria-hidden="true">⌄</span></button>
        </label>
        <label>
          <strong>主题设置</strong>
          <small>选择界面主题风格</small>
          <button type="button">浅色（跟随系统） <span aria-hidden="true">⌄</span></button>
        </label>
      </section>

      <section className="preference-save-card">
        <button type="button">保存设置</button>
        <span>更改将自动保存</span>
      </section>
    </>
  );
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

function activeClass(mode: ProfilePageProps["mode"], label: string) {
  if (mode === "settings" && label === "账号与资料设置") return "active";
  if (mode === "content" && label === "我的内容") return "active";
  if (mode === "preferences" && label === "偏好设置") return "active";
  if (mode === "overview" && label === "个人中心") return "active";
  return "";
}

export default ProfilePage;
