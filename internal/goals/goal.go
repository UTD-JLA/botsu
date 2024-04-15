package goals

import (
	"context"
	"slices"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/adhocore/gronx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Goal struct {
	ID              int64
	UserID          string
	Name            string
	ActivityType    *string
	MediaType       *string
	YoutubeChannels []string
	Target          time.Duration
	Current         time.Duration
	Cron            string
	DueAt           time.Time
	CreatedAt       time.Time
}

func (g *Goal) MatchesActivity(a *activities.Activity) bool {
	if g.ActivityType != nil && *g.ActivityType != a.PrimaryType {
		return false
	}

	if g.MediaType != nil && (a.MediaType == nil || *g.MediaType != *a.MediaType) {
		return false
	}

	if len(g.YoutubeChannels) == 0 {
		return true
	}

	meta, ok := a.Meta.(*activities.VideoInfo)

	if !ok {
		return false
	}

	if len(g.YoutubeChannels) > 0 && !slices.Contains(g.YoutubeChannels, meta.ChannelHandle) {
		return false
	}

	return true
}

func (g *Goal) IsDue() bool {
	return g.DueAt.Before(time.Now())
}

func (g *Goal) NextDueAt(location *time.Location) (t time.Time, err error) {
	if !g.IsDue() {
		return g.DueAt, nil
	}

	return gronx.NextTickAfter(g.Cron, time.Now().In(location), true)
}

func (g *Goal) PreviousDueAt(location *time.Location) (t time.Time, err error) {
	if g.IsDue() {
		return g.DueAt, nil
	}

	return gronx.PrevTickBefore(g.Cron, time.Now().In(location), false)
}

type GoalRepository struct {
	pool *pgxpool.Pool
}

func NewGoalRepository(pool *pgxpool.Pool) *GoalRepository {
	return &GoalRepository{pool}
}

func (r *GoalRepository) Create(ctx context.Context, g *Goal) (err error) {
	err = r.pool.QueryRow(
		ctx,
		`INSERT INTO goals (user_id, name, activity_type, media_type, youtube_channels, target, current, cron, due_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id`,
		g.UserID,
		g.Name,
		g.ActivityType,
		g.MediaType,
		g.YoutubeChannels,
		g.Target,
		g.Current,
		g.Cron,
		g.DueAt,
	).Scan(&g.ID)

	return
}

func (r *GoalRepository) FindByID(ctx context.Context, id int64) (goal *Goal, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, activity_type, media_type, youtube_channels, target, current, cron, due_at, created_at
		FROM goals		
		WHERE deleted_at IS NULL
		AND id = $1
	`, id)

	goal = &Goal{}

	err = row.Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Name,
		&goal.ActivityType,
		&goal.MediaType,
		&goal.YoutubeChannels,
		&goal.Target,
		&goal.Current,
		&goal.Cron,
		&goal.DueAt,
		&goal.CreatedAt,
	)

	return
}

func (r *GoalRepository) FindByUserID(ctx context.Context, userID string) (goals []*Goal, err error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, user_id, name, activity_type, media_type, youtube_channels, target, current, cron, due_at, created_at
		FROM goals
		WHERE deleted_at IS NULL
		AND user_id = $1`,
		userID,
	)

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		g := &Goal{}
		err = rows.Scan(
			&g.ID,
			&g.UserID,
			&g.Name,
			&g.ActivityType,
			&g.MediaType,
			&g.YoutubeChannels,
			&g.Target,
			&g.Current,
			&g.Cron,
			&g.DueAt,
			&g.CreatedAt,
		)

		if err != nil {
			return
		}

		goals = append(goals, g)
	}

	return
}

func (r *GoalRepository) BeginUpdateByUserID(ctx context.Context, userID string) (goals []*Goal, tx pgx.Tx, err error) {
	tx, err = r.pool.Begin(ctx)

	if err != nil {
		return
	}

	rows, err := tx.Query(
		ctx,
		`SELECT id, user_id, name, activity_type, media_type, youtube_channels, target, current, cron, due_at, created_at
		FROM goals
		WHERE user_id = $1
		AND DELETED_AT IS NULL
		FOR UPDATE`,
		userID,
	)

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		g := &Goal{}
		err = rows.Scan(
			&g.ID,
			&g.UserID,
			&g.Name,
			&g.ActivityType,
			&g.MediaType,
			&g.YoutubeChannels,
			&g.Target,
			&g.Current,
			&g.Cron,
			&g.DueAt,
			&g.CreatedAt,
		)

		if err != nil {
			return
		}

		goals = append(goals, g)
	}

	return
}

func (r *GoalRepository) UpdateTx(ctx context.Context, tx pgx.Tx, g *Goal) (err error) {
	_, err = tx.Exec(
		ctx,
		`UPDATE goals
		SET name = $1, activity_type = $2, media_type = $3, youtube_channels = $4, target = $5, current = $6, cron = $7, due_at = $8
		WHERE id = $9`,
		g.Name,
		g.ActivityType,
		g.MediaType,
		g.YoutubeChannels,
		g.Target,
		g.Current,
		g.Cron,
		g.DueAt,
		g.ID,
	)

	return
}

func (r *GoalRepository) FinishUpdate(ctx context.Context, tx pgx.Tx) (err error) {
	err = tx.Commit(ctx)
	return
}

func (r *GoalRepository) DeleteByID(ctx context.Context, id int64) (err error) {
	conn, err := r.pool.Acquire(ctx)

	if err != nil {
		return
	}

	defer conn.Release()

	_, err = conn.Exec(ctx, "UPDATE GOALS SET deleted_at = (NOW() AT TIME ZONE 'UTC') WHERE id = $1", id)

	return
}

type GoalService struct {
	repo *GoalRepository
	ts   *users.UserTimeService
}

func NewGoalService(repo *GoalRepository, ts *users.UserTimeService) *GoalService {
	return &GoalService{repo, ts}
}

func (s *GoalService) NextCron(ctx context.Context, g *Goal) (t time.Time, err error) {
	location, err := s.ts.GetTimeLocation(ctx, g.UserID, "")

	if err != nil {
		return
	}

	return g.NextDueAt(location)
}

func (s *GoalService) PreviousCron(ctx context.Context, g *Goal) (t time.Time, err error) {
	location, err := s.ts.GetTimeLocation(ctx, g.UserID, "")

	if err != nil {
		return
	}

	return g.PreviousDueAt(location)
}

func (s *GoalService) Create(ctx context.Context, g *Goal) (err error) {
	err = s.repo.Create(ctx, g)
	return
}

func (s *GoalService) Delete(ctx context.Context, goalID int64) (err error) {
	err = s.repo.DeleteByID(ctx, goalID)
	return
}

func (s *GoalService) FindByID(ctx context.Context, goalID int64) (goal *Goal, err error) {
	goal, err = s.repo.FindByID(ctx, goalID)
	return
}

func (s *GoalService) FindByUserID(ctx context.Context, userID string) (goals []*Goal, err error) {
	goals, err = s.repo.FindByUserID(ctx, userID)
	return
}

func (s *GoalService) CheckCompleted(ctx context.Context, a *activities.Activity) (completed []*Goal, err error) {
	goals, tx, err := s.repo.BeginUpdateByUserID(ctx, a.UserID)

	if err != nil {
		return
	}

	defer tx.Rollback(ctx) //nolint:errcheck

	for _, g := range goals {
		if g.IsDue() {
			g.Current = 0
			g.DueAt, err = s.NextCron(ctx, g)

			if err != nil {
				return
			}
		}

		var (
			nextDueAt time.Time
			prevDueAt time.Time
		)

		nextDueAt, err = s.NextCron(ctx, g)

		if err != nil {
			return
		}

		prevDueAt, err = s.PreviousCron(ctx, g)

		if err != nil {
			return
		}

		// We don't care about activities that are not within the range of the previous and next due time
		if a.Date.Before(prevDueAt) || a.Date.After(nextDueAt) {
			continue
		}

		// We don't care about activities that don't match the goal
		if !g.MatchesActivity(a) {
			continue
		}

		alreadyCompleted := g.Current >= g.Target

		g.Current += a.Duration

		if g.Current >= g.Target && !alreadyCompleted {
			completed = append(completed, g)
		}

		err = s.repo.UpdateTx(ctx, tx, g)

		if err != nil {
			return
		}
	}

	err = s.repo.FinishUpdate(ctx, tx)

	return
}
