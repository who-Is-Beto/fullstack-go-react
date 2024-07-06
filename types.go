package main

import "math/rand"


type User struct {
	ID       int    `json:"id"`
	Username string  `json:"username"`
}

func NewAccount(username string) *User {
	return &User{
		ID: rand.Intn(10000),
		Username: username,
	}
}