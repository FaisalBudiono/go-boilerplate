package domain

import "time"

type ClientCredential struct {
	ID        string
	UserID    string
	ClientID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewClientCredential(
	id string,
	userID string,
	clientID string,
	createdAt time.Time,
	updatedAt time.Time,
) ClientCredential {
	return ClientCredential{
		ID:        id,
		UserID:    userID,
		ClientID:  clientID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

type ClientCredentialData struct {
	UserID       string
	ClientID     string
	HashedSecret string
}

func NewClientCredentialData(
	userID string,
	clientID string,
	hashedSecret string,
) ClientCredentialData {
	return ClientCredentialData{
		UserID:       userID,
		ClientID:     clientID,
		HashedSecret: hashedSecret,
	}
}

type ClientCredentialSecret struct {
	CC ClientCredential

	Secret string
}

func NewClientCredentialSecret(
	id string,
	userID string,
	clientID string,
	secret string,
	createdAt time.Time,
	updatedAt time.Time,
) ClientCredentialSecret {
	return ClientCredentialSecret{
		CC: NewClientCredential(
			id,
			userID,
			clientID,
			createdAt,
			updatedAt,
		),
		Secret: secret,
	}
}

type ClientCredentialSecretEagerLoad struct {
	CC ClientCredentialSecret

	User UserEagerLoad
}

func NewClientCredentialSecretEagerLoad(
	cc ClientCredentialSecret,
	user UserEagerLoad,
) ClientCredentialSecretEagerLoad {
	return ClientCredentialSecretEagerLoad{
		CC:   cc,
		User: user,
	}
}

type ClientCredentialEagerLoad struct {
	CC ClientCredential

	User UserEagerLoad
}

func NewClientCredentialEagerLoad(
	cc ClientCredential,
	user UserEagerLoad,
) ClientCredentialEagerLoad {
	return ClientCredentialEagerLoad{
		CC:   cc,
		User: user,
	}
}
