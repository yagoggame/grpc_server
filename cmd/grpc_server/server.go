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
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster"
	server "github.com/yagoggame/grpc_server"
	"github.com/yagoggame/grpc_server/authorization/dummy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// ErrMissCred occurs when metadata missing credentials
	ErrMissCred = status.Error(codes.Unauthenticated, "missing credentials")
	// ErrServerCast occurs when unaryInterceptor called not for the *Server
	ErrServerCast = status.Error(codes.Internal, "unable to cast server")
)

// iniDataContainertype is a container of initial data to run server.
type iniDataContainer struct {
	port     int
	ip       string
	certFile string
	keyFile  string
}

// init parses cmd line arguments into iniDataContainertype.
func (d *iniDataContainer) init() {
	flag.IntVar(&d.port, "p", 7777, "port")
	flag.StringVar(&d.ip, "a", "localhost", "server's host address")
	flag.StringVar(&d.certFile, "c", "../cert/server.crt", "tls certificate file")
	flag.StringVar(&d.keyFile, "k", "../cert/server.key", "tls key file")
	flag.Parse()
}

// private type for Context keys.
type contextKey int

// set of context keys
const (
	clientIDKey contextKey = iota
)

// authenticateAgent checks the client credentials.
func authenticateClient(ctx context.Context, s *Server) (int, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		requisites := server.Requisites{
			Login:    strings.Join(md["login"], ""),
			Password: strings.Join(md["password"], ""),
		}

		id, err := s.authorizator.Authorize(&requisites)
		if err != nil {
			return 0, status.Error(codes.Unauthenticated, err.Error())
		}
		return id, nil

	}
	return 0, ErrMissCred
}

// unaryInterceptor calls authenticateClient with current context.
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*Server)
	if !ok {
		return nil, ErrServerCast
	}

	if !strings.HasSuffix(info.FullMethod, "RegisterUser") {
		clientID, err := authenticateClient(ctx, s)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, clientIDKey, clientID)
	}

	return handler(ctx, req)
}

func main() {
	initData := &iniDataContainer{}
	initData.init()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", initData.ip, initData.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gamePool := gomaster.NewGamersPool()
	// gameGeter is separated from the object for testing purposes
	gameGeter := newGameGeter(gamePool)
	s := newServer(dummy.New(), gamePool, gameGeter)
	defer s.Release()

	creds, err := credentials.NewServerTLSFromFile(initData.certFile, initData.keyFile)
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor)}

	grpcServer := grpc.NewServer(opts...)
	api.RegisterGoGameServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
