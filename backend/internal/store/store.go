package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	DB *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
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
func (s *Store) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.DB.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, query), args...)
}

func (s *Store) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.DB.QueryContext(ctx, sqlx.Rebind(sqlx.DOLLAR, query), args...)
}

func (s *Store) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.DB.QueryRowContext(ctx, sqlx.Rebind(sqlx.DOLLAR, query), args...)
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
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
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
		verified_drivers_only BOOLEAN NOT NULL DEFAULT FALSE,
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
		status TEXT NOT NULL DEFAULT 'active',
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
	);

	CREATE TABLE IF NOT EXISTS trips (
		id TEXT PRIMARY KEY,
		ride_request_id TEXT NOT NULL UNIQUE,
		rider_user_id TEXT NOT NULL,
		driver_user_id TEXT NOT NULL,
		published_ride_id TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'confirmed',
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (ride_request_id) REFERENCES ride_requests(id),
		FOREIGN KEY (rider_user_id) REFERENCES users(id),
		FOREIGN KEY (driver_user_id) REFERENCES users(id),
		FOREIGN KEY (published_ride_id) REFERENCES published_rides(id)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		body TEXT NOT NULL DEFAULT '',
		read INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS chat_conversations (
		id TEXT PRIMARY KEY,
		rider_user_id TEXT NOT NULL,
		driver_user_id TEXT NOT NULL,
		ride_request_id TEXT NOT NULL DEFAULT '',
		published_ride_id TEXT NOT NULL DEFAULT '',
		trip_id TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (rider_user_id) REFERENCES users(id),
		FOREIGN KEY (driver_user_id) REFERENCES users(id),
		FOREIGN KEY (ride_request_id) REFERENCES ride_requests(id),
		FOREIGN KEY (published_ride_id) REFERENCES published_rides(id),
		FOREIGN KEY (trip_id) REFERENCES trips(id)
	);

	CREATE TABLE IF NOT EXISTS chat_messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		sender_user_id TEXT NOT NULL,
		body TEXT NOT NULL,
		message_type TEXT NOT NULL DEFAULT 'text',
		image_data TEXT NOT NULL DEFAULT '',
		location_label TEXT NOT NULL DEFAULT '',
		location_lat REAL NOT NULL DEFAULT 0,
		location_lon REAL NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (conversation_id) REFERENCES chat_conversations(id),
		FOREIGN KEY (sender_user_id) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(strings.TrimSpace(migration)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}

	alterStatements := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_data TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS interests TEXT NOT NULL DEFAULT '[]';`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS rating_count INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS trips_completed INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_code TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_sent_at TIMESTAMP;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider TEXT NOT NULL DEFAULT 'password';`,

		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS max_budget REAL NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS minimum_rating REAL NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS verified_drivers_only BOOLEAN NOT NULL DEFAULT FALSE;`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS route_miles INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS route_duration TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_requests ADD COLUMN IF NOT EXISTS price_estimate TEXT NOT NULL DEFAULT '';`,

		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';`,
		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS vehicle_info TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS route_miles INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS route_duration TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE published_rides ADD COLUMN IF NOT EXISTS earnings_estimate REAL NOT NULL DEFAULT 0;`,

		`ALTER TABLE ride_request_actions ADD COLUMN IF NOT EXISTS published_ride_id TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ride_request_actions ADD COLUMN IF NOT EXISTS message TEXT NOT NULL DEFAULT '';`,

		`ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS message_type TEXT NOT NULL DEFAULT 'text';`,
		`ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS image_data TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS location_label TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS location_lat REAL NOT NULL DEFAULT 0;`,
		`ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS location_lon REAL NOT NULL DEFAULT 0;`,
	}

	for _, statement := range alterStatements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("apply alter statement: %w", err)
		}
	}

	if _, err := db.Exec(`UPDATE users SET rating = 0 WHERE rating_count = 0 AND rating = 5.0`); err != nil {
		return fmt.Errorf("normalize unrated users: %w", err)
	}

	return nil
}
