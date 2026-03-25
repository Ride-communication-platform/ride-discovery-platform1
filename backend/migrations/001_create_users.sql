CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  rating REAL DEFAULT 5.0,
  email_verified INTEGER NOT NULL DEFAULT 0,
  verification_code TEXT NOT NULL DEFAULT '',
  password_reset_code TEXT NOT NULL DEFAULT '',
  password_reset_sent_at TIMESTAMP,
  auth_provider TEXT NOT NULL DEFAULT 'password',
  created_at TIMESTAMP NOT NULL
);
