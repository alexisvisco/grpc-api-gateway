generate:
	protoc -I=. \
            -I=$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
    		-I=$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
            --grpc-gateway_out=logtostderr=true:. \
            --go_out=plugins=grpc:. \
    		--swagger_out=logtostderr=true:. \
    	     	./proto/*.proto

install:
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/golang/protobuf/protoc-gen-go