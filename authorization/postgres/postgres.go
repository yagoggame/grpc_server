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
	err = authorizator.db.QueryRow("select id, password from users where name = ? limit 1", requisites.Login).Scan(&id, &password)

	if err := checkAuthorization(password, requisites.Password, err); err != nil {
		return 0, err
	}

	return id, nil
}

// Register attempts to register a new user and returns the id if success
func (authorizator *Authorizator) Register(requisites *interfaces.Requisites) (id int, err error) {
	tx, err := authorizator.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		err = closeTransaction(tx, err)
	}()

	err = tx.QueryRow("select id from users where name = ? limit 1", requisites.Login).Scan(&id)
	if err != sql.ErrNoRows {
		if err != nil {
			return 0, err
		}
		return 0, interfaces.ErrLoginOccupied
	}

	err = tx.QueryRow("INSERT INTO users (id,username,password) VALUES(DEFAULT,?,?) RETURNING id",
		requisites.Login, requisites.Password).Scan(&id)
	if err := checkModification(id, err); err != nil {
		return 0, err
	}

	return id, nil
}

// // Remove attempts to remove a user and returns the id if success
// func (authorizator *Authorizator) Remove(requisites *interfaces.Requisites) (id int, err error) {
// 	tx, err := authorizator.db.Begin()
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer func() {
// 		err = closeTransaction(tx, err)
// 	}()

// 	var password string
// 	err = tx.QueryRow("select id, password from users where name = ? limit 1", requisites.Login).Scan(&id, &password)
// 	if err := checkAuthorization(password, requisites.Password, err); err != nil {
// 		return 0, err
// 	}

// 	result, err := tx.Exec("DELETE FROM users (id,username,password) VALUES(DEFAULT,?,?)", requisites.Login, requisites.Password)

// 	err = tx.QueryRow("DELETE FROM users (id,username,password) VALUES(DEFAULT,?,?) RETURNING id",
// 		requisites.Login, requisites.Password).Scan(&id)
// 	if err := checkModification(id, err); err != nil {
// 		return 0, err
// 	}

// 	return 0, nil
// }

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
