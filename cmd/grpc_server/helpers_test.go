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
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/gomaster/game"
	"github.com/yagoggame/grpc_server/interfaces"
	"github.com/yagoggame/grpc_server/interfaces/mocks"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
	if ((got == nil) != (want == nil)) ||
		!errors.Is(got, want) &&
			!strings.Contains(fmt.Sprintf("%v", got), fmt.Sprintf("%v", want)) {
		t.Errorf("Unexpected err:\nwant: %v,\ngot: %v.", want, got)
	}

	if status.Code(got) != status.Code(want) {
		t.Errorf("Unexpected error code:\nwant: %v,\ngot: %v.", status.Code(want), status.Code(got))
	}

}

func userContext(login, password string) context.Context {
	md := metadata.New(map[string]string{"login": login, "password": password})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return ctx
}

func genSrv(a interfaces.Authorizator, p interfaces.Pooler, g interfaces.GameGeter) interface{} {
	return NewServer(a, p, g)

}

func genFakeSrv(a interfaces.Authorizator, p interfaces.Pooler, g interfaces.GameGeter) interface{} {
	return nil

}

func composeTests(common []commonTestCase, specific []commonTestCase) []commonTestCase {
	tests := make([]commonTestCase, len(common), len(common)+len(specific))
	copy(tests, common)
	tests = append(tests, specific...)
	return tests
}

type byGamerPtr struct{ g *game.Gamer }

func matchByGamerPtr(g *game.Gamer) gomock.Matcher {
	return &byGamerPtr{g}
}

func (o *byGamerPtr) Matches(x interface{}) bool {
	g2, ok := x.(*game.Gamer)
	if !ok {
		return false
	}

	return reflect.DeepEqual(*o.g, *g2)
}

func (o *byGamerPtr) String() string {
	return "has value " + (*o.g).String()
}

type byTurnDataPtr struct{ m *game.TurnData }

func matchByTurnDataPtr(m *game.TurnData) gomock.Matcher {
	return &byTurnDataPtr{m}
}

func (o *byTurnDataPtr) Matches(x interface{}) bool {
	m2, ok := x.(*game.TurnData)
	if !ok {
		return false
	}

	return reflect.DeepEqual(*o.m, *m2)
}

func (o *byTurnDataPtr) String() string {
	return fmt.Sprintf("has value %v", o.m)
}

func launchVariants(t *testing.T, arg launchVariantsArgs) {
	for _, test := range arg.tests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			gameGeter := mocks.NewMockGameGeter(controller)
			_ = mocks.NewMockGameManager(controller)
			s := NewServer(authorizator, pooler, gameGeter)
			defer s.Release()

			args := singleTestArgs{
				authorizator: authorizator,
				pooler:       pooler,
				test:         &test,
				s:            s,
			}
			arg.fnc(t, args)
		})
	}
}
