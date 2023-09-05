DROP INDEX activities_date_index;
DROP INDEX activities_primary_type_index;
DROP INDEX activities_media_type_index;
DROP INDEX activities_duration_index;
DROP INDEX activities_meta_index;
DROP INDEX activities_user_id_index;

DROP TRIGGER create_guild_member_on_activity_insert ON activities;
DROP FUNCTION create_guild_member_on_activity_insert;

DROP TABLE activities;
DROP TABLE guild_members;
DROP TABLE guilds;
DROP TABLE users;

DROP TYPE activity_primary_type;
DROP TYPE activity_media_type;


