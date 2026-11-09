package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *Store) CreatePet(ctx context.Context, p Pet) (Pet, error) {
	id, err := newID()
	if err != nil {
		return Pet{}, err
	}
	var birth *int
	if p.BirthYear != nil {
		birth = p.BirthYear
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO pets (id, user_id, name, species, sex, birth_year, breed, traits)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id[:], binID(p.UserID), p.Name, p.Species, p.Sex, birth, p.Breed, p.Traits)
	if err != nil {
		return Pet{}, err
	}
	p.ID = id.String()
	p.CreatedAt = time.Now().UTC()
	return p, nil
}

func (s *Store) ListPets(ctx context.Context, userID string) ([]Pet, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, species, sex, birth_year, breed, traits, created_at
		   FROM pets WHERE user_id = ? ORDER BY created_at`, binID(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Pet
	for rows.Next() {
		var (
			p        Pet
			id, owner []byte
			birth    sql.NullInt64
		)
		if err := rows.Scan(&id, &owner, &p.Name, &p.Species, &p.Sex, &birth, &p.Breed, &p.Traits, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.ID = uuid.UUID(id).String()
		p.UserID = uuid.UUID(owner).String()
		if birth.Valid {
			v := int(birth.Int64)
			p.BirthYear = &v
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) PetOwned(ctx context.Context, userID, petID string) (Pet, error) {
	var (
		p        Pet
		id, owner []byte
		birth    sql.NullInt64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, species, sex, birth_year, breed, traits, created_at
		   FROM pets WHERE id = ? AND user_id = ?`, binID(petID), binID(userID)).
		Scan(&id, &owner, &p.Name, &p.Species, &p.Sex, &birth, &p.Breed, &p.Traits, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Pet{}, ErrNotFound
	}
	if err != nil {
		return Pet{}, err
	}
	p.ID = uuid.UUID(id).String()
	p.UserID = uuid.UUID(owner).String()
	return p, nil
}
