CREATE TABLE users (
    id VARCHAR(20) PRIMARY KEY,
    active_guilds VARCHAR(20)[] NOT NULL DEFAULT '{}',
    timezone TEXT
);
