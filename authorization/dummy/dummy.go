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

// provides dummy realization of server.Authorizator interface
package dummy

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User contains user attributes
type User struct {
	Password string
	Id       int
}

// Authorizator implements server.Authorizator interface
type Authorizator map[string]User

// New constructs new Authorizator
func New() Authorizator {
	return map[string]User{
		"Joe":  User{"aaa", 1},
		"Nick": User{"bbb", 2},
	}
}

// Authorize attempts to authorize user and returns the id if success
func (users Authorizator) Authorize(login, password string) (id int, err error) {
	usr, ok := users[login]
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "unknown user %s", login)
	}

	if password != usr.Password {
		return 0, status.Errorf(codes.Unauthenticated, "bad password %s", password)
	}
	log.Printf("authenticated client: %s", login)
	return usr.Id, nil
}
