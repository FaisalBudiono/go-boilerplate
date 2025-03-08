package res

type response[T any] struct {
	Data T `json:"data"`
}
