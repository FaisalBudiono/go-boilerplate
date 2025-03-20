package domain

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"fmt"
	"strings"
)

func NewErrVerbose(
	code errcode.Code,
	message string,
) VerboseError {
	return VerboseError{
		Code:    code,
		Message: message,
	}
}

type VerboseError struct {
	Code    errcode.Code
	Message string
}

func (e VerboseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type VerboseErrorList []verboseErrorWithFieldName

func (v VerboseErrorList) Error() string {
	errMsgs := make([]string, len(v))
	for i, err := range v {
		errMsgs[i] = err.Err.Error()
	}

	return strings.Join(errMsgs, " \n")
}

type verboseErrorWithFieldName struct {
	FieldName string
	Err       VerboseError
}

func NewVerboseErrorItem(
	fieldName string,
	err VerboseError,
) verboseErrorWithFieldName {
	return verboseErrorWithFieldName{
		FieldName: fieldName,
		Err:       err,
	}
}
