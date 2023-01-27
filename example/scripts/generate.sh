#!/bin/bash -e

SOURCE_PACKAGE=github.com/lastbackend/toolkit/example
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
PROTO_DIR=$ROOT_DIR/apis

find $ROOT_DIR -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete

mkdir -p $PROTO_DIR/google/api
mkdir -p $PROTO_DIR/validate

curl -s -o $PROTO_DIR/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
curl -s -o $PROTO_DIR/google/api/http.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
curl -s -o $PROTO_DIR/validate/validate.proto -L https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto

PROTOS=$(find $PROTO_DIR -type f -name '*.proto' | grep -v $PROTO_DIR/google/api | grep -v $PROTO_DIR/validate)

# Generate for toolkit service
for PROTO in $PROTOS; do
  protoc \
    -I. \
    -I$GOPATH/src \
    -I$PROTO_DIR \
    -I$(dirname $PROTO) \
    --validate_out=lang=go:$GOPATH/src \
    --go_out=:$GOPATH/src \
    --go-grpc_out=require_unimplemented_servers=false:$GOPATH/src \
    --toolkit_out=source_package=$SOURCE_PACKAGE:$GOPATH/src \
    $PROTO
done

rm -r $PROTO_DIR/google
rm -r $PROTO_DIR/validate
