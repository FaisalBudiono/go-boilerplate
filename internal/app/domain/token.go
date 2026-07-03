package domain

type RefreshToken struct {
	ClientID     string
	ClientSecret string
}

func NewRefreshToken(clientID, clientSecret string) RefreshToken {
	return RefreshToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewTokenPair(accessToken, refreshToken string) TokenPair {
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

type TokenCredentialData struct {
	UserID       string
	ClientID     string
	HashedSecret string
	LoginMethod  LoginMethod
	LoginID      string
}

func NewTokenCredentialData(
	userID, clientID, hashedSecret string,
	loginMethod LoginMethod,
	loginID string,
) TokenCredentialData {
	return TokenCredentialData{
		UserID:       userID,
		ClientID:     clientID,
		HashedSecret: hashedSecret,
		LoginMethod:  loginMethod,
		LoginID:      loginID,
	}
}

type TokenCredential struct {
	ID           string
	UserID       string
	ClientID     string
	HashedSecret string
	LoginMethod  LoginMethod
	LoginID      string
}

func NewTokenCredential(
	id string,
	userID, clientID, hashedSecret string,
	loginMethod LoginMethod,
	loginID string,
) TokenCredential {
	return TokenCredential{
		ID:           id,
		UserID:       userID,
		ClientID:     clientID,
		HashedSecret: hashedSecret,
		LoginMethod:  loginMethod,
		LoginID:      loginID,
	}
}

type LoginMethod string

const (
	LoginMethodEmail    LoginMethod = "email"
	LoginMethodClientID LoginMethod = "clientID"
)

type UserTokenInfo struct {
	ID string

	LoginMethod LoginMethod
	LoginID     string
}

func NewUserTokenInfo(
	id string,
	loginMethod LoginMethod,
	loginID string,
) UserTokenInfo {
	return UserTokenInfo{
		ID: id,

		LoginMethod: loginMethod,
		LoginID:     loginID,
	}
}
