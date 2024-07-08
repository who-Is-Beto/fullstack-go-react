package main

import (
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Username string  `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type createUserRequest struct {
	Username string `json:"username"`
}

func NewAccount(username string) *User {
	return &User{
		Username: username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}