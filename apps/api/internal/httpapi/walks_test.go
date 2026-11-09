package httpapi

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func iso(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func walkBody(started time.Time, d time.Duration, points ...map[string]any) map[string]any {
	return map[string]any{
		"started_at": iso(started),
		"ended_at":   iso(started.Add(d)),
		"points":     points,
	}
}

func pt(lat, lng float64, at time.Time) map[string]any {
	return map[string]any{"lat": lat, "lng": lng, "at": iso(at)}
}

func TestWalkValidation(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("walk@example.com")
	now := time.Now().Add(-time.Hour)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"одна точка", walkBody(now, time.Minute, pt(12.9, 100.9, now))},
		{"широта вне диапазона", walkBody(now, time.Minute, pt(120, 100.9, now), pt(12.9, 100.9, now))},
		{"дольше 12 часов", walkBody(now, 13*time.Hour, pt(12.9, 100.9, now), pt(12.91, 100.91, now))},
		{"конец раньше начала", map[string]any{
			"started_at": iso(now), "ended_at": iso(now.Add(-time.Minute)),
			"points": []map[string]any{pt(12.9, 100.9, now), pt(12.91, 100.91, now)},
		}},
	}
	for _, tc := range cases {
		code, body := c.do(http.MethodPost, "/api/v1/walks", tc.body)
		if code != http.StatusBadRequest {
			t.Fatalf("%s: код %d, ожидали 400 (тело %v)", tc.name, code, body)
		}
	}
}

func TestWalkCreateGetList(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("walker@example.com")
	start := time.Now().Add(-30 * time.Minute)

	body := walkBody(start, 20*time.Minute,
		pt(12.9236, 100.8824, start),
		pt(12.9240, 100.8830, start.Add(10*time.Minute)),
		pt(12.9244, 100.8836, start.Add(20*time.Minute)),
	)
	code, created := c.do(http.MethodPost, "/api/v1/walks", body)
	if code != http.StatusCreated {
		t.Fatalf("создание прогулки: код %d, тело %v", code, created)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("нет id в ответе: %v", created)
	}
	if d, _ := created["distance_m"].(float64); d <= 0 {
		t.Fatalf("дистанция не посчитана: %v", created)
	}

	code, one := c.do(http.MethodGet, "/api/v1/walks/"+id, nil)
	if code != http.StatusOK {
		t.Fatalf("чтение прогулки: код %d, тело %v", code, one)
	}
	points, _ := one["points"].([]any)
	if len(points) != 3 {
		t.Fatalf("точек %d, ожидали 3", len(points))
	}

	code, list := c.do(http.MethodGet, "/api/v1/walks?limit=10", nil)
	if code != http.StatusOK {
		t.Fatalf("список прогулок: код %d", code)
	}
	items, _ := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("прогулок в списке %d, ожидали 1", len(items))
	}

	code, _ = c.do(http.MethodGet, "/api/v1/walks?limit=0", nil)
	if code != http.StatusBadRequest {
		t.Fatalf("limit=0 дал %d, ожидали 400", code)
	}
}

func TestForeignWalkIsHidden(t *testing.T) {
	srv := newTestServer(t)
	owner := newClient(t, srv)
	owner.register("owner2@example.com")
	start := time.Now().Add(-10 * time.Minute)
	_, created := owner.do(http.MethodPost, "/api/v1/walks",
		walkBody(start, 5*time.Minute, pt(12.9, 100.9, start), pt(12.91, 100.91, start.Add(time.Minute))))
	id, _ := created["id"].(string)

	stranger := newClient(t, srv)
	stranger.register("stranger@example.com")
	code, body := stranger.do(http.MethodGet, "/api/v1/walks/"+id, nil)
	if code != http.StatusNotFound {
		t.Fatalf("чужая прогулка: код %d, ожидали 404 (тело %v)", code, body)
	}
}

func TestLiveWalksAndMeetups(t *testing.T) {
	srv := newTestServer(t)
	c := newClient(t, srv)
	c.register("live@example.com")

	// прогулка, которая только что закончилась рядом с точкой поиска
	start := time.Now().Add(-15 * time.Minute)
	code, body := c.do(http.MethodPost, "/api/v1/walks",
		walkBody(start, 14*time.Minute,
			pt(12.9236, 100.8824, start),
			pt(12.9238, 100.8826, start.Add(14*time.Minute))))
	if code != http.StatusCreated {
		t.Fatalf("прогулка: код %d, тело %v", code, body)
	}

	code, live := c.do(http.MethodGet, "/api/v1/live/walks?lat=12.9237&lng=100.8825&meters=500", nil)
	if code != http.StatusOK {
		t.Fatalf("живые прогулки: код %d, тело %v", code, live)
	}
	items, _ := live["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("живых прогулок %d, ожидали 1: %v", len(items), live)
	}

	// далёкая точка поиска ничего не находит
	code, far := c.do(http.MethodGet, "/api/v1/live/walks?lat=55.75&lng=37.61&meters=500", nil)
	if code != http.StatusOK {
		t.Fatalf("живые прогулки (далеко): код %d", code)
	}
	farItems, _ := far["items"].([]any)
	if len(farItems) != 0 {
		t.Fatalf("в Москве нашлись паттайские прогулки: %v", far)
	}

	// встреча рядом
	starts := time.Now().Add(24 * time.Hour)
	code, meetup := c.do(http.MethodPost, "/api/v1/meetups", map[string]any{
		"title": "Утренний выгул у пляжа", "lat": 12.9240, "lng": 100.8830, "starts_at": iso(starts),
	})
	if code != http.StatusCreated {
		t.Fatalf("встреча: код %d, тело %v", code, meetup)
	}

	code, nearby := c.do(http.MethodGet, "/api/v1/meetups?lat=12.9236&lng=100.8824&radius_m=2000", nil)
	if code != http.StatusOK {
		t.Fatalf("встречи: код %d", code)
	}
	nearItems, _ := nearby["items"].([]any)
	if len(nearItems) != 1 {
		t.Fatalf("встреч рядом %d, ожидали 1: %v", len(nearItems), nearby)
	}

	// встреча в прошлом не принимается
	past := time.Now().Add(-2 * time.Hour)
	code, body = c.do(http.MethodPost, "/api/v1/meetups", map[string]any{
		"title": "Прошлое", "lat": 12.9240, "lng": 100.8830, "starts_at": iso(past),
	})
	if code != http.StatusBadRequest {
		t.Fatalf("встреча в прошлом: код %d, ожидали 400 (тело %v)", code, body)
	}

	_ = fmt.Sprint() // держим fmt в импортах для отладочных сообщений
}
