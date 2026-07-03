package domain

import (
	"slices"
	"time"
)

type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	id string,
	name string,
	email string,
	createdAt time.Time,
	updatedAt time.Time,
) User {
	return User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

type UserWithPassword struct {
	User
	HashedPassword string
}

func NewUserWithPassword(
	id string,
	name string,
	email string,
	hashedPassword string,
	createdAt time.Time,
	updatedAt time.Time,
) UserWithPassword {
	return UserWithPassword{
		User: User{
			ID:        id,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		HashedPassword: hashedPassword,
	}
}

type UserEagerLoad struct {
	User User

	Roles []Role
}

func (u UserEagerLoad) HasRoles(roles ...Role) bool {
	for _, role := range roles {
		if slices.Contains(u.Roles, role) {
			return true
		}
	}
	return false
}

func NewUserEagerLoad(user User, roles []Role) UserEagerLoad {
	return UserEagerLoad{
		User:  user,
		Roles: roles,
	}
}

type Userinfo struct {
	User UserEagerLoad

	UserTokenInfo UserTokenInfo
}

func NewUserinfo(
	user UserEagerLoad,
	userTokenInfo UserTokenInfo,
) Userinfo {
	return Userinfo{
		User:          user,
		UserTokenInfo: userTokenInfo,
	}
}
