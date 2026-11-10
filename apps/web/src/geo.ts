// Чистые гео-хелперы для карты и истории прогулок (покрыты unit-тестами).
export interface LatLng {
  lat: number;
  lng: number;
}

// Расстояние между двумя точками в метрах (haversine).
export function distanceM(a: LatLng, b: LatLng): number {
  const R = 6371000;
  const dLat = ((b.lat - a.lat) * Math.PI) / 180;
  const dLng = ((b.lng - a.lng) * Math.PI) / 180;
  const s1 = Math.sin(dLat / 2);
  const s2 = Math.sin(dLng / 2);
  const h =
    s1 * s1 + Math.cos((a.lat * Math.PI) / 180) * Math.cos((b.lat * Math.PI) / 180) * s2 * s2;
  return 2 * R * Math.asin(Math.sqrt(h));
}

// Длина трека в метрах.
export function trackLengthM(points: LatLng[]): number {
  let total = 0;
  for (let i = 1; i < points.length; i++) {
    total += distanceM(points[i - 1], points[i]);
  }
  return total;
}

export function formatDistance(m: number): string {
  if (m < 1000) return `${Math.round(m)} м`;
  return `${(m / 1000).toFixed(1)} км`;
}

export function formatDuration(startISO: string, endISO: string): string {
  const ms = Date.parse(endISO) - Date.parse(startISO);
  if (!Number.isFinite(ms) || ms < 0) return "—";
  const min = Math.round(ms / 60000);
  if (min < 60) return `${min} мин`;
  return `${Math.floor(min / 60)} ч ${min % 60} мин`;
}
