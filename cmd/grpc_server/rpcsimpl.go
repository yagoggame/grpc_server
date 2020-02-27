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
	"errors"
	"log"
	"strings"

	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster/game"
	server "github.com/yagoggame/grpc_server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server represents the gRPC server.
type Server struct {
	pool         server.Pooler
	authorizator server.Authorizator
}

// newServer Creates a new Server instance.
// After using, it mast be destroyed by Release call.
func newServer(authorizator server.Authorizator, pool server.Pooler) *Server {
	return &Server{
		pool:         pool,
		authorizator: authorizator,
	}
}

// idFromCtx gets id as integer from context.
func idFromCtx(ctx context.Context) (id int, err error) {
	iid := ctx.Value(clientIDKey)
	if iid == nil {
		return -1, status.Errorf(codes.Internal, "can't get gamer's ID from context")
	}

	id, ok := iid.(int)
	if !ok {
		return -1, status.Errorf(codes.Internal, "unpredicted type %T of interface value", iid)
	}

	return id, nil
}

// EnterTheLobby puts a player into the lobby.
func (s *Server) EnterTheLobby(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	gamer, err := userFromContext(ctx)
	if err != nil {
		log.Printf("EnterTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.pool.AddGamer(gamer); err != nil {
		err = status.Errorf(codes.Internal, "can't add gamer %s to the Lobby: %s", gamer, err)
		log.Printf("EnterTheLobby error: %s", err)

		return &api.EmptyMessage{}, err
	}

	log.Printf("gamer %s added to the Lobby", gamer)
	return &api.EmptyMessage{}, nil
}

// LeaveTheLobby deletes a player from the lobby.
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

// JoinTheGame joins a player to another player or starts a game and waits of another player.
func (s *Server) JoinTheGame(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.joinGamer(id); err != nil {
		log.Printf("joinGamer error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.waitGame(ctx, id); err != nil {
		log.Printf("waitGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("game for gamer with id %d has been begun", id)
	return &api.EmptyMessage{}, nil
}

// WaitTheTurn waits for gamers turn.
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
	if err := gamer.InGame.WaitTurn(ctx, gamer.ID); err != nil {
		err = status.Errorf(codes.Internal, "failed to wait turn for gamer %v: %s", gamer, err)
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("Gamer's %s turn has been begun", gamer)

	return &api.EmptyMessage{}, nil
}

// LeaveTheGame exits the gamer from the game.
// Gamer stays in lobby, so JoinTheGame can be invoked again.
func (s *Server) LeaveTheGame(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's ID: %s", err)
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	//leave the gamer's game, if it is.
	if err := s.pool.ReleaseGame(id); err != nil {
		err = status.Errorf(codes.Internal, "gamer with id %d: failed to Release game: %q", id, err)
		log.Printf("WaitTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer with id %d left his game", id)

	return &api.EmptyMessage{}, nil
}

// MakeTurn tries to make a turn with data of "in" message.
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

	err = makeTurn(gamer,
		&game.TurnData{X: int(in.X), Y: int(in.Y)})
	if err != nil {
		log.Printf("MakeTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer %s made a turn: %v %v", gamer, in.X, in.Y)

	return &api.EmptyMessage{}, nil
}

// Release - sop the server intity.
func (s *Server) Release() {
	s.pool.Release()
}

func userFromContext(ctx context.Context) (gamer *game.Gamer, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := status.Errorf(codes.Unknown, "can't get metadata")
		return nil, err
	}
	name := strings.Join(md["login"], "")
	if len(name) < 1 {
		err := status.Errorf(codes.Unknown, "name of a player is to short to be accepted by a lobby")
		return nil, err
	}

	id, err := idFromCtx(ctx)
	if err != nil {
		err = status.Errorf(codes.Unknown, "can't get gamer's %s ID: %s", name, err)
		return nil, err
	}

	return &game.Gamer{Name: name, ID: id}, nil
}

func (s *Server) joinGamer(id int) error {
	gamer, err := s.pool.GetGamer(id)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to get a gamer with id %d: %s", id, err)
		return err
	}

	if err := s.pool.JoinGame(gamer.ID); err != nil {
		err = status.Errorf(codes.Internal, "failed to join the game for gamer %s: %s", gamer, err)
		return err
	}
	return nil
}

func (s *Server) waitGame(ctx context.Context, id int) error {
	gamer, err := s.pool.GetGamer(id)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to get a gamer with id %d: %s", id, err)
		return err
	}

	if err := gamer.InGame.WaitBegin(ctx, gamer.ID); err != nil {
		err = status.Errorf(codes.Internal, "failed to wait game for gamer %v: %s", gamer, err)
		//gamer joined a game, so it's must be released.
		if errl := s.pool.ReleaseGame(gamer.ID); errl != nil {
			err = status.Errorf(codes.Internal, "gamer %v: failed to Release game: %q, after failed game awaiting: %q:", gamer, errl, err)
			return err
		}
		return err
	}
	return nil
}

func makeTurn(gamer *game.Gamer, move *game.TurnData) error {
	if err := gamer.InGame.MakeTurn(gamer.ID, move); err != nil {
		if errors.Is(err, game.ErrWrongTurn) {
			err := status.Errorf(codes.InvalidArgument, "%s", err)
			return err
		}
		err = status.Errorf(codes.Internal, "fail on  MakeTurnGame for gamer %s : %s", gamer, err)
		return err
	}
	return nil
}
