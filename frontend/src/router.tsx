import React from "react";
import { createBrowserRouter, Navigate, Outlet } from "react-router-dom";
import { useAuth } from "./AuthContext";

// Pages
import { LandingPage, SegmentLandingPage } from "./ui/landing";
import {
  CompanyInfoPage,
  TermsOfServicePage,
  PrivacyPolicyPage,
} from "./ui/landing/legal";
import { LoginPage } from "./ui/LoginPage";
import { OnboardingPage } from "./ui/OnboardingPage";
import { AppShellLayout } from "./ui/AppShellLayout";
import { CallInboxPage } from "./ui/CallInboxPage";
import { CallDetailPage } from "./ui/CallDetailPage";
import { SettingsPage } from "./ui/SettingsPage";
import { AdminPhoneNumbersPage } from "./ui/AdminPhoneNumbersPage";
import { AdminLogsPage } from "./ui/AdminLogsPage";
import { AdminUsersPage } from "./ui/AdminUsersPage";
import { AdminGlobalConfigPage } from "./ui/AdminGlobalConfigPage";
import { AdminShellLayout } from "./ui/AdminShellLayout";

// Protected route wrapper - requires authentication
function ProtectedRoute() {
  const { isLoading, isAuthenticated, needsOnboarding } = useAuth();

  if (isLoading) {
    return null; // Or a loading spinner
  }

  if (!isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  if (needsOnboarding) {
    return <Navigate to="/onboarding" replace />;
  }

  return <Outlet />;
}

// Onboarding route - requires auth but not completed onboarding
function OnboardingRoute() {
  const { isLoading, isAuthenticated, needsOnboarding, onboardingInProgress } = useAuth();

  if (isLoading) {
    return null;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Stay on onboarding if needsOnboarding is true OR if we're in the middle of onboarding flow
  // (onboardingInProgress prevents redirect when loadUser() sets needsOnboarding to false)
  if (!needsOnboarding && !onboardingInProgress) {
    return <Navigate to="/" replace />;
  }

  return <OnboardingPage />;
}

// Public route - redirects to app if already authenticated
function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isLoading, isAuthenticated, needsOnboarding } = useAuth();

  if (isLoading) {
    return null;
  }

  if (isAuthenticated) {
    if (needsOnboarding) {
      return <Navigate to="/onboarding" replace />;
    }
    return <Navigate to="/inbox" replace />;
  }

  return <>{children}</>;
}

// Home route - shows landing for unauthenticated, inbox for authenticated
function HomeRoute() {
  const { isLoading, isAuthenticated, needsOnboarding } = useAuth();

  if (isLoading) {
    return null;
  }

  if (!isAuthenticated) {
    return <LandingPage />;
  }

  if (needsOnboarding) {
    return <Navigate to="/onboarding" replace />;
  }

  return <Navigate to="/inbox" replace />;
}

export const router = createBrowserRouter([
  // Home - landing or redirect to inbox
  {
    path: "/",
    element: <HomeRoute />,
  },

  // Segment landing pages (public)
  {
    path: "/pro-techniky",
    element: <SegmentLandingPage segmentKey="technicians" />,
  },
  {
    path: "/pro-lekare",
    element: <SegmentLandingPage segmentKey="professionals" />,
  },
  {
    path: "/pro-maklere",
    element: <SegmentLandingPage segmentKey="sales" />,
  },
  {
    path: "/pro-manazery",
    element: <SegmentLandingPage segmentKey="managers" />,
  },

  // Legal pages (public)
  {
    path: "/obchodni-podminky",
    element: <TermsOfServicePage />,
  },
  {
    path: "/ochrana-osobnich-udaju",
    element: <PrivacyPolicyPage />,
  },
  {
    path: "/informace-o-provozovateli",
    element: <CompanyInfoPage />,
  },

  // Login - public only
  {
    path: "/login",
    element: (
      <PublicRoute>
        <LoginPage />
      </PublicRoute>
    ),
  },

  // Onboarding - authenticated but needs setup
  {
    path: "/onboarding",
    element: <OnboardingRoute />,
  },

  // Protected app routes
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppShellLayout />,
        children: [
          { path: "/inbox", element: <CallInboxPage /> },
          { path: "/calls/:providerCallId", element: <CallDetailPage /> },
        ],
      },
      { path: "/settings", element: <SettingsPage /> },
      {
        element: <AdminShellLayout />,
        children: [
          { path: "/admin", element: <AdminPhoneNumbersPage /> },
          { path: "/admin/users", element: <AdminUsersPage /> },
          { path: "/admin/logs", element: <AdminLogsPage /> },
          { path: "/admin/logs/:providerCallId", element: <AdminLogsPage /> },
          { path: "/admin/config", element: <AdminGlobalConfigPage /> },
        ],
      },
    ],
  },
]);
