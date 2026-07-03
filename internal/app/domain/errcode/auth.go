package errcode

const (
	AuthInvalidCredentials Code = "auth.credential.invalid"
	AuthUnauthorized       Code = "auth.unauthorized"
	AuthPermissionDenied   Code = "auth.permission.denied"
	AuthForbidden          Code = "auth.forbidden"

	AuthLoginDenied Code = "auth.login.denied"

	AuthTokenExpired Code = "auth.token.expired"
	AuthTokenInvalid Code = "auth.token.invalid"
)
