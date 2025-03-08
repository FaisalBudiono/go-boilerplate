package res

import "FaisalBudiono/go-boilerplate/internal/app/domain"

type Auth struct {
	Type         string `json:"type"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func ToAuth(token domain.Token) response[Auth] {
	return response[Auth]{
		Data: Auth{
			Type:         "Bearer",
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		},
	}
}
