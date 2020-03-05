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
	"io/ioutil"
	"log"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/api"
	"github.com/yagoggame/grpc_server/mocks"
)

var RegisterUserTests = []commonTestCase{
	{
		caseName: "No ID",
		times:    []int{1},
		ret:      []error{nil},
		want:     ErrGetIDFailed,
		ctx:      userContext(someLogin, somePassword)},
	{
		caseName: "Normal",
		times:    []int{1},
		ret:      []error{nil},
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
