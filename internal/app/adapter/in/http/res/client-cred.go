package res

import (
	"time"

	"komdigi-immigration/internal/app/domain"
)

type clientCred struct {
	ID        string `json:"id"`
	UserID    string `json:"userID"`
	ClientID  string `json:"clientID"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func newClientCred(cc domain.ClientCredential) clientCred {
	return clientCred{
		ID:        cc.ID,
		UserID:    cc.UserID,
		ClientID:  cc.ClientID,
		CreatedAt: cc.CreatedAt.Format(time.RFC3339),
		UpdatedAt: cc.UpdatedAt.Format(time.RFC3339),
	}
}

type clientCredEagerLoad struct {
	clientCred

	User user `json:"user"`
}

func newClientCredEagerLoad(
	cc domain.ClientCredentialEagerLoad,
) clientCredEagerLoad {
	return clientCredEagerLoad{
		clientCred: newClientCred(cc.CC),
		User:       newUser(cc.User),
	}
}

type clientCredSecret struct {
	clientCred
	ClientSecret string `json:"clientSecret"`
}

func newClientCredSecret(cc domain.ClientCredentialSecret) clientCredSecret {
	return clientCredSecret{
		clientCred:   newClientCred(cc.CC),
		ClientSecret: cc.Secret,
	}
}

type clientCredSecretEagerLoad struct {
	clientCredSecret

	User user `json:"user"`
}

func newClientCredSecretEagerLoad(
	cc domain.ClientCredentialSecretEagerLoad,
) clientCredSecretEagerLoad {
	return clientCredSecretEagerLoad{
		clientCredSecret: newClientCredSecret(cc.CC),
		User:             newUser(cc.User),
	}
}

func ClientCredSecret(cc domain.ClientCredentialSecretEagerLoad) response[clientCredSecretEagerLoad] {
	return response[clientCredSecretEagerLoad]{
		Data: newClientCredSecretEagerLoad(cc),
	}
}

func ClientCreds(
	ccs []domain.ClientCredentialEagerLoad, pg domain.Pagination,
) responsePaginated[clientCredEagerLoad] {
	res := make([]clientCredEagerLoad, len(ccs))
	for i, cc := range ccs {
		res[i] = newClientCredEagerLoad(cc)
	}

	return responsePaginated[clientCredEagerLoad]{
		response: response[[]clientCredEagerLoad]{
			Data: res,
		},
		Meta: newPaginatedMeta(pg),
	}
}
