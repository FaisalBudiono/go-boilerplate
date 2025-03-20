package res

import "FaisalBudiono/go-boilerplate/internal/app/domain"

type auth struct {
	Type         string `json:"type"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func Auth(token domain.Token) response[auth] {
	return response[auth]{
		Data: auth{
			Type:         "Bearer",
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		},
	}
}
