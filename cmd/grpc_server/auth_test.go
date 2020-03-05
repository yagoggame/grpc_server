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
	"testing"

	"github.com/golang/mock/gomock"
	server "github.com/yagoggame/grpc_server"
	"github.com/yagoggame/grpc_server/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	someLogin    string = "JoeLogin"
	somePassword        = "JoePassword"
	correctID           = 1
)

type iderr struct {
	id  int
	err error
}

var authorizationTests = []struct {
	caseName     string
	ret          *iderr
	want         *iderr
	timesAuth    int
	timesRel     int
	ctx          context.Context
	fncGenServer func(server.Authorizator, server.Pooler, server.GameGeter) interface{}
}{
	{
		caseName: "authorized user", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: correctID, err: nil},
		ctx:          userContext(someLogin, somePassword)},
	{
		caseName: "missing credentials", timesAuth: 0, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: 0, err: ErrMissCred},
		ctx:          context.Background()},
	{
		caseName: "wrong login", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: 0, err: server.ErrLogin},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, server.ErrLogin.Error())},
		ctx:          userContext(someLogin, somePassword)},
	{
		caseName: "wrong password", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: 0, err: server.ErrPassword},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, server.ErrPassword.Error())},
		ctx:          userContext(someLogin, somePassword)},
	{
		caseName: "fake server", timesAuth: 0, timesRel: 0,
		fncGenServer: genFakeSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: 0, err: ErrServerCast},
		ctx:          userContext(someLogin, somePassword)},
}

func TestUnaryInterceptor(t *testing.T) {
	for _, test := range authorizationTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			s := test.fncGenServer(authorizator, pooler, nil)
			if s, ok := s.(*Server); ok {
				defer s.Release()
			}

			requisites := server.Requisites{
				Login:    someLogin,
				Password: somePassword,
			}

			gomock.InOrder(
				authorizator.EXPECT().
					Authorize(&requisites).
					Return(test.ret.id, test.ret.err).
					Times(test.timesAuth),
				pooler.EXPECT().
					Release().
					Times(test.timesRel),
			)

			val, err := unaryInterceptor(test.ctx, nil,
				&grpc.UnaryServerInfo{Server: s}, handler)
			ival := transform(t, val, err)
			testIDErr(t, &iderr{id: ival, err: err}, test.want)
		})
	}
}

func TestUnarySkipAuth(t *testing.T) {
	funcNames := []string{"RegisterUser", "RemoveUser", "ChangeUserRequisits",
		"EnterTheLobby", "LeaveTheLobby", "JoinTheGame", "WaitTheTurn", "LeaveTheGame", "MakeTurn"}
	skipper := "RegisterUser"

	for _, funcName := range funcNames {
		t.Run(funcName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			s := newServer(authorizator, pooler, nil)
			defer s.Release()

			requisites := server.Requisites{
				Login:    someLogin,
				Password: somePassword,
			}

			var want iderr
			if funcName == skipper {
				want = iderr{id: correctID, err: nil}

				gomock.InOrder(
					authorizator.EXPECT().
						Register(&requisites).
						Return(correctID, nil).
						Times(1),
					pooler.EXPECT().
						Release().
						Times(1))
			} else {
				want = iderr{id: 0, err: status.Error(codes.Unauthenticated, server.ErrLogin.Error())}

				gomock.InOrder(
					authorizator.EXPECT().
						Authorize(&requisites).
						Return(0, server.ErrLogin).
						Times(1),
					pooler.EXPECT().
						Release().
						Times(1))
			}

			val, err := unaryInterceptor(userContext(someLogin, somePassword), nil,
				&grpc.UnaryServerInfo{Server: s, FullMethod: "SomePrefix" + funcName}, handler)
			ival := transform(t, val, err)
			testIDErr(t, &iderr{id: ival, err: err}, &want)
		})
	}
}
