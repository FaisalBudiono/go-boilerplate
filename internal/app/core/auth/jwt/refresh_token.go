package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		return "", errors.Join(errors.New("failed to signed string"), err)
	}

	return ss, nil
}

func (signer *refreshTokenSigner) ParsePayload(token string) (string, error) {
	tok, err := jwt.ParseWithClaims(token, &refreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		return signer.key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return "", (ErrTokenMalformed)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return "", (ErrSignatureInvalid)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", (ErrTokenExpired)
		}

		return "", errors.Join(errors.New("failed to parse claims"), err)
	}

	claims, ok := tok.Claims.(*refreshTokenClaims)
	if !ok {
		return "", errors.New("Failed to fetch claims")
	}

	return claims.Payload, nil
}
