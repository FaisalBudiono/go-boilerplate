package rules

import (
	"errors"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
)

func Required[T comparable]() httpfmt.Rule {
	return &ruleRequired[T]{}
}

type ruleRequired[T comparable] struct {
	empyVal T
}

func (r *ruleRequired[T]) Validate(
	uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput,
) error {
	val, ok := input.Input.(T)
	if !ok {
		return errors.New("type not supported")
	}

	if val == r.empyVal {
		uerr.Add(input.Name, invalid.Required, "")
	}

	return nil
}
