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
	"io/ioutil"
	"log"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster/game"
	"github.com/yagoggame/grpc_server/mocks"
)

type commonTestCase struct {
	caseName   string
	times      []int
	ret        []error
	want       error
	ctx        context.Context
	nilManager bool
	move       *api.TurnMessage
}

type singleTestArgs struct {
	authorizator *mocks.MockAuthorizator
	pooler       *mocks.MockPooler
	gameGeter    *mocks.MockGameGeter
	gameManager  *mocks.MockGameManager
	test         *commonTestCase
	s            *Server
	nilManager   bool
}

type launchVariantsArgs struct {
	fnc   func(t *testing.T, args singleTestArgs)
	tests []commonTestCase
}

var commonRPCTests = []commonTestCase{
	{
		caseName: "No ID",
		times:    []int{0, 1},
		ret:      []error{nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
	{
		caseName: "Normal",
		times:    []int{1, 1},
		ret:      []error{nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var EnterTheLobbyTests = []commonTestCase{
	{
		caseName: "No Cred",
		times:    []int{0, 1},
		ret:      []error{nil},
		want:     ErrMissCred,
		ctx:      context.Background()},
	{
		caseName: "No Name",
		times:    []int{0, 1},
		ret:      []error{nil},
		want:     ErrLoginEmpty,
		ctx:      userContext("", "")},
	{
		caseName: "Main action fail",
		times:    []int{1, 1},
		ret:      []error{errors.New("some internal error")},
		want:     ErrAddGamer,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var LeaveTheLobbyTests = []commonTestCase{
	{
		caseName: "Main action fail",
		times:    []int{1, 1},
		ret:      []error{errors.New("some internal error")},
		want:     ErrLeaveLobby,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var LeaveTheGameTests = []commonTestCase{
	{
		caseName: "Main action fail",
		times:    []int{1, 1},
		ret:      []error{errors.New("some internal error")},
		want:     ErrReleaseGame,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

func TestEnterTheLobby(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := composeTests(commonRPCTests, EnterTheLobbyTests)

	args := launchVariantsArgs{
		fnc:   performTestEnterTheLobby,
		tests: tests,
	}

	launchVariants(t, args)
}

func TestLeaveTheLobby(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := composeTests(commonRPCTests, LeaveTheLobbyTests)

	args := launchVariantsArgs{
		fnc:   performTestLeaveTheLobby,
		tests: tests,
	}

	launchVariants(t, args)
}

func TestLeaveTheGame(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := composeTests(commonRPCTests, LeaveTheGameTests)

	args := launchVariantsArgs{
		fnc:   performTestLeaveTheGame,
		tests: tests,
	}

	launchVariants(t, args)
}

func performTestEnterTheLobby(t *testing.T, args singleTestArgs) {
	gomock.InOrder(
		args.pooler.EXPECT().
			AddGamer(matchByGamerPtr(&game.Gamer{Name: someLogin, ID: correctID})).
			Return(args.test.ret[0]).
			Times(args.test.times[0]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[1]),
	)
	_, err := args.s.EnterTheLobby(args.test.ctx, &api.EmptyMessage{})
	testErr(t, err, args.test.want)
}

func performTestLeaveTheLobby(t *testing.T, args singleTestArgs) {
	gomock.InOrder(
		args.pooler.EXPECT().
			RmGamer(correctID).
			Return(&game.Gamer{Name: someLogin, ID: correctID}, args.test.ret[0]).
			Times(args.test.times[0]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[1]),
	)
	_, err := args.s.LeaveTheLobby(args.test.ctx, &api.EmptyMessage{})
	testErr(t, err, args.test.want)
}

func performTestLeaveTheGame(t *testing.T, args singleTestArgs) {
	gomock.InOrder(
		args.pooler.EXPECT().
			ReleaseGame(correctID).
			Return(args.test.ret[0]).
			Times(args.test.times[0]),
		args.pooler.EXPECT().
			Release().
			Times(args.test.times[1]),
	)
	_, err := args.s.LeaveTheGame(args.test.ctx, &api.EmptyMessage{})
	testErr(t, err, args.test.want)
}
