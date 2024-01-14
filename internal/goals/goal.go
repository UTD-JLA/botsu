package goals

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/hashicorp/cronexpr"
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

	expr, err := cronexpr.Parse(g.Cron)

	if err != nil {
		return
	}

	localTime := time.Now().In(location)
	return expr.Next(localTime), nil
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

func (r *GoalRepository) FindByUserID(ctx context.Context, userID string) (goals []*Goal, err error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, user_id, name, activity_type, media_type, youtube_channels, target, current, cron, due_at, created_at
		FROM goals
		WHERE user_id = $1`,
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

func (s *GoalService) Create(ctx context.Context, g *Goal) (err error) {
	err = s.repo.Create(ctx, g)
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

	defer tx.Rollback(ctx)

	for _, g := range goals {
		fmt.Printf("goal: %+v\n", g)

		if g.IsDue() {
			g.Current = 0
			g.DueAt, err = s.NextCron(ctx, g)

			if err != nil {
				return
			}

			err = s.repo.UpdateTx(ctx, tx, g)

			if err != nil {
				return
			}

			continue
		}

		if g.Current >= g.Target {
			continue
		}

		if g.MatchesActivity(a) {
			fmt.Printf("matched activity: %+v\n", a)
			g.Current += a.Duration
		} else {
			fmt.Printf("did not match activity: %+v\n", a)
		}

		if g.Current >= g.Target {
			completed = append(completed, g)
		}

		err = s.repo.UpdateTx(ctx, tx, g)

		if err != nil {
			return
		}
	}

	err = s.repo.FinishUpdate(ctx, tx)

	if err != nil {
		return
	}

	return
}
