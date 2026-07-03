package rules

import (
	"errors"

	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/domain"
)

func Luhn() httpfmt.Rule {
	return &luhnRule{}
}

type luhnRule struct{}

func (r *luhnRule) Validate(
	uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput,
) error {
	raw, ok := input.Input.(string)
	if !ok {
		return errors.New("type should be string")
	}

	if len(raw) != 15 {
		uerr.Add(input.Name, invalid.ShouldString, "Should be 15 digits")
	}

	imei := domain.IMEI(raw)
	if !imei.IsValid() {
		uerr.Add(input.Name, invalid.Luhn, "Not passed luhn check")
	}

	return nil
}
