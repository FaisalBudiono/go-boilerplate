package httpfmt_test

import (
	"testing"

	"komdigi-immigration/internal/app/core/util/httpfmt"

	"github.com/stretchr/testify/suite"
)

type unprocessableErrSuite struct {
	suite.Suite
}

func TestUnprocessableErrSuite(t *testing.T) {
	suite.Run(t, new(unprocessableErrSuite))
}

func (suite *unprocessableErrSuite) TestIsError_should_return_false_when_no_meta() {
	err := httpfmt.NewUnprocessableErr()

	suite.False(err.IsError())
}

func (suite *unprocessableErrSuite) TestIsError_should_return_true_when_has_meta() {
	err := httpfmt.NewUnprocessableErr()

	err.Add("key", "code", "msg")

	suite.True(err.IsError())
}
