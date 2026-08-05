package httpfmt

import (
	"context"
	"fmt"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
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

type RuleReq struct {
	input ValidationInput
	rule  []Rule
}

func NewRuleReq(name string, input any, rules ...Rule) RuleReq {
	return RuleReq{
		input: NewValidationInput(name, input),
		rule:  rules,
	}
}

func Validate(
	ctx context.Context,
	uerr *UnprocessableErr,
	inputs []RuleReq,
) error {
	ctx, span := mon.Tracer().Start(ctx, "http.validate")
	defer span.End()

	for _, input := range inputs {
		for i, rule := range input.rule {
			err := rule.Validate(uerr, input.input)
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
