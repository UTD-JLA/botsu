package activities

import (
	"context"
	"time"

	orderedmap "github.com/UTD-JLA/botsu/pkg/ordered_map"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberStats struct {
	UserID        string
	TotalDuration time.Duration
}

type UserActivityPage struct {
	Activities []*Activity
	PageCount  int
	Page       int
}

type ActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *Activity) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		`INSERT INTO activities (user_id, name, primary_type, media_type, duration, date, meta)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id;`,
		activity.UserID,
		activity.Name,
		activity.PrimaryType,
		activity.MediaType,
		activity.Duration,
		activity.Date,
		activity.Meta).
		Scan(&activity.ID)

	return err
}

func (r *ActivityRepository) GetTotalByUserIDGroupByVideoChannel(ctx context.Context, userID string, start, end time.Time) (orderedmap.Map[time.Duration], error) {
	query := `
		SELECT
			COALESCE(SUM(duration), 0) AS total_duration,
			meta->>'channel_handle' AS channel_handle
		FROM activities
		WHERE user_id = $1
		AND media_type = 'video'
		AND meta->>'platform' = 'youtube'
		AND meta->>'channel_handle' IS NOT NULL
		AND date >= $2
		AND date <= $3
		AND deleted_at IS NULL
		GROUP BY meta->>'channel_handle'
		ORDER BY total_duration DESC
	`

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, query, userID, start, end)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	channels := orderedmap.New[time.Duration]()

	for rows.Next() {
		var channel string
		var duration time.Duration

		if err := rows.Scan(&duration, &channel); err != nil {
			return nil, err
		}

		channels.Set(channel, duration)
	}

	return channels, nil
}

// Returns map of day (YYYY-MM-DD) to total duration
// filling in missing days with 0 (string formatted according to user's timezone)
func (r *ActivityRepository) GetTotalByUserIDGroupedByDay(ctx context.Context, userID string, start, end time.Time) (orderedmap.Map[time.Duration], error) {
	// day should be truncated to a string `YYYY-MM-DD` in the user's timezone
	query := `
		SELECT
			to_char(date_series.day, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(duration), 0) AS total_duration
		FROM (
			SELECT
				generate_series(
					$2::date,
					$3::date,
					interval '1 day'
				) AS day
		) AS date_series
		LEFT JOIN users u ON u.id = $1
		LEFT JOIN activities
			ON date_series.day = date_trunc('day', activities.date at time zone COALESCE(u.timezone, 'UTC'))
			AND activities.user_id = $1
			AND activities.deleted_at IS NULL
		GROUP BY day
		ORDER BY day ASC
	`

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, query, userID, start, end)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	durations := orderedmap.NewWithCapacity[time.Duration](int(end.Sub(start).Hours()/24) + 1)

	for rows.Next() {
		var date string
		var duration time.Duration

		if err := rows.Scan(&date, &duration); err != nil {
			return nil, err
		}

		durations.Set(date, duration)
	}

	return durations, nil
}

func (r *ActivityRepository) GetLatestByUserID(ctx context.Context, userID string) (*Activity, error) {
	query := `
		SELECT activities.id,
			   user_id,	
			   name,
			   primary_type,
			   media_type,
			   duration,
			   date at time zone COALESCE(u.timezone, 'UTC'),
			   created_at,
			   deleted_at,
			   meta
		FROM activities
		INNER JOIN users u ON activities.user_id = u.id
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	var activity Activity

	err = conn.QueryRow(ctx, query, userID).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.Name,
		&activity.PrimaryType,
		&activity.MediaType,
		&activity.Duration,
		&activity.Date,
		&activity.CreatedAt,
		&activity.DeletedAt,
		&activity.Meta,
	)

	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *ActivityRepository) GetByID(ctx context.Context, id uint64) (*Activity, error) {
	query := `
		SELECT activities.id,
			   user_id,
			   name,
			   primary_type,
			   media_type,
			   duration,
			   date at time zone COALESCE(u.timezone, 'UTC'),
			   created_at,
			   deleted_at,
			   meta
		FROM activities
		INNER JOIN users u ON activities.user_id = u.id
		WHERE activities.id = $1
		AND deleted_at IS NULL
	`

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	var activity Activity

	err = conn.QueryRow(ctx, query, id).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.Name,
		&activity.PrimaryType,
		&activity.MediaType,
		&activity.Duration,
		&activity.Date,
		&activity.CreatedAt,
		&activity.DeletedAt,
		&activity.Meta,
	)

	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *ActivityRepository) PageByUserID(ctx context.Context, userID string, limit, offset int) (*UserActivityPage, error) {
	query := `
		SELECT activities.id,
			   user_id,
			   name,
			   primary_type,
			   media_type,
			   duration,
			   date at time zone COALESCE(u.timezone, 'UTC'),
			   created_at,
			   deleted_at,
			   meta,
			   CEIL(COUNT(*) OVER() / $2::float) AS page_count,
			   CEIL($3::float / $2::float) + 1 AS page
		FROM activities
		INNER JOIN users u ON activities.user_id = u.id
		WHERE user_id = $1
		AND deleted_at IS NULL
		ORDER BY date DESC
		LIMIT $2
		OFFSET $3
	`

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, query, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	page := &UserActivityPage{
		Activities: make([]*Activity, 0),
	}

	for rows.Next() {
		var activity Activity
		if err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Name,
			&activity.PrimaryType,
			&activity.MediaType,
			&activity.Duration,
			&activity.Date,
			&activity.CreatedAt,
			&activity.DeletedAt,
			&activity.Meta,
			&page.PageCount,
			&page.Page,
		); err != nil {
			return nil, err
		}
		page.Activities = append(page.Activities, &activity)
	}
	return page, nil
}

func (r *ActivityRepository) DeleteById(ctx context.Context, id uint64) error {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx, `
		UPDATE activities
		SET deleted_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

func (r *ActivityRepository) GetTopMembers(ctx context.Context, guildId string, limit int, start, end time.Time) ([]*MemberStats, error) {
	members := make([]*MemberStats, 0)

	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, `
		SELECT u.id, SUM(a.duration) AS total_duration
		FROM users u
		INNER JOIN activities a ON u.id = a.user_id
		WHERE $1 = ANY(u.active_guilds)
		AND a.deleted_at IS NULL
		AND a.date >= $2
		AND a.date <= $3
		GROUP BY u.id
		ORDER BY total_duration DESC
		LIMIT $4
	`, guildId, start, end, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var member MemberStats
		if err := rows.Scan(
			&member.UserID,
			&member.TotalDuration,
		); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}

	return members, nil
}
