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
          <span className="brand-mark" aria-hidden="true" />
          <span>智活AI · OPC</span>
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
                <Link to="/membership">会员中心</Link>
                <button onClick={logout} type="button">
                  退出登录
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="header-actions">
            <Link className="ghost-action compact" to="/analysis">
              查看案例
            </Link>
            <Link className="primary-action compact" to="/login">
              开始体验
            </Link>
          </div>
        )}
      </header>

      <main className="hero">
        <h1 aria-label="一框输入，开始增长">一框输入，<span>开始增长</span></h1>
        <p>告诉 AI 你的资源、目标或问题，快速获得方向与行动建议。</p>

        <div className="intent-box">
          <label htmlFor="business-intent">输入你的资源、行业、目标客户，或你现在遇到的问题...</label>
          <div className="intent-control">
            <span className="input-spark" aria-hidden="true">✦</span>
            <textarea
              id="business-intent"
              rows={3}
              placeholder="我有 10 年教培经验，3 万预算，想在线上做副业，适合从什么方向切入？"
            />
            <Link to="/analysis" aria-label="提交需求">
              <span>开始分析</span>
            </Link>
          </div>
          <div className="intent-presets" aria-label="快捷问题">
            <Link to="/analysis?mode=direction">我适合做什么</Link>
            <Link to="/analysis?mode=competitor">拆解一个对标公司</Link>
            <Link to="/leads">帮我找成都教培客户</Link>
            <Link to="/tools">推荐适合我的工具</Link>
          </div>
        </div>

        <div className="hero-benefits" aria-label="产品能力">
          <div>
            <span>💬</span>
            <strong>AI 追问补充</strong>
            <small>更准确理解需求</small>
          </div>
          <div>
            <span>📈</span>
            <strong>输出行动建议</strong>
            <small>给出下一步方向</small>
          </div>
          <div>
            <span>🔄</span>
            <strong>持续优化结果</strong>
            <small>边用边迭代</small>
          </div>
        </div>
      </main>
    </div>
  );
}

function maskPhone(phone: string) {
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, "$1****$2");
}

export default HomePage;
