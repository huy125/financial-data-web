CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(100) NOT NULL UNIQUE,
	hash VARCHAR(256) NOT NULL UNIQUE,
	CHECK(username <> '' AND hash <> '')
);
