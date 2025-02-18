package res

import "FaisalBudiono/go-boilerplate/internal/http/res/errcode"

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

type metaErr struct {
	Code string `json:"code"`
}

type UnprocessableErrResponse struct {
	errResponse

	Meta map[string][]metaErr `json:"meta"`
}

func (e *UnprocessableErrResponse) Error() string {
	return "Invalid param request"
}

func NewErrorUnprocessable(meta map[string][]string) *UnprocessableErrResponse {
	metaMap := make(map[string][]metaErr, 0)

	for k, codes := range meta {
		for _, c := range codes {
			metaMap[k] = append(metaMap[k], metaErr{
				Code: c,
			})
		}
	}

	return &UnprocessableErrResponse{
		errResponse: errResponse{
			Msg:     "Structure body/param might be invalid.",
			ErrCode: errcode.InvalidParam,
		},
		Meta: metaMap,
	}
}
