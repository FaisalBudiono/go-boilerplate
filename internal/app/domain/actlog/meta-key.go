package actlog

type MetaKey string

func (mk MetaKey) String() string { return string(mk) }

const (
	MetaKeyClientID       MetaKey = "clientID"
	MetaKeyCreatedForID   MetaKey = "createdForID"
	MetaKeyCreatedForName MetaKey = "createdForName"

	MetaKeyReqIMEI MetaKey = "reqIMEI"

	MetaKeyFinalResponse MetaKey = "finalResponse"

	MetaKeyResultRaw    MetaKey = "resultRaw"
	MetaKeyResultStatus MetaKey = "resultStatus"

	MetaKeyResultPassportNumber MetaKey = "resultPassportNumber"
	MetaKeyResultCountryCode    MetaKey = "resultCountryCode"

	MetaKeyResultFirstName     MetaKey = "resultFirstName"
	MetaKeyResultLastName      MetaKey = "resultLastName"
	MetaKeyResultGender        MetaKey = "resultGender"
	MetaKeyResultPlaceOfBirth  MetaKey = "resultPlaceOfBirth"
	MetaKeyResultDateOfBirth   MetaKey = "resultDateOfBirth"
	MetaKeyResultIsInIndonesia MetaKey = "resultIsInIndonesia"
	MetaKeyResultIssuedAt      MetaKey = "resultIssuedAt"
	MetaKeyResultExpiredAt     MetaKey = "resultExpiredAt"
)
