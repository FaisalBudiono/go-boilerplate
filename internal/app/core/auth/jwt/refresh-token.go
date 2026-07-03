package jwt

import (
	"errors"
	"time"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/domain"

	"github.com/golang-jwt/jwt/v5"
)

func NewRefreshTokenSigner(key []byte) *refreshTokenSigner {
	return &refreshTokenSigner{
		key:             key,
		expiredDuration: 24 * time.Hour,
	}
}

type refreshTokenSigner struct {
	key             []byte
	expiredDuration time.Duration
}

func (signer *refreshTokenSigner) Sign(id, secret string) (string, error) {
	claims := refreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(signer.expiredDuration)),
		},

		ClientID:     id,
		ClientSecret: secret,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signer.key)
	if err != nil {
		return "", errors.Join(errors.New("failed to signed string"), err)
	}

	return ss, nil
}

func (signer *refreshTokenSigner) Parse(token string) (domain.RefreshToken, error) {
	emptyVal := domain.RefreshToken{}

	tok, err := jwt.ParseWithClaims(token, &refreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		return signer.key, nil
	})
	if err != nil {
		expectedErrs := []error{ErrTokenMalformed, ErrSignatureInvalid, ErrTokenExpired}
		if errs.Is(err, expectedErrs...) {
			return emptyVal, err
		}

		return emptyVal, errors.Join(errors.New("failed to parse claims"), err)
	}

	claims, ok := tok.Claims.(*refreshTokenClaims)
	if !ok {
		return emptyVal, errors.New("failed to fetch claims")
	}

	return domain.NewRefreshToken(
		claims.ClientID,
		claims.ClientSecret,
	), nil
}

type refreshTokenClaims struct {
	jwt.RegisteredClaims

	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}
