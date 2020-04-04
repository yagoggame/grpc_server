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

// Package postgres provides postgres realization of interfaces.Authorizator interface
package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	// registers pgx postgresSQL driver
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/yagoggame/grpc_server/interfaces"
)

var (
	// ErrWrongAffectedRows occures if some request affects on unexpected number of rows
	ErrWrongAffectedRows = errors.New("unexpected number of rows affected")
	// ErrModificationResult occures when request of modification produses strange result
	ErrModificationResult = errors.New("data changing produsec strange result")
)

// ConnectionData struct stores all database requisites
type ConnectionData struct {
	Host     string
	Port     int
	DBname   string
	User     string
	Password string
}

// Authorizator implements interfaces.Authorizator interface
type Authorizator struct {
	db *sql.DB
}

// NewWithDB constructs new Authorizator
// This approach provided for testing purpose
func NewWithDB(db *sql.DB) *Authorizator {
	return &Authorizator{db: db}
}

// NewPgx constructs new Authorizator with underlying pgx interface.
func NewPgx(conData *ConnectionData) (*Authorizator, error) {
	connectionString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s",
		conData.Host, conData.Port, conData.DBname, conData.User, conData.Password)

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	return &Authorizator{db: db}, nil
}

// Close closes underlying database connection - not nececcary
func (authorizator *Authorizator) Close() error {
	return authorizator.db.Close()
}

// Authorize attempts to authorize a user and returns the id if success
func (authorizator *Authorizator) Authorize(requisites *interfaces.Requisites) (id int, err error) {
	var password string
	err = authorizator.db.QueryRow("SELECT id, password FROM users WHERE username = $1 LIMIT 1", requisites.Login).Scan(&id, &password)
	if err := checkAuthorization(password, requisites.Password, err); err != nil {
		return 0, err
	}

	return id, nil
}

// Register attempts to register a new user and returns the id if success
func (authorizator *Authorizator) Register(requisites *interfaces.Requisites) (err error) {
	tx, err := authorizator.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = closeTransaction(tx, err)
	}()

	var id int
	err = tx.QueryRow("SELECT id FROM users WHERE username = $1 LIMIT 1", requisites.Login).Scan(&id)
	if err != sql.ErrNoRows {
		if err != nil {
			return err
		}
		return interfaces.ErrLoginOccupied
	}

	err = tx.QueryRow("INSERT INTO users (id,username,password) VALUES(DEFAULT,$1,$2) RETURNING id",
		requisites.Login, requisites.Password).Scan(&id)
	if err := checkModification(id, err); err != nil {
		return err
	}

	return nil
}

// Remove attempts to remove a user and returns the id if success
func (authorizator *Authorizator) Remove(requisites *interfaces.Requisites) (err error) {
	tx, err := authorizator.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = closeTransaction(tx, err)
	}()

	var (
		password string
		id       int
	)
	err = tx.QueryRow("SELECT id, password FROM users WHERE username = $1 LIMIT 1", requisites.Login).Scan(&id, &password)
	if err := checkAuthorization(password, requisites.Password, err); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM users WHERE id=$1", id)
	if err := checkResult(result, err); err != nil {
		return err
	}

	return nil
}

// ChangeRequisites changes requisites of user from requisitesOld to requisitesNew
func (authorizator *Authorizator) ChangeRequisites(requisitesOld, requisitesNew *interfaces.Requisites) (err error) {
	tx, err := authorizator.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = closeTransaction(tx, err)
	}()

	rows, err := tx.Query("SELECT id, username, password FROM users WHERE username IN ($1,$2)", requisitesOld.Login, requisitesNew.Login)
	id, err := processRows(rows, err, requisitesOld, requisitesNew)
	if err != nil {
		return err
	}

	result, err := tx.Exec("UPDATE users SET username=$1,password=$2 WHERE id=$3", requisitesNew.Login, requisitesNew.Password, id)
	if err := checkResult(result, err); err != nil {
		return err
	}

	return nil
}

func processRows(rows *sql.Rows, err error, requisitesOld, requisitesNew *interfaces.Requisites) (int, error) {
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, interfaces.ErrLogin
		}
		return 0, err
	}
	defer rows.Close()

	var (
		password, username string
		id, tmpid          int
	)

	for rows.Next() {
		err := rows.Scan(&tmpid, &username, &password)
		if err != nil {
			return 0, err
		}
		if username == requisitesNew.Login && requisitesNew.Login != requisitesOld.Login {
			return 0, interfaces.ErrLoginOccupied
		}

		if username == requisitesOld.Login {
			if requisitesOld.Password != password {
				return 0, interfaces.ErrPassword
			}
			id = tmpid
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, err
	}

	if id < 1 {
		return 0, interfaces.ErrLogin
	}
	return id, nil
}

func checkResult(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: unexpected number of rows affected: %d", ErrModificationResult, rowsAffected)
	}
	return nil
}

func closeTransaction(tx *sql.Tx, err error) error {
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			err = fmt.Errorf("%w: %q", err, rbErr)
		}
		return err
	}

	cmtErr := tx.Commit()
	if cmtErr != nil {
		err = cmtErr
	}
	return err
}

func checkAuthorization(passwordIn, passwordRef string, err error) error {
	if err == sql.ErrNoRows {
		return interfaces.ErrLogin
	}
	if err != nil {
		return err
	}

	if passwordIn != passwordRef {
		return interfaces.ErrPassword
	}
	return nil
}

func checkModification(id int, err error) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: no rows affected", ErrModificationResult)
	}
	if err != nil {
		return err
	}
	if id < 1 {
		return fmt.Errorf("%w: affected row with strange id %d", ErrModificationResult, id)
	}
	return nil
}
