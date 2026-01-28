export type ScreeningResult = {
  legitimacy_label: string;
  legitimacy_confidence: number;
  lead_label: string;
  intent_category: string;
  intent_text: string;
  entities_json: unknown;
  created_at: string;
};

export type CallListItem = {
  provider: string;
  provider_call_id: string;
  from_number: string;
  to_number: string;
  status: string;
  rejection_reason?: string | null;
  started_at: string;
  ended_at?: string | null;
  ended_by?: string | null;
  first_viewed_at?: string | null;
  resolved_at?: string | null;
  resolved_by?: string | null;
  screening?: ScreeningResult | null;
};

export type Utterance = {
  speaker: string;
  text: string;
  sequence: number;
  started_at?: string | null;
  ended_at?: string | null;
  stt_confidence?: number | null;
  interrupted: boolean;
};

export type CallDetail = CallListItem & {
  utterances: Utterance[];
};

export type User = {
  id: string;
  phone: string;
  name?: string;
  tenant_id?: string;
};

export type Tenant = {
  id: string;
  name: string;
  system_prompt: string;
  greeting_text?: string;
  voice_id?: string;
  language: string;
  vip_names?: string[];
  marketing_email?: string;
  forward_number?: string;
  max_turn_timeout_ms?: number;
  plan: string;
  status: string;
};

export type TenantPhoneNumber = {
  id: string;
  twilio_number: string;
  is_primary: boolean;
};

export type AdminPhoneNumber = {
  id: string;
  twilio_number: string;
  twilio_sid?: string;
  forwarding_source?: string;
  is_primary: boolean;
  tenant_id?: string;
  tenant_name?: string;
  created_at: string;
};

export type AdminTenant = {
  id: string;
  name: string;
};

export type AdminTenantDetail = {
  id: string;
  name: string;
  system_prompt: string;
  greeting_text?: string;
  voice_id?: string;
  language: string;
  vip_names?: string[];
  marketing_email?: string;
  forward_number?: string;
  max_turn_timeout_ms?: number;
  plan: string;
  status: string;
  user_count: number;
  call_count: number;
  created_at: string;
  updated_at: string;
  // Billing fields
  stripe_customer_id?: string;
  stripe_subscription_id?: string;
  trial_ends_at?: string;
  current_period_start?: string;
  current_period_calls: number;
  time_saved_seconds: number;
  spam_calls_blocked: number;
  // Admin-only fields
  admin_notes?: string;
};

export type AdminUser = {
  id: string;
  phone: string;
  phone_verified: boolean;
  name?: string;
  role: string;
  last_login_at?: string;
  created_at: string;
};

export type CallEvent = {
  id: string;
  call_id: string;
  event_type: string;
  event_data: Record<string, unknown>;
  created_at: string;
};

export type AuthResponse = {
  token: string;
  expires_at: string;
  user: User;
};

export type OnboardingResponse = {
  tenant: Tenant;
  token: string;
  expires_at: string;
  phone_number?: TenantPhoneNumber;
};

export type CallStatus = {
  can_receive: boolean;
  reason: string; // "ok", "trial_expired", "limit_exceeded"
  calls_used: number;
  calls_limit: number; // -1 = unlimited
  trial_days_left?: number;
  trial_calls_left?: number;
};

export type CurrentUsage = {
  calls_count: number;
  minutes_used: number;
  time_saved_seconds: number;
  spam_calls_blocked: number;
  period_start: string;
  period_end: string;
};

export type BillingInfo = {
  plan: string;
  status: string;
  call_status: CallStatus;
  trial_ends_at?: string;
  total_time_saved: number; // seconds
  current_usage?: CurrentUsage;
};

export type TenantCostSummary = {
  tenant_id: string;
  period: string; // YYYY-MM format
  call_count: number;
  total_duration_seconds: number;
  twilio_cost_cents: number;
  stt_cost_cents: number;
  llm_cost_cents: number;
  tts_cost_cents: number;
  total_api_cost_cents: number;
  phone_number_count: number;
  phone_rental_cents: number;
  total_cost_cents: number;
  // Raw metrics for debugging
  total_stt_seconds: number;
  total_llm_input_tokens: number;
  total_llm_output_tokens: number;
  total_tts_characters: number;
};

export type Voice = {
  id: string;
  name: string;
  description: string;
  gender: "male" | "female";
};

export type GlobalConfigEntry = {
  key: string;
  value: string;
  description?: string;
  updated_at: string;
};

const API_BASE = import.meta.env.VITE_API_BASE_URL as string;

let authToken: string | null = localStorage.getItem("karen_token");

// Token refresh state to prevent concurrent refreshes
let isRefreshing = false;
let refreshPromise: Promise<AuthResponse> | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
  if (token) {
    localStorage.setItem("karen_token", token);
  } else {
    localStorage.removeItem("karen_token");
  }
}

export function getAuthToken(): string | null {
  return authToken;
}

export function isAuthenticated(): boolean {
  return authToken !== null;
}

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

async function http<T>(
  path: string,
  options?: RequestInit,
  _isRetry = false
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options?.headers as Record<string, string>),
  };

  if (authToken) {
    headers["Authorization"] = `Bearer ${authToken}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    if (res.status === 401 && !_isRetry) {
      // Try to refresh token before giving up (only on first attempt)
      const currentToken = authToken;
      if (currentToken && !isRefreshing) {
        isRefreshing = true;
        refreshPromise = (async () => {
          const refreshRes = await fetch(`${API_BASE}/auth/refresh`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ token: currentToken }),
          });
          if (!refreshRes.ok) throw new Error("refresh failed");
          return refreshRes.json() as Promise<AuthResponse>;
        })();

        try {
          const data = await refreshPromise;
          setAuthToken(data.token);
          isRefreshing = false;
          refreshPromise = null;
          // Retry original request with new token (mark as retry to prevent loop)
          return http<T>(path, options, true);
        } catch {
          isRefreshing = false;
          refreshPromise = null;
          setAuthToken(null);
          throw new ApiError(401, "unauthorized");
        }
      } else if (isRefreshing && refreshPromise) {
        // Another request is already refreshing, wait for it
        try {
          await refreshPromise;
          return http<T>(path, options, true);
        } catch {
          throw new ApiError(401, "unauthorized");
        }
      }
      setAuthToken(null);
      throw new ApiError(401, "unauthorized");
    }
    if (res.status === 401) {
      // Retry after refresh still got 401, clear token and fail
      setAuthToken(null);
      throw new ApiError(401, "unauthorized");
    }
    const text = await res.text();
    throw new ApiError(res.status, text || `HTTP ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

export const api = {
  // Calls
  listCalls: () => http<CallListItem[]>("/api/calls"),
  getCall: (providerCallId: string) =>
    http<CallDetail>(`/api/calls/${encodeURIComponent(providerCallId)}`),
  markCallViewed: (providerCallId: string) =>
    http<{ success: boolean }>(`/api/calls/${encodeURIComponent(providerCallId)}/viewed`, {
      method: "PATCH",
    }),
  markCallResolved: (providerCallId: string) =>
    http<{ success: boolean }>(`/api/calls/${encodeURIComponent(providerCallId)}/resolve`, {
      method: "PATCH",
    }),
  markCallUnresolved: (providerCallId: string) =>
    http<{ success: boolean }>(`/api/calls/${encodeURIComponent(providerCallId)}/resolve`, {
      method: "DELETE",
    }),
  getUnresolvedCount: () => http<{ count: number }>("/api/calls/unresolved-count"),

  // Auth
  sendCode: (phone: string) =>
    http<{ success: boolean }>("/auth/send-code", {
      method: "POST",
      body: JSON.stringify({ phone }),
    }),

  verifyCode: (phone: string, code: string) =>
    http<AuthResponse>("/auth/verify-code", {
      method: "POST",
      body: JSON.stringify({ phone, code }),
    }),

  refreshToken: (token: string) =>
    http<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ token }),
    }),

  logout: () =>
    http<void>("/auth/logout", {
      method: "POST",
    }),

  // User & Tenant
  getMe: () => http<{ user: User; tenant?: Tenant; is_admin?: boolean }>("/api/me"),

  getTenant: () => http<{ tenant: Tenant; phone_numbers: TenantPhoneNumber[] }>("/api/tenant"),

  // Billing
  getBilling: () => http<BillingInfo>("/api/billing"),

  createCheckout: (plan: "basic" | "pro", interval: "monthly" | "annual") =>
    http<{ checkout_url: string; session_id: string }>("/api/billing/checkout", {
      method: "POST",
      body: JSON.stringify({ plan, interval }),
    }),

  createPortal: () =>
    http<{ portal_url: string }>("/api/billing/portal", {
      method: "POST",
    }),

  updateTenant: (data: Partial<Tenant>) =>
    http<{ tenant: Tenant }>("/api/tenant", {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  // Voices
  getVoices: () => http<{ voices: Voice[] }>("/api/voices"),

  previewVoice: async (voiceId: string): Promise<Blob> => {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (authToken) {
      headers["Authorization"] = `Bearer ${authToken}`;
    }
    const res = await fetch(`${API_BASE}/api/voices/preview`, {
      method: "POST",
      headers,
      body: JSON.stringify({ voice_id: voiceId }),
    });
    if (!res.ok) {
      throw new Error(`Failed to preview voice: ${res.status}`);
    }
    return res.blob();
  },

  // Onboarding
  completeOnboarding: (name: string, greetingText: string) =>
    http<OnboardingResponse>("/api/onboarding/complete", {
      method: "POST",
      body: JSON.stringify({ name, greeting_text: greetingText }),
    }),

  // Admin
  adminListPhoneNumbers: () =>
    http<{ phone_numbers: AdminPhoneNumber[] }>("/admin/phone-numbers"),

  adminAddPhoneNumber: (twilioNumber: string, twilioSid?: string) =>
    http<TenantPhoneNumber>("/admin/phone-numbers", {
      method: "POST",
      body: JSON.stringify({ twilio_number: twilioNumber, twilio_sid: twilioSid }),
    }),

  adminDeletePhoneNumber: (id: string) =>
    http<{ success: boolean }>(`/admin/phone-numbers/${encodeURIComponent(id)}`, {
      method: "DELETE",
    }),

  adminUpdatePhoneNumber: (id: string, tenantId: string | null) =>
    http<{ success: boolean }>(`/admin/phone-numbers/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify({ tenant_id: tenantId }),
    }),

  adminListTenants: () => http<{ tenants: AdminTenant[] }>("/admin/tenants"),

  // Admin call logs
  adminListCalls: (limit?: number) =>
    http<{ calls: CallListItem[] }>(`/admin/calls${limit ? `?limit=${limit}` : ""}`),

  adminGetCallDetail: (providerCallId: string) =>
    http<CallDetail>(`/admin/calls/${encodeURIComponent(providerCallId)}`),

  adminGetCallEvents: (providerCallId: string) =>
    http<{ events: CallEvent[] }>(
      `/admin/calls/${encodeURIComponent(providerCallId)}/events`
    ),

  // Admin users dashboard
  adminListTenantsWithDetails: () =>
    http<{ tenants: AdminTenantDetail[] }>("/admin/tenants/details"),

  adminGetTenantUsers: (tenantId: string) =>
    http<{ users: AdminUser[] }>(`/admin/tenants/${encodeURIComponent(tenantId)}/users`),

  adminGetTenantCalls: (tenantId: string, limit = 20) =>
    http<{ calls: CallDetail[] }>(
      `/admin/tenants/${encodeURIComponent(tenantId)}/calls?limit=${limit}`
    ),

  adminUpdateTenant: (
    tenantId: string,
    data: {
      plan: string;
      status: string;
      max_turn_timeout_ms?: number;
      trial_ends_at?: string;
      current_period_calls?: number;
      admin_notes?: string;
    }
  ) =>
    http<{ success: boolean }>(`/admin/tenants/${encodeURIComponent(tenantId)}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  adminResetUserOnboarding: (userId: string) =>
    http<{ success: boolean; previous_tenant_id?: string }>(
      `/admin/users/${encodeURIComponent(userId)}/reset-onboarding`,
      { method: "PATCH" }
    ),

  adminDeleteTenant: (tenantId: string) =>
    http<{ success: boolean }>(`/admin/tenants/${encodeURIComponent(tenantId)}`, {
      method: "DELETE",
    }),

  adminGetTenantCosts: (tenantId: string, period?: string) =>
    http<TenantCostSummary>(
      `/admin/tenants/${encodeURIComponent(tenantId)}/costs${period ? `?period=${period}` : ""}`
    ),

  // Admin global config
  adminListGlobalConfig: () => http<{ config: GlobalConfigEntry[] }>("/admin/config"),

  adminUpdateGlobalConfig: (key: string, value: string) =>
    http<{ success: boolean }>(`/admin/config/${encodeURIComponent(key)}`, {
      method: "PATCH",
      body: JSON.stringify({ value }),
    }),
};
