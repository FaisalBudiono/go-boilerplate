package getall

type queryOpt struct {
	UserIDs []string

	SearchClientID string
}

func NewQueryOpt() *queryOpt {
	return &queryOpt{
		UserIDs: nil,
	}
}

type QueryOption func(*queryOpt)

func WithUserIDs(userIDs ...string) QueryOption {
	return func(qo *queryOpt) {
		qo.UserIDs = userIDs
	}
}

func WithSearchClientID(clientID string) QueryOption {
	return func(qo *queryOpt) {
		qo.SearchClientID = clientID
	}
}
