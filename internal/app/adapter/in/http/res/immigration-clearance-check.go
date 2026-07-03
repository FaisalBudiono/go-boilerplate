package res

import (
	"komdigi-immigration/internal/app/core/util/timeutil"
	"komdigi-immigration/internal/app/domain"
)

type immigrationClearance struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Gender    string `json:"gender"`

	PlaceOfBirth string `json:"placeOfBirth"`
	DateOfBirth  string `json:"dateOfBirth"`

	PassportNumber string `json:"passportNumber"`
	CountryCode    string `json:"countryCode"`

	IsInIndonesia bool `json:"isInIndonesia"`

	IssuedAt  string `json:"issuedAt"`
	ExpiredAt string `json:"expiredAt"`
}

func newImmigrationClearance(ic domain.ImmigrationClearance) immigrationClearance {
	return immigrationClearance{
		FirstName: ic.FirstName,
		LastName:  ic.LastName,
		Gender:    string(ic.Gender),

		PlaceOfBirth: ic.PlaceOfBirth,
		DateOfBirth:  ic.DateOfBirth.Format(timeutil.FormatDate),

		PassportNumber: ic.PassportNumber,
		CountryCode:    string(ic.CountryCode),

		IsInIndonesia: ic.IsInIndonesia,

		IssuedAt:  ic.IssuedAt.Format(timeutil.FormatDate),
		ExpiredAt: ic.ExpiredAt.Format(timeutil.FormatDate),
	}
}

func ImmigrationClearance(
	ic domain.ImmigrationClearance,
) response[immigrationClearance] {
	return response[immigrationClearance]{
		Data: newImmigrationClearance(ic),
	}
}
