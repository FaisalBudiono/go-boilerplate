package httputil

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/ztrue/tracerr"
)

func ValidateStruct(
	input any, structFieldJson map[string]string,
) (map[string][]domain.VerboseError, error) {
	err := validator.New().Struct(input)
	if err == nil {
		return nil, nil
	}

	var valErr validator.ValidationErrors
	if !errors.As(err, &valErr) {
		return nil, tracerr.Wrap(err)
	}

	errMsgs := make(map[string][]domain.VerboseError, 0)

	for _, fe := range valErr {
		vMsg := domain.NewErrVerbose(errcode.Code(fe.Tag()), "")

		for fieldName, jsonFieldName := range structFieldJson {
			if fe.Field() == fieldName {
				errMsgs[jsonFieldName] = append(errMsgs[jsonFieldName], vMsg)
			}
		}
	}

	return errMsgs, nil
}
