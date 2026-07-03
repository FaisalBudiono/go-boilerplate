package res

import (
	"time"

	"komdigi-immigration/internal/app/core/util/timeutil"
	"komdigi-immigration/internal/app/domain"
)

type immigrationClearanceLog struct {
	ID          string `json:"id"`
	RawResponse string `json:"rawResponse"`
	StatusCode  int64  `json:"statusCode"`

	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Gender    *string `json:"gender"`

	PlaceOfBirth *string `json:"placeOfBirth"`
	DateOfBirth  *string `json:"dateOfBirth"`

	PassportNumber *string `json:"passportNumber"`
	CountryCode    *string `json:"countryCode"`

	IsInIndonesia *bool `json:"isInIndonesia"`

	IssuedAt  *string `json:"issuedAt"`
	ExpiredAt *string `json:"expiredAt"`

	ActorID *string `json:"actorID"`
	IMEI    string  `json:"imei"`

	CreatedAt string `json:"createdAt"`
}

func newImmigrationClearanceLog(icl domain.ImmigrationClearanceLog) immigrationClearanceLog {
	var gender *string
	if icl.Gender != nil {
		gender = new(icl.Gender.String())
	}

	var countryCode *string
	if icl.CountryCode != nil {
		countryCode = new(icl.CountryCode.String())
	}

	return immigrationClearanceLog{
		ID:          icl.ID,
		RawResponse: string(icl.RawResponse),
		StatusCode:  icl.StatusCode,

		FirstName: icl.FirstName,
		LastName:  icl.LastName,
		Gender:    gender,

		PlaceOfBirth: icl.PlaceOfBirth,
		DateOfBirth:  optionalDateOnly(icl.DateOfBirth),

		PassportNumber: icl.PassportNumber,
		CountryCode:    countryCode,

		IsInIndonesia: icl.IsInIndonesia,

		IssuedAt:  optionalDateOnly(icl.IssuedAt),
		ExpiredAt: optionalDateOnly(icl.ExpiredAt),

		ActorID: icl.ActorID,
		IMEI:    icl.IMEI.String(),

		CreatedAt: icl.CreatedAt.Format(timeutil.FormatDate),
	}
}

type immigrationClearanceLogEagerLoad struct {
	immigrationClearanceLog

	CreatedBy *user `json:"createdBy"`
}

func newImmigrationClearanceLogEagerLoad(
	icl domain.ImmigrationClearanceLogEagerLoad,
) immigrationClearanceLogEagerLoad {
	var createdBy *user
	if icl.Actor != nil {
		createdBy = new(newUser(*icl.Actor))
	}

	return immigrationClearanceLogEagerLoad{
		immigrationClearanceLog: newImmigrationClearanceLog(icl.ICL),
		CreatedBy:               createdBy,
	}
}

func optionalDateOnly(d *time.Time) *string {
	if d == nil {
		return nil
	}
	return new(d.Format(timeutil.FormatDate))
}

func PaginatedImmigrationClearanceLog(
	icls []domain.ImmigrationClearanceLogEagerLoad,
	pg domain.Pagination,
) responsePaginated[immigrationClearanceLogEagerLoad] {
	res := make([]immigrationClearanceLogEagerLoad, len(icls))
	for i, icl := range icls {
		res[i] = newImmigrationClearanceLogEagerLoad(icl)
	}

	return responsePaginated[immigrationClearanceLogEagerLoad]{
		response: response[[]immigrationClearanceLogEagerLoad]{
			Data: res,
		},
		Meta: newPaginatedMeta(pg),
	}
}
