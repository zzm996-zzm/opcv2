import {
  Activity,
  BarChart3,
  Calculator,
  CircleCheckBig,
  CircleHelp,
  Database,
  Diamond,
  Flag,
  House,
  MessageCircle,
  Radar,
  Store,
  Users,
  type LucideIcon
} from "lucide-react";
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
      { label: "增长洞察", href: "/growth-calculator", icon: "calc" }
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

const accountLinks: Array<[string, string, string]> = [
  ["个人中心", "/profile", "user"],
  ["账号与资料设置", "/profile/settings", "settings"],
  ["会员与账单", "/membership", "wallet"],
  ["我的内容", "/profile/content", "content"],
  ["偏好设置", "/profile/preferences", "gear"]
];

const sidebarIcons: Record<string, LucideIcon> = {
  calc: Calculator,
  chart: BarChart3,
  check: CircleCheckBig,
  diamond: Diamond,
  flag: Flag,
  grid: Store,
  home: House,
  pulse: Activity,
  stack: Database,
  target: Radar,
  user: Users
};

function SidebarIcon({ name }: { name: string }) {
  const Icon = sidebarIcons[name];
  return Icon ? <Icon aria-hidden="true" className="v4-lucide-icon" /> : null;
}

type V4PageShellProps = {
  accountSlot?: ReactNode;
  children: ReactNode;
  className?: string;
  mainClassName?: string;
  showCopilotMini?: boolean;
};

function V4PageShell({ accountSlot, children, className = "", mainClassName = "" }: V4PageShellProps) {
  const session = useAuthSession();
  const location = useLocation();
  const [accountOpen, setAccountOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const nickname = session.user?.nickname || "张婧";
  const isTopNavActive = (href: string) => {
    if (href !== "/") {
      return location.pathname === href || location.pathname.startsWith(`${href}/`);
    }
    // Profile, billing and utility routes are still part of the workbench.
    return !topNav.some((item) => item.href !== "/" && (location.pathname === item.href || location.pathname.startsWith(`${item.href}/`)));
  };

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
          {defaultSidebarGroups.map((group) => (
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
                  <SidebarIcon name={item.icon} />
                  <span>{item.label}</span>
                </Link>
              ))}
            </section>
          ))}
        </nav>

        <div className="v4-sidebar-bottom">
          <Link className={`v4-side-link with-dot ${location.pathname.startsWith("/messages") ? "active" : ""}`} to="/messages">
            <MessageCircle aria-hidden="true" className="v4-lucide-icon" />
            <span>消息中心</span>
          </Link>
          <Link className="v4-side-link" to="/help">
            <CircleHelp aria-hidden="true" className="v4-lucide-icon" />
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

          {accountSlot ?? <div className="v4-account-area">
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
                <img className="v4-avatar" alt="" src="/public-components/avatar.jpg" />
                <span className="v4-user-copy">
                  <strong>{nickname} · 智活AI</strong>
                  <small>企业管理员</small>
                </span>
                <span aria-hidden="true">⌄</span>
              </button>
              {accountOpen && (
                <div className="v4-account-menu page-shell-account" role="dialog" aria-label="头像下拉框">
                  <div className="account-card-head">
                    <img className="v4-avatar large" alt="" src="/public-components/avatar.jpg" />
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
          </div>}
        </header>

        <main className={`v4-page-main ${mainClassName}`}>
          {children}
        </main>

      </div>
    </div>
  );
}

export default V4PageShell;
