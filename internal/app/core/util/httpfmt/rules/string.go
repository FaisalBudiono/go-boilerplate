package rules

import (
	"errors"
	"strconv"
	"strings"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
)

func StrToInt() httpfmt.Rule {
	return &strToInt{}
}

type strToInt struct{}

func (r *strToInt) Validate(
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

	_, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		uerr.Add(input.Name, invalid.ShouldNumber, "")
	}
	return nil
}
