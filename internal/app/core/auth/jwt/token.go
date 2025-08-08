package jwt

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func (s *userSigner) Sign(u domain.UserTokenInfo) (string, error) {
	claims := userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.expiredDuration)),
		},

		ID: string(u.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.key)
	if err != nil {
		return "", errors.Join(errors.New("failed to signed user signer"), err)
	}

	return ss, nil
}

func (s *userSigner) Parse(token string) (domain.UserTokenInfo, error) {
	tok, err := jwt.ParseWithClaims(token, &userClaims{}, func(token *jwt.Token) (any, error) {
		return s.key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return domain.UserTokenInfo{}, (ErrTokenMalformed)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return domain.UserTokenInfo{}, (ErrSignatureInvalid)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.UserTokenInfo{}, (ErrTokenExpired)
		}

		return domain.UserTokenInfo{}, errors.Join(errors.New("failed to parse claims"), err)
	}

	claims, ok := tok.Claims.(*userClaims)
	if !ok {
		return domain.UserTokenInfo{}, errors.New("Failed to fetch claims")
	}

	return domain.NewUserBasicInfo(domid.UserID(claims.ID)), nil
}
