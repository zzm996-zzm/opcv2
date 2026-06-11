import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";

import { authApi } from "./lib/authApi";
import { authSession, useAuthSession } from "./lib/authSession";
import HomePage from "./pages/HomePage";
import LegalPage from "./pages/LegalPage";
import LoginPage from "./pages/LoginPage";
import ProductPlaceholder from "./pages/ProductPlaceholder";

const productRoutes = [
  ["/analysis", "免费分析", "把你的资源和目标整理成可执行的方向。"],
  ["/leads", "VIP获客", "从公开来源发现企业线索，并沉淀到客户库。"],
  ["/insights", "咨询通", "查看行业动态、案例和商业机会证据。"],
  ["/tools", "工具箱", "集中管理适合一人公司的效率工具。"],
  ["/community", "社群", "查看活动、权益与同行交流入口。"]
] as const;

function App() {
  useEffect(() => {
    const session = authSession.get();
    if (session.accessToken || session.ready) return;

    let active = true;
    authApi
      .restore()
      .then((result) => {
        if (active) authSession.set(result);
      })
      .catch(() => {
        // A missing refresh cookie is the normal signed-out state.
      })
      .finally(() => {
        if (active) authSession.finishRestore();
      });

    return () => {
      active = false;
    };
  }, []);

  return (
    <Routes>
      <Route element={<HomePage />} path="/" />
      <Route element={<LoginPage />} path="/login" />
      <Route element={<LegalPage kind="terms" />} path="/terms" />
      <Route element={<LegalPage kind="privacy" />} path="/privacy" />
      {productRoutes.map(([path, title, description]) => (
        <Route
          key={path}
          element={
            <RequireAuth>
              <ProductPlaceholder description={description} title={title} />
            </RequireAuth>
          }
          path={path}
        />
      ))}
    </Routes>
  );
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const session = useAuthSession();
  const location = useLocation();

  if (!session.ready) {
    return <main className="session-loading">正在恢复登录状态...</main>;
  }
  if (!session.user) {
    return <Navigate replace state={{ returnTo: location.pathname + location.search }} to="/login" />;
  }
  return children;
}

export default App;
