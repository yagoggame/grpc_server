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

// Package dummy provides dummy realization of interfaces.Authorizator interface
package dummy

import (
	"log"
	"sort"
	"sync"

	"github.com/yagoggame/grpc_server/authorization"
	"github.com/yagoggame/grpc_server/interfaces"
)

// Authorizator implements interfaces.Authorizator interface
type Authorizator struct {
	users map[string]*authorization.User
	mutex sync.RWMutex
}

// New constructs new Authorizator
func New() *Authorizator {
	return &Authorizator{
		users: map[string]*authorization.User{
			"Joe":  &authorization.User{Password: "aaa", ID: 2},
			"Nick": &authorization.User{Password: "bbb", ID: 3},
		},
	}
}

// Authorize attempts to authorize a user and returns the id if success
func (authorizator *Authorizator) Authorize(requisites *interfaces.Requisites) (id int, err error) {
	authorizator.mutex.RLock()
	defer authorizator.mutex.RUnlock()

	user, ok := authorizator.users[requisites.Login]
	if !ok {
		return 0, interfaces.ErrLogin
	}
	if requisites.Password != user.Password {
		return 0, interfaces.ErrPassword
	}

	log.Printf("authenticated client: %s, %d", requisites.Login, user.ID)
	return user.ID, nil
}

// Register attempts to register a new user and returns the id if success
func (authorizator *Authorizator) Register(requisites *interfaces.Requisites) error {
	authorizator.mutex.Lock()
	defer authorizator.mutex.Unlock()

	user, ok := authorizator.users[requisites.Login]
	if ok {
		return interfaces.ErrLoginOccupied
	}

	user = &authorization.User{
		Password: requisites.Password,
		ID:       authorizator.getFirstVacantID(),
	}
	authorizator.users[requisites.Login] = user

	log.Printf("client has been registred: %s, %d", requisites.Login, user.ID)
	return nil
}

// Remove attempts to remove a user and returns the id if success
func (authorizator *Authorizator) Remove(requisites *interfaces.Requisites) error {
	authorizator.mutex.Lock()
	defer authorizator.mutex.Unlock()

	user, ok := authorizator.users[requisites.Login]
	if !ok {
		return interfaces.ErrLogin
	}
	if requisites.Password != user.Password {
		return interfaces.ErrPassword
	}

	delete(authorizator.users, requisites.Login)

	log.Printf("client removed: %s", requisites.Login)
	return nil
}

// ChangeRequisites changes requisites of user from requisitesOld to requisitesNew
func (authorizator *Authorizator) ChangeRequisites(requisitesOld, requisitesNew *interfaces.Requisites) error {
	authorizator.mutex.Lock()
	defer authorizator.mutex.Unlock()

	user, ok := authorizator.users[requisitesOld.Login]
	if !ok {
		return interfaces.ErrLogin
	}
	if requisitesOld.Password != user.Password {
		return interfaces.ErrPassword
	}

	if requisitesNew.Login != requisitesOld.Login {
		if _, ok := authorizator.users[requisitesNew.Login]; ok {
			return interfaces.ErrLoginOccupied
		}

		authorizator.users[requisitesNew.Login] = authorizator.users[requisitesOld.Login]
		delete(authorizator.users, requisitesOld.Login)
	}
	authorizator.users[requisitesNew.Login].Password = requisitesNew.Password
	log.Printf("requisites changed from: %v to %v", requisitesOld, requisitesNew)
	return nil
}

// Len returns number of users
func (authorizator *Authorizator) Len() int {
	authorizator.mutex.RLock()
	defer authorizator.mutex.RUnlock()

	return len(authorizator.users)
}

func (authorizator *Authorizator) getFirstVacantID() (id int) {
	ids := make([]int, 0, len(authorizator.users))
	for _, user := range authorizator.users {
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
