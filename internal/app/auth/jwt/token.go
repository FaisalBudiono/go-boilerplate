package jwt

import (
	"FaisalBudiono/go-boilerplate/internal/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ztrue/tracerr"
)

var (
	ErrTokenMalformed   = errors.New("Invalid malformed format")
	ErrSignatureInvalid = errors.New("Invalid signature")
	ErrTokenExpired     = errors.New("Token expired")
)

type userSigner struct {
	key             []byte
	expiredDuration time.Duration
}

func NewUserSigner(
	key []byte,
	expiredDuration time.Duration,
) *userSigner {
	return &userSigner{
		key:             key,
		expiredDuration: expiredDuration,
	}
}

type userClaims struct {
	jwt.RegisteredClaims

	ID string `json:"uid"`
}

func (s *userSigner) Sign(u domain.UserBasicInfo) (string, error) {
	claims := userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.expiredDuration)),
		},

		ID: u.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.key)
	if err != nil {
		return "", tracerr.Wrap(err)
	}

	return ss, nil
}

func (s *userSigner) Parse(token string) (domain.UserBasicInfo, error) {
	tok, err := jwt.ParseWithClaims(token, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return domain.UserBasicInfo{}, tracerr.Wrap(ErrTokenMalformed)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return domain.UserBasicInfo{}, tracerr.Wrap(ErrSignatureInvalid)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.UserBasicInfo{}, tracerr.Wrap(ErrTokenExpired)
		}

		return domain.UserBasicInfo{}, tracerr.Wrap(err)
	}

	claims, ok := tok.Claims.(*userClaims)
	if !ok {
		return domain.UserBasicInfo{}, tracerr.New("Failed to fetch claims")
	}

	return domain.NewUserBasicInfo(claims.ID), nil
}
