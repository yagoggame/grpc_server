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

// Package dummy provides dummy realization of server.Authorizator interface
package dummy

import (
	"log"
	"sort"

	server "github.com/yagoggame/grpc_server"
)

// User contains user attributes
type User struct {
	Password string
	ID       int
}

// Authorizator implements server.Authorizator interface
type Authorizator map[string]*User

// New constructs new Authorizator
func New() Authorizator {
	return map[string]*User{
		"Joe":  &User{Password: "aaa", ID: 2},
		"Nick": &User{Password: "bbb", ID: 3},
	}
}

// Authorize attempts to authorize a user and returns the id if success
func (users Authorizator) Authorize(requisites *server.Requisites) (id int, err error) {
	user, ok := users[requisites.Login]
	if !ok {
		return 0, server.ErrLogin
	}
	if requisites.Password != user.Password {
		return 0, server.ErrPassword
	}

	log.Printf("authenticated client: %s, %d", requisites.Login, user.ID)
	return user.ID, nil
}

// Register attempts to register a new user and returns the id if success
func (users Authorizator) Register(requisites *server.Requisites) (id int, err error) {
	user, ok := users[requisites.Login]
	if ok {
		return 0, server.ErrLoginOccupied
	}

	user = &User{
		Password: requisites.Password,
		ID:       users.getFirstVacantID(),
	}
	users[requisites.Login] = user

	log.Printf("client has been registred: %s, %d", requisites.Login, user.ID)
	return user.ID, nil
}

// Remove attempts to remove a user and returns the id if success
func (users Authorizator) Remove(requisites *server.Requisites) (id int, err error) {
	user, ok := users[requisites.Login]
	if !ok {
		return 0, server.ErrLogin
	}
	if requisites.Password != user.Password {
		return 0, server.ErrPassword
	}

	delete(users, requisites.Login)

	log.Printf("client removed: %s", requisites.Login)
	return user.ID, nil
}

// ChangeRequisites changes requisites of user from requisitesOld to requisitesNew
func (users Authorizator) ChangeRequisites(requisitesOld, requisitesNew *server.Requisites) (id int, err error) {
	user, ok := users[requisitesOld.Login]
	if !ok {
		return 0, server.ErrLogin
	}
	if requisitesOld.Password != user.Password {
		return 0, server.ErrPassword
	}

	if _, ok := users[requisitesNew.Login]; ok {
		return 0, server.ErrLoginOccupied
	}

	users[requisitesNew.Login] = users[requisitesOld.Login]
	delete(users, requisitesOld.Login)
	users[requisitesNew.Login].Password = requisitesNew.Password
	log.Printf("requisites changed from: %v to %v", requisitesOld, requisitesNew)
	return user.ID, nil
}

// Len returns number of users
func (users Authorizator) Len() int {
	return len(users)
}

func (users Authorizator) getFirstVacantID() (id int) {
	ids := make([]int, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	sort.Ints(ids)

	var i int
	for i = range ids {
		if i+1 != ids[i] {
			return i + 1
		}
	}
	return i + 2
}
