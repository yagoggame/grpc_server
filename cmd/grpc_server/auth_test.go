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
	"testing"

	"github.com/golang/mock/gomock"
	server "github.com/yagoggame/grpc_server"
	"github.com/yagoggame/grpc_server/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	correctLogin    string = "JoeLogin"
	correctPassword        = "JoePassword"
	correctID              = 1
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
	fncGenServer func(server.Authorizator, server.Pooler) interface{}
}{
	{
		caseName: "authorized user", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: correctID, err: nil},
		ctx:          userContext(correctLogin, correctPassword)},
	{
		caseName: "missing credentials", timesAuth: 0, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: 0, err: ErrMissCred},
		ctx:          context.Background()},
	{
		caseName: "wrong ogin", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: 0, err: server.ErrLogin},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, server.ErrLogin.Error())},
		ctx:          userContext(correctLogin, correctPassword)},
	{
		caseName: "wrong password", timesAuth: 1, timesRel: 1,
		fncGenServer: genSrv,
		ret:          &iderr{id: 0, err: server.ErrPassword},
		want:         &iderr{id: 0, err: status.Error(codes.Unauthenticated, server.ErrPassword.Error())},
		ctx:          userContext(correctLogin, correctPassword)},
	{
		caseName: "fake server", timesAuth: 0, timesRel: 0,
		fncGenServer: genFakeSrv,
		ret:          &iderr{id: correctID, err: nil},
		want:         &iderr{id: 0, err: ErrServerCast},
		ctx:          userContext(correctLogin, correctPassword)},
}

func TestUnaryInterceptor(t *testing.T) {
	for _, test := range authorizationTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			s := test.fncGenServer(authorizator, pooler)
			if s, ok := s.(*Server); ok {
				defer s.Release()
			}

			gomock.InOrder(
				authorizator.EXPECT().
					Authorize(correctLogin, correctPassword).
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

func transform(t *testing.T, val interface{}, err error) int {
	if err != nil {
		return 0
	}
	id, ok := val.(int)
	if !ok {
		t.Fatalf("Unexpected id type:\nwant: %T,\ngot: %T.", 1, val)
	}

	return id
}

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	iid := ctx.Value(clientIDKey)
	if iid == nil {
		return 0, ErrGetIDFailed
	}

	return iid, nil
}

func testIDErr(t *testing.T, got, want *iderr) {
	if got.id != want.id {
		t.Errorf("Unexpected id:\nwant: %d,\ngot: %d.", want.id, got.id)
	}

	testErr(t, got.err, want.err)

}

func testErr(t *testing.T, got, want error) {
	if !errors.Is(got, want) {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want, got)
	}

	if status.Code(got) != status.Code(want) {
		t.Errorf("Unexpected error code:\nwant: %v,\ngot: %v.", status.Code(want), status.Code(got))
	}

}

func userContext(login, password string) context.Context {
	md := metadata.New(map[string]string{"login": correctLogin, "password": correctPassword})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return ctx
}

func genSrv(a server.Authorizator, p server.Pooler) interface{} {
	return newServer(a, p)

}

func genFakeSrv(a server.Authorizator, p server.Pooler) interface{} {
	return nil

}
