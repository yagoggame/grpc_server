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
	"fmt"

	// registers pgx postgresSQL driver
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/yagoggame/grpc_server/interfaces"
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

	if err == sql.ErrNoRows {
		return 0, interfaces.ErrLogin
	}
	if err != nil {
		return 0, err
	}

	if requisites.Password != password {
		return 0, interfaces.ErrPassword
	}

	return id, nil
}

// Register attempts to register a new user and returns the id if success
func (authorizator *Authorizator) Register(requisites *interfaces.Requisites) (id int, err error) {
	err = authorizator.db.QueryRow("select id from users where name = ? limit 1", requisites.Login).Scan(&id)
	if err != sql.ErrNoRows {
		if err != nil {
			return 0, err
		}
		return 0, interfaces.ErrLoginOccupied
	}

	return 1, nil
}
