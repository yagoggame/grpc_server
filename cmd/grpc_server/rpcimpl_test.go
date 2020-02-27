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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/yagoggame/grpc_server/mocks"
)

func TestEnterTheLobby(t *testing.T) {
	ontroller := gomock.NewController(t)
	defer controller.Finish()

	authorizator := mocks.NewMockAuthorizator(controller)
	pooler := mocks.NewMockPooler(controller)
	s := newServer(authorizator, pooler)
	defer s.Release()
}
