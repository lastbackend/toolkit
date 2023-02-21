/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

// BootstrapScript is the bootstrap.sh script template used for new services.
var BootstrapScript = `#!/bin/sh

installGoDependencies() {
	echo "=> Install the protobuf for Go"
	go get -u google.golang.org/protobuf/proto

	echo "=> Install the protocol compiler plugins for Go"
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	echo "=> Install protoc-gen-toolkit into $GOPATH/bin"
	go get -d github.com/lastbackend/toolkit;
	go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest;

	echo "=> Install protoc-gen-validate into $GOPATH/bin"
	go get -d github.com/envoyproxy/protoc-gen-validate;
	go install github.com/envoyproxy/protoc-gen-validate;

	echo "=> Install protoc-gen-openapiv2 into $GOPATH/bin"
	go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2;

	echo "=> Install mockery into $GOPATH/bin"
	go install github.com/vektra/mockery/v2@latest
}

installForLinux() {
	echo "=> Configure Linux"
	sudo apt-get update;
	sudo apt-get install -y make protobuf-compiler;
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2;

	echo "Set up go dependencies\n"
	installGoDependencies;
}

installForDarwin() {
	echo "=> Configure OS X\n"
	brew install make;
	brew install protobuf;
	brew install golangci-lint;

	echo "Set up go dependencies\n"
	installGoDependencies;
}

unameOut="$(uname -s)"

case "${unameOut}" in
   Linux*)     installForLinux;;
   Darwin*)    installForDarwin;;
   *)          machine="UNKNOWN:${unameOut}"
esac
`

// GenerateScript is the generate.sh script template used for new services.
var GenerateScript = `#!/bin/bash -e
{{ if not .Vendor }}
SOURCE_PACKAGE={{.Service}}
ROOT_DIR=.
OUTPUT_DIR=.
{{ else }}
SOURCE_PACKAGE={{.Vendor}}{{.Service}}
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
OUTPUT_DIR=$GOPATH/src
{{ end -}}
PROTO_DIR=$ROOT_DIR/proto

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
    --validate_out=lang=go:$OUTPUT_DIR \
    --go_out=:$OUTPUT_DIR \
    --go-grpc_out=require_unimplemented_servers=false:$OUTPUT_DIR \
    --toolkit_out=$OUTPUT_DIR \
    $PROTO
done

rm -r $PROTO_DIR/google
rm -r $PROTO_DIR/validate
`
