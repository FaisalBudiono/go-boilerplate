package domain

type UserBasicInfo struct {
	ID string
}

func NewUserBasicInfo(id string) UserBasicInfo {
	return UserBasicInfo{
		ID: id,
	}
}

type User struct {
	ID          string
	Name        string
	PhoneNumber string
	Email       string
	Password    string
	Roles       []Role
}

func NewUser(
	id string,
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
