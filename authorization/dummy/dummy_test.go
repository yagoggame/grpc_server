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

package dummy

import (
	"io/ioutil"
	"log"
	"testing"

	server "github.com/yagoggame/grpc_server"
)

type iderr struct {
	id  int
	err error
}

var testsAuthorize = []struct {
	caseName   string
	requisites server.Requisites
	want       iderr
}{
	{
		caseName: "unregistred login",
		requisites: server.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: iderr{id: 0, err: server.ErrLogin},
	},
	{
		caseName: "registred login wrong password",
		requisites: server.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		want: iderr{id: 0, err: server.ErrPassword},
	},
	{
		caseName: "registred login",
		requisites: server.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: iderr{id: 2, err: nil},
	},
	{
		caseName: "registred login",
		requisites: server.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		want: iderr{id: 3, err: nil},
	},
}

var testsRemove = []struct {
	caseName   string
	requisites server.Requisites
	want       error
}{
	{
		caseName: "unregistred login",
		requisites: server.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: server.ErrLogin,
	},
	{
		caseName: "registred login wrong password",
		requisites: server.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		want: server.ErrPassword,
	},
	{
		caseName: "registred login",
		requisites: server.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: nil,
	},
	{
		caseName: "registred login",
		requisites: server.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		want: nil,
	},
}

var testsRegister = []struct {
	caseName   string
	requisites server.Requisites
	want       error
}{
	{
		caseName: "registred login",
		requisites: server.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: server.ErrLoginOccupied},
	{
		caseName: "unregistred login first",
		requisites: server.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		want: nil},
	{
		caseName: "unregistred login second",
		requisites: server.Requisites{
			Login:    "Mike",
			Password: "mmm",
		},
		want: nil},
}

var testsChangeRequisites = []struct {
	caseName      string
	requisitesOld server.Requisites
	requisitesNew server.Requisites
	want          error
}{
	{
		caseName: "from unregistred login",
		requisitesOld: server.Requisites{
			Login:    "Piter",
			Password: "aaa",
		},
		requisitesNew: server.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: server.ErrLogin,
	},
	{
		caseName: "from registred login wrong password",
		requisitesOld: server.Requisites{
			Login:    "Joe",
			Password: "ababab",
		},
		requisitesNew: server.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: server.ErrPassword,
	},
	{
		caseName: "to registred",
		requisitesOld: server.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		requisitesNew: server.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		want: server.ErrLoginOccupied,
	},
	{
		caseName: "registred to unregistred",
		requisitesOld: server.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		requisitesNew: server.Requisites{
			Login:    "Teodor",
			Password: "ttt",
		},
		want: nil,
	},
	{
		caseName: "only change password",
		requisitesOld: server.Requisites{
			Login:    "Nick",
			Password: "bbb",
		},
		requisitesNew: server.Requisites{
			Login:    "Nick",
			Password: "ccc",
		},
		want: nil,
	},
}

func TestAuthorize(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator := New()
	for _, test := range testsAuthorize {
		t.Run(test.caseName, func(t *testing.T) {
			id, err := authorizator.Authorize(&test.requisites)

			testIDErr(t, test.want, iderr{id: id, err: err})
		})
	}
}

func TestRegister(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator := New()

	for _, test := range testsRegister {
		t.Run(test.caseName, func(t *testing.T) {
			err := authorizator.Register(&test.requisites)

			testErr(t, test.want, err)
		})
	}
}

func TestRemove(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator := New()

	for _, test := range testsRemove {
		t.Run(test.caseName, func(t *testing.T) {
			usersLen := authorizator.Len()
			err := authorizator.Remove(&test.requisites)

			testErr(t, test.want, err)

			if err != nil && usersLen != authorizator.Len() {
				t.Errorf("Unexpected count of user delta:\nwant: 0\ngot: %d.", authorizator.Len()-usersLen)
			} else if err == nil && usersLen == authorizator.Len() {
				t.Errorf("Unexpected count of user delta:\nwant: -1\ngot: %d.", authorizator.Len()-usersLen)
			}

			_, authErr := authorizator.Authorize(&test.requisites)
			if err == nil && authErr != server.ErrLogin {
				t.Errorf("Unexpected Authorize err:\nwant: %v\ngot: %v.", server.ErrLogin, authErr)
			}

		})
	}
}

func TestChangeRequisites(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	authorizator := New()
	for _, test := range testsChangeRequisites {
		t.Run(test.caseName, func(t *testing.T) {
			err := authorizator.ChangeRequisites(&test.requisitesOld, &test.requisitesNew)

			testErr(t, test.want, err)

			user, ok := authorizator[test.requisitesNew.Login]
			if err == nil {
				if !ok {
					t.Fatalf("Unexpected behavior of authorizator:\nwant: login changed\n:got: user with new login not found.")
				}
				if user.Password != test.requisitesNew.Password {
					t.Fatalf("Unexpected user's password:\nwant: %q\n:got: %q.", test.requisitesNew.Password, user.Password)
				}
				if _, ok := authorizator[test.requisitesOld.Login]; ok && test.requisitesOld.Login != test.requisitesNew.Login {
					t.Fatalf("Unexpected behavior of authorizator:\nwant: user with old login not found\n:got: user with old login found.")
				}
			}
		})
	}
}

func testIDErr(t *testing.T, want, got iderr) {
	if got.id != want.id {
		t.Errorf("Unexpected id:\nwant: %d,\ngot: %d.", want.id, got.id)
	}
	if got.err != want.err {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want.err, got.err)
	}
}

func testErr(t *testing.T, want, got error) {
	if got != want {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want, got)
	}
}
