package domain

import "FaisalBudiono/go-boilerplate/internal/domain/domid"

type UserTokenInfo struct {
	ID domid.UserID
}

func NewUserBasicInfo(id domid.UserID) UserTokenInfo {
	return UserTokenInfo{
		ID: id,
	}
}

type User struct {
	ID          domid.UserID
	Name        string
	PhoneNumber string
	Email       string
	Password    string
	Roles       []Role
}

func NewUser(
	id domid.UserID,
	name string,
	phoneNumber string,
	email string,
	password string,
	roles []Role,
) User {
	return User{
		ID:          id,
		Name:        name,
		PhoneNumber: phoneNumber,
		Email:       email,
		Password:    password,
		Roles:       roles,
	}
}
