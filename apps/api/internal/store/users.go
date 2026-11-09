package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")
var ErrEmailTaken = errors.New("email taken")

func newID() (uuid.UUID, error) { return uuid.NewV7() }

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, displayName string) (User, error) {
	id, err := newID()
	if err != nil {
		return User{}, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, display_name) VALUES (?, ?, ?, ?)",
		id[:], email, passwordHash, displayName)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	return User{ID: id.String(), Email: email, DisplayName: displayName, PasswordHash: passwordHash}, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, display_name, city, created_at FROM users WHERE email = ?",
		strings.ToLower(strings.TrimSpace(email)))
	return scanUser(row)
}

func (s *Store) UserByID(ctx context.Context, id string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, display_name, city, created_at FROM users WHERE id = ?", binID(id))
	return scanUser(row)
}

func scanUser(row *sql.Row) (User, error) {
	var (
		u  User
		id []byte
	)
	err := row.Scan(&id, &u.Email, &u.PasswordHash, &u.DisplayName, &u.City, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.ID = uuid.UUID(id).String()
	return u, nil
}

func (s *Store) CreateSession(ctx context.Context, userID, token string, ttl time.Duration) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, binID(userID), time.Now().UTC().Add(ttl))
	return err
}

func (s *Store) UserByToken(ctx context.Context, token string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.password_hash, u.display_name, u.city, u.created_at
		   FROM sessions s JOIN users u ON u.id = s.user_id
		  WHERE s.token = ? AND s.expires_at > UTC_TIMESTAMP()`, token)
	return scanUser(row)
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func binID(id string) []byte {
	u, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	b := u[:]
	return b
}
