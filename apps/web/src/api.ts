// Тонкий клиент к API Petfamily (по умолчанию https://api.petfamily.beliberde.net).
// Токен сессии — в localStorage, подставляется в Authorization: Bearer.
export const API_BASE =
  import.meta.env.VITE_API_BASE ?? "https://api.petfamily.beliberde.net";

const TOKEN_KEY = "petfamily_token";

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setToken(token: string | null): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token);
    else localStorage.removeItem(TOKEN_KEY);
  } catch {
    /* приватный режим — сессия не сохранится */
  }
}

export interface ApiUser {
  id: string;
  email: string;
  display_name: string;
}

export interface AuthPayload {
  token: string;
  user: ApiUser;
}

export interface Pet {
  id: string;
  name: string;
  species: string;
  sex: string;
  birth_year: number | null;
  breed: string;
  traits: string;
}

export interface NewPet {
  name: string;
  species?: string;
  sex?: string;
  birth_year?: number | null;
  breed?: string;
  traits?: string;
}

function humanError(status: number, body: unknown): string {
  if (
    body !== null &&
    typeof body === "object" &&
    "error" in body &&
    typeof (body as { error: unknown }).error === "string"
  ) {
    return (body as { error: string }).error;
  }
  if (status === 401) return "нужно войти заново";
  if (status >= 500) return "сервер недоступен, попробуйте позже";
  return `запрос не удался (код ${status})`;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  } catch {
    throw new Error("нет связи с сервером, проверьте интернет");
  }
  let body: unknown = null;
  try {
    body = await res.json();
  } catch {
    /* пустой ответ */
  }
  if (!res.ok) throw new Error(humanError(res.status, body));
  return body as T;
}

export interface WalkPoint {
  lat: number;
  lng: number;
  at: string;
}

export interface NewWalk {
  pet_id?: string;
  started_at: string;
  ended_at: string;
  points: WalkPoint[];
}

export interface Walk {
  id: string;
  user_id: string;
  pet_id: string | null;
  started_at: string;
  ended_at: string;
  distance_m: number;
  h3_cells: number;
  points?: WalkPoint[];
}

export interface LiveWalk {
  walk_id: string;
  user_id: string;
  lat: number;
  lng: number;
  distance_m: number;
  ended_at: string;
}

export const api = {
  register: (email: string, password: string, display_name: string) =>
    request<AuthPayload>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, display_name }),
    }),
  login: (email: string, password: string) =>
    request<AuthPayload>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<{ ok: boolean }>("/api/v1/auth/logout", { method: "POST" }),
  me: () => request<ApiUser>("/api/v1/me"),
  listPets: async (): Promise<Pet[]> => {
    const data = await request<Pet[] | { items: Pet[] }>("/api/v1/pets");
    return Array.isArray(data) ? data : data.items;
  },
  createPet: (pet: NewPet) =>
    request<Pet>("/api/v1/pets", { method: "POST", body: JSON.stringify(pet) }),
  createWalk: (walk: NewWalk) =>
    request<Walk>("/api/v1/walks", { method: "POST", body: JSON.stringify(walk) }),
  listWalks: async (limit = 20): Promise<Walk[]> => {
    const data = await request<{ items: Walk[] }>(`/api/v1/walks?limit=${limit}`);
    return data.items;
  },
  getWalk: (id: string) => request<Walk>(`/api/v1/walks/${encodeURIComponent(id)}`),
  liveWalks: async (lat: number, lng: number, meters = 2000): Promise<LiveWalk[]> => {
    const data = await request<{ items: LiveWalk[] }>(
      `/api/v1/live/walks?lat=${lat}&lng=${lng}&meters=${meters}`,
    );
    return data.items;
  },
};
