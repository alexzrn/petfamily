package main

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed templates/*.html
var templateFS embed.FS

type app struct {
	db       *sql.DB
	tmpl     map[string]*template.Template
	user     string
	password string
	base     string
}

func main() {
	dsn := mustEnv("MYSQL_DSN")
	user := mustEnv("ADMIN_USER")
	password := mustEnv("ADMIN_PASSWORD")
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("dsn: %v", err)
	}
	db.SetMaxOpenConns(10)
	if err := db.Ping(); err != nil {
		log.Fatalf("база недоступна: %v", err)
	}
	defer db.Close()

	// base — префикс пути, если админка живёт не в корне домена,
	// например /admin за общим nginx. Шаблоны строят ссылки от него.
	base := os.Getenv("ADMIN_BASE_PATH")
	a := &app{db: db, tmpl: parseTemplates(), user: user, password: password, base: base}
	log.Printf("admin listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.routes()))
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s не задан", k)
	}
	return v
}

func parseTemplates() map[string]*template.Template {
	out := map[string]*template.Template{}
	for _, name := range []string{"index", "users", "walks", "meetups"} {
		out[name] = template.Must(template.ParseFS(templateFS,
			"templates/layout.html", "templates/"+name+".html"))
	}
	return out
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /{$}", a.guard(a.handleIndex))
	mux.HandleFunc("GET /users", a.guard(a.handleUsers))
	mux.HandleFunc("POST /users/{id}/block", a.guard(a.handleBlock))
	mux.HandleFunc("POST /users/{id}/unblock", a.guard(a.handleUnblock))
	mux.HandleFunc("GET /walks", a.guard(a.handleWalks))
	mux.HandleFunc("GET /meetups", a.guard(a.handleMeetups))
	return mux
}

// guard пускает только по Basic Auth: больше ничего в админке нет.
func (a *app) guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != a.user || p != a.password {
			w.Header().Set("WWW-Authenticate", `Basic realm="petfamily admin"`)
			http.Error(w, "нужен логин и пароль", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

type counters struct {
	Users   int
	Pets    int
	Walks   int
	Meetups int
	Blocked int
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	var c counters
	err := a.db.QueryRowContext(r.Context(), `
		SELECT (SELECT COUNT(*) FROM users),
		       (SELECT COUNT(*) FROM pets),
		       (SELECT COUNT(*) FROM walks),
		       (SELECT COUNT(*) FROM meetups),
		       (SELECT COUNT(*) FROM user_blocks)`).
		Scan(&c.Users, &c.Pets, &c.Walks, &c.Meetups, &c.Blocked)
	if err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	a.render(w, "index", map[string]any{"Counters": c})
}

type userRow struct {
	ID          string
	Email       string
	DisplayName string
	Pets        int
	Walks       int
	CreatedAt   time.Time
	Blocked     bool
}

func (a *app) handleUsers(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 50, 200)
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT u.id, u.email, u.display_name, u.created_at,
		       (SELECT COUNT(*) FROM pets p WHERE p.user_id = u.id),
		       (SELECT COUNT(*) FROM walks wl WHERE wl.user_id = u.id),
		       EXISTS(SELECT 1 FROM user_blocks b WHERE b.user_id = u.id)
		  FROM users u
		 ORDER BY u.created_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []userRow
	for rows.Next() {
		var (
			u  userRow
			id []byte
		)
		if err := rows.Scan(&id, &u.Email, &u.DisplayName, &u.CreatedAt, &u.Pets, &u.Walks, &u.Blocked); err != nil {
			http.Error(w, "чтение: "+err.Error(), http.StatusInternalServerError)
			return
		}
		u.ID = formatID(id)
		items = append(items, u)
	}
	a.render(w, "users", map[string]any{"Items": items})
}

func (a *app) handleBlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := a.db.ExecContext(r.Context(),
		"INSERT IGNORE INTO user_blocks (user_id, reason) VALUES (UNHEX(REPLACE(?, '-', '')), 'вручную из админки')", id); err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, a.base+"/users", http.StatusSeeOther)
}

func (a *app) handleUnblock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := a.db.ExecContext(r.Context(),
		"DELETE FROM user_blocks WHERE user_id = UNHEX(REPLACE(?, '-', ''))", id); err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, a.base+"/users", http.StatusSeeOther)
}

type walkRow struct {
	ID        string
	Email     string
	StartedAt time.Time
	EndedAt   time.Time
	DistanceM int
	Points    int
}

func (a *app) handleWalks(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 50, 200)
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT wl.id, u.email, wl.started_at, wl.ended_at, wl.distance_m,
		       (SELECT COUNT(*) FROM walk_points wp WHERE wp.walk_id = wl.id)
		  FROM walks wl JOIN users u ON u.id = wl.user_id
		 ORDER BY wl.started_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []walkRow
	for rows.Next() {
		var (
			wr walkRow
			id []byte
		)
		if err := rows.Scan(&id, &wr.Email, &wr.StartedAt, &wr.EndedAt, &wr.DistanceM, &wr.Points); err != nil {
			http.Error(w, "чтение: "+err.Error(), http.StatusInternalServerError)
			return
		}
		wr.ID = formatID(id)
		items = append(items, wr)
	}
	a.render(w, "walks", map[string]any{"Items": items})
}

type meetupRow struct {
	ID       string
	Title    string
	Email    string
	Lat      float64
	Lng      float64
	StartsAt time.Time
}

func (a *app) handleMeetups(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT m.id, m.title, u.email, m.lat, m.lng, m.starts_at
		  FROM meetups m JOIN users u ON u.id = m.user_id
		 ORDER BY m.starts_at DESC
		 LIMIT 200`)
	if err != nil {
		http.Error(w, "ошибка базы: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []meetupRow
	for rows.Next() {
		var (
			m  meetupRow
			id []byte
		)
		if err := rows.Scan(&id, &m.Title, &m.Email, &m.Lat, &m.Lng, &m.StartsAt); err != nil {
			http.Error(w, "чтение: "+err.Error(), http.StatusInternalServerError)
			return
		}
		m.ID = formatID(id)
		items = append(items, m)
	}
	a.render(w, "meetups", map[string]any{"Items": items})
}

func (a *app) render(w http.ResponseWriter, name string, data map[string]any) {
	t, ok := a.tmpl[name]
	if !ok {
		http.Error(w, "нет шаблона", http.StatusInternalServerError)
		return
	}
	data["Base"] = a.base
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("шаблон %s: %v", name, err)
	}
}

func parseLimit(r *http.Request, def, max int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > max {
		return def
	}
	return n
}

// formatID превращает BINARY(16) в uuid-строку, не завязываясь на пакет uuid.
func formatID(b []byte) string {
	if len(b) != 16 {
		return ""
	}
	h := make([]byte, 32)
	const hexdigits = "0123456789abcdef"
	for i, v := range b {
		h[i*2] = hexdigits[v>>4]
		h[i*2+1] = hexdigits[v&0x0f]
	}
	return string(h[0:8]) + "-" + string(h[8:12]) + "-" + string(h[12:16]) + "-" + string(h[16:20]) + "-" + string(h[20:32])
}
