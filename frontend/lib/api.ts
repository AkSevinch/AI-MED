export type ApiError = { error: string };

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(path, {
    credentials: "include",
    ...options,
    headers: {
      ...(options.body && !(options.body instanceof FormData)
        ? { "Content-Type": "application/json" }
        : {}),
      ...(options.headers || {}),
    },
  });

  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = null;
    }
  }

  if (!res.ok) {
    const message =
      data && typeof data === "object" && "error" in data
        ? (data as ApiError).error
        : text && text.length <= 200
          ? text
          : `Ошибка запроса (${res.status})`;
    throw new Error(message);
  }
  return data as T;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body instanceof FormData ? body : JSON.stringify(body ?? {}),
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body ?? {}) }),
  del: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};

export type MeResponse = {
  authenticated: boolean;
  id?: number;
  email?: string;
  has_consent?: boolean;
};

export type SummaryProfile = {
  birth_date?: string;
  sex?: string;
  known_diagnoses?: string;
  current_medications?: string;
  cardiovascular_history?: string;
  weight_kg?: number | null;
  risk_factors?: string;
};

export type SummaryIndicator = {
  id: number;
  code: string;
  name: string;
  value: string;
  unit: string;
  reference_range: string;
  test_date: string;
  source: string;
  status: string;
  deviation: "below" | "within" | "above" | "unknown";
};

export type SummaryMeasurement = {
  kind: string;
  measured_at: string;
  systolic: number | null;
  diastolic: number | null;
  pulse: number | null;
  weight: number | null;
};

export type SummaryResponse = {
  profile: SummaryProfile;
  diagnoses: { name: string; comment: string }[];
  medications: {
    name: string;
    drug_class: string;
    frequency: string;
    status: string;
  }[];
  measurements: SummaryMeasurement[];
  indicators: SummaryIndicator[];
  abnormal: SummaryIndicator[];
  empty: boolean;
  updated_at: string;
};

export type OverviewResponse = {
  text: string;
  generated_at: string;
  empty: boolean;
};

export type ChatMessage = { role: "user" | "assistant"; content: string };
export type ChatResponse = { answer: string };
