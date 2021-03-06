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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/grpc_server/interfaces"
	"github.com/yagoggame/grpc_server/interfaces/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	someLogin    string = "JoeLogin"
	somePassword        = "JoePassword"
	correctID           = 1
)

var usualRequisites = interfaces.Requisites{
	Login:    someLogin,
	Password: somePassword,
}

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
	fncGenServer func(interfaces.Authorizator, interfaces.Pooler, interfaces.GameGeter) interface{}
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
		ret:          &iderr{id: 0, err: interfaces.ErrLogin},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, interfaces.ErrLogin.Error())},
		ctx:          userContext(someLogin, somePassword)},
	{
		caseName: "wrong password", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: 0, err: interfaces.ErrPassword},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, interfaces.ErrPassword.Error())},
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

			gomock.InOrder(
				authorizator.EXPECT().
					Authorize(&usualRequisites).
					Return(test.ret.id, test.ret.err).
					Times(test.timesAuth),
				pooler.EXPECT().
					Release().
					Times(test.timesRel),
			)

			val, err := UnaryInterceptor(test.ctx, nil,
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
			s := NewServer(authorizator, pooler, nil)
			defer s.Release()

			want := gomockDependsOnSkiper(funcName == skipper, authorizator, pooler)

			_, err := UnaryInterceptor(userContext(someLogin, somePassword), nil,
				&grpc.UnaryServerInfo{Server: s, FullMethod: "SomePrefix" + funcName}, handler)
			testErr(t, err, want)
		})
	}
}

func gomockDependsOnSkiper(isSkiper bool,
	authorizator *mocks.MockAuthorizator, pooler *mocks.MockPooler) error {
	if isSkiper {
		gomock.InOrder(
			authorizator.EXPECT().
				Register(&usualRequisites).
				Return(nil).
				Times(1),
			pooler.EXPECT().
				Release().
				Times(1))
		return ErrGetIDFailed
	}
	gomock.InOrder(
		authorizator.EXPECT().
			Authorize(&usualRequisites).
			Return(0, interfaces.ErrLogin).
			Times(1),
		pooler.EXPECT().
			Release().
			Times(1))

	return status.Error(codes.Unauthenticated, interfaces.ErrLogin.Error())
}
