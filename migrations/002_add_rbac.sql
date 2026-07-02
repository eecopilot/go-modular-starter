CREATE TABLE IF NOT EXISTS roles (
	name text PRIMARY KEY,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS permissions (
	name text PRIMARY KEY,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS role_permissions (
	role_name text NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
	permission_name text NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (role_name, permission_name)
);

CREATE TABLE IF NOT EXISTS user_roles (
	user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role_name text NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (user_id, role_name)
);

CREATE INDEX IF NOT EXISTS user_roles_role_name_idx
	ON user_roles (role_name);

INSERT INTO roles (name)
VALUES
	('super_admin'),
	('user_admin'),
	('viewer'),
	('user')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name)
VALUES
	('user:read'),
	('user:list'),
	('user:update:self'),
	('user:update:any'),
	('user:disable'),
	('user:enable'),
	('user:reset_password'),
	('role:assign')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
VALUES
	('super_admin', 'user:read'),
	('super_admin', 'user:list'),
	('super_admin', 'user:update:self'),
	('super_admin', 'user:update:any'),
	('super_admin', 'user:disable'),
	('super_admin', 'user:enable'),
	('super_admin', 'user:reset_password'),
	('super_admin', 'role:assign'),
	('user_admin', 'user:read'),
	('user_admin', 'user:list'),
	('user_admin', 'user:update:self'),
	('user_admin', 'user:update:any'),
	('user_admin', 'user:disable'),
	('user_admin', 'user:enable'),
	('user_admin', 'user:reset_password'),
	('viewer', 'user:read'),
	('viewer', 'user:list'),
	('user', 'user:update:self')
ON CONFLICT (role_name, permission_name) DO NOTHING;
