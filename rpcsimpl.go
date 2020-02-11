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

package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster"
	"github.com/yagoggame/gomaster/game"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	// 	"log"
)

// Server represents the gRPC server
type Server struct {
	pool gomaster.GamersPool
}

// idFromCtx - gets id as integer from context
func idFromCtx(ctx context.Context) (id int, err error) {
	iid := ctx.Value(clientIDKey)
	if iid == nil {
		return -1, status.Errorf(codes.Internal, "can't get gamer's ID from context")
	}

	sid, ok := iid.(string)
	if !ok {
		return -1, status.Errorf(codes.Internal, "Unpredicted type %T of interface value", iid)
	}

	id, err = strconv.Atoi(sid)
	if err != nil {
		return -1, status.Errorf(codes.Internal, "can't convert gamer's ID to integer value: %s", err)
	}

	return id, nil
}

// EnterTheLobby - puts a player into the lobby
func (s *Server) EnterTheLobby(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	//get name
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := status.Errorf(codes.Unknown, "can't get metadata")
		log.Printf("EnterTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}
	name := strings.Join(md["login"], "")
	if len(name) < 1 {
		err := status.Errorf(codes.Unknown, "name of a player is to short to be accepted by a lobby")
		log.Printf("EnterTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's %s ID: %s", name, err)
		log.Printf("EnterTheLobby error: %s", err)

		return &api.EmptyMessage{}, err
	}

	gamer := game.Gamer{Name: name, Id: id}
	if err := s.pool.AddGamer(&gamer); err != nil {
		err = status.Errorf(codes.Internal, "can't add gamer %s to the Lobby: %s", &gamer, err)
		log.Printf("EnterTheLobby error: %s", err)

		return &api.EmptyMessage{}, err
	}

	log.Printf("gamer %s added to the Lobby", &gamer)
	return &api.EmptyMessage{}, nil
}

// LeaveTheLobby - deletes a player from the lobby
func (s *Server) LeaveTheLobby(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("LeaveTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gamer, err := s.pool.RmGamer(id)
	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, "%s", err)
		log.Printf("LeaveTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("gamer %s removed from the Lobby", gamer)
	return &api.EmptyMessage{}, nil
}

// JoinTheGame - joins a player to another player or start a game and wait of another player
func (s *Server) JoinTheGame(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gamer, err := s.pool.GetGamer(id)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to get a gamer with id %d: %s", id, err)
		log.Printf("WaitTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err = s.pool.JoinGame(id); err != nil {
		err = status.Errorf(codes.Internal, "failed to join the game for gamer %s: %s", gamer, err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer %s joined the game, awaiting of other player...", gamer)

	// if gamer joined to a game, he must wait when all preparation and other gamer are ready
	if err := gamer.InGame.WaitBegin(ctx, gamer); err != nil {
		err = status.Errorf(codes.Internal, "failed to wait game for gamer %v: %s", gamer, err)
		log.Printf("WaitTheGame error: %s", err)
		//gamer joined a game, so it's must be released
		if errl := s.pool.ReleaseGame(id); errl!=nil {
			err = status.Errorf(codes.Internal, "gamer %v: failed to Release game: %q, after failed game awaiting: %q:", gamer, errl, err)
			log.Printf("WaitTheGame error: %s", err)
			return &api.EmptyMessage{}, err
		}
		log.Printf("gamer %s left his game", gamer)
		return &api.EmptyMessage{}, err
	}
	log.Printf("Gamer's %s game has been begun", gamer)

	return &api.EmptyMessage{}, nil
}

// WaitTheTurn - if both gamers joined to a game, they must wait of their turn
func (s *Server) WaitTheTurn(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gamer, err := s.pool.GetGamer(id)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to get a gamer with id %d: %s", id, err)
		log.Printf("WaitTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("Gamer %s waiting for his turn...", gamer)
	if err := gamer.InGame.WaitTurn(ctx, gamer); err != nil {
		err = status.Errorf(codes.Internal, "failed to wait turn for gamer %v: %s", gamer, err)
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("Gamer's %s turn has been begun", gamer)

	return &api.EmptyMessage{}, nil
}

// LeaveTheGame - leave the game, but not a lobby, so JoinTheGame can be invoked again
func (s *Server) LeaveTheGame(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	//leave the gamer's game, if it is
	if err := s.pool.ReleaseGame(id); err!=nil {
		err = status.Errorf(codes.Internal, "gamer with id %d: failed to Release game: %q", id, err)
		log.Printf("WaitTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer with id %d left his game", id)

	return &api.EmptyMessage{}, nil
}

// MakeTurn - tries to make a turn with data of "in" message
func (s *Server) MakeTurn(ctx context.Context, in *api.TurnMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gamer, err := s.pool.GetGamer(id)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to get a gamer with id %d: %s", id, err)
		log.Printf("WaitTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	//leave the gamer's game, if it is
	if err := gamer.InGame.MakeTurn(gamer, &game.TurnData{X: int(in.X), Y: int(in.Y)}); err != nil {
		if err, ok := err.(*game.TurnError); ok {
			err := status.Errorf(codes.InvalidArgument, "%s", err)
			log.Printf("gamer %s made a wrong turn: %s", gamer, err)
			return &api.EmptyMessage{}, err
		}
		err = status.Errorf(codes.Internal, "fail on  MakeTurnGame for gamer %s : %s", gamer, err)
		log.Printf("MakeTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer %s made a turn: %v %v", gamer, in.X, in.Y)

	return &api.EmptyMessage{}, nil
}

// newServer - Create a new Server instance
// after using, it mast be destroyed by Release call
func newServer() *Server {
	return &Server{pool: gomaster.NewGamersPool()}
}

// Release - sop the server intity
func (s *Server) Release() {
	s.pool.Release()
}
