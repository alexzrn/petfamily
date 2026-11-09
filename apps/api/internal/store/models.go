package store

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	City         string    `json:"city"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Pet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Species   string    `json:"species"`
	Sex       string    `json:"sex"`
	BirthYear *int      `json:"birth_year"`
	Breed     string    `json:"breed"`
	Traits    string    `json:"traits"`
	CreatedAt time.Time `json:"created_at"`
}

type Point struct {
	Lat   float64   `json:"lat"`
	Lng   float64   `json:"lng"`
	AtUTC time.Time `json:"at"`
}

type Walk struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PetID     *string   `json:"pet_id"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	DistanceM int       `json:"distance_m"`
	H3Cells   int       `json:"h3_cells"`
	Points    []Point   `json:"points,omitempty"`
}

type Meetup struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Title    string    `json:"title"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	StartsAt time.Time `json:"starts_at"`
}

type LiveWalk struct {
	WalkID   string    `json:"walk_id"`
	UserID   string    `json:"user_id"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	Distance float64   `json:"distance_m"`
	EndedAt  time.Time `json:"ended_at"`
}
