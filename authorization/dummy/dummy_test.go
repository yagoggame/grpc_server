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
	"testing"

	server "github.com/yagoggame/grpc_server"
)

type iderr struct {
	id  int
	err error
}

var tests = []struct {
	caseName string
	login    string
	password string
	want     iderr
}{
	{
		caseName: "correct",
		login:    "Joe", password: "aaa",
		want: iderr{id: 1, err: nil}},
	{
		caseName: "wrong login",
		login:    "JJooee", password: "aaa",
		want: iderr{id: 0, err: server.ErrLogin}},
	{
		caseName: "wrong password",
		login:    "Joe", password: "ababab",
		want: iderr{id: 0, err: server.ErrPassword}},
}

func TestMain(t *testing.T) {
	authorizator := New()
	for _, test := range tests {
		t.Run(test.caseName, func(t *testing.T) {

		})
		id, err := authorizator.Authorize(test.login, test.password)
		if id != test.want.id {
			t.Errorf("Unexpected id:\nwant: %d,\ngot: %d.", test.want.id, id)
		}
		if err != test.want.err {
			t.Errorf("Unexpected id:\nwant: %v,\ngot: %v.", test.want.err, err)
		}
	}
}
