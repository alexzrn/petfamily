package geo

import "math"

const earthRadiusM = 6371000.0

// Distance считает расстояние по большому кругу между двумя точками в метрах.
func Distance(lat1, lng1, lat2, lng2 float64) float64 {
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dp := (lat2 - lat1) * math.Pi / 180
	dl := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return 2 * earthRadiusM * math.Asin(math.Min(1, math.Sqrt(a)))
}

// BoundingBox возвращает прямоугольник вокруг точки, в который заведомо влезает
// окружность радиуса radiusM. Точная проверка всё равно делается через Distance.
func BoundingBox(lat, lng, radiusM float64) (minLat, maxLat, minLng, maxLng float64) {
	dLat := radiusM / earthRadiusM * 180 / math.Pi
	cos := math.Cos(lat * math.Pi / 180)
	if math.Abs(cos) < 1e-9 {
		cos = 1e-9
	}
	dLng := radiusM / (earthRadiusM * cos) * 180 / math.Pi
	return lat - dLat, lat + dLat, lng - dLng, lng + dLng
}
