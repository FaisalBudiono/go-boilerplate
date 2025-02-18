package jwt_test

import (
	"FaisalBudiono/go-boilerplate/internal/app/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type userSignerSuite struct {
	suite.Suite

	key []byte
}

func TestUserSigner(t *testing.T) {
	suite.Run(t, new(userSignerSuite))
}

func (suite *userSignerSuite) SetupTest() {
	suite.key = []byte("AllYourBase")
}

func (suite *userSignerSuite) TestSign_should_return_signed_token() {
	u := domain.NewUserBasicInfo("some-id")

	srv := jwt.NewUserSigner(suite.key, time.Second)

	res, err := srv.Sign(u)
	suite.Nil(err)
	suite.NotEqual("", res)
}

func (suite *userSignerSuite) TestParse_should_return_error_when_token_malformed() {
	srv := jwt.NewUserSigner(suite.key, time.Second)

	_, err := srv.Parse("asd")

	suite.ErrorIs(err, jwt.ErrTokenMalformed)
}

func (suite *userSignerSuite) TestParse_should_return_error_when_key_invalid() {
	u := domain.NewUserBasicInfo("some-id")

	srv := jwt.NewUserSigner(suite.key, time.Second)
	tok, _ := srv.Sign(u)

	srv2 := jwt.NewUserSigner([]byte("invalid-key"), time.Second)
	_, err := srv2.Parse(tok)

	suite.ErrorIs(err, jwt.ErrSignatureInvalid)
}

func (suite *userSignerSuite) TestParse_should_return_error_when_token_expired() {
	u := domain.NewUserBasicInfo("some-user-id")

	srv := jwt.NewUserSigner(suite.key, -1*time.Second)
	tok, _ := srv.Sign(u)

	_, err := srv.Parse(tok)
	suite.ErrorIs(err, jwt.ErrTokenExpired)
}

func (suite *userSignerSuite) TestParse_should_return_basic_user_info() {
	u := domain.NewUserBasicInfo("some-user-id")

	srv := jwt.NewUserSigner(suite.key, time.Second)
	tok, _ := srv.Sign(u)

	ures, err := srv.Parse(tok)
	suite.Nil(err)
	suite.Equal(u, ures)
}
