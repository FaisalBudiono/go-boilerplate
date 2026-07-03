package activitylog

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
)

func New(db *sql.DB, repo port.ActivityLogRepo) *ActivityLog {
	return &ActivityLog{
		db:   db,
		repo: repo,
	}
}

type ActivityLog struct {
	db   *sql.DB
	repo port.ActivityLogRepo
}

func (srv *ActivityLog) New(action actlog.Action) *LogInstance {
	return newLogInstance(srv.db, srv.repo, action)
}

func newLogInstance(
	db *sql.DB,
	repo port.ActivityLogRepo,
	action actlog.Action,
) *LogInstance {
	return &LogInstance{
		db:   db,
		repo: repo,

		action: action,

		meta: make(map[actlog.MetaKey]string),
	}
}

type LogInstance struct {
	db   *sql.DB
	repo port.ActivityLogRepo

	action actlog.Action

	mu sync.Mutex

	userID      *string
	loginMethod *domain.LoginMethod
	loginID     *string

	meta map[actlog.MetaKey]string
}

func (log *LogInstance) Actor(actor *domain.Userinfo) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if actor == nil {
		return
	}

	log.userID = &actor.User.User.ID
	log.loginMethod = &actor.UserTokenInfo.LoginMethod
	log.loginID = &actor.UserTokenInfo.LoginID
}

func (log *LogInstance) AddMeta(key actlog.MetaKey, value string) {
	log.mu.Lock()
	defer log.mu.Unlock()

	log.meta[key] = value
}

func (log *LogInstance) Log(ctx context.Context) error {
	log.mu.Lock()
	defer log.mu.Unlock()

	ctx, span := monitoring.Tracer().Start(
		ctx, "core.logger.activitylog.log-instance.Log",
	)
	defer span.End()

	tx, err := log.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin tx"),
		)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if !errors.Is(err, port.ErrTxDone) {
				otelutil.SpanLogError(
					span, err,
					otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to rollback transaction"),
				)
			}
		}
	}()

	id, err := log.repo.Insert(
		ctx, log.db,
		domain.NewActivityLogData(
			log.userID, log.loginMethod, log.loginID, log.action,
		),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert activity log"),
		)
		return err
	}

	err = log.repo.AddMetadata(ctx, tx, id, log.meta)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to add metadata"),
		)
		return err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
		return err
	}

	return nil
}
