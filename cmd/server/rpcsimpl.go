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
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster/game"
	"github.com/yagoggame/gomaster/game/igame"
	"github.com/yagoggame/grpc_server/interfaces"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	standartSize = 9
	standartKomi = 0.0
)

var (
	// ErrGetIDFailed occurs when context not contains an ID
	ErrGetIDFailed = status.Errorf(codes.Internal, "can't get gamer's ID from context")
	// ErrLoginEmpty occurs login is empty
	ErrLoginEmpty = status.Errorf(codes.Internal, "lobby don't accepts empty login")
	// ErrPasswordEmpty occurs password is empty
	ErrPasswordEmpty = status.Errorf(codes.Internal, "lobby don't accepts empty password")
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
	// ErrIDMatching occurs when id of context not match with result of authorizator work
	ErrIDMatching = status.Errorf(codes.Internal, "id doesn't match")
	// ErrRemovingUser occurs when authorizator fails to remove user
	ErrRemovingUser = status.Errorf(codes.Unknown, "can't remove user")
	// ErrChangeUser occurs when authorizator fails to change user's requisites
	ErrChangeUser = status.Errorf(codes.Unknown, "can't change user requisites")
	// ErrGameState occurs when authorizatorgame manager failed to get game state
	ErrGameState = status.Errorf(codes.Internal, "can't get game state")
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
	pool         interfaces.Pooler
	authorizator interfaces.Authorizator
	gameGeter    interfaces.GameGeter
}

// NewServer Creates a new Server instance.
// After using, it mast be destroyed by Release call.
func NewServer(authorizator interfaces.Authorizator, pool interfaces.Pooler, gameGeter interfaces.GameGeter) *Server {
	return &Server{
		pool:         pool,
		authorizator: authorizator,
		gameGeter:    gameGeter,
	}
}

// RegisterUser provides registration of user by authorizator.
func (s *Server) RegisterUser(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &api.EmptyMessage{}, ErrMissCred
	}
	login := strings.Join(md["login"], "")

	log.Printf("user with login %q registred", login)
	return &api.EmptyMessage{}, nil
}

// RemoveUser provides removing of user by authorizator.
func (s *Server) RemoveUser(ctx context.Context, in *api.EmptyMessage) (*api.EmptyMessage, error) {
	requisites, id, err := requisitesFromContext(ctx)
	if err != nil {
		log.Printf("RegisterUser error: %s", err)
		return &api.EmptyMessage{}, err
	}

	err = s.authorizator.Remove(requisites)

	if err != nil {
		err := extGrpcError(ErrRemovingUser, fmt.Sprintf("user with login %q, id %d: %v", requisites.Login, id, err))
		log.Printf("RegisterUser error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("user with login %q, id %d removed", requisites.Login, id)

	return &api.EmptyMessage{}, nil
}

// ChangeUserRequisits provides change of requisits for user by authorizator.
func (s *Server) ChangeUserRequisits(ctx context.Context, requisits *api.RequisitsMessage) (*api.EmptyMessage, error) {
	requisitesOld, id, err := requisitesFromContext(ctx)
	if err != nil {
		log.Printf("ChangeUserRequisits error: %s", err)
		return &api.EmptyMessage{}, err
	}

	requisitesNew := &interfaces.Requisites{
		Login:    requisits.Login,
		Password: requisits.Password,
	}

	err = s.authorizator.ChangeRequisites(requisitesOld, requisitesNew)

	if err != nil {
		err := extGrpcError(ErrChangeUser, fmt.Sprintf("user with login %q, id %d (new login %q): %v",
			requisitesOld.Login, id, requisitesNew.Login, err))
		log.Printf("ChangeUserRequisits error: %s", err)
		return &api.EmptyMessage{}, err
	}

	log.Printf("user with login %q, id %d changed his requisites (new login %q)",
		requisitesOld.Login, id, requisitesNew.Login)

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
func (s *Server) JoinTheGame(ctx context.Context, in *api.EmptyMessage) (*api.State, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("JoinTheGame error: %s", err)
		return &api.State{}, err
	}

	if err := s.pool.JoinGame(id, standartSize, standartKomi); err != nil {
		err := extGrpcError(ErrJoinGame, err.Error())
		log.Printf("JoinTheGame error: %s", err)
		return &api.State{}, err
	}

	state, err := s.waitGame(ctx, id)
	if err != nil {
		log.Printf("JoinTheGame error: %s", err)
		return &api.State{}, err
	}

	log.Printf("game for gamer with id %d has been begun", id)
	return state, nil
}

// WaitTheTurn waits for gamers turn.
func (s *Server) WaitTheTurn(ctx context.Context, in *api.EmptyMessage) (*api.State, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.State{}, err
	}

	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.State{}, err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		log.Printf("WaitTheTurn error: %s", err)
		return &api.State{}, err
	}

	log.Printf("Gamer with id %d waiting for his turn...", id)
	state, err := s.waitTurn(ctx, gameManager, id)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.State{}, err
	}

	log.Printf("turn of gamer with id %d has been begun", id)

	return state, nil
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
func (s *Server) MakeTurn(ctx context.Context, in *api.TurnMessage) (*api.State, error) {
	id, err := idFromCtx(ctx)
	if err != nil {
		log.Printf("JoinTheGame error: %s", err)
		return &api.State{}, err
	}

	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		log.Printf("MakeTurn error: %s", err)
		return &api.State{}, err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		log.Printf("MakeTurn error: %s", err)
		return &api.State{}, err
	}

	state, err := s.makeTurn(gameManager, id,
		&igame.TurnData{X: int(in.X), Y: int(in.Y)})
	if err != nil {
		log.Printf("MakeTurn error: %s", err)
		return &api.State{}, err
	}

	log.Printf("gamer with id %d made a turn: %v %v", id, in.X, in.Y)

	return state, nil
}

// Release sops the server intity.
func (s *Server) Release() {
	s.pool.Release()
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

func userFromContext(ctx context.Context) (gamer *game.Gamer, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMissCred
	}
	login := strings.Join(md["login"], "")
	if len(login) < 1 {
		return nil, ErrLoginEmpty
	}

	id, err := idFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	return &game.Gamer{Name: login, ID: id}, nil
}

func requisitesFromContext(ctx context.Context) (requisites *interfaces.Requisites, id int, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, 0, ErrMissCred
	}
	requisites = &interfaces.Requisites{
		Login:    strings.Join(md["login"], ""),
		Password: strings.Join(md["password"], ""),
	}
	if len(requisites.Login) < 1 {
		return nil, 0, ErrLoginEmpty
	}
	if len(requisites.Password) < 1 {
		return nil, 0, ErrPasswordEmpty
	}

	id, err = idFromCtx(ctx)
	if err != nil {
		return nil, 0, err
	}

	return requisites, id, nil
}

func (s *Server) waitGame(ctx context.Context, id int) (*api.State, error) {
	gameManager, err := s.gameGeter.GetGame(id)
	if err != nil {
		return &api.State{}, err
	}
	if gameManager == nil {
		err = extGrpcError(ErrNilGame, fmt.Sprintf(" with id %d: %v", id, err))
		return &api.State{}, err
	}

	if err := gameManager.WaitBegin(ctx, id); err != nil {
		//gamer joined a game, so it's must be released.
		if errl := s.pool.ReleaseGame(id); errl != nil {
			err = extGrpcError(err, fmt.Sprintf(", gamer with id %d: failed to Release game: %q, after failed game awaiting", id, errl))
			return &api.State{}, err
		}
		return &api.State{}, err
	}

	state, err := s.getGameState(gameManager, id)
	if err != nil {
		return &api.State{}, err
	}

	return state, nil
}

func (s *Server) waitTurn(ctx context.Context, gameManager interfaces.GameManager, id int) (*api.State, error) {
	if err := gameManager.WaitTurn(ctx, id); err != nil {
		return &api.State{}, err
	}

	state, err := s.getGameState(gameManager, id)
	if err != nil {
		return &api.State{}, err
	}
	return state, nil
}

func (s *Server) makeTurn(gameManager interfaces.GameManager, id int, move *igame.TurnData) (*api.State, error) {
	if err := gameManager.MakeTurn(id, move); err != nil {
		if errors.Is(err, game.ErrWrongTurn) {
			err = extGrpcError(ErrWrongTurn, fmt.Sprintf(" with id %d: %v", id, err))
			return &api.State{}, err
		}
		err = extGrpcError(ErrMakeTurn, fmt.Sprintf(" with id %d: %v", id, err))
		return &api.State{}, err
	}

	state, err := s.getGameState(gameManager, id)
	if err != nil {
		log.Printf("WaitTheTurn error: %s", err)
		return &api.State{}, err
	}
	return state, nil
}

func (s *Server) getGameState(gameManager interfaces.GameManager, id int) (*api.State, error) {
	gameState := &api.State{
		Black: &api.State_ColourState{},
		White: &api.State_ColourState{},
	}

	fs, err := gameManager.FieldSize(id)
	if err != nil {
		err := extGrpcError(ErrGameState, fmt.Sprintf("user with id %d on FieldSize: %v", id, err))
		return &api.State{}, err
	}
	gameState.Size = int64(fs)

	state, err := gameManager.GameState(id)
	if err != nil {
		err := extGrpcError(ErrGameState, fmt.Sprintf("user with id %d on GameState: %v", id, err))
		return &api.State{}, err
	}

	gameState.Komi = state.Komi
	gameState.GameOver = state.GameOver
	fillForColour(gameState.White, state, igame.White)
	fillForColour(gameState.Black, state, igame.Black)

	return gameState, nil
}

func fillForColour(apiState *api.State_ColourState, state *igame.FieldState, colour igame.ChipColour) {
	apiState.ChipsCaptured = int64(state.ChipsCuptured[colour])
	apiState.ChipsInCap = int64(state.ChipsInCup[colour])
	apiState.Scores = state.Scores[colour]

	chipsOnBoard := make([]*api.TurnMessage, len(state.ChipsOnBoard[colour]))
	for i, pt := range state.ChipsOnBoard[colour] {
		chipsOnBoard[i] = &api.TurnMessage{
			X: int64(pt.X),
			Y: int64(pt.Y),
		}
	}
	apiState.ChipsOnBoard = chipsOnBoard

	pointsUnderControl := make([]*api.TurnMessage, len(state.PointsUnderControl[colour]))
	for i, pt := range state.PointsUnderControl[colour] {
		pointsUnderControl[i] = &api.TurnMessage{
			X: int64(pt.X),
			Y: int64(pt.Y),
		}
	}
	apiState.PointsUnderControl = pointsUnderControl
}
