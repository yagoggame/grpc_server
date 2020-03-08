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

package server

import (
	"fmt"

	"github.com/yagoggame/grpc_server/interfaces"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrNoGamerID occurs when failed to EnterTheLobby GetGamer(id).
	ErrNoGamerID = status.Errorf(codes.Internal, "failed to get a gamer")
)

// GameGeter implements GameGeter interface
// it is separated from the Server for testing purposes.
type GameGeter struct {
	pool interfaces.Pooler
}

// NewGameGeter creates a new GameGeter instance.
func NewGameGeter(pool interfaces.Pooler) *GameGeter {
	return &GameGeter{
		pool: pool,
	}
}

// GetGame method returns Game of Gamer with specified ID
// as interfaces.GameManager intyerface with testing purposes.
func (gg *GameGeter) GetGame(id int) (interfaces.GameManager, error) {
	gamer, err := gg.pool.GetGamer(id)
	if err != nil {
		err = extGrpcError(ErrNoGamerID, fmt.Sprintf(" with id %d: %v", id, err))
		return nil, err
	}
	return gamer.GetGame(), nil
}
