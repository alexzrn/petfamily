import { useCallback, useEffect, useRef, useState } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { api } from "../api";
import type { Pet, Walk, WalkPoint } from "../api";
import { formatDistance, formatDuration, trackLengthM } from "../geo";

const CENTER: [number, number] = [55.7558, 37.6173];

// /walks: запись трека (старт/стоп, геолокация только во время записи),
// сохранение через POST /api/v1/walks, история через GET /api/v1/walks.
export default function Walks() {
  const ref = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const lineRef = useRef<L.Polyline | null>(null);

  const [pets, setPets] = useState<Pet[]>([]);
  const [petId, setPetId] = useState("");
  const [recording, setRecording] = useState(false);
  const [points, setPoints] = useState<WalkPoint[]>([]);
  const [history, setHistory] = useState<Walk[]>([]);
  const [selected, setSelected] = useState<Walk | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const watchRef = useRef<number | null>(null);
  const startedRef = useRef<string>("");

  useEffect(() => {
    if (!ref.current || mapRef.current) return;
    const map = L.map(ref.current).setView(CENTER, 12);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
    }).addTo(map);
    mapRef.current = map;
    return () => {
      map.remove();
      mapRef.current = null;
    };
  }, []);

  const drawTrack = useCallback((pts: WalkPoint[]) => {
    const map = mapRef.current;
    if (!map) return;
    if (lineRef.current) {
      lineRef.current.remove();
      lineRef.current = null;
    }
    if (pts.length === 0) return;
    const latlngs = pts.map((p) => [p.lat, p.lng]) as [number, number][];
    lineRef.current = L.polyline(latlngs, { color: "#2a7ae2" }).addTo(map);
    map.fitBounds(lineRef.current.getBounds().pad(0.2));
  }, []);

  const loadHistory = useCallback(async () => {
    try {
      setHistory(await api.listWalks(20));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить историю");
    }
  }, []);

  useEffect(() => {
    void api
      .listPets()
      .then(setPets)
      .catch(() => {});
    void loadHistory();
  }, [loadHistory]);

  function start() {
    if (!("geolocation" in navigator)) {
      setError("Браузер не поддерживает геолокацию");
      return;
    }
    setError("");
    setSelected(null);
    setPoints([]);
    startedRef.current = new Date().toISOString();
    watchRef.current = navigator.geolocation.watchPosition(
      (pos) => {
        const p: WalkPoint = {
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
          at: new Date().toISOString(),
        };
        setPoints((prev) => {
          const next = [...prev, p];
          drawTrack(next);
          return next;
        });
      },
      () => setError("Геолокация недоступна — проверьте разрешение браузера"),
      { enableHighAccuracy: true },
    );
    setRecording(true);
  }

  function stop() {
    if (watchRef.current !== null) {
      navigator.geolocation.clearWatch(watchRef.current);
      watchRef.current = null;
    }
    setRecording(false);
  }

  useEffect(
    () => () => {
      if (watchRef.current !== null) navigator.geolocation.clearWatch(watchRef.current);
    },
    [],
  );

  async function save() {
    if (points.length < 2) {
      setError("Нужно минимум две точки трека");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const walk = await api.createWalk({
        pet_id: petId || undefined,
        started_at: startedRef.current,
        ended_at: new Date().toISOString(),
        points,
      });
      setPoints([]);
      setHistory((prev) => [walk, ...prev]);
      setSelected(walk);
      drawTrack(walk.points ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить прогулку");
    } finally {
      setBusy(false);
    }
  }

  async function openWalk(id: string) {
    setError("");
    try {
      const walk = await api.getWalk(id);
      setSelected(walk);
      drawTrack(walk.points ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось открыть прогулку");
    }
  }

  return (
    <section>
      <h1>Прогулки</h1>
      {error ? (
        <p role="alert" style={{ color: "crimson" }}>
          {error}
        </p>
      ) : null}

      <h2>Запись</h2>
      <p style={{ display: "flex", gap: "0.5rem", alignItems: "center", flexWrap: "wrap" }}>
        <select value={petId} onChange={(e) => setPetId(e.target.value)} aria-label="Питомец">
          <option value="">Без питомца</option>
          {pets.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
        {!recording ? (
          <button type="button" onClick={start}>
            Начать прогулку
          </button>
        ) : (
          <button type="button" onClick={stop}>
            Остановить
          </button>
        )}
        {!recording && points.length > 0 ? (
          <button type="button" onClick={() => void save()} disabled={busy}>
            {busy ? "Сохраняем…" : `Сохранить (${points.length} точек)`}
          </button>
        ) : null}
      </p>
      {recording ? (
        <p>
          Идёт запись… точек: {points.length}, длина:{" "}
          {Math.round(trackLengthM(points.map((p) => ({ lat: p.lat, lng: p.lng }))))} м
        </p>
      ) : null}
      <div ref={ref} style={{ height: "50vh", minHeight: 280 }} />

      <h2>История</h2>
      {history.length === 0 ? (
        <p>Пока нет сохранённых прогулок.</p>
      ) : (
        <ul style={{ padding: 0, listStyle: "none", display: "grid", gap: "0.5rem" }}>
          {history.map((w) => (
            <li
              key={w.id}
              style={{ border: "1px solid #ddd", borderRadius: 8, padding: "0.5rem 0.75rem" }}
            >
              <button
                type="button"
                onClick={() => void openWalk(w.id)}
                style={{ background: "none", border: 0, padding: 0, cursor: "pointer", textAlign: "left" }}
              >
                <strong>{new Date(w.started_at).toLocaleString("ru-RU")}</strong> ·{" "}
                {formatDistance(w.distance_m)} · {formatDuration(w.started_at, w.ended_at)}
              </button>
              {selected?.id === w.id && selected.points ? (
                <p style={{ margin: "0.25rem 0 0" }}>Точек в маршруте: {selected.points.length}</p>
              ) : null}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
