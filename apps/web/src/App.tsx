import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";

import { authApi } from "./lib/authApi";
import { authSession, useAuthSession } from "./lib/authSession";
import AnalysisPage from "./pages/AnalysisPage";
import HelpPage from "./pages/HelpPage";
import HomePage from "./pages/HomePage";
import LegalPage from "./pages/LegalPage";
import LearningDiagnosisPage from "./pages/LearningDiagnosisPage";
import LearningPage from "./pages/LearningPage";
import LoginPage from "./pages/LoginPage";
import MembershipPage from "./pages/MembershipPage";
import MessagesPage from "./pages/MessagesPage";
import ProductPlaceholder from "./pages/ProductPlaceholder";
import ProjectsPage from "./pages/ProjectsPage";
import ProfilePage from "./pages/ProfilePage";
import RegisterDetailsPage from "./pages/RegisterDetailsPage";
import TasksPage from "./pages/TasksPage";
import ToolsPage from "./pages/ToolsPage";

const productRoutes = [
  ["/copilot", "智活 Copilot", "全屏对话、多模型问答与全站能力调度。"],
  ["/insights", "咨询通", "查看商业与 AI 资讯，追踪行业动态。"],
  ["/community", "AI社群", "会员社群与企业社群的真实引流入口。"],
  ["/sandbox", "商业沙盘", "用多角色推演判断项目可行性。"],
  ["/competitor-data", "竞品全盘数据破解", "发起脚本代查，拿到竞品数据与 AI 结论。"],
  ["/competitor-monitoring", "竞品动态监测", "持续盯招聘、内容、投放与新品动态。"],
  ["/growth-calculator", "增长测算", "用成本、客单价和转化假设测算营收。"],
  ["/geo", "GEO获客", "占位展示 AI 搜索收录获客能力。"],
  ["/leads", "AI线索开发", "占位展示精准线索开发能力。"],
  ["/dashboard", "仪表盘", "占位展示经营数据看板。"],
  ["/crm", "CRM客户管理", "占位展示客户全生命周期管理。"],
  ["/enterprise", "企业定制化陪跑", "展示企业陪跑服务、案例与咨询入口。"]
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
      <Route
        element={
          <RequireAuth>
            <HomePage menuState="notice" />
          </RequireAuth>
        }
        path="/home/notice"
      />
      <Route
        element={
          <RequireAuth>
            <HomePage menuState="account" />
          </RequireAuth>
        }
        path="/home/account"
      />
      <Route
        element={
          <RequireAuth>
            <HomePage assistantState="settings" />
          </RequireAuth>
        }
        path="/assistant/settings"
      />
      <Route
        element={
          <RequireAuth>
            <HomePage assistantState="files" />
          </RequireAuth>
        }
        path="/assistant/files"
      />
      <Route
        element={
          <RequireAuth>
            <HomePage assistantState="collapsed" />
          </RequireAuth>
        }
        path="/assistant/collapsed"
      />
      <Route element={<LoginPage />} path="/login" />
      <Route element={<RegisterDetailsPage />} path="/register/details" />
      <Route element={<LegalPage kind="terms" />} path="/terms" />
      <Route element={<LegalPage kind="privacy" />} path="/privacy" />
      <Route
        element={
          <RequireAuth>
            <MembershipPage />
          </RequireAuth>
        }
        path="/membership"
      />
      <Route
        element={
          <RequireAuth>
            <MembershipPage showUpgrade />
          </RequireAuth>
        }
        path="/membership/upgrade"
      />
      <Route
        element={
          <RequireAuth>
            <AnalysisPage />
          </RequireAuth>
        }
        path="/analysis"
      />
      <Route
        element={
          <RequireAuth>
            <MessagesPage />
          </RequireAuth>
        }
        path="/messages"
      />
      <Route
        element={
          <RequireAuth>
            <MessagesPage />
          </RequireAuth>
        }
        path="/messages/:messageId"
      />
      <Route
        element={
          <RequireAuth>
            <ToolsPage />
          </RequireAuth>
        }
        path="/tools"
      />
      <Route
        element={
          <RequireAuth>
            <LearningPage />
          </RequireAuth>
        }
        path="/learning"
      />
      <Route
        element={
          <RequireAuth>
            <LearningDiagnosisPage />
          </RequireAuth>
        }
        path="/learning/diagnosis"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage />
          </RequireAuth>
        }
        path="/projects"
      />
      <Route
        element={
          <RequireAuth>
            <TasksPage />
          </RequireAuth>
        }
        path="/tasks"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage />
          </RequireAuth>
        }
        path="/profile"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="settings" />
          </RequireAuth>
        }
        path="/profile/settings"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="settings" overlay="password" />
          </RequireAuth>
        }
        path="/profile/settings/password"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="settings" overlay="logout" />
          </RequireAuth>
        }
        path="/profile/settings/logout"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="settings" overlay="delete" />
          </RequireAuth>
        }
        path="/profile/settings/delete"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="settings" overlay="complete" />
          </RequireAuth>
        }
        path="/profile/settings/complete"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage binding="unbound" mode="settings" />
          </RequireAuth>
        }
        path="/profile/settings/unbound"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="content" />
          </RequireAuth>
        }
        path="/profile/content"
      />
      <Route
        element={
          <RequireAuth>
            <ProfilePage mode="preferences" />
          </RequireAuth>
        }
        path="/profile/preferences"
      />
      <Route
        element={
          <RequireAuth>
            <HelpPage />
          </RequireAuth>
        }
        path="/help"
      />
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
