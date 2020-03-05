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
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/api"
	server "github.com/yagoggame/grpc_server"
	"github.com/yagoggame/grpc_server/mocks"
)

var RegisterUserTests = []commonTestCase{
	{
		caseName: "No Cred",
		times:    []int{1},
		ret:      []error{nil},
		want:     ErrMissCred,
		ctx:      context.Background()},
	{
		caseName: "Normal",
		times:    []int{1},
		ret:      []error{nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var commonAuthTests = []commonTestCase{
	{
		caseName: "No Cred",
		times:    []int{0, 1},
		ret:      []error{nil},
		want:     ErrMissCred,
		ctx:      context.Background()},
	{
		caseName: "Empty login",
		times:    []int{0, 1},
		ret:      []error{nil, nil},
		want:     ErrLoginEmpty,
		ctx:      userContext("", somePassword)},
	{
		caseName: "Empty password",
		times:    []int{0, 1},
		ret:      []error{nil, nil},
		want:     ErrPasswordEmpty,
		ctx:      userContext(someLogin, "")},
	{
		caseName: "No ID",
		times:    []int{0, 1},
		ret:      []error{nil, nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
}

var removingUserTests = []commonTestCase{
	{
		caseName: "removing error",
		times:    []int{1, 1},
		ret:      []error{server.ErrPassword, nil},
		want:     ErrRemovingUser,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "removing success",
		times:    []int{1, 1},
		ret:      []error{nil, nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

var changingUserTests = []commonTestCase{
	{
		caseName: "changing error",
		times:    []int{1, 1},
		ret:      []error{server.ErrLoginOccupied, nil},
		want:     ErrChangeUser,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
	{
		caseName: "changing success",
		times:    []int{1, 1},
		ret:      []error{nil, nil},
		want:     nil,
		ctx: context.WithValue(userContext(someLogin, somePassword),
			clientIDKey, correctID)},
}

func TestRegisterUser(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	for _, test := range RegisterUserTests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			s := newServer(authorizator, pooler, nil)
			defer s.Release()

			pooler.EXPECT().
				Release().
				Times(test.times[0])

			_, err := s.RegisterUser(test.ctx, &api.EmptyMessage{})
			testErr(t, err, test.want)
		})
	}
}

func TestRemoveUser(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := composeTests(commonAuthTests, removingUserTests)

	for _, test := range tests {
		t.Run(test.caseName, func(t *testing.T) {
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

			gomock.InOrder(
				authorizator.EXPECT().
					Remove(matchByRequisites(&requisites)).
					Return(test.ret[0]).
					Times(test.times[0]),
				pooler.EXPECT().
					Release().
					Times(test.times[1]))

			_, err := s.RemoveUser(test.ctx, &api.EmptyMessage{})
			testErr(t, err, test.want)
		})
	}
}

func TestChangeUserRequisites(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := composeTests(commonAuthTests, changingUserTests)

	for _, test := range tests {
		t.Run(test.caseName, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			authorizator := mocks.NewMockAuthorizator(controller)
			pooler := mocks.NewMockPooler(controller)
			s := newServer(authorizator, pooler, nil)
			defer s.Release()

			requisitesOld := server.Requisites{
				Login:    someLogin,
				Password: somePassword,
			}
			requisitesNew := server.Requisites{
				Login:    someLogin + "New",
				Password: somePassword + "New",
			}

			gomock.InOrder(
				authorizator.EXPECT().
					ChangeRequisites(matchByRequisites(&requisitesOld), matchByRequisites(&requisitesNew)).
					Return(test.ret[0]).
					Times(test.times[0]),
				pooler.EXPECT().
					Release().
					Times(test.times[1]))

			_, err := s.ChangeUserRequisits(test.ctx, &api.RequisitsMessage{
				Login:    requisitesNew.Login,
				Password: requisitesNew.Password})
			testErr(t, err, test.want)
		})
	}
}

type byRequisites struct{ r *server.Requisites }

func matchByRequisites(r *server.Requisites) gomock.Matcher {
	return &byRequisites{r}
}

func (o *byRequisites) Matches(x interface{}) bool {
	r2, ok := x.(*server.Requisites)
	if !ok {
		return false
	}

	return reflect.DeepEqual(*o.r, *r2)
}

func (o *byRequisites) String() string {
	return fmt.Sprintf("has value %v", o.r)
}
