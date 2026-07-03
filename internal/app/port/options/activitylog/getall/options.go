package getall

import (
	"time"

	"komdigi-immigration/internal/app/domain/actlog"
)

type queryOpt struct {
	UserIDFlags []string
	ActionFlags []actlog.Action

	CreatedFrom time.Time
	CreatedTo   time.Time
}

func NewQueryOpt() *queryOpt {
	return &queryOpt{
		UserIDFlags: []string{},
		ActionFlags: []actlog.Action{},
	}
}

type QueryOption func(*queryOpt)

func WithUserIDFlags(userIDs ...string) QueryOption {
	return func(qo *queryOpt) {
		qo.UserIDFlags = userIDs
	}
}

func WithActionFlags(actions ...actlog.Action) QueryOption {
	return func(qo *queryOpt) {
		qo.ActionFlags = actions
	}
}

func WithCreatedFrom(from time.Time) QueryOption {
	return func(qo *queryOpt) {
		qo.CreatedFrom = from
	}
}

func WithCreatedTo(to time.Time) QueryOption {
	return func(qo *queryOpt) {
		qo.CreatedTo = to
	}
}
