package res

import "FaisalBudiono/go-boilerplate/internal/app/domain"

type tokenPair struct {
	Type         string `json:"tokenType"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func newTokenPair(token domain.TokenPair) tokenPair {
	return tokenPair{
		Type:         "Bearer",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}

func TokenPair(token domain.TokenPair) response[tokenPair] {
	return response[tokenPair]{
		Data: newTokenPair(token),
	}
}
