package jwt

import (
	"errors"
	"time"

	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/domain"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenMalformed   = jwt.ErrTokenMalformed
	ErrSignatureInvalid = jwt.ErrSignatureInvalid
	ErrTokenExpired     = jwt.ErrTokenExpired
)

func NewUserSigner(
	key []byte,
	expiredDuration time.Duration,
) *userSigner {
	return &userSigner{
		key:             key,
		expiredDuration: expiredDuration,
	}
}

type userSigner struct {
	key             []byte
	expiredDuration time.Duration
}

func (s *userSigner) Sign(u domain.UserTokenInfo) (string, error) {
	claims := userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.expiredDuration)),
		},

		ID:          string(u.ID),
		LoginMethod: string(u.LoginMethod),
		LoginID:     u.LoginID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.key)
	if err != nil {
		return "", errors.Join(errors.New("failed to signed user signer"), err)
	}

	return ss, nil
}

func (s *userSigner) Parse(token string) (domain.UserTokenInfo, error) {
	emptyVal := domain.UserTokenInfo{}

	tok, err := jwt.ParseWithClaims(token, &userClaims{}, func(token *jwt.Token) (any, error) {
		return s.key, nil
	})
	if err != nil {
		expectedErrs := []error{ErrTokenMalformed, ErrSignatureInvalid, ErrTokenExpired}
		if errs.Is(err, expectedErrs...) {
			return emptyVal, err
		}

		return emptyVal, errors.Join(errors.New("failed to parse claims"), err)
	}

	claims, ok := tok.Claims.(*userClaims)
	if !ok {
		return emptyVal, errors.New("failed to fetch claims")
	}

	return domain.NewUserTokenInfo(
		claims.ID,
		domain.LoginMethod(claims.LoginMethod),
		claims.LoginID,
	), nil
}

type userClaims struct {
	jwt.RegisteredClaims

	ID          string `json:"uid"`
	LoginMethod string `json:"loginMethod"`
	LoginID     string `json:"loginID"`
}
