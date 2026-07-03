package domain

type Promise[T any] struct {
	Val T
	Err error
}

func NewPromise[T any](val T, err error) Promise[T] {
	return Promise[T]{
		Val: val,
		Err: err,
	}
}
