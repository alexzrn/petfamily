package main

import (
	"log"
	"net/http"
	"os"

	"github.com/alexzrn/petfamily/apps/api/internal/httpapi"
	"github.com/alexzrn/petfamily/apps/api/internal/store"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN не задан")
	}
	db, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("база недоступна: %v", err)
	}
	defer db.Close()
	if err := store.Migrate(db); err != nil {
		log.Fatalf("миграции: %v", err)
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: buildRouter(httpapi.NewServer(store.NewStore(db)).Handler()),
	}
	log.Printf("api listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
