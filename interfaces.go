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

//go:generate mockgen -destination=./mocks/mock_grpc_server.go -package=mocks github.com/yagoggame/grpc_server Authorizator,Pooler,GameManager,GameGeter
//go:generate goimports -w ./mocks/mock_grpc_server.go

package server

import (
	"context"

	"github.com/yagoggame/gomaster/game"
)

// Authorizator is the interface that wraps Authorize method
//
// Authorize performs authorization of user by login and password
// and returns id of user in the case of success
type Authorizator interface {
	Authorize(requisites *Requisites) (id int, err error)
	Register(requisites *Requisites) error
	Remove(requisites *Requisites) error
	ChangeRequisites(requisitesOld, requisitesNew *Requisites) error
}

// Requisites contains login and password of user
type Requisites struct {
	Login    string
	Password string
}

// Pooler is the interface that groups the AddGamer, RmGamer, JoinGame,
// ReleaseGame, GetGamer, Release methods.
//
// AddGamer adds the gamer to a game
//
// RmGamer removes the gamer with specified id from a game.
//
// JoinerGame joins the gamer with specified id to the game
//
// ReleaseGame releases the game of gamer with specified id
//
// GetGamer gets the gamer with specified id from the pool
//
// Release releases the pool of gamer
// Release should be invoked after last use of this interface
type Pooler interface {
	AddGamer(gamer *game.Gamer) error
	RmGamer(id int) (gamer *game.Gamer, err error)
	JoinGame(id int) error
	ReleaseGame(id int) error
	GetGamer(id int) (*game.Gamer, error)
	Release()
}

// GameManager is the interface that groups the WaitBegin, WaitTurn,
// MakeTurn methods.
//
// WaitBegin awaits of game begin for the gamer with specified id
//
// WaitTurn awaits of turn begin for the gamer with specified id
//
// MakeTurn performs a move for the gamer with specified id
type GameManager interface {
	WaitBegin(ctx context.Context, id int) (err error)
	WaitTurn(ctx context.Context, id int) (err error)
	MakeTurn(id int, turn *game.TurnData) (err error)
}

// GameGeter is the interface that wraps the GetGame method.
//
// GetGame gets the gamer's game's interface
type GameGeter interface {
	GetGame(id int) (GameManager, error)
}
