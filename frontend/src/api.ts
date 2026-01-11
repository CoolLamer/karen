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
  plan: string;
  status: string;
  user_count: number;
  call_count: number;
  created_at: string;
  updated_at: string;
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

const API_BASE = import.meta.env.VITE_API_BASE_URL as string;

let authToken: string | null = localStorage.getItem("karen_token");

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

async function http<T>(path: string, options?: RequestInit): Promise<T> {
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
    if (res.status === 401) {
      setAuthToken(null);
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

  refreshToken: () =>
    http<AuthResponse>("/auth/refresh", {
      method: "POST",
    }),

  logout: () =>
    http<void>("/auth/logout", {
      method: "POST",
    }),

  // User & Tenant
  getMe: () => http<{ user: User; tenant?: Tenant; is_admin?: boolean }>("/api/me"),

  getTenant: () => http<{ tenant: Tenant; phone_numbers: TenantPhoneNumber[] }>("/api/tenant"),

  updateTenant: (data: Partial<Tenant>) =>
    http<{ tenant: Tenant }>("/api/tenant", {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  // Onboarding
  completeOnboarding: (name: string) =>
    http<OnboardingResponse>("/api/onboarding/complete", {
      method: "POST",
      body: JSON.stringify({ name }),
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

  adminUpdateTenantPlanStatus: (tenantId: string, plan: string, status: string) =>
    http<{ success: boolean }>(`/admin/tenants/${encodeURIComponent(tenantId)}`, {
      method: "PATCH",
      body: JSON.stringify({ plan, status }),
    }),
};
