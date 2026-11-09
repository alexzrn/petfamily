package store

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Store — тонкая обёртка над пулом соединений: никакой бизнес-логики, только SQL.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Open открывает пул соединений и проверяет, что база отвечает.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
