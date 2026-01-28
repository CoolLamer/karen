import React, { createContext, useContext, useEffect, useState } from "react";
import { flushSync } from "react-dom";
import { api, User, Tenant, setAuthToken, getAuthToken } from "./api";

type AuthState = {
  isLoading: boolean;
  isAuthenticated: boolean;
  user: User | null;
  tenant: Tenant | null;
  needsOnboarding: boolean;
  isAdmin: boolean;
  onboardingInProgress: boolean; // Session-only, not persisted
};

type AuthContextType = AuthState & {
  login: (token: string, user: User) => boolean;
  logout: () => Promise<void>;
  setTenant: (tenant: Tenant) => void;
  refreshUser: () => Promise<void>;
  startOnboarding: () => void;
  finishOnboarding: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    isLoading: true,
    isAuthenticated: false,
    user: null,
    tenant: null,
    needsOnboarding: false,
    isAdmin: false,
    onboardingInProgress: false,
  });

  const loadUser = async () => {
    const token = getAuthToken();
    if (!token) {
      setState((prev) => ({
        isLoading: false,
        isAuthenticated: false,
        user: null,
        tenant: null,
        needsOnboarding: false,
        isAdmin: false,
        onboardingInProgress: prev.onboardingInProgress,
      }));
      return;
    }

    try {
      const data = await api.getMe();
      setState((prev) => ({
        isLoading: false,
        isAuthenticated: true,
        user: data.user,
        tenant: data.tenant ?? null,
        // Only needs onboarding if no tenant AND not currently in onboarding flow
        needsOnboarding: !data.tenant && !prev.onboardingInProgress,
        isAdmin: data.is_admin ?? false,
        onboardingInProgress: prev.onboardingInProgress, // Preserve session flag
      }));
    } catch (err) {
      // Only clear auth on 401 (invalid/expired token).
      // Network errors (e.g. backend briefly down during deploy) should
      // preserve the token so the user isn't logged out unnecessarily.
      const isAuthError = err instanceof Error && "status" in err && (err as { status: number }).status === 401;
      if (isAuthError) {
        setAuthToken(null);
      }
      setState((prev) => ({
        isLoading: false,
        isAuthenticated: isAuthError ? false : prev.isAuthenticated,
        user: isAuthError ? null : prev.user,
        tenant: isAuthError ? null : prev.tenant,
        needsOnboarding: isAuthError ? false : prev.needsOnboarding,
        isAdmin: isAuthError ? false : prev.isAdmin,
        onboardingInProgress: prev.onboardingInProgress,
      }));
    }
  };

  useEffect(() => {
    loadUser();
  }, []);

  // Background token refresh - keeps session alive for active users
  useEffect(() => {
    if (!state.isAuthenticated) return;

    // Refresh token every 6 days (before 7-day expiry)
    const REFRESH_INTERVAL = 6 * 24 * 60 * 60 * 1000; // 6 days in ms

    const refreshToken = async () => {
      const token = getAuthToken();
      if (!token) return;

      try {
        const data = await api.refreshToken(token);
        setAuthToken(data.token);
      } catch {
        // Silent fail - will be handled on next API call
      }
    };

    const interval = setInterval(refreshToken, REFRESH_INTERVAL);
    return () => clearInterval(interval);
  }, [state.isAuthenticated]);

  const login = (token: string, user: User): boolean => {
    setAuthToken(token);
    const needsOnboarding = !user.tenant_id;
    // Use flushSync to ensure state is committed synchronously before navigation.
    // This prevents route guards from reading stale state during the redirect.
    flushSync(() => {
      setState((prev) => ({
        ...prev,
        isAuthenticated: true,
        user,
        needsOnboarding,
        isAdmin: false, // Will be set correctly when loadUser completes
        onboardingInProgress: prev.onboardingInProgress,
      }));
    });
    // Load full user data including tenant and admin status
    loadUser();
    // Return needsOnboarding for immediate use in navigation
    return needsOnboarding;
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // Ignore errors during logout
    }
    setAuthToken(null);
    setState({
      isLoading: false,
      isAuthenticated: false,
      user: null,
      tenant: null,
      needsOnboarding: false,
      isAdmin: false,
      onboardingInProgress: false,
    });
  };

  const setTenant = (tenant: Tenant) => {
    setState((prev) => ({
      ...prev,
      tenant,
      needsOnboarding: false,
    }));
  };

  const refreshUser = async () => {
    await loadUser();
  };

  const startOnboarding = () => {
    setState((prev) => ({ ...prev, onboardingInProgress: true }));
  };

  const finishOnboarding = async () => {
    setState((prev) => ({ ...prev, onboardingInProgress: false }));
    await loadUser(); // Refresh state - now backend tenant determines needsOnboarding
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        logout,
        setTenant,
        refreshUser,
        startOnboarding,
        finishOnboarding,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
