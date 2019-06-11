PROTOS := proto/service.proto
INCLUDES := /usr/local/include

do:
	protoc -I$(INCLUDES) -I. -I$$GOPATH/src -I$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. $(PROTOS)
	protoc -I$(INCLUDES) -I. -I$$GOPATH/src -I$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. $(PROTOS)
	protoc -I$(INCLUDES) -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true:. $(PROTOS)