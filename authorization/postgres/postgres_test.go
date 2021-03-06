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
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/yagoggame/grpc_server/authorization/postgres"
	"github.com/yagoggame/grpc_server/interfaces"
)

type iderr struct {
	id  int
	err error
}

var errSome = errors.New("some error")

var (
	joe = &interfaces.Requisites{
		Login:    "Joe",
		Password: "aaa",
	}
	joePas = &interfaces.Requisites{
		Login:    "Joe",
		Password: "bbb",
	}
	nick = &interfaces.Requisites{
		Login:    "Nick",
		Password: "bbb",
	}
)

type commonTestCase struct {
	name              string
	userRequisites    *interfaces.Requisites
	newUserRequisites *interfaces.Requisites
	want              iderr
	retRowsSel1       []*sqlmock.Rows
	retErrSel1        error
	retRowsSel2       []*sqlmock.Rows
	retErrSel2        error
	retResult2        sql.Result
	returnedID        int
}

var authorizeTests = []*commonTestCase{
	&commonTestCase{
		name:           "authorized user",
		userRequisites: joe,
		want:           iderr{id: 1, err: nil},
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password)},
	},
	&commonTestCase{
		name:           "wrong password",
		userRequisites: joe,
		want:           iderr{id: 0, err: interfaces.ErrPassword},
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password+"FFF")},
	},
	&commonTestCase{
		name:           "login not found",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		want:           iderr{id: 0, err: interfaces.ErrLogin},
		retRowsSel1:    []*sqlmock.Rows{},
	},
	&commonTestCase{
		name:           "some request error",
		userRequisites: joe,
		retErrSel1:     sql.ErrTxDone,
		want:           iderr{id: 0, err: sql.ErrTxDone},
		retRowsSel1:    []*sqlmock.Rows{},
	},
}

var registerTests = []*commonTestCase{
	&commonTestCase{
		name:           "some query error",
		userRequisites: joe,
		retErrSel1:     sql.ErrTxDone,
		want:           iderr{id: 0, err: sql.ErrTxDone},
		retRowsSel1:    []*sqlmock.Rows{},
	},
	&commonTestCase{
		name:           "login occupied",
		userRequisites: joe,
		retErrSel1:     nil,
		want:           iderr{id: 0, err: interfaces.ErrLoginOccupied},
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id"}).
				AddRow(1)},
	},
	&commonTestCase{
		name:           "some query2 error",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		retRowsSel1:    []*sqlmock.Rows{},
		retErrSel2:     sql.ErrTxDone,
		retRowsSel2:    []*sqlmock.Rows{},
		want:           iderr{id: 0, err: sql.ErrTxDone},
	},
	&commonTestCase{
		name:           "not registred",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		retRowsSel1:    []*sqlmock.Rows{},
		retErrSel2:     sql.ErrNoRows,
		retRowsSel2:    []*sqlmock.Rows{},
		want:           iderr{id: 0, err: postgres.ErrModificationResult},
	},
	&commonTestCase{
		name:           "strange id",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		retRowsSel1:    []*sqlmock.Rows{},
		retErrSel2:     nil,
		retRowsSel2: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id"}).
				AddRow(0)},
		want: iderr{id: 0, err: postgres.ErrModificationResult},
	},
	&commonTestCase{
		name:           "success",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		retRowsSel1:    []*sqlmock.Rows{},
		retErrSel2:     nil,
		retRowsSel2: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id"}).
				AddRow(1)},
		want: iderr{id: 1, err: nil},
	},
}

var removeTests = []*commonTestCase{
	&commonTestCase{
		name:           "some query error",
		userRequisites: joe,
		retErrSel1:     sql.ErrTxDone,
		retRowsSel1:    []*sqlmock.Rows{},
		want:           iderr{id: 0, err: sql.ErrTxDone},
	},
	&commonTestCase{
		name:           "login not found",
		userRequisites: joe,
		retErrSel1:     sql.ErrNoRows,
		retRowsSel1:    []*sqlmock.Rows{},
		want:           iderr{id: 0, err: interfaces.ErrLogin},
	},
	&commonTestCase{
		name:           "wrong password",
		userRequisites: joe,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password+"FFF")},
		want: iderr{id: 0, err: interfaces.ErrPassword},
	},
	&commonTestCase{
		name:           "some exec error",
		userRequisites: joe,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password)},
		retErrSel2: sql.ErrTxDone,
		returnedID: 1,
		retResult2: sqlmock.NewResult(0, 0),
		want:       iderr{id: 0, err: sql.ErrTxDone},
	},
	&commonTestCase{
		name:           "some RowsAffected error",
		userRequisites: joe,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewErrorResult(errSome),
		want:       iderr{id: 0, err: errSome},
	},
	&commonTestCase{
		name:           "some RowsAffected error",
		userRequisites: joe,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewResult(0, 0),
		want:       iderr{id: 0, err: postgres.ErrModificationResult},
	},
	&commonTestCase{
		name:           "success",
		userRequisites: joe,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "password"}).
				AddRow(1, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewResult(1, 1),
		want:       iderr{id: 1, err: nil},
	},
}

var changeRequisitesTests = []*commonTestCase{
	&commonTestCase{
		name:              "some query error",
		userRequisites:    joe,
		newUserRequisites: nick,
		retErrSel1:        sql.ErrTxDone,
		retRowsSel1:       []*sqlmock.Rows{},
		want:              iderr{err: sql.ErrTxDone},
	},
	&commonTestCase{
		name:              "login not found",
		userRequisites:    joe,
		newUserRequisites: nick,
		retErrSel1:        sql.ErrNoRows,
		retRowsSel1:       []*sqlmock.Rows{},
		want:              iderr{err: interfaces.ErrLogin},
	},
	&commonTestCase{
		name:              "wrong password",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password+"FFF")},
		want: iderr{err: interfaces.ErrPassword},
	},
	&commonTestCase{
		name:              "login occupied",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password).
				AddRow(2, nick.Login, nick.Password)},
		want: iderr{err: interfaces.ErrLoginOccupied},
	},
	&commonTestCase{
		name:              "some exec error",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password)},
		retErrSel2: sql.ErrTxDone,
		returnedID: 1,
		retResult2: sqlmock.NewResult(0, 0),
		want:       iderr{err: sql.ErrTxDone},
	},
	&commonTestCase{
		name:              "some RowsAffected error",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewErrorResult(errSome),
		want:       iderr{err: errSome},
	},
	&commonTestCase{
		name:              "some RowsAffected error",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewResult(0, 0),
		want:       iderr{err: postgres.ErrModificationResult},
	},
	&commonTestCase{
		name:              "chNamePas success",
		userRequisites:    joe,
		newUserRequisites: nick,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewResult(1, 1),
		want:       iderr{err: nil},
	},
	&commonTestCase{
		name:              "chPas success",
		userRequisites:    joe,
		newUserRequisites: joePas,
		retRowsSel1: []*sqlmock.Rows{
			sqlmock.NewRows([]string{"id", "username", "password"}).
				AddRow(1, joe.Login, joe.Password)},
		returnedID: 1,
		retResult2: sqlmock.NewResult(1, 1),
		want:       iderr{err: nil},
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

	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1 LIMIT 1").
		WithArgs(test.userRequisites.Login).
		WillReturnRows(test.retRowsSel1...).
		WillReturnError(test.retErrSel1)

	id, err := authorizator.Authorize(test.userRequisites)

	testIDErr(t, test.want, iderr{id: id, err: err})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRegister(t *testing.T) {
	for _, test := range registerTests {
		t.Run(test.name, func(t *testing.T) {
			performRegisterTest(t, test)
		})
	}
}

func performRegisterTest(t *testing.T, test *commonTestCase) {
	authorizator, mock := initMock(t)
	defer authorizator.Close()

	makeRegisterExpectations(mock, test)

	err := authorizator.Register(test.userRequisites)

	testErr(t, test.want.err, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func makeRegisterExpectations(mock sqlmock.Sqlmock, test *commonTestCase) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1 LIMIT 1").
		WithArgs(test.userRequisites.Login).
		WillReturnRows(test.retRowsSel1...).
		WillReturnError(test.retErrSel1)
	if test.retErrSel1 == sql.ErrNoRows {
		mock.ExpectQuery("INSERT INTO users \\(id,username,password\\) VALUES\\(DEFAULT,\\$1,\\$2\\)").
			WithArgs(test.userRequisites.Login, test.userRequisites.Password).
			WillReturnRows(test.retRowsSel2...).
			WillReturnError(test.retErrSel2)
	}
	if test.name == "success" {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
}

func TestRemove(t *testing.T) {
	for _, test := range removeTests {
		t.Run(test.name, func(t *testing.T) {
			performRemoveTest(t, test)
		})
	}
}

func performRemoveTest(t *testing.T, test *commonTestCase) {
	authorizator, mock := initMock(t)
	defer authorizator.Close()

	makeRemoveExpectations(mock, test)

	err := authorizator.Remove(test.userRequisites)

	testErr(t, test.want.err, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func makeRemoveExpectations(mock sqlmock.Sqlmock, test *commonTestCase) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1 LIMIT 1").
		WithArgs(test.userRequisites.Login).
		WillReturnRows(test.retRowsSel1...).
		WillReturnError(test.retErrSel1)
	if test.retErrSel1 == nil && test.name != "wrong password" {
		mock.ExpectExec("DELETE FROM users WHERE id=\\$1").
			WithArgs(test.returnedID).
			WillReturnResult(test.retResult2).
			WillReturnError(test.retErrSel2)
	}
	if test.name == "success" {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
}

func TestChangeRequisites(t *testing.T) {
	for _, test := range changeRequisitesTests {
		t.Run(test.name, func(t *testing.T) {
			performChangeRequisitesTest(t, test)
		})
	}
}

func performChangeRequisitesTest(t *testing.T, test *commonTestCase) {
	authorizator, mock := initMock(t)
	defer authorizator.Close()

	makeChangeRequisitesExpectations(mock, test)

	err := authorizator.ChangeRequisites(test.userRequisites, test.newUserRequisites)

	testErr(t, test.want.err, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func makeChangeRequisitesExpectations(mock sqlmock.Sqlmock, test *commonTestCase) {
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT id, username, password FROM users WHERE username IN \\(\\$1,\\$2\\)").
		WithArgs(test.userRequisites.Login, test.newUserRequisites.Login).
		WillReturnRows(test.retRowsSel1...).
		WillReturnError(test.retErrSel1)

	if test.retErrSel1 == nil && test.name != "wrong password" && test.name != "login occupied" {
		mock.ExpectExec("UPDATE users SET username=\\$1,password=\\$2 WHERE id=\\$3").
			WithArgs(test.newUserRequisites.Login, test.newUserRequisites.Password, test.returnedID).
			WillReturnResult(test.retResult2).
			WillReturnError(test.retErrSel2)
	}

	if test.name == "chNamePas success" || test.name == "chPas success" {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
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
	if !errors.Is(got.err, want.err) {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want.err, got.err)
	}
}
func testErr(t *testing.T, want, got error) {
	if !errors.Is(got, want) {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want, got)
	}
}
