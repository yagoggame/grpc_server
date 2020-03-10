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
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/yagoggame/grpc_server/authorization"
)

var twoUsers = map[string]*authorization.User{
	"Joe":  &authorization.User{Password: "aaa", ID: 2},
	"Nick": &authorization.User{Password: "bbb", ID: 3},
}

var twoJSON = `[
	{
		"login": "Joe",
		"password": "aaa",
		"id": 2
	},
	{
		"login": "Nick",
		"password": "bbb",
		"id": 3
	}
]
`

var oneUser = map[string]*authorization.User{
	"Joe": &authorization.User{Password: "aaa", ID: 2},
}

var oneJSON = `[
	{
		"login": "Joe",
		"password": "aaa",
		"id": 2
	}
]
`

var noUsers = map[string]*authorization.User{}

var noJSON = `[]
`

var errKeysJSON = `[
	{
		"logAn": "Joe",
		"paFFword": "aaa",
		"id": 2
	},
	{
		"login": "Nick",
		"password": "bbb",
		"id": 3
	}
]
`
var errFmtJSON = `[
	{
		"login": "Joe",
		"password": "aaa",
		"id": "ERR"
	},
	{
		"login": "Nick",
		"password": "bbb",
		"id": 3
	}
]
`

var encodeTests = []struct {
	name     string
	users    map[string]*authorization.User
	wantJSON string
	wantErr  error
}{
	{
		name:     "two users",
		users:    twoUsers,
		wantJSON: twoJSON,
		wantErr:  nil,
	},
	{
		name:     "one user",
		users:    oneUser,
		wantJSON: oneJSON,
		wantErr:  nil,
	},
	{
		name:     "no users",
		users:    noUsers,
		wantJSON: noJSON,
		wantErr:  nil,
	},
}

var decodeTests = []struct {
	name      string
	jsonValue string
	wantUsers map[string]*authorization.User
	wantErr   error
}{
	{
		name:      "two users",
		jsonValue: twoJSON,
		wantUsers: twoUsers,
		wantErr:   nil,
	},
	{
		name:      "one user",
		jsonValue: oneJSON,
		wantUsers: oneUser,
		wantErr:   nil,
	},
	{
		name:      "no users",
		jsonValue: noJSON,
		wantUsers: noUsers,
		wantErr:   nil,
	},
	{
		name:      "err keys json",
		jsonValue: errKeysJSON,
		wantUsers: nil,
		wantErr:   ErrCOrruptedUser,
	},
	{
		name:      "err format json",
		jsonValue: errFmtJSON,
		wantUsers: nil,
		wantErr:   ErrDecode,
	},
}

func TestSave(t *testing.T) {
	maper := New()
	for _, test := range encodeTests {
		t.Run(test.name, func(t *testing.T) {
			writer := &strings.Builder{}

			err := maper.Save(test.users, writer)
			if !errors.Is(err, test.wantErr) {
				t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", test.wantErr, err)
			}
			if writer.String() != test.wantJSON {
				t.Errorf("Unexpected json:\nwant: %q,\n got: %q", test.wantJSON, writer.String())
			}
		})
	}
}

func TestLoad(t *testing.T) {
	maper := New()
	for _, test := range decodeTests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.jsonValue)

			users, err := maper.Load(reader)
			if !errors.Is(err, test.wantErr) {
				t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", test.wantErr, err)
			}

			if !reflect.DeepEqual(users, test.wantUsers) {
				t.Errorf("Unexpected users:\nwant: %v,\n got: %v", test.wantUsers, users)
			}
		})
	}
}
