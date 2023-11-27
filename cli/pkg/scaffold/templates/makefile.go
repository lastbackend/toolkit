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

// Makefile is the Makefile template used for new services.
var Makefile = `GOPATH:=$(shell go env GOPATH)

.PHONY: init
init:
	@sh ./scripts/bootstrap.sh

.PHONY: proto
proto:
	@sh ./scripts/generate.sh

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: build
build:
	@go build -o {{.Service}} *.go

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: gotest
gotest:
	@gotestsum --format testname

.PHONY: lint
lint:
	@golangci-lint run -v
`
