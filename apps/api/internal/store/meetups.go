package store

import (
	"context"
	"time"

	"github.com/alexzrn/petfamily/apps/api/internal/geo"
	"github.com/google/uuid"
)

func (s *Store) CreateMeetup(ctx context.Context, m Meetup) (Meetup, error) {
	id, err := newID()
	if err != nil {
		return Meetup{}, err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO meetups (id, user_id, title, lat, lng, starts_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id[:], binID(m.UserID), m.Title, m.Lat, m.Lng, m.StartsAt.UTC())
	if err != nil {
		return Meetup{}, err
	}
	m.ID = id.String()
	return m, nil
}

func (s *Store) ListMeetups(ctx context.Context, lat, lng, radiusM float64, limit int) ([]Meetup, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusM)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, title, lat, lng, starts_at
		   FROM meetups
		  WHERE starts_at > UTC_TIMESTAMP()
		    AND lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
		  ORDER BY starts_at
		  LIMIT ?`,
		minLat, maxLat, minLng, maxLng, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Meetup
	for rows.Next() {
		var (
			m        Meetup
			id, own  []byte
			startsAt time.Time
		)
		if err := rows.Scan(&id, &own, &m.Title, &m.Lat, &m.Lng, &startsAt); err != nil {
			return nil, err
		}
		if geo.Distance(lat, lng, m.Lat, m.Lng) > radiusM {
			continue
		}
		m.ID = uuid.UUID(id).String()
		m.UserID = uuid.UUID(own).String()
		m.StartsAt = startsAt
		out = append(out, m)
	}
	return out, rows.Err()
}
