CREATE TABLE goals (
    id BIGSERIAL PRIMARY KEY,
    user_id varchar(20) NOT NULL,
    name varchar(100) NOT NULL,
    activity_type activity_primary_type,
    media_type activity_media_type,
    youtube_channels TEXT[],
    cron TEXT NOT NULL,
    target BIGINT NOT NULL,
    current BIGINT NOT NULL DEFAULT 0,
    due_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    deleted_at TIMESTAMP
);
