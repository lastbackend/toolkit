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

package engine

import (
	"gitlab.com/lastbackend/engine/client"
	"gitlab.com/lastbackend/engine/config"
	"gitlab.com/lastbackend/engine/server"
	"gitlab.com/lastbackend/engine/storage"
)

type Service interface {
	Name() string
	Version() string
	Options() Options
	Client() client.Client
	Server() server.Server
	Storage() storage.Storage
	Config() config.Config
	Start() error
	Stop() error
	Run() error
}

type Option func(*Options)

type serviceOptions struct {
}

func NewService(opts ...Option) (Service, error) {
	return newService(opts...)
}

// ServiceOption configures how we set up the service.
type ServiceOption interface {
	apply(*serviceOptions)
}

func WithServer(opt ServiceOption) {

}
