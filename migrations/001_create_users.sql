CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	email text NULL,
	username text NULL,
	full_name text NOT NULL DEFAULT '',
	password_hash text NOT NULL,
	status text NOT NULL DEFAULT 'enabled',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT users_email_or_username_check CHECK (
		NULLIF(email, '') IS NOT NULL OR NULLIF(username, '') IS NOT NULL
	),
	CONSTRAINT users_status_check CHECK (status IN ('enabled', 'disabled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx
	ON users (lower(email))
	WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_username_unique_idx
	ON users (lower(username))
	WHERE username IS NOT NULL;

CREATE INDEX IF NOT EXISTS users_created_at_idx
	ON users (created_at DESC, id DESC);
