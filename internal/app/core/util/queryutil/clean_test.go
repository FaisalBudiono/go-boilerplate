package queryutil_test

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type cleanerSuite struct {
	suite.Suite
}

func TestCleaner(t *testing.T) {
	suite.Run(t, new(cleanerSuite))
}

func (suite *cleanerSuite) TestCleaner() {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input: `
something


        like
  this 
            `,
			expected: "something like this",
		},
		{
			input: `
kuda

        lumping .

    #2  1     39  ".
            `,
			expected: "kuda lumping . #2 1 39 \".",
		},
	}

	for i, c := range cases {
		suite.Run(fmt.Sprintf("case-%d", i+1), func() {
			suite.Equal(c.expected, queryutil.Clean(c.input))
		})
	}
}
