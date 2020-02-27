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

//go:generate mockgen -destination=./mocks/mock_grpc_server.go -package=mocks github.com/yagoggame/grpc_server Authorizator,Pooler

package server

import "github.com/yagoggame/gomaster/game"

// Authorizator requests id of user by login and password
type Authorizator interface {
	Authorize(login, password string) (id int, err error)
}

// Pooler provides interface to manage gomaster
type Pooler interface {
	AddGamer(gamer *game.Gamer) error
	RmGamer(id int) (gamer *game.Gamer, err error)
	ListGamers() []*game.Gamer
	JoinGame(id int) error
	ReleaseGame(id int) error
	GetGamer(id int) (*game.Gamer, error)
	Release()
}
