package res

type errCode string

const (
	ECInvalidParam errCode = "invalid-structure-validation"
)

type errResponse struct {
	Msg     string  `json:"message"`
	ErrCode errCode `json:"errorCode"`
}

type errMetaValue struct {
	Code string `json:"code"`
	Msg  string `json:"message,omitempty"`
}

type errMetaUnprocessableKeyValue map[string][]errMetaValue

type unprocessableErrResponse struct {
	errResponse

	Meta errMetaUnprocessableKeyValue `json:"meta"`
}

// func NewErrUnprocessable(meta OLDVerboseMetaMsgs) *unprocessableErrResponse {
// 	return &unprocessableErrResponse{
// 		errResponse: errResponse{
// 			Msg:     "Structure body/param might be invalid.",
// 			ErrCode: ECInvalidParam,
// 		},
// 	}
// }

func (e *unprocessableErrResponse) Error() string {
	return "Invalid param request"
}
