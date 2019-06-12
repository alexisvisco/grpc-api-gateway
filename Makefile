do:
	protoc -I=. \
            -I=$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
    		-I=$$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
            --grpc-gateway_out=logtostderr=true:. \
            --go_out=plugins=grpc:. \
    		--swagger_out=logtostderr=true:. \
    	     	./proto/*.proto