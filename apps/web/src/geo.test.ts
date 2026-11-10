import { describe, expect, it } from "vitest";
import { distanceM, formatDistance, formatDuration, trackLengthM } from "./geo";

describe("distanceM", () => {
  it("ноль для одной точки", () => {
    expect(distanceM({ lat: 55.75, lng: 37.61 }, { lat: 55.75, lng: 37.61 })).toBe(0);
  });
  it("Москва–Питер около 630 км", () => {
    const m = distanceM({ lat: 55.7558, lng: 37.6173 }, { lat: 59.9343, lng: 30.3351 });
    expect(m).toBeGreaterThan(600000);
    expect(m).toBeLessThan(670000);
  });
});

describe("trackLengthM", () => {
  it("суммирует отрезки", () => {
    const pts = [
      { lat: 55.75, lng: 37.61 },
      { lat: 55.751, lng: 37.61 },
      { lat: 55.751, lng: 37.611 },
    ];
    expect(trackLengthM(pts)).toBeCloseTo(
      distanceM(pts[0], pts[1]) + distanceM(pts[1], pts[2]),
      6,
    );
  });
  it("пустой трек — ноль", () => {
    expect(trackLengthM([])).toBe(0);
  });
});

describe("formatDistance", () => {
  it("метры и километры", () => {
    expect(formatDistance(850)).toBe("850 м");
    expect(formatDistance(1500)).toBe("1.5 км");
  });
});

describe("formatDuration", () => {
  it("минуты и часы", () => {
    expect(formatDuration("2026-01-01T10:00:00Z", "2026-01-01T10:25:00Z")).toBe("25 мин");
    expect(formatDuration("2026-01-01T10:00:00Z", "2026-01-01T11:30:00Z")).toBe("1 ч 30 мин");
  });
  it("битый интервал — прочерк", () => {
    expect(formatDuration("2026-01-01T11:00:00Z", "2026-01-01T10:00:00Z")).toBe("—");
  });
});
