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
	"fmt"
	"log"
	"strings"

	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster/game"
	server "github.com/yagoggame/grpc_server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// ErrGetIDFailed occurs when context not contains an ID
	ErrGetIDFailed = status.Errorf(codes.Internal, "can't get gamer's ID from context")
	// ErrNameEmpty occurs login is empty
	ErrNameEmpty = status.Errorf(codes.Internal, "lobby don't accept empty name(login)")
	// ErrWrongIDType occurs when context not contains an ID of valid type
	ErrWrongIDType = status.Errorf(codes.Internal, "unpredicted type of interface value")
	// ErrAddGamer occurs when failed to EnterTheLobby
	ErrAddGamer = status.Errorf(codes.Internal, "can't add gamer to the Lobby")
	// ErrLeaveLobby occurs when failed to LeaveTheLobby
	ErrLeaveLobby = status.Errorf(codes.Internal, "can't leave the lobby")
	// ErrJoinGame occurs when failed to JoinGame
	ErrJoinGame = status.Errorf(codes.Internal, "can't join the game")
	// ErrWaitGame occurs when failed to wait a game begin
	ErrWaitGame = status.Errorf(codes.Internal, "can't await the game begin")
	// ErrNilGame occurs when game obtained by gamer.GetGame is nil
	ErrNilGame = status.Errorf(codes.Internal, "nil game getted for a gamer")
	// ErrReleaseGame occurs when failed to ReleaseGame
	ErrReleaseGame = status.Errorf(codes.Internal, "can't release Game")
	// ErrMakeTurn occurs when failed to MakeTurn
	ErrMakeTurn = status.Errorf(codes.Internal, "can't make turn for a gamer")
	// ErrWrongTurn occurs when wrong data providet to MakeTurn
	ErrWrongTurn = status.Errorf(codes.InvalidArgument, "wrong turn for a gamer")
)

func extGrpcError(err error, ext string) error {
	st, ok := status.FromError(err)
	if !ok {
		return status.Errorf(codes.Unknown, fmt.Errorf("%w: %s", err, ext).Error())
	}
	return status.Errorf(st.Code(), fmt.Errorf("%s: %s", st.Message(), ext).Error())
}

// Server represents the gRPC server.
type Server struct {
	pool         server.Pooler
	authorizator server.Authorizator
	gameGeter    server.GameGeter
}

// newServer Creates a new Server instance.
// After using, it mast be destroyed by Release call.
func newServer(authorizator server.Authorizator, pool server.Pooler, gameGeter server.GameGeter) *Server {
	return &Server{
		pool:         pool,
		authorizator: authorizator,
		gameGeter:    gameGeter,
	}
}

// idFromCtx gets id as integer from context.
func idFromCtx(ctx context.Context) (id int, err error) {
	iid := ctx.Value(clientIDKey)
	if iid == nil {
		return 0, ErrGetIDFailed
	}

	id, ok := iid.(int)
	if !ok {
		return 0, fmt.Errorf("%w: %T", ErrWrongIDType, iid)
	}

	return id, nil
}

// RegisterUser provides registration of user by authorizator.
func (s *Server) RegisterUser(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	_, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("RegisterUser error: %s", err)
		return &api.EmptyMessage{}, err
	}

	return &api.EmptyMessage{}, nil
}

// RemoveUser provides removing of user by authorizator.
func (s *Server) RemoveUser(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	return &api.EmptyMessage{}, nil
}

// ChangeUserRequisits provides change of requisits for user by authorizator.
func (s *Server) ChangeUserRequisits(ctx context.Context, requisits *api.RequisitsMessage) (*api.EmptyMessage, error) {
	return &api.EmptyMessage{}, nil
}

// EnterTheLobby puts a player into the lobby.
func (s *Server) EnterTheLobby(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	gamer, err := userFromContext(ctx)
	if err != nil {
		log.Printf("EnterTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.pool.AddGamer(gamer); err != nil {
		err := extGrpcError(ErrAddGamer, err.Error())
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
		log.Printf("LeaveTheLobby error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gamer, err := s.pool.RmGamer(id)
	if err != nil {
		err := extGrpcError(ErrLeaveLobby, err.Error())
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
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.pool.JoinGame(id); err != nil {
		err := extGrpcError(ErrJoinGame, err.Error())
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	if err := s.waitGame(ctx, id); err != nil {
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("game for gamer with id %d has been begun", id)
	return &api.EmptyMessage{}, nil
}

// WaitTheTurn waits for gamers turn.
func (s *Server) WaitTheTurn(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("Gamer with id %d waiting for his turn...", id)
	if err := gameManager.WaitTurn(ctx, id); err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("turn of gamer with id %d has been begun", id)

	return &api.EmptyMessage{}, nil
}

// LeaveTheGame exits the gamer from the game.
// Gamer stays in lobby, so JoinTheGame can be invoked again.
func (s *Server) LeaveTheGame(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("LeaveTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	//leave the gamer's game, if it is.
	if err := s.pool.ReleaseGame(id); err != nil {
		err = extGrpcError(ErrReleaseGame, fmt.Sprintf("failed to ReleaseGame for gamer with id %d: %v", id, err))
		log.Printf("LeaveTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer with id %d left his game", id)

	return &api.EmptyMessage{}, nil
}

// MakeTurn tries to make a turn with data of "in" message.
func (s *Server) MakeTurn(ctx context.Context, in *api.TurnMessage) (*api.EmptyMessage, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("JoinTheGame error: %s", err)
		return &api.EmptyMessage{}, err
	}

	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		log.Printf("MakeTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		log.Printf("MakeTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}

	err = makeTurn(gameManager, id,
		&game.TurnData{X: int(in.X), Y: int(in.Y)})
	if err != nil {
		log.Printf("MakeTurn error: %s", err)
		return &api.EmptyMessage{}, err
	}
	log.Printf("gamer with id %d made a turn: %v %v", id, in.X, in.Y)

	return &api.EmptyMessage{}, nil
}

// Release sops the server intity.
func (s *Server) Release() {
	s.pool.Release()
}

func userFromContext(ctx context.Context) (gamer *game.Gamer, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMissCred
	}
	name := strings.Join(md["login"], "")
	if len(name) < 1 {
		return nil, ErrNameEmpty
	}

	id, err := idFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	return &game.Gamer{Name: name, ID: id}, nil
}

func (s *Server) waitGame(ctx context.Context, id int) error {
	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		return err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		return err
	}

	if err := gameManager.WaitBegin(ctx, id); err != nil {
		//gamer joined a game, so it's must be released.
		if errl := s.pool.ReleaseGame(id); errl != nil {
			err = extGrpcError(err, fmt.Sprintf(", gamer with id %d: failed to Release game: %q, after failed game awaiting", id, errl))
			return err
		}
		return err
	}
	return nil
}

func makeTurn(gm server.GameManager, id int, move *game.TurnData) error {
	if err := gm.MakeTurn(id, move); err != nil {
		if errors.Is(err, game.ErrWrongTurn) {
			err = extGrpcError(ErrWrongTurn, fmt.Sprintf(" with id %d: %v", id, err))
			return err
		}
		err = extGrpcError(ErrMakeTurn, fmt.Sprintf(" with id %d: %v", id, err))
		return err
	}
	return nil
}
