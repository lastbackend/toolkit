#!/bin/bash -e

find . -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete
PROTOS=$(find . -type f -name '*.proto' | grep -v proto/google/api)

mkdir -p proto/google/api
curl -s -o proto/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto

for PROTO in $PROTOS; do
  echo $PROTO
  protoc --experimental_allow_proto3_optional \
    -I./proto \
    -I. \
    -I$(dirname $PROTO) \
    --go_out=paths=source_relative:. \
    $PROTO
done

rm -r proto
