// Copyright ©2020 BlinnikovAA. All rights reserved.
// This file is part of yagogame.
//
// yagogame is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// yagogame is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with yagogame.  If not, see <https://www.gnu.org/licenses/>.

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

//checkClientEntry checks user registration
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
