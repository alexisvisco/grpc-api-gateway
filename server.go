package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	pb "test-grpc-apigateway/proto"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

func (server) EchoStream(input *pb.StringMessage, output pb.EchoService_EchoStreamServer) error {
	i := 0
	for {
		if i > 10 {
			output.Context().Done()
			break
		}
		err := output.Send(input)
		if err != nil {
			fmt.Println("stream error", err)
			output.Context().Done()
		}
		time.Sleep(2 * time.Second)
		i++
	}
	return nil
}

func (server) Echo(c context.Context, s *pb.StringMessage) (*pb.StringMessage, error) {
	md, _ := metadata.FromIncomingContext(c)
	fmt.Println("authorization: ", md.Get("authorization"))
	fmt.Println("user: ", c.Value("user"))

	return &pb.StringMessage{
		Value: strings.ToUpper(s.Value),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	registerGRPCServer(lis)
	registerGRPCGateway()
}

func registerGRPCServer(lis net.Listener) {
	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(userAuthenticationMiddleware)))
	pb.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func registerGRPCGateway() {
	fmt.Println("listening on 8080")
	if err := Run(":8080"); err != nil {
		glog.Fatal(err)
	}
}

func gatewayHandler(ctx context.Context) (http.Handler, error) {
	mux := runtime.NewServeMux(headerMatcher())
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterEchoServiceHandlerFromEndpoint(ctx, mux, ":5050", dialOpts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func Run(address string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", serveSwagger)

	gw, err := gatewayHandler(ctx)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)

	return http.ListenAndServe(address, allowCORS(mux))
}

func userAuthenticationMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Println("mw method", info.FullMethod)
	md, _ := metadata.FromIncomingContext(ctx)
	fmt.Println("authorization from mw: ", md.Get("authorization"))

	switch req.(type) {
	case *pb.StringMessage:
		fmt.Println("mw : StringMessage")
	}

	return handler(context.WithValue(ctx, "user", 159), req)
}

func headerMatcher() runtime.ServeMuxOption {
	return runtime.WithIncomingHeaderMatcher(func(s string) (s2 string, b bool) {
		return s, s == "authorization"
	})
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		glog.Errorf("Swagger JSON not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	glog.Infof("Serving %s", r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("proto", p)
	http.ServeFile(w, r, p)
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				setCorsHeaders(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func setCorsHeaders(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
	return
}
