package main

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type user struct {
	Password, Id string
}

//temporary user holder
var users map[string]user = map[string]user{
	"Joe":  user{"aaa", "1"},
	"Nick": user{"bbb", "2"},
}

//checkClientEntry - checks user registration
func checkClientEntry(login, password string) (id string, err error) {
	usr, ok := users[login]
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "unknown user %s", login)
	}

	if password != usr.Password {
		return "", status.Errorf(codes.Unauthenticated, "bad password %s", password)
	}
	log.Printf("authenticated client: %s", login)
	return usr.Id, nil
}
