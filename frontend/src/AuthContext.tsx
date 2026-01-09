import React, { createContext, useContext, useEffect, useState } from "react";
import { api, User, Tenant, setAuthToken, getAuthToken } from "./api";

type AuthState = {
  isLoading: boolean;
  isAuthenticated: boolean;
  user: User | null;
  tenant: Tenant | null;
  needsOnboarding: boolean;
};

type AuthContextType = AuthState & {
  login: (token: string, user: User) => void;
  logout: () => Promise<void>;
  setTenant: (tenant: Tenant) => void;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    isLoading: true,
    isAuthenticated: false,
    user: null,
    tenant: null,
    needsOnboarding: false,
  });

  const loadUser = async () => {
    const token = getAuthToken();
    if (!token) {
      setState({
        isLoading: false,
        isAuthenticated: false,
        user: null,
        tenant: null,
        needsOnboarding: false,
      });
      return;
    }

    try {
      const data = await api.getMe();
      setState({
        isLoading: false,
        isAuthenticated: true,
        user: data.user,
        tenant: data.tenant ?? null,
        needsOnboarding: !data.tenant,
      });
    } catch {
      setAuthToken(null);
      setState({
        isLoading: false,
        isAuthenticated: false,
        user: null,
        tenant: null,
        needsOnboarding: false,
      });
    }
  };

  useEffect(() => {
    loadUser();
  }, []);

  const login = (token: string, user: User) => {
    setAuthToken(token);
    setState((prev) => ({
      ...prev,
      isAuthenticated: true,
      user,
      needsOnboarding: !user.tenant_id,
    }));
    // Load full user data including tenant
    loadUser();
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

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        logout,
        setTenant,
        refreshUser,
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
