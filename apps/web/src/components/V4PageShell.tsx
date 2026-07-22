import { useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";

import { authApi } from "../lib/authApi";
import { authSession, useAuthSession } from "../lib/authSession";

const topNav = [
  { label: "工作台", href: "/" },
  { label: "智活 Copilot", href: "/copilot", featured: true },
  { label: "工具箱", href: "/tools" },
  { label: "咨询通", href: "/insights" },
  { label: "AI社群", href: "/community" },
  { label: "AI教学", href: "/learning" }
];

const defaultSidebarGroups = [
  {
    title: "项目确定及拆解",
    items: [
      { label: "项目超市", href: "/projects", icon: "grid" },
      { label: "商业沙盘", href: "/sandbox", icon: "home" }
    ]
  },
  {
    title: "落地",
    items: [
      { label: "任务中心", href: "/tasks", icon: "check" },
      { label: "竞品全盘数据破解", href: "/competitor-data", icon: "stack" },
      { label: "竞品动态监测", href: "/competitor-monitoring", icon: "pulse" },
      { label: "增长测算", href: "/growth-calculator", icon: "calc" }
    ]
  },
  {
    title: "增长",
    items: [
      { label: "GEO获客", href: "/geo", icon: "target" },
      { label: "AI线索开发", href: "/leads", icon: "diamond" },
      { label: "仪表盘", href: "/dashboard", icon: "chart" },
      { label: "CRM客户管理", href: "/crm", icon: "user" },
      { label: "企业定制化陪跑", href: "/enterprise", icon: "flag" }
    ]
  }
];

const taskSidebarGroups = [
  {
    title: "项目管理及系统",
    items: [
      { label: "项目列阵", href: "/projects", icon: "grid" },
      { label: "商业沙盘", href: "/sandbox", icon: "home" }
    ]
  },
  {
    title: "落地",
    items: [
      { label: "任务中心", href: "/tasks", icon: "check" },
      { label: "我负责的管理看板", href: "/tasks/board", icon: "stack" },
      { label: "发起的待协调", href: "/tasks/calendar", icon: "pulse" },
      { label: "增长视图", href: "/growth-calculator", icon: "calc" }
    ]
  },
  {
    title: "增长",
    items: [
      { label: "CEO智管", href: "/geo", icon: "target" },
      { label: "AI探索开发", href: "/leads", icon: "diamond" },
      { label: "发现AI", href: "/dashboard", icon: "chart" },
      { label: "CRM客户管理", href: "/crm", icon: "user" },
      { label: "企业定制化应用", href: "/enterprise", icon: "flag" }
    ]
  }
];

const crmSidebarGroups = [
  {
    title: "项目管理",
    items: [
      { label: "项目看板", href: "/projects", icon: "grid" },
      { label: "商业沙盘", href: "/sandbox", icon: "home" }
    ]
  },
  {
    title: "落地执行",
    items: [
      { label: "任务中心", href: "/tasks", icon: "check" },
      { label: "竞品情报", href: "/competitor-data", icon: "stack" },
      { label: "销售线索池", href: "/leads", icon: "pulse" },
      { label: "增长分析", href: "/growth-calculator", icon: "calc" }
    ]
  },
  {
    title: "增长引擎",
    items: [
      { label: "GEO获客", href: "/geo", icon: "target" },
      { label: "AI线索开发", href: "/leads", icon: "diamond" },
      { label: "仪表盘", href: "/dashboard", icon: "chart" },
      { label: "CRM客户管理", href: "/crm", icon: "user" },
      { label: "企业定制化陪跑", href: "/enterprise", icon: "flag" }
    ]
  }
];

const accountLinks: Array<[string, string, string]> = [
  ["个人中心", "/profile", "user"],
  ["账号与资料设置", "/profile/settings", "settings"],
  ["会员与账单", "/membership", "wallet"],
  ["我的内容", "/profile/content", "content"],
  ["偏好设置", "/profile/preferences", "gear"]
];

type V4PageShellProps = {
  children: ReactNode;
  className?: string;
  showCopilotMini?: boolean;
};

function V4PageShell({ children, className = "" }: V4PageShellProps) {
  const session = useAuthSession();
  const location = useLocation();
  const [accountOpen, setAccountOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const nickname = session.user?.nickname || "张婧";
  const sidebarGroups = location.pathname.startsWith("/crm")
    ? crmSidebarGroups
    : location.pathname.startsWith("/tasks") ? taskSidebarGroups : defaultSidebarGroups;
  const isTopNavActive = (href: string) =>
    location.pathname === href || (href !== "/" && location.pathname.startsWith(`${href}/`));

  async function logout() {
    try {
      await authApi.logout();
    } catch {
      // Local logout must still work when the API is temporarily unavailable.
    } finally {
      authSession.clear();
      setAccountOpen(false);
    }
  }

  return (
    <div className={`v4-shell ${className} ${sidebarCollapsed ? "v4-shell-sidebar-collapsed" : ""}`}>
      <aside className="v4-sidebar" aria-label="产品侧边导航">
        <Link className="v4-brand" to="/" aria-label="智活AI OPC V4.0 首页">
          <span className="v4-logo" aria-hidden="true" />
          <span className="v4-brand-name">智活AI</span>
          <small>OPC V4.0</small>
        </Link>

        <nav className="v4-side-nav" aria-label="三大板块导航">
          {sidebarGroups.map((group) => (
            <section key={group.title} className="v4-side-group">
              <button className="v4-group-title" type="button">
                <span>{group.title}</span>
                <span aria-hidden="true">⌄</span>
              </button>
              {group.items.map((item) => (
                <Link
                  key={item.href}
                  className={`v4-side-link ${location.pathname.startsWith(item.href) ? "active" : ""}`}
                  to={item.href}
                >
                  <span className={`v4-line-icon ${item.icon}`} aria-hidden="true" />
                  <span>{item.label}</span>
                </Link>
              ))}
            </section>
          ))}
        </nav>

        <div className="v4-sidebar-bottom">
          <Link className={`v4-side-link with-dot ${location.pathname.startsWith("/messages") ? "active" : ""}`} to="/messages">
            <span className="v4-line-icon chat" aria-hidden="true" />
            <span>消息中心</span>
          </Link>
          <Link className="v4-side-link" to="/help">
            <span className="v4-line-icon help" aria-hidden="true" />
            <span>帮助与反馈</span>
          </Link>
          <button
            aria-label={sidebarCollapsed ? "展开侧栏" : "收起侧栏"}
            className="sidebar-collapse"
            onClick={() => setSidebarCollapsed((collapsed) => !collapsed)}
            type="button"
          >
            <span aria-hidden="true">{sidebarCollapsed ? "›" : "‹"}</span>
            <span className="sidebar-collapse-label">
              {sidebarCollapsed ? "展开侧栏" : "收起侧栏"}
            </span>
          </button>
        </div>
      </aside>

      <div className="v4-workspace">
        <header className="v4-topbar v4-page-topbar">
          <nav className="v4-topnav" aria-label="顶部全局功能区">
            {topNav.map((item) => (
              <Link
                key={item.href}
                className={isTopNavActive(item.href) ? "active" : item.featured ? "featured" : ""}
                to={item.href}
              >
                {item.featured && <span className="mini-logo" aria-hidden="true" />}
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="v4-account-area">
            <Link aria-label="通知" className="bell-button" to="/messages">
              <span />
            </Link>
            <div className="v4-account">
              <button
                aria-expanded={accountOpen}
                aria-label={`${nickname}的账号菜单`}
                className="v4-account-trigger"
                onClick={() => setAccountOpen((open) => !open)}
                type="button"
              >
                <span className="v4-avatar" aria-hidden="true">张</span>
                <span className="v4-user-copy">
                  <strong>{nickname} · 智活AI</strong>
                  <small>企业管理员</small>
                </span>
                <span aria-hidden="true">⌄</span>
              </button>
              {accountOpen && (
                <div className="v4-account-menu page-shell-account" role="dialog" aria-label="头像下拉框">
                  <div className="account-card-head">
                    <span className="v4-avatar large" aria-hidden="true">张</span>
                    <div>
                      <strong>{nickname}</strong>
                      <p>
                        <small>企业管理员</small>
                        <small className="gold">企业版</small>
                      </p>
                    </div>
                  </div>
                  <div className="account-menu-list">
                    {accountLinks.map(([label, href, icon]) => (
                      <Link key={label} to={href}>
                        <span className={`account-menu-icon ${icon}`} aria-hidden="true" />
                        {label}
                        <b aria-hidden="true">›</b>
                      </Link>
                    ))}
                  </div>
                  <button className="logout-button" onClick={logout} type="button">
                    <span className="account-menu-icon logout" aria-hidden="true" />
                    退出登录
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        <main className="v4-page-main">
          {children}
        </main>

      </div>
    </div>
  );
}

export default V4PageShell;
