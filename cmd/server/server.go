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
	"strings"

	"github.com/yagoggame/grpc_server/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// ErrMissCred occurs when metadata missing credentials
	ErrMissCred = status.Error(codes.Unauthenticated, "missing credentials")
	// ErrServerCast occurs when UnaryInterceptor called not for the *Server
	ErrServerCast = status.Error(codes.Internal, "unable to cast server")
)

// IniDataContainer is a container of initial data to run server.
type IniDataContainer struct {
	Port       int
	IP         string
	CertFile   string
	KeyFile    string
	Authorizer string
	Filename   string
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
}

// private type for Context keys.
type contextKey int

// set of context keys
const (
	clientIDKey contextKey = iota
)

// authenticateAgent checks the client credentials.
func authenticateClient(ctx context.Context, s *Server) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, ErrMissCred
	}
	requisites := interfaces.Requisites{
		Login:    strings.Join(md["login"], ""),
		Password: strings.Join(md["password"], ""),
	}

	id, err := s.authorizator.Authorize(&requisites)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return id, nil
}

// authenticateAgent checks the client credentials.
func registerClient(ctx context.Context, s *Server) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrMissCred
	}

	requisites := interfaces.Requisites{
		Login:    strings.Join(md["login"], ""),
		Password: strings.Join(md["password"], ""),
	}

	err := s.authorizator.Register(&requisites)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	return nil
}

// UnaryInterceptor calls authenticateClient with current context.
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*Server)
	if !ok {
		return nil, ErrServerCast
	}

	if strings.HasSuffix(info.FullMethod, "RegisterUser") {
		err := registerClient(ctx, s)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}

	clientID, err := authenticateClient(ctx, s)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}
