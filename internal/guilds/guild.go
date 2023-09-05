package guilds

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Guild struct {
	ID       string
	Timezone *string
}

func NewGuild(id string) *Guild {
	return &Guild{
		ID:       id,
		Timezone: nil,
	}
}

type GuildRepository struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

func NewGuildRepository(pool *pgxpool.Pool) *GuildRepository {
	return &GuildRepository{pool: pool, cache: sync.Map{}}
}

func (r *GuildRepository) Create(ctx context.Context, guild *Guild) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		`INSERT INTO guilds (id, timezone)
			VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE SET timezone = $2
			RETURNING id;`,
		guild.ID,
		guild.Timezone).
		Scan(&guild.ID)

	if err != nil {
		return err
	}

	r.cache.Store(guild.ID, guild)
	return nil
}

func (r *GuildRepository) FindByID(ctx context.Context, id string) (*Guild, error) {
	entry, ok := r.cache.Load(id)

	if ok {
		return entry.(*Guild), nil
	}

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	var guild Guild

	err = conn.QueryRow(ctx,
		`SELECT id, timezone
		FROM guilds
		WHERE id = $1;`,
		id).Scan(&guild.ID, &guild.Timezone)

	if err != nil {
		return nil, err
	}

	r.cache.Store(guild.ID, &guild)
	return &guild, nil
}

func (r *GuildRepository) FindOrCreate(ctx context.Context, id string) (*Guild, error) {
	entry, ok := r.cache.Load(id)

	if ok {
		return entry.(*Guild), nil
	}

	guild, err := r.FindByID(ctx, id)

	if err != nil {
		if err == pgx.ErrNoRows {
			guild = NewGuild(id)
			err = r.Create(ctx, guild)

			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return guild, nil
}

func (r *GuildRepository) SetGuildTimezone(ctx context.Context, guildId, timezone string) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`UPDATE guilds
		SET timezone = $2
		WHERE id = $1;`,
		guildId, timezone)

	if err != nil {
		return err
	}

	entry, ok := r.cache.Load(guildId)

	if ok {
		guild := entry.(*Guild)
		guild.Timezone = &timezone
	}

	return nil
}

func (r *GuildRepository) RemoveMembers(ctx context.Context, guildId string, userId []string) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`DELETE FROM guild_members
		WHERE guild_id = $1
		AND user_id = ANY($2);`,
		guildId, userId)

	return err
}
