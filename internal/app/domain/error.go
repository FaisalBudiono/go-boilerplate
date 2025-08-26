package domain

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"fmt"
	"strings"
)

func OLDNewErrVerbose(
	code errcode.Code,
	message string,
) OLDVerboseError {
	return OLDVerboseError{
		Code:    code,
		Message: message,
	}
}

type OLDVerboseError struct {
	Code    errcode.Code
	Message string
}

func (e OLDVerboseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type OLDVerboseErrorList []oldverboseErrorWithFieldName

func (v OLDVerboseErrorList) Error() string {
	errMsgs := make([]string, len(v))
	for i, err := range v {
		errMsgs[i] = err.Err.Error()
	}

	return strings.Join(errMsgs, " \n")
}

type oldverboseErrorWithFieldName struct {
	FieldName string
	Err       OLDVerboseError
}

func OLDNewVerboseErrorItem(
	fieldName string,
	err OLDVerboseError,
) oldverboseErrorWithFieldName {
	return oldverboseErrorWithFieldName{
		FieldName: fieldName,
		Err:       err,
	}
}
