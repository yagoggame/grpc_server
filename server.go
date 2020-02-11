package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/yagoggame/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// private type for Context keys
type contextKey int

const (
	clientIDKey contextKey = iota
)

// authenticateAgent - checks the client credentials
func authenticateClient(ctx context.Context, s *Server) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")

		id, err := checkClientEntry(clientLogin, clientPassword)
		if err != nil {
			return "", err
		}
		return id, nil

	}
	return "", status.Error(codes.Unauthenticated, "missing credentials")
}

// unaryInterceptor - calls authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*Server)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to cast server")
	}

	clientID, err := authenticateClient(ctx, s)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func main() {
	// create a listener on TCP port 7777
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "localhost", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// create a server instance
	s := newServer()
	defer s.Release()

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile("../cert/server.crt", "../cert/server.key")
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}
	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor)}

	// create a gRPC server object
	grpcServer := grpc.NewServer(opts...)
	// attach the Ping service to the server
	api.RegisterGoGameServer(grpcServer, s)
	// start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
