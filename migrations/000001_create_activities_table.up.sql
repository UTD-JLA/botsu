CREATE TYPE activity_primary_type as ENUM('reading', 'listening');
CREATE TYPE activity_media_type as ENUM('book', 'anime', 'manga', 'video',  'visual_novel');

CREATE TABLE activities (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(20),
    name TEXT NOT NULL,
    primary_type activity_primary_type,
    media_type activity_media_type,
    duration BIGINT NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    meta JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX activities_date_index ON activities (date);
CREATE INDEX activities_user_id_index ON activities (user_id);
CREATE INDEX activities_primary_type_index ON activities (primary_type);
CREATE INDEX activities_media_type_index ON activities (media_type);
CREATE INDEX activities_duration_index ON activities (duration);
CREATE INDEX activities_meta_index ON activities USING GIN (meta);
