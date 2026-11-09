// Базовый URL API: import.meta.env.VITE_API_BASE,
// по умолчанию https://api.petfamily.beliberde.net
export const API_BASE =
  import.meta.env.VITE_API_BASE ?? "https://api.petfamily.beliberde.net";

export async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${path}`);
  }
  return res.json() as Promise<T>;
}
