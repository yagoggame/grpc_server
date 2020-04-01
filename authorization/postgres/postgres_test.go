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

package postgres_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/yagoggame/grpc_server/authorization/postgres"
	"github.com/yagoggame/grpc_server/interfaces"
)

type iderr struct {
	id  int
	err error
}

type commonTestCase struct {
	name           string
	userRequisites *interfaces.Requisites
	want           iderr
	retRows        []*sqlmock.Rows
	retErr         error
}

var authorizeTests = []*commonTestCase{
	&commonTestCase{
		name: "authorized user",
		userRequisites: &interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: iderr{id: 1, err: nil},
		retRows: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, "aaa")},
	},
	&commonTestCase{
		name: "wrong password",
		userRequisites: &interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: iderr{id: 0, err: interfaces.ErrPassword},
		retRows: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, "bbb")},
	},
	&commonTestCase{
		name: "login not found",
		userRequisites: &interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		retErr:  sql.ErrNoRows,
		want:    iderr{id: 0, err: interfaces.ErrLogin},
		retRows: []*sqlmock.Rows{},
	},
	&commonTestCase{
		name: "some request error",
		userRequisites: &interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		retErr:  sql.ErrTxDone,
		want:    iderr{id: 0, err: sql.ErrTxDone},
		retRows: []*sqlmock.Rows{},
	},
}

var registerTests = []*commonTestCase{
	&commonTestCase{
		name: "success",
		userRequisites: &interfaces.Requisites{
			Login:    "Joe",
			Password: "aaa",
		},
		want: iderr{id: 1, err: nil},
		retRows: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, "aaa")},
		retErr: nil,
	},
}

func TestNewPgx(t *testing.T) {
	conData := &postgres.ConnectionData{
		Host:     "localhost",
		Port:     5432,
		DBname:   "postgres",
		User:     "authentificator",
		Password: "authentificator",
	}

	db, err := postgres.NewPgx(conData)
	if err != nil {
		t.Fatalf("Unexpected NewPgx() err: %q", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Unexpected db.Close() err: %q", err)
	}
}

func TestAuthorize(t *testing.T) {
	for _, test := range authorizeTests {
		t.Run(test.name, func(t *testing.T) {
			performAuthorizeTest(t, test)
		})
	}
}

func performAuthorizeTest(t *testing.T, test *commonTestCase) {
	authorizator, mock := initMock(t)
	defer authorizator.Close()

	mock.ExpectQuery("select id, password from users where name = \\? limit 1").
		WithArgs(test.userRequisites.Login).
		WillReturnRows(test.retRows...).
		WillReturnError(test.retErr)

	id, err := authorizator.Authorize(test.userRequisites)

	testIDErr(t, test.want, iderr{id: id, err: err})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRegister(t *testing.T) {
	for _, test := range registerTests {
		t.Run(test.name, func(t *testing.T) {
			performAuthorizeTest(t, test)
		})
	}
}

func performRegisterTest(t *testing.T, test *commonTestCase) {
	authorizator, mock := initMock(t)
	defer authorizator.Close()

	// rows := sqlmock.NewRows([]string{"id", "password"}).
	// 	AddRow(test.want.id, test.actualPassword)
	// mock.ExpectQuery("select id, password from users where name = \\? limit 1").
	// 	WithArgs(test.userRequisites.Login).
	// 	WillReturnRows(rows)

	id, err := authorizator.Register(test.userRequisites)

	testIDErr(t, test.want, iderr{id: id, err: err})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func initMock(t *testing.T) (*postgres.Authorizator, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}
	authorizator := postgres.NewWithDB(db)

	return authorizator, mock
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
