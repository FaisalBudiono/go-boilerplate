package res

import (
	"time"

	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
)

type activityLog struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	CreatedAt string `json:"createdAt"`
}

func newActivityLog(al domain.ActivityLog) activityLog {
	return activityLog{
		ID:        al.ID,
		Action:    al.Action.String(),
		CreatedAt: al.CreatedAt.Format(time.RFC3339),
	}
}

type activityLogMeta struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newActivityLogMeta(key actlog.MetaKey, val string) activityLogMeta {
	return activityLogMeta{
		Key:   key.String(),
		Value: val,
	}
}

type activityAuthMethod struct {
	LoginMethod string `json:"method"`
	LoginID     string `json:"id"`
}

func newActivityAuthMethod(lm domain.LoginMethod, lid string) activityAuthMethod {
	return activityAuthMethod{
		LoginMethod: string(lm),
		LoginID:     lid,
	}
}

type activityLogEagerLoad struct {
	activityLog

	Auth      *activityAuthMethod `json:"auth"`
	Metas     []activityLogMeta   `json:"metas"`
	CreatedBy *user               `json:"createdBy"`
}

func newActivityLogEagerLoad(
	al domain.ActivityLogEagerLoad,
) activityLogEagerLoad {
	authMethod := func() *activityAuthMethod {
		if al.User == nil {
			return nil
		}
		if al.Log.LoginMethod == nil {
			return nil
		}
		if al.Log.LoginID == nil {
			return nil
		}
		return new(newActivityAuthMethod(*al.Log.LoginMethod, *al.Log.LoginID))
	}()

	metas := make([]activityLogMeta, len(al.Metas))
	i := 0
	for key, val := range al.Metas {
		metas[i] = newActivityLogMeta(key, val)
		i++
	}

	actor := func() *user {
		if al.User == nil {
			return nil
		}
		return new(newUser(*al.User))
	}()

	return activityLogEagerLoad{
		activityLog: newActivityLog(al.Log),

		Auth:      authMethod,
		Metas:     metas,
		CreatedBy: actor,
	}
}

func ActivityLogs(
	logs []domain.ActivityLogEagerLoad, pg domain.Pagination,
) responsePaginated[activityLogEagerLoad] {
	res := make([]activityLogEagerLoad, len(logs))
	for i, log := range logs {
		res[i] = newActivityLogEagerLoad(log)
	}

	return responsePaginated[activityLogEagerLoad]{
		response: response[[]activityLogEagerLoad]{
			Data: res,
		},
		Meta: newPaginatedMeta(pg),
	}
}
