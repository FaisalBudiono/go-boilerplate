package res

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
)

type errResponse struct {
	Msg     string       `json:"message"`
	ErrCode errcode.Code `json:"errorCode"`
}

func NewErrorGeneric() errResponse {
	return errResponse{
		Msg:     "Something wrong in the server",
		ErrCode: errcode.Generic,
	}
}

func NewError(msg string, code errcode.Code) errResponse {
	return errResponse{
		Msg:     msg,
		ErrCode: code,
	}
}

type verboseMetaErr struct {
	Code string `json:"code"`
	Msg  string `json:"message,omitempty"`
}

func NewVerboseMeta(code string, msg string) verboseMetaErr {
	return verboseMetaErr{
		Code: code,
		Msg:  msg,
	}
}

type VerboseMetaMsgs map[string][]verboseMetaErr

func (v VerboseMetaMsgs) Append(key string, vErr ...verboseMetaErr) VerboseMetaMsgs {
	v[key] = append(v[key], vErr...)

	return v
}

func (v VerboseMetaMsgs) AppendDom(key string, err ...domain.VerboseError) VerboseMetaMsgs {
	items := make([]verboseMetaErr, len(err))
	for i := range err {
		domErr := err[i]

		items[i] = NewVerboseMeta(string(domErr.Code), domErr.Message)
	}

	v.Append(key, items...)

	return v
}

func (v VerboseMetaMsgs) AppendDomMap(mapErr map[string][]domain.VerboseError) VerboseMetaMsgs {
	for key := range mapErr {
		vErrs := mapErr[key]

		v.AppendDom(key, vErrs...)
	}

	return v
}

func (v VerboseMetaMsgs) AppendVList(e domain.VerboseErrorList) VerboseMetaMsgs {
	for _, eVal := range e {
		v.Append(eVal.FieldName, NewVerboseMeta(string(eVal.Err.Code), eVal.Err.Message))
	}

	return v
}

type UnprocessableErrResponse struct {
	errResponse

	Meta VerboseMetaMsgs `json:"meta"`
}

func (e *UnprocessableErrResponse) Error() string {
	return "Invalid param request"
}

func NewErrorUnprocessable(meta VerboseMetaMsgs) *UnprocessableErrResponse {
	return &UnprocessableErrResponse{
		errResponse: errResponse{
			Msg:     "Structure body/param might be invalid.",
			ErrCode: errcode.InvalidParam,
		},
		Meta: meta,
	}
}
