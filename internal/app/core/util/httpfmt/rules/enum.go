package rules

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
)

type enumValidable interface {
	String() string
}

func EnumFlag[T enumValidable](
	allowed []T,
) httpfmt.Rule {
	return &enumFlag[T]{
		allowed: allowed,
	}
}

type enumFlag[T enumValidable] struct {
	allowed []T
}

func (r *enumFlag[T]) Validate(
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

	allowedStrings := make([]string, len(r.allowed))
	for i, allowed := range r.allowed {
		allowedStrings[i] = allowed.String()
	}
	allowedMsg := fmt.Sprintf("Only accept: [%s]", strings.Join(allowedStrings, ","))

	raws := strings.SplitSeq(trimmed, ",")
	for raw := range raws {
		if !slices.Contains(allowedStrings, raw) {
			uerr.Add(input.Name, invalid.ShouldEnum, allowedMsg)
		}
	}

	return nil
}

func Enum[T enumValidable](
	allowed []T,
) httpfmt.Rule {
	return &enumSingle[T]{
		allowed: allowed,
	}
}

type enumSingle[T enumValidable] struct {
	allowed []T
}

func (r *enumSingle[T]) Validate(uerr *httpfmt.UnprocessableErr, input httpfmt.ValidationInput) error {
	raw, ok := input.Input.(string)
	if !ok {
		return errors.New("type should be string")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	allowedStrings := make([]string, len(r.allowed))
	for i, allowed := range r.allowed {
		allowedStrings[i] = allowed.String()
	}
	allowedMsg := fmt.Sprintf("Only accept: [%s]", strings.Join(allowedStrings, ","))

	if !slices.Contains(allowedStrings, raw) {
		uerr.Add(input.Name, invalid.ShouldEnum, allowedMsg)
	}

	return nil
}
