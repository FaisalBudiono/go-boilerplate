package rules

import (
	"errors"
	"fmt"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
)

func EnumFlag[T enumValidable](
	allowed []T,
) httpfmt.Rule {
	return &enumFlag[T]{
		allowed: allowed,
	}
}

type enumValidable interface {
	String() string
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
		if !strings.Contains(allowedMsg, raw) {
			uerr.Add(input.Name, invalid.ShouldEnum, allowedMsg)
		}
	}

	return nil
}
