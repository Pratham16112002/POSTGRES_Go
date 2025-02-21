
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
	id bigserial PRIMARY KEY,
	username varchar(255) UNIQUE not NULL,
	email citext UNIQUE NOT NULL,
	password bytea NOT NULL,
	created_at timestamp(0
   ) WITH time zone NOT NULL DEFAULT now()
);