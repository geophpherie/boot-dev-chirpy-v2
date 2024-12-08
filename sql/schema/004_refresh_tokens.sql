-- +goose Up
CREATE TABLE refresh_tokens (
	token TEXT PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE cascade,
	expires_at TIMESTAMP,
	revoked_at TIMESTAMP
	);
-- +goose Down
DROP TABLE refresh_tokens;
