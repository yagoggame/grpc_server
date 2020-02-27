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

//go:generate mockgen -destination=./mocks/mock_grpc_server.go -package=mocks github.com/yagoggame/grpc_server Authorizator,Pooler

package server

import "errors"

var (
	// ErrLogin occurs when login not recognized by Authorizator interface
	ErrLogin = errors.New("wrong login")
	// ErrPassword occurs when password not recognized by Authorizator interface
	ErrPassword = errors.New("wrong password")
)
