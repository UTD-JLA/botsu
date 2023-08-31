package users

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, cache: sync.Map{}}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		`INSERT INTO users (id, active_guilds, timezone)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET active_guilds = $2, timezone = $3
			RETURNING id;`,
		user.ID,
		user.ActiveGuilds,
		user.Timezone).
		Scan(&user.ID)

	if err != nil {
		return err
	}

	r.cacheUser(user)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	var user User
	err = conn.QueryRow(ctx,
		`SELECT id, active_guilds, timezone
		FROM users
		WHERE id = $1;`,
		id).Scan(&user.ID, &user.ActiveGuilds, &user.Timezone)

	if err != nil {
		return nil, err
	}

	r.cacheUser(&user)

	return &user, nil
}

func (r *UserRepository) FindOrCreate(ctx context.Context, id string) (*User, error) {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			user = NewUser(id)
			err = r.Create(ctx, user)
			if err != nil {
				log.Printf("Error creating user: %v", err)
				return nil, err
			}
		} else {
			log.Printf("Error finding user: %v", err)
			return nil, err
		}
	}

	return user, nil
}

func (r *UserRepository) SetUserTimezone(ctx context.Context, userId, timezone string) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`INSERT INTO users (id, timezone)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET timezone = $2;`,
		userId, timezone)

	if err != nil {
		return err
	}

	user := r.getCachedUser(userId)
	user.Timezone = &timezone
	return nil
}

func (r *UserRepository) AppendActiveGuild(ctx context.Context, userId, guildId string) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`INSERT INTO users (id, active_guilds)
		VALUES ($1, ARRAY[$2])
		ON CONFLICT (id) DO UPDATE SET active_guilds = array_append(users.active_guilds, $2);`,
		userId, guildId)

	if err != nil {
		return err
	}

	user := r.getCachedUser(userId)
	user.ActiveGuilds = append(user.ActiveGuilds, guildId)
	return nil
}

func (r *UserRepository) RemoveActiveGuild(ctx context.Context, userId, guildId string) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`UPDATE users
		SET active_guilds = array_remove(users.active_guilds, $1)
		WHERE id = $2;`,
		guildId, userId)

	if err != nil {
		return err
	}

	user := r.getCachedUser(userId)

	if user != nil {
		for i, g := range user.ActiveGuilds {
			if g == guildId {
				user.ActiveGuilds = append(user.ActiveGuilds[:i], user.ActiveGuilds[i+1:]...)
				break
			}
		}
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`UPDATE users
		SET active_guilds = $1, timezone = $2
		WHERE id = $3;`,
		user.ActiveGuilds, user.Timezone, user.ID)

	if err != nil {
		return err
	}

	r.cacheUser(user)

	return nil
}

func (r *UserRepository) cacheUser(user *User) {
	r.cache.Store(user.ID, user)
}

func (r *UserRepository) getCachedUser(id string) *User {
	user, ok := r.cache.Load(id)

	if ok {
		return user.(*User)
	}

	return nil
}
