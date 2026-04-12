package models

type UserInfo struct {
	ID    string
	Name  string
	Login string
}

type UserLogin struct {
	Email        string
	PasswordHash string
}

type UserReg struct {
	Email    string
	Login    string
	Name     string
	Password string
}
