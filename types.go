package main

type CreateUserRequest struct {
	Sub string `json:"sub"`
}

type User struct {
	Sub string `json:"sub"`
}

func NewUser(sub string) *User {
	return &User{
		Sub: sub,
	}
}
