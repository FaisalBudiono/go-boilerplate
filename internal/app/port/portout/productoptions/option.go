package productoptions

type queryOpt struct {
	ShowAll bool
}

func NewQueryOpt() *queryOpt {
	return &queryOpt{}
}

type QueryOption func(*queryOpt)

func WithShowFlag(showAll bool) QueryOption {
	return func(qo *queryOpt) {
		qo.ShowAll = showAll
	}
}
