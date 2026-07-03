package rules

import (
	"errors"
	"strings"
	"time"

	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/timeutil"
)

func Datetime() httpfmt.Rule {
	return &datetime{}
}

type datetime struct{}

func (d *datetime) Validate(
	uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput,
) error {
	raw, ok := input.Input.(string)
	if !ok {
		return errors.New("type should be string")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	_, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		uerr.Add(input.Name, invalid.ShouldDate, "")
	}

	return nil
}

func Date() httpfmt.Rule {
	return &date{}
}

type date struct{}

func (d *date) Validate(
	uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput,
) error {
	raw, ok := input.Input.(string)
	if !ok {
		return errors.New("type should be string")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	_, err := time.Parse(timeutil.FormatDate, trimmed)
	if err != nil {
		uerr.Add(input.Name, invalid.ShouldDateOnly, "Should be in format YYYY-MM-DD")
	}

	return nil
}
