package users

import (
	"context"
	"errors"
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
		`INSERT INTO users (
			   id,
			   timezone,
			   vn_reading_speed,
			   book_reading_speed,
			   manga_reading_speed,
			   daily_goal
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
			    timezone = $2,
				vn_reading_speed = $3,
				book_reading_speed = $4,
				manga_reading_speed = $5,
				daily_goal = $6
			RETURNING id;`,
		user.ID,
		user.Timezone,
		user.VisualNovelReadingSpeed,
		user.BookReadingSpeed,
		user.MangaReadingSpeed,
		user.DailyGoal,
	).Scan(&user.ID)

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

	cached := r.getCachedUser(id)

	if cached != nil {
		return cached, nil
	}

	var user User
	err = conn.QueryRow(ctx,
		`SELECT id,
       		timezone,
       		vn_reading_speed,
       		book_reading_speed,
       		manga_reading_speed,
       		daily_goal
		FROM users
		WHERE id = $1;`, id).Scan(
		&user.ID,
		&user.Timezone,
		&user.VisualNovelReadingSpeed,
		&user.BookReadingSpeed,
		&user.MangaReadingSpeed,
		&user.DailyGoal,
	)

	if err != nil {
		return nil, err
	}

	r.cacheUser(&user)

	return &user, nil
}

func (r *UserRepository) FindOrCreate(ctx context.Context, id string) (*User, error) {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (r *UserRepository) SetVisualNovelReadingSpeed(ctx context.Context, userID string, speed float32) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	query := `
		INSERT INTO users (id, vn_reading_speed)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET vn_reading_speed = $2;
	`

	if _, err = conn.Exec(ctx, query, userID, speed); err != nil {
		return err
	}

	if user := r.getCachedUser(userID); user != nil {
		user.VisualNovelReadingSpeed = speed
	}

	return nil
}

func (r *UserRepository) SetBookReadingSpeed(ctx context.Context, userID string, speed float32) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	query := `
		INSERT INTO users (id, book_reading_speed)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET book_reading_speed = $2;
	`

	if _, err = conn.Exec(ctx, query, userID, speed); err != nil {
		return err
	}

	if user := r.getCachedUser(userID); user != nil {
		user.BookReadingSpeed = speed
	}

	return nil
}

func (r *UserRepository) SetMangaReadingSpeed(ctx context.Context, userID string, speed float32) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	query := `
		INSERT INTO users (id, manga_reading_speed)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET manga_reading_speed = $2;
	`

	if _, err = conn.Exec(ctx, query, userID, speed); err != nil {
		return err
	}

	if user := r.getCachedUser(userID); user != nil {
		user.MangaReadingSpeed = speed
	}

	return nil
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

	if user != nil {
		user.Timezone = &timezone
	}

	return nil
}

func (r *UserRepository) SetDailyGoal(ctx context.Context, userID string, goal int) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	query := `
		INSERT INTO users (id, daily_goal)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET daily_goal = $2;
	`

	if _, err = conn.Exec(ctx, query, userID, goal); err != nil {
		return err
	}

	if user := r.getCachedUser(userID); user != nil {
		user.DailyGoal = goal
	}

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
