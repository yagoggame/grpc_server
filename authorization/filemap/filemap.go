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

package filemap

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"

	"github.com/yagoggame/grpc_server/authorization"
	"github.com/yagoggame/grpc_server/authorization/filemap/json"
	"github.com/yagoggame/grpc_server/interfaces"
)

var (
	// ErrNotImpl error occurs when file name extension not recognized
	ErrNotImpl = errors.New("not Implemented")
	// ErrFillUsers  error occurs when failed to fill users onauthorizator creation
	ErrFillUsers = errors.New("cant't fill users")
	// ErrStoreUsers  error occurs when failed to store users on disk
	ErrStoreUsers = errors.New("cant't store users")
)

// FileMaper wraps Load, Save methods
type FileMaper interface {
	Load(io.Reader) (map[string]*authorization.User, error)
	Save(map[string]*authorization.User, io.Writer) error
}

// Authorizator implements interfaces.Authorizator interface
type Authorizator struct {
	fileName string
	maper    FileMaper
	users    map[string]*authorization.User
}

// New constructs new Authorizator
func New(fileName string) (*Authorizator, error) {
	maper, err := choseMaper(fileName)
	if err != nil {
		return nil, err
	}

	authorizator := &Authorizator{
		fileName: fileName,
		maper:    maper,
		users:    nil,
	}

	if err := authorizator.fillUsers(); err != nil {
		return nil, err
	}
	return authorizator, nil
}

// Authorize attempts to authorize a user and returns the id if success
func (authorizator *Authorizator) Authorize(requisites *interfaces.Requisites) (id int, err error) {
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
	user, ok := authorizator.users[requisites.Login]
	if ok {
		return interfaces.ErrLoginOccupied
	}

	user = &authorization.User{
		Password: requisites.Password,
		ID:       getFirstVacantID(authorizator.users),
	}
	authorizator.users[requisites.Login] = user

	err := authorizator.storeUsers()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUsers, err)
	}

	log.Printf("client has been registred: %s, %d", requisites.Login, user.ID)
	return nil
}

// Remove attempts to remove a user and returns the id if success
func (authorizator *Authorizator) Remove(requisites *interfaces.Requisites) error {
	user, ok := authorizator.users[requisites.Login]
	if !ok {
		return interfaces.ErrLogin
	}
	if requisites.Password != user.Password {
		return interfaces.ErrPassword
	}

	delete(authorizator.users, requisites.Login)

	err := authorizator.storeUsers()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUsers, err)
	}

	log.Printf("client removed: %s", requisites.Login)
	return nil
}

// ChangeRequisites changes requisites of user from requisitesOld to requisitesNew
func (authorizator *Authorizator) ChangeRequisites(requisitesOld, requisitesNew *interfaces.Requisites) error {
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

	err := authorizator.storeUsers()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUsers, err)
	}

	log.Printf("requisites changed from: %v to %v", requisitesOld, requisitesNew)
	return nil
}

// Len returns number of users
func (authorizator *Authorizator) Len() int {
	return len(authorizator.users)
}

func (authorizator *Authorizator) fillUsers() error {
	err := authorizator.loadUsers()
	if err == nil {
		return nil
	}

	authorizator.users = map[string]*authorization.User{
		"Joe":  &authorization.User{Password: "aaa", ID: 2},
		"Nick": &authorization.User{Password: "bbb", ID: 3},
	}

	errS := authorizator.storeUsers()
	if errS == nil {
		return nil
	}

	return fmt.Errorf("%w: default users store failed: %v,  after load failed: %v", ErrFillUsers, errS, err)
}

func (authorizator *Authorizator) loadUsers() error {
	file, err := os.Open(authorizator.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	users, err := authorizator.maper.Load(file)
	if err != nil {
		return err
	}
	authorizator.users = users
	return nil
}

func (authorizator *Authorizator) storeUsers() error {
	file, err := os.Create(authorizator.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = authorizator.maper.Save(authorizator.users, file)
	if err != nil {
		return err
	}
	return nil
}

func choseMaper(fileName string) (FileMaper, error) {
	ext := path.Ext(fileName)
	switch ext {
	case ".json":
		return json.New(), nil
	}
	return nil, ErrNotImpl
}

func getFirstVacantID(users map[string]*authorization.User) (id int) {
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
