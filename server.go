package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	pb "test-grpc-apigateway/proto"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

func (server) Echo(c context.Context, s *pb.StringMessage) (*pb.StringMessage, error) {
	fmt.Println(s.Value)
	return &pb.StringMessage{
		Value: strings.ToUpper(s.Value),
	}, nil
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func main() {

	grpcServer := grpc.NewServer()
	pb.RegisterEchoServiceServer(grpcServer, server{})
	ctx := context.Background()

	dopts := []grpc.DialOption{grpc.WithInsecure()}

	mux := http.NewServeMux()

	gwmux := runtime.NewServeMux()
	err := pb.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, ":5050", dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", ":5051")
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    "5050",
		Handler: grpcHandlerFunc(grpcServer, mux),
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}
