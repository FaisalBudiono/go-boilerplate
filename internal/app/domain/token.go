package domain

type Token struct {
	AccessToken  string
	RefreshToken string
}

func NewToken(
	accessToken string,
	refreshToken string,
) Token {
	return Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
