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
	"io/ioutil"
	"log"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster/game"
	gi "github.com/yagoggame/gomaster/game/interfaces"
	"github.com/yagoggame/grpc_server/interfaces"
	"github.com/yagoggame/grpc_server/interfaces/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	usualSize = 9
	usualKomi = 0.0
)

var JoinTheGameTests = []commonTestCase{
	{
		caseName: "No ID",
		times:    []int{0, 0, 0, 0, 0, 0, 1},
		ret:      []error{nil, nil, nil, nil, nil, nil, nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
	{
		caseName: "Normal",
		times:    []int{1, 1, 1, 0, 1, 1, 1},
		ret:      []error{nil, nil, nil, nil, nil, nil, nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "pool.JoinGame error",
		times:    []int{1, 0, 0, 0, 0, 0, 1},
		ret:      []error{ErrJoinGame, nil, nil, nil, nil, nil, nil},
		want:     ErrJoinGame,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "gameGeter error",
		times:    []int{1, 1, 0, 0, 0, 0, 1},
		ret:      []error{nil, ErrNoGamerID, nil, nil, nil, nil, nil},
		want:     ErrNoGamerID,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "waitGame error",
		times:    []int{1, 1, 1, 1, 0, 0, 1},
		ret:      []error{nil, nil, status.Errorf(codes.Canceled, "ERROR"), nil, nil, nil, nil},
		want:     status.Errorf(codes.Canceled, "ERROR"),
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "waitGame and release error",
		times:    []int{1, 1, 1, 1, 0, 0, 1},
		ret:      []error{nil, nil, status.Errorf(codes.Canceled, "ERROR"), errors.New("some internal error"), nil, nil, nil},
		want:     status.Errorf(codes.Canceled, "ERROR"),
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "FieldSize and release error",
		times:    []int{1, 1, 1, 0, 1, 0, 1},
		ret:      []error{nil, nil, nil, nil, errors.New("some internal error"), nil, nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "GameState and release error",
		times:    []int{1, 1, 1, 0, 1, 1, 1},
		ret:      []error{nil, nil, nil, nil, nil, errors.New("some internal error"), nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName:   "nil Game geted",
		nilManager: true,
		times:      []int{1, 1, 0, 0, 0, 0, 1},
		ret:        []error{nil, nil, nil, nil, nil, nil, nil},
		want:       ErrNilGame,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var WaitTheTurnTests = []commonTestCase{
	{
		caseName: "No ID",
		times:    []int{0, 0, 0, 0, 1},
		ret:      []error{nil, nil, nil, nil, nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
	{
		caseName: "Normal",
		times:    []int{1, 1, 1, 1, 1},
		ret:      []error{nil, nil, nil, nil, nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "gameGeter error",
		times:    []int{1, 0, 0, 0, 1},
		ret:      []error{ErrNoGamerID, nil, nil, nil, nil},
		want:     ErrNoGamerID,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "waitTurn error",
		times:    []int{1, 1, 0, 0, 1},
		ret:      []error{nil, status.Errorf(codes.Canceled, "ERROR"), nil, nil, nil},
		want:     status.Errorf(codes.Canceled, "ERROR"),
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "FieldSize and release error",
		times:    []int{1, 1, 1, 0, 1},
		ret:      []error{nil, nil, errors.New("some internal error"), nil, nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "GameState and release error",
		times:    []int{1, 1, 1, 1, 1},
		ret:      []error{nil, nil, nil, errors.New("some internal error"), nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName:   "nil Game geted",
		nilManager: true,
		times:      []int{1, 0, 0, 0, 1},
		ret:        []error{nil, nil, nil, nil, nil},
		want:       ErrNilGame,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var MakeTurnTests = []commonTestCase{
	{
		caseName: "No ID",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{0, 0, 0, 0, 1},
		ret:      []error{nil, nil, nil, nil, nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
	{
		caseName: "Normal",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 1, 1, 1, 1},
		ret:      []error{nil, nil, nil, nil, nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "gameGeter error",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 0, 0, 0, 1},
		ret:      []error{ErrNoGamerID, nil, nil, nil, nil},
		want:     ErrNoGamerID,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "waitTurn move error",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 1, 0, 0, 1},
		ret:      []error{nil, game.ErrWrongTurn, nil, nil, nil},
		want:     ErrWrongTurn,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "waitTurn regular error",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 1, 0, 0, 1},
		ret:      []error{nil, errors.New("some internal error"), nil, nil, nil},
		want:     ErrMakeTurn,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "FieldSize and release error",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 1, 1, 0, 1},
		ret:      []error{nil, nil, errors.New("some internal error"), nil, nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "GameState and release error",
		move:     &api.TurnMessage{X: 1, Y: 1},
		times:    []int{1, 1, 1, 1, 1},
		ret:      []error{nil, nil, nil, errors.New("some internal error"), nil},
		want:     ErrGameState,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName:   "nil Game geted",
		move:       &api.TurnMessage{X: 1, Y: 1},
		nilManager: true,
		times:      []int{1, 0, 0, 0, 1},
		ret:        []error{nil, nil, nil, nil, nil},
		want:       ErrNilGame,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

func TestJoinTheGame(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	for _, test := range JoinTheGameTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			gameGeter := mocks.NewMockGameGeter(controller)
			gameManager := mocks.NewMockGameManager(controller)
			s := NewServer(authorizator, pooler, gameGeter)
			defer s.Release()

			args := singleTestArgs{
				authorizator: authorizator,
				pooler:       pooler,
				test:         &test,
				s:            s,
				gameGeter:    gameGeter,
				gameManager:  gameManager,
				nilManager:   test.nilManager,
			}

			performTestJoinTheGame(t, args)
		})
	}
}

func TestWaitTheTurn(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	for _, test := range WaitTheTurnTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			gameGeter := mocks.NewMockGameGeter(controller)
			gameManager := mocks.NewMockGameManager(controller)
			s := NewServer(authorizator, pooler, gameGeter)
			defer s.Release()

			args := singleTestArgs{
				authorizator: authorizator,
				pooler:       pooler,
				test:         &test,
				s:            s,
				gameGeter:    gameGeter,
				gameManager:  gameManager,
				nilManager:   test.nilManager,
			}

			performTestWaitTheTurn(t, args)
		})
	}
}

func TestMakeTurn(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	for _, test := range MakeTurnTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			gameGeter := mocks.NewMockGameGeter(controller)
			gameManager := mocks.NewMockGameManager(controller)
			s := NewServer(authorizator, pooler, gameGeter)
			defer s.Release()

			args := singleTestArgs{
				authorizator: authorizator,
				pooler:       pooler,
				test:         &test,
				s:            s,
				gameGeter:    gameGeter,
				gameManager:  gameManager,
				nilManager:   test.nilManager,
			}

			performTestMakeTurn(t, args)
		})
	}
}

func performTestJoinTheGame(t *testing.T, args singleTestArgs) {
	var gm interfaces.GameManager
	if !args.nilManager {
		gm = args.gameManager
	}
	fieldState := genFieldState()

	gomock.InOrder(
		args.pooler.EXPECT().
			JoinGame(correctID, usualSize, usualKomi).
			Return(args.test.ret[0]).
			Times(args.test.times[0]),
		args.gameGeter.EXPECT().
			GetGame(correctID).
			Return(gm, args.test.ret[1]).
			Times(args.test.times[1]),
		args.gameManager.EXPECT().
			WaitBegin(args.test.ctx, correctID).
			Return(args.test.ret[2]).
			Times(args.test.times[2]),
		args.pooler.EXPECT().
			ReleaseGame(correctID).
			Return(args.test.ret[3]).
			Times(args.test.times[3]),
		args.gameManager.EXPECT().
			FieldSize(correctID).
			Return(usualSize, args.test.ret[4]).
			Times(args.test.times[4]),
		args.gameManager.EXPECT().
			GameState(correctID).
			Return(fieldState, args.test.ret[5]).
			Times(args.test.times[5]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[6]),
	)

	gameState, err := args.s.JoinTheGame(args.test.ctx, &api.EmptyMessage{})
	testErr(t, err, args.test.want)
	if err == nil {
		testGameState(t, gameState, fieldState)
	}
}

func performTestWaitTheTurn(t *testing.T, args singleTestArgs) {
	var gm interfaces.GameManager
	if !args.nilManager {
		gm = args.gameManager
	}
	fieldState := genFieldState()

	gomock.InOrder(
		args.gameGeter.EXPECT().
			GetGame(correctID).
			Return(gm, args.test.ret[0]).
			Times(args.test.times[0]),
		args.gameManager.EXPECT().
			WaitTurn(args.test.ctx, correctID).
			Return(args.test.ret[1]).
			Times(args.test.times[1]),
		args.gameManager.EXPECT().
			FieldSize(correctID).
			Return(usualSize, args.test.ret[2]).
			Times(args.test.times[2]),
		args.gameManager.EXPECT().
			GameState(correctID).
			Return(fieldState, args.test.ret[3]).
			Times(args.test.times[3]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[4]),
	)
	gameState, err := args.s.WaitTheTurn(args.test.ctx, &api.EmptyMessage{})
	testErr(t, err, args.test.want)
	if err == nil {
		testGameState(t, gameState, fieldState)
	}
}

func performTestMakeTurn(t *testing.T, args singleTestArgs) {
	var gm interfaces.GameManager
	if !args.nilManager {
		gm = args.gameManager
	}
	move := gi.TurnData{
		X: int(args.test.move.X),
		Y: int(args.test.move.Y),
	}
	fieldState := genFieldState()

	gomock.InOrder(
		args.gameGeter.EXPECT().
			GetGame(correctID).
			Return(gm, args.test.ret[0]).
			Times(args.test.times[0]),
		args.gameManager.EXPECT().
			MakeTurn(correctID, matchByTurnDataPtr(&move)).
			Return(args.test.ret[1]).
			Times(args.test.times[1]),
		args.gameManager.EXPECT().
			FieldSize(correctID).
			Return(usualSize, args.test.ret[2]).
			Times(args.test.times[2]),
		args.gameManager.EXPECT().
			GameState(correctID).
			Return(fieldState, args.test.ret[3]).
			Times(args.test.times[3]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[4]),
	)
	gameState, err := args.s.MakeTurn(args.test.ctx, args.test.move)
	testErr(t, err, args.test.want)
	if err == nil {
		testGameState(t, gameState, fieldState)
	}
}

func genFieldState() *gi.FieldState {
	state := &gi.FieldState{
		ChipsInCup:         make(map[gi.ChipColour]int, 2),
		ChipsCuptured:      make(map[gi.ChipColour]int, 2),
		PointsUnderControl: make(map[gi.ChipColour][]*gi.TurnData, 2),
		Scores:             make(map[gi.ChipColour]float64, 2),
		ChipsOnBoard:       make(map[gi.ChipColour][]*gi.TurnData, 2),
	}
	state.GameOver = true
	state.Komi = usualKomi

	colours := []gi.ChipColour{gi.White, gi.Black}

	for _, colour := range colours {
		state.ChipsInCup[colour] = 2
		state.ChipsOnBoard[colour] = []*gi.TurnData{&gi.TurnData{X: 10, Y: 11}}
		state.ChipsCuptured[colour] = 3
		state.PointsUnderControl[colour] = []*gi.TurnData{
			&gi.TurnData{X: 12, Y: 13},
			&gi.TurnData{X: 14, Y: 15}}
		state.Scores[colour] = float64(4)
	}
	return state
}
