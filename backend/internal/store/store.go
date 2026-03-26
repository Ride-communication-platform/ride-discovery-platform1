package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func New(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func runMigrations(db *sql.DB) error {
	migration := `
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  avatar_data TEXT NOT NULL DEFAULT '',
  interests TEXT NOT NULL DEFAULT '[]',
  password_hash TEXT NOT NULL,
  rating REAL NOT NULL DEFAULT 0,
  rating_count INTEGER NOT NULL DEFAULT 0,
  trips_completed INTEGER NOT NULL DEFAULT 0,
  email_verified INTEGER NOT NULL DEFAULT 0,
  verification_code TEXT NOT NULL DEFAULT '',
  password_reset_code TEXT NOT NULL DEFAULT '',
  password_reset_sent_at TIMESTAMP,
  auth_provider TEXT NOT NULL DEFAULT 'password',
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS ride_requests (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  from_label TEXT NOT NULL,
  from_lat REAL NOT NULL,
  from_lon REAL NOT NULL,
  to_label TEXT NOT NULL,
  to_lat REAL NOT NULL,
  to_lon REAL NOT NULL,
  ride_date TEXT NOT NULL,
  ride_time TEXT NOT NULL,
  flexibility TEXT NOT NULL,
  passengers INTEGER NOT NULL,
  luggage TEXT NOT NULL,
  max_budget REAL NOT NULL DEFAULT 0,
  ride_type TEXT NOT NULL,
  vehicle_preference TEXT NOT NULL,
  minimum_rating REAL NOT NULL DEFAULT 0,
  verified_drivers_only INTEGER NOT NULL DEFAULT 0,
  notes TEXT NOT NULL DEFAULT '',
  route_miles INTEGER NOT NULL DEFAULT 0,
  route_duration TEXT NOT NULL DEFAULT '',
  price_estimate TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS published_rides (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  from_label TEXT NOT NULL,
  from_lat REAL NOT NULL,
  from_lon REAL NOT NULL,
  to_label TEXT NOT NULL,
  to_lat REAL NOT NULL,
  to_lon REAL NOT NULL,
  ride_date TEXT NOT NULL,
  ride_time TEXT NOT NULL,
  flexibility TEXT NOT NULL,
  available_seats INTEGER NOT NULL,
  total_seats INTEGER NOT NULL,
  price_per_seat REAL NOT NULL,
  vehicle_type TEXT NOT NULL,
  luggage_allowed TEXT NOT NULL,
  ride_type TEXT NOT NULL,
  vehicle_info TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  route_miles INTEGER NOT NULL DEFAULT 0,
  route_duration TEXT NOT NULL DEFAULT '',
  earnings_estimate REAL NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS ride_request_actions (
  id TEXT PRIMARY KEY,
  ride_request_id TEXT NOT NULL,
  driver_user_id TEXT NOT NULL,
  published_ride_id TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL,
  message TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (ride_request_id) REFERENCES ride_requests(id),
  FOREIGN KEY (driver_user_id) REFERENCES users(id),
  FOREIGN KEY (published_ride_id) REFERENCES published_rides(id)
);`

	if _, err := db.Exec(strings.TrimSpace(migration)); err != nil {
		return fmt.Errorf("run users migration: %w", err)
	}

	for _, statement := range []string{
		`ALTER TABLE users ADD COLUMN avatar_data TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN interests TEXT NOT NULL DEFAULT '[]';`,
		`ALTER TABLE users ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN trips_completed INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN verification_code TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN password_reset_code TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN password_reset_sent_at TIMESTAMP;`,
		`ALTER TABLE users ADD COLUMN auth_provider TEXT NOT NULL DEFAULT 'password';`,
	} {
		if _, err := db.Exec(statement); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return fmt.Errorf("apply users alter: %w", err)
		}
	}

	if _, err := db.Exec(`UPDATE users SET rating = 0 WHERE rating_count = 0 AND rating = 5.0`); err != nil {
		return fmt.Errorf("normalize unrated users: %w", err)
	}

	for _, statement := range []string{
		`ALTER TABLE ride_requests ADD COLUMN max_budget REAL NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN minimum_rating REAL NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN verified_drivers_only INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN notes TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_requests ADD COLUMN route_miles INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN route_duration TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_requests ADD COLUMN price_estimate TEXT NOT NULL DEFAULT '';`,
	} {
		if _, err := db.Exec(statement); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") && !strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return fmt.Errorf("apply ride requests alter: %w", err)
		}
	}

	for _, statement := range []string{
		`ALTER TABLE published_rides ADD COLUMN vehicle_info TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN notes TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN route_miles INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE published_rides ADD COLUMN route_duration TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN earnings_estimate REAL NOT NULL DEFAULT 0;`,
	} {
		if _, err := db.Exec(statement); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") && !strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return fmt.Errorf("apply published rides alter: %w", err)
		}
	}

	for _, statement := range []string{
		`ALTER TABLE ride_request_actions ADD COLUMN published_ride_id TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_request_actions ADD COLUMN message TEXT NOT NULL DEFAULT '';`,
	} {
		if _, err := db.Exec(statement); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") && !strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return fmt.Errorf("apply ride request actions alter: %w", err)
		}
	}

	return nil
}
