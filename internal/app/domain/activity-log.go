package domain

import (
	"time"

	"komdigi-immigration/internal/app/domain/actlog"
)

type ActivityLog struct {
	ID string

	ActorID     *string
	LoginMethod *LoginMethod
	LoginID     *string

	Action    actlog.Action
	CreatedAt time.Time
}

func NewActivityLog(
	id string,
	actorID *string,
	loginMethod *LoginMethod,
	loginID *string,
	action actlog.Action,
	createdAt time.Time,
) ActivityLog {
	return ActivityLog{
		ID: id,

		ActorID:     actorID,
		LoginMethod: loginMethod,
		LoginID:     loginID,

		Action:    action,
		CreatedAt: createdAt,
	}
}

type ActivityLogData struct {
	ActorID     *string
	LoginMethod *LoginMethod
	LoginID     *string

	Action actlog.Action
}

func NewActivityLogData(
	actorID *string,
	loginMethod *LoginMethod,
	loginID *string,
	action actlog.Action,
) ActivityLogData {
	return ActivityLogData{
		ActorID:     actorID,
		LoginMethod: loginMethod,
		LoginID:     loginID,

		Action: action,
	}
}

type ActivityLogEagerLoad struct {
	Log ActivityLog

	Metas map[actlog.MetaKey]string
	User  *UserEagerLoad
}

func NewActivityLogEagerLoad(
	log ActivityLog,
	metas map[actlog.MetaKey]string,
	user *UserEagerLoad,
) ActivityLogEagerLoad {
	return ActivityLogEagerLoad{
		Log: log,

		Metas: metas,
		User:  user,
	}
}
