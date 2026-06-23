import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";

import { authApi } from "./lib/authApi";
import { authSession, useAuthSession } from "./lib/authSession";
import AnalysisPage from "./pages/AnalysisPage";
import CommunityEnterprisePage from "./pages/CommunityEnterprisePage";
import CommunityMembersPage from "./pages/CommunityMembersPage";
import CommunityPage from "./pages/CommunityPage";
import CopilotPage from "./pages/CopilotPage";
import CompetitorDataPage from "./pages/CompetitorDataPage";
import HelpPage from "./pages/HelpPage";
import HomePage from "./pages/HomePage";
import InsightsPage from "./pages/InsightsPage";
import LegalPage from "./pages/LegalPage";
import LearningAssessmentPage from "./pages/LearningAssessmentPage";
import LearningCourseDetailPage from "./pages/LearningCourseDetailPage";
import LearningCourseIntroPage from "./pages/LearningCourseIntroPage";
import LearningCoursesPage from "./pages/LearningCoursesPage";
import LearningDiagnosisPage from "./pages/LearningDiagnosisPage";
import LearningGapAnalysisPage from "./pages/LearningGapAnalysisPage";
import LearningHistoryPage from "./pages/LearningHistoryPage";
import LearningPage from "./pages/LearningPage";
import LearningPlanPage from "./pages/LearningPlanPage";
import LearningRecommendationPage from "./pages/LearningRecommendationPage";
import LearningRecommendedCoursesPage from "./pages/LearningRecommendedCoursesPage";
import LearningReportPage from "./pages/LearningReportPage";
import LoginPage from "./pages/LoginPage";
import MembershipPage from "./pages/MembershipPage";
import MessagesPage from "./pages/MessagesPage";
import ProductPlaceholder from "./pages/ProductPlaceholder";
import ProjectsPage from "./pages/ProjectsPage";
import ProfilePage from "./pages/ProfilePage";
import RegisterDetailsPage from "./pages/RegisterDetailsPage";
import SandboxPage from "./pages/SandboxPage";
import TasksPage from "./pages/TasksPage";
import ToolsPage from "./pages/ToolsPage";

const productRoutes = [
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
            <ToolsPage variant="all" />
          </RequireAuth>
        }
        path="/tools/all"
      />
      <Route
        element={
          <RequireAuth>
            <ToolsPage variant="recommend" />
          </RequireAuth>
        }
        path="/tools/recommend"
      />
      <Route
        element={
          <RequireAuth>
            <ToolsPage variant="plan" />
          </RequireAuth>
        }
        path="/tools/recommendation-plan"
      />
      <Route
        element={
          <RequireAuth>
            <ToolsPage variant="detail" />
          </RequireAuth>
        }
        path="/tools/detail"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage />
          </RequireAuth>
        }
        path="/copilot"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="new" />
          </RequireAuth>
        }
        path="/copilot/new"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="models" />
          </RequireAuth>
        }
        path="/copilot/models"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="files" />
          </RequireAuth>
        }
        path="/copilot/files"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="compare" />
          </RequireAuth>
        }
        path="/copilot/compare"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="rename" />
          </RequireAuth>
        }
        path="/copilot/rename"
      />
      <Route
        element={
          <RequireAuth>
            <CopilotPage variant="delete" />
          </RequireAuth>
        }
        path="/copilot/delete"
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
            <LearningAssessmentPage />
          </RequireAuth>
        }
        path="/learning/assessment"
      />
      <Route
        element={
          <RequireAuth>
            <LearningGapAnalysisPage />
          </RequireAuth>
        }
        path="/learning/gap-analysis"
      />
      <Route
        element={
          <RequireAuth>
            <LearningRecommendationPage />
          </RequireAuth>
        }
        path="/learning/recommendation"
      />
      <Route
        element={
          <RequireAuth>
            <LearningPlanPage />
          </RequireAuth>
        }
        path="/learning/plan"
      />
      <Route
        element={
          <RequireAuth>
            <LearningReportPage />
          </RequireAuth>
        }
        path="/learning/report"
      />
      <Route
        element={
          <RequireAuth>
            <LearningCoursesPage />
          </RequireAuth>
        }
        path="/learning/courses"
      />
      <Route
        element={
          <RequireAuth>
            <LearningCourseIntroPage />
          </RequireAuth>
        }
        path="/learning/courses/intro"
      />
      <Route
        element={
          <RequireAuth>
            <LearningCourseDetailPage />
          </RequireAuth>
        }
        path="/learning/courses/detail"
      />
      <Route
        element={
          <RequireAuth>
            <LearningRecommendedCoursesPage />
          </RequireAuth>
        }
        path="/learning/recommended-courses"
      />
      <Route
        element={
          <RequireAuth>
            <LearningHistoryPage />
          </RequireAuth>
        }
        path="/learning/history"
      />
      <Route
        element={
          <RequireAuth>
            <CommunityPage />
          </RequireAuth>
        }
        path="/community"
      />
      <Route
        element={
          <RequireAuth>
            <CommunityMembersPage />
          </RequireAuth>
        }
        path="/community/members"
      />
      <Route
        element={
          <RequireAuth>
            <CommunityEnterprisePage />
          </RequireAuth>
        }
        path="/community/enterprise"
      />
      <Route
        element={
          <RequireAuth>
            <InsightsPage />
          </RequireAuth>
        }
        path="/insights"
      />
      <Route
        element={
          <RequireAuth>
            <InsightsPage variant="detail" />
          </RequireAuth>
        }
        path="/insights/detail"
      />
      <Route
        element={
          <RequireAuth>
            <InsightsPage variant="fileAnalysis" />
          </RequireAuth>
        }
        path="/insights/file-analysis"
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
            <SandboxPage />
          </RequireAuth>
        }
        path="/sandbox"
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
            <CompetitorDataPage />
          </RequireAuth>
        }
        path="/competitor-data"
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
