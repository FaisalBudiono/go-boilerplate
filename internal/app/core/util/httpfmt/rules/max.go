package rules

import (
	"errors"
	"fmt"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
)

func MaxString(maxTotal int) httpfmt.Rule {
	return &maxRule{maxTotal}
}

type maxRule struct {
	maxTotal int
}

func (m *maxRule) Validate(
	uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput,
) error {
	raw, ok := input.Input.(string)
	if !ok {
		return errors.New("type should be string")
	}

	if len(raw) > m.maxTotal {
		uerr.Add(
			input.Name, invalid.MaxLength,
			fmt.Sprintf("Max length exceeded %d", m.maxTotal),
		)
	}

	return nil
}
