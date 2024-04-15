#!/bin/bash -e

find . -type d -name example -prune -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete

mkdir -p proto/google/api

curl -s -f -o proto/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto || { echo "Error: Request failed"; exit 1; }

PROTOS=$(find . -type f -name '*.proto' | grep -v examples | grep -v proto/google/api | grep -v proto/toolkit )

for PROTO in $PROTOS; do
  protoc \
    -I. \
    -I$GOPATH/src \
    -I$(dirname $PROTO) \
    --go_out=paths=source_relative:. \
    $PROTO
done

rm -r proto

echo "Generation is ok"
