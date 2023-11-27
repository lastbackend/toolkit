#!/bin/bash -e

SOURCE_PACKAGE=github.com/lastbackend/toolkit/examples/gateway
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
PROTO_DIR=$ROOT_DIR/apis
SWAGGER_DIR_NAME=swagger

find $ROOT_DIR -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete

rm -rf ./${SWAGGER_DIR_NAME}

mkdir -p ${SWAGGER_DIR_NAME}
mkdir -p $PROTO_DIR/google/api
mkdir -p $PROTO_DIR/validate
mkdir -p $PROTO_DIR/protoc-gen-openapiv2/options

curl -s -o $PROTO_DIR/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
curl -s -o $PROTO_DIR/google/api/http.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
curl -s -o $PROTO_DIR/validate/validate.proto -L https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto
curl -s -o $PROTO_DIR/protoc-gen-openapiv2/options/annotations.proto -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-openapiv2/options/annotations.proto
curl -s -o $PROTO_DIR/protoc-gen-openapiv2/options/openapiv2.proto -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-openapiv2/options/openapiv2.proto

PROTOS=$(find $PROTO_DIR -type f -name '*.proto' | grep -v $PROTO_DIR/google/api | grep -v $PROTO_DIR/router/options)

# Generate for toolkit service
for PROTO in $PROTOS; do
  protoc \
    -I. \
    -I$GOPATH/src \
    -I$PROTO_DIR \
    -I$(dirname $PROTO) \
    --openapiv2_out ./${SWAGGER_DIR_NAME} --openapiv2_opt logtostderr=true \
    --validate_out=lang=go:$GOPATH/src \
    --go_out=:$GOPATH/src \
    --go-grpc_out=require_unimplemented_servers=false:$GOPATH/src \
    --toolkit_out=$GOPATH/src \
    $PROTO
done

mv ./${SWAGGER_DIR_NAME}/$SOURCE_PACKAGE/apis/* ./${SWAGGER_DIR_NAME}/

rm -rf ./${SWAGGER_DIR_NAME}/github.com
rm -r $PROTO_DIR/google
rm -r $PROTO_DIR/validate
rm -r $PROTO_DIR/protoc-gen-openapiv2
