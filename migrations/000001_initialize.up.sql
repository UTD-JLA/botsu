CREATE TABLE guilds (
    id VARCHAR(20) PRIMARY KEY,
    timezone TEXT
);

CREATE TABLE users (
    id VARCHAR(20) PRIMARY KEY,
    timezone TEXT,
    vn_reading_speed REAL NOT NULL DEFAULT 0,
    book_reading_speed REAL NOT NULL DEFAULT 0,
    manga_reading_speed REAL NOT NULL DEFAULT 0,
    daily_goal INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE guild_members (
    PRIMARY KEY (guild_id, user_id),
    guild_id VARCHAR(20) NOT NULL REFERENCES guilds(id),
    user_id VARCHAR(20) NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    last_seen_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
);

CREATE TYPE activity_primary_type as ENUM('reading', 'listening');
CREATE TYPE activity_media_type as ENUM('book', 'anime', 'manga', 'video',  'visual_novel');

CREATE TABLE activities (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL REFERENCES users(id),
    guild_id VARCHAR(20) REFERENCES guilds(id),
    name TEXT NOT NULL,
    primary_type activity_primary_type NOT NULL,
    media_type activity_media_type,
    duration BIGINT NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    deleted_at TIMESTAMP,
    imported_at TIMESTAMP,
    meta JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX activities_date_index ON activities (date);
CREATE INDEX activities_user_id_index ON activities (user_id);
CREATE INDEX activities_primary_type_index ON activities (primary_type);
CREATE INDEX activities_media_type_index ON activities (media_type);
CREATE INDEX activities_duration_index ON activities (duration);
CREATE INDEX activities_meta_index ON activities USING GIN (meta);

CREATE FUNCTION create_guild_member_on_activity_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users (id) VALUES (NEW.user_id)
    ON CONFLICT DO NOTHING;

    IF NEW.guild_id IS NOT NULL THEN
        INSERT INTO guilds (id) VALUES (NEW.guild_id)
        ON CONFLICT DO NOTHING;

        INSERT INTO guild_members (guild_id, user_id)
        VALUES (NEW.guild_id, NEW.user_id)
        ON CONFLICT (guild_id, user_id) DO UPDATE
        SET last_seen_at = (NOW() AT TIME ZONE 'utc');
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_guild_member_on_activity_insert
BEFORE INSERT ON activities
FOR EACH ROW EXECUTE PROCEDURE create_guild_member_on_activity_insert();

