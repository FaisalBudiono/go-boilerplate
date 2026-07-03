package rules

import (
	"errors"
	"fmt"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
)

func SliceInside[T any](rules ...httpfmt.Rule) httpfmt.Rule {
	return &sliceInside[T]{rules}
}

type sliceInside[T any] struct {
	rules []httpfmt.Rule
}

func (s *sliceInside[T]) Validate(
	uerr *httpfmt.UnprocessableErr,
	input httpfmt.ValidationInput,
) error {
	raws, ok := input.Input.([]T)
	if !ok {
		return errors.New("type not supported")
	}

	for i, raw := range raws {
		valInput := httpfmt.NewValidationInput(
			fmt.Sprintf("%s.%d", input.Name, i), raw,
		)
		for _, rule := range s.rules {
			err := rule.Validate(uerr, valInput)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
