package rules

import (
	"errors"
	"strings"
	"time"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
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
