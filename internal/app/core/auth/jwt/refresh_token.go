package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ztrue/tracerr"
)

type refreshTokenSigner struct {
	key []byte
}

func NewRefreshTokenSigner(key []byte) *refreshTokenSigner {
	return &refreshTokenSigner{
		key: key,
	}
}

type refreshTokenClaims struct {
	jwt.RegisteredClaims

	Payload string `json:"payload"`
}

func (signer *refreshTokenSigner) Sign(payload string) (string, error) {
	claims := refreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		},

		Payload: payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signer.key)
	if err != nil {
		return "", tracerr.Wrap(err)
	}

	return ss, nil
}

func (signer *refreshTokenSigner) ParsePayload(token string) (string, error) {
	tok, err := jwt.ParseWithClaims(token, &refreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		return signer.key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return "", tracerr.Wrap(ErrTokenMalformed)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return "", tracerr.Wrap(ErrSignatureInvalid)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", tracerr.Wrap(ErrTokenExpired)
		}

		return "", tracerr.Wrap(err)
	}

	claims, ok := tok.Claims.(*refreshTokenClaims)
	if !ok {
		return "", tracerr.New("Failed to fetch claims")
	}

	return claims.Payload, nil
}
