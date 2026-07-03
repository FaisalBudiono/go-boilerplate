package httpfmt

import (
	"context"
	"fmt"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
)

type ValidationInput struct {
	Name  string
	Input any
}

func NewValidationInput(name string, input any) ValidationInput {
	return ValidationInput{
		Name:  name,
		Input: input,
	}
}

type Rule interface {
	Validate(uerr *UnprocessableErr, input ValidationInput) error
}

func Validate(
	ctx context.Context,
	uerr *UnprocessableErr,
	inputMap map[ValidationInput][]Rule,
) error {
	ctx, span := monitoring.Tracer().Start(ctx, "http.validate")
	defer span.End()

	for input, rules := range inputMap {
		for i, rule := range rules {
			err := rule.Validate(uerr, input)
			if err != nil {
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage(
						fmt.Sprintf("failed to validate on rules #%d", i),
					),
				)
				return err
			}
		}
	}

	return nil
}
