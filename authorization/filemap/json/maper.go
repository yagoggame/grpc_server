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

package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/yagoggame/grpc_server/authorization"
)

var (
	// ErrCOrruptedUser occurs when trying to load a corrupted user
	ErrCOrruptedUser = errors.New("Loading of corrupted user")
	// ErrDecode occurs when decoding fails
	ErrDecode = errors.New("Loading of corrupted user")
)

type dbItem struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	ID       int    `json:"id"`
}

func (item *dbItem) isUserCorrupted() bool {
	return len(item.Login) < 1 || len(item.Password) < 1 || item.ID < 1
}

// Maper implements FileMaper interface for json files
type Maper struct{}

// New creates Maper fo file string
func New() *Maper {
	return &Maper{}
}

// Save stores users elements into writer in json format
func (*Maper) Save(users map[string]*authorization.User, writer io.Writer) error {
	items := users2Items(users)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder.Encode(items)
}

// Load reades users elements from writer in json format
func (*Maper) Load(reader io.Reader) (map[string]*authorization.User, error) {
	var items []dbItem
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&items)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrDecode, err)
	}

	return items2Users(items)
}

func users2Items(users map[string]*authorization.User) []dbItem {
	items := make([]dbItem, len(users))
	i := 0
	for login, user := range users {
		items[i] = dbItem{Login: login, Password: user.Password, ID: user.ID}
		i++
	}
	return items
}

func items2Users(items []dbItem) (map[string]*authorization.User, error) {
	users := make(map[string]*authorization.User, len(items))
	for _, item := range items {
		if (&item).isUserCorrupted() {
			return nil, fmt.Errorf("%w, %v", ErrCOrruptedUser, item)
		}

		users[item.Login] = &authorization.User{Password: item.Password, ID: item.ID}
	}
	return users, nil
}
