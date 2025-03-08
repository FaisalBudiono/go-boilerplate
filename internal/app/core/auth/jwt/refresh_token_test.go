package jwt_test

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type refreshTokenSignerSuite struct {
	suite.Suite

	key []byte
}

func TestRefreshTokenSigner(t *testing.T) {
	suite.Run(t, new(refreshTokenSignerSuite))
}

func (suite *refreshTokenSignerSuite) SetupTest() {
	suite.key = []byte("AllYourBase")
}

func (suite *refreshTokenSignerSuite) TestSign_should_return_signed_token() {
	srv := jwt.NewRefreshTokenSigner(suite.key)

	res, err := srv.Sign("some-random-payload")
	suite.Nil(err)
	suite.NotEqual("", res)
}

func (suite *refreshTokenSignerSuite) TestParse_should_return_error_when_token_malformed() {
	srv := jwt.NewRefreshTokenSigner(suite.key)

	_, err := srv.ParsePayload("asd")

	suite.ErrorIs(err, jwt.ErrTokenMalformed)
}

func (suite *refreshTokenSignerSuite) TestParse_should_return_error_when_key_invalid() {
	srv := jwt.NewRefreshTokenSigner(suite.key)
	tok, _ := srv.Sign("some-random-payload")

	srv2 := jwt.NewRefreshTokenSigner([]byte("invalid-key"))
	_, err := srv2.ParsePayload(tok)

	suite.ErrorIs(err, jwt.ErrSignatureInvalid)
}

func (suite *refreshTokenSignerSuite) TestParse_should_return_refresh_token_payload() {
	payload := "some-random-payload"

	srv := jwt.NewRefreshTokenSigner(suite.key)
	tok, _ := srv.Sign(payload)

	res, err := srv.ParsePayload(tok)
	suite.Nil(err)
	suite.Equal(payload, res)
}
