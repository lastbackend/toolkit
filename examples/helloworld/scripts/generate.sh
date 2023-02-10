#!/bin/bash -e

SOURCE_PACKAGE=github.com/lastbackend/toolkit/examples/helloworld
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
PROTO_DIR=$ROOT_DIR/apis

find $ROOT_DIR -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete

mkdir -p $PROTO_DIR/google/api

curl -s -o $PROTO_DIR/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
curl -s -o $PROTO_DIR/google/api/http.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto

PROTOS=$(find $PROTO_DIR -type f -name '*.proto' | grep -v $PROTO_DIR/google/api | grep -v $PROTO_DIR/validate)

# Generate GRPC service
for PROTO in $PROTOS; do
  protoc \
    -I. \
    -I$GOPATH/src \
    -I$PROTO_DIR \
    -I$(dirname $PROTO) \
    --go_out=:$GOPATH/src \
    --go-grpc_out=require_unimplemented_servers=false:$GOPATH/src \
    $PROTO
done

rm -r $PROTO_DIR/google
