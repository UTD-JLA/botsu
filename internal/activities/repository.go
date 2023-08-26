package activities

import (
	"context"

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

func (r *ActivityRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*Activity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, primary_type, media_type, duration, date, meta
		FROM activities
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY date DESC
		LIMIT $2
		OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
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
		); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}
	return activities, nil
}

func (r *ActivityRepository) DeleteById(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE activities
		SET deleted_at = NOW()
		WHERE id = $1
	`, id)
	return err
}
