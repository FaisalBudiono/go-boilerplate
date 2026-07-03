package port

import (
	"context"
	"errors"

	"komdigi-immigration/internal/app/domain"
)

var ErrClearanceNotFound = errors.New("immigration clearance not found")

type ImmigrationClearanceService interface {
	// Check return raw response and status code regardless of error
	Check(
		ctx context.Context,
		countryCode domain.CountryCode,
		passportNumber string,
	) (domain.ImmigrationClearance, []byte, int64, error)
}
