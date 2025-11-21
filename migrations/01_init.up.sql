CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        email VARCHAR(255) NOT NULL UNIQUE,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now()
);


CREATE TABLE events (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        event_date DATE NOT NULL,
        description TEXT NOT NULL,
        remind_at TIMESTAMP NULL,
        sent BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP NOT NULL DEFAULT now(),
        updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_user_date ON events (user_id, event_date);

CREATE TABLE events_archive (
        id UUID PRIMARY KEY,
        user_id UUID NOT NULL,
        event_date DATE NOT NULL,
        description TEXT NOT NULL,
        archived_at TIMESTAMP NOT NULL DEFAULT now(),
        original_created_at TIMESTAMP,
        original_updated_at TIMESTAMP
);

CREATE INDEX idx_events_archive_user ON events_archive (user_id);
