package users

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	err := r.db.
		QueryRow(
			ctx,
			`INSERT INTO users (id, active_guilds, timezone)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET active_guilds = $2, timezone = $3
			RETURNING id;`,
			user.ID,
			user.ActiveGuilds,
			user.Timezone).
		Scan(&user.ID)

	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		`SELECT id, active_guilds, timezone
		FROM users
		WHERE id = $1;`,
		id).Scan(&user.ID, &user.ActiveGuilds, &user.Timezone)

	if err != nil {
		return nil, err
	}

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
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, timezone)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET timezone = $2;`,
		userId, timezone)

	return err
}

func (r *UserRepository) AppendActiveGuild(ctx context.Context, userId, guildId string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, active_guilds)
		VALUES ($1, ARRAY[$2])
		ON CONFLICT (id) DO UPDATE SET active_guilds = array_append(users.active_guilds, $2);`,
		userId, guildId)

	return err
}

func (r *UserRepository) RemoveActiveGuild(ctx context.Context, userId, guildId string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users
		SET active_guilds = array_remove(users.active_guilds, $1)
		WHERE id = $2;`,
		guildId, userId)

	return err
}
