import type { ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";

import { useAuthSession } from "../lib/authSession";

const topNav = [
  ["工作台", "/"],
  ["智活 Copilot", "/copilot"],
  ["工具箱", "/tools"],
  ["咨询通", "/insights"],
  ["AI社群", "/community"],
  ["AI教学", "/learning"]
] as const;

const sideGroups = [
  {
    title: "项目确定及拆解",
    items: [["项目超市", "/projects", "grid"], ["商业沙盘", "/sandbox", "diamond"]]
  },
  {
    title: "落地",
    items: [
      ["任务中心", "/tasks", "check"],
      ["竞品全盘数据破解", "/competitor-data", "stack"],
      ["竞品动态监测", "/competitor-monitoring", "pulse"],
      ["增长测算", "/growth-calculator", "calc"]
    ]
  },
  {
    title: "增长",
    items: [
      ["GEO获客", "/geo", "target"],
      ["AI线索开发", "/leads", "diamond"],
      ["仪表盘", "/dashboard", "dashboard"],
      ["CRM客户管理", "/crm", "crm"],
      ["企业定制化陪跑", "/enterprise", "enterprise"]
    ]
  }
] as const;

type ReferenceShellProps = {
  accountSlot?: ReactNode;
  children: ReactNode;
  className?: string;
  mainClassName?: string;
};

function ReferenceShell({ accountSlot, children, className = "", mainClassName = "" }: ReferenceShellProps) {
  const location = useLocation();
  const session = useAuthSession();
  const nickname = session.user?.nickname || "张婧";
  const isActive = (href: string) => location.pathname === href || (href !== "/" && location.pathname.startsWith(`${href}/`));

  return (
    <div className={`ref-shell ${className}`}>
      <aside className="ref-sidebar" aria-label="产品侧边导航">
        <Link className="ref-brand" to="/" aria-label="智活AI OPC V4.0 首页">
          <img alt="" src="/home/logo.png" />
          <strong>智活AI</strong>
          <small>OPC V4.0</small>
        </Link>

        <nav className="ref-side-nav" aria-label="三大板块导航">
          {sideGroups.map((group) => (
            <section className="ref-side-group" key={group.title}>
              <div className="ref-side-title"><strong>{group.title}</strong><span>⌄</span></div>
              {group.items.map(([label, href, icon]) => (
                <Link className={`ref-side-link ${isActive(href) ? "active" : ""}`} key={href} to={href}>
                  <i className={`ref-side-icon ${icon}`} aria-hidden="true" />
                  <span>{label}</span>
                </Link>
              ))}
            </section>
          ))}
        </nav>

        <div className="ref-side-footer">
          <Link className={`ref-side-link ref-message-link ${isActive("/messages") ? "active" : ""}`} to="/messages">
            <i className="ref-side-icon chat" aria-hidden="true" />
            <span>消息中心</span>
          </Link>
          <Link className={`ref-side-link ${isActive("/help") ? "active" : ""}`} to="/help">
            <i className="ref-side-icon help" aria-hidden="true" />
            <span>帮助与反馈</span>
          </Link>
          <button className="ref-collapse" type="button"><span>‹</span> 收起侧栏</button>
        </div>
      </aside>

      <div className="ref-workspace">
        <header className="ref-topbar">
          <Link className="ref-mobile-brand" to="/" aria-label="智活AI 首页">
            <img alt="" src="/home/logo.png" />
            <strong>智活AI</strong>
          </Link>
          <nav className="ref-topnav" aria-label="顶部全局功能区">
            {topNav.map(([label, href]) => (
              <Link className={`${isActive(href) ? "active" : ""} ${href === "/copilot" ? "copilot" : ""}`} key={href} to={href}>
                {href === "/copilot" ? <img alt="" src="/home/logo.png" /> : null}
                {label}
              </Link>
            ))}
          </nav>

          {accountSlot ?? (
            <div className="ref-account-area">
              <Link aria-label="通知" className="ref-notice" to="/messages"><span /></Link>
              <Link className="ref-account" to="/profile">
                <img alt="" src="/public-components/avatar.jpg" />
                <span><strong>{nickname} · 智活AI</strong><small>企业管理员</small></span>
                <b>⌄</b>
              </Link>
            </div>
          )}
        </header>

        <main className={`ref-page-main ${mainClassName}`}>{children}</main>
      </div>
    </div>
  );
}

export default ReferenceShell;
