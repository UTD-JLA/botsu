package activities

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type ActivityRepository struct {
	db *pgx.Conn
}

func NewActivityRepository(db *pgx.Conn) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *Activity) error {
	err := r.db.
		QueryRow(
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

func (r *ActivityRepository) PageByUserID(ctx context.Context, userID string, limit, offset int) (*UserActivityPage, error) {
	query := `
		SELECT activities.id,
			   user_id,
			   name,
			   primary_type,
			   media_type,
			   duration, date at time zone COALESCE(u.timezone, 'UTC'),
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

	rows, err := r.db.Query(ctx, query, userID, limit, offset)

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
	_, err := r.db.Exec(ctx, `
		UPDATE activities
		SET deleted_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *ActivityRepository) GetTopMembers(ctx context.Context, guildId string, limit int, start, end time.Time) ([]*MemberStats, error) {
	members := make([]*MemberStats, 0)

	rows, err := r.db.Query(ctx, `
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
