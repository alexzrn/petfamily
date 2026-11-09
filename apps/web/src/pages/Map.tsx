import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

// Заглушка карты: центр Москвы, один маркер. Настоящие точки подтянем из API позже.
const CENTER: [number, number] = [55.7558, 37.6173];

export default function MapPage() {
  const ref = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);

  useEffect(() => {
    if (!ref.current || mapRef.current) return;
    const map = L.map(ref.current).setView(CENTER, 11);
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
    }).addTo(map);
    L.marker(CENTER).addTo(map).bindPopup("Petfamily: скоро здесь будут точки");
    mapRef.current = map;
    return () => {
      map.remove();
      mapRef.current = null;
    };
  }, []);

  return (
    <section>
      <h1>Карта</h1>
      <div ref={ref} style={{ height: "60vh", minHeight: 320 }} />
    </section>
  );
}
