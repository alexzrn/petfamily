package store

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/alexzrn/petfamily/apps/api/internal/geo"
	"github.com/google/uuid"
)

type NewWalk struct {
	UserID    string
	PetID     *string
	StartedAt time.Time
	EndedAt   time.Time
	Points    []Point
}

// CreateWalk пишет прогулку и её точки в одной транзакции. Дистанцию считает
// сервер, клиенту тут верить нельзя.
func (s *Store) CreateWalk(ctx context.Context, w NewWalk) (Walk, error) {
	id, err := newID()
	if err != nil {
		return Walk{}, err
	}
	dist := 0.0
	for i := 1; i < len(w.Points); i++ {
		dist += geo.Distance(w.Points[i-1].Lat, w.Points[i-1].Lng, w.Points[i].Lat, w.Points[i].Lng)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Walk{}, err
	}
	defer tx.Rollback()

	var petID any
	if w.PetID != nil {
		petID = binID(*w.PetID)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO walks (id, user_id, pet_id, started_at, ended_at, distance_m)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id[:], binID(w.UserID), petID, w.StartedAt.UTC(), w.EndedAt.UTC(), int(math.Round(dist))); err != nil {
		return Walk{}, err
	}
	for i, p := range w.Points {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO walk_points (walk_id, seq, lat, lng, at_utc) VALUES (?, ?, ?, ?, ?)`,
			id[:], i, p.Lat, p.Lng, p.AtUTC.UTC()); err != nil {
			return Walk{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Walk{}, err
	}
	return Walk{
		ID:        id.String(),
		UserID:    w.UserID,
		PetID:     w.PetID,
		StartedAt: w.StartedAt.UTC(),
		EndedAt:   w.EndedAt.UTC(),
		DistanceM: int(math.Round(dist)),
		Points:    w.Points,
	}, nil
}

func (s *Store) ListWalks(ctx context.Context, userID string, limit int) ([]Walk, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, pet_id, started_at, ended_at, distance_m, h3_cells
		   FROM walks WHERE user_id = ? ORDER BY started_at DESC LIMIT ?`, binID(userID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Walk
	for rows.Next() {
		var (
			w        Walk
			id, own  []byte
			pet      []byte
			petIsNil bool
		)
		if err := rows.Scan(&id, &own, &pet, &w.StartedAt, &w.EndedAt, &w.DistanceM, &w.H3Cells); err != nil {
			return nil, err
		}
		w.ID = uuid.UUID(id).String()
		w.UserID = uuid.UUID(own).String()
		if !petIsNil && len(pet) == 16 {
			s := uuid.UUID(pet).String()
			w.PetID = &s
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) WalkByID(ctx context.Context, userID, walkID string) (Walk, error) {
	var (
		w       Walk
		id, own []byte
		pet     []byte
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, pet_id, started_at, ended_at, distance_m, h3_cells
		   FROM walks WHERE id = ? AND user_id = ?`, binID(walkID), binID(userID)).
		Scan(&id, &own, &pet, &w.StartedAt, &w.EndedAt, &w.DistanceM, &w.H3Cells)
	if errors.Is(err, sql.ErrNoRows) {
		return Walk{}, ErrNotFound
	}
	if err != nil {
		return Walk{}, err
	}
	w.ID = uuid.UUID(id).String()
	w.UserID = uuid.UUID(own).String()
	if len(pet) == 16 {
		s := uuid.UUID(pet).String()
		w.PetID = &s
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT lat, lng, at_utc FROM walk_points WHERE walk_id = ? ORDER BY seq`, binID(walkID))
	if err != nil {
		return Walk{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Point
		if err := rows.Scan(&p.Lat, &p.Lng, &p.AtUTC); err != nil {
			return Walk{}, err
		}
		w.Points = append(w.Points, p)
	}
	return w, rows.Err()
}

// LiveWalks отдаёт прогулки, которые закончились не позже чем window назад и
// чья последняя точка попадает в окружность. Прямоугольник режет индекс,
// точное расстояние считается в Go.
func (s *Store) LiveWalks(ctx context.Context, lat, lng, meters float64, window time.Duration) ([]LiveWalk, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, meters)
	rows, err := s.db.QueryContext(ctx,
		`SELECT w.id, w.user_id, p.lat, p.lng, w.ended_at
		   FROM walks w
		   JOIN walk_points p ON p.walk_id = w.id
		   JOIN (SELECT walk_id, MAX(seq) AS seq FROM walk_points GROUP BY walk_id) m
		        ON m.walk_id = p.walk_id AND m.seq = p.seq
		  WHERE w.ended_at > (UTC_TIMESTAMP() - INTERVAL ? SECOND)
		    AND p.lat BETWEEN ? AND ? AND p.lng BETWEEN ? AND ?
		  ORDER BY w.ended_at DESC
		  LIMIT 200`,
		int(window.Seconds()), minLat, maxLat, minLng, maxLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LiveWalk
	for rows.Next() {
		var (
			lw  LiveWalk
			id  []byte
			own []byte
		)
		if err := rows.Scan(&id, &own, &lw.Lat, &lw.Lng, &lw.EndedAt); err != nil {
			return nil, err
		}
		d := geo.Distance(lat, lng, lw.Lat, lw.Lng)
		if d > meters {
			continue
		}
		lw.WalkID = uuid.UUID(id).String()
		lw.UserID = uuid.UUID(own).String()
		lw.Distance = d
		out = append(out, lw)
	}
	return out, rows.Err()
}
