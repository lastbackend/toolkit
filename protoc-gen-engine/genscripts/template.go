/*
Copyright [2014] - [2021] The Last.Backend authors.

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

package genscripts

import (
	"bytes"
	"text/template"
)

type tplDockerfileOptions struct {
}

func applyDockerfileTemplate(to tplDockerfileOptions) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := dockerfileTemplate.Execute(w, to); err != nil {
		return "", err
	}

	return w.String(), nil
}

type tplMakefileOptions struct {
}

func applyMakefileTemplate(to tplMakefileOptions) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := makefileTemplate.Execute(w, to); err != nil {
		return "", err
	}

	return w.String(), nil
}

var (
	dockerfileTemplate = template.Must(template.New("dockerfile").Parse(
		`# Script generated by protoc-gen-engine.
# Build manifest
FROM golang:1.16.6-alpine3.13 as build

RUN apk add --no-cache ca-certificates \
  linux-headers \
  gcc \
  musl-dev

RUN set -ex \
	&& apk add --no-cache --virtual .build-deps \
    bash \
    git \
    make \
	\
	&& rm -rf /*.patch

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV GO111MODULE on
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD . $GOPATH/src/ gitlab.com/dummy/dummy
WORKDIR $GOPATH/src/gitlab.com/dummy/dummy

RUN make build && make install
RUN apk del --purge .build-deps

WORKDIR $GOPATH/bin

RUN rm -rf $GOPATH/pkg \
    && rm -rf $GOPATH/src \
    && rm -rf /var/cache/apk/*


# Release manifest
FROM alpine:3.13 as production

RUN apk add --no-cache ca-certificates

COPY --from=build /usr/bin/service /usr/bin/service

EXPOSE 80 443 50005

CMD ["/usr/bin/service"]

`))

	makefileTemplate = template.Must(template.New("makefile").Parse(
		`# Script generated by protoc-gen-engine.
.PHONY : default deps test build install

default: deps test build install

deps:
	@echo "Configuring dependencies for service"
	#Here you need to describe the script for installing dependencies for your service. 

test:
	@echo "Testing service"
	#Here you need to describe the script for testing your service.

build:
	@echo "Building service"
	#Here you need to describe the script for building your service.

install:
	@echo "== Install binaries"
	#Here you need to describe the script for install binaries your service.

`))
)
