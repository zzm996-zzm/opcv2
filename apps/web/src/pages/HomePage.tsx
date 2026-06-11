import { useState } from "react";
import { Link } from "react-router-dom";

import { authApi } from "../lib/authApi";
import { authSession, useAuthSession } from "../lib/authSession";

const navigation = [
  { label: "首页", href: "/" },
  { label: "免费分析", href: "/analysis" },
  { label: "VIP获客", href: "/leads" },
  { label: "咨询通", href: "/insights" },
  { label: "工具箱", href: "/tools" },
  { label: "社群", href: "/community" }
];

function HomePage() {
  const session = useAuthSession();
  const [accountOpen, setAccountOpen] = useState(false);

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
    <div className="app-shell antialiased">
      <header className="site-header">
        <Link className="brand" to="/" aria-label="智活AI OPC 首页">
          <span className="brand-mark">智</span>
          <span>智活AI</span>
        </Link>

        <nav className="main-nav" aria-label="主导航">
          {navigation.map((item) => (
            <Link key={item.href} to={item.href}>
              {item.label}
            </Link>
          ))}
        </nav>

        {session.user ? (
          <div className="account-control">
            <button
              aria-expanded={accountOpen}
              aria-label={`${session.user.nickname}的账号菜单`}
              className="account-trigger"
              onClick={() => setAccountOpen((open) => !open)}
              type="button"
            >
              <span className="account-avatar">{session.user.nickname.slice(0, 1)}</span>
              <span>{session.user.nickname}</span>
              <span aria-hidden="true">⌄</span>
            </button>
            {accountOpen && (
              <div className="account-menu">
                <strong>{session.user.nickname}</strong>
                <span>{maskPhone(session.user.phone)}</span>
                <button onClick={logout} type="button">
                  退出登录
                </button>
              </div>
            )}
          </div>
        ) : (
          <Link className="primary-action compact" to="/login">
            开始体验
          </Link>
        )}
      </header>

      <main className="hero">
        <div className="hero-kicker">
          <span />
          从判断方向，到找到客户
        </div>
        <h1>把商业想法，变成下一步行动</h1>
        <p>
          描述你的资源、经验或目标。AI 会补充关键问题，给出可执行的方向，
          再帮你找到真实企业线索。
        </p>

        <div className="intent-box">
          <label htmlFor="business-intent">今天想解决什么问题？</label>
          <div className="intent-control">
            <textarea
              id="business-intent"
              rows={3}
              placeholder="例如：我有 10 年教培经验和 5 万预算，在成都适合做什么？"
            />
            <button type="button" aria-label="提交需求">
              <span>开始分析</span>
              <span aria-hidden="true">↗</span>
            </button>
          </div>
        </div>

        <div className="quick-intents" aria-label="快捷问题">
          <Link to="/analysis?mode=direction">我适合做什么</Link>
          <Link to="/analysis?mode=competitor">拆解一个对标公司</Link>
          <Link to="/leads">帮我找成都教培客户</Link>
          <Link to="/tools">推荐适合我的工具</Link>
        </div>
      </main>
    </div>
  );
}

function maskPhone(phone: string) {
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, "$1****$2");
}

export default HomePage;
