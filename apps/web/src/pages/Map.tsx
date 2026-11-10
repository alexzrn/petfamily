import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { api } from "../api";
import type { LiveWalk } from "../api";

const CENTER: [number, number] = [55.7558, 37.6173];

// /map: карта, позиционирование только по кнопке (с согласия пользователя),
// «кто гуляет рядом» по центру карты через GET /api/v1/live/walks.
export default function MapPage() {
  const ref = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const meRef = useRef<L.Marker | null>(null);
  const liveRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!ref.current || mapRef.current) return;
    const map = L.map(ref.current).setView(CENTER, 11);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
    }).addTo(map);
    liveRef.current = L.layerGroup().addTo(map);
    mapRef.current = map;
    return () => {
      map.remove();
      mapRef.current = null;
      meRef.current = null;
      liveRef.current = null;
    };
  }, []);

  function locate() {
    const map = mapRef.current;
    if (!map || !("geolocation" in navigator)) return;
    // Позиция запрашивается только по нажатию кнопки.
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const at: [number, number] = [pos.coords.latitude, pos.coords.longitude];
        map.setView(at, 14);
        if (!meRef.current) {
          meRef.current = L.marker(at).addTo(map);
        } else {
          meRef.current.setLatLng(at);
        }
        meRef.current.bindPopup("Вы здесь").openPopup();
      },
      () => {
        alert("Не удалось определить местоположение");
      },
    );
  }

  async function showLive() {
    const map = mapRef.current;
    const layer = liveRef.current;
    if (!map || !layer) return;
    const c = map.getCenter();
    let items: LiveWalk[];
    try {
      items = await api.liveWalks(c.lat, c.lng, 2000);
    } catch (err) {
      alert(err instanceof Error ? err.message : "Не удалось загрузить прогулки рядом");
      return;
    }
    layer.clearLayers();
    if (items.length === 0) {
      alert("Рядом никто не гуляет");
      return;
    }
    for (const w of items) {
      L.marker([w.lat, w.lng])
        .addTo(layer)
        .bindPopup(`Прогулка рядом, ${Math.round(w.distance_m)} м от вас`);
    }
  }

  return (
    <section>
      <h1>Карта</h1>
      <p style={{ display: "flex", gap: "0.5rem" }}>
        <button type="button" onClick={locate}>
          Моё местоположение
        </button>
        <button type="button" onClick={() => void showLive()}>
          Кто гуляет рядом
        </button>
      </p>
      <div ref={ref} style={{ height: "60vh", minHeight: 320 }} />
    </section>
  );
}
