export type ScreeningResult = {
  legitimacy_label: string;
  legitimacy_confidence: number;
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

const API_BASE = import.meta.env.VITE_API_BASE_URL as string;

async function http<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return (await res.json()) as T;
}

export const api = {
  listCalls: () => http<CallListItem[]>("/api/calls"),
  getCall: (providerCallId: string) => http<CallDetail>(`/api/calls/${encodeURIComponent(providerCallId)}`),
};


