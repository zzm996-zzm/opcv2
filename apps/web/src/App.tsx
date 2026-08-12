import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";

import FloatingCopilotOrb from "./components/FloatingCopilotOrb";
import { CopilotPanelVisibilityProvider, useCopilotPanelVisibility } from "./components/CopilotPanelVisibility";
import { authApi } from "./lib/authApi";
import { authSession, useAuthSession } from "./lib/authSession";
import AnalysisHistoryPage from "./pages/AnalysisHistoryPage";
import AnalysisPage from "./pages/AnalysisPage";
import AnalysisReportPage from "./pages/AnalysisReportPage";
import CommunityEnterprisePage from "./pages/CommunityEnterprisePage";
import CommunityMembersPage from "./pages/CommunityMembersPage";
import CommunityPage from "./pages/CommunityPage";
import CopilotPage from "./pages/CopilotPage";
import CompetitorMonitoringPage from "./pages/CompetitorMonitoringPage";
import CrmPage from "./pages/CrmPage";
import DashboardPage from "./pages/DashboardPage";
import EnterprisePage from "./pages/EnterprisePage";
import EnterpriseReferencePage from "./pages/EnterpriseReferencePage";
import GeoAcquisitionPage from "./pages/GeoAcquisitionPage";
import GrowthCalculatorPage from "./pages/GrowthCalculatorPage";
import HelpPage from "./pages/HelpPage";
import HomePage from "./pages/HomePage";
import InsightsPage from "./pages/InsightsPage";
import LeadDevelopmentPage from "./pages/LeadDevelopmentPage";
import LandingReferencePage from "./pages/LandingReferencePage";
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
import MembershipPaymentPage from "./pages/MembershipPaymentPage";
import MessagesPage from "./pages/MessagesPage";
import NotFoundPage from "./pages/NotFoundPage";
import ProjectsPage from "./pages/ProjectsPage";
import ProfilePage from "./pages/ProfilePage";
import RegisterDetailsPage from "./pages/RegisterDetailsPage";
import SandboxPage from "./pages/SandboxPage";
import ToolsPage from "./pages/ToolsPage";
import V4PageShell from "./components/V4PageShell";

function AppRoutes() {
  const location = useLocation();
  const session = useAuthSession();
  const showGlobalCopilotOrb = session.ready && shouldShowGlobalCopilotOrb(location.pathname);
  const keepSandboxOrbVisible = location.pathname === "/sandbox";
  const copilotPanel = useCopilotPanelVisibility();

  useEffect(() => {
    const session = authSession.get();
    if (session.accessToken || session.ready) return;

    const allowDevAuth = import.meta.env.DEV && (
      window.location.hostname === "127.0.0.1" ||
      window.location.hostname === "localhost" ||
      window.location.hostname === "::1"
    );

    if (allowDevAuth && new URLSearchParams(window.location.search).get("devAuth") === "1") {
      authSession.set({
        access_token: "dev-access-token",
        access_token_expires_at: "2026-12-31T23:59:59Z",
        is_new_user: false,
        user: { id: 7, nickname: "张晨", phone: "13800138000", status: "active" }
      });
      return;
    }

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
    <>
      <div className="route-motion-frame" key={location.pathname}>
      <Routes location={location}>
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
      <Route element={<LoginPage initialMode="register" />} path="/register" />
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
      <Route element={<RequireAuth><MembershipPaymentPage mode="checkout" /></RequireAuth>} path="/membership/checkout" />
      <Route element={<RequireAuth><MembershipPaymentPage mode="quota" /></RequireAuth>} path="/membership/quota" />
      <Route element={<RequireAuth><MembershipPaymentPage mode="success" /></RequireAuth>} path="/membership/success" />
      <Route element={<RequireAuth><Navigate replace to="/profile/settings" /></RequireAuth>} path="/account/settings" />
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
            <AnalysisHistoryPage />
          </RequireAuth>
        }
        path="/analysis/history"
      />
      <Route
        element={
          <RequireAuth>
            <AnalysisReportPage />
          </RequireAuth>
        }
        path="/analysis/sessions/:sessionId"
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
            <CopilotPage variant="memories" />
          </RequireAuth>
        }
        path="/copilot/memories"
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
            <LearningCourseDetailPage />
          </RequireAuth>
        }
        path="/learning/courses/:courseSlug/study"
      />
      <Route
        element={
          <RequireAuth>
            <LearningCourseIntroPage />
          </RequireAuth>
        }
        path="/learning/courses/:courseSlug"
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
            <ProjectsPage variant="match" />
          </RequireAuth>
        }
        path="/projects/match"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="explore" />
          </RequireAuth>
        }
        path="/projects/explore"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="cases" />
          </RequireAuth>
        }
        path="/projects/cases"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="caseDetail" />
          </RequireAuth>
        }
        path="/project-cases/:caseRef"
      />
      <Route
        element={
          <RequireAuth>
            <Navigate replace to="/projects/match" />
          </RequireAuth>
        }
        path="/projects/questions"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="questions" />
          </RequireAuth>
        }
        path="/projects/matches/:matchId/questions"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="results" />
          </RequireAuth>
        }
        path="/projects/results"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="results" />
          </RequireAuth>
        }
        path="/projects/matches/:matchId/results"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="paywall" />
          </RequireAuth>
        }
        path="/projects/results/paywall"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="paywall" />
          </RequireAuth>
        }
        path="/projects/matches/:matchId/results/paywall"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="history" />
          </RequireAuth>
        }
        path="/projects/history"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="detail" />
          </RequireAuth>
        }
        path="/projects/opportunities/:opportunitySlug"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="detailUnlock" />
          </RequireAuth>
        }
        path="/projects/opportunities/:opportunitySlug/unlock"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="detailUnlock" />
          </RequireAuth>
        }
        path="/projects/:projectRef/unlock"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="detail" />
          </RequireAuth>
        }
        path="/projects/:projectRef"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="diagnosis" />
          </RequireAuth>
        }
        path="/projects/opportunities/:opportunitySlug/diagnosis"
      />
      <Route
        element={
          <RequireAuth>
            <Navigate replace to="/projects/explore" />
          </RequireAuth>
        }
        path="/projects/detail"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="detail" />
          </RequireAuth>
        }
        path="/projects/matches/:matchId"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="compare" />
          </RequireAuth>
        }
        path="/projects/compare"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="export" />
          </RequireAuth>
        }
        path="/projects/export"
      />
      <Route
        element={
          <RequireAuth>
            <ProjectsPage variant="export" />
          </RequireAuth>
        }
        path="/projects/matches/:matchId/export"
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
            <SandboxPage variant="setup" />
          </RequireAuth>
        }
        path="/sandbox/setup"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="roles" />
          </RequireAuth>
        }
        path="/sandbox/roles"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="start" />
          </RequireAuth>
        }
        path="/sandbox/start"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="questions" />
          </RequireAuth>
        }
        path="/sandbox/questions"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="run" />
          </RequireAuth>
        }
        path="/sandbox/run"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="report" />
          </RequireAuth>
        }
        path="/sandbox/report"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="report" />
          </RequireAuth>
        }
        path="/sandbox/sessions/:sessionId/report"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="history" />
          </RequireAuth>
        }
        path="/sandbox/history"
      />
      <Route
        element={
          <RequireAuth>
            <SandboxPage variant="quota" />
          </RequireAuth>
        }
        path="/sandbox/quota"
      />
      <Route
        element={
          <RequireAuth>
            <LandingReferencePage module="tasks" view="create" />
          </RequireAuth>
        }
        path="/tasks/new"
      />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="create-menu" /></RequireAuth>} path="/tasks/new/menu" />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="board" /></RequireAuth>} path="/tasks/board" />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="calendar" /></RequireAuth>} path="/tasks/calendar" />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="ai" /></RequireAuth>} path="/tasks/ai" />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="detail" /></RequireAuth>} path="/tasks/detail" />
      <Route element={<RequireAuth><LandingReferencePage module="tasks" view="list-menu" /></RequireAuth>} path="/tasks/menu" />
      <Route
        element={
          <RequireAuth>
            <LandingReferencePage module="tasks" view="list" />
          </RequireAuth>
        }
        path="/tasks"
      />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="progress" /></RequireAuth>} path="/competitor-data/progress" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="history" /></RequireAuth>} path="/competitor-data/history" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="overview" /></RequireAuth>} path="/competitor-data/results/overview" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="content" /></RequireAuth>} path="/competitor-data/results/content" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="live" /></RequireAuth>} path="/competitor-data/results/live" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="product" /></RequireAuth>} path="/competitor-data/results/product" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="audience" /></RequireAuth>} path="/competitor-data/results/audience" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="ads" /></RequireAuth>} path="/competitor-data/results/ads" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="sentiment" /></RequireAuth>} path="/competitor-data/results/sentiment" />
      <Route element={<RequireAuth><LandingReferencePage module="data" view="compare" /></RequireAuth>} path="/competitor-data/results/compare" />
      <Route
        element={
          <RequireAuth>
            <LandingReferencePage module="data" view="home" />
          </RequireAuth>
        }
        path="/competitor-data"
      />
      <Route element={<RequireAuth><LandingReferencePage module="monitoring" view="history" /></RequireAuth>} path="/competitor-monitoring/history" />
      <Route element={<RequireAuth><LandingReferencePage module="monitoring" view="progress" /></RequireAuth>} path="/competitor-monitoring/progress" />
      <Route element={<RequireAuth><LandingReferencePage module="monitoring" view="analysis" /></RequireAuth>} path="/competitor-monitoring/analysis" />
      <Route
        element={
          <RequireAuth>
            <CompetitorMonitoringPage />
          </RequireAuth>
        }
        path="/competitor-monitoring"
      />
      <Route element={<RequireAuth><LandingReferencePage module="growth" view="questions" /></RequireAuth>} path="/growth-calculator/questions" />
      <Route element={<RequireAuth><LandingReferencePage module="growth" view="history" /></RequireAuth>} path="/growth-calculator/history" />
      <Route element={<RequireAuth><LandingReferencePage module="growth" view="report" /></RequireAuth>} path="/growth-calculator/report" />
      <Route
        element={
          <RequireAuth>
            <GrowthCalculatorPage />
          </RequireAuth>
        }
        path="/growth-calculator"
      />
      <Route
        element={
          <RequireAuth>
            <GeoAcquisitionPage />
          </RequireAuth>
        }
        path="/geo"
      />
      <Route
        element={
          <RequireAuth>
            <LeadDevelopmentPage />
          </RequireAuth>
        }
        path="/leads"
      />
      <Route
        element={
          <RequireAuth>
            <DashboardPage />
          </RequireAuth>
        }
        path="/dashboard"
      />
      <Route
        element={
          <RequireAuth>
            <CrmPage />
          </RequireAuth>
        }
        path="/crm"
      />
      <Route
        element={
          <RequireAuth>
            <CrmPage variant="followUps" />
          </RequireAuth>
        }
        path="/crm/follow-ups"
      />
      <Route
        element={<RequireAuth><EnterpriseReferencePage variant="form" /></RequireAuth>}
        path="/enterprise/form"
      />
      <Route
        element={<RequireAuth><EnterpriseReferencePage variant="cases" /></RequireAuth>}
        path="/enterprise/cases"
      />
      <Route
        element={<RequireAuth><EnterpriseReferencePage variant="detail" /></RequireAuth>}
        path="/enterprise/cases/:caseSlug"
      />
      <Route
        element={<RequireAuth><EnterpriseReferencePage variant="success" /></RequireAuth>}
        path="/enterprise/success"
      />
      <Route
        element={<RequireAuth><EnterpriseReferencePage variant="contact" /></RequireAuth>}
        path="/enterprise/contact"
      />
      <Route
        element={
          <RequireAuth>
            <EnterprisePage />
          </RequireAuth>
        }
        path="/enterprise"
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
      <Route element={<RequireAuth><NotFoundPage /></RequireAuth>} path="*" />
      </Routes>
      </div>
      {(keepSandboxOrbVisible || (copilotPanel?.hasPanel ? !copilotPanel.isPanelOpen : showGlobalCopilotOrb)) && (
        <FloatingCopilotOrb
          className={keepSandboxOrbVisible ? "sandbox-floating-orb" : location.pathname.startsWith("/membership") ? "membership-floating-orb" : ""}
          onActivate={copilotPanel?.hasPanel && !copilotPanel.isPanelOpen ? copilotPanel.openPanel : undefined}
        />
      )}
    </>
  );
}

function shouldShowGlobalCopilotOrb(pathname: string) {
  const homeRoutes = pathname === "/" || pathname.startsWith("/home/") || pathname.startsWith("/assistant/");
  const standaloneRoutes = pathname === "/login" || pathname.startsWith("/register/") || pathname === "/terms" || pathname === "/privacy";
  return !homeRoutes && !standaloneRoutes && !pathname.startsWith("/copilot") && !pathname.startsWith("/crm");
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const session = useAuthSession();
  const location = useLocation();

  if (!session.ready) {
    return <SessionRestoreShell />;
  }
  if (!session.user) {
    return <Navigate replace state={{ returnTo: location.pathname + location.search }} to="/login" />;
  }
  return children;
}

function SessionRestoreShell() {
  return (
    <V4PageShell className="session-restore-shell">
      <section className="session-restore-content" aria-label="正在恢复登录状态">
        <div className="session-restore-heading">
          <span className="session-restore-kicker">智活AI OPC</span>
          <h1>正在恢复工作台</h1>
          <p role="status" aria-label="恢复登录状态">正在恢复登录状态...</p>
        </div>
        <div className="session-restore-grid" aria-hidden="true">
          <span className="session-restore-block wide" />
          <span className="session-restore-block" />
          <span className="session-restore-block" />
          <span className="session-restore-block large" />
        </div>
      </section>
    </V4PageShell>
  );
}

function App() {
  return (
    <CopilotPanelVisibilityProvider>
      <AppRoutes />
    </CopilotPanelVisibilityProvider>
  );
}

export default App;
