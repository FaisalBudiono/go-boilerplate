package res

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
)

type oldErrResponse struct {
	Msg     string       `json:"message"`
	ErrCode errcode.Code `json:"errorCode"`
}

func OLDNewErrorGeneric() oldErrResponse {
	return oldErrResponse{
		Msg:     "Something wrong in the server",
		ErrCode: errcode.Generic,
	}
}

func OLDNewError(msg string, code errcode.Code) oldErrResponse {
	return oldErrResponse{
		Msg:     msg,
		ErrCode: code,
	}
}

type oldverboseMetaErr struct {
	Code string `json:"code"`
	Msg  string `json:"message,omitempty"`
}

func OLDNewVerboseMeta(code string, msg string) oldverboseMetaErr {
	return oldverboseMetaErr{
		Code: code,
		Msg:  msg,
	}
}

type OLDVerboseMetaMsgs map[string][]oldverboseMetaErr

func (v OLDVerboseMetaMsgs) Append(key string, vErr ...oldverboseMetaErr) OLDVerboseMetaMsgs {
	v[key] = append(v[key], vErr...)

	return v
}

func (v OLDVerboseMetaMsgs) AppendDom(key string, err ...domain.OLDVerboseError) OLDVerboseMetaMsgs {
	items := make([]oldverboseMetaErr, len(err))
	for i := range err {
		domErr := err[i]

		items[i] = OLDNewVerboseMeta(string(domErr.Code), domErr.Message)
	}

	v.Append(key, items...)

	return v
}

func (v OLDVerboseMetaMsgs) AppendDomMap(mapErr map[string][]domain.OLDVerboseError) OLDVerboseMetaMsgs {
	for key := range mapErr {
		vErrs := mapErr[key]

		v.AppendDom(key, vErrs...)
	}

	return v
}

func (v OLDVerboseMetaMsgs) AppendVList(e domain.OLDVerboseErrorList) OLDVerboseMetaMsgs {
	for _, eVal := range e {
		v.Append(eVal.FieldName, OLDNewVerboseMeta(string(eVal.Err.Code), eVal.Err.Message))
	}

	return v
}

type OLDUnprocessableErrResponse struct {
	oldErrResponse

	Meta OLDVerboseMetaMsgs `json:"meta"`
}

func (e *OLDUnprocessableErrResponse) Error() string {
	return "Invalid param request"
}

func OLDNewErrorUnprocessable(meta OLDVerboseMetaMsgs) *OLDUnprocessableErrResponse {
	return &OLDUnprocessableErrResponse{
		oldErrResponse: oldErrResponse{
			Msg:     "Structure body/param might be invalid.",
			ErrCode: errcode.InvalidParam,
		},
		Meta: meta,
	}
}
